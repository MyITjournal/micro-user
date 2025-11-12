package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/clients"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/handlers"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/repository"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/services"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationHandler_Create_Integration(t *testing.T) {
	// Setup dependencies
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	redisClient, redisCleanup := helpers.SetupTestRedis(t)
	defer redisCleanup()

	userServer := helpers.NewMockUserServiceServer(t)
	defer userServer.Close()

	templateServer := helpers.NewMockTemplateServiceServer(t)
	defer templateServer.Close()

	// Configure mocks
	userServer.SetPreferences("user-handler-1", helpers.UserPreferencesResponse{
		EmailEnabled: true,
		PushEnabled:  true,
	})

	templateServer.SetTemplate("welcome_email", helpers.TemplateResponse{
		ID:        "welcome_email",
		Code:      "welcome_email",
		Subject:   "Welcome {{name}}!",
		Body:      "Hello {{name}}!",
		Language:  "en",
		Variables: []string{"name"},
	})

	// Create services
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

	idempotencyService := services.NewIdempotencyService(redisClient, 1*time.Hour)

	// Create handler
	handler := handlers.NewNotificationHandler(orchService, idempotencyService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.POST("/api/v1/notifications", handler.Create)

	// Create request
	reqBody := models.NotificationRequest{
		RequestID:        "req-handler-1",
		UserID:           "user-handler-1",
		TemplateCode:     "welcome_email",
		NotificationType: models.NotificationEmail,
		Variables:        map[string]interface{}{"name": "Test User"},
		Priority:         2,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Verify idempotency header
	assert.Equal(t, "false", w.Header().Get("X-Idempotent-Replay"))
}

func TestNotificationHandler_Create_Idempotency_Integration(t *testing.T) {
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	redisClient, redisCleanup := helpers.SetupTestRedis(t)
	defer redisCleanup()

	userServer := helpers.NewMockUserServiceServer(t)
	defer userServer.Close()

	templateServer := helpers.NewMockTemplateServiceServer(t)
	defer templateServer.Close()

	userServer.SetPreferences("user-idempotent", helpers.UserPreferencesResponse{
		EmailEnabled: true,
		PushEnabled:  true,
	})

	// Set up template for the test
	templateServer.SetTemplate("test_template", helpers.TemplateResponse{
		ID:        "test_template",
		Code:      "test_template",
		Subject:   "Test Subject",
		Body:      "Test Body",
		Language:  "en",
		Variables: []string{},
	})

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

	idempotencyService := services.NewIdempotencyService(redisClient, 1*time.Hour)

	handler := handlers.NewNotificationHandler(orchService, idempotencyService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.POST("/api/v1/notifications", handler.Create)

	reqBody := models.NotificationRequest{
		RequestID:        "req-idempotent-1", // Same request ID
		UserID:           "user-idempotent",
		TemplateCode:     "test_template",
		NotificationType: models.NotificationEmail,
		Variables:        map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)

	// First request
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Second request with same request_id - should return cached response
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code) // Should be 200 for cached response
	assert.Equal(t, "true", w2.Header().Get("X-Idempotent-Replay"))

	var response2 models.Response
	err := json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)
	assert.True(t, response2.Success)

	// Verify only one Kafka message was published (first request)
	assert.Equal(t, 1, len(mockKafka.PublishedMessages))
}

func TestNotificationHandler_Create_InvalidRequest_Integration(t *testing.T) {
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	redisClient, redisCleanup := helpers.SetupTestRedis(t)
	defer redisCleanup()

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

	idempotencyService := services.NewIdempotencyService(redisClient, 1*time.Hour)
	handler := handlers.NewNotificationHandler(orchService, idempotencyService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.POST("/api/v1/notifications", handler.Create)

	// Invalid request - missing required fields
	invalidBody := `{"user_id": "user-123"}` // Missing request_id, template_code, etc.
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewBufferString(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	// Error message contains field name "RequestID" (struct field) or "request_id" (JSON tag)
	assert.True(t,
		strings.Contains(response.Error, "request_id") ||
			strings.Contains(response.Error, "RequestID"),
		"Error should contain 'request_id' or 'RequestID', got: %s", response.Error)
}

func TestNotificationHandler_UpdateStatus_Integration(t *testing.T) {
	db, dbCleanup := helpers.SetupTestDB(t)
	defer dbCleanup()

	redisClient, redisCleanup := helpers.SetupTestRedis(t)
	defer redisCleanup()

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

	idempotencyService := services.NewIdempotencyService(redisClient, 1*time.Hour)
	handler := handlers.NewNotificationHandler(orchService, idempotencyService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.PUT("/api/v1/email/status", handler.UpdateStatus)

	// Create a notification first
	notificationID := uuid.New().String()
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
	ctx := context.Background()
	err := notificationRepo.Create(ctx, record)
	require.NoError(t, err)

	// Update status
	statusUpdate := map[string]interface{}{
		"notification_id": notificationID,
		"status":          "delivered",
	}

	body, _ := json.Marshal(statusUpdate)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/email/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify status was updated in database
	updated, err := notificationRepo.GetByID(ctx, notificationID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDelivered, updated.Status)
}
