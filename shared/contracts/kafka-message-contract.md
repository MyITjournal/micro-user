# Kafka Message Contract

This document defines the message format that the Orchestrator Service publishes to Kafka queues for consumption by the Email and Push services.

## Topics

The orchestrator publishes messages to the following topics:

- **`email.queue`** - Messages for email notifications
- **`push.queue`** - Messages for push notifications
- **`failed.queue`** - Dead letter queue for failed messages (future use)

## Message Format

All messages are published as **JSON** with the following structure:

```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "notification_type": "email",
  "user_id": "usr_123",
  "template_code": "welcome_email",
  "subject": "Welcome to Our Service!",
  "body": "<h1>Welcome John!</h1><p>Thank you for joining.</p>",
  "text_body": "Welcome John!\n\nThank you for joining.",
  "priority": "normal",
  "metadata": {
    "campaign_id": "campaign-123",
    "source": "api"
  },
  "created_at": "2025-11-12T16:30:00Z",
  "retry_count": 0
}
```

## Field Specifications

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `notification_id` | string (UUID) | Unique identifier for the notification |
| `notification_type` | string | Either `"email"` or `"push"` |
| `user_id` | string | ID of the user receiving the notification |
| `template_code` | string | Code identifying the template used |
| `body` | string | Main content of the notification |
| `priority` | string | Priority level: `"low"`, `"normal"`, `"high"`, or `"urgent"` |
| `created_at` | string (ISO 8601) | Timestamp when the notification was created |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `subject` | string | Subject line (for email) or title (for push) |
| `text_body` | string | Plain text version of the body (for email) |
| `metadata` | object | Additional metadata (campaign info, source, etc.) |
| `retry_count` | integer | Number of retry attempts (default: 0) |
| `last_retry_at` | string (ISO 8601) | Timestamp of last retry attempt |

## Message Key

Messages are published with the **notification_id** as the key. This ensures:
- Messages with the same notification_id go to the same partition
- Enables idempotent processing
- Allows for message ordering per notification

## Email Queue Messages

For `email.queue` topic, messages include:

```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "notification_type": "email",
  "user_id": "usr_123",
  "template_code": "welcome_email",
  "subject": "Welcome to Our Service!",
  "body": "<h1>Welcome John!</h1><p>Thank you for joining.</p>",
  "text_body": "Welcome John!\n\nThank you for joining.",
  "priority": "normal",
  "created_at": "2025-11-12T16:30:00Z"
}
```

**Email Service Responsibilities:**
- Extract `user_id` and fetch user's email address from User Service
- Use `subject` for email subject line
- Use `body` for HTML email body
- Use `text_body` for plain text alternative
- Send email via SMTP or email service API
- Update notification status via orchestrator's status endpoint

## Push Queue Messages

For `push.queue` topic, messages include:

```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "notification_type": "push",
  "user_id": "usr_123",
  "template_code": "welcome_push",
  "subject": "Welcome!",
  "body": "Thank you for joining our service!",
  "priority": "normal",
  "created_at": "2025-11-12T16:30:00Z"
}
```

**Push Service Responsibilities:**
- Extract `user_id` and fetch user's device tokens from User Service
- Use `subject` as push notification title
- Use `body` as push notification message
- Send push notification via FCM, OneSignal, or other push service
- Update notification status via orchestrator's status endpoint

## Priority Levels

Priority is mapped from integer input (0-4) to string:

- `0` or unset → `"normal"`
- `1` → `"low"`
- `2` → `"normal"`
- `3` → `"high"`
- `4` → `"urgent"`

## Status Updates

After processing a message, services **MUST** update the notification status by calling:

```
PUT /api/v1/{notification_type}/status
```

Request body:
```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "delivered",
  "timestamp": "2025-11-12T16:30:05Z",
  "error": null
}
```

Status values:
- `"pending"` - Initial state (set by orchestrator)
- `"delivered"` - Successfully sent
- `"failed"` - Failed to send (include error message)

## Error Handling

### Retry Logic

If a message fails to process:
1. Service should retry with exponential backoff
2. Increment `retry_count` in metadata
3. After max retries, move to `failed.queue` (future implementation)

### Dead Letter Queue

Permanently failed messages will be published to `failed.queue` with additional error information:
```json
{
  "original_message": { /* original KafkaNotificationPayload */ },
  "error": "Error message",
  "error_type": "processing_error",
  "failed_at": "2025-11-12T16:30:00Z",
  "retry_count": 3,
  "source_topic": "email.queue"
}
```

## Testing Locally

See `KAFKA_TESTING_GUIDE.md` for instructions on:
- Starting Kafka locally
- Consuming messages for testing
- Sample consumer scripts
- Debugging tips

## Example Messages

### Email Notification Example

```json
{
  "notification_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "notification_type": "email",
  "user_id": "usr_abc123",
  "template_code": "password_reset",
  "subject": "Reset Your Password",
  "body": "<h1>Password Reset</h1><p>Click <a href=\"https://example.com/reset?token=xyz\">here</a> to reset your password.</p>",
  "text_body": "Password Reset\n\nClick here to reset your password: https://example.com/reset?token=xyz",
  "priority": "high",
  "metadata": {
    "campaign_id": "reset-2025-11",
    "source": "web"
  },
  "created_at": "2025-11-12T16:30:00Z",
  "retry_count": 0
}
```

### Push Notification Example

```json
{
  "notification_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "notification_type": "push",
  "user_id": "usr_xyz789",
  "template_code": "order_confirmation",
  "subject": "Order Confirmed",
  "body": "Your order #12345 has been confirmed and will arrive soon!",
  "priority": "normal",
  "metadata": {
    "order_id": "order-12345",
    "source": "mobile_app"
  },
  "created_at": "2025-11-12T16:30:00Z",
  "retry_count": 0
}
```

## Consumer Requirements

Your service should:

1. **Consume from the correct topic:**
   - Email Service → `email.queue`
   - Push Service → `push.queue`

2. **Handle message deserialization:**
   - Parse JSON payload
   - Validate required fields
   - Handle missing optional fields gracefully

3. **Process messages:**
   - Fetch user data (email/device tokens) from User Service
   - Send notification via appropriate channel
   - Update status via orchestrator API

4. **Handle errors:**
   - Retry transient failures
   - Log errors appropriately
   - Update status with error message

5. **Ensure idempotency:**
   - Check if notification was already processed (using `notification_id`)
   - Handle duplicate messages gracefully

## Integration Checklist

- [ ] Service can connect to Kafka broker at `localhost:9092`
- [ ] Service can consume from `email.queue` or `push.queue`
- [ ] Service can parse JSON message format
- [ ] Service can fetch user data from User Service
- [ ] Service can send notifications (email/push)
- [ ] Service can update status via orchestrator API
- [ ] Service handles errors and retries appropriately
- [ ] Service implements idempotency checks

