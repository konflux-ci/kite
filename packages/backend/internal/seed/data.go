package seed

import (
	"fmt"
	"time"

	"github.com/konflux-ci/kite/internal/models"
	"gorm.io/gorm"
)

// SeedData seeds the database with sample data
func SeedData(db *gorm.DB) error {
	// Check if data already exists
	var count int64
	if err := db.Model(&models.Issue{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		fmt.Printf("Database already has %d issues, skipping seed\n", count)
		return nil
	}

	// Seed in transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Create scopes first and get their generated IDs
		scopes, err := seedIssueScopes(tx)
		if err != nil {
			return fmt.Errorf("failed to seed issue scopes: %w", err)
		}

		if err := seedIssues(tx, scopes); err != nil {
			return fmt.Errorf("failed to seed issues: %w", err)
		}

		if err := seedLinks(tx); err != nil {
			return fmt.Errorf("failed to seed links: %w", err)
		}

		if err := seedRelatedIssues(tx); err != nil {
			return fmt.Errorf("failed to seed related issues: %w", err)
		}

		fmt.Println("Database seeded successfully")
		return nil
	})
}

func seedIssueScopes(tx *gorm.DB) (map[string]string, error) {
	scopes := []models.IssueScope{
		{
			ResourceType:      "component",
			ResourceName:      "frontend-ui",
			ResourceNamespace: "team-alpha",
		},
		{
			ResourceType:      "component",
			ResourceName:      "backend-api",
			ResourceNamespace: "team-alpha",
		},
		{
			ResourceType:      "application",
			ResourceName:      "e-commerce-app",
			ResourceNamespace: "team-beta",
		},
		{
			ResourceType:      "component",
			ResourceName:      "frontend-ui-deps",
			ResourceNamespace: "team-alpha",
		},
		{
			ResourceType:      "workspace",
			ResourceName:      "data-processing-workspace",
			ResourceNamespace: "team-gamma",
		},
		{
			ResourceType:      "pipelinerun",
			ResourceName:      "analytics-service-deploy-123",
			ResourceNamespace: "team-delta",
		},
		{
			ResourceType:      "component",
			ResourceName:      "mobile-app",
			ResourceNamespace: "team-alpha",
		},
		{
			ResourceType:      "application",
			ResourceName:      "e-commerce-app-db",
			ResourceNamespace: "team-beta",
		},
		{
			ResourceType:      "component",
			ResourceName:      "logging-service",
			ResourceNamespace: "team-gamma",
		},
		{
			ResourceType:      "workspace",
			ResourceName:      "main-workspace",
			ResourceNamespace: "team-alpha",
		},
		{
			ResourceType:      "application",
			ResourceName:      "analytics-service",
			ResourceNamespace: "team-delta",
		},
		{
			ResourceType:      "application",
			ResourceName:      "e-commerce-app-quota",
			ResourceNamespace: "team-beta",
		},
	}

	// Create scopes and let GORM generate UUIDs
	if err := tx.Create(&scopes).Error; err != nil {
		return nil, err
	}

	// Return a map of logical names to generated UUIDs for reference
	scopeMap := map[string]string{
		"scope-failed-build-frontend":             scopes[0].ID,
		"scope-failed-test-api":                   scopes[1].ID,
		"scope-release-failed-production":         scopes[2].ID,
		"scope-dependency-update-needed-frontend": scopes[3].ID,
		"scope-pipeline-outdated":                 scopes[4].ID,
		"scope-failed-pipeline-run":               scopes[5].ID,
		"scope-test-flaky-mobile":                 scopes[6].ID,
		"scope-outdated-dependency-database":      scopes[7].ID,
		"scope-build-warning-logging":             scopes[8].ID,
		"scope-database-connection-timeout":       scopes[9].ID,
		"scope-permission-config-incorrect":       scopes[10].ID,
		"scope-resource-quota-exceeded":           scopes[11].ID,
	}

	return scopeMap, nil
}

func seedIssues(tx *gorm.DB, scopeMap map[string]string) error {
	now := time.Now()
	issues := []models.Issue{
		{
			Title:       "Frontend build failed due to dependency conflict",
			Description: "The build process for the frontend component failed because of conflicting versions of React dependencies",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypeBuild,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 30, 15, 45, 30, 0, time.UTC),
			Namespace:   "team-alpha",
			ScopeID:     scopeMap["scope-failed-build-frontend"],
		},
		{
			Title:       "API integration tests failing on database connection",
			Description: "Integration tests for the API component are failing because the database connection is timing out",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeTest,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 5, 1, 9, 15, 22, 0, time.UTC),
			Namespace:   "team-alpha",
			ScopeID:     scopeMap["scope-failed-test-api"],
		},
		{
			Title:       "Production release failed during deployment",
			Description: "The production release of the e-commerce application failed during the deployment phase due to resource limits",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeRelease,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 29, 18, 30, 45, 0, time.UTC),
			Namespace:   "team-beta",
			ScopeID:     scopeMap["scope-release-failed-production"],
		},
		{
			Title:       "Frontend dependency updates available",
			Description: "Security vulnerabilities found in current dependencies. Updates are available and recommended.",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypeDependency,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 28, 14, 20, 10, 0, time.UTC),
			Namespace:   "team-alpha",
			ScopeID:     scopeMap["scope-dependency-update-needed-frontend"],
		},
		{
			Title:       "Pipeline tasks using deprecated API versions",
			Description: "Several pipeline tasks are using API versions that will be deprecated in the next Konflux update",
			Severity:    models.SeverityMinor,
			IssueType:   models.IssueTypePipeline,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 25, 11, 10, 30, 0, time.UTC),
			Namespace:   "team-gamma",
			ScopeID:     scopeMap["scope-pipeline-outdated"],
		},
		{
			Title:       "Pipeline run failed during deployment stage",
			Description: "The pipeline run for the analytics service failed during the deployment stage due to insufficient permissions",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypePipeline,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 30, 16, 45, 20, 0, time.UTC),
			Namespace:   "team-delta",
			ScopeID:     scopeMap["scope-failed-pipeline-run"],
		},
		{
			Title:       "Mobile app tests showing intermittent failures",
			Description: "The integration tests for the mobile app component are showing intermittent failures that may be related to test environment stability",
			Severity:    models.SeverityMinor,
			IssueType:   models.IssueTypeTest,
			State:       models.IssueStateResolved,
			DetectedAt:  time.Date(2025, 4, 28, 10, 25, 15, 0, time.UTC),
			ResolvedAt:  &[]time.Time{time.Date(2025, 4, 29, 14, 35, 40, 0, time.UTC)}[0],
			Namespace:   "team-alpha",
			ScopeID:     scopeMap["scope-test-flaky-mobile"],
		},
		{
			Title:       "Database client library needs security update",
			Description: "The database client library used by multiple components has a critical security vulnerability that needs to be addressed",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeDependency,
			State:       models.IssueStateResolved,
			DetectedAt:  time.Date(2025, 4, 25, 9, 20, 30, 0, time.UTC),
			ResolvedAt:  &[]time.Time{time.Date(2025, 4, 30, 13, 40, 15, 0, time.UTC)}[0],
			Namespace:   "team-beta",
			ScopeID:     scopeMap["scope-outdated-dependency-database"],
		},
		{
			Title:       "Build warnings in logging component",
			Description: "The logging component is generating build warnings about deprecated APIs that should be addressed",
			Severity:    models.SeverityInfo,
			IssueType:   models.IssueTypeBuild,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 27, 15, 30, 45, 0, time.UTC),
			Namespace:   "team-gamma",
			ScopeID:     scopeMap["scope-build-warning-logging"],
		},
		{
			Title:       "Database connection timeouts affecting multiple components",
			Description: "Database connection timeouts are occurring across multiple components, potentially due to configuration or resource constraints",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeRelease,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 5, 1, 8, 10, 25, 0, time.UTC),
			Namespace:   "team-alpha",
			ScopeID:     scopeMap["scope-database-connection-timeout"],
		},
		{
			Title:       "Incorrect permission configuration for deployment service account",
			Description: "The service account used for deployments has insufficient permissions, causing pipeline failures",
			Severity:    models.SeverityMajor,
			IssueType:   models.IssueTypeRelease,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 30, 15, 30, 10, 0, time.UTC),
			Namespace:   "team-delta",
			ScopeID:     scopeMap["scope-permission-config-incorrect"],
		},
		{
			Title:       "Namespace resource quota exceeded during deployment",
			Description: "The namespace resource quota was exceeded during the deployment phase, causing the release to fail",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeRelease,
			State:       models.IssueStateActive,
			DetectedAt:  time.Date(2025, 4, 29, 18, 15, 30, 0, time.UTC),
			Namespace:   "team-beta",
			ScopeID:     scopeMap["scope-resource-quota-exceeded"],
		},
	}

	// Set timestamps and let GORM generate UUIDs for issues
	for i := range issues {
		issues[i].CreatedAt = now
		issues[i].UpdatedAt = now
	}

	return tx.Create(&issues).Error
}

func seedLinks(tx *gorm.DB) error {
	// First, get the issue IDs by querying the database
	var issues []models.Issue
	if err := tx.Find(&issues).Error; err != nil {
		return err
	}

	// Create a map of issue titles to IDs for easier linking
	issueMap := make(map[string]string)
	for _, issue := range issues {
		switch issue.Title {
		case "Frontend build failed due to dependency conflict":
			issueMap["failed-build-frontend"] = issue.ID
		case "API integration tests failing on database connection":
			issueMap["failed-test-api"] = issue.ID
		case "Production release failed during deployment":
			issueMap["release-failed-production"] = issue.ID
		case "Frontend dependency updates available":
			issueMap["dependency-update-needed-frontend"] = issue.ID
		case "Pipeline tasks using deprecated API versions":
			issueMap["pipeline-outdated"] = issue.ID
		case "Pipeline run failed during deployment stage":
			issueMap["failed-pipeline-run"] = issue.ID
		}
	}

	links := []models.Link{
		// failed-build-frontend links
		{
			Title:   "Build Logs",
			URL:     "https://konflux.dev/logs/build/frontend-ui/12345",
			IssueID: issueMap["failed-build-frontend"],
		},
		{
			Title:   "Fix Instructions",
			URL:     "https://konflux.dev/docs/fixing-dependency-conflicts",
			IssueID: issueMap["failed-build-frontend"],
		},
		// failed-test-api links
		{
			Title:   "Test Logs",
			URL:     "https://konflux.dev/logs/test/backend-api/23456",
			IssueID: issueMap["failed-test-api"],
		},
		{
			Title:   "Database Connection Guide",
			URL:     "https://konflux.dev/docs/database-connection-troubleshooting",
			IssueID: issueMap["failed-test-api"],
		},
		// release-failed-production links
		{
			Title:   "Release Logs",
			URL:     "https://konflux.dev/logs/release/e-commerce-app/34567",
			IssueID: issueMap["release-failed-production"],
		},
		{
			Title:   "Resource Configuration Guide",
			URL:     "https://konflux.dev/docs/resource-configuration",
			IssueID: issueMap["release-failed-production"],
		},
		// dependency-update-needed-frontend links
		{
			Title:   "Dependency Report",
			URL:     "https://konflux.dev/security/dependencies/frontend-ui/78901",
			IssueID: issueMap["dependency-update-needed-frontend"],
		},
		{
			Title:   "Update Instructions",
			URL:     "https://konflux.dev/docs/updating-dependencies-safely",
			IssueID: issueMap["dependency-update-needed-frontend"],
		},
		// pipeline-outdated links
		{
			Title:   "Pipeline Configuration",
			URL:     "https://konflux.dev/pipelines/data-processing-workspace/45678",
			IssueID: issueMap["pipeline-outdated"],
		},
		{
			Title:   "API Migration Guide",
			URL:     "https://konflux.dev/docs/api-migration-guide",
			IssueID: issueMap["pipeline-outdated"],
		},
		// failed-pipeline-run links
		{
			Title:   "Pipeline Run Logs",
			URL:     "https://konflux.dev/logs/pipelinerun/analytics-service-deploy-123/56789",
			IssueID: issueMap["failed-pipeline-run"],
		},
		{
			Title:   "Permissions Configuration Guide",
			URL:     "https://konflux.dev/docs/pipeline-permissions",
			IssueID: issueMap["failed-pipeline-run"],
		},
	}

	return tx.Create(&links).Error
}

func seedRelatedIssues(tx *gorm.DB) error {
	// Get issues by querying the database
	var issues []models.Issue
	if err := tx.Find(&issues).Error; err != nil {
		return err
	}

	// Create a map of issue titles to IDs
	issueMap := make(map[string]string)
	for _, issue := range issues {
		switch issue.Title {
		case "Frontend build failed due to dependency conflict":
			issueMap["failed-build-frontend"] = issue.ID
		case "Frontend dependency updates available":
			issueMap["dependency-update-needed-frontend"] = issue.ID
		case "API integration tests failing on database connection":
			issueMap["failed-test-api"] = issue.ID
		case "Database connection timeouts affecting multiple components":
			issueMap["database-connection-timeout"] = issue.ID
		case "Production release failed during deployment":
			issueMap["release-failed-production"] = issue.ID
		case "Namespace resource quota exceeded during deployment":
			issueMap["quota-issue"] = issue.ID
		case "Pipeline run failed during deployment stage":
			issueMap["failed-pipeline-run"] = issue.ID
		case "Incorrect permission configuration for deployment service account":
			issueMap["permission-config-incorrect"] = issue.ID
		case "Database client library needs security update":
			issueMap["outdated-dependency-database"] = issue.ID
		}
	}

	relatedIssues := []models.RelatedIssue{
		{
			SourceID: issueMap["failed-build-frontend"],
			TargetID: issueMap["dependency-update-needed-frontend"],
		},
		{
			SourceID: issueMap["failed-test-api"],
			TargetID: issueMap["database-connection-timeout"],
		},
		{
			SourceID: issueMap["release-failed-production"],
			TargetID: issueMap["quota-issue"],
		},
		{
			SourceID: issueMap["failed-pipeline-run"],
			TargetID: issueMap["permission-config-incorrect"],
		},
		{
			SourceID: issueMap["outdated-dependency-database"],
			TargetID: issueMap["database-connection-timeout"],
		},
	}

	// Filter out any relationships where we couldn't find the issue IDs
	var validRelations []models.RelatedIssue
	for _, rel := range relatedIssues {
		if rel.SourceID != "" && rel.TargetID != "" {
			validRelations = append(validRelations, rel)
		}
	}

	if len(validRelations) > 0 {
		return tx.Create(&validRelations).Error
	}

	return nil
}
