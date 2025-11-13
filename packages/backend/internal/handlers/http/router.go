package http

import (
	"github.com/gin-gonic/gin"
	kiteConf "github.com/konflux-ci/kite/internal/config"
	"github.com/konflux-ci/kite/internal/middleware"
	"github.com/konflux-ci/kite/internal/repository"
	"github.com/konflux-ci/kite/internal/services"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, logger *logrus.Logger) (*gin.Engine, error) {
	// Set Gin mode based on environment
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
		// custom webhook for mintmaker
		webhooksGroup.POST("/mintmaker-custom", webhookHandler.MintmakerIssues)
		// custom webhooks for release-service
		webhooksGroup.POST("/release-failure", webhookHandler.ReleaseFailure)
		webhooksGroup.POST("/release-success", webhookHandler.ReleaseSuccess)
	}

	// Health and version endpoints
	healthGroup := v1.Group("/health")
	healthGroup.GET("/", NewHealthHandler(db, logger))

	versionGroup := v1.Group("/version")
	versionGroup.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "Konflux Issues Dashboard API",
			"description": "The backend service that powers the Konflux Issues Dashboard",
			"version":     kiteConf.GetEnvOrDefault("KITE_VERSION", "0.0.1"),
		})
	})

	return router, nil
}
