package fixtures

import (
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/google/uuid"
)

// GetTestNotificationRecord returns a test notification record
func GetTestNotificationRecord() *models.NotificationRecord {
	now := time.Now()
	return &models.NotificationRecord{
		ID:               uuid.New().String(),
		UserID:           "test-user-123",
		TemplateCode:     "test_template",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "Test User"},
		Metadata:         &models.JSONB{"source": "test"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// GetTestNotificationRequest returns a test notification request
func GetTestNotificationRequest() *models.NotificationRequest {
	return &models.NotificationRequest{
		RequestID:        uuid.New().String(),
		UserID:           "test-user-123",
		TemplateCode:     "test_template",
		NotificationType: models.NotificationEmail,
		Variables:        map[string]interface{}{"name": "Test User"},
		Priority:         2, // normal
	}
}

// GetTestNotificationResponse returns a test notification response
func GetTestNotificationResponse() *models.NotificationResponse {
	return &models.NotificationResponse{
		NotificationID: uuid.New().String(),
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}
}
