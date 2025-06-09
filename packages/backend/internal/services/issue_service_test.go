package services

import (
	"context"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/konflux-ci/kite/internal/testhelpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// setupServiceDependents sets up the dependent components used in the IssueService
func setupServiceDependents(t *testing.T) (context.Context, *logrus.Logger, repository.IssueRepository, *gorm.DB) {
	db := testhelpers.SetupTestDB(t)
	logger := logrus.New()
	repo := repository.NewIssueRepository(db, logger)
	ctx := context.Background()

	return ctx, logger, repo, db
}

func createTestService(t *testing.T) (*IssueService, context.Context, *gorm.DB) {
	ctx, logger, repo, db := setupServiceDependents(t)
	return NewIssueService(repo, logger), ctx, db
}

func TestIssueService_CreateIssue(t *testing.T) {
	service, ctx, _ := createTestService(t)

	req := dto.CreateIssueRequest{
		Title:       "Test Service Issue",
		Description: "Testing service layer",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "test-service-namespace",
		Scope: dto.ScopeReqBody{
			ResourceType:      "component",
			ResourceName:      "test-component",
			ResourceNamespace: "test-service-namespace",
		},
	}

	issue, err := service.CreateIssue(ctx, req)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if issue == nil {
		t.Fatal("Expected issue to be created, got nil")
	}

	if issue.Title != req.Title {
		t.Errorf("Expected title '%s', got '%s'", req.Title, issue.Title)
	}

	if issue.Namespace != req.Namespace {
		t.Errorf("Expected namespace '%s', got '%s'", req.Namespace, issue.Namespace)
	}
}

func TestIssueService_FindIssuesByID(t *testing.T) {
	service, ctx, db := createTestService(t)

	req := dto.CreateIssueRequest{
		Title:       "Test Service Service Find By ID",
		Description: "Testing service layer",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "test-service-namespace",
		Scope: dto.ScopeReqBody{
			ResourceType:      "component",
			ResourceName:      "test-component",
			ResourceNamespace: "test-service-namespace",
		},
	}

	newIssue, err := service.CreateIssue(ctx, req)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if newIssue == nil {
		t.Fatal("Expected issue to be created, got nil")
	}

	var latestIssue models.Issue
	db.Last(&latestIssue)

	foundIssue, err := service.FindIssueByID(ctx, latestIssue.ID)

	if foundIssue == nil {
		t.Fatal("Expected to find issue, got nil")
	}

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if foundIssue.ID != newIssue.ID {
		t.Errorf("Expected ID '%s', got '%s'", newIssue.ID, foundIssue.ID)
	}
}
