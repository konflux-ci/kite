package repository

import (
	"context"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
)

type IssueRepository interface {
	Create(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error)
	FindByID(ctx context.Context, id string) (*models.Issue, error)
	Update(ctx context.Context, id string, updates dto.UpdateIssueRequest) (*models.Issue, error)
	Delete(ctx context.Context, id string) error
	// TODO - move IssueQueryFilters somewhere else
	FindAll(ctx context.Context, filters IssueQueryFilters) ([]models.Issue, int64, error)
	CheckDuplicate(ctx context.Context, req dto.CreateIssueRequest) (*DuplicateCheckResult, error)
	ResolveByScope(ctx context.Context, resourceType, resourceName, namespace string) (int64, error)
	AddRelatedIssue(ctx context.Context, sourceID, targetID string) error
	RemoveRelatedIssue(ctx context.Context, sourceID, targetID string) error
}

type LinkRepository interface {
	CreateBatch(ctx context.Context, issueID string, links []models.Link) error
	DeleteByIssueID(ctx context.Context, issueID string) error
}
