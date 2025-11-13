package http

import (
	"bytes"
	"encoding/json"
	"testing"

	net_http "net/http"
	net_httptest "net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/testhelpers"
	"github.com/sirupsen/logrus"
)

func setupTestWebhookHandler(mockService *MockIssueService) *WebhookHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return NewWebhookHandler(mockService, logger)
}

func setupTestWebhookRouter(handler *WebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	v1 := router.Group("/webhooks")
	{
		v1.POST("/pipeline-failure", handler.PipelineFailure)
		v1.POST("/pipeline-success", handler.PipelineSuccess)
		v1.POST("/release-failure", handler.ReleaseFailure)
		v1.POST("/release-success", handler.ReleaseSuccess)
	}

	return router
}

func TestWebhookHandler_PipelineFailure(t *testing.T) {
	// What gets sent to the webhook endpoint
	pipelineFailureRequest := PipelineFailureRequest{
		PipelineName:  "pipeline-xyz",
		Namespace:     "team-failed-pr",
		FailureReason: "task run timed out",
		RunID:         "pipeline-xyz-123",
	}

	// Expected issue created
	expectedIssue := &models.Issue{
		Title:       "Pipeline run failed: pipeline-xyz",
		Description: "The pipeline run pipeline-xyz failed with reason: task run timed out",
		Severity:    models.SeverityMajor,
		Namespace:   "team-failed-pr",
		Scope: models.IssueScope{
			ResourceType:      "pipelinerun",
			ResourceName:      "pipeline-xyz",
			ResourceNamespace: "team-failed-pr",
		},
		Links: []models.Link{
			{
				Title: "Pipeline Run Logs",
				URL:   "https://cluster.dev/pipelineruns/pipeline-xyz-123/logs",
			},
		},
	}

	mockService := &MockIssueService{
		// This should not be a duplicate
		findDuplicateIssueResult:      nil,
		findDuplicateIssueResultError: nil,
		// Issue should get created without any...issues.
		createOrUpdateIssueResult: expectedIssue,
		createOrUpdateIssueError:  nil,
	}

	handler := setupTestWebhookHandler(mockService)
	router := setupTestWebhookRouter(handler)

	// Create request body
	reqBody, err := json.Marshal(pipelineFailureRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Make request
	req, err := net_http.NewRequest("POST", "/webhooks/pipeline-failure", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedStatus := "success"
	if response["status"] != expectedStatus {
		t.Errorf("expected response with status '%s', got '%s'", expectedStatus, response["status"])
	}

	// Convert response data to JSON
	issueData, err := json.Marshal(response["issue"])
	if err != nil {
		t.Fatalf("Failed to marshal issue data: %v", err)
	}

	// Convert JSON to struct
	var createdIssue models.Issue
	err = json.Unmarshal(issueData, &createdIssue)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Compare the issue created to the expected issue
	err = testhelpers.CompareIssues(createdIssue, *expectedIssue)
	if err != nil {
		t.Errorf("issue comparison failed: %v", err)
	}
}

func TestWebhookHandler_PipelineSuccess(t *testing.T) {
	// What gets sent to the webhook endpoint
	pipelineSuccessRequest := PipelineSuccessRequest{
		PipelineName: "pipeline-xyz",
		Namespace:    "team-failed-pr",
	}

	// Mock service results
	mockService := &MockIssueService{
		resolveIssuesByScopeResult: 2,
		resolveIssuesByScopeError:  nil,
	}

	// Setup
	handler := setupTestWebhookHandler(mockService)
	router := setupTestWebhookRouter(handler)

	// Create request body
	reqBody, err := json.Marshal(pipelineSuccessRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Make request
	req, err := net_http.NewRequest("POST", "/webhooks/pipeline-success", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Extract response onto map
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check status of response
	expectedStatus := "success"
	if response["status"] != expectedStatus {
		t.Errorf("expected response with status '%s', got '%s'", expectedStatus, response["status"])
	}

	// Check message in response
	expectedMessage := "Resolved 2 issue(s) for pipeline pipeline-xyz"
	if response["message"] != expectedMessage {
		t.Errorf("expected response with message '%s', got '%s'", expectedMessage, response["message"])
	}
}

func TestWebhookHandler_ReleaseFailure(t *testing.T) {
	// What gets sent to the webhook endpoint
	releaseFailureRequest := ReleaseFailureRequest{
		Application:    "fancy-app",
		Namespace:      "team-failed-release",
		FailurePhase:   "ManagedProcessing",
		ReleaseName:    "release-to-prod-123",
		PipelineRunURL: "logs.com/managed-123",
	}

	// Expected issue created
	expectedIssue := &models.Issue{
		Title:       "Release release-to-prod-123 failed for application fancy-app",
		Description: "The release failed in phase: ManagedProcessing. Link to logs: logs.com/managed-123",
		Severity:    models.SeverityMajor,
		Namespace:   "team-failed-release",
		Scope: models.IssueScope{
			ResourceType:      "application",
			ResourceName:      "fancy-app",
			ResourceNamespace: "team-failed-release",
		},
	}

	mockService := &MockIssueService{
		// This should not be a duplicate
		findDuplicateIssueResult:      nil,
		findDuplicateIssueResultError: nil,
		// Issue should get created without any...issues.
		createOrUpdateIssueResult: expectedIssue,
		createOrUpdateIssueError:  nil,
	}

	handler := setupTestWebhookHandler(mockService)
	router := setupTestWebhookRouter(handler)

	// Create request body
	reqBody, err := json.Marshal(releaseFailureRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Make request
	req, err := net_http.NewRequest("POST", "/webhooks/release-failure", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedStatus := "success"
	if response["status"] != expectedStatus {
		t.Errorf("expected response with status '%s', got '%s'", expectedStatus, response["status"])
	}

	// Convert response data to JSON
	issueData, err := json.Marshal(response["issue"])
	if err != nil {
		t.Fatalf("Failed to marshal issue data: %v", err)
	}

	// Convert JSON to struct
	var createdIssue models.Issue
	err = json.Unmarshal(issueData, &createdIssue)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Compare the issue created to the expected issue
	err = testhelpers.CompareIssues(createdIssue, *expectedIssue)
	if err != nil {
		t.Errorf("issue comparison failed: %v", err)
	}
}

func TestWebhookHandler_ReleaseSuccess(t *testing.T) {
	// What gets sent to the webhook endpoint
	releaseSuccessRequest := ReleaseSuccessRequest{
		Application: "fancy-app",
		Namespace:   "team-failed-release",
	}

	// Mock service results
	mockService := &MockIssueService{
		resolveIssuesByScopeResult: 2,
		resolveIssuesByScopeError:  nil,
	}

	// Setup
	handler := setupTestWebhookHandler(mockService)
	router := setupTestWebhookRouter(handler)

	// Create request body
	reqBody, err := json.Marshal(releaseSuccessRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Make request
	req, err := net_http.NewRequest("POST", "/webhooks/release-success", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := net_httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != net_http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Extract response onto map
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check status of response
	expectedStatus := "success"
	if response["status"] != expectedStatus {
		t.Errorf("expected response with status '%s', got '%s'", expectedStatus, response["status"])
	}

	// Check message in response
	expectedMessage := "Resolved 2 issue(s) for application fancy-app"
	if response["message"] != expectedMessage {
		t.Errorf("expected response with message '%s', got '%s'", expectedMessage, response["message"])
	}
}
