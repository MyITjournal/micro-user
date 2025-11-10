# Orchestrator Service

The Orchestrator Service is the central coordinator for the notification system. It orchestrates the flow of notifications by:

1. Fetching user preferences
2. Validating notification eligibility
3. Rendering templates
4. Queuing notifications for delivery

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (optional)

### Local Development
```bash
# Install dependencies
make deps

# Run the service
make run

# Run tests
make test
```

The service will start on `http://localhost:8080`

### Using Docker
```bash
# Build and run
make docker-run

# Stop
make docker-stop
```

## API Endpoints

### Create Notification
```http
POST /api/v1/notifications
Content-Type: application/json

{
  "user_id": "usr_7x9k2p",
  "template_id": "welcome_email",
  "channel": "email",
  "variables": {
    "user_name": "John Doe",
    "app_name": "MyApp"
  },
  "priority": "normal"
}
```

**Response (201 Created):**
```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "message": "Notification queued for delivery",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Health Checks
```http
GET /health          # Simple health check
GET /health/live     # Liveness probe
GET /health/ready    # Readiness probe
```

## Testing

### Test with cURL
```bash
# Health check
curl http://localhost:8080/health

# Create notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_123",
    "template_id": "welcome_email",
    "channel": "email",
    "variables": {
      "user_name": "John Doe",
      "app_name": "MyApp"
    }
  }'
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `USE_MOCK_SERVICES` | `true` | Use mock services (set false for real services) |
| `USER_SERVICE_URL` | `http://user-service:8081` | User service base URL |
| `TEMPLATE_SERVICE_URL` | `http://template-service:8082` | Template service base URL |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, console) |

## Development

### Project Structure