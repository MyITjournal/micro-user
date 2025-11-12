package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/config"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
	_ "github.com/lib/pq"
)

// TestDBConfig returns a test database configuration
func TestDBConfig() config.PostgreSQLConfig {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "orchestrator_test"
	}

	return config.PostgreSQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  "disable",
		MaxConns: 10,
	}
}

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) (*database.DB, func()) {
	cfg := TestDBConfig()

	// Connect to database
	db, err := database.NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db.DB, t); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		cleanupTestData(db.DB, t)
		db.Close()
	}

	return db, cleanup
}

// runMigrations executes SQL migration files
func runMigrations(db *sql.DB, t *testing.T) error {
	migrationsDir := filepath.Join("..", "..", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try alternative path
		migrationsDir = filepath.Join("migrations")
	}

	// Read and execute migration file
	migrationFile := filepath.Join(migrationsDir, "001_create_notifications_table.up.sql")
	migrationSQL, err := os.ReadFile(migrationFile)
	if err != nil {
		// If file doesn't exist, try to create tables directly
		return createTablesDirectly(db)
	}

	// Execute migration
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		// Check if tables already exist
		if isTableExistsError(err) {
			return nil // Tables already exist, that's fine
		}
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// createTablesDirectly creates tables if migration file is not found
func createTablesDirectly(db *sql.DB) error {
	// Check if tables already exist
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'notifications'
		)
	`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		return nil // Tables already exist
	}

	// Create enums
	_, err = db.Exec(`
		DO $$ BEGIN
			CREATE TYPE notification_type_enum AS ENUM ('email', 'push');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to create notification_type_enum: %w", err)
	}

	_, err = db.Exec(`
		DO $$ BEGIN
			CREATE TYPE notification_status_enum AS ENUM ('pending', 'delivered', 'failed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to create notification_status_enum: %w", err)
	}

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			template_code VARCHAR(255) NOT NULL,
			notification_type notification_type_enum NOT NULL,
			status notification_status_enum NOT NULL DEFAULT 'pending',
			priority VARCHAR(50) NOT NULL DEFAULT 'normal',
			variables JSONB NOT NULL DEFAULT '{}',
			metadata JSONB,
			error_message TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			scheduled_for TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notifications table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status)",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create trigger function and trigger
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger function: %w", err)
	}

	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
		CREATE TRIGGER update_notifications_updated_at
			BEFORE UPDATE ON notifications
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	return nil
}

// isTableExistsError checks if error is due to table already existing
func isTableExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "already exists") || contains(errStr, "duplicate")
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return false
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// cleanupTestData removes all test data from the database
func cleanupTestData(db *sql.DB, t *testing.T) {
	ctx := context.Background()

	// Disable foreign key checks temporarily (if any)
	// Truncate tables
	tables := []string{"notifications"}

	for _, table := range tables {
		// Use TRUNCATE with CASCADE to handle any dependencies
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
			// Fallback to DELETE
			deleteQuery := fmt.Sprintf("DELETE FROM %s", table)
			if _, err := db.ExecContext(ctx, deleteQuery); err != nil {
				t.Logf("Warning: Failed to delete from table %s: %v", table, err)
			}
		}
	}
}
