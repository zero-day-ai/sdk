package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Endpoint represents a web endpoint or URL discovered through crawling or fuzzing.
// Endpoints are specific HTTP/HTTPS URLs on a service with a method (GET, POST, etc.).
//
// Hierarchy: Host -> Port -> Service -> Endpoint
//
// Identifying Properties: service_id, url, method
// Parent: Service (via RUNS_SERVICE relationship)
//
// Example:
//
//	endpoint := &Endpoint{
//	    ServiceID:  "192.168.1.1:443:tcp:https",
//	    URL:        "/api/users",
//	    Method:     "GET",
//	    StatusCode: 200,
//	    Headers:    map[string]string{"Content-Type": "application/json"},
//	}
type Endpoint struct {
	// ServiceID is the composite ID of the parent service ("{host_id}:{number}:{protocol}:{service_name}").
	// This is an identifying property.
	ServiceID string `json:"service_id"`

	// URL is the endpoint path or full URL.
	// This is an identifying property.
	URL string `json:"url"`

	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.).
	// This is an identifying property.
	Method string `json:"method"`

	// StatusCode is the HTTP response status code (optional).
	StatusCode int `json:"status_code,omitempty"`

	// Headers contains HTTP response headers (optional).
	Headers map[string]string `json:"headers,omitempty"`

	// ResponseTime is the response time in milliseconds (optional).
	ResponseTime int64 `json:"response_time,omitempty"`

	// ContentType is the response Content-Type header (optional).
	ContentType string `json:"content_type,omitempty"`

	// ContentLength is the response Content-Length in bytes (optional).
	ContentLength int64 `json:"content_length,omitempty"`
}

// NodeType returns the canonical node type for endpoints.
func (e *Endpoint) NodeType() string {
	return graphrag.NodeTypeEndpoint
}

// IdentifyingProperties returns the properties that uniquely identify this endpoint.
// An endpoint is identified by its service, URL path, and HTTP method.
func (e *Endpoint) IdentifyingProperties() map[string]any {
	return map[string]any{
		"service_id": e.ServiceID,
		"url":        e.URL,
		"method":     e.Method,
	}
}

// Properties returns all properties to set on the endpoint node.
func (e *Endpoint) Properties() map[string]any {
	props := map[string]any{
		"service_id": e.ServiceID,
		"url":        e.URL,
		"method":     e.Method,
	}

	// Add optional properties if present
	if e.StatusCode != 0 {
		props["status_code"] = e.StatusCode
	}
	if e.Headers != nil && len(e.Headers) > 0 {
		props["headers"] = e.Headers
	}
	if e.ResponseTime != 0 {
		props["response_time"] = e.ResponseTime
	}
	if e.ContentType != "" {
		props["content_type"] = e.ContentType
	}
	if e.ContentLength != 0 {
		props["content_length"] = e.ContentLength
	}

	return props
}

// ParentRef returns a reference to the parent Service node.
// The service is identified by its composite service_id.
func (e *Endpoint) ParentRef() *NodeRef {
	if e.ServiceID == "" {
		return nil
	}

	// ServiceID format: "{host_id}:{number}:{protocol}:{service_name}"
	// We need to parse it to get the port_id and service name
	// For now, we use the full ServiceID as a unique identifier
	return &NodeRef{
		NodeType: graphrag.NodeTypeService,
		Properties: map[string]any{
			"service_id": e.ServiceID,
		},
	}
}

// RelationshipType returns the relationship type to the parent service.
func (e *Endpoint) RelationshipType() string {
	return graphrag.RelTypeHasEndpoint
}
