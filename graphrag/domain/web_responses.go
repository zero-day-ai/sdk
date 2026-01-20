package domain

// Response represents an HTTP response from an endpoint.
// This type supports testing, monitoring, and API behavior analysis.
//
// Example:
//
//	response := &Response{
//	    EndpointID:  "https://api.example.com/users",
//	    StatusCode:  200,
//	    ContentType: "application/json",
//	}
type Response struct {
	// EndpointID is the ID of the endpoint that returned this response.
	// Required. Should match an APIEndpoint or Endpoint ID.
	EndpointID string

	// StatusCode is the HTTP status code (200, 404, 500, etc.).
	// Required.
	StatusCode int

	// ContentType is the Content-Type header value.
	// Optional.
	ContentType string

	// Headers is a map of response headers.
	// Optional.
	Headers map[string]string

	// BodySize is the size of the response body in bytes.
	// Optional.
	BodySize int

	// ResponseTime is the time taken to receive the response in milliseconds.
	// Optional.
	ResponseTime int

	// Example is an example response body.
	// Optional.
	Example string
}

func (r *Response) NodeType() string { return "response" }

func (r *Response) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint_id": r.EndpointID,
		"status_code": r.StatusCode,
	}
}

func (r *Response) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.ContentType != "" {
		props["content_type"] = r.ContentType
	}
	if len(r.Headers) > 0 {
		props["headers"] = r.Headers
	}
	if r.BodySize > 0 {
		props["body_size"] = r.BodySize
	}
	if r.ResponseTime > 0 {
		props["response_time"] = r.ResponseTime
	}
	if r.Example != "" {
		props["example"] = r.Example
	}
	return props
}

func (r *Response) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "api_endpoint",
		Properties: map[string]any{
			"id": r.EndpointID,
		},
	}
}

func (r *Response) RelationshipType() string { return "HAS_RESPONSE" }

// StatusCode represents an HTTP status code definition for an API.
// This type supports API contract analysis and expected behavior documentation.
//
// Example:
//
//	statusCode := &StatusCode{
//	    EndpointID:  "https://api.example.com/users",
//	    Code:        404,
//	    Description: "User not found",
//	}
type StatusCode struct {
	// EndpointID is the ID of the endpoint that can return this status code.
	// Required. Should match an APIEndpoint or Endpoint ID.
	EndpointID string

	// Code is the HTTP status code (200, 404, 500, etc.).
	// Required.
	Code int

	// Description is a description of what this status code means for this endpoint.
	// Optional.
	Description string

	// Example is an example response body for this status code.
	// Optional.
	Example string
}

func (s *StatusCode) NodeType() string { return "status_code" }

func (s *StatusCode) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint_id": s.EndpointID,
		"code":        s.Code,
	}
}

func (s *StatusCode) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Description != "" {
		props["description"] = s.Description
	}
	if s.Example != "" {
		props["example"] = s.Example
	}
	return props
}

func (s *StatusCode) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "api_endpoint",
		Properties: map[string]any{
			"id": s.EndpointID,
		},
	}
}

func (s *StatusCode) RelationshipType() string { return "HAS_STATUS_CODE" }
