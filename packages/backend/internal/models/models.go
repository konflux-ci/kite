package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Enums
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityMinor    Severity = "minor"
	SeverityMajor    Severity = "major"
	SeverityCritical Severity = "critical"
)

type IssueType string

const (
	IssueTypeBuild      IssueType = "build"
	IssueTypeTest       IssueType = "test"
	IssueTypeRelease    IssueType = "release"
	IssueTypeDependency IssueType = "dependency"
	IssueTypePipeline   IssueType = "pipeline"
)

type IssueState string

const (
	IssueStateActive   IssueState = "ACTIVE"
	IssueStateResolved IssueState = "RESOLVED"
)

// Issue represents an issue in the cluster
type Issue struct {
	ID          string     `gorm:"type:uuid;primaryKey;" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `gorm:"not null" json:"description"`
	Severity    Severity   `gorm:"type:varchar(20);not null" json:"severity"`
	IssueType   IssueType  `gorm:"type:varchar(20);not null" json:"issueType"`
	State       IssueState `gorm:"type:varchar(20);default:ACTIVE" json:"state"`
	DetectedAt  time.Time  `gorm:"not null" json:"detectedAt"`
	ResolvedAt  *time.Time `json:"resolvedAt"`
	Namespace   string     `gorm:"not null" json:"namespace"`

	// Foreign key to IssueScope
	ScopeID string     `gorm:"type:uuid;not null;unique" json:"scopeId"`
	Scope   IssueScope `gorm:"foreignKey:ScopeID" json:"scope"`

	// Relationships
	Links       []Link         `gorm:"foreignKey:IssueID" json:"links"`
	RelatedFrom []RelatedIssue `gorm:"foreignKey:SourceID" json:"relatedFrom"`
	RelatedTo   []RelatedIssue `gorm:"foreignKey:TargetID" json:"relatedTo"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate hook to set UUID if not provided
func (i *Issue) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// IssueScope represents the scope of an Issue
type IssueScope struct {
	ID                string `gorm:"type:uuid;primaryKey" json:"id"`
	ResourceType      string `gorm:"not null" json:"resourceType"`
	ResourceName      string `gorm:"not null" json:"resourceName"`
	ResourceNamespace string `gorm:"not null" json:"resourceNamespace"`

	// Relationship - one issue scope has one issue
	Issue *Issue `gorm:"foreignKey:ScopeID" json:"issue,omitempty"`
}

// BeforeCreate hook to set UUID if not provided
func (s *IssueScope) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// RelatedIssue represents relationships between issues
type RelatedIssue struct {
	ID       string `gorm:"type:uuid;primaryKey" json:"id"`
	SourceID string `gorm:"type:uuid;not null" json:"sourceId"`
	TargetID string `gorm:"type:uuid;not null" json:"targetId"`

	// Relationships
	Source Issue `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	Target Issue `gorm:"foreignKey:TargetID" json:"target,omitempty"`
}

// BeforeCreate hook to set UUID if not provided
func (r *RelatedIssue) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// Link represents a link associated with an issue
type Link struct {
	ID      string `gorm:"type:uuid;primaryKey" json:"id"`
	Title   string `gorm:"not null" json:"title"`
	URL     string `gorm:"not null" json:"url"`
	IssueID string `gorm:"type:uuid;not null" json:"issueId"`
	// Omit field when converting to JSON or deconverting from JSON
	Issue Issue `gorm:"foreignKey:IssueID" json:"-"`
}

// BeforeCreate hook to set UUID if not provided
func (l *Link) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
