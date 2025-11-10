# Heroku Deployment Guide

This guide will help you deploy the User Notification Preferences microservice to Heroku.

## Prerequisites

- [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed
- Git repository initialized
- Heroku account

## Quick Deployment Steps

### 1. Create Heroku App

```bash
heroku create your-app-name
```

### 2. Add PostgreSQL Database

```bash
heroku addons:create heroku-postgresql:essential-0
```

This automatically sets the `DATABASE_URL` environment variable.

### 3. Set Environment Variables

```bash
heroku config:set NODE_ENV=production
heroku config:set PORT=8080
heroku config:set GRAPHQL_PLAYGROUND=false
```

**Note:** `DATABASE_URL` is automatically set by Heroku PostgreSQL addon.

### 4. Create Procfile

Create a file named `Procfile` in the root directory:

```
web: node dist/main.js
```

### 5. Update package.json

Ensure you have these scripts in `package.json`:

```json
{
  "scripts": {
    "build": "nest build",
    "start:prod": "node dist/main",
    "heroku-postbuild": "npm run build"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  }
}
```

### 6. Deploy to Heroku

```bash
git add .
git commit -m "Prepare for Heroku deployment"
git push heroku main
```

### 7. Run Database Migrations

```bash
# Connect to Heroku PostgreSQL
heroku pg:psql

# Copy and paste the contents of database/schema.sql
# Or run it directly:
heroku pg:psql < database/schema.sql
```

### 8. Open Your App

```bash
heroku open
```

Your GraphQL endpoint will be at:

```
https://your-app-name.herokuapp.com/api/v1/graphql
```

## Environment Variables Reference

### Automatically Set by Heroku

- `DATABASE_URL` - PostgreSQL connection string (set by `heroku-postgresql` addon)
- `PORT` - Port number (managed by Heroku)

### You Need to Set

```bash
heroku config:set NODE_ENV=production
heroku config:set GRAPHQL_PLAYGROUND=false  # Disable in production
```

### View All Environment Variables

```bash
heroku config
```

## Database Configuration

The app automatically detects `DATABASE_URL` and uses it for database connection. If `DATABASE_URL` is not set, it falls back to individual connection parameters (`DB_HOST`, `DB_PORT`, etc.).

**Heroku PostgreSQL format:**

```
DATABASE_URL=postgresql://user:password@host:5432/database
```

## SSL Configuration

The app automatically enables SSL for production environments when using `DATABASE_URL`, which is required by Heroku PostgreSQL.

## Monitoring

### View Logs

```bash
heroku logs --tail
```

### Check App Status

```bash
heroku ps
```

### Database Info

```bash
heroku pg:info
```

## Troubleshooting

### Build Failures

```bash
# Check build logs
heroku logs --tail

# Ensure all dependencies are in package.json
npm install --save-prod <missing-package>
```

### Database Connection Issues

```bash
# Verify DATABASE_URL is set
heroku config:get DATABASE_URL

# Check database status
heroku pg:info

# View database logs
heroku logs --ps postgres
```

### App Not Starting

```bash
# Check if Procfile exists
cat Procfile

# Verify build succeeded
heroku builds

# Check dyno status
heroku ps
```

## Performance Optimization

### Scale Dynos

```bash
# Scale to multiple dynos
heroku ps:scale web=2

# Use hobby dyno (always on)
heroku dyno:type hobby
```

### Database Performance

```bash
# Upgrade database plan for better performance
heroku addons:upgrade heroku-postgresql:standard-0
```

## Security Best Practices

1. **Disable GraphQL Playground in production:**

   ```bash
   heroku config:set GRAPHQL_PLAYGROUND=false
   ```

2. **Enable CORS** (if needed for frontend):
   Add to `main.ts`:

   ```typescript
   app.enableCors({
     origin: process.env.ALLOWED_ORIGINS?.split(',') || [],
   });
   ```

3. **Add Authentication** (recommended):
   Implement JWT or API key authentication before deploying to production.

4. **Use environment variables** for all sensitive data - never commit them to git.

## Continuous Deployment

### GitHub Integration

1. Connect your GitHub repository to Heroku
2. Enable automatic deploys from your main branch

```bash
# Via Heroku Dashboard:
# App → Deploy → GitHub → Connect Repository → Enable Automatic Deploys
```

### Using Heroku Pipelines

```bash
# Create a pipeline
heroku pipelines:create your-pipeline-name

# Add staging and production apps
heroku pipelines:add your-pipeline-name --app your-staging-app
heroku pipelines:add your-pipeline-name --app your-production-app
```

## Useful Commands

```bash
# Restart app
heroku restart

# Open Heroku dashboard
heroku open --dashboard

# Run commands on Heroku
heroku run npm run migration

# Access database
heroku pg:psql

# Download database backup
heroku pg:backups:capture
heroku pg:backups:download

# View config
heroku config

# Set config
heroku config:set KEY=value

# Unset config
heroku config:unset KEY
```

## Cost Optimization

- **Free tier:** Use `essential-0` PostgreSQL (free for hobby projects)
- **Eco dynos:** Cost-effective for low-traffic apps
- **Scheduled scaling:** Scale down during off-hours

## Next Steps

1. Set up monitoring with Heroku metrics
2. Configure custom domain
3. Set up SSL certificate (automatic with custom domains)
4. Implement health check endpoint
5. Add application performance monitoring (APM)

## Support

- [Heroku Dev Center](https://devcenter.heroku.com/)
- [Heroku Status](https://status.heroku.com/)
- [NestJS Documentation](https://docs.nestjs.com/)
