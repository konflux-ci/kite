package http

import (
	"context"
	"encoding/json"
	"testing"

	net_http "net/http"
	net_httptest "net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/sirupsen/logrus"
)

// MockIssueService is a mock implementation for testing handlers
type MockIssueService struct {
	findIssueResults    *dto.IssueResponse
	findIssuesError     error
	findIssueByIDResult *models.Issue
	findIssueByIDError  error
	createIssueResult   *models.Issue
	createIssueError    error
	deleteIssueError    error
	updateIssueResult   *models.Issue
	updateIssueError    error
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

func (m *MockIssueService) CheckForDuplicateIssue(ctx context.Context, req dto.CreateIssueRequest) (*repository.DuplicateCheckResult, error) {
	return nil, nil
}

func (m *MockIssueService) ResolveIssuesByScope(ctx context.Context, resourceType, resourceName, namespace string) (int64, error) {
	return 0, nil
}

func (m *MockIssueService) AddRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	return nil
}

func (m *MockIssueService) RemoveRelatedIssue(ctx context.Context, sourceID, targetID string) error {
	return nil
}

// setTestHandler creates a test handler with mock service
func setupTestHandler(mockService *MockIssueService) *IssueHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return NewIssueHandler(mockService, logger)
}

// setupTestRouter creates a test router with HTTP tests
func setupTestRouter(handler *IssueHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Add routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/issues", handler.GetIssues)
		v1.POST("/issues", handler.CreateIssue)
		v1.GET("/issues/:id", handler.GetIssue)
		v1.PUT("/issues/:id", handler.UpdateIssue)
		v1.DELETE("/issues/:id", handler.DeleteIssue)
		v1.POST("/issues/:id/resolve", handler.ResolveIssue)
	}

	return router
}

func TestIssueHandler_GetIssues(t *testing.T) {
	mockIssues := []models.Issue{
		{
			ID:        "abc-1",
			Title:     "Test Issue 1",
			Namespace: "team-alpha",
			Severity:  models.SeverityMajor,
		},
		{
			ID:        "def-2",
			Title:     "Test Issue 2",
			Namespace: "team-alpha",
			Severity:  models.SeverityMajor,
		},
	}

	mockService := &MockIssueService{
		findIssueResults: &dto.IssueResponse{
			Data:   mockIssues,
			Total:  2,
			Limit:  50,
			Offset: 0,
		},
	}

	handler := setupTestHandler(mockService)
	router := setupTestRouter(handler)

	// Create test request
	req, err := net_http.NewRequest("GET", "/api/v1/issues?namespace=team-alpha", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	w := net_httptest.NewRecorder()

	// Serve
	router.ServeHTTP(w, req)

	// Check
	if w.Code != net_http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response body
	var response dto.IssueResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("expected 2 issues, got %d", len(response.Data))
	}

	if response.Total != 2 {
		t.Errorf("expected total 2, got %d", response.Total)
	}
}
