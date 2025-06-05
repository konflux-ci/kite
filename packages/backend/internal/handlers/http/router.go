package http

import (
	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/middleware"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/konflux-ci/kite/internal/services"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, logger *logrus.Logger) (*gin.Engine, error) {
	// Set Gin mode based on environmetn
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Setup middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Initialize repository
	issueRepo := repository.NewIssueRepository(db, logger)
	// Initialize services
	issueService := services.NewIssueService(issueRepo, logger)

	// Initialize handlers
	issueHandler := NewIssueHandler(issueService, logger)
	webhookHandler := NewWebhookHandler(issueService, logger)

	// Initialize namespace checker
	namespaceChecker, err := middleware.NewNamespaceChecker(logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize namespace checker")
	}

	// Health and version endpoints
	router.GET("/health", middleware.HealthCheck(logger))
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":     "1.0.0",
			"name":        "Konflux Issues API",
			"description": "API for managing issues in Konflux",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Issues routes with namespace checking
	issuesGroup := v1.Group("/issues")
	if namespaceChecker != nil {
		issuesGroup.Use(namespaceChecker.CheckNamespacessAccess())
	}
	{
		issuesGroup.GET("/", issueHandler.GetIssues)
		issuesGroup.POST("/", issueHandler.CreateIssue)
		issuesGroup.GET("/:id", middleware.ValidateID(), issueHandler.GetIssue)
		issuesGroup.PUT("/:id", middleware.ValidateID(), issueHandler.UpdateIssue)
		issuesGroup.DELETE("/:id", middleware.ValidateID(), issueHandler.DeleteIssue)
		issuesGroup.POST("/:id/resolve", middleware.ValidateID(), issueHandler.ResolveIssue)
		issuesGroup.POST("/:id/related", middleware.ValidateID(), issueHandler.AddRelatedIssue)
		issuesGroup.DELETE("/:id/related/:relatedId", middleware.ValidateID(), issueHandler.RemoveRelatedIssue)
	}

	// Webhook routes with namespace checking
	webhooksGroup := v1.Group("/webhooks")
	if namespaceChecker != nil {
		webhooksGroup.Use(namespaceChecker.CheckNamespacessAccess())
	}
	{
		webhooksGroup.POST("/pipeline-failure", webhookHandler.PipelineFailure)
		webhooksGroup.POST("/pipeline-success", webhookHandler.PipelineSuccess)
	}

	return router, nil
}
