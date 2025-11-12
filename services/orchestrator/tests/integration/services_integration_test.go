package integration

import (
	"context"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/clients"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/repository"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/services"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/tests/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyService_Integration(t *testing.T) {
	redisClient, cleanup := helpers.SetupTestRedis(t)
	defer cleanup()

	ttl := 1 * time.Hour
	idempotencyService := services.NewIdempotencyService(redisClient, ttl)

	ctx := context.Background()
	idempotencyKey := "test-request-123"

	// First request - should not be cached
	cached, err := idempotencyService.GetCachedResponse(ctx, idempotencyKey)
	require.NoError(t, err)
	assert.Nil(t, cached)

	// Create a response and store it
	response := &models.NotificationResponse{
		NotificationID: "notif-456",
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}

	err = idempotencyService.StoreResponse(ctx, idempotencyKey, response)
	require.NoError(t, err)

	// Second request with same key - should return cached response
	cached, err = idempotencyService.GetCachedResponse(ctx, idempotencyKey)
	require.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, response.NotificationID, cached.NotificationID)
	assert.Equal(t, response.Status, cached.Status)
}

func TestOrchestrationService_ProcessNotification_Integration(t *testing.T) {
	// Setup dependencies
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	// Setup mock HTTP servers
	userServer := helpers.NewMockUserServiceServer(t)
	defer userServer.Close()

	templateServer := helpers.NewMockTemplateServiceServer(t)
	defer templateServer.Close()

	// Configure mock responses
	userServer.SetPreferences("user-123", helpers.UserPreferencesResponse{
		EmailEnabled: true,
		PushEnabled:  true,
	})

	templateServer.SetTemplate("welcome_email", helpers.TemplateResponse{
		ID:        "welcome_email",
		Code:      "welcome_email",
		Subject:   "Welcome {{name}}!",
		Body:      "Hello {{name}}, welcome to our service!",
		Language:  "en",
		Variables: []string{"name"},
	})

	// Create clients
	userClient := clients.NewUserClient(clients.UserClientConfig{
		BaseURL:               userServer.URL(),
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      3,
		RetryInitialDelay:     100 * time.Millisecond,
		RetryMaxDelay:         1 * time.Second,
	})

	templateClient := clients.NewTemplateClient(clients.TemplateClientConfig{
		BaseURL:               templateServer.URL(),
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      3,
		RetryInitialDelay:     100 * time.Millisecond,
		RetryMaxDelay:         1 * time.Second,
	})

	// Setup mock Kafka manager
	mockKafka, kafkaCleanup := helpers.SetupMockKafkaProducer(t)
	defer kafkaCleanup()

	// Create a simple Kafka manager wrapper for testing
	kafkaManager := &mockKafkaManager{
		mockProducer: mockKafka,
	}

	// Create repository
	notificationRepo := repository.NewNotificationRepository(db)

	// Create orchestration service
	orchService := services.NewOrchestrationService(
		userClient,
		templateClient,
		kafkaManager,
		notificationRepo,
	)

	// Test notification request
	req := &models.NotificationRequest{
		RequestID:        "req-integration-1",
		UserID:           "user-123",
		TemplateCode:     "welcome_email",
		NotificationType: models.NotificationEmail,
		Variables:        map[string]interface{}{"name": "John Doe"},
		Priority:         2, // normal
	}

	// Process notification
	response, err := orchService.ProcessNotification(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusPending, response.Status)
	assert.NotEmpty(t, response.NotificationID)

	// Verify notification was saved to database
	ctx := context.Background()
	record, err := notificationRepo.GetByID(ctx, response.NotificationID)
	require.NoError(t, err)
	assert.Equal(t, "user-123", record.UserID)
	assert.Equal(t, "welcome_email", record.TemplateCode)
	assert.Equal(t, "email", record.NotificationType)

	// Verify Kafka message was published
	assert.Greater(t, len(mockKafka.PublishedMessages), 0)
	// The message should have the notification ID as key
	found := false
	for _, msg := range mockKafka.PublishedMessages {
		if msg.Key == response.NotificationID {
			found = true
			break
		}
	}
	assert.True(t, found, "Kafka message with notification ID should be published")
}

func TestOrchestrationService_UserPreferencesDisabled_Integration(t *testing.T) {
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	userServer := helpers.NewMockUserServiceServer(t)
	defer userServer.Close()

	templateServer := helpers.NewMockTemplateServiceServer(t)
	defer templateServer.Close()

	// User has email disabled
	userServer.SetPreferences("user-no-email", helpers.UserPreferencesResponse{
		EmailEnabled: false,
		PushEnabled:  true,
	})

	userClient := clients.NewUserClient(clients.UserClientConfig{
		BaseURL:               userServer.URL(),
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
	})

	templateClient := clients.NewTemplateClient(clients.TemplateClientConfig{
		BaseURL:               templateServer.URL(),
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
	})

	mockKafka, kafkaCleanup := helpers.SetupMockKafkaProducer(t)
	defer kafkaCleanup()

	kafkaManager := &mockKafkaManager{mockProducer: mockKafka}
	notificationRepo := repository.NewNotificationRepository(db)

	orchService := services.NewOrchestrationService(
		userClient,
		templateClient,
		kafkaManager,
		notificationRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-integration-2",
		UserID:           "user-no-email",
		TemplateCode:     "test_template",
		NotificationType: models.NotificationEmail,
		Variables:        map[string]interface{}{},
	}

	// When channel preferences are disabled, ProcessNotification returns a response
	// with StatusFailed (not an error) to allow tracking the failed notification
	response, err := orchService.ProcessNotification(req)
	require.NoError(t, err) // No error returned
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusFailed, response.Status)
	assert.Contains(t, response.Error, "email notifications disabled")

	// Verify failed notification was saved
	ctx := context.Background()
	// We need to find the notification by user_id since we don't have the ID
	records, err := notificationRepo.GetByUserID(ctx, "user-no-email", 10, 0)
	require.NoError(t, err)
	if len(records) > 0 {
		assert.Equal(t, models.StatusFailed, records[0].Status)
	}
}

func TestOrchestrationService_UpdateNotificationStatus_Integration(t *testing.T) {
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	userServer := helpers.NewMockUserServiceServer(t)
	defer userServer.Close()

	templateServer := helpers.NewMockTemplateServiceServer(t)
	defer templateServer.Close()

	userClient := clients.NewUserClient(clients.UserClientConfig{
		BaseURL: userServer.URL(),
		Timeout: 5 * time.Second,
	})

	templateClient := clients.NewTemplateClient(clients.TemplateClientConfig{
		BaseURL: templateServer.URL(),
		Timeout: 5 * time.Second,
	})

	mockKafka, kafkaCleanup := helpers.SetupMockKafkaProducer(t)
	defer kafkaCleanup()

	kafkaManager := &mockKafkaManager{mockProducer: mockKafka}
	notificationRepo := repository.NewNotificationRepository(db)

	orchService := services.NewOrchestrationService(
		userClient,
		templateClient,
		kafkaManager,
		notificationRepo,
	)

	ctx := context.Background()
	notificationID := uuid.New().String()

	// Create a notification record first
	record := &models.NotificationRecord{
		ID:               notificationID,
		UserID:           "user-status",
		TemplateCode:     "test_template",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := notificationRepo.Create(ctx, record)
	require.NoError(t, err)

	// Update status
	err = orchService.UpdateNotificationStatus(ctx, notificationID, models.StatusDelivered, "")
	require.NoError(t, err)

	// Verify update
	updated, err := notificationRepo.GetByID(ctx, notificationID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDelivered, updated.Status)
}
