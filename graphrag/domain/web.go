package domain

import "github.com/zero-day-ai/sdk/graphrag"

// APIEndpoint represents a specific endpoint within a web API.
// Endpoints combine a path and HTTP method to form a unique resource.
//
// Example:
//
//	endpoint := &APIEndpoint{
//	    APIID:       "api.example.com",
//	    Path:        "/api/v1/users",
//	    Method:      "POST",
//	    Description: "Create a new user",
//	}
//
// Identifying Properties:
//   - api_id (required): The API this endpoint belongs to
//   - path (required): The URL path
//   - method (required): HTTP method (GET, POST, PUT, DELETE, etc.)
//
// Relationships:
//   - Parent: API node (via HAS_ENDPOINT relationship)
type APIEndpoint struct {
	// APIID is the identifier of the parent API.
	// This is an identifying property and is required.
	APIID string

	// Path is the URL path for this endpoint.
	// This is an identifying property and is required.
	// Example: "/api/v1/users", "/graphql"
	Path string

	// Method is the HTTP method.
	// This is an identifying property and is required.
	// Common values: "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"
	Method string

	// Description is a human-readable description of what this endpoint does.
	// Optional.
	Description string

	// AuthRequired indicates if authentication is required.
	// Optional. Default: false
	AuthRequired bool

	// AuthType is the type of authentication required.
	// Optional. Common values: "bearer", "basic", "api_key", "oauth2"
	AuthType string

	// Deprecated indicates if this endpoint is deprecated.
	// Optional. Default: false
	Deprecated bool
}

func (a *APIEndpoint) NodeType() string { return "api_endpoint" }

func (a *APIEndpoint) IdentifyingProperties() map[string]any {
	return map[string]any{
		"api_id":            a.APIID,
		"path":              a.Path,
		graphrag.PropMethod: a.Method,
	}
}

func (a *APIEndpoint) Properties() map[string]any {
	props := a.IdentifyingProperties()
	if a.Description != "" {
		props[graphrag.PropDescription] = a.Description
	}
	props["auth_required"] = a.AuthRequired
	if a.AuthType != "" {
		props["auth_type"] = a.AuthType
	}
	props["deprecated"] = a.Deprecated
	return props
}

func (a *APIEndpoint) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: graphrag.NodeTypeApi,
		Properties: map[string]any{
			"base_url": a.APIID,
		},
	}
}

func (a *APIEndpoint) RelationshipType() string { return graphrag.RelTypeHasEndpoint }

// Parameter represents a parameter accepted by an API endpoint.
// Parameters can be in the query string, path, header, or body.
//
// Example:
//
//	param := &Parameter{
//	    EndpointID: "/api/v1/users:POST",
//	    Name:       "username",
//	    Type:       "string",
//	    Location:   "body",
//	    Required:   true,
//	}
//
// Identifying Properties:
//   - endpoint_id (required): The endpoint this parameter belongs to
//   - name (required): Parameter name
//
// Relationships:
//   - Parent: APIEndpoint node (via HAS_PARAMETER relationship)
type Parameter struct {
	// EndpointID is the identifier of the parent endpoint.
	// This is an identifying property and is required.
	EndpointID string

	// Name is the parameter name.
	// This is an identifying property and is required.
	Name string

	// Type is the data type.
	// Optional. Common values: "string", "integer", "boolean", "array", "object"
	Type string

	// Location is where the parameter is provided.
	// Optional. Common values: "query", "path", "header", "body", "cookie"
	Location string

	// Required indicates if the parameter is required.
	// Optional. Default: false
	Required bool

	// Description is a description of the parameter.
	// Optional.
	Description string

	// DefaultValue is the default value if not provided.
	// Optional.
	DefaultValue string
}

func (p *Parameter) NodeType() string { return "parameter" }

func (p *Parameter) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint_id":     p.EndpointID,
		graphrag.PropName: p.Name,
	}
}

func (p *Parameter) Properties() map[string]any {
	props := p.IdentifyingProperties()
	if p.Type != "" {
		props["type"] = p.Type
	}
	if p.Location != "" {
		props["location"] = p.Location
	}
	props["required"] = p.Required
	if p.Description != "" {
		props[graphrag.PropDescription] = p.Description
	}
	if p.DefaultValue != "" {
		props["default_value"] = p.DefaultValue
	}
	return props
}

func (p *Parameter) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "api_endpoint",
		Properties: map[string]any{
			"id": p.EndpointID,
		},
	}
}

func (p *Parameter) RelationshipType() string { return "HAS_PARAMETER" }

// Header represents an HTTP header used in requests or responses.
// Headers carry metadata about the HTTP transaction.
//
// Example:
//
//	header := &Header{
//	    Name:        "Authorization",
//	    Description: "Bearer token for authentication",
//	    Required:    true,
//	}
//
// Identifying Properties:
//   - name (required): Header name (case-insensitive per HTTP spec)
//
// Relationships:
//   - None (root node)
type Header struct {
	// Name is the HTTP header name.
	// This is an identifying property and is required.
	// Example: "Authorization", "Content-Type", "X-API-Key"
	Name string

	// Description is a description of the header's purpose.
	// Optional.
	Description string

	// Required indicates if this header is required.
	// Optional. Default: false
	Required bool

	// Example is an example value for documentation.
	// Optional.
	Example string
}

func (h *Header) NodeType() string { return "header" }

func (h *Header) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: h.Name,
	}
}

func (h *Header) Properties() map[string]any {
	props := h.IdentifyingProperties()
	if h.Description != "" {
		props[graphrag.PropDescription] = h.Description
	}
	props["required"] = h.Required
	if h.Example != "" {
		props["example"] = h.Example
	}
	return props
}

func (h *Header) ParentRef() *NodeRef      { return nil }
func (h *Header) RelationshipType() string { return "" }

// Cookie represents an HTTP cookie.
// Cookies store state between client and server.
//
// Example:
//
//	cookie := &Cookie{
//	    Name:     "session_id",
//	    Domain:   ".example.com",
//	    Path:     "/",
//	    Secure:   true,
//	    HTTPOnly: true,
//	}
//
// Identifying Properties:
//   - name (required): Cookie name
//   - domain (required): Cookie domain
//
// Relationships:
//   - None (root node)
type Cookie struct {
	// Name is the cookie name.
	// This is an identifying property and is required.
	Name string

	// Domain is the domain scope for the cookie.
	// This is an identifying property and is required.
	// Example: ".example.com", "api.example.com"
	Domain string

	// Path is the URL path scope for the cookie.
	// Optional. Default: "/"
	Path string

	// Secure indicates if the cookie should only be sent over HTTPS.
	// Optional. Default: false
	Secure bool

	// HTTPOnly indicates if the cookie is inaccessible to JavaScript.
	// Optional. Default: false
	HTTPOnly bool

	// SameSite is the SameSite attribute value.
	// Optional. Common values: "Strict", "Lax", "None"
	SameSite string

	// MaxAge is the cookie lifetime in seconds.
	// Optional.
	MaxAge int
}

func (c *Cookie) NodeType() string { return "cookie" }

func (c *Cookie) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
		"domain":          c.Domain,
	}
}

func (c *Cookie) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if c.Path != "" {
		props["path"] = c.Path
	}
	props["secure"] = c.Secure
	props["http_only"] = c.HTTPOnly
	if c.SameSite != "" {
		props["same_site"] = c.SameSite
	}
	if c.MaxAge > 0 {
		props["max_age"] = c.MaxAge
	}
	return props
}

func (c *Cookie) ParentRef() *NodeRef      { return nil }
func (c *Cookie) RelationshipType() string { return "" }

// Form represents an HTML form on a web page.
// Forms collect user input and submit it to a server.
//
// Example:
//
//	form := &Form{
//	    EndpointID: "/login",
//	    Action:     "/api/auth/login",
//	    Method:     "POST",
//	    Name:       "login-form",
//	}
//
// Identifying Properties:
//   - endpoint_id (required): The endpoint where this form appears
//   - action (required): The URL where the form submits to
//
// Relationships:
//   - Parent: Endpoint node (via HAS_FORM relationship)
type Form struct {
	// EndpointID is the identifier of the endpoint where this form appears.
	// This is an identifying property and is required.
	EndpointID string

	// Action is the URL where the form data is submitted.
	// This is an identifying property and is required.
	Action string

	// Method is the HTTP method used for form submission.
	// Optional. Common values: "GET", "POST"
	Method string

	// Name is the form name attribute.
	// Optional.
	Name string

	// Encoding is the form encoding type.
	// Optional. Common values: "application/x-www-form-urlencoded", "multipart/form-data"
	Encoding string
}

func (f *Form) NodeType() string { return "form" }

func (f *Form) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint_id": f.EndpointID,
		"action":      f.Action,
	}
}

func (f *Form) Properties() map[string]any {
	props := f.IdentifyingProperties()
	if f.Method != "" {
		props[graphrag.PropMethod] = f.Method
	}
	if f.Name != "" {
		props[graphrag.PropName] = f.Name
	}
	if f.Encoding != "" {
		props["encoding"] = f.Encoding
	}
	return props
}

func (f *Form) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: graphrag.NodeTypeEndpoint,
		Properties: map[string]any{
			graphrag.PropURL: f.EndpointID,
		},
	}
}

func (f *Form) RelationshipType() string { return "HAS_FORM" }

// FormField represents a single input field in an HTML form.
// Fields collect specific pieces of user input.
//
// Example:
//
//	field := &FormField{
//	    FormID:      "/api/auth/login",
//	    Name:        "username",
//	    Type:        "text",
//	    Required:    true,
//	}
//
// Identifying Properties:
//   - form_id (required): The form this field belongs to
//   - name (required): Field name attribute
//
// Relationships:
//   - Parent: Form node (via HAS_FIELD relationship)
type FormField struct {
	// FormID is the identifier of the parent form.
	// This is an identifying property and is required.
	FormID string

	// Name is the field name attribute.
	// This is an identifying property and is required.
	Name string

	// Type is the input type.
	// Optional. Common values: "text", "password", "email", "checkbox", "hidden", "submit"
	Type string

	// Required indicates if the field is required.
	// Optional. Default: false
	Required bool

	// MaxLength is the maximum input length.
	// Optional.
	MaxLength int

	// Pattern is a regex pattern for validation.
	// Optional.
	Pattern string

	// Placeholder is the placeholder text.
	// Optional.
	Placeholder string
}

func (f *FormField) NodeType() string { return "form_field" }

func (f *FormField) IdentifyingProperties() map[string]any {
	return map[string]any{
		"form_id":         f.FormID,
		graphrag.PropName: f.Name,
	}
}

func (f *FormField) Properties() map[string]any {
	props := f.IdentifyingProperties()
	if f.Type != "" {
		props["type"] = f.Type
	}
	props["required"] = f.Required
	if f.MaxLength > 0 {
		props["max_length"] = f.MaxLength
	}
	if f.Pattern != "" {
		props["pattern"] = f.Pattern
	}
	if f.Placeholder != "" {
		props["placeholder"] = f.Placeholder
	}
	return props
}

func (f *FormField) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "form",
		Properties: map[string]any{
			"action": f.FormID,
		},
	}
}

func (f *FormField) RelationshipType() string { return "HAS_FIELD" }

// WebSocket represents a WebSocket connection endpoint.
// WebSockets provide full-duplex communication channels over TCP.
//
// Example:
//
//	ws := &WebSocket{
//	    URL:      "wss://api.example.com/ws",
//	    Protocol: "chat-protocol-v1",
//	    Origin:   "https://example.com",
//	}
//
// Identifying Properties:
//   - url (required): WebSocket URL
//
// Relationships:
//   - None (root node)
type WebSocket struct {
	// URL is the WebSocket endpoint URL.
	// This is an identifying property and is required.
	// Example: "ws://example.com/socket", "wss://api.example.com/ws"
	URL string

	// Protocol is the WebSocket sub-protocol.
	// Optional. Example: "chat", "mqtt"
	Protocol string

	// Origin is the allowed origin for connections.
	// Optional. Example: "https://example.com"
	Origin string

	// AuthRequired indicates if authentication is required.
	// Optional. Default: false
	AuthRequired bool

	// MaxMessageSize is the maximum message size in bytes.
	// Optional.
	MaxMessageSize int
}

func (w *WebSocket) NodeType() string { return "websocket" }

func (w *WebSocket) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropURL: w.URL,
	}
}

func (w *WebSocket) Properties() map[string]any {
	props := w.IdentifyingProperties()
	if w.Protocol != "" {
		props[graphrag.PropProtocol] = w.Protocol
	}
	if w.Origin != "" {
		props["origin"] = w.Origin
	}
	props["auth_required"] = w.AuthRequired
	if w.MaxMessageSize > 0 {
		props["max_message_size"] = w.MaxMessageSize
	}
	return props
}

func (w *WebSocket) ParentRef() *NodeRef      { return nil }
func (w *WebSocket) RelationshipType() string { return "" }

// GraphQLSchema represents a GraphQL API schema.
// Schemas define the types, queries, and mutations available.
//
// Example:
//
//	schema := &GraphQLSchema{
//	    Endpoint:    "https://api.example.com/graphql",
//	    Version:     "1.0",
//	    Description: "Main GraphQL API schema",
//	}
//
// Identifying Properties:
//   - endpoint (required): The GraphQL endpoint URL
//
// Relationships:
//   - None (root node)
//   - Children: GraphQLQuery, GraphQLMutation nodes
type GraphQLSchema struct {
	// Endpoint is the GraphQL API endpoint URL.
	// This is an identifying property and is required.
	Endpoint string

	// Version is the schema version.
	// Optional.
	Version string

	// Description is a description of the schema.
	// Optional.
	Description string

	// HasIntrospection indicates if introspection is enabled.
	// Optional. Default: true
	HasIntrospection bool

	// HasSubscriptions indicates if subscriptions are supported.
	// Optional. Default: false
	HasSubscriptions bool
}

func (g *GraphQLSchema) NodeType() string { return "graphql_schema" }

func (g *GraphQLSchema) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint": g.Endpoint,
	}
}

func (g *GraphQLSchema) Properties() map[string]any {
	props := g.IdentifyingProperties()
	if g.Version != "" {
		props["version"] = g.Version
	}
	if g.Description != "" {
		props[graphrag.PropDescription] = g.Description
	}
	props["has_introspection"] = g.HasIntrospection
	props["has_subscriptions"] = g.HasSubscriptions
	return props
}

func (g *GraphQLSchema) ParentRef() *NodeRef      { return nil }
func (g *GraphQLSchema) RelationshipType() string { return "" }

// GraphQLQuery represents a GraphQL query operation.
// Queries fetch data from the GraphQL API.
//
// Example:
//
//	query := &GraphQLQuery{
//	    SchemaID:    "https://api.example.com/graphql",
//	    Name:        "getUser",
//	    Description: "Fetch user by ID",
//	    ReturnType:  "User",
//	}
//
// Identifying Properties:
//   - schema_id (required): The schema this query belongs to
//   - name (required): Query name
//
// Relationships:
//   - Parent: GraphQLSchema node (via HAS_QUERY relationship)
type GraphQLQuery struct {
	// SchemaID is the identifier of the parent schema.
	// This is an identifying property and is required.
	SchemaID string

	// Name is the query name.
	// This is an identifying property and is required.
	Name string

	// Description is a description of what the query does.
	// Optional.
	Description string

	// ReturnType is the GraphQL return type.
	// Optional. Example: "User", "[Post]", "String!"
	ReturnType string

	// Deprecated indicates if this query is deprecated.
	// Optional. Default: false
	Deprecated bool
}

func (g *GraphQLQuery) NodeType() string { return "graphql_query" }

func (g *GraphQLQuery) IdentifyingProperties() map[string]any {
	return map[string]any{
		"schema_id":       g.SchemaID,
		graphrag.PropName: g.Name,
	}
}

func (g *GraphQLQuery) Properties() map[string]any {
	props := g.IdentifyingProperties()
	if g.Description != "" {
		props[graphrag.PropDescription] = g.Description
	}
	if g.ReturnType != "" {
		props["return_type"] = g.ReturnType
	}
	props["deprecated"] = g.Deprecated
	return props
}

func (g *GraphQLQuery) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "graphql_schema",
		Properties: map[string]any{
			"endpoint": g.SchemaID,
		},
	}
}

func (g *GraphQLQuery) RelationshipType() string { return "HAS_QUERY" }

// GraphQLMutation represents a GraphQL mutation operation.
// Mutations modify data on the server.
//
// Example:
//
//	mutation := &GraphQLMutation{
//	    SchemaID:    "https://api.example.com/graphql",
//	    Name:        "createUser",
//	    Description: "Create a new user account",
//	    ReturnType:  "User",
//	}
//
// Identifying Properties:
//   - schema_id (required): The schema this mutation belongs to
//   - name (required): Mutation name
//
// Relationships:
//   - Parent: GraphQLSchema node (via HAS_MUTATION relationship)
type GraphQLMutation struct {
	// SchemaID is the identifier of the parent schema.
	// This is an identifying property and is required.
	SchemaID string

	// Name is the mutation name.
	// This is an identifying property and is required.
	Name string

	// Description is a description of what the mutation does.
	// Optional.
	Description string

	// ReturnType is the GraphQL return type.
	// Optional. Example: "User", "Boolean!", "CreateUserPayload"
	ReturnType string

	// Deprecated indicates if this mutation is deprecated.
	// Optional. Default: false
	Deprecated bool
}

func (g *GraphQLMutation) NodeType() string { return "graphql_mutation" }

func (g *GraphQLMutation) IdentifyingProperties() map[string]any {
	return map[string]any{
		"schema_id":       g.SchemaID,
		graphrag.PropName: g.Name,
	}
}

func (g *GraphQLMutation) Properties() map[string]any {
	props := g.IdentifyingProperties()
	if g.Description != "" {
		props[graphrag.PropDescription] = g.Description
	}
	if g.ReturnType != "" {
		props["return_type"] = g.ReturnType
	}
	props["deprecated"] = g.Deprecated
	return props
}

func (g *GraphQLMutation) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "graphql_schema",
		Properties: map[string]any{
			"endpoint": g.SchemaID,
		},
	}
}

func (g *GraphQLMutation) RelationshipType() string { return "HAS_MUTATION" }

// RESTResource represents a REST API resource.
// Resources are entities that can be manipulated via HTTP methods.
//
// Example:
//
//	resource := &RESTResource{
//	    APIID:       "api.example.com",
//	    Name:        "users",
//	    BasePath:    "/api/v1/users",
//	    Description: "User management resource",
//	}
//
// Identifying Properties:
//   - api_id (required): The API this resource belongs to
//   - name (required): Resource name
//
// Relationships:
//   - Parent: API node (via HAS_RESOURCE relationship)
type RESTResource struct {
	// APIID is the identifier of the parent API.
	// This is an identifying property and is required.
	APIID string

	// Name is the resource name.
	// This is an identifying property and is required.
	// Example: "users", "posts", "comments"
	Name string

	// BasePath is the base URL path for this resource.
	// Optional. Example: "/api/v1/users"
	BasePath string

	// Description is a description of the resource.
	// Optional.
	Description string

	// SupportsCRUD indicates if this resource supports standard CRUD operations.
	// Optional. Default: false
	SupportsCRUD bool
}

func (r *RESTResource) NodeType() string { return "rest_resource" }

func (r *RESTResource) IdentifyingProperties() map[string]any {
	return map[string]any{
		"api_id":          r.APIID,
		graphrag.PropName: r.Name,
	}
}

func (r *RESTResource) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.BasePath != "" {
		props["base_path"] = r.BasePath
	}
	if r.Description != "" {
		props[graphrag.PropDescription] = r.Description
	}
	props["supports_crud"] = r.SupportsCRUD
	return props
}

func (r *RESTResource) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: graphrag.NodeTypeApi,
		Properties: map[string]any{
			"base_url": r.APIID,
		},
	}
}

func (r *RESTResource) RelationshipType() string { return "HAS_RESOURCE" }

// HTTPMethod represents an HTTP method/verb.
// Methods define the type of operation being performed.
//
// Example:
//
//	method := &HTTPMethod{
//	    Method:      "POST",
//	    Description: "Create a new resource",
//	    Idempotent:  false,
//	    Safe:        false,
//	}
//
// Identifying Properties:
//   - method (required): HTTP method name
//
// Relationships:
//   - None (root node)
type HTTPMethod struct {
	// Method is the HTTP method name.
	// This is an identifying property and is required.
	// Common values: "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"
	Method string

	// Description is a description of the method's semantics.
	// Optional.
	Description string

	// Idempotent indicates if the method is idempotent.
	// Optional. Idempotent methods can be called multiple times with the same result.
	Idempotent bool

	// Safe indicates if the method is safe (read-only).
	// Optional. Safe methods don't modify resources.
	Safe bool
}

func (h *HTTPMethod) NodeType() string { return "http_method" }

func (h *HTTPMethod) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropMethod: h.Method,
	}
}

func (h *HTTPMethod) Properties() map[string]any {
	props := h.IdentifyingProperties()
	if h.Description != "" {
		props[graphrag.PropDescription] = h.Description
	}
	props["idempotent"] = h.Idempotent
	props["safe"] = h.Safe
	return props
}

func (h *HTTPMethod) ParentRef() *NodeRef      { return nil }
func (h *HTTPMethod) RelationshipType() string { return "" }

// ContentType represents a MIME content type.
// Content types specify the format of HTTP message bodies.
//
// Example:
//
//	ct := &ContentType{
//	    MimeType:    "application/json",
//	    Description: "JSON data format",
//	}
//
// Identifying Properties:
//   - mime_type (required): MIME type string
//
// Relationships:
//   - None (root node)
type ContentType struct {
	// MimeType is the MIME type string.
	// This is an identifying property and is required.
	// Example: "application/json", "text/html", "image/png"
	MimeType string

	// Description is a description of the content type.
	// Optional.
	Description string

	// Extensions are common file extensions for this type.
	// Optional. Example: [".json", ".js"]
	Extensions []string

	// Charset is the character encoding.
	// Optional. Example: "utf-8"
	Charset string
}

func (c *ContentType) NodeType() string { return "content_type" }

func (c *ContentType) IdentifyingProperties() map[string]any {
	return map[string]any{
		"mime_type": c.MimeType,
	}
}

func (c *ContentType) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	if len(c.Extensions) > 0 {
		props["extensions"] = c.Extensions
	}
	if c.Charset != "" {
		props["charset"] = c.Charset
	}
	return props
}

func (c *ContentType) ParentRef() *NodeRef      { return nil }
func (c *ContentType) RelationshipType() string { return "" }

// CORSPolicy represents a Cross-Origin Resource Sharing policy.
// CORS policies control which origins can access web resources.
//
// Example:
//
//	cors := &CORSPolicy{
//	    Origin:           "https://example.com",
//	    AllowedMethods:   []string{"GET", "POST"},
//	    AllowCredentials: true,
//	}
//
// Identifying Properties:
//   - origin (required): Allowed origin
//
// Relationships:
//   - None (root node)
type CORSPolicy struct {
	// Origin is the allowed origin.
	// This is an identifying property and is required.
	// Example: "https://example.com", "*"
	Origin string

	// AllowedMethods are the HTTP methods allowed for CORS requests.
	// Optional. Example: ["GET", "POST", "PUT"]
	AllowedMethods []string

	// AllowedHeaders are the headers allowed in CORS requests.
	// Optional. Example: ["Content-Type", "Authorization"]
	AllowedHeaders []string

	// ExposedHeaders are the headers exposed to the client.
	// Optional.
	ExposedHeaders []string

	// MaxAge is the preflight cache duration in seconds.
	// Optional.
	MaxAge int

	// AllowCredentials indicates if credentials are allowed.
	// Optional. Default: false
	AllowCredentials bool
}

func (c *CORSPolicy) NodeType() string { return "cors_policy" }

func (c *CORSPolicy) IdentifyingProperties() map[string]any {
	return map[string]any{
		"origin": c.Origin,
	}
}

func (c *CORSPolicy) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if len(c.AllowedMethods) > 0 {
		props["allowed_methods"] = c.AllowedMethods
	}
	if len(c.AllowedHeaders) > 0 {
		props["allowed_headers"] = c.AllowedHeaders
	}
	if len(c.ExposedHeaders) > 0 {
		props["exposed_headers"] = c.ExposedHeaders
	}
	if c.MaxAge > 0 {
		props["max_age"] = c.MaxAge
	}
	props["allow_credentials"] = c.AllowCredentials
	return props
}

func (c *CORSPolicy) ParentRef() *NodeRef      { return nil }
func (c *CORSPolicy) RelationshipType() string { return "" }

// RateLimit represents an API rate limiting policy.
// Rate limits control how many requests can be made in a time window.
//
// Example:
//
//	limit := &RateLimit{
//	    Name:     "api-rate-limit",
//	    Requests: 1000,
//	    Window:   3600,
//	    Scope:    "user",
//	}
//
// Identifying Properties:
//   - name (required): Rate limit policy name
//
// Relationships:
//   - None (root node)
type RateLimit struct {
	// Name is the rate limit policy name.
	// This is an identifying property and is required.
	Name string

	// Requests is the maximum number of requests allowed.
	// Optional.
	Requests int

	// Window is the time window in seconds.
	// Optional. Example: 3600 (1 hour), 86400 (1 day)
	Window int

	// Scope is the scope of the rate limit.
	// Optional. Common values: "global", "user", "ip", "api_key"
	Scope string

	// Description is a description of the rate limit policy.
	// Optional.
	Description string
}

func (r *RateLimit) NodeType() string { return "rate_limit" }

func (r *RateLimit) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

func (r *RateLimit) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.Requests > 0 {
		props["requests"] = r.Requests
	}
	if r.Window > 0 {
		props["window"] = r.Window
	}
	if r.Scope != "" {
		props["scope"] = r.Scope
	}
	if r.Description != "" {
		props[graphrag.PropDescription] = r.Description
	}
	return props
}

func (r *RateLimit) ParentRef() *NodeRef      { return nil }
func (r *RateLimit) RelationshipType() string { return "" }

// RequestBody represents the body of an HTTP request.
// Request bodies carry data to be processed by the server.
//
// Example:
//
//	body := &RequestBody{
//	    EndpointID:  "/api/v1/users:POST",
//	    ContentType: "application/json",
//	    Schema:      `{"type": "object", "properties": {...}}`,
//	    Required:    true,
//	}
//
// Identifying Properties:
//   - endpoint_id (required): The endpoint this body belongs to
//
// Relationships:
//   - Parent: APIEndpoint node (via HAS_REQUEST_BODY relationship)
type RequestBody struct {
	// EndpointID is the identifier of the parent endpoint.
	// This is an identifying property and is required.
	EndpointID string

	// ContentType is the expected content type.
	// Optional. Example: "application/json", "multipart/form-data"
	ContentType string

	// Schema is a JSON schema or description of the body structure.
	// Optional.
	Schema string

	// Required indicates if the body is required.
	// Optional. Default: false
	Required bool

	// MaxSize is the maximum body size in bytes.
	// Optional.
	MaxSize int

	// Example is an example request body.
	// Optional.
	Example string
}

func (r *RequestBody) NodeType() string { return "request_body" }

func (r *RequestBody) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint_id": r.EndpointID,
	}
}

func (r *RequestBody) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.ContentType != "" {
		props["content_type"] = r.ContentType
	}
	if r.Schema != "" {
		props["schema"] = r.Schema
	}
	props["required"] = r.Required
	if r.MaxSize > 0 {
		props["max_size"] = r.MaxSize
	}
	if r.Example != "" {
		props["example"] = r.Example
	}
	return props
}

func (r *RequestBody) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "api_endpoint",
		Properties: map[string]any{
			"id": r.EndpointID,
		},
	}
}

func (r *RequestBody) RelationshipType() string { return "HAS_REQUEST_BODY" }
