package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/kafka"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// TestKafkaConfig returns test Kafka configuration
func TestKafkaConfig() kafka.ProducerConfig {
	brokers := os.Getenv("TEST_KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	// Initialize logger if not already done
	if logger.Log == nil {
		if err := logger.Initialize("info", "console"); err != nil {
			// Use a no-op logger for tests
			logger.Log = zap.NewNop()
		}
	}

	return kafka.ProducerConfig{
		Brokers: []string{brokers},
		Topic:   "test_topic",
		Logger:  logger.Log,
	}
}

// SetupTestKafkaProducer creates a test Kafka producer
// Note: For integration tests, you may want to use a mock producer instead
func SetupTestKafkaProducer(t *testing.T, topic string) (*kafka.Producer, func()) {
	cfg := TestKafkaConfig()
	cfg.Topic = topic

	producer := kafka.NewProducer(cfg)

	// Test connection by checking if we can create the producer
	// In a real scenario, you might want to verify connectivity
	// For now, we'll just return the producer

	cleanup := func() {
		if err := producer.Close(); err != nil {
			t.Logf("Warning: Failed to close Kafka producer: %v", err)
		}
	}

	return producer, cleanup
}

// MockKafkaProducer is a mock implementation for testing without real Kafka
type MockKafkaProducer struct {
	PublishedMessages []kafka.Message
	PublishError      error
	CloseError        error
}

// Publish simulates publishing a message
func (m *MockKafkaProducer) Publish(ctx context.Context, key string, value interface{}) error {
	if m.PublishError != nil {
		return m.PublishError
	}

	m.PublishedMessages = append(m.PublishedMessages, kafka.Message{
		Key:   key,
		Value: value,
	})
	return nil
}

// PublishBatch simulates publishing multiple messages
func (m *MockKafkaProducer) PublishBatch(ctx context.Context, messages []kafka.Message) error {
	if m.PublishError != nil {
		return m.PublishError
	}

	m.PublishedMessages = append(m.PublishedMessages, messages...)
	return nil
}

// Close simulates closing the producer
func (m *MockKafkaProducer) Close() error {
	return m.CloseError
}

// Stats returns empty stats for mock
func (m *MockKafkaProducer) Stats() kafkago.WriterStats {
	return kafkago.WriterStats{}
}

// SetupMockKafkaProducer creates a mock Kafka producer for testing
func SetupMockKafkaProducer(t *testing.T) (*MockKafkaProducer, func()) {
	mock := &MockKafkaProducer{
		PublishedMessages: make([]kafka.Message, 0),
	}

	cleanup := func() {
		mock.PublishedMessages = nil
	}

	return mock, cleanup
}

// VerifyKafkaMessage checks if a message was published with expected content
func VerifyKafkaMessage(mock *MockKafkaProducer, expectedKey string, expectedValue interface{}) error {
	for _, msg := range mock.PublishedMessages {
		if msg.Key == expectedKey {
			// Simple comparison - in real tests you might want deeper comparison
			if fmt.Sprintf("%v", msg.Value) == fmt.Sprintf("%v", expectedValue) {
				return nil
			}
		}
	}
	return fmt.Errorf("message with key %s and value %v not found", expectedKey, expectedValue)
}

// WaitForKafkaMessage waits for a message to be published (useful for async testing)
func WaitForKafkaMessage(mock *MockKafkaProducer, expectedKey string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, msg := range mock.PublishedMessages {
			if msg.Key == expectedKey {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for message with key %s", expectedKey)
}
