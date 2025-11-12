package kafka

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
	defer logger.Sync()
	code := m.Run()
	os.Exit(code)
}

// MockProducer mocks the Producer for testing
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockProducer) PublishBatch(ctx context.Context, messages []Message) error {
	args := m.Called(ctx, messages)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProducer) Stats() kafka.WriterStats {
	args := m.Called()
	if args.Get(0) == nil {
		return kafka.WriterStats{}
	}
	return args.Get(0).(kafka.WriterStats)
}

func TestNewManager_Success(t *testing.T) {
	cfg := ManagerConfig{
		Brokers:    []string{"localhost:9092"},
		EmailTopic: "email.queue",
		PushTopic:  "push.queue",
		Logger:     logger.Log,
	}

	manager, err := NewManager(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.emailProducer)
	assert.NotNil(t, manager.pushProducer)
	assert.Equal(t, logger.Log, manager.logger)
}

func TestNewManager_MultipleBrokers(t *testing.T) {
	cfg := ManagerConfig{
		Brokers:    []string{"localhost:9092", "localhost:9093"},
		EmailTopic: "email.queue",
		PushTopic:  "push.queue",
		Logger:     logger.Log,
	}

	manager, err := NewManager(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestNewManager_NoBrokers(t *testing.T) {
	cfg := ManagerConfig{
		Brokers:    []string{},
		EmailTopic: "email.queue",
		PushTopic:  "push.queue",
		Logger:     logger.Log,
	}

	manager, err := NewManager(cfg)

	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "at least one broker is required")
}

func TestNewManager_EmptyBrokers(t *testing.T) {
	cfg := ManagerConfig{
		Brokers:    nil,
		EmailTopic: "email.queue",
		PushTopic:  "push.queue",
		Logger:     logger.Log,
	}

	manager, err := NewManager(cfg)

	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "at least one broker is required")
}

func TestManager_PublishEmail_Success(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
		"user_id":         "user-456",
	}

	mockEmailProducer.On("Publish", context.Background(), "notif-123", payload).Return(nil)

	err := manager.PublishEmail(context.Background(), "notif-123", payload)

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertNotCalled(t, "Publish")
}

func TestManager_PublishEmail_Error(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	expectedError := errors.New("kafka publish failed")
	mockEmailProducer.On("Publish", context.Background(), "notif-123", payload).Return(expectedError)

	err := manager.PublishEmail(context.Background(), "notif-123", payload)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockEmailProducer.AssertExpectations(t)
}

func TestManager_PublishPush_Success(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
		"user_id":         "user-456",
	}

	mockPushProducer.On("Publish", context.Background(), "notif-123", payload).Return(nil)

	err := manager.PublishPush(context.Background(), "notif-123", payload)

	assert.NoError(t, err)
	mockPushProducer.AssertExpectations(t)
	mockEmailProducer.AssertNotCalled(t, "Publish")
}

func TestManager_PublishPush_Error(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	expectedError := errors.New("kafka publish failed")
	mockPushProducer.On("Publish", context.Background(), "notif-123", payload).Return(expectedError)

	err := manager.PublishPush(context.Background(), "notif-123", payload)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_PublishByType_Email(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	mockEmailProducer.On("Publish", context.Background(), "notif-123", payload).Return(nil)

	err := manager.PublishByType(context.Background(), "email", "notif-123", payload)

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertNotCalled(t, "Publish")
}

func TestManager_PublishByType_Push(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	mockPushProducer.On("Publish", context.Background(), "notif-123", payload).Return(nil)

	err := manager.PublishByType(context.Background(), "push", "notif-123", payload)

	assert.NoError(t, err)
	mockPushProducer.AssertExpectations(t)
	mockEmailProducer.AssertNotCalled(t, "Publish")
}

func TestManager_PublishByType_UnsupportedType(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	err := manager.PublishByType(context.Background(), "sms", "notif-123", payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification type")
	mockEmailProducer.AssertNotCalled(t, "Publish")
	mockPushProducer.AssertNotCalled(t, "Publish")
}

func TestManager_PublishByType_EmptyType(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	err := manager.PublishByType(context.Background(), "", "notif-123", payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification type")
}

func TestManager_Close_Success(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	mockEmailProducer.On("Close").Return(nil)
	mockPushProducer.On("Close").Return(nil)

	err := manager.Close()

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_Close_EmailProducerError(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	expectedError := errors.New("email producer close failed")
	mockEmailProducer.On("Close").Return(expectedError)
	// Push producer should still be called even if email fails
	mockPushProducer.On("Close").Return(nil)

	err := manager.Close()

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_Close_PushProducerError(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	mockEmailProducer.On("Close").Return(nil)
	expectedError := errors.New("push producer close failed")
	mockPushProducer.On("Close").Return(expectedError)

	err := manager.Close()

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_Close_BothProducersError(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	emailError := errors.New("email producer close failed")
	pushError := errors.New("push producer close failed")
	mockEmailProducer.On("Close").Return(emailError)
	mockPushProducer.On("Close").Return(pushError)

	err := manager.Close()

	assert.Error(t, err)
	// Should return the first error (email producer error)
	assert.Equal(t, emailError, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_HealthCheck_Success(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	emailStats := kafka.WriterStats{
		Errors: 0,
		Writes: 10,
	}
	pushStats := kafka.WriterStats{
		Errors: 0,
		Writes: 5,
	}

	mockEmailProducer.On("Stats").Return(emailStats)
	mockPushProducer.On("Stats").Return(pushStats)

	err := manager.HealthCheck()

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_HealthCheck_EmailProducerHasErrors(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	emailStats := kafka.WriterStats{
		Errors: 5,
		Writes: 10,
	}
	pushStats := kafka.WriterStats{
		Errors: 0,
		Writes: 5,
	}

	mockEmailProducer.On("Stats").Return(emailStats)
	mockPushProducer.On("Stats").Return(pushStats)

	// HealthCheck should not return error even if stats show errors
	// It only logs warnings
	err := manager.HealthCheck()

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_HealthCheck_PushProducerHasErrors(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	emailStats := kafka.WriterStats{
		Errors: 0,
		Writes: 10,
	}
	pushStats := kafka.WriterStats{
		Errors: 3,
		Writes: 5,
	}

	mockEmailProducer.On("Stats").Return(emailStats)
	mockPushProducer.On("Stats").Return(pushStats)

	err := manager.HealthCheck()

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_HealthCheck_BothProducersHaveErrors(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	emailStats := kafka.WriterStats{
		Errors: 5,
		Writes: 10,
	}
	pushStats := kafka.WriterStats{
		Errors: 3,
		Writes: 5,
	}

	mockEmailProducer.On("Stats").Return(emailStats)
	mockPushProducer.On("Stats").Return(pushStats)

	err := manager.HealthCheck()

	assert.NoError(t, err)
	mockEmailProducer.AssertExpectations(t)
	mockPushProducer.AssertExpectations(t)
}

func TestManager_PublishEmail_ContextCancellation(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	ctxError := context.Canceled
	mockEmailProducer.On("Publish", ctx, "notif-123", payload).Return(ctxError)

	err := manager.PublishEmail(ctx, "notif-123", payload)

	assert.Error(t, err)
	assert.Equal(t, ctxError, err)
	mockEmailProducer.AssertExpectations(t)
}

func TestManager_PublishByType_CaseSensitive(t *testing.T) {
	mockEmailProducer := new(MockProducer)
	mockPushProducer := new(MockProducer)

	manager := &Manager{
		emailProducer: mockEmailProducer,
		pushProducer:  mockPushProducer,
		logger:        logger.Log,
	}

	payload := map[string]interface{}{
		"notification_id": "notif-123",
	}

	// Test that type matching is case-sensitive
	err := manager.PublishByType(context.Background(), "Email", "notif-123", payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification type")
	mockEmailProducer.AssertNotCalled(t, "Publish")
	mockPushProducer.AssertNotCalled(t, "Publish")
}
