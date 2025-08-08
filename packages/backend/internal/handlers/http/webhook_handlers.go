package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/config"
	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/services"
	"github.com/sirupsen/logrus"
)

// WebhookHandler handles incoming webhook requests for pipeline events.
type WebhookHandler struct {
	issueService services.IssueServiceInterface // Issue service for managing issues
	logger       *logrus.Logger                 // Logger for structured logging
}

// NewWebhookHandler returns a new handler for the webhooks router
func NewWebhookHandler(issueService services.IssueServiceInterface, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		issueService: issueService,
		logger:       logger,
	}
}

// PipelineFailureRequest represents the payload for a pipeline failure webhook.
//
// Fields:
//   - pipelineName:  (string, required) - Name of the failed pipeline.
//   - namespace:     (string, required) - Kubernetes namespace where the pipeline ran.
//   - failureReason: (string, required) - Why the pipeline failed. (required)
//   - severity:      (string. optional, - defaults to "major") Issue severity.
//   - runId:         (string, optional) - Pipeline run identifier.
//   - logsUrl:       (string, optional) - Direct URL to logs.
type PipelineFailureRequest struct {
	PipelineName  string `json:"pipelineName" binding:"required"`
	Namespace     string `json:"namespace" binding:"required"`
	Severity      string `json:"severity"`
	FailureReason string `json:"failureReason" binding:"required"`
	RunID         string `json:"runId"`
	LogsURL       string `json:"logsUrl"`
}

// PipelineSuccessRequest represents the payload for a pipeline success webhook.
//
// Fields:
//   - pipelineName: (string, required) - Name of the successful pipeline.
//   - namespace:    (string, required) - Kubernetes namespace where the pipeline ran.
type PipelineSuccessRequest struct {
	PipelineName string `json:"pipelineName" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
}

// PipelineFailure handles pipeline failure webhooks with idempotent behavior.
// If the same issue payload is sent multiple times, only one issue will be created or updated.
//
// Request Body:
//   - pipelineName:   (string, required) - Name of the failed pipeline.
//   - namespace:      (string, required) - Namespace where the pipeline ran.
//   - failureReason:  (string, required) - Description of why the pipeline failed.
//   - severity:       (string, optional, default: "major") - Issue severity level.
//   - runId:          (string, optional) - Pipeline run identifier for log URLs.
//   - logsUrl:        (string, optional) - Direct URL to logs. Generated if omitted.
//
// Response:
//   - 201 Created: Issue was created or updated successfully
//   - 400 Bad Request: Missing required fields
//   - 500 Internal Server Error: Database or processing error
//
// Example:
//
//	 POST /api/v1/webhooks/pipeline-failure
//	 Content-Type: application/json
//		{
//		  "pipelineName": "frontend-build-xyz",
//		  "namespace": "team-alpha",
//		  "failureReason": "Docker build failed"
//		}
func (h *WebhookHandler) PipelineFailure(c *gin.Context) {
	var req PipelineFailureRequest
	// Check if the request binds to proper JSON, in the format specified
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields", "details": err.Error()})
		return
	}

	// Format issue data
	logsURL := req.LogsURL
	if logsURL == "" {
		baseURL := config.GetEnvOrDefault("KITE_CLUSTER_URL", "https://konflux.dev")
		logsEndpoint := config.GetEnvOrDefault("KITE_LOGS_ENDPOINT", "/logs/pipelineruns/")
		logsURL = fmt.Sprintf("%s%s%s", baseURL, logsEndpoint, req.RunID)
	}

	severity := models.SeverityMajor
	if req.Severity != "" {
		severity = models.Severity(req.Severity)
	}

	issueData := dto.CreateIssueRequest{
		Title:       fmt.Sprintf("Pipeline run failed: %s", req.PipelineName),
		Description: fmt.Sprintf("The pipeline run %s failed with reason: %s", req.PipelineName, req.FailureReason),
		Severity:    severity,
		IssueType:   models.IssueTypePipeline,
		Namespace:   req.Namespace,
		Scope: dto.ScopeReqBody{
			ResourceType:      "pipelinerun",
			ResourceName:      req.PipelineName,
			ResourceNamespace: req.Namespace,
		},
		Links: []dto.CreateLinkRequest{
			{
				Title: "Pipeline Run Logs",
				URL:   logsURL,
			},
		},
	}

	// Create or update the issue
	issue, err := h.issueService.CreateOrUpdateIssue(c, issueData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create or update pipeline issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	h.logger.WithField("issue_id", issue.ID).Info("Processed pipeline failure webhook")

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"issue":  issue,
	})
}

// PipelineSuccess handles pipeline success webhooks.
//
// Request Body:
//   - pipelineName: (string, required) - Name of the successful pipeline
//   - namespace:    (string, required) -  Namespace where the pipeline ran
//
// Response:
//   - 200 OK: Issues related to the pipeline are resolved
//   - 400 Bad Request: Missing required fields
//   - 500 Internal Server Error: Database or processing error
//
// Issues that match the pipeline name and namespace will be marked as resolved using
// the scope:
//   - ResourceName: <pipeline name>
//   - ResourceType: "pipelinerun"
//   - ResourceNamespace: <pipeline namespace>
//
// Example:
//
//	    Content-Type: application/json
//		  POST /api/v1/webhooks/pipeline-success
//			 {
//			   "pipelineName": "frontend-build",
//			   "namespace": "team-alpha"
//			 }
func (h *WebhookHandler) PipelineSuccess(c *gin.Context) {
	var req PipelineSuccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields", "details": err.Error()})
		return
	}

	// Resolve any active issues for this pipeline
	resolved, err := h.issueService.ResolveIssuesByScope(c.Request.Context(), "pipelinerun", req.PipelineName, req.Namespace)
	if err != nil {
		h.logger.WithError(err).Errorf("failed to resolve issues for pipeline run %s : %v", req.PipelineName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve pipeline issues",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"pipeline":  req.PipelineName,
		"namespace": req.Namespace,
		"resolved":  resolved,
	}).Info("Pipeline success webhook processed")

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Resolved %d issue(s) for pipeline %s", resolved, req.PipelineName),
	})
}
