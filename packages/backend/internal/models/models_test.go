package models

import (
	"testing"
	"time"
)

// TestIssueCreation tests basic Issue struct creation
func TestIssueCreation(t *testing.T) {
	expectedTitle := "Test Issue"
	expectedDescription := "This is a test issue"
	expectedNamespace := "test-namespace"
	// setup test data
	issue := Issue{
		Title:       expectedTitle,
		Description: expectedDescription,
		Severity:    SeverityMajor,
		IssueType:   IssueTypeBuild,
		State:       IssueStateActive,
		DetectedAt:  time.Now(),
		Namespace:   expectedNamespace,
	}

	// Run assertions
	if issue.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, issue.Title)
	}

	if issue.Description != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedTitle, issue.Title)
	}

	if issue.Severity != SeverityMajor {
		t.Errorf("Expected severity %s, got '%s'", SeverityMajor, issue.Severity)
	}

	if issue.IssueType != IssueTypeBuild {
		t.Errorf("Expected IssueType %s, got '%s'", IssueTypeBuild, issue.IssueType)
	}

	if issue.State != IssueStateActive {
		t.Errorf("Expected state %s, got '%s'", IssueStateActive, issue.State)
	}

	if issue.Namespace != expectedNamespace {
		t.Errorf("Expected namespace '%s', got '%s'", expectedNamespace, issue.Namespace)
	}
}
