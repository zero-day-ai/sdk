package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Model Context Protocol (MCP) Types
// These types represent the MCP protocol components used for AI agent communication.

// MCPServer represents an MCP server providing tools, resources, and prompts.
// An MCP server is identified by its name.
// MCPServer is a root-level node with no parent relationships.
//
// Example:
//
//	server := &MCPServer{
//	    Name:           "filesystem-server",
//	    Transport:      "stdio",
//	    Version:        "1.0.0",
//	    Description:    "MCP server providing filesystem access tools",
//	    Capabilities:   []string{"tools", "resources"},
//	    ToolsCount:     5,
//	    ResourcesCount: 10,
//	    PromptsCount:   2,
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the MCP server
//
// Relationships:
//   - None (root node)
//   - Children: MCPTool, MCPResource, MCPPrompt nodes
type MCPServer struct {
	// Name is the unique identifier for this MCP server.
	// This is an identifying property and is required.
	Name string

	// Transport is the transport mechanism used by the server.
	// Optional. Common values: "stdio", "sse", "http"
	Transport string

	// Version is the MCP protocol version or server version.
	// Optional. Example: "1.0.0", "2024-11-05"
	Version string

	// Description describes the server's purpose and capabilities.
	// Optional.
	Description string

	// Capabilities is a list of MCP capabilities supported.
	// Optional. Example: ["tools", "resources", "prompts", "sampling"]
	Capabilities []string

	// ToolsCount is the number of tools provided by this server.
	// Optional.
	ToolsCount int

	// ResourcesCount is the number of resources provided.
	// Optional.
	ResourcesCount int

	// PromptsCount is the number of prompts provided.
	// Optional.
	PromptsCount int
}

// NodeType returns the canonical node type for MCPServer nodes.
// Implements GraphNode interface.
func (m *MCPServer) NodeType() string {
	return graphrag.NodeTypeMCPServer
}

// IdentifyingProperties returns the properties that uniquely identify this server.
// For MCPServer nodes, only name is identifying.
// Implements GraphNode interface.
func (m *MCPServer) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPServer node.
// Implements GraphNode interface.
func (m *MCPServer) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.Transport != "" {
		props[graphrag.PropTransport] = m.Transport
	}
	if m.Version != "" {
		props["version"] = m.Version
	}
	if m.Description != "" {
		props[graphrag.PropDescription] = m.Description
	}
	if len(m.Capabilities) > 0 {
		props["capabilities"] = m.Capabilities
	}
	if m.ToolsCount > 0 {
		props["tools_count"] = m.ToolsCount
	}
	if m.ResourcesCount > 0 {
		props["resources_count"] = m.ResourcesCount
	}
	if m.PromptsCount > 0 {
		props["prompts_count"] = m.PromptsCount
	}

	return props
}

// ParentRef returns nil because MCPServer is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPServer) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPServer is a root node.
// Implements GraphNode interface.
func (m *MCPServer) RelationshipType() string {
	return ""
}

// MCPTool represents a tool provided by an MCP server.
// An MCPTool is identified by its server ID and name.
// MCPTool nodes are children of MCPServer nodes.
//
// Example:
//
//	tool := &MCPTool{
//	    ServerID:    "filesystem-server",
//	    Name:        "read_file",
//	    Description: "Read contents of a file",
//	    InputSchema: map[string]any{
//	        "type": "object",
//	        "properties": map[string]any{
//	            "path": map[string]any{"type": "string"},
//	        },
//	    },
//	    Annotations: map[string]any{"safe": true},
//	}
//
// Identifying Properties:
//   - server_id (required): Parent MCP server name
//   - name (required): Tool name
//
// Relationships:
//   - Parent: MCPServer node (via PROVIDES_TOOL relationship)
type MCPTool struct {
	// ServerID is the identifier of the parent MCP server.
	// This is an identifying property and is required.
	ServerID string

	// Name is the unique name of this tool within the server.
	// This is an identifying property and is required.
	Name string

	// Description describes what the tool does.
	// Optional.
	Description string

	// InputSchema is the JSON schema defining the tool's input parameters.
	// Optional. Typically a map representing a JSON schema object.
	InputSchema map[string]any

	// Annotations contains additional metadata about the tool.
	// Optional. Example: {"safe": true, "requiresAuth": false}
	Annotations map[string]any
}

// NodeType returns the canonical node type for MCPTool nodes.
// Implements GraphNode interface.
func (m *MCPTool) NodeType() string {
	return graphrag.NodeTypeMCPTool
}

// IdentifyingProperties returns the properties that uniquely identify this tool.
// For MCPTool nodes, server_id and name are both identifying.
// Implements GraphNode interface.
func (m *MCPTool) IdentifyingProperties() map[string]any {
	return map[string]any{
		"server_id":       m.ServerID,
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPTool node.
// Implements GraphNode interface.
func (m *MCPTool) Properties() map[string]any {
	props := map[string]any{
		"server_id":       m.ServerID,
		graphrag.PropName: m.Name,
	}

	if m.Description != "" {
		props[graphrag.PropDescription] = m.Description
	}
	if m.InputSchema != nil {
		props["input_schema"] = m.InputSchema
	}
	if m.Annotations != nil {
		props["annotations"] = m.Annotations
	}

	return props
}

// ParentRef returns a reference to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPTool) ParentRef() *NodeRef {
	if m.ServerID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeMCPServer,
		Properties: map[string]any{
			graphrag.PropName: m.ServerID,
		},
	}
}

// RelationshipType returns the relationship type to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPTool) RelationshipType() string {
	return graphrag.RelTypeProvidesTool
}

// MCPResource represents a resource provided by an MCP server.
// An MCPResource is identified by its server ID and URI.
// MCPResource nodes are children of MCPServer nodes.
//
// Example:
//
//	resource := &MCPResource{
//	    ServerID:    "filesystem-server",
//	    URI:         "file:///home/user/project/README.md",
//	    Name:        "Project README",
//	    Description: "Main project documentation",
//	    MimeType:    "text/markdown",
//	}
//
// Identifying Properties:
//   - server_id (required): Parent MCP server name
//   - uri (required): Resource URI
//
// Relationships:
//   - Parent: MCPServer node (via PROVIDES_RESOURCE relationship)
type MCPResource struct {
	// ServerID is the identifier of the parent MCP server.
	// This is an identifying property and is required.
	ServerID string

	// URI is the unique resource identifier.
	// This is an identifying property and is required.
	// Example: "file:///path/to/file", "http://api.example.com/data"
	URI string

	// Name is the human-readable name of the resource.
	// Optional.
	Name string

	// Description describes the resource.
	// Optional.
	Description string

	// MimeType is the MIME type of the resource content.
	// Optional. Example: "text/plain", "application/json", "image/png"
	MimeType string
}

// NodeType returns the canonical node type for MCPResource nodes.
// Implements GraphNode interface.
func (m *MCPResource) NodeType() string {
	return graphrag.NodeTypeMCPResource
}

// IdentifyingProperties returns the properties that uniquely identify this resource.
// For MCPResource nodes, server_id and uri are both identifying.
// Implements GraphNode interface.
func (m *MCPResource) IdentifyingProperties() map[string]any {
	return map[string]any{
		"server_id": m.ServerID,
		"uri":       m.URI,
	}
}

// Properties returns all properties to set on the MCPResource node.
// Implements GraphNode interface.
func (m *MCPResource) Properties() map[string]any {
	props := map[string]any{
		"server_id": m.ServerID,
		"uri":       m.URI,
	}

	if m.Name != "" {
		props[graphrag.PropName] = m.Name
	}
	if m.Description != "" {
		props[graphrag.PropDescription] = m.Description
	}
	if m.MimeType != "" {
		props["mime_type"] = m.MimeType
	}

	return props
}

// ParentRef returns a reference to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPResource) ParentRef() *NodeRef {
	if m.ServerID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeMCPServer,
		Properties: map[string]any{
			graphrag.PropName: m.ServerID,
		},
	}
}

// RelationshipType returns the relationship type to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPResource) RelationshipType() string {
	return graphrag.RelTypeProvidesResource
}

// MCPPrompt represents a prompt template provided by an MCP server.
// An MCPPrompt is identified by its server ID and name.
// MCPPrompt nodes are children of MCPServer nodes.
//
// Example:
//
//	prompt := &MCPPrompt{
//	    ServerID:    "analysis-server",
//	    Name:        "analyze-code",
//	    Description: "Analyze code for security vulnerabilities",
//	    Arguments:   []string{"code", "language", "rules"},
//	}
//
// Identifying Properties:
//   - server_id (required): Parent MCP server name
//   - name (required): Prompt name
//
// Relationships:
//   - Parent: MCPServer node (via PROVIDES_PROMPT relationship)
type MCPPrompt struct {
	// ServerID is the identifier of the parent MCP server.
	// This is an identifying property and is required.
	ServerID string

	// Name is the unique name of this prompt template.
	// This is an identifying property and is required.
	Name string

	// Description describes what the prompt does.
	// Optional.
	Description string

	// Arguments is a list of argument names the prompt accepts.
	// Optional. Example: ["input", "context", "format"]
	Arguments []string
}

// NodeType returns the canonical node type for MCPPrompt nodes.
// Implements GraphNode interface.
func (m *MCPPrompt) NodeType() string {
	return graphrag.NodeTypeMCPPrompt
}

// IdentifyingProperties returns the properties that uniquely identify this prompt.
// For MCPPrompt nodes, server_id and name are both identifying.
// Implements GraphNode interface.
func (m *MCPPrompt) IdentifyingProperties() map[string]any {
	return map[string]any{
		"server_id":       m.ServerID,
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPPrompt node.
// Implements GraphNode interface.
func (m *MCPPrompt) Properties() map[string]any {
	props := map[string]any{
		"server_id":       m.ServerID,
		graphrag.PropName: m.Name,
	}

	if m.Description != "" {
		props[graphrag.PropDescription] = m.Description
	}
	if len(m.Arguments) > 0 {
		props["arguments"] = m.Arguments
	}

	return props
}

// ParentRef returns a reference to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPPrompt) ParentRef() *NodeRef {
	if m.ServerID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeMCPServer,
		Properties: map[string]any{
			graphrag.PropName: m.ServerID,
		},
	}
}

// RelationshipType returns the relationship type to the parent MCPServer node.
// Implements GraphNode interface.
func (m *MCPPrompt) RelationshipType() string {
	return graphrag.RelTypeProvidesPrompt
}

// MCPClient represents an MCP client connecting to servers.
// An MCPClient is identified by its name.
// MCPClient is a root-level node with no parent relationships.
//
// Example:
//
//	client := &MCPClient{
//	    Name:             "claude-desktop",
//	    Transport:        "stdio",
//	    ServersConnected: 3,
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the MCP client
//
// Relationships:
//   - None (root node)
type MCPClient struct {
	// Name is the unique identifier for this MCP client.
	// This is an identifying property and is required.
	Name string

	// Transport is the transport mechanism used by the client.
	// Optional. Common values: "stdio", "sse", "http"
	Transport string

	// ServersConnected is the number of servers this client is connected to.
	// Optional.
	ServersConnected int
}

// NodeType returns the canonical node type for MCPClient nodes.
// Implements GraphNode interface.
func (m *MCPClient) NodeType() string {
	return graphrag.NodeTypeMCPClient
}

// IdentifyingProperties returns the properties that uniquely identify this client.
// For MCPClient nodes, only name is identifying.
// Implements GraphNode interface.
func (m *MCPClient) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPClient node.
// Implements GraphNode interface.
func (m *MCPClient) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.Transport != "" {
		props[graphrag.PropTransport] = m.Transport
	}
	if m.ServersConnected > 0 {
		props["servers_connected"] = m.ServersConnected
	}

	return props
}

// ParentRef returns nil because MCPClient is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPClient) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPClient is a root node.
// Implements GraphNode interface.
func (m *MCPClient) RelationshipType() string {
	return ""
}

// MCPTransport represents an MCP transport configuration.
// An MCPTransport is identified by its type and configuration.
// MCPTransport is a root-level node with no parent relationships.
//
// Example:
//
//	transport := &MCPTransport{
//	    Type: "stdio",
//	    Config: map[string]any{
//	        "command": "/usr/local/bin/mcp-server",
//	        "args":    []string{"--verbose"},
//	        "env":     map[string]string{"LOG_LEVEL": "debug"},
//	    },
//	}
//
// Identifying Properties:
//   - type (required): Transport type (stdio, sse, http)
//
// Relationships:
//   - None (root node)
type MCPTransport struct {
	// Type is the transport mechanism type.
	// This is an identifying property and is required.
	// Common values: "stdio", "sse", "http"
	Type string

	// Config contains transport-specific configuration.
	// Optional. Structure varies by transport type:
	// - stdio: {"command": string, "args": []string, "env": map[string]string}
	// - sse: {"url": string, "headers": map[string]string}
	// - http: {"url": string, "method": string, "headers": map[string]string}
	Config map[string]any
}

// NodeType returns the canonical node type for MCPTransport nodes.
// Implements GraphNode interface.
func (m *MCPTransport) NodeType() string {
	return graphrag.NodeTypeMCPTransport
}

// IdentifyingProperties returns the properties that uniquely identify this transport.
// For MCPTransport nodes, only type is identifying.
// Implements GraphNode interface.
func (m *MCPTransport) IdentifyingProperties() map[string]any {
	return map[string]any{
		"type": m.Type,
	}
}

// Properties returns all properties to set on the MCPTransport node.
// Implements GraphNode interface.
func (m *MCPTransport) Properties() map[string]any {
	props := map[string]any{
		"type": m.Type,
	}

	if m.Config != nil {
		props["config"] = m.Config
	}

	return props
}

// ParentRef returns nil because MCPTransport is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPTransport) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPTransport is a root node.
// Implements GraphNode interface.
func (m *MCPTransport) RelationshipType() string {
	return ""
}

// MCPCapability represents an MCP capability supported by a server or client.
// An MCPCapability is identified by its name.
// MCPCapability is a root-level node with no parent relationships.
//
// Example:
//
//	capability := &MCPCapability{
//	    Name:    "tools",
//	    Version: "2024-11-05",
//	    Enabled: true,
//	}
//
// Identifying Properties:
//   - name (required): Capability name
//
// Relationships:
//   - None (root node)
type MCPCapability struct {
	// Name is the capability identifier.
	// This is an identifying property and is required.
	// Common values: "tools", "resources", "prompts", "sampling", "logging"
	Name string

	// Version is the capability version.
	// Optional. Example: "2024-11-05", "1.0"
	Version string

	// Enabled indicates if the capability is currently enabled.
	// Optional. Default: true
	Enabled bool
}

// NodeType returns the canonical node type for MCPCapability nodes.
// Implements GraphNode interface.
func (m *MCPCapability) NodeType() string {
	return graphrag.NodeTypeMCPCapability
}

// IdentifyingProperties returns the properties that uniquely identify this capability.
// For MCPCapability nodes, only name is identifying.
// Implements GraphNode interface.
func (m *MCPCapability) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPCapability node.
// Implements GraphNode interface.
func (m *MCPCapability) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.Version != "" {
		props["version"] = m.Version
	}
	// Always include enabled field as it's a boolean with meaningful false value
	props["enabled"] = m.Enabled

	return props
}

// ParentRef returns nil because MCPCapability is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPCapability) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPCapability is a root node.
// Implements GraphNode interface.
func (m *MCPCapability) RelationshipType() string {
	return ""
}

// MCPSampling represents an MCP sampling configuration for LLM calls.
// An MCPSampling is identified by its name.
// MCPSampling is a root-level node with no parent relationships.
//
// Example:
//
//	sampling := &MCPSampling{
//	    Name:          "default-sampling",
//	    MaxTokens:     1000,
//	    Temperature:   0.7,
//	    TopP:          0.95,
//	    StopSequences: []string{"\n\n", "###"},
//	}
//
// Identifying Properties:
//   - name (required): Sampling configuration name
//
// Relationships:
//   - None (root node)
type MCPSampling struct {
	// Name is the unique identifier for this sampling configuration.
	// This is an identifying property and is required.
	Name string

	// MaxTokens is the maximum number of tokens to generate.
	// Optional. Example: 1000, 4096
	MaxTokens int

	// Temperature controls randomness (0.0 = deterministic, 1.0+ = creative).
	// Optional. Typical range: 0.0 to 2.0
	Temperature float64

	// TopP controls nucleus sampling (cumulative probability cutoff).
	// Optional. Typical range: 0.0 to 1.0
	TopP float64

	// StopSequences are sequences that stop generation when encountered.
	// Optional. Example: ["\n\n", "###", "END"]
	StopSequences []string
}

// NodeType returns the canonical node type for MCPSampling nodes.
// Implements GraphNode interface.
func (m *MCPSampling) NodeType() string {
	return graphrag.NodeTypeMCPSampling
}

// IdentifyingProperties returns the properties that uniquely identify this sampling config.
// For MCPSampling nodes, only name is identifying.
// Implements GraphNode interface.
func (m *MCPSampling) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPSampling node.
// Implements GraphNode interface.
func (m *MCPSampling) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.MaxTokens > 0 {
		props["max_tokens"] = m.MaxTokens
	}
	if m.Temperature > 0 {
		props["temperature"] = m.Temperature
	}
	if m.TopP > 0 {
		props["top_p"] = m.TopP
	}
	if len(m.StopSequences) > 0 {
		props["stop_sequences"] = m.StopSequences
	}

	return props
}

// ParentRef returns nil because MCPSampling is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPSampling) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPSampling is a root node.
// Implements GraphNode interface.
func (m *MCPSampling) RelationshipType() string {
	return ""
}

// MCPRoots represents MCP root directory configurations.
// An MCPRoots is identified by its name.
// MCPRoots is a root-level node with no parent relationships.
//
// Example:
//
//	roots := &MCPRoots{
//	    Name:        "project-roots",
//	    URI:         "file:///home/user/projects",
//	    Description: "Project directory roots for file access",
//	}
//
// Identifying Properties:
//   - name (required): Roots configuration name
//
// Relationships:
//   - None (root node)
type MCPRoots struct {
	// Name is the unique identifier for this roots configuration.
	// This is an identifying property and is required.
	Name string

	// URI is the root URI or path.
	// Optional. Example: "file:///home/user/projects", "/"
	URI string

	// Description describes the purpose of this roots configuration.
	// Optional.
	Description string
}

// NodeType returns the canonical node type for MCPRoots nodes.
// Implements GraphNode interface.
func (m *MCPRoots) NodeType() string {
	return graphrag.NodeTypeMCPRoots
}

// IdentifyingProperties returns the properties that uniquely identify this roots config.
// For MCPRoots nodes, only name is identifying.
// Implements GraphNode interface.
func (m *MCPRoots) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the MCPRoots node.
// Implements GraphNode interface.
func (m *MCPRoots) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.URI != "" {
		props["uri"] = m.URI
	}
	if m.Description != "" {
		props[graphrag.PropDescription] = m.Description
	}

	return props
}

// ParentRef returns nil because MCPRoots is a root node with no parent.
// Implements GraphNode interface.
func (m *MCPRoots) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MCPRoots is a root node.
// Implements GraphNode interface.
func (m *MCPRoots) RelationshipType() string {
	return ""
}
