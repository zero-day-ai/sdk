package domain

import "github.com/zero-day-ai/sdk/graphrag"

// API represents a web API service with multiple endpoints.
// APIs are root entities that can have multiple endpoints associated with them.
//
// Hierarchy: API is a root node (no parent)
//
// Identifying Properties: base_url
// Parent: None (root node)
//
// Example:
//
//	api := &API{
//	    BaseURL:     "https://api.example.com",
//	    Name:        "Example API",
//	    Version:     "v1",
//	    Description: "RESTful API for example service",
//	}
type API struct {
	// BaseURL is the API base URL (e.g., "https://api.example.com").
	// This is an identifying property.
	BaseURL string

	// Name is the API name (optional).
	Name string

	// Version is the API version (e.g., "v1", "2.0") (optional).
	Version string

	// Description is the API description (optional).
	Description string

	// SwaggerURL is the URL to the Swagger/OpenAPI specification (optional).
	SwaggerURL string

	// AuthType is the authentication type (e.g., "bearer", "basic", "oauth2") (optional).
	AuthType string

	// RateLimit is the rate limit (requests per period) (optional).
	RateLimit string

	// Status is the API status (e.g., "active", "deprecated") (optional).
	Status string
}

// NodeType returns the canonical node type for APIs.
func (a *API) NodeType() string {
	return graphrag.NodeTypeApi
}

// IdentifyingProperties returns the properties that uniquely identify this API.
// An API is identified by its base URL.
func (a *API) IdentifyingProperties() map[string]any {
	return map[string]any{
		"base_url": a.BaseURL,
	}
}

// Properties returns all properties to set on the API node.
func (a *API) Properties() map[string]any {
	props := map[string]any{
		"base_url": a.BaseURL,
	}

	// Add optional properties if present
	if a.Name != "" {
		props["name"] = a.Name
	}
	if a.Version != "" {
		props["version"] = a.Version
	}
	if a.Description != "" {
		props["description"] = a.Description
	}
	if a.SwaggerURL != "" {
		props["swagger_url"] = a.SwaggerURL
	}
	if a.AuthType != "" {
		props["auth_type"] = a.AuthType
	}
	if a.RateLimit != "" {
		props["rate_limit"] = a.RateLimit
	}
	if a.Status != "" {
		props["status"] = a.Status
	}

	return props
}

// ParentRef returns nil because API is a root node.
func (a *API) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because API has no parent.
func (a *API) RelationshipType() string {
	return ""
}
