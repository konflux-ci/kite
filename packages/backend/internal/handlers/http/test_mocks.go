package http

import (
	"context"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
)

// MockIssueService is a mock implementation for testing handlers
type MockIssueService struct {
	findIssueResults              *dto.IssueResponse
	findIssuesError               error
	findIssueByIDResult           *models.Issue
	findIssueByIDError            error
	createIssueResult             *models.Issue
	createIssueError              error
	deleteIssueError              error
	updateIssueResult             *models.Issue
	updateIssueError              error
	findDuplicateIssueResult      *models.Issue
	findDuplicateIssueResultError error
	resolveIssuesByScopeResult    int64
	resolveIssuesByScopeError     error
	createOrUpdateIssueResult     *models.Issue
	createOrUpdateIssueError      error
}

func (m *MockIssueService) FindIssues(ctx context.Context, filters repository.IssueQueryFilters) (*dto.IssueResponse, error) {
	return m.findIssueResults, m.findIssuesError
}

func (m *MockIssueService) FindIssueByID(ctx context.Context, id string) (*models.Issue, error) {
	return m.findIssueByIDResult, m.findIssueByIDError
}

func (m *MockIssueService) CreateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	return m.createIssueResult, m.createIssueError
}

func (m *MockIssueService) UpdateIssue(ctx context.Context, id string, req dto.UpdateIssueRequest) (*models.Issue, error) {
	return m.updateIssueResult, m.updateIssueError
}

func (m *MockIssueService) DeleteIssue(ctx context.Context, id string) error {
	return m.deleteIssueError
}

func (m *MockIssueService) FindDuplicateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	return m.findDuplicateIssueResult, m.findDuplicateIssueResultError
}

func (m *MockIssueService) CreateOrUpdateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	return m.createOrUpdateIssueResult, m.findDuplicateIssueResultError
}

func (m *MockIssueService) ResolveIssuesByScope(ctx context.Context, resourceType, resourceName, namespace string) (int64, error) {
	return m.resolveIssuesByScopeResult, m.resolveIssuesByScopeError
}

func (m *MockIssueService) AddRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	return nil
}

func (m *MockIssueService) RemoveRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	return nil
}
