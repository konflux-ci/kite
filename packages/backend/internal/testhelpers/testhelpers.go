package testhelpers

import (
	"fmt"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing.
// This database does not persist changes and is isolated per connection.
// Use this when your tests need a clean database.
func SetupTestDB(t *testing.T) *gorm.DB {
	// Mark as a test helper for better error reporting
	t.Helper()

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

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				t.Fatalf("Failed to close test database: %v", err)
			}
		}
	})

	return db
}

// SetupConcurrentTestDB creates an in-memory SQLite database for testing.
// Uses shared cache mode to ensure goroutines access the same database.
//
// Use this for tests that:
//   - Launch multiple goroutines
//   - Test race conditions
//   - Validation transaction isolation
func SetupConcurrentTestDB(t *testing.T) *gorm.DB {
	// Mark as a test helper for better error reporting
	t.Helper()

	// Use shared cache mode for concurrent access
	// This ensures that all connections share the same in-memory database.
	// Without this, each goroutine gets its own isolated DB instance.
	dsn := "file::memory:?cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Get underlying SQL DB for configuration
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying SQL database: %v", err)
	}

	// Configure connection pool for concurrent testing
	// SQLite only supports one writer at a time, so we can limit
	// connections to prevent "database is locked" errors.
	sqlDB.SetMaxOpenConns(1) // One writer at a time
	sqlDB.SetMaxIdleConns(1) // Keep one connection ready

	// Apply SQLite optimizers for better concurrent access:
	//
	// Enable WAL mode for better concurrency in memory.
	// Without WAL: Writer locks the entire DB.
	// With WAL: Writers append to log, readers read from main file
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		t.Logf("Warning: Could not enable WAL mode: %v", err)
	}
	// Wait up to 5 seconds if DB is locked instead of failing.
	if err := db.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
		t.Logf("Warning: Could not set busy timeout: %v", err)
	}
	// Ensure foreign key constraints are enforced.
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		t.Logf("Warning: Could not enable foreign keys: %v", err)
	}

	// Run DB migration
	err = db.AutoMigrate(
		&models.IssueScope{},
		&models.Issue{},
		&models.Link{},
		&models.RelatedIssue{},
	)

	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Cleanup with test finishes
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				t.Fatalf("Failed to close test database: %v", err)
			}
		}
	})

	return db
}

// CompareIssues performs a simple comparison on two issues.
//
// The values compared are: ID, Title, Namespace, Severity, IssueType, State
//
// Returns an error if a comparison fails
func CompareIssues(a, b models.Issue) error {
	if a.ID != b.ID {
		return fmt.Errorf("expected ID '%s', got '%s'", a.ID, b.ID)
	}
	if a.Title != b.Title {
		return fmt.Errorf("expected Title '%s', got '%s'", a.Title, b.Title)
	}
	if a.Namespace != b.Namespace {
		return fmt.Errorf("expected Namespace '%s', got '%s'", a.Namespace, b.Namespace)
	}
	if a.Description != b.Description {
		return fmt.Errorf("expected Description '%s', got '%s'", a.Description, b.Description)
	}
	if a.Severity != b.Severity {
		return fmt.Errorf("expected Severity '%s', got '%s'", a.Severity, b.Severity)
	}
	if a.IssueType != b.IssueType {
		return fmt.Errorf("expected IssueType '%s', got '%s'", a.IssueType, b.IssueType)
	}
	if a.State != b.State {
		return fmt.Errorf("expected State '%s', got '%s'", a.State, b.State)
	}

	return nil
}

// CompareIssueToDTO performs a simple comparison on an Issue and the CreateIssueRequest struct
// used to create that issue.
//
// The values compared are: Title, Namespace, Severity, IssueType, State
//
// Returns an error if a comparison fails
func CompareIssueToDTO(a models.Issue, b dto.CreateIssueRequest) error {
	if a.Title != "" && a.Title != b.Title {
		return fmt.Errorf("expected Title '%s', got '%s'", a.Title, b.Title)
	}
	if a.Namespace != "" && a.Namespace != b.Namespace {
		return fmt.Errorf("expected Namespace '%s', got '%s'", a.Namespace, b.Namespace)
	}
	if a.Description != "" && a.Description != b.Description {
		return fmt.Errorf("expected Description '%s', got '%s'", a.Description, b.Description)
	}
	if string(a.Severity) != "" && a.Severity != b.Severity {
		return fmt.Errorf("expected Severity '%s', got '%s'", a.Severity, b.Severity)
	}
	if string(a.IssueType) != "" && a.IssueType != b.IssueType {
		return fmt.Errorf("expected IssueType '%s', got '%s'", a.IssueType, b.IssueType)
	}

	if string(b.State) != "" && a.State != b.State {
		return fmt.Errorf("expected State '%s', got '%s'", a.State, b.State)
	}

	return nil
}
