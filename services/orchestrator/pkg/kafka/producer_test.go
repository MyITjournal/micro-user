package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWriter is a mock implementation of kafkaWriter interface for testing
type mockWriter struct {
	writeMessagesFunc func(ctx context.Context, msgs ...kafka.Message) error
	closeFunc         func() error
	statsFunc         func() kafka.WriterStats
	topic             string
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.writeMessagesFunc != nil {
		return m.writeMessagesFunc(ctx, msgs...)
	}
	return nil
}

func (m *mockWriter) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockWriter) Stats() kafka.WriterStats {
	if m.statsFunc != nil {
		return m.statsFunc()
	}
	return kafka.WriterStats{}
}

func TestNewProducer(t *testing.T) {
	cfg := ProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		Logger:  logger.Log,
	}

	producer := NewProducer(cfg)

	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
	assert.Equal(t, logger.Log, producer.logger)
	assert.Equal(t, "test-topic", producer.topic)
}

func TestNewProducer_MultipleBrokers(t *testing.T) {
	cfg := ProducerConfig{
		Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:   "test-topic",
		Logger:  logger.Log,
	}

	producer := NewProducer(cfg)

	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
}

func TestProducer_Publish_Success(t *testing.T) {
	// Create a producer with a mock writer
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 1)
				assert.Equal(t, "test-key", string(msgs[0].Key))

				var payload map[string]interface{}
				err := json.Unmarshal(msgs[0].Value, &payload)
				assert.NoError(t, err)
				assert.Equal(t, "test-value", payload["key"])

				return nil
			},
		},
		logger: logger.Log,
	}

	payload := map[string]interface{}{
		"key": "test-value",
	}

	err := producer.Publish(context.Background(), "test-key", payload)
	assert.NoError(t, err)
}

func TestProducer_Publish_MarshalError(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{},
		logger: logger.Log,
	}

	// Create a value that cannot be marshaled (circular reference)
	type Circular struct {
		Self *Circular
	}
	circular := &Circular{}
	circular.Self = circular

	err := producer.Publish(context.Background(), "test-key", circular)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal message")
}

func TestProducer_Publish_WriteError(t *testing.T) {
	expectedError := errors.New("kafka write failed")
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				return expectedError
			},
		},
		logger: logger.Log,
	}

	payload := map[string]interface{}{
		"key": "test-value",
	}

	err := producer.Publish(context.Background(), "test-key", payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish message")
}

func TestProducer_Publish_ContextCancellation(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				// Simulate context cancellation
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			},
		},
		logger: logger.Log,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	payload := map[string]interface{}{
		"key": "test-value",
	}

	err := producer.Publish(ctx, "test-key", payload)
	assert.Error(t, err)
}

func TestProducer_Publish_ComplexPayload(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 1)

				var payload map[string]interface{}
				err := json.Unmarshal(msgs[0].Value, &payload)
				require.NoError(t, err)

				assert.Equal(t, "test", payload["string"])
				assert.Equal(t, float64(123), payload["number"])
				assert.Equal(t, true, payload["bool"])
				assert.NotNil(t, payload["nested"])

				return nil
			},
		},
		logger: logger.Log,
	}

	payload := map[string]interface{}{
		"string": "test",
		"number": 123,
		"bool":   true,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	err := producer.Publish(context.Background(), "test-key", payload)
	assert.NoError(t, err)
}

func TestProducer_PublishBatch_Success(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 3)
				assert.Equal(t, "key-1", string(msgs[0].Key))
				assert.Equal(t, "key-2", string(msgs[1].Key))
				assert.Equal(t, "key-3", string(msgs[2].Key))
				return nil
			},
		},
		logger: logger.Log,
	}

	messages := []Message{
		{Key: "key-1", Value: map[string]interface{}{"msg": "message 1"}},
		{Key: "key-2", Value: map[string]interface{}{"msg": "message 2"}},
		{Key: "key-3", Value: map[string]interface{}{"msg": "message 3"}},
	}

	err := producer.PublishBatch(context.Background(), messages)
	assert.NoError(t, err)
}

func TestProducer_PublishBatch_EmptySlice(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 0)
				return nil
			},
		},
		logger: logger.Log,
	}

	err := producer.PublishBatch(context.Background(), []Message{})
	assert.NoError(t, err)
}

func TestProducer_PublishBatch_MarshalError(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{},
		logger: logger.Log,
	}

	type Circular struct {
		Self *Circular
	}
	circular := &Circular{}
	circular.Self = circular

	messages := []Message{
		{Key: "key-1", Value: map[string]interface{}{"msg": "message 1"}},
		{Key: "key-2", Value: circular}, // This will fail to marshal
	}

	err := producer.PublishBatch(context.Background(), messages)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal batch message")
}

func TestProducer_PublishBatch_WriteError(t *testing.T) {
	expectedError := errors.New("batch write failed")
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				return expectedError
			},
		},
		logger: logger.Log,
	}

	messages := []Message{
		{Key: "key-1", Value: map[string]interface{}{"msg": "message 1"}},
		{Key: "key-2", Value: map[string]interface{}{"msg": "message 2"}},
	}

	err := producer.PublishBatch(context.Background(), messages)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish batch")
}

func TestProducer_Close_Success(t *testing.T) {
	closed := false
	producer := &Producer{
		writer: &mockWriter{
			closeFunc: func() error {
				closed = true
				return nil
			},
		},
		logger: logger.Log,
	}

	err := producer.Close()
	assert.NoError(t, err)
	assert.True(t, closed)
}

func TestProducer_Close_Error(t *testing.T) {
	expectedError := errors.New("close failed")
	producer := &Producer{
		writer: &mockWriter{
			closeFunc: func() error {
				return expectedError
			},
		},
		logger: logger.Log,
	}

	err := producer.Close()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestProducer_Stats(t *testing.T) {
	expectedStats := kafka.WriterStats{
		Writes:   10,
		Messages: 25,
		Bytes:    1024,
		Errors:   0,
	}

	producer := &Producer{
		writer: &mockWriter{
			statsFunc: func() kafka.WriterStats {
				return expectedStats
			},
		},
		logger: logger.Log,
	}

	stats := producer.Stats()
	assert.Equal(t, expectedStats, stats)
}

func TestProducer_Publish_WithNilLogger(t *testing.T) {
	// Test that producer handles nil logger gracefully
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				return nil
			},
		},
		logger: nil,
	}

	payload := map[string]interface{}{
		"key": "test-value",
	}

	// Should not panic
	err := producer.Publish(context.Background(), "test-key", payload)
	assert.NoError(t, err)
}

func TestProducer_Publish_EmptyKey(t *testing.T) {
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 1)
				assert.Equal(t, "", string(msgs[0].Key))
				return nil
			},
		},
		logger: logger.Log,
	}

	payload := map[string]interface{}{
		"key": "test-value",
	}

	err := producer.Publish(context.Background(), "", payload)
	assert.NoError(t, err)
}

func TestProducer_Publish_TimeStamp(t *testing.T) {
	var messageTime time.Time
	producer := &Producer{
		writer: &mockWriter{
			writeMessagesFunc: func(ctx context.Context, msgs ...kafka.Message) error {
				assert.Len(t, msgs, 1)
				messageTime = msgs[0].Time
				return nil
			},
		},
		logger: logger.Log,
	}

	payload := map[string]interface{}{
		"key": "test-value",
	}

	err := producer.Publish(context.Background(), "test-key", payload)
	assert.NoError(t, err)

	// Verify timestamp is set (should be very recent)
	assert.False(t, messageTime.IsZero())
	assert.WithinDuration(t, time.Now(), messageTime, 1*time.Second)
}
