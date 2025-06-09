package testhelpers

import (
	"testing"

	"github.com/konflux-ci/kite/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	// Use SQLite in-memory DB for tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to created test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.IssueScope{},
		&models.Issue{},
		&models.Link{},
		&models.RelatedIssue{},
	)

	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}
