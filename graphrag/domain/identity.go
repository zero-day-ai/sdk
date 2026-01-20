package domain

import "github.com/zero-day-ai/sdk/graphrag"

// User represents a user account in a system.
// Users authenticate and access resources with specific permissions.
//
// Example:
//
//	user := &User{
//	    ID:       "user-12345",
//	    Username: "john.doe",
//	    Email:    "john@example.com",
//	    Active:   true,
//	}
//
// Identifying Properties:
//   - id (required): Unique user identifier
//
// Relationships:
//   - None (root node)
type User struct {
	// ID is the unique identifier for this user.
	// This is an identifying property and is required.
	ID string

	// Username is the user's login name.
	// Optional. Example: "john.doe", "admin"
	Username string

	// Email is the user's email address.
	// Optional. Example: "john@example.com"
	Email string

	// FullName is the user's full name.
	// Optional. Example: "John Doe"
	FullName string

	// Active indicates if the user account is active.
	// Optional. Default: true
	Active bool

	// CreatedAt is when the user account was created.
	// Optional. Unix timestamp.
	CreatedAt int64

	// LastLogin is when the user last authenticated.
	// Optional. Unix timestamp.
	LastLogin int64
}

func (u *User) NodeType() string { return "user" }

func (u *User) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": u.ID,
	}
}

func (u *User) Properties() map[string]any {
	props := u.IdentifyingProperties()
	if u.Username != "" {
		props["username"] = u.Username
	}
	if u.Email != "" {
		props["email"] = u.Email
	}
	if u.FullName != "" {
		props["full_name"] = u.FullName
	}
	props["active"] = u.Active
	if u.CreatedAt > 0 {
		props["created_at"] = u.CreatedAt
	}
	if u.LastLogin > 0 {
		props["last_login"] = u.LastLogin
	}
	return props
}

func (u *User) ParentRef() *NodeRef      { return nil }
func (u *User) RelationshipType() string { return "" }

// Group represents a collection of users.
// Groups simplify permission management by applying policies to multiple users.
//
// Example:
//
//	group := &Group{
//	    ID:          "group-admin",
//	    Name:        "Administrators",
//	    Description: "System administrators with full access",
//	}
//
// Identifying Properties:
//   - id (required): Unique group identifier
//
// Relationships:
//   - None (root node)
type Group struct {
	// ID is the unique identifier for this group.
	// This is an identifying property and is required.
	ID string

	// Name is the group name.
	// Optional. Example: "Administrators", "Developers", "ReadOnly"
	Name string

	// Description is a description of the group's purpose.
	// Optional.
	Description string

	// Type is the type of group.
	// Optional. Common values: "security", "distribution", "role"
	Type string

	// MemberCount is the number of users in the group.
	// Optional.
	MemberCount int
}

func (g *Group) NodeType() string { return "group" }

func (g *Group) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": g.ID,
	}
}

func (g *Group) Properties() map[string]any {
	props := g.IdentifyingProperties()
	if g.Name != "" {
		props[graphrag.PropName] = g.Name
	}
	if g.Description != "" {
		props[graphrag.PropDescription] = g.Description
	}
	if g.Type != "" {
		props["type"] = g.Type
	}
	if g.MemberCount > 0 {
		props["member_count"] = g.MemberCount
	}
	return props
}

func (g *Group) ParentRef() *NodeRef      { return nil }
func (g *Group) RelationshipType() string { return "" }

// Role represents a named set of permissions.
// Roles define what actions can be performed on which resources.
//
// Example:
//
//	role := &Role{
//	    ID:          "role-editor",
//	    Name:        "Editor",
//	    Description: "Can create and edit content",
//	    Scope:       "application",
//	}
//
// Identifying Properties:
//   - id (required): Unique role identifier
//
// Relationships:
//   - None (root node)
type Role struct {
	// ID is the unique identifier for this role.
	// This is an identifying property and is required.
	ID string

	// Name is the role name.
	// Optional. Example: "Admin", "Editor", "Viewer"
	Name string

	// Description is a description of the role's permissions.
	// Optional.
	Description string

	// Scope is the scope where this role applies.
	// Optional. Common values: "application", "organization", "project"
	Scope string

	// System indicates if this is a system-defined role.
	// Optional. Default: false
	System bool
}

func (r *Role) NodeType() string { return "role" }

func (r *Role) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": r.ID,
	}
}

func (r *Role) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.Name != "" {
		props[graphrag.PropName] = r.Name
	}
	if r.Description != "" {
		props[graphrag.PropDescription] = r.Description
	}
	if r.Scope != "" {
		props["scope"] = r.Scope
	}
	props["system"] = r.System
	return props
}

func (r *Role) ParentRef() *NodeRef      { return nil }
func (r *Role) RelationshipType() string { return "" }

// Permission represents a specific action that can be performed.
// Permissions are the atomic units of access control.
//
// Example:
//
//	perm := &Permission{
//	    ID:          "perm-users-read",
//	    Resource:    "users",
//	    Action:      "read",
//	    Description: "View user information",
//	}
//
// Identifying Properties:
//   - id (required): Unique permission identifier
//
// Relationships:
//   - None (root node)
type Permission struct {
	// ID is the unique identifier for this permission.
	// This is an identifying property and is required.
	ID string

	// Resource is the resource this permission applies to.
	// Optional. Example: "users", "posts", "settings"
	Resource string

	// Action is the action that can be performed.
	// Optional. Common values: "read", "write", "delete", "execute"
	Action string

	// Description is a description of the permission.
	// Optional.
	Description string

	// Effect is whether the permission allows or denies access.
	// Optional. Common values: "allow", "deny"
	Effect string
}

func (p *Permission) NodeType() string { return "permission" }

func (p *Permission) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": p.ID,
	}
}

func (p *Permission) Properties() map[string]any {
	props := p.IdentifyingProperties()
	if p.Resource != "" {
		props["resource"] = p.Resource
	}
	if p.Action != "" {
		props["action"] = p.Action
	}
	if p.Description != "" {
		props[graphrag.PropDescription] = p.Description
	}
	if p.Effect != "" {
		props["effect"] = p.Effect
	}
	return props
}

func (p *Permission) ParentRef() *NodeRef      { return nil }
func (p *Permission) RelationshipType() string { return "" }

// Policy represents an access control policy document.
// Policies define complex rules for granting or denying access.
//
// Example:
//
//	policy := &Policy{
//	    ID:          "policy-s3-read",
//	    Name:        "S3 Read Access",
//	    Description: "Allow reading from S3 buckets",
//	    Type:        "resource",
//	}
//
// Identifying Properties:
//   - id (required): Unique policy identifier
//
// Relationships:
//   - None (root node)
type Policy struct {
	// ID is the unique identifier for this policy.
	// This is an identifying property and is required.
	ID string

	// Name is the policy name.
	// Optional. Example: "S3ReadAccess", "AdminPolicy"
	Name string

	// Description is a description of the policy.
	// Optional.
	Description string

	// Type is the type of policy.
	// Optional. Common values: "resource", "identity", "service"
	Type string

	// Document is the policy document content (JSON, YAML, etc.).
	// Optional.
	Document string

	// Version is the policy version.
	// Optional.
	Version string
}

func (p *Policy) NodeType() string { return "policy" }

func (p *Policy) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": p.ID,
	}
}

func (p *Policy) Properties() map[string]any {
	props := p.IdentifyingProperties()
	if p.Name != "" {
		props[graphrag.PropName] = p.Name
	}
	if p.Description != "" {
		props[graphrag.PropDescription] = p.Description
	}
	if p.Type != "" {
		props["type"] = p.Type
	}
	if p.Document != "" {
		props["document"] = p.Document
	}
	if p.Version != "" {
		props["version"] = p.Version
	}
	return props
}

func (p *Policy) ParentRef() *NodeRef      { return nil }
func (p *Policy) RelationshipType() string { return "" }

// Credential represents authentication credentials.
// Credentials prove identity (passwords, keys, certificates, etc.).
//
// Example:
//
//	cred := &Credential{
//	    ID:        "cred-12345",
//	    Type:      "password",
//	    UserID:    "user-12345",
//	    ExpiresAt: 1735689600,
//	}
//
// Identifying Properties:
//   - id (required): Unique credential identifier
//
// Relationships:
//   - None (root node)
type Credential struct {
	// ID is the unique identifier for this credential.
	// This is an identifying property and is required.
	ID string

	// Type is the credential type.
	// Optional. Common values: "password", "ssh_key", "certificate", "token"
	Type string

	// UserID is the user this credential belongs to.
	// Optional.
	UserID string

	// Description is a description of the credential.
	// Optional.
	Description string

	// ExpiresAt is when the credential expires.
	// Optional. Unix timestamp.
	ExpiresAt int64

	// Active indicates if the credential is active.
	// Optional. Default: true
	Active bool
}

func (c *Credential) NodeType() string { return "credential" }

func (c *Credential) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": c.ID,
	}
}

func (c *Credential) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if c.Type != "" {
		props["type"] = c.Type
	}
	if c.UserID != "" {
		props["user_id"] = c.UserID
	}
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	if c.ExpiresAt > 0 {
		props["expires_at"] = c.ExpiresAt
	}
	props["active"] = c.Active
	return props
}

func (c *Credential) ParentRef() *NodeRef      { return nil }
func (c *Credential) RelationshipType() string { return "" }

// APIKey represents an API key for programmatic access.
// API keys authenticate applications and services.
//
// Example:
//
//	key := &APIKey{
//	    ID:        "key-abcd1234",
//	    Name:      "Production API Key",
//	    Prefix:    "pk_live_",
//	    ExpiresAt: 1735689600,
//	}
//
// Identifying Properties:
//   - id (required): Unique API key identifier
//
// Relationships:
//   - None (root node)
type APIKey struct {
	// ID is the unique identifier for this API key.
	// This is an identifying property and is required.
	ID string

	// Name is a descriptive name for the key.
	// Optional. Example: "Production API Key", "Test Environment Key"
	Name string

	// Prefix is the visible prefix of the key.
	// Optional. Example: "pk_live_", "sk_test_"
	Prefix string

	// UserID is the user who created this key.
	// Optional.
	UserID string

	// Scopes are the scopes/permissions this key has.
	// Optional. Example: ["read", "write"]
	Scopes []string

	// ExpiresAt is when the key expires.
	// Optional. Unix timestamp.
	ExpiresAt int64

	// Active indicates if the key is active.
	// Optional. Default: true
	Active bool

	// LastUsed is when the key was last used.
	// Optional. Unix timestamp.
	LastUsed int64
}

func (a *APIKey) NodeType() string { return "api_key" }

func (a *APIKey) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": a.ID,
	}
}

func (a *APIKey) Properties() map[string]any {
	props := a.IdentifyingProperties()
	if a.Name != "" {
		props[graphrag.PropName] = a.Name
	}
	if a.Prefix != "" {
		props["prefix"] = a.Prefix
	}
	if a.UserID != "" {
		props["user_id"] = a.UserID
	}
	if len(a.Scopes) > 0 {
		props["scopes"] = a.Scopes
	}
	if a.ExpiresAt > 0 {
		props["expires_at"] = a.ExpiresAt
	}
	props["active"] = a.Active
	if a.LastUsed > 0 {
		props["last_used"] = a.LastUsed
	}
	return props
}

func (a *APIKey) ParentRef() *NodeRef      { return nil }
func (a *APIKey) RelationshipType() string { return "" }

// Token represents an authentication or access token.
// Tokens are time-limited credentials used for API access.
//
// Example:
//
//	token := &Token{
//	    ID:        "token-xyz789",
//	    Type:      "access",
//	    UserID:    "user-12345",
//	    ExpiresAt: 1704067200,
//	}
//
// Identifying Properties:
//   - id (required): Unique token identifier
//
// Relationships:
//   - None (root node)
type Token struct {
	// ID is the unique identifier for this token.
	// This is an identifying property and is required.
	ID string

	// Type is the token type.
	// Optional. Common values: "access", "refresh", "id", "session"
	Type string

	// UserID is the user this token belongs to.
	// Optional.
	UserID string

	// Scopes are the scopes/permissions this token grants.
	// Optional. Example: ["read:users", "write:posts"]
	Scopes []string

	// IssuedAt is when the token was issued.
	// Optional. Unix timestamp.
	IssuedAt int64

	// ExpiresAt is when the token expires.
	// Optional. Unix timestamp.
	ExpiresAt int64

	// Revoked indicates if the token has been revoked.
	// Optional. Default: false
	Revoked bool
}

func (t *Token) NodeType() string { return "token" }

func (t *Token) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

func (t *Token) Properties() map[string]any {
	props := t.IdentifyingProperties()
	if t.Type != "" {
		props["type"] = t.Type
	}
	if t.UserID != "" {
		props["user_id"] = t.UserID
	}
	if len(t.Scopes) > 0 {
		props["scopes"] = t.Scopes
	}
	if t.IssuedAt > 0 {
		props["issued_at"] = t.IssuedAt
	}
	if t.ExpiresAt > 0 {
		props["expires_at"] = t.ExpiresAt
	}
	props["revoked"] = t.Revoked
	return props
}

func (t *Token) ParentRef() *NodeRef      { return nil }
func (t *Token) RelationshipType() string { return "" }

// OAuthClient represents an OAuth 2.0 client application.
// OAuth clients request access to resources on behalf of users.
//
// Example:
//
//	client := &OAuthClient{
//	    ClientID:     "client-abc123",
//	    Name:         "Mobile App",
//	    GrantTypes:   []string{"authorization_code", "refresh_token"},
//	    RedirectURIs: []string{"https://app.example.com/callback"},
//	}
//
// Identifying Properties:
//   - client_id (required): Unique client identifier
//
// Relationships:
//   - None (root node)
type OAuthClient struct {
	// ClientID is the unique identifier for this OAuth client.
	// This is an identifying property and is required.
	ClientID string

	// Name is the client application name.
	// Optional. Example: "Mobile App", "Web Dashboard"
	Name string

	// Description is a description of the client.
	// Optional.
	Description string

	// GrantTypes are the OAuth grant types allowed.
	// Optional. Example: ["authorization_code", "client_credentials"]
	GrantTypes []string

	// RedirectURIs are the allowed redirect URIs.
	// Optional. Example: ["https://app.example.com/callback"]
	RedirectURIs []string

	// Scopes are the scopes this client can request.
	// Optional.
	Scopes []string

	// Public indicates if this is a public client (cannot keep secrets).
	// Optional. Default: false
	Public bool
}

func (o *OAuthClient) NodeType() string { return "oauth_client" }

func (o *OAuthClient) IdentifyingProperties() map[string]any {
	return map[string]any{
		"client_id": o.ClientID,
	}
}

func (o *OAuthClient) Properties() map[string]any {
	props := o.IdentifyingProperties()
	if o.Name != "" {
		props[graphrag.PropName] = o.Name
	}
	if o.Description != "" {
		props[graphrag.PropDescription] = o.Description
	}
	if len(o.GrantTypes) > 0 {
		props["grant_types"] = o.GrantTypes
	}
	if len(o.RedirectURIs) > 0 {
		props["redirect_uris"] = o.RedirectURIs
	}
	if len(o.Scopes) > 0 {
		props["scopes"] = o.Scopes
	}
	props["public"] = o.Public
	return props
}

func (o *OAuthClient) ParentRef() *NodeRef      { return nil }
func (o *OAuthClient) RelationshipType() string { return "" }

// OAuthScope represents an OAuth 2.0 scope.
// Scopes define granular permissions for API access.
//
// Example:
//
//	scope := &OAuthScope{
//	    Name:        "read:users",
//	    Description: "Read user information",
//	    Category:    "users",
//	}
//
// Identifying Properties:
//   - name (required): Scope name
//
// Relationships:
//   - None (root node)
type OAuthScope struct {
	// Name is the scope name.
	// This is an identifying property and is required.
	// Example: "read:users", "write:posts", "admin:all"
	Name string

	// Description is a human-readable description of the scope.
	// Optional.
	Description string

	// Category is a category for organizing scopes.
	// Optional. Example: "users", "posts", "admin"
	Category string

	// System indicates if this is a system-defined scope.
	// Optional. Default: false
	System bool
}

func (o *OAuthScope) NodeType() string { return "oauth_scope" }

func (o *OAuthScope) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: o.Name,
	}
}

func (o *OAuthScope) Properties() map[string]any {
	props := o.IdentifyingProperties()
	if o.Description != "" {
		props[graphrag.PropDescription] = o.Description
	}
	if o.Category != "" {
		props[graphrag.PropCategory] = o.Category
	}
	props["system"] = o.System
	return props
}

func (o *OAuthScope) ParentRef() *NodeRef      { return nil }
func (o *OAuthScope) RelationshipType() string { return "" }

// SAMLProvider represents a SAML identity provider.
// SAML providers enable single sign-on (SSO) authentication.
//
// Example:
//
//	provider := &SAMLProvider{
//	    EntityID:   "https://idp.example.com/metadata",
//	    Name:       "Corporate IdP",
//	    SSOUrl:     "https://idp.example.com/sso",
//	    SigningCert: "-----BEGIN CERTIFICATE-----...",
//	}
//
// Identifying Properties:
//   - entity_id (required): Unique SAML entity ID
//
// Relationships:
//   - None (root node)
type SAMLProvider struct {
	// EntityID is the unique SAML entity identifier.
	// This is an identifying property and is required.
	EntityID string

	// Name is the provider name.
	// Optional. Example: "Corporate IdP", "Okta"
	Name string

	// SSOUrl is the single sign-on URL.
	// Optional. Example: "https://idp.example.com/sso"
	SSOUrl string

	// SLOUrl is the single logout URL.
	// Optional.
	SLOUrl string

	// SigningCert is the X.509 certificate for signature verification.
	// Optional.
	SigningCert string

	// EncryptionCert is the X.509 certificate for encryption.
	// Optional.
	EncryptionCert string
}

func (s *SAMLProvider) NodeType() string { return "saml_provider" }

func (s *SAMLProvider) IdentifyingProperties() map[string]any {
	return map[string]any{
		"entity_id": s.EntityID,
	}
}

func (s *SAMLProvider) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Name != "" {
		props[graphrag.PropName] = s.Name
	}
	if s.SSOUrl != "" {
		props["sso_url"] = s.SSOUrl
	}
	if s.SLOUrl != "" {
		props["slo_url"] = s.SLOUrl
	}
	if s.SigningCert != "" {
		props["signing_cert"] = s.SigningCert
	}
	if s.EncryptionCert != "" {
		props["encryption_cert"] = s.EncryptionCert
	}
	return props
}

func (s *SAMLProvider) ParentRef() *NodeRef      { return nil }
func (s *SAMLProvider) RelationshipType() string { return "" }

// IdentityProvider represents a generic identity provider.
// Identity providers authenticate users from external systems.
//
// Example:
//
//	idp := &IdentityProvider{
//	    Name:        "Google",
//	    Type:        "oauth2",
//	    Domain:      "accounts.google.com",
//	}
//
// Identifying Properties:
//   - name (required): Provider name
//
// Relationships:
//   - None (root node)
type IdentityProvider struct {
	// Name is the unique name of the identity provider.
	// This is an identifying property and is required.
	Name string

	// Type is the provider type.
	// Optional. Common values: "oauth2", "saml", "ldap", "oidc"
	Type string

	// Domain is the provider domain.
	// Optional. Example: "accounts.google.com", "login.microsoft.com"
	Domain string

	// Description is a description of the provider.
	// Optional.
	Description string

	// Active indicates if the provider is active.
	// Optional. Default: true
	Active bool
}

func (i *IdentityProvider) NodeType() string { return "identity_provider" }

func (i *IdentityProvider) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: i.Name,
	}
}

func (i *IdentityProvider) Properties() map[string]any {
	props := i.IdentifyingProperties()
	if i.Type != "" {
		props["type"] = i.Type
	}
	if i.Domain != "" {
		props["domain"] = i.Domain
	}
	if i.Description != "" {
		props[graphrag.PropDescription] = i.Description
	}
	props["active"] = i.Active
	return props
}

func (i *IdentityProvider) ParentRef() *NodeRef      { return nil }
func (i *IdentityProvider) RelationshipType() string { return "" }

// ServiceAccount represents a non-human identity for services.
// Service accounts are used by applications and automated processes.
//
// Example:
//
//	sa := &ServiceAccount{
//	    Name:      "backup-service",
//	    Namespace: "production",
//	    Purpose:   "Automated backup operations",
//	}
//
// Identifying Properties:
//   - name (required): Service account name
//   - namespace (required): Namespace or scope
//
// Relationships:
//   - None (root node)
type ServiceAccount struct {
	// Name is the service account name.
	// This is an identifying property and is required.
	Name string

	// Namespace is the namespace or scope.
	// This is an identifying property and is required.
	// Example: "default", "production", "kube-system"
	Namespace string

	// Purpose is a description of what the service account is used for.
	// Optional.
	Purpose string

	// Type is the type of service account.
	// Optional. Common values: "kubernetes", "aws", "gcp", "azure"
	Type string

	// Active indicates if the service account is active.
	// Optional. Default: true
	Active bool
}

func (s *ServiceAccount) NodeType() string { return "service_account" }

func (s *ServiceAccount) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: s.Name,
		"namespace":       s.Namespace,
	}
}

func (s *ServiceAccount) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Purpose != "" {
		props["purpose"] = s.Purpose
	}
	if s.Type != "" {
		props["type"] = s.Type
	}
	props["active"] = s.Active
	return props
}

func (s *ServiceAccount) ParentRef() *NodeRef      { return nil }
func (s *ServiceAccount) RelationshipType() string { return "" }

// Session represents an active user session.
// Sessions track authenticated user state over time.
//
// Example:
//
//	session := &Session{
//	    ID:        "sess-abc123",
//	    UserID:    "user-12345",
//	    CreatedAt: 1704067200,
//	    ExpiresAt: 1704070800,
//	}
//
// Identifying Properties:
//   - id (required): Unique session identifier
//
// Relationships:
//   - None (root node)
type Session struct {
	// ID is the unique identifier for this session.
	// This is an identifying property and is required.
	ID string

	// UserID is the user this session belongs to.
	// Optional.
	UserID string

	// IPAddress is the IP address of the client.
	// Optional.
	IPAddress string

	// UserAgent is the client user agent string.
	// Optional.
	UserAgent string

	// CreatedAt is when the session was created.
	// Optional. Unix timestamp.
	CreatedAt int64

	// ExpiresAt is when the session expires.
	// Optional. Unix timestamp.
	ExpiresAt int64

	// LastActivity is when the session was last active.
	// Optional. Unix timestamp.
	LastActivity int64

	// Active indicates if the session is active.
	// Optional. Default: true
	Active bool
}

func (s *Session) NodeType() string { return "session" }

func (s *Session) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": s.ID,
	}
}

func (s *Session) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.UserID != "" {
		props["user_id"] = s.UserID
	}
	if s.IPAddress != "" {
		props["ip_address"] = s.IPAddress
	}
	if s.UserAgent != "" {
		props["user_agent"] = s.UserAgent
	}
	if s.CreatedAt > 0 {
		props["created_at"] = s.CreatedAt
	}
	if s.ExpiresAt > 0 {
		props["expires_at"] = s.ExpiresAt
	}
	if s.LastActivity > 0 {
		props["last_activity"] = s.LastActivity
	}
	props["active"] = s.Active
	return props
}

func (s *Session) ParentRef() *NodeRef      { return nil }
func (s *Session) RelationshipType() string { return "" }

// AccessKey represents a cloud provider access key.
// Access keys provide programmatic access to cloud resources.
//
// Example:
//
//	key := &AccessKey{
//	    ID:       "AKIAIOSFODNN7EXAMPLE",
//	    UserID:   "user-12345",
//	    Provider: "aws",
//	    Status:   "active",
//	}
//
// Identifying Properties:
//   - id (required): Access key ID
//
// Relationships:
//   - None (root node)
type AccessKey struct {
	// ID is the access key identifier.
	// This is an identifying property and is required.
	// Example: "AKIAIOSFODNN7EXAMPLE" (AWS)
	ID string

	// UserID is the user this key belongs to.
	// Optional.
	UserID string

	// Provider is the cloud provider.
	// Optional. Common values: "aws", "gcp", "azure"
	Provider string

	// Status is the key status.
	// Optional. Common values: "active", "inactive", "deleted"
	Status string

	// CreatedAt is when the key was created.
	// Optional. Unix timestamp.
	CreatedAt int64

	// LastUsed is when the key was last used.
	// Optional. Unix timestamp.
	LastUsed int64
}

func (a *AccessKey) NodeType() string { return "access_key" }

func (a *AccessKey) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": a.ID,
	}
}

func (a *AccessKey) Properties() map[string]any {
	props := a.IdentifyingProperties()
	if a.UserID != "" {
		props["user_id"] = a.UserID
	}
	if a.Provider != "" {
		props["provider"] = a.Provider
	}
	if a.Status != "" {
		props["status"] = a.Status
	}
	if a.CreatedAt > 0 {
		props["created_at"] = a.CreatedAt
	}
	if a.LastUsed > 0 {
		props["last_used"] = a.LastUsed
	}
	return props
}

func (a *AccessKey) ParentRef() *NodeRef      { return nil }
func (a *AccessKey) RelationshipType() string { return "" }

// MFADevice represents a multi-factor authentication device.
// MFA devices provide additional authentication factors beyond passwords.
//
// Example:
//
//	device := &MFADevice{
//	    ID:       "mfa-device-123",
//	    UserID:   "user-12345",
//	    Type:     "totp",
//	    Name:     "Google Authenticator",
//	    Verified: true,
//	}
//
// Identifying Properties:
//   - id (required): Unique device identifier
//
// Relationships:
//   - None (root node)
type MFADevice struct {
	// ID is the unique identifier for this MFA device.
	// This is an identifying property and is required.
	ID string

	// UserID is the user this device belongs to.
	// Optional.
	UserID string

	// Type is the MFA device type.
	// Optional. Common values: "totp", "sms", "email", "hardware", "biometric"
	Type string

	// Name is a user-friendly name for the device.
	// Optional. Example: "Google Authenticator", "YubiKey"
	Name string

	// Verified indicates if the device has been verified.
	// Optional. Default: false
	Verified bool

	// Active indicates if the device is active.
	// Optional. Default: true
	Active bool

	// CreatedAt is when the device was registered.
	// Optional. Unix timestamp.
	CreatedAt int64

	// LastUsed is when the device was last used.
	// Optional. Unix timestamp.
	LastUsed int64
}

func (m *MFADevice) NodeType() string { return "mfa_device" }

func (m *MFADevice) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": m.ID,
	}
}

func (m *MFADevice) Properties() map[string]any {
	props := m.IdentifyingProperties()
	if m.UserID != "" {
		props["user_id"] = m.UserID
	}
	if m.Type != "" {
		props["type"] = m.Type
	}
	if m.Name != "" {
		props[graphrag.PropName] = m.Name
	}
	props["verified"] = m.Verified
	props["active"] = m.Active
	if m.CreatedAt > 0 {
		props["created_at"] = m.CreatedAt
	}
	if m.LastUsed > 0 {
		props["last_used"] = m.LastUsed
	}
	return props
}

func (m *MFADevice) ParentRef() *NodeRef      { return nil }
func (m *MFADevice) RelationshipType() string { return "" }
