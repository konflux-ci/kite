package testhelpers

import (
	"fmt"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
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
