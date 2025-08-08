package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/konflux-ci/kite/internal/services"
	"github.com/sirupsen/logrus"
)

type IssueHandler struct {
	issueService services.IssueServiceInterface
	logger       *logrus.Logger
}

func NewIssueHandler(issueService services.IssueServiceInterface, logger *logrus.Logger) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
		logger:       logger,
	}
}

// GetIssues handles GET /issues
func (h *IssueHandler) GetIssues(c *gin.Context) {
	// Esxtract query params
	filters := repository.IssueQueryFilters{
		Namespace:    c.Query("namespace"),
		ResourceType: c.Query("resourceType"),
		ResourceName: c.Query("resourceName"),
		Search:       c.Query("search"),
	}

	// Parse optional enum params
	if severity := c.Query("severity"); severity != "" {
		// Convert to custom type, then assign
		sev := models.Severity(severity)
		filters.Severity = &sev
	}
	if issueType := c.Query("issueType"); issueType != "" {
		it := models.IssueType(issueType)
		filters.IssueType = &it
	}
	if state := c.Query("state"); state != "" {
		st := models.IssueState(state)
		filters.State = &st
	}

	// Parse pagination parameters
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	// Default limit
	if filters.Limit == 0 {
		filters.Limit = 50
	}

	result, err := h.issueService.FindIssues(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("failed to fetch issues")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch issues"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetIssue handles GET /issues/:id
func (h *IssueHandler) GetIssue(c *gin.Context) {
	id := c.Param("id")
	namespace := c.Query("namespace")

	issue, err := h.issueService.FindIssueByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to fetch issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch issue"})
		return
	}

	if issue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	if namespace != "" && issue.Namespace != namespace {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this namespace"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// CreateIssue handles POST /issues
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	var req dto.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.validateCreateIssueRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	issue, err := h.issueService.CreateIssue(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create issue"})
		return
	}

	c.JSON(http.StatusCreated, issue)
}

// UpdateIssue handles PUT /issues/:id
func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	id := c.Param("id")
	namespace := c.Query("namespace")

	var req dto.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Check if issue exists and verify namespace exists
	existingIssue, err := h.issueService.FindIssueByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to find issue for update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update issue"})
		return
	}
	if existingIssue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	// Verify namespace access
	if namespace != "" && existingIssue.Namespace != namespace {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this namespace"})
		return
	}

	updatedIssue, err := h.issueService.UpdateIssue(c.Request.Context(), id, req)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to update issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update issue"})
		return
	}

	c.JSON(http.StatusOK, updatedIssue)
}

// DeleteIssue handles DELETE /issues/:id
func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	id := c.Param("id")
	namespace := c.Query("namespace")

	existingIssue, err := h.issueService.FindIssueByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to find issue for deletion")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete issue"})
		return
	}
	if existingIssue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	// Namespace access check
	if namespace != "" && existingIssue.Namespace != namespace {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this namespace"})
		return
	}

	if err := h.issueService.DeleteIssue(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to delete issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete issue"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ResolveIssue handles POST /issues/:id/resolve
func (h *IssueHandler) ResolveIssue(c *gin.Context) {
	id := c.Param("id")
	namespace := c.Query("namespace")

	existingIssue, err := h.issueService.FindIssueByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("failed to find issue for resolution")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve issue"})
		return
	}

	if existingIssue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	// Namespace access check
	if namespace != "" && existingIssue.Namespace != namespace {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this namespace"})
		return
	}

	now := time.Now()
	state := models.IssueStateResolved
	req := dto.UpdateIssueRequest{
		State:      state,
		ResolvedAt: now,
	}

	updatedIssue, err := h.issueService.UpdateIssue(c.Request.Context(), id, req)
	if err != nil {
		h.logger.WithError(err).WithField("issue_id", id).Error("Failed to mark issue resolved")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve issue"})
		return
	}

	c.JSON(http.StatusOK, updatedIssue)
}

// AddRelatedIssue handles POST /issues/:id/related
func (h *IssueHandler) AddRelatedIssue(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		RelatedID string `json:"relatedId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing relatdId field"})
		return
	}

	if err := h.issueService.AddRelatedIssue(c.Request.Context(), id, req.RelatedID); err != nil {
		if err.Error() == "one or both issues not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "relationship already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		h.logger.WithError(err).Error("Failed to add related issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create issue relationship"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Relationship created"})
}

// RemoveRelatedIssue handles DELETE /issues/:id/related/:relatedId
func (h *IssueHandler) RemoveRelatedIssue(c *gin.Context) {
	id := c.Param("id")
	relatedID := c.Param("relatedId")

	if err := h.issueService.RemoveRelatedIssue(c.Request.Context(), id, relatedID); err != nil {
		if err.Error() == "relationship not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.WithError(err).Error("Failed to remove related issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete issue relationship"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper function for validation issue creation
func (h *IssueHandler) validateCreateIssueRequest(req dto.CreateIssueRequest) error {
	// Validate severity
	validSeverities := []models.Severity{
		models.SeverityInfo, models.SeverityMinor,
		models.SeverityMajor, models.SeverityCritical,
	}

	if !slices.Contains(validSeverities, req.Severity) {
		return errors.New("invalid severity value")
	}

	// Validate issue type
	validTypes := []models.IssueType{
		models.IssueTypeBuild, models.IssueTypeTest,
		models.IssueTypeRelease, models.IssueTypeDependency,
		models.IssueTypePipeline,
	}
	if !slices.Contains(validTypes, req.IssueType) {
		return errors.New("invalid issueType value")
	}

	// validate state if provided
	if req.State != "" {
		validStates := []models.IssueState{models.IssueStateActive, models.IssueStateResolved}
		if !slices.Contains(validStates, req.State) {
			return errors.New("invalid state value")
		}
	}

	return nil
}
