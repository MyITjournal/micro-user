package integration

import (
	"context"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/tests/helpers"
)

// mockKafkaManager implements KafkaManagerInterface for testing
type mockKafkaManager struct {
	mockProducer *helpers.MockKafkaProducer
}

func (m *mockKafkaManager) PublishByType(ctx context.Context, notificationType, notificationID string, payload interface{}) error {
	key := notificationID
	return m.mockProducer.Publish(ctx, key, payload)
}
