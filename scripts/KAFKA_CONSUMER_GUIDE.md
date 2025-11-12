# Kafka Consumer Guide for Email and Push Services

This guide provides scripts and instructions for testing your Email and Push services by consuming messages from Kafka queues.

## Quick Start

### 1. Start Kafka

```bash
# From project root
docker-compose -f docker-compose.kafka.yml up -d

# Wait for Kafka to be ready
docker-compose -f docker-compose.kafka.yml logs -f kafka
```

### 2. Verify Topics

```bash
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092
```

Expected topics:
- `email.queue`
- `push.queue`
- `failed.queue`

### 3. Generate Test Messages

Create notifications via the orchestrator API (see examples below).

### 4. Consume Messages

Choose one of the methods below.

## Consumption Methods

### Method 1: Kafka Console Consumer (No Dependencies)

**Consume Email Queue:**
```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

**Consume Push Queue:**
```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic push.queue \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

**Consume Only New Messages:**
Remove `--from-beginning` flag.

### Method 2: Python Consumer Script

**Prerequisites:**
```bash
pip install kafka-python
```

**Usage:**
```bash
# Consume email queue
python3 scripts/test-kafka-consumer.py email

# Consume push queue from beginning
python3 scripts/test-kafka-consumer.py push --from-beginning

# Custom bootstrap server
python3 scripts/test-kafka-consumer.py email --bootstrap localhost:9093
```

### Method 3: Node.js Consumer Script

**Prerequisites:**
```bash
npm install kafkajs
```

**Usage:**
```bash
# Consume email queue
node scripts/test-kafka-consumer.js email

# Consume push queue from beginning
node scripts/test-kafka-consumer.js push --from-beginning
```

### Method 4: Go Consumer Script

**Prerequisites:**
```bash
go get github.com/segmentio/kafka-go
```

**Usage:**
```bash
# Build
go build -o test-kafka-consumer scripts/test-kafka-consumer.go

# Run
./test-kafka-consumer -topic email
./test-kafka-consumer -topic push -from-beginning
```

## Generating Test Messages

### Via API Gateway

```bash
# Email notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-Request-ID: test-$(date +%s)" \
  -d '{
    "request_id": "test-email-123",
    "user_id": "test-user",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"name": "Test User"},
    "priority": 2
  }'

# Push notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-Request-ID: test-$(date +%s)" \
  -d '{
    "request_id": "test-push-123",
    "user_id": "test-user",
    "template_code": "welcome_push",
    "notification_type": "push",
    "variables": {"name": "Test User"},
    "priority": 2
  }'
```

### Via Direct Orchestrator

```bash
curl -X POST http://localhost:8081/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-$(date +%s)" \
  -d '{
    "request_id": "test-direct-123",
    "user_id": "test-user",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"name": "Test User"}
  }'
```

## Message Format

Messages are JSON with this structure:

```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "notification_type": "email",
  "user_id": "usr_123",
  "template_code": "welcome_email",
  "subject": "Welcome to Our Service!",
  "body": "<h1>Welcome John!</h1>",
  "text_body": "Welcome John!",
  "priority": "normal",
  "metadata": {},
  "created_at": "2025-11-12T16:30:00Z",
  "retry_count": 0
}
```

**Full specification:** See `shared/contracts/kafka-message-contract.md`

## Integration Steps

1. **Start Kafka** (if not already running)
2. **Start your consumer** (using one of the scripts above)
3. **Generate test messages** (via API)
4. **Verify messages are received** in your consumer
5. **Integrate into your service:**
   - Parse JSON messages
   - Fetch user data (email/device tokens) from User Service
   - Send notifications
   - Update status via orchestrator API

## Updating Notification Status

After processing a message, update the status:

```bash
curl -X PUT http://localhost:8080/api/v1/email/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "notification_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "delivered",
    "timestamp": "2025-11-12T16:30:05Z"
  }'
```

## Troubleshooting

### No messages appearing

1. Check if orchestrator is running: `docker ps | grep orchestrator`
2. Check orchestrator logs: `docker logs orchestrator | grep -i kafka`
3. Verify topics exist: `docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092`
4. Check message count: `docker exec -it kafka kafka-run-class kafka.tools.GetOffsetShell --broker-list localhost:9092 --topic email.queue`

### Connection refused

- Ensure Kafka is running: `docker ps | grep kafka`
- Check Kafka logs: `docker logs kafka`
- Verify port: `nc -zv localhost 9092`

### Consumer not receiving messages

- Check consumer group: `docker exec -it kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list`
- Reset consumer group offset: `docker exec -it kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group test-consumer-email.queue --reset-offsets --to-earliest --topic email.queue --execute`

## Next Steps

1. Review `shared/contracts/kafka-message-contract.md` for full message specification
2. Integrate Kafka consumer into your service
3. Implement message processing logic
4. Add status update calls to orchestrator
5. Test end-to-end flow

