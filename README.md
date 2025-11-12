# User Notification Preferences Microservice

A NestJS-based GraphQL microservice for managing user notification preferences with support for multiple channels (email, push) and customizable settings.

## Features

- 🔔 User notification preference management
- 📧 Email channel configuration with quiet hours
- 📱 Push notification management with device tracking
- 🌍 Multi-timezone support
- 🎯 GraphQL + REST API with type safety
- 💾 PostgreSQL database with TypeORM
- ✅ Input validation with class-validator
- 📦 Batch operations for multiple users
- ⚡ Fast opt-out status checks (<100ms)
- 📊 Last notification tracking

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

# Module Separation - Simple Users vs Complex Users

## Overview

The application has been successfully split into two separate modules:

1. **Simple Users Module** - `/api/v1/users`
2. **Complex Users Module** - `/api/v1/cusers`

## Module Structure

### Simple Users Module (`src/simple-users/`)

**Purpose:** Lightweight user management with basic preferences and notification tracking

**Database:** `simple_users` table

**Endpoints:**

- `POST /api/v1/users` - Create a simple user
- `GET /api/v1/users/:user_id/preferences` - Get user preferences
- `POST /api/v1/users/preferences/batch` - Batch get user preferences (max 100)
- `POST /api/v1/users/:user_id/last-notification` - Update last notification time (fire-and-forget)

**Entity Fields:**

- `user_id` - Primary key (usr_xxxxxxxx)
- `name` - User's name
- `email` - Unique email
- `password` - Bcrypt hashed password
- `push_token` - Optional push notification token
- `email_preference` - Boolean for email notifications
- `push_preference` - Boolean for push notifications
- `last_notification_email` - Last email notification timestamp
- `last_notification_push` - Last push notification timestamp
- `last_notification_id` - Last notification ID
- `created_at` - Creation timestamp
- `updated_at` - Update timestamp

**Files:**

- `simple-users.module.ts` - Module definition
- `simple-users.controller.ts` - REST controller
- `simple-users.service.ts` - Business logic
- `entity/simple-user.entity.ts` - TypeORM entity
- `dto/simple-user.dto.ts` - DTOs for validation

### Complex Users Module (`src/complex-users/`)

**Purpose:** Full-featured notification preference system with channels and devices

**Database:** `users`, `user_channels`, `user_devices` tables

**Endpoints:**

- `GET /api/v1/cusers/:user_id/preferences` - Get comprehensive user preferences
- `POST /api/v1/cusers/preferences` - Create/update user preferences
- `POST /api/v1/cusers/preferences/batch` - Batch get user preferences (max 100)
- `GET /api/v1/cusers/:user_id/opt-out-status` - Check opt-out status
- `POST /api/v1/cusers/:user_id/last-notification` - Update last notification time
- GraphQL endpoint at `/api/v1/graphql`

**Entity Features:**

- Full user profile with timezone, language
- Notification preferences (marketing, transactional, reminders)
- Digest settings (frequency, time)
- Email and push channels with quiet hours
- Multiple devices per user
- Verified status for channels

**Files:**

- `users.module.ts` - Module definition (exported as ComplexUsersModule)
- `user.controller.ts` - REST controller
- `user.service.ts` - Business logic
- `user.resolver.ts` - GraphQL resolver
- `entity/user.entity.ts` - User entity
- `entity/usersChannel.entity.ts` - Channel entity
- `entity/userDevices.entity.ts` - Device entity
- `dto/user.dto.ts` - DTOs for validation and GraphQL

## Testing

**Run the comprehensive test:**

```bash
node test-both-modules.js
```

**Test simple users only:**

```bash
node test-endpoint.js
```

**Example - Create Simple User:**

```bash
curl -X POST http://localhost:8000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "preferences": {
      "email": true,
      "push": false
    }
  }'
```

## Key Differences

| Feature         | Simple Users      | Complex Users           |
| --------------- | ----------------- | ----------------------- |
| Route           | `/api/v1/users`   | `/api/v1/cusers`        |
| Database        | single table      | 3 tables with relations |
| Authentication  | password field    | no auth fields          |
| Channels        | just booleans     | full channel config     |
| Devices         | single push_token | multiple devices        |
| Quiet Hours     | no                | yes                     |
| GraphQL         | no                | yes                     |
| Timezone        | no                | yes                     |
| Digest          | no                | yes                     |
| Marketing prefs | no                | yes                     |

## Benefits of Separation

1. **Simplicity** - Simple users for basic use cases
2. **Performance** - Faster queries on simple_users table
3. **Flexibility** - Each module can evolve independently
4. **Clear Separation** - Different routes avoid confusion
5. **Scalability** - Can deploy modules separately if needed

## Database Migration

The `simple_users` table is created by the migration script:

- `database/migrations/create_simple_users_table.sql`

This runs automatically on Heroku deployment via Procfile.

## Server Status

Both modules are loaded successfully:

- ✓ SimpleUsersModule initialized
- ✓ ComplexUsersModule initialized
- ✓ All endpoints mapped correctly
- ✓ 0 compilation errors

## Next Steps

1. Test both modules with real data
2. Add authentication middleware if needed
3. Consider rate limiting
4. Add API documentation (Swagger/OpenAPI)
5. Monitor performance of both modules
