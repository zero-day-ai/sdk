package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// Container represents a running container instance in the knowledge graph.
// Containers are runtime instances of container images, typically managed by Docker, Kubernetes, or other orchestrators.
//
// Example:
//
//	container := &Container{
//	    ID:     "abc123def456",
//	    Name:   "web-server-prod",
//	    Image:  "nginx:1.24",
//	    Status: "running",
//	    Ports:  []string{"80/tcp", "443/tcp"},
//	    Environment: map[string]string{"ENV": "production"},
//	    Labels: map[string]string{"app": "web"},
//	}
//
// Identifying Properties:
//   - id (required): The unique container ID
//
// Relationships:
//   - None (root node)
//   - Children: Can be related to ContainerImage, Host nodes
type Container struct {
	// ID is the unique identifier for the container instance.
	// This is the identifying property and is required.
	// Example: "abc123def456", "sha256:a1b2c3..."
	ID string

	// Name is the human-readable name assigned to the container.
	// Optional. Example: "web-server-prod", "db-primary"
	Name string

	// Image is the container image this instance was created from.
	// Optional. Example: "nginx:1.24", "postgres:15-alpine"
	Image string

	// Status represents the current state of the container.
	// Optional. Common values: "running", "stopped", "paused", "exited"
	Status string

	// Ports is the list of exposed ports for this container.
	// Optional. Example: ["80/tcp", "443/tcp", "5432/tcp"]
	Ports []string

	// Environment contains environment variables set on the container.
	// Optional. Key-value pairs of environment configuration.
	Environment map[string]string

	// Labels contains metadata labels attached to the container.
	// Optional. Key-value pairs used for organization and filtering.
	Labels map[string]string
}

// NodeType returns the canonical node type for Container nodes.
// Implements GraphNode interface.
func (c *Container) NodeType() string {
	return graphrag.NodeTypeCloudAsset // Containers are cloud/infrastructure assets
}

// IdentifyingProperties returns the properties that uniquely identify this container.
// For Container nodes, only the ID is identifying.
// Implements GraphNode interface.
func (c *Container) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": c.ID,
	}
}

// Properties returns all properties to set on the container node.
// Includes identifying properties (id) and optional descriptive properties.
// Implements GraphNode interface.
func (c *Container) Properties() map[string]any {
	props := map[string]any{
		"id":   c.ID,
		"type": "container", // Subtype of cloud_asset
	}

	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Image != "" {
		props["image"] = c.Image
	}
	if c.Status != "" {
		props["status"] = c.Status
	}
	if len(c.Ports) > 0 {
		props["ports"] = c.Ports
	}
	if len(c.Environment) > 0 {
		props["environment"] = c.Environment
	}
	if len(c.Labels) > 0 {
		props["labels"] = c.Labels
	}

	return props
}

// ParentRef returns nil because Container is a root node with no parent.
// Implements GraphNode interface.
func (c *Container) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Container is a root node.
// Implements GraphNode interface.
func (c *Container) RelationshipType() string {
	return ""
}

// ContainerImage represents a container image in the knowledge graph.
// Container images are templates used to create container instances, stored in registries.
//
// Example:
//
//	image := &ContainerImage{
//	    Repository: "nginx",
//	    Tag:        "1.24",
//	    Digest:     "sha256:abc123...",
//	    Size:       142000000,
//	    OS:         "linux",
//	    Arch:       "amd64",
//	}
//
// Identifying Properties:
//   - repository (required): The image repository name
//   - tag (required): The image tag or version
//
// Relationships:
//   - None (root node)
type ContainerImage struct {
	// Repository is the name of the image repository.
	// This is an identifying property and is required.
	// Example: "nginx", "postgres", "myapp"
	Repository string

	// Tag is the version tag for this image.
	// This is an identifying property and is required.
	// Example: "1.24", "latest", "15-alpine"
	Tag string

	// Digest is the SHA256 digest uniquely identifying this image version.
	// Optional. Example: "sha256:abc123def456..."
	Digest string

	// Size is the total size of the image in bytes.
	// Optional. Example: 142000000 (142 MB)
	Size int64

	// OS is the operating system the image is built for.
	// Optional. Example: "linux", "windows"
	OS string

	// Arch is the architecture the image is built for.
	// Optional. Example: "amd64", "arm64", "arm"
	Arch string
}

// NodeType returns the canonical node type for ContainerImage nodes.
// Implements GraphNode interface.
func (ci *ContainerImage) NodeType() string {
	return graphrag.NodeTypeCloudAsset
}

// IdentifyingProperties returns the properties that uniquely identify this container image.
// For ContainerImage nodes, both repository and tag are identifying.
// Implements GraphNode interface.
func (ci *ContainerImage) IdentifyingProperties() map[string]any {
	return map[string]any{
		"repository": ci.Repository,
		"tag":        ci.Tag,
	}
}

// Properties returns all properties to set on the container image node.
// Implements GraphNode interface.
func (ci *ContainerImage) Properties() map[string]any {
	props := map[string]any{
		"repository": ci.Repository,
		"tag":        ci.Tag,
		"type":       "container_image", // Subtype of cloud_asset
	}

	if ci.Digest != "" {
		props["digest"] = ci.Digest
	}
	if ci.Size > 0 {
		props["size"] = ci.Size
	}
	if ci.OS != "" {
		props["os"] = ci.OS
	}
	if ci.Arch != "" {
		props["arch"] = ci.Arch
	}

	return props
}

// ParentRef returns nil because ContainerImage is a root node with no parent.
// Implements GraphNode interface.
func (ci *ContainerImage) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ContainerImage is a root node.
// Implements GraphNode interface.
func (ci *ContainerImage) RelationshipType() string {
	return ""
}

// ContainerRegistry represents a container registry in the knowledge graph.
// Registries store and distribute container images.
//
// Example:
//
//	registry := &ContainerRegistry{
//	    URL:  "https://registry.example.com",
//	    Name: "Example Registry",
//	    Type: "docker",
//	    Auth: "basic",
//	}
//
// Identifying Properties:
//   - url (required): The registry URL
//
// Relationships:
//   - None (root node)
type ContainerRegistry struct {
	// URL is the registry endpoint URL.
	// This is the identifying property and is required.
	// Example: "https://registry.example.com", "https://gcr.io"
	URL string

	// Name is the human-readable name of the registry.
	// Optional. Example: "Example Registry", "Google Container Registry"
	Name string

	// Type is the registry implementation type.
	// Optional. Common values: "docker", "ecr", "gcr", "acr", "harbor"
	Type string

	// Auth is the authentication method used by the registry.
	// Optional. Common values: "basic", "token", "oauth", "none"
	Auth string
}

// NodeType returns the canonical node type for ContainerRegistry nodes.
// Implements GraphNode interface.
func (cr *ContainerRegistry) NodeType() string {
	return graphrag.NodeTypeCloudAsset
}

// IdentifyingProperties returns the properties that uniquely identify this registry.
// For ContainerRegistry nodes, only the URL is identifying.
// Implements GraphNode interface.
func (cr *ContainerRegistry) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropURL: cr.URL,
	}
}

// Properties returns all properties to set on the container registry node.
// Implements GraphNode interface.
func (cr *ContainerRegistry) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropURL: cr.URL,
		"type":           "container_registry", // Subtype of cloud_asset
	}

	if cr.Name != "" {
		props[graphrag.PropName] = cr.Name
	}
	if cr.Type != "" {
		props["registry_type"] = cr.Type
	}
	if cr.Auth != "" {
		props["auth"] = cr.Auth
	}

	return props
}

// ParentRef returns nil because ContainerRegistry is a root node with no parent.
// Implements GraphNode interface.
func (cr *ContainerRegistry) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ContainerRegistry is a root node.
// Implements GraphNode interface.
func (cr *ContainerRegistry) RelationshipType() string {
	return ""
}

// Dockerfile represents a Dockerfile in the knowledge graph.
// Dockerfiles contain instructions for building container images.
//
// Example:
//
//	dockerfile := &Dockerfile{
//	    Path:        "/app/Dockerfile",
//	    BaseImage:   "node:18-alpine",
//	    Stages:      []string{"builder", "runtime"},
//	    Instructions: []string{"FROM", "RUN", "COPY", "CMD"},
//	}
//
// Identifying Properties:
//   - path (required): The file path to the Dockerfile
//
// Relationships:
//   - None (root node)
type Dockerfile struct {
	// Path is the file system path to the Dockerfile.
	// This is the identifying property and is required.
	// Example: "/app/Dockerfile", "./docker/api.Dockerfile"
	Path string

	// BaseImage is the base image specified in the FROM instruction.
	// Optional. Example: "node:18-alpine", "ubuntu:22.04"
	BaseImage string

	// Stages lists the named build stages in multi-stage builds.
	// Optional. Example: ["builder", "runtime"]
	Stages []string

	// Instructions lists the types of instructions used in the Dockerfile.
	// Optional. Example: ["FROM", "RUN", "COPY", "EXPOSE", "CMD"]
	Instructions []string
}

// NodeType returns the canonical node type for Dockerfile nodes.
// Implements GraphNode interface.
func (d *Dockerfile) NodeType() string {
	return graphrag.NodeTypeCloudAsset
}

// IdentifyingProperties returns the properties that uniquely identify this Dockerfile.
// For Dockerfile nodes, only the path is identifying.
// Implements GraphNode interface.
func (d *Dockerfile) IdentifyingProperties() map[string]any {
	return map[string]any{
		"path": d.Path,
	}
}

// Properties returns all properties to set on the Dockerfile node.
// Implements GraphNode interface.
func (d *Dockerfile) Properties() map[string]any {
	props := map[string]any{
		"path": d.Path,
		"type": "dockerfile", // Subtype of cloud_asset
	}

	if d.BaseImage != "" {
		props["base_image"] = d.BaseImage
	}
	if len(d.Stages) > 0 {
		props["stages"] = d.Stages
	}
	if len(d.Instructions) > 0 {
		props["instructions"] = d.Instructions
	}

	return props
}

// ParentRef returns nil because Dockerfile is a root node with no parent.
// Implements GraphNode interface.
func (d *Dockerfile) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Dockerfile is a root node.
// Implements GraphNode interface.
func (d *Dockerfile) RelationshipType() string {
	return ""
}
