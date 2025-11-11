package clients

import (
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type UserClient interface {
	GetPreferences(userID string) (*models.UserPreferences, error)
}
