package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/services"
	"github.com/sirupsen/logrus"
)

type WebhookHandler struct {
	issueService services.IssueServiceInterface // IssueService instance
	logger       *logrus.Logger                 // Logging Instance
}

// NewWebhookHandler returns a new handler for the webhooks route
func NewWebhookHandler(issueService services.IssueServiceInterface, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		issueService: issueService,
		logger:       logger,
	}
}

type PipelineFailureRequest struct {
	PipelineName  string `json:"pipelineName" binding:"required"`
	Namespace     string `json:"namespace" binding:"required"`
	FailureReason string `json:"failureReason" binding:"required"`
	RunID         string `json:"runId"`
	LogsURL       string `json:"logsUrl"`
}

type PipelineSuccessRequest struct {
	PipelineName string `json:"pipelineName" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
}

// PipelineFailure handles pipeline failure webhooks
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
		// TODO - Update this to the actual cluster URL
		// Can probably be configured in the config package and referenced here.
		logsURL = fmt.Sprintf("https://konflux.dev/logs/pipelinerun/%s", req.RunID)
	}

	issueData := dto.CreateIssueRequest{
		Title:       fmt.Sprintf("Pipeline run failed: %s", req.PipelineName),
		Description: fmt.Sprintf("The pipeline run %s failed with reason: %s", req.PipelineName, req.FailureReason),
		Severity:    models.SeverityMajor, // TODO - check if we should make this configurable via the request.
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

	// Check for duplicates
	duplicateResult, err := h.issueService.CheckForDuplicateIssue(c.Request.Context(), issueData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check for duplicate pipeline issues")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	var issue *models.Issue
	// If an existing issue already exists, run an update
	if duplicateResult.IsDuplicate && duplicateResult.ExistingIssue != nil {
		// Update existing issue
		updateReq := dto.UpdateIssueRequest{
			Title:       &issueData.Title,
			Description: &issueData.Description,
			Severity:    &issue.Severity,
			IssueType:   &issue.IssueType,
			Links:       issueData.Links,
		}
		issue, err = h.issueService.UpdateIssue(c.Request.Context(), duplicateResult.ExistingIssue.ID, updateReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to update existing pipeline issue")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
			return
		}
		h.logger.WithField("issue_id", duplicateResult.ExistingIssue.ID).Info("Updated existing pipeline issue")
	} else {
		// Create new issue
		issue, err = h.issueService.CreateIssue(c.Request.Context(), issueData)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create pipeline issue: %w", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
			return
		}
		h.logger.WithField("issue_id", issue.ID).Info("Created new pipeline issue")
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"issue":  issue,
	})
}

// PipelineSuccess handles pipeline success webhooks
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
