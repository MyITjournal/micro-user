# GraphQL Query Examples

## Get User Preferences (Full)

```graphql
query GetUserPreferencesFull {
  getUserPreferences(user_id: "usr_7x9k2p", include_channels: true) {
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

## Get User Preferences (Without Channels)

```graphql
query GetUserPreferencesBasic {
  getUserPreferences(user_id: "usr_7x9k2p", include_channels: false) {
    user_id
    email
    phone
    timezone
    language
    notification_enabled
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

## Get User Preferences (Default - includes channels)

```graphql
query GetUserPreferencesDefault {
  getUserPreferences(user_id: "usr_7x9k2p") {
    user_id
    email
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
          active
        }
      }
    }
    preferences {
      marketing
      transactional
    }
  }
}
```

## Using Variables

```graphql
query GetUserPreferences($userId: String!, $includeChannels: Boolean) {
  getUserPreferences(user_id: $userId, include_channels: $includeChannels) {
    user_id
    email
    notification_enabled
    channels @include(if: $includeChannels) {
      email {
        enabled
      }
      push {
        enabled
      }
    }
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

## cURL Examples

### Using GraphQL HTTP Request

```bash
curl -X POST http://localhost:8080/api/v1/graphql \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "query": "query GetUserPreferences($userId: String!, $includeChannels: Boolean) { getUserPreferences(user_id: $userId, include_channels: $includeChannels) { user_id email notification_enabled } }",
    "variables": {
      "userId": "usr_7x9k2p",
      "includeChannels": true
    }
  }'
```

### Inline Query

```bash
curl -X POST http://localhost:8080/api/v1/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ getUserPreferences(user_id: \"usr_7x9k2p\") { user_id email notification_enabled } }"
  }'
```

## Testing in GraphQL Playground

1. Start the application: `npm run start:dev`
2. Open browser: `http://localhost:8080/api/v1/graphql`
3. Paste any query from above
4. Click "Play" button to execute

## Expected Responses

### Success (200 OK)

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
              "last_seen": "2025-01-15T10:25:00.000Z",
              "active": true
            },
            {
              "device_id": "dev_def456",
              "platform": "android",
              "token": "fcm_token_abc...",
              "last_seen": "2025-01-14T18:30:00.000Z",
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
      "updated_at": "2025-01-15T08:00:00.000Z"
    }
  }
}
```

### Error - User Not Found

```json
{
  "errors": [
    {
      "message": "{\"code\":\"USER_NOT_FOUND\",\"message\":\"User with ID usr_invalid does not exist\",\"details\":{\"user_id\":\"usr_invalid\"}}",
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "path": [
        "getUserPreferences"
      ],
      "extensions": {
        "code": "INTERNAL_SERVER_ERROR",
        "stacktrace": [...]
      }
    }
  ],
  "data": null
}
```

## GraphQL Schema (Auto-generated)

The schema is automatically generated by NestJS. To view it:

1. Run the application
2. Open GraphQL Playground
3. Click "SCHEMA" tab on the right side

You'll see the complete type definitions for:

- `UserPreferencesResponse`
- `ChannelsDto`
- `EmailChannelDto`
- `PushChannelDto`
- `DeviceDto`
- `QuietHoursDto`
- `UserPreferencesDto`
- `DigestPreferenceDto`
