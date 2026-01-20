package domain

import "github.com/zero-day-ai/sdk/graphrag"

// CloudAsset represents a cloud infrastructure resource (AWS, GCP, Azure).
// CloudAssets can host services or hosts via HOSTS relationships.
//
// Hierarchy: CloudAsset is a root node (no parent)
//
// Identifying Properties: provider, resource_id
// Parent: None (root node)
//
// Example:
//
//	cloudAsset := &CloudAsset{
//	    Provider:   "aws",
//	    ResourceID: "i-0123456789abcdef0",
//	    Region:     "us-east-1",
//	    Type:       "ec2-instance",
//	    Name:       "web-server-01",
//	}
type CloudAsset struct {
	// Provider is the cloud provider (e.g., "aws", "gcp", "azure").
	// This is an identifying property.
	Provider string

	// ResourceID is the cloud-specific resource identifier (e.g., AWS instance ID).
	// This is an identifying property.
	ResourceID string

	// Region is the cloud region (e.g., "us-east-1", "europe-west1") (optional).
	Region string

	// Type is the resource type (e.g., "ec2-instance", "s3-bucket", "vm") (optional).
	Type string

	// Name is the resource name or tag (optional).
	Name string

	// AccountID is the cloud account/project ID (optional).
	AccountID string

	// VPC is the VPC/VNet ID (optional).
	VPC string

	// SubnetID is the subnet ID (optional).
	SubnetID string

	// SecurityGroups is the list of security group IDs (optional).
	SecurityGroups []string

	// Tags is the map of resource tags (optional).
	Tags map[string]string

	// State is the resource state (e.g., "running", "stopped") (optional).
	State string
}

// NodeType returns the canonical node type for cloud assets.
func (c *CloudAsset) NodeType() string {
	return graphrag.NodeTypeCloudAsset
}

// IdentifyingProperties returns the properties that uniquely identify this cloud asset.
// A cloud asset is identified by its provider and resource ID.
func (c *CloudAsset) IdentifyingProperties() map[string]any {
	return map[string]any{
		"provider":    c.Provider,
		"resource_id": c.ResourceID,
	}
}

// Properties returns all properties to set on the cloud asset node.
func (c *CloudAsset) Properties() map[string]any {
	props := map[string]any{
		"provider":    c.Provider,
		"resource_id": c.ResourceID,
	}

	// Add optional properties if present
	if c.Region != "" {
		props["region"] = c.Region
	}
	if c.Type != "" {
		props["type"] = c.Type
	}
	if c.Name != "" {
		props["name"] = c.Name
	}
	if c.AccountID != "" {
		props["account_id"] = c.AccountID
	}
	if c.VPC != "" {
		props["vpc"] = c.VPC
	}
	if c.SubnetID != "" {
		props["subnet_id"] = c.SubnetID
	}
	if c.SecurityGroups != nil && len(c.SecurityGroups) > 0 {
		props["security_groups"] = c.SecurityGroups
	}
	if c.Tags != nil && len(c.Tags) > 0 {
		props["tags"] = c.Tags
	}
	if c.State != "" {
		props["state"] = c.State
	}

	return props
}

// ParentRef returns nil because CloudAsset is a root node.
func (c *CloudAsset) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because CloudAsset has no parent.
func (c *CloudAsset) RelationshipType() string {
	return ""
}
