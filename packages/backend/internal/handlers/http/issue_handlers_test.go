package http

import (
	"bytes"
	"encoding/json"
	"testing"

	net_http "net/http"
	net_httptest "net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/sirupsen/logrus"
)

// setTestIssueHandler creates a test handler with mock service
func setupTestIssueHandler(mockService *MockIssueService) *IssueHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return NewIssueHandler(mockService, logger)
}

// setupTestIssueRouter creates a test router with HTTP tests
func setupTestIssueRouter(handler *IssueHandler) *gin.Engine {
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

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

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

func TestIssueHandler_GetIssue_Found(t *testing.T) {
	mockIssue := &models.Issue{
		ID:        "test-issue-abc",
		Title:     "Test Issue",
		Namespace: "team-alpha",
		Severity:  models.SeverityMajor,
	}

	mockService := &MockIssueService{
		findIssueByIDResult: mockIssue,
	}

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	// Create request
	req, err := net_http.NewRequest("GET", "/api/v1/issues/test-issue-abc", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.Issue
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.ID != mockIssue.ID {
		t.Errorf("Expected ID '%s', got '%s'", mockIssue.ID, response.ID)
	}

	if response.Title != mockIssue.Title {
		t.Errorf("expected title '%s', got '%s'", mockIssue.Title, response.Title)
	}
}

func TestIssueHandler_GetIssue_NotFound(t *testing.T) {
	mockService := &MockIssueService{
		findIssueByIDResult: nil,
	}

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	req, err := net_http.NewRequest("GET", "/api/v1/issues/do-not-exist-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedErrorMessage := "Issue not found"
	if response["error"] != expectedErrorMessage {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMessage, response["error"])
	}
}

func TestIssueHandler_CreateIssue_Success(t *testing.T) {
	createRequest := dto.CreateIssueRequest{
		Title:       "New Test Issue",
		Description: "This is a test issue",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "team-gamma",
		Scope: dto.ScopeReqBody{
			ResourceType:      "component",
			ResourceName:      "test-component",
			ResourceNamespace: "team-gamma",
		},
	}

	createdIssue := &models.Issue{
		ID:          "new-issue-abc",
		Title:       createRequest.Title,
		Description: createRequest.Description,
		Severity:    createRequest.Severity,
		IssueType:   createRequest.IssueType,
		Namespace:   createRequest.Namespace,
	}

	mockService := &MockIssueService{
		createIssueResult: createdIssue,
	}

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	// Create request body
	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request,: %v", err)
	}

	req, err := net_http.NewRequest("POST", "/api/v1/issues", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response models.Issue
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Title != createRequest.Title {
		t.Errorf("expected title '%s', got '%s'", createRequest.Title, response.Title)
	}

	if response.ID != createdIssue.ID {
		t.Errorf("expected ID '%s', got '%s'", createdIssue.ID, response.ID)
	}
}

func TestIssueHandler_CreateIssue_InvalidRequest(t *testing.T) {
	mockService := &MockIssueService{}
	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	// Create invalid request with missing required fields
	invalidRequest := map[string]interface{}{
		"title": "Test Issue",
	}

	reqBody, err := json.Marshal(invalidRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := net_http.NewRequest("POST", "/api/v1/issues", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error message in response")
	}
}

func TestIssueHandler_DeleteIssue_Success(t *testing.T) {
	mockIssue := &models.Issue{
		ID:        "delete-test-abc",
		Title:     "Issue for deletion",
		Namespace: "team-deleted",
	}

	mockService := &MockIssueService{
		findIssueByIDResult: mockIssue,
		deleteIssueError:    nil,
	}

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	req, err := net_http.NewRequest("DELETE", "/api/v1/issues/delete-test-abc", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Body should be empty for 204 No Content
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %s", w.Body.String())
	}
}

func TestIssueHandler_ResolveIssue(t *testing.T) {
	originalIssue := &models.Issue{
		ID:        "resolve-test-abc",
		Title:     "Issue to Resolve",
		State:     models.IssueStateActive,
		Namespace: "team-resolved",
	}

	resolvedIssue := &models.Issue{
		ID:        "resolve-test-abc",
		Title:     "Issue to Resolve",
		State:     models.IssueStateResolved,
		Namespace: "team-resolved",
	}

	mockService := &MockIssueService{
		findIssueByIDResult: originalIssue,
		updateIssueResult:   resolvedIssue,
	}

	handler := setupTestIssueHandler(mockService)
	router := setupTestIssueRouter(handler)

	req, err := net_http.NewRequest("POST", "/api/v1/issues/resolve-test-abc/resolve", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.Issue
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.State != models.IssueStateResolved {
		t.Errorf("expeted state 'RESOLVED', got '%s'", response.State)
	}
}
