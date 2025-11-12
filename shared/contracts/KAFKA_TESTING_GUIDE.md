# Kafka Testing Guide for Email and Push Services

This guide helps you test your Email and Push services by consuming messages from Kafka queues published by the Orchestrator Service.

## Prerequisites

1. **Docker and Docker Compose** installed
2. **Kafka running locally** (see setup instructions below)
3. **Orchestrator service running** (to publish messages)

## Quick Start

### 1. Start Kafka

```bash
# From project root
docker-compose -f docker-compose.kafka.yml up -d

# Wait for Kafka to be ready (30-60 seconds)
docker-compose -f docker-compose.kafka.yml logs -f kafka
# Look for: "KafkaServer id=1 started"
```

### 2. Verify Topics Exist

```bash
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092
```

You should see:
- `email.queue`
- `push.queue`
- `failed.queue`

### 3. Start Orchestrator (to generate messages)

```bash
# Start full stack
docker-compose -f docker-compose.gateway.yml up -d

# Or just orchestrator + dependencies
cd services/orchestrator
make run  # or use docker-compose
```

## Consuming Messages

### Option 1: Using Kafka Console Consumer (Quick Test)

#### Consume Email Queue

```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

#### Consume Push Queue

```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic push.queue \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

#### Consume Only New Messages

Remove `--from-beginning` to consume only new messages:

```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --property print.key=true
```

### Option 2: Using Python Consumer Script

See `scripts/test-kafka-consumer.py` for a Python script that:
- Connects to Kafka
- Consumes messages from a topic
- Parses JSON
- Prints formatted output
- Can be modified to call your service

### Option 3: Using Node.js Consumer Script

See `scripts/test-kafka-consumer.js` for a Node.js script.

### Option 4: Using Go Consumer Script

See `scripts/test-kafka-consumer.go` for a Go script.

## Generating Test Messages

### Via API Gateway

```bash
# Create an email notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-Request-ID: test-email-$(date +%s)" \
  -d '{
    "request_id": "test-email-123",
    "user_id": "test-user-email",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {
      "name": "Test User"
    },
    "priority": 2
  }'

# Create a push notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-Request-ID: test-push-$(date +%s)" \
  -d '{
    "request_id": "test-push-123",
    "user_id": "test-user-push",
    "template_code": "welcome_push",
    "notification_type": "push",
    "variables": {
      "name": "Test User"
    },
    "priority": 2
  }'
```

### Via Direct Orchestrator API

```bash
# Direct to orchestrator (port 8081)
curl -X POST http://localhost:8081/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-direct-$(date +%s)" \
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

See `kafka-message-contract.md` for full specification.

## Testing Your Service

### Step 1: Consume Messages

Start consuming from your topic:

```bash
# Email service
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic email.queue \
  --from-beginning
```

### Step 2: Generate Test Messages

Use the API to create notifications (see above).

### Step 3: Verify Your Service Receives Messages

Your service should:
1. Receive the JSON message
2. Parse the fields
3. Fetch user data (email/device tokens) from User Service
4. Send the notification
5. Update status via orchestrator API

### Step 4: Update Status

After processing, update the status:

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

## Consumer Scripts

We provide consumer scripts in multiple languages. See the `scripts/` directory:

- `test-kafka-consumer.py` - Python consumer
- `test-kafka-consumer.js` - Node.js consumer
- `test-kafka-consumer.go` - Go consumer

## Troubleshooting

### No messages appearing

1. **Check if orchestrator is running:**
   ```bash
   docker ps | grep orchestrator
   ```

2. **Check orchestrator logs:**
   ```bash
   docker logs orchestrator | grep -i kafka
   ```

3. **Verify topics exist:**
   ```bash
   docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092
   ```

4. **Check message count:**
   ```bash
   docker exec -it kafka kafka-run-class kafka.tools.GetOffsetShell \
     --broker-list localhost:9092 \
     --topic email.queue
   ```

### Connection refused

- Ensure Kafka is running: `docker ps | grep kafka`
- Check Kafka logs: `docker logs kafka`
- Verify port 9092 is accessible: `nc -zv localhost 9092`

### Messages not in expected format

- Check orchestrator version
- Verify template service is returning correct format
- Check orchestrator logs for errors

## Next Steps

1. Integrate Kafka consumer into your service
2. Parse JSON messages
3. Fetch user data from User Service
4. Send notifications
5. Update status via orchestrator API

See `kafka-message-contract.md` for detailed message specifications.

