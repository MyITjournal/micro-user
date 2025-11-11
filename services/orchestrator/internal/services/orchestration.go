package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/clients"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/kafka"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrchestrationService struct {
	userClient     clients.UserClient
	templateClient clients.TemplateClient
	kafkaManager   *kafka.Manager
}

func NewOrchestrationService(
	userClient clients.UserClient,
	templateClient clients.TemplateClient,
	kafkaManager *kafka.Manager,
) *OrchestrationService {
	return &OrchestrationService{
		userClient:     userClient,
		templateClient: templateClient,
		kafkaManager:   kafkaManager,
	}
}

func (s *OrchestrationService) ProcessNotification(req *models.NotificationRequest) (*models.NotificationResponse, error) {
	notificationID := uuid.New().String()
	ctx := context.Background()

	logger.Log.Info("Processing notification",
		zap.String("notification_id", notificationID),
		zap.String("user_id", req.UserID),
		zap.String("template_code", req.TemplateCode),
		zap.String("notification_type", string(req.NotificationType)),
	)

	// Step 1: Get user preferences
	userPrefs, err := s.userClient.GetPreferences(req.UserID)
	if err != nil {
		logger.Log.Error("Failed to get user preferences",
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Step 2: Check channel-specific preferences
	if err := s.validateChannelPreferences(req.NotificationType, userPrefs); err != nil {
		logger.Log.Warn("Channel validation failed",
			zap.String("user_id", req.UserID),
			zap.String("notification_type", string(req.NotificationType)),
			zap.Error(err),
		)
		return &models.NotificationResponse{
			NotificationID: notificationID,
			Status:         models.StatusFailed,
			Timestamp:      time.Now(),
			Error:          err.Error(),
		}, nil
	}

	// Step 3: Get and render template
	// Note: You'll need to update this based on your templateClient API
	// If language is no longer in preferences, you may need to remove it or set a default
	rendered, err := s.templateClient.RenderTemplate(
		req.TemplateCode,
		"en", // Default language - adjust as needed
		req.Variables,
	)
	if err != nil {
		logger.Log.Error("Failed to render template",
			zap.String("template_code", req.TemplateCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Step 4: Create notification object
	notification := &models.Notification{
		NotificationType: string(req.NotificationType),
		UserID:           req.UserID,
		TemplateCode:     req.TemplateCode,
		Variables:        req.Variables,
		RequestID:        req.ID,
		Priority:         s.getPriority(req.Priority),
		Metadata:         req.Metadata,
	}

	// Step 5: Publish to Kafka
	if err := s.publishToKafka(ctx, notification, rendered); err != nil {
		logger.Log.Error("Failed to publish to Kafka",
			zap.String("notification_id", notificationID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to queue notification: %w", err)
	}

	logger.Log.Info("Notification queued successfully",
		zap.String("notification_id", notificationID),
		zap.String("notification_type", string(req.NotificationType)),
	)

	return &models.NotificationResponse{
		NotificationID: notificationID,
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}, nil
}

// publishToKafka sends the notification to the appropriate Kafka topic
func (s *OrchestrationService) publishToKafka(ctx context.Context, notification *models.Notification, rendered interface{}) error {
	// You'll need to define what the rendered template structure looks like
	// and how to extract subject/body from it
	// This is a placeholder - adjust based on your template client response

	// Publish to the correct topic based on notification type
	return s.kafkaManager.PublishByType(
		ctx,
		notification.NotificationType,
		notification.RequestID,
		notification, // Or create a proper payload struct
	)
}

func (s *OrchestrationService) validateChannelPreferences(notificationType models.NotificationType, prefs *models.UserPreferences) error {
	switch notificationType {
	case models.NotificationEmail:
		if !prefs.Email {
			return fmt.Errorf("email notifications disabled")
		}
	case models.NotificationPush:
		if !prefs.Push {
			return fmt.Errorf("push notifications disabled")
		}
	default:
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}
	return nil
}

func (s *OrchestrationService) getPriority(priority string) string {
	if priority == "" {
		return "normal"
	}
	return priority
}
