# User Notification Preferences Microservice

A NestJS-based GraphQL microservice for managing user notification preferences with support for multiple channels (email, push) and customizable settings.

## Features

- 🔔 User notification preference management
- 📧 Email channel configuration with quiet hours
- 📱 Push notification management with device tracking
- 🌍 Multi-timezone support
- 🎯 GraphQL API with type safety
- 💾 PostgreSQL database with TypeORM
- ✅ Input validation with class-validator

## Architecture

This service follows NestJS best practices with a modular architecture:

```
src/
├── users/
│   ├── entity/           # TypeORM entities
│   │   ├── user.entity.ts
│   │   ├── usersChannel.entity.ts
│   │   └── userDevices.entity.ts
│   ├── dto/              # Data Transfer Objects
│   │   └── user.dto.ts
│   ├── user.service.ts   # Business logic
│   ├── user.resolver.ts  # GraphQL resolver
│   └── users.module.ts   # Module configuration
└── app.module.ts         # Root module with GraphQL & TypeORM setup
```

## Database Schema

### Entities

1. **User** - Core user information and global preferences
2. **UserChannel** - Channel-specific settings (email, push)
3. **UserDevice** - Push notification device registry

### Relationships

- User → UserChannel (One-to-Many)
- UserChannel → UserDevice (One-to-Many)

## Setup

### Prerequisites

- Node.js >= 18
- PostgreSQL >= 14
- npm or yarn

### Installation

1. Clone the repository
2. Install dependencies:

   ```bash
   npm install
   ```

3. Create `.env` file from example:

   ```bash
   copy .env.example .env
   ```

4. Configure database connection in `.env`:

   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USERNAME=postgres
   DB_PASSWORD=postgres
   DB_NAME=user_service
   ```

5. Initialize database:
   ```bash
   psql -U postgres -f database/schema.sql
   ```

### Running the Application

```bash
# Development mode
npm run start:dev

# Production mode
npm run build
npm run start:prod
```

The service will be available at:

- **GraphQL Endpoint**: `http://localhost:8080/api/v1/graphql`
- **GraphQL Playground**: `http://localhost:8080/api/v1/graphql` (in browser)

## API Usage

### GraphQL Query

#### Get User Preferences

```graphql
query GetUserPreferences($userId: String!, $includeChannels: Boolean) {
  getUserPreferences(user_id: $userId, include_channels: $includeChannels) {
    user_id
    email
    phone
    timezone
    language
    notification_enabled
    channels {
      email {
        enabled
        verified
        frequency
        quiet_hours {
          enabled
          start
          end
          timezone
        }
      }
      push {
        enabled
        devices {
          device_id
          platform
          token
          last_seen
          active
        }
        quiet_hours {
          enabled
          start
          end
          timezone
        }
      }
    }
    preferences {
      marketing
      transactional
      reminders
      digest {
        enabled
        frequency
        time
      }
    }
    updated_at
  }
}
```

**Variables:**

```json
{
  "userId": "usr_7x9k2p",
  "includeChannels": true
}
```

### Example Response

```json
{
  "data": {
    "getUserPreferences": {
      "user_id": "usr_7x9k2p",
      "email": "user@example.com",
      "phone": "+254712345678",
      "timezone": "Africa/Nairobi",
      "language": "en",
      "notification_enabled": true,
      "channels": {
        "email": {
          "enabled": true,
          "verified": true,
          "frequency": "immediate",
          "quiet_hours": {
            "enabled": true,
            "start": "22:00",
            "end": "07:00",
            "timezone": "Africa/Nairobi"
          }
        },
        "push": {
          "enabled": true,
          "devices": [
            {
              "device_id": "dev_abc123",
              "platform": "ios",
              "token": "fcm_token_xyz...",
              "last_seen": "2025-01-15T10:25:00Z",
              "active": true
            }
          ],
          "quiet_hours": {
            "enabled": false
          }
        }
      },
      "preferences": {
        "marketing": false,
        "transactional": true,
        "reminders": true,
        "digest": {
          "enabled": true,
          "frequency": "daily",
          "time": "09:00"
        }
      },
      "updated_at": "2025-01-15T08:00:00Z"
    }
  }
}
```

## Error Handling

The service implements proper error handling:

### User Not Found (404)

```json
{
  "errors": [
    {
      "message": "User with ID usr_invalid does not exist",
      "extensions": {
        "code": "USER_NOT_FOUND"
      }
    }
  ]
}
```

## Testing

```bash
# Unit tests
npm run test

# E2E tests
npm run test:e2e

# Test coverage
npm run test:cov
```

## Development

### Adding New Features

1. Create/update entities in `src/users/entity/`
2. Define DTOs in `src/users/dto/`
3. Implement business logic in `src/users/user.service.ts`
4. Add GraphQL queries/mutations in `src/users/user.resolver.ts`
5. Update module imports if needed

### Code Style

```bash
# Lint
npm run lint

# Format
npm run format
```

## Deployment

### Docker

Create a `Dockerfile`:

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 8080
CMD ["node", "dist/main"]
```

Build and run:

```bash
docker build -t user-service .
docker run -p 8080:8080 user-service
```

## Environment Variables

| Variable      | Description       | Default        |
| ------------- | ----------------- | -------------- |
| `PORT`        | Server port       | `8080`         |
| `NODE_ENV`    | Environment       | `development`  |
| `DB_HOST`     | Database host     | `localhost`    |
| `DB_PORT`     | Database port     | `5432`         |
| `DB_USERNAME` | Database user     | `postgres`     |
| `DB_PASSWORD` | Database password | `postgres`     |
| `DB_NAME`     | Database name     | `user_service` |

## License

UNLICENSED
