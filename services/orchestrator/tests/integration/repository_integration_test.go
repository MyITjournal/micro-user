package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/repository"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/tests/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize logger
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Skip tests if TEST_DB_HOST is not set (integration tests require database)
	if os.Getenv("TEST_DB_HOST") == "" && os.Getenv("CI") == "" {
		os.Exit(0) // Skip tests in local development without test DB
	}

	code := m.Run()
	os.Exit(code)
}

func TestNotificationRepository_Create_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	notification := &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "user-123",
		TemplateCode:     "welcome_email",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "John Doe"},
		Metadata:         &models.JSONB{"source": "api"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := repo.GetByID(context.Background(), notification.ID)
	require.NoError(t, err)
	assert.Equal(t, notification.ID, retrieved.ID)
	assert.Equal(t, notification.UserID, retrieved.UserID)
	assert.Equal(t, notification.TemplateCode, retrieved.TemplateCode)
	assert.Equal(t, notification.Status, retrieved.Status)
}

func TestNotificationRepository_UpdateStatus_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	// Create a notification
	notification := &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "user-456",
		TemplateCode:     "test_template",
		NotificationType: "push",
		Status:           models.StatusPending,
		Priority:         "high",
		Variables:        models.JSONB{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	// Update status
	err = repo.UpdateStatus(context.Background(), notification.ID, models.StatusDelivered, "")
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), notification.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDelivered, retrieved.Status)
	assert.NotEqual(t, notification.UpdatedAt, retrieved.UpdatedAt) // Updated timestamp should change
}

func TestNotificationRepository_UpdateStatus_WithError_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	notification := &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "user-789",
		TemplateCode:     "test_template",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	errorMsg := "Failed to send email: connection timeout"
	err = repo.UpdateStatus(context.Background(), notification.ID, models.StatusFailed, errorMsg)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), notification.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusFailed, retrieved.Status)
	assert.NotNil(t, retrieved.ErrorMessage)
	assert.Equal(t, errorMsg, *retrieved.ErrorMessage)
}

func TestNotificationRepository_GetByUserID_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	userID := "user-multi"
	ctx := context.Background()

	// Create multiple notifications for the same user
	notifications := []*models.NotificationRecord{
		{
			ID:               uuid.New().String(),
			UserID:           userID,
			TemplateCode:     "template1",
			NotificationType: "email",
			Status:           models.StatusPending,
			Priority:         "normal",
			Variables:        models.JSONB{},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               uuid.New().String(),
			UserID:           userID,
			TemplateCode:     "template2",
			NotificationType: "push",
			Status:           models.StatusDelivered,
			Priority:         "high",
			Variables:        models.JSONB{},
			CreatedAt:        time.Now().Add(1 * time.Minute),
			UpdatedAt:        time.Now().Add(1 * time.Minute),
		},
		{
			ID:               uuid.New().String(),
			UserID:           userID,
			TemplateCode:     "template3",
			NotificationType: "email",
			Status:           models.StatusFailed,
			Priority:         "normal",
			Variables:        models.JSONB{},
			CreatedAt:        time.Now().Add(2 * time.Minute),
			UpdatedAt:        time.Now().Add(2 * time.Minute),
		},
	}

	for _, notif := range notifications {
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// Retrieve all notifications for the user
	retrieved, err := repo.GetByUserID(ctx, userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Verify they are ordered by created_at DESC
	assert.True(t, retrieved[0].CreatedAt.After(retrieved[1].CreatedAt) || retrieved[0].CreatedAt.Equal(retrieved[1].CreatedAt))
	assert.True(t, retrieved[1].CreatedAt.After(retrieved[2].CreatedAt) || retrieved[1].CreatedAt.Equal(retrieved[2].CreatedAt))
}

func TestNotificationRepository_GetByUserID_Pagination_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	userID := "user-pagination"
	ctx := context.Background()

	// Create 5 notifications
	for i := 0; i < 5; i++ {
		notification := &models.NotificationRecord{
			ID:               uuid.New().String(),
			UserID:           userID,
			TemplateCode:     "template",
			NotificationType: "email",
			Status:           models.StatusPending,
			Priority:         "normal",
			Variables:        models.JSONB{},
			CreatedAt:        time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt:        time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	// Test pagination: first page
	page1, err := repo.GetByUserID(ctx, userID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Test pagination: second page
	page2, err := repo.GetByUserID(ctx, userID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	// Test pagination: third page
	page3, err := repo.GetByUserID(ctx, userID, 2, 4)
	require.NoError(t, err)
	assert.Len(t, page3, 1) // Only 1 remaining
}

func TestNotificationRepository_GetByUserID_EmptyResult_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	retrieved, err := repo.GetByUserID(context.Background(), "non-existent-user", 10, 0)
	require.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

func TestNotificationRepository_ComplexMetadata_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	metadata := models.JSONB{
		"source":      "api",
		"campaign_id": "campaign-123",
		"tags":        []interface{}{"welcome", "onboarding"},
		"metadata": map[string]interface{}{
			"nested": "value",
		},
	}

	notification := &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "user-metadata",
		TemplateCode:     "test_template",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "Test User"},
		Metadata:         &metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), notification.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.Metadata)
	assert.Equal(t, "api", (*retrieved.Metadata)["source"])
	assert.Equal(t, "campaign-123", (*retrieved.Metadata)["campaign_id"])
}

func TestNotificationRepository_ScheduledFor_Integration(t *testing.T) {
	db, cleanup := helpers.SetupTestDB(t)
	defer cleanup()

	repo := repository.NewNotificationRepository(db)

	// Create time in UTC to match PostgreSQL's storage (PostgreSQL stores timestamps in UTC)
	scheduledTime := time.Now().UTC().Add(24 * time.Hour)
	notification := &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "user-scheduled",
		TemplateCode:     "scheduled_template",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{},
		ScheduledFor:     &scheduledTime,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), notification.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.ScheduledFor)
	// Both times are in UTC, so compare directly
	assert.WithinDuration(t, scheduledTime, *retrieved.ScheduledFor, time.Second)
}
