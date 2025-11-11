release: psql $DATABASE_URL -f database/migrations/add_last_notification_columns.sql && node scripts/seed.js
web: npm run start:prod
