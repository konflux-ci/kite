package repository

import (
	"context"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/seed"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
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

	// Seed the DB with test data
	if err := seed.SeedData(db); err != nil {
		t.Fatalf("Failed to seed test database: %v", err)
	}

	return db
}

// createTestIssue is a helper function to create test issues
func createTestIssue(title, namespace string) dto.CreateIssueRequest {
	return dto.CreateIssueRequest{
		Title:       title,
		Description: "Test description",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   namespace,
		Scope: dto.ScopeReqBody{
			ResourceType:      "component",
			ResourceName:      "test-component",
			ResourceNamespace: namespace,
		},
		Links: []dto.CreateLinkRequest{
			{
				URL:   "konflux.test/pipelineruns/failure-xyz",
				Title: "Failed Pipeline Run: xyz",
			},
		},
	}
}

func TestIssueRepository_Create(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	logger := logrus.New()
	repo := NewIssueRepository(db, logger)
	ctx := context.Background()

	// Get initial count of DB
	var initialDBCount int64
	db.Model(&models.Issue{}).Count(&initialDBCount)

	// Test issue data
	req := createTestIssue("Test Issue", "test-namespace")

	// Create it
	issue, err := repo.Create(ctx, req)

	// Check
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	if issue == nil {
		t.Fatalf("Expected issue to be created, got nil")
	}

	if issue.Title != req.Title {
		t.Errorf("Expected title '%s', got '%s'", req.Title, issue.Title)
	}

	if issue.Namespace != req.Namespace {
		t.Errorf("Expected namespace '%s', got '%s'", req.Namespace, issue.Namespace)
	}

	// Confirm that issue was saved to the database
	var currentCount int64
	db.Model(&models.Issue{}).Count(&currentCount)
	expectedCount := initialDBCount + 1
	if currentCount != expectedCount {
		t.Errorf("Expected %d, got %d", expectedCount, currentCount)
	}
}
