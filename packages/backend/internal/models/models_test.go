package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestIssueCreation tests basic Issue struct creation
func TestIssueStructInit(t *testing.T) {
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
		t.Errorf("Expected Title '%s', got '%s'", expectedTitle, issue.Title)
	}

	if issue.Description != expectedDescription {
		t.Errorf("Expected Description '%s', got '%s'", expectedDescription, issue.Description)
	}

	if issue.Severity != SeverityMajor {
		t.Errorf("Expected Severity %s, got '%s'", SeverityMajor, issue.Severity)
	}

	if issue.IssueType != IssueTypeBuild {
		t.Errorf("Expected IssueType %s, got '%s'", IssueTypeBuild, issue.IssueType)
	}

	if issue.State != IssueStateActive {
		t.Errorf("Expected State %s, got '%s'", IssueStateActive, issue.State)
	}

	if issue.Namespace != expectedNamespace {
		t.Errorf("Expected Namespace '%s', got '%s'", expectedNamespace, issue.Namespace)
	}

	if issue.DetectedAt.IsZero() {
		t.Errorf("Expected DetectedAt to be set")
	}
}

func TestIssueScopeStructInit(t *testing.T) {
	expectedResourceType := "PipelineRun"
	expectedResourceName := "pipeline-run-xyz"
	expectedResourceNamespace := "test-namespace"

	issueScope := IssueScope{
		ResourceType:      expectedResourceType,
		ResourceName:      expectedResourceName,
		ResourceNamespace: expectedResourceNamespace,
	}

	if issueScope.ResourceType != expectedResourceType {
		t.Errorf("Expected ResourceType '%s', got '%s'", expectedResourceType, issueScope.ResourceType)
	}
	if issueScope.ResourceName != expectedResourceName {
		t.Errorf("Expected ResourceName '%s', got '%s'", expectedResourceName, issueScope.ResourceName)
	}
	if issueScope.ResourceNamespace != expectedResourceNamespace {
		t.Errorf("Expected ResourceType '%s', got '%s'", expectedResourceNamespace, issueScope.ResourceNamespace)
	}
}

func TestRelatedIssueStructInit(t *testing.T) {
	issueA := Issue{
		ID:          uuid.New().String(),
		Title:       "Test Issue A",
		Description: "Description of Test Issue A",
		Severity:    SeverityMajor,
		IssueType:   IssueTypeBuild,
		State:       IssueStateActive,
		DetectedAt:  time.Now(),
		Namespace:   "test-namespace",
		Links:       []Link{}, // Empty
		Scope: IssueScope{
			ResourceType:      "Pipeline Run",
			ResourceName:      "pr-run-a",
			ResourceNamespace: "test-namespace",
		},
	}

	issueB := Issue{
		ID:          uuid.New().String(),
		Title:       "Test Issue B",
		Description: "Description of Test Issue B",
		Severity:    SeverityMajor,
		IssueType:   IssueTypeBuild,
		State:       IssueStateActive,
		DetectedAt:  time.Now(),
		Namespace:   "test-namespace",
		Links:       []Link{}, // Empty
		Scope: IssueScope{
			ResourceType:      "Pipeline Run",
			ResourceName:      "pr-run-b",
			ResourceNamespace: "test-namespace",
		},
	}

	relatedIssue := RelatedIssue{
		SourceID: issueA.ID,
		TargetID: issueB.ID,
		Source:   issueA,
		Target:   issueB,
	}

	if relatedIssue.SourceID != issueA.ID {
		t.Errorf("Expected SourceID '%s', got '%s'", relatedIssue.SourceID, issueA.ID)
	}

	if relatedIssue.TargetID != issueB.ID {
		t.Errorf("Expected TargetID '%s', got '%s'", relatedIssue.TargetID, issueB.ID)
	}
}

func TestLinkStructInit(t *testing.T) {
	expectedLinkTitle := "Pipeline Run Failure"
	expectedLinkUrl := "konflux.dev/pipelineruns/xyz"
	link := Link{
		Title:   expectedLinkTitle,
		URL:     expectedLinkUrl,
		IssueID: uuid.NewString(),
	}

	if link.Title != expectedLinkTitle {
		t.Errorf("Expected Title '%s', got '%s'", expectedLinkTitle, link.Title)
	}

	if link.URL != expectedLinkUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedLinkUrl, link.URL)
	}
}
