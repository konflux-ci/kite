package services

import (
	"context"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
)

// IssueServiceInterface defines what an issue service should do
// This allows us to mock it for testing
type IssueServiceInterface interface {
	FindIssues(ctx context.Context, filters repository.IssueQueryFilters) (*dto.IssueResponse, error)
	FindIssueByID(ctx context.Context, id string) (*models.Issue, error)
	CreateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error)
	UpdateIssue(ctx context.Context, id string, req dto.UpdateIssueRequest) (*models.Issue, error)
	DeleteIssue(ctx context.Context, id string) error
	CheckForDuplicateIssue(ctx context.Context, req dto.CreateIssueRequest) (*repository.DuplicateCheckResult, error)
	ResolveIssuesByScope(ctx context.Context, resourceType, resourceName, namespace string) (int64, error)
	AddRelatedIssue(ctx context.Context, sourceID, targetID string) error
	RemoveRelatedIssue(ctx context.Context, sourceID, targetID string) error
}

// Compile-time interface check to verify that IssueService implements the interface
var _ IssueServiceInterface = (*IssueService)(nil)
