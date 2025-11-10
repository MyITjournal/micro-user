# Quick Start Guide

## 🚀 Getting Started in 5 Minutes

### Step 1: Install Dependencies

```bash
npm install
```

### Step 2: Setup Database

Make sure PostgreSQL is running, then:

```bash
# Connect to PostgreSQL
psql -U postgres

# Run the schema file
\i database/schema.sql

# Or on Windows PowerShell:
# Get-Content database/schema.sql | psql -U postgres
```

This creates:

- Database: `user_service`
- Tables: `users`, `user_channels`, `user_devices`
- Sample user: `usr_7x9k2p`

### Step 3: Configure Environment

```bash
# Copy the example environment file
copy .env.example .env

# Update .env with your database credentials if needed
```

### Step 4: Run the Application

```bash
npm run start:dev
```

You should see:

```
[Nest] INFO [NestApplication] Nest application successfully started
[Nest] INFO Mapped {/api/v1/graphql, POST}
```

### Step 5: Test the API

Open your browser to:
**http://localhost:8080/api/v1/graphql**

Paste this query in GraphQL Playground:

```graphql
{
  getUserPreferences(user_id: "usr_7x9k2p") {
    user_id
    email
    notification_enabled
    channels {
      email {
        enabled
        verified
      }
      push {
        enabled
        devices {
          device_id
          platform
        }
      }
    }
  }
}
```

Click ▶️ **Play** button!

## 📝 What You Just Built

A complete user notification preference system with:

✅ **GraphQL API** for querying user preferences  
✅ **TypeORM entities** with proper relationships  
✅ **Multi-channel support** (email, push)  
✅ **Device management** for push notifications  
✅ **Quiet hours** configuration  
✅ **Timezone support**  
✅ **Marketing & digest preferences**

## 🎯 Next Steps

1. **Explore the API**: Check `GRAPHQL_EXAMPLES.md` for more queries
2. **Read the docs**: See `USER_PREFERENCES_API.md` for full documentation
3. **Add mutations**: Create update/create operations in `user.resolver.ts`
4. **Add REST endpoints**: Create a controller if you need REST API alongside GraphQL
5. **Add authentication**: Integrate JWT or OAuth
6. **Add tests**: Write unit and e2e tests

## 📂 Project Structure

```
micro-user/
├── src/
│   ├── users/
│   │   ├── entity/           # Database models
│   │   ├── dto/              # GraphQL types & validation
│   │   ├── user.service.ts   # Business logic
│   │   ├── user.resolver.ts  # GraphQL queries
│   │   └── users.module.ts   # Module config
│   └── app.module.ts         # App setup (GraphQL + TypeORM)
├── database/
│   └── schema.sql            # Database schema
├── .env.example              # Environment template
├── GRAPHQL_EXAMPLES.md       # Query examples
└── USER_PREFERENCES_API.md   # Full API docs
```

## 🐛 Troubleshooting

### Database Connection Error

```bash
# Check PostgreSQL is running
pg_isready

# Verify credentials in .env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
```

### Port Already in Use

Change port in `.env`:

```env
PORT=3000
```

### TypeORM Synchronize Warning

In production, set in `.env`:

```env
NODE_ENV=production
```

This disables auto-schema sync. Use migrations instead.

## 💡 Tips

- **Auto-reload**: The app watches for file changes in dev mode
- **GraphQL Playground**: Great for testing queries interactively
- **Type Safety**: TypeScript + GraphQL gives you end-to-end types
- **Validation**: DTOs automatically validate input with `class-validator`

## 🔗 Useful Commands

```bash
# Development
npm run start:dev

# Production build
npm run build
npm run start:prod

# Linting
npm run lint

# Testing
npm run test
npm run test:e2e

# Format code
npm run format
```

Happy coding! 🎉
