// DTOs (Data Transfer Objects)
// These allow us to carry and format data between layers or services, without embedding any business logic.
package dto

import (
	"time"

	"github.com/konflux-ci/kite/internal/models"
)

// ScopePayload is the interface implemented by both required and optional scope
// payload structs. It allows handlers/service to accept the same scope
// or both CREATE (required fields) and UPDATE (optional/patch) requests.
type ScopePayload interface {
	GetResourceType() string
	GetResourceName() string
	GetResourceNamespace() string
	// AsOptional returns an optional/patch form of the scope payload.
	// this is useful when you need to forward scope data to an API that accepts
	// partial updates.
	AsOptional() ScopeReqBodyOptional
}

// ScopeReqBody represents a required scope in CREATE requests.
// All fields excepted ResourceNamespace are required.
type ScopeReqBody struct {
	ResourceType      string `json:"resourceType" binding:"required"`
	ResourceName      string `json:"resourceName" binding:"required"`
	ResourceNamespace string `json:"resourceNamespace"`
}

func (s ScopeReqBody) GetResourceType() string      { return s.ResourceType }
func (s ScopeReqBody) GetResourceName() string      { return s.ResourceName }
func (s ScopeReqBody) GetResourceNamespace() string { return s.ResourceNamespace }
func (s ScopeReqBody) AsOptional() ScopeReqBodyOptional {
	return ScopeReqBodyOptional(s)
}

// ScopeReqBody represents an optional/patch scope in UPDATE requests.
// All fields are optional.
type ScopeReqBodyOptional struct {
	ResourceType      string `json:"resourceType"`
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
}

func (s ScopeReqBodyOptional) GetResourceType() string      { return s.ResourceType }
func (s ScopeReqBodyOptional) GetResourceName() string      { return s.ResourceName }
func (s ScopeReqBodyOptional) GetResourceNamespace() string { return s.ResourceNamespace }
func (s ScopeReqBodyOptional) AsOptional() ScopeReqBodyOptional {
	return s
}

// CreateIssueRequest is the payload for creating a new issue.
// Required Fields: Title, Description, Severity, IssueType, Namespace, Scope.
// State is optional, defaults to "ACTIVE".
type CreateIssueRequest struct {
	Title       string              `json:"title" binding:"required"`
	Description string              `json:"description" binding:"required"`
	Severity    models.Severity     `json:"severity" binding:"required"`
	IssueType   models.IssueType    `json:"issueType" binding:"required"`
	State       models.IssueState   `json:"state"`
	Namespace   string              `json:"namespace" binding:"required"`
	Scope       ScopeReqBody        `json:"scope" binding:"required"`
	Links       []CreateLinkRequest `json:"links"`
}

// CreateLinkRequest represents a link associated with an issue.
type CreateLinkRequest struct {
	Title string `json:"title" binding:"required"`
	URL   string `json:"url" binding:"required"`
}

// UpdateIssueRequest is the payload for updating an existing issue.
// All fields are optional. Only provided fields will be updated.
// If ResolvedAt is non-zero, the issue will be considered resolved by the service.
type UpdateIssueRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Severity    models.Severity      `json:"severity"`
	IssueType   models.IssueType     `json:"issueType"`
	State       models.IssueState    `json:"state"`
	Namespace   string               `json:"namespace"`
	Scope       ScopeReqBodyOptional `json:"scope"`
	Links       []CreateLinkRequest  `json:"links"`
	ResolvedAt  time.Time            `json:"resolvedAt"`
}

// IssuePayload unifies CREATE and UPDATE payloads for issues so services can accept either.
type IssuePayload interface {
	GetTitle() string
	GetDescription() string
	GetSeverity() models.Severity
	GetIssueType() models.IssueType
	GetState() models.IssueState
	GetLinks() []CreateLinkRequest
	GetResolvedAt() time.Time
	GetNamespace() string
	GetScope() ScopePayload
}

func (c CreateIssueRequest) GetTitle() string               { return c.Title }
func (c CreateIssueRequest) GetDescription() string         { return c.Description }
func (c CreateIssueRequest) GetSeverity() models.Severity   { return c.Severity }
func (c CreateIssueRequest) GetIssueType() models.IssueType { return c.IssueType }
func (c CreateIssueRequest) GetState() models.IssueState    { return c.State }
func (c CreateIssueRequest) GetLinks() []CreateLinkRequest  { return c.Links }
func (c CreateIssueRequest) GetScope() ScopePayload         { return c.Scope }
func (c CreateIssueRequest) GetNamespace() string           { return c.Namespace }
func (c CreateIssueRequest) GetResolvedAt() time.Time {
	// CREATE requests do not set a resolved time. Return a zero time value.
	return time.Time{}
}

func (u UpdateIssueRequest) GetTitle() string               { return u.Title }
func (u UpdateIssueRequest) GetDescription() string         { return u.Description }
func (u UpdateIssueRequest) GetSeverity() models.Severity   { return u.Severity }
func (u UpdateIssueRequest) GetIssueType() models.IssueType { return u.IssueType }
func (u UpdateIssueRequest) GetState() models.IssueState    { return u.State }
func (u UpdateIssueRequest) GetLinks() []CreateLinkRequest  { return u.Links }
func (u UpdateIssueRequest) GetScope() ScopePayload         { return u.Scope }
func (u UpdateIssueRequest) GetNamespace() string           { return u.Namespace }
func (u UpdateIssueRequest) GetResolvedAt() time.Time       { return u.ResolvedAt }
