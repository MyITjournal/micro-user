# Submit User Preferences - API Examples

## REST API

### Endpoint

```
POST http://localhost:8080/api/v1/users/preferences
Content-Type: application/json
```

### Request Body Example

```json
{
  "user_id": "usr_abc123",
  "email": "newuser@example.com",
  "phone": "+1234567890",
  "timezone": "America/New_York",
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
        "timezone": "America/New_York"
      }
    },
    "push": {
      "enabled": true,
      "devices": [
        {
          "device_id": "dev_mobile_001",
          "platform": "ios",
          "token": "fcm_token_example_123",
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
  }
}
```

### cURL Example

```bash
curl -X POST http://localhost:8080/api/v1/users/preferences \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_abc123",
    "email": "newuser@example.com",
    "phone": "+1234567890",
    "timezone": "America/New_York",
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
          "timezone": "America/New_York"
        }
      },
      "push": {
        "enabled": true,
        "devices": [
          {
            "device_id": "dev_mobile_001",
            "platform": "ios",
            "token": "fcm_token_example_123",
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
    }
  }'
```

### Success Response (200 OK)

```json
{
  "user_id": "usr_abc123",
  "email": "newuser@example.com",
  "phone": "+1234567890",
  "timezone": "America/New_York",
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
        "timezone": "America/New_York"
      }
    },
    "push": {
      "enabled": true,
      "devices": [
        {
          "device_id": "dev_mobile_001",
          "platform": "ios",
          "token": "fcm_token_example_123",
          "last_seen": "2025-11-10T14:30:00.000Z",
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
  "updated_at": "2025-11-10T14:30:00.000Z"
}
```

## GraphQL API

### Mutation

```graphql
mutation SubmitUserPreferences($input: CreateUserPreferencesInput!) {
  submitUserPreferences(input: $input) {
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

### Variables

```json
{
  "input": {
    "user_id": "usr_abc123",
    "email": "newuser@example.com",
    "phone": "+1234567890",
    "timezone": "America/New_York",
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
          "timezone": "America/New_York"
        }
      },
      "push": {
        "enabled": true,
        "devices": [
          {
            "device_id": "dev_mobile_001",
            "platform": "ios",
            "token": "fcm_token_example_123",
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
    }
  }
}
```

### cURL for GraphQL

```bash
curl -X POST http://localhost:8080/api/v1/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation SubmitUserPreferences($input: CreateUserPreferencesInput!) { submitUserPreferences(input: $input) { user_id email notification_enabled updated_at } }",
    "variables": {
      "input": {
        "user_id": "usr_abc123",
        "email": "newuser@example.com",
        "timezone": "America/New_York",
        "language": "en",
        "notification_enabled": true,
        "preferences": {
          "marketing": false,
          "transactional": true,
          "reminders": true,
          "digest": {
            "enabled": true,
            "frequency": "daily",
            "time": "09:00"
          }
        }
      }
    }
  }'
```

## Minimal Example (Without Channels)

You can also create user preferences without specifying channels (they'll be created with defaults later):

### REST

```bash
curl -X POST http://localhost:8080/api/v1/users/preferences \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_simple",
    "email": "simple@example.com",
    "timezone": "UTC",
    "language": "en",
    "notification_enabled": true,
    "preferences": {
      "marketing": false,
      "transactional": true,
      "reminders": true,
      "digest": {
        "enabled": false,
        "frequency": "daily",
        "time": "09:00"
      }
    }
  }'
```

## Update Existing User

The same endpoint works for both creating new users and updating existing ones. If the `user_id` already exists, the preferences will be updated.

## Field Validations

- `user_id`: Required, string
- `email`: Required, valid email format
- `phone`: Optional, string
- `timezone`: Required, string (IANA timezone)
- `language`: Required, string (ISO 639-1 code)
- `notification_enabled`: Required, boolean
- `channels`: Optional object
- `preferences`: Required object

### Frequency Options

- `immediate`
- `batched`
- `digest`

### Platform Options

- `ios`
- `android`
- `web`

### Digest Frequency Options

- `daily`
- `weekly`
- `monthly`

## Error Responses

### Invalid Input (400)

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid input data",
    "details": {
      "error": "email must be a valid email"
    }
  }
}
```
