package services

import (
	"context"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/sirupsen/logrus"
)

type IssueService struct {
	repo   repository.IssueRepository // Repository instance
	logger *logrus.Logger             // Logging instance
}

type IssueQueryFilters struct {
	Namespace    string
	Severity     *models.Severity
	IssueType    *models.IssueType
	State        *models.IssueState
	ResourceType string
	ResourceName string
	Search       string
	Limit        int
	Offset       int
}

type DuplicateCheckResult struct {
	IsDuplicate   bool
	ExistingIssue *models.Issue
}

func NewIssueService(repo repository.IssueRepository, logger *logrus.Logger) *IssueService {
	return &IssueService{
		repo:   repo,
		logger: logger,
	}
}

// CheckForDuplicateIssue checks if a similar issue already exists
func (s *IssueService) FindDuplicateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	issueFound, err := s.repo.FindDuplicate(ctx, req)
	if err != nil {
		return nil, err
	}
	return issueFound, nil
}

// CreateOrUpdateIssue creates an issue if a duplicate is not found and updates the record if it is.
//
// NOTE: This method is mainly used for webhook endpoints.
func (s *IssueService) CreateOrUpdateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	issue, err := s.repo.CreateOrUpdate(ctx, req)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// FindIssues retrieves issues with optional filters
func (s *IssueService) FindIssues(ctx context.Context, filters repository.IssueQueryFilters) (*dto.IssueResponse, error) {
	issues, total, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &dto.IssueResponse{
		Data:   issues,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}

// FindIssueByID retrieves a single issue by ID
func (s *IssueService) FindIssueByID(ctx context.Context, id string) (*models.Issue, error) {
	issue, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// CreateIssue creates a new issue if a duplicate is not found and updates the record if it is.
func (s *IssueService) CreateIssue(ctx context.Context, req dto.CreateIssueRequest) (*models.Issue, error) {
	issue, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// UpdateIssue updates and existing issue
func (s *IssueService) UpdateIssue(ctx context.Context, id string, req dto.UpdateIssueRequest) (*models.Issue, error) {
	issue, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// DeleteIssue deletes an issue and related entities
func (s *IssueService) DeleteIssue(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// AddRelatedIsue creates a relationship between two issues
func (s *IssueService) AddRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	if err := s.repo.AddRelatedIssue(ctx, sourceID, targetID); err != nil {
		return err
	}
	return nil
}

// RemoveRelatedIssue removes a relationship between issues
func (s *IssueService) RemoveRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	if err := s.repo.RemoveRelatedIssue(ctx, sourceID, targetID); err != nil {
		return err
	}
	return nil
}

// ResolveIssuesByScope resolves all active issues for a given scope
func (s *IssueService) ResolveIssuesByScope(ctx context.Context, resourceType, resourceName, namespace string) (int64, error) {
	count, err := s.repo.ResolveByScope(ctx, resourceType, resourceName, namespace)
	if err != nil {
		return 0, nil
	}
	return count, nil
}
