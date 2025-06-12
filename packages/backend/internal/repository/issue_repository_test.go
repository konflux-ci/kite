package repository

import (
	"context"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/testhelpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// setupTestScenario sets up a context and repository for test scenarios
func setupTestScenario(t *testing.T) (context.Context, *gorm.DB, IssueRepository) {
	db := testhelpers.SetupTestDB(t)
	logger := logrus.New()
	repo := NewIssueRepository(db, logger)
	ctx := context.Background()

	return ctx, db, repo
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
	ctx, db, repo := setupTestScenario(t)

	// Test issue data
	req := createTestIssue("Test Issue", "test-namespace")

	// Create it
	issue, err := repo.Create(ctx, req)

	// Check
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}

	err = testhelpers.CompareIssueToDTO(*issue, req)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	// Confirm that issue was saved to the database
	var currentCount int64
	db.Model(&models.Issue{}).Count(&currentCount)
	if currentCount != 1 {
		t.Errorf("Expected 1 issue in DB, got %d", currentCount)
	}
}

func TestIssueRepository_FindByID(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t)

	// Create a test issue first
	req := createTestIssue("Find Test Issue", "test-namespace")
	createdIssue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if createdIssue == nil {
		t.Fatalf("Expected issue to be created, got nil")
	}

	// Find the issue
	foundIssue, err := repo.FindByID(ctx, createdIssue.ID)
	if err != nil {
		t.Fatalf("unexpected error, got: %v", err)
	}
	if foundIssue == nil {
		t.Fatalf("Expected issue to be found, got nil")
	}

	// Verify
	err = testhelpers.CompareIssues(*createdIssue, *foundIssue)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}
}

func TestIssueRepository_FindByID_NotFound(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t)
	// Try to find non-existent issue
	foundIssue, err := repo.FindByID(ctx, "does-not-exist")

	// Verify
	if err != nil {
		t.Fatalf("Expected no error for non-existent issue, got %v", err)
	}

	if foundIssue != nil {
		t.Errorf("Expected nil for non-existent issue, got an issue")
	}
}

func TestIssueRepository_FindAll_WithFilters(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t)

	// Create test issues
	issues := []dto.CreateIssueRequest{
		createTestIssue("Build Issue", "team-test"),
		{
			Title:       "Test Issue",
			Description: "Test Description",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeTest,
			Namespace:   "team-test",
			Scope: dto.ScopeReqBody{
				ResourceType:      "component",
				ResourceName:      "test-component",
				ResourceNamespace: "team-test",
			},
		},
		createTestIssue("Release Issue", "team-beta"),
	}

	// Write issues to DB
	for _, req := range issues {
		_, err := repo.Create(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create test issue: %v", err)
		}
	}

	// Check: Find all issues in team-alpha
	filters := IssueQueryFilters{
		Namespace: "team-test",
		Limit:     10,
	}

	foundIssues, total, err := repo.FindAll(ctx, filters)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 issues in team-test, got %d", total)
	}

	if len(foundIssues) != 2 {
		t.Errorf("Expected 2 issues returned, got %d", len(foundIssues))
	}

	// Check that all returned issues belong to team-test
	for _, issue := range foundIssues {
		if issue.Namespace != "team-test" {
			t.Errorf("Expected namespace 'team-test', got '%s'", issue.Namespace)
		}
	}
}

func TestIssueRepository_CheckDuplicate(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t)

	// Create an issue
	req := createTestIssue("Duplicate Test", "test-namespace")
	_, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Check for duplicates with the same properties
	result, err := repo.CheckDuplicate(ctx, req)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	if !result.IsDuplicate {
		t.Error("Expected issue to be a duplicate")
	}

	if result.ExistingIssue == nil {
		t.Error("Expected existing issue to be returned")
	}
}

func TestIssueRepository_Update(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t)

	// Create an issue
	req := createTestIssue("Some Issue", "test-namespace")
	issue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Get latest issue
	expectedID := issue.ID
	expectedTitle := "Updated Issue"

	updatedIssueReq := dto.UpdateIssueRequest{
		Title: &expectedTitle,
	}
	// Update
	updatedIssue, err := repo.Update(ctx, expectedID, updatedIssueReq)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	if updatedIssue == nil {
		t.Fatal("Expected issue to be returned")
	}

	if updatedIssue.ID != expectedID {
		t.Errorf("Wrong issue returned, got issue with ID %s, expected %s", updatedIssue.ID, expectedID)
	}

	if updatedIssue.Title != expectedTitle {
		t.Errorf("Wrong title, got '%s', expected '%s'", updatedIssue.Title, expectedTitle)
	}
}

func TestIssueRepository_Delete(t *testing.T) {
	ctx, db, repo := setupTestScenario(t)

	// Create issue with links
	req := createTestIssue("Delete Test", "test-namespace")
	req.Links = append(req.Links,
		dto.CreateLinkRequest{
			Title: "Delete Test Link",
			URL:   "https://konflux.test/some-link",
		},
	)

	createdIssue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Verify issue and link exists
	var issueCount, linkCount int64
	db.Model(&models.Issue{}).Count(&issueCount)
	db.Model(&models.Link{}).Count(&linkCount)

	if issueCount != 1 {
		t.Errorf("Expected 1 issue before delete, got %d", issueCount)
	}

	if linkCount != 2 {
		t.Errorf("Expected 2 links before delete, got %d", linkCount)
	}

	// Delete the issue
	err = repo.Delete(ctx, createdIssue.ID)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Update variables after deletion
	db.Model(&models.Issue{}).Count(&issueCount)
	db.Model(&models.Link{}).Count(&linkCount)

	if issueCount != 0 {
		t.Errorf("Expected 0 issues after delete, got %d", issueCount)
	}

	if linkCount != 0 {
		t.Errorf("Expected 0 links after delete, got %d", linkCount)
	}
}
