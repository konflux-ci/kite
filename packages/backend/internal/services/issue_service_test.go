package services

import (
	"context"
	"testing"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/konflux-ci/kite/internal/testingtools"
	"github.com/sirupsen/logrus"
)

// setupServiceDependents sets up the dependent components used in the IssueService
func setupServiceDependents(t *testing.T) (context.Context, *logrus.Logger, repository.IssueRepository) {
	db := testingtools.SetupTestDB(t)
	logger := logrus.New()
	repo := repository.NewIssueRepository(db, logger)
	ctx := context.Background()

	return ctx, logger, repo
}

func createTestService(t *testing.T) (*IssueService, context.Context) {
	ctx, logger, repo := setupServiceDependents(t)
	return NewIssueService(repo, logger), ctx
}

func TestIssueService_CreateIssue(t *testing.T) {
	service, ctx := createTestService(t)

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
