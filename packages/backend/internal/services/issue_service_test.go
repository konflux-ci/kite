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

	err = testhelpers.CompareIssueToDTO(*issue, req)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
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

	err = testhelpers.CompareIssues(*foundIssue, *newIssue)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}
}

func TestIssueService_FindIssue_WithFilters(t *testing.T) {
	// Setup
	service, ctx, _ := createTestService(t)

	req := []dto.CreateIssueRequest{
		{
			Title:       "Test Issue A",
			Description: "Testing service filters",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypeBuild,
			Namespace:   "test-service-namespace",
			Scope: dto.ScopeReqBody{
				ResourceType:      "component",
				ResourceName:      "test-component",
				ResourceNamespace: "test-service-namespace",
			},
		},
		{
			Title:       "Test Issue B",
			Description: "Testing dependency issue",
			Severity:    models.SeverityInfo,
			IssueType:   models.IssueTypeDependency,
			Namespace:   "team-gamma",
			Scope: dto.ScopeReqBody{
				ResourceType:      "dependency",
				ResourceName:      "test-dependency",
				ResourceNamespace: "team-gamma",
			},
		},
		{
			Title:       "Test Issue C",
			Description: "Testing release issue",
			Severity:    models.SeverityInfo,
			IssueType:   models.IssueTypeDependency,
			Namespace:   "team-alpha",
			Scope: dto.ScopeReqBody{
				ResourceType:      "release",
				ResourceName:      "release-xyz",
				ResourceNamespace: "team-alpha",
			},
		},
	}

	for _, issueReq := range req {
		_, err := service.CreateIssue(ctx, issueReq)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Simple namespace filter
	response, err := service.FindIssues(ctx, repository.IssueQueryFilters{
		Namespace: "team-alpha",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 issue returned, got %d", len(response.Data))
	}

	firstIssueFound := response.Data[0]
	expectedIssue := req[2]
	err = testhelpers.CompareIssueToDTO(firstIssueFound, expectedIssue)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// More filters
	response, err = service.FindIssues(ctx, repository.IssueQueryFilters{
		Namespace:    "team-gamma",
		Search:       "Test Issue B",
		ResourceType: "dependency",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 issue returned, got %d", len(response.Data))
	}

	secondFoundIssue := response.Data[0]
	expectedIssue = req[1]
	err = testhelpers.CompareIssueToDTO(secondFoundIssue, expectedIssue)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	// Should return no results
	response, err = service.FindIssues(ctx, repository.IssueQueryFilters{
		Namespace:    "void",
		Search:       "should not exist",
		ResourceType: "non-existent",
	})

	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected 0 issues returned, got %d", len(response.Data))
	}

	severity := models.SeverityMajor
	issueType := models.IssueTypeBuild
	response, err = service.FindIssues(ctx, repository.IssueQueryFilters{
		Severity:  &severity,
		IssueType: &issueType,
	})

	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 issue returned, got %d", len(response.Data))
	}

	thirdFoundIssue := response.Data[0]
	expectedIssue = req[0]
	err = testhelpers.CompareIssueToDTO(thirdFoundIssue, expectedIssue)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
	}
}

func TestIssueService_ResolveIssuesByScope(t *testing.T) {
	// Setup
	service, ctx, _ := createTestService(t)
	req := []dto.CreateIssueRequest{
		{
			Title:       "Test Issue A",
			Description: "Testing service filters",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypeBuild,
			Namespace:   "team-gamma",
			Scope: dto.ScopeReqBody{
				ResourceType:      "component",
				ResourceName:      "test-component",
				ResourceNamespace: "team-gamma",
			},
		},
		{
			Title:       "Test Issue B",
			Description: "Testing dependency issue",
			Severity:    models.SeverityInfo,
			IssueType:   models.IssueTypeDependency,
			Namespace:   "team-gamma",
			Scope: dto.ScopeReqBody{
				ResourceType:      "component",
				ResourceName:      "test-component",
				ResourceNamespace: "team-gamma",
			},
		},
		{
			Title:       "Test Issue C",
			Description: "Testing release issue",
			Severity:    models.SeverityInfo,
			IssueType:   models.IssueTypeDependency,
			Namespace:   "team-alpha",
			Scope: dto.ScopeReqBody{
				ResourceType:      "release",
				ResourceName:      "release-xyz",
				ResourceNamespace: "team-alpha",
			},
		},
	}
	// Create test issues
	for _, issueReq := range req {
		_, err := service.CreateIssue(ctx, issueReq)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Should resolve two issues
	count, err := service.ResolveIssuesByScope(ctx, "component", "test-component", "team-gamma")
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 issues resolved, got %d", count)
	}

	// Should resolve 1 issue
	count, err = service.ResolveIssuesByScope(ctx, "release", "release-xyz", "team-alpha")
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 issue resolved, got %d", count)
	}

	// Should resolve non, returning 0
	count, err = service.ResolveIssuesByScope(ctx, "void", "void", "void")
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 issues resolves, got %d", count)
	}
}

func TestIssueService_CheckForDuplicates(t *testing.T) {
	// Setup
	service, ctx, _ := createTestService(t)
	req := dto.CreateIssueRequest{
		Title:       "Test Issue",
		Description: "Testing issue duplication check",
		Severity:    models.SeverityInfo,
		IssueType:   models.IssueTypeDependency,
		Namespace:   "team-alpha",
		Scope: dto.ScopeReqBody{
			ResourceType:      "release",
			ResourceName:      "release-xyz",
			ResourceNamespace: "team-alpha",
		},
	}
	issue, err := service.CreateIssue(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}

	err = testhelpers.CompareIssueToDTO(*issue, req)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	result, err := service.CheckForDuplicateIssue(ctx, req)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	if !result.IsDuplicate {
		t.Error("expected duplicate to be found")
	}

	if result.ExistingIssue == nil {
		t.Error("expected existing issue to be found")
	}

	if result.ExistingIssue.ID != issue.ID {
		t.Errorf("expected issue with id '%s', got '%s'", result.ExistingIssue.ID, issue.ID)
	}
}
