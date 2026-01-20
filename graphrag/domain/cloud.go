package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// CloudAccount represents a cloud provider account in the knowledge graph.
//
// Identifying Properties:
//   - provider (required): The cloud provider name
//   - account_id (required): The account identifier
//
// Relationships:
//   - None (root node)
//   - Children: CloudVPC, CloudIAMRole, CloudIAMPolicy, CloudTrail
type CloudAccount struct {
	Provider  string // Identifying: "aws", "gcp", "azure"
	AccountID string // Identifying
	Name      string
	Email     string
	Region    string // Default region
}

func (c *CloudAccount) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudAccount) IdentifyingProperties() map[string]any {
	return map[string]any{
		"provider":   c.Provider,
		"account_id": c.AccountID,
	}
}
func (c *CloudAccount) Properties() map[string]any {
	props := map[string]any{
		"provider":   c.Provider,
		"account_id": c.AccountID,
		"type":       "cloud_account",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Email != "" {
		props["email"] = c.Email
	}
	if c.Region != "" {
		props["region"] = c.Region
	}
	return props
}
func (c *CloudAccount) ParentRef() *NodeRef      { return nil }
func (c *CloudAccount) RelationshipType() string { return "" }

// CloudVPC represents a Virtual Private Cloud in the knowledge graph.
//
// Identifying Properties:
//   - account_id (required): The parent account ID
//   - vpc_id (required): The VPC identifier
//
// Relationships:
//   - Parent: CloudAccount (via PART_OF relationship)
//   - Children: CloudSubnet, CloudSecurityGroup
type CloudVPC struct {
	AccountID string // Identifying (parent reference)
	VPCID     string // Identifying
	Name      string
	CIDR      string
	Region    string
	IsDefault bool
}

func (c *CloudVPC) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudVPC) IdentifyingProperties() map[string]any {
	return map[string]any{
		"account_id": c.AccountID,
		"vpc_id":     c.VPCID,
	}
}
func (c *CloudVPC) Properties() map[string]any {
	props := map[string]any{
		"account_id": c.AccountID,
		"vpc_id":     c.VPCID,
		"type":       "cloud_vpc",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.CIDR != "" {
		props["cidr"] = c.CIDR
	}
	if c.Region != "" {
		props["region"] = c.Region
	}
	props["is_default"] = c.IsDefault
	return props
}
func (c *CloudVPC) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"account_id": c.AccountID, "type": "cloud_account"},
	}
}
func (c *CloudVPC) RelationshipType() string { return graphrag.RelTypePartOf }

// CloudSubnet represents a subnet within a VPC in the knowledge graph.
//
// Identifying Properties:
//   - vpc_id (required): The parent VPC ID
//   - subnet_id (required): The subnet identifier
//
// Relationships:
//   - Parent: CloudVPC (via PART_OF relationship)
type CloudSubnet struct {
	VPCID            string // Identifying (parent reference)
	SubnetID         string // Identifying
	Name             string
	CIDR             string
	AvailabilityZone string
	Public           bool
}

func (c *CloudSubnet) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudSubnet) IdentifyingProperties() map[string]any {
	return map[string]any{
		"vpc_id":    c.VPCID,
		"subnet_id": c.SubnetID,
	}
}
func (c *CloudSubnet) Properties() map[string]any {
	props := map[string]any{
		"vpc_id":    c.VPCID,
		"subnet_id": c.SubnetID,
		"type":      "cloud_subnet",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.CIDR != "" {
		props["cidr"] = c.CIDR
	}
	if c.AvailabilityZone != "" {
		props["availability_zone"] = c.AvailabilityZone
	}
	props["public"] = c.Public
	return props
}
func (c *CloudSubnet) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"vpc_id": c.VPCID, "type": "cloud_vpc"},
	}
}
func (c *CloudSubnet) RelationshipType() string { return graphrag.RelTypePartOf }

// CloudSecurityGroup represents a security group in the knowledge graph.
//
// Identifying Properties:
//   - vpc_id (required): The parent VPC ID
//   - group_id (required): The security group identifier
//
// Relationships:
//   - Parent: CloudVPC (via PART_OF relationship)
type CloudSecurityGroup struct {
	VPCID         string // Identifying (parent reference)
	GroupID       string // Identifying
	Name          string
	Description   string
	InboundRules  []string // Simplified rule representation
	OutboundRules []string
}

func (c *CloudSecurityGroup) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudSecurityGroup) IdentifyingProperties() map[string]any {
	return map[string]any{
		"vpc_id":   c.VPCID,
		"group_id": c.GroupID,
	}
}
func (c *CloudSecurityGroup) Properties() map[string]any {
	props := map[string]any{
		"vpc_id":   c.VPCID,
		"group_id": c.GroupID,
		"type":     "cloud_security_group",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	if len(c.InboundRules) > 0 {
		props["inbound_rules"] = c.InboundRules
	}
	if len(c.OutboundRules) > 0 {
		props["outbound_rules"] = c.OutboundRules
	}
	return props
}
func (c *CloudSecurityGroup) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"vpc_id": c.VPCID, "type": "cloud_vpc"},
	}
}
func (c *CloudSecurityGroup) RelationshipType() string { return graphrag.RelTypePartOf }

// CloudInstance represents a cloud compute instance in the knowledge graph.
//
// Identifying Properties:
//   - instance_id (required): The instance identifier
//
// Relationships:
//   - None (root node)
type CloudInstance struct {
	InstanceID     string // Identifying
	Name           string
	Type           string // Instance type/size
	State          string // "running", "stopped", "terminated"
	PublicIP       string
	PrivateIP      string
	SubnetID       string
	SecurityGroups []string
	ImageID        string
	LaunchTime     string
}

func (c *CloudInstance) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudInstance) IdentifyingProperties() map[string]any {
	return map[string]any{
		"instance_id": c.InstanceID,
	}
}
func (c *CloudInstance) Properties() map[string]any {
	props := map[string]any{
		"instance_id": c.InstanceID,
		"type":        "cloud_instance",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Type != "" {
		props["instance_type"] = c.Type
	}
	if c.State != "" {
		props[graphrag.PropState] = c.State
	}
	if c.PublicIP != "" {
		props["public_ip"] = c.PublicIP
	}
	if c.PrivateIP != "" {
		props["private_ip"] = c.PrivateIP
	}
	if c.SubnetID != "" {
		props["subnet_id"] = c.SubnetID
	}
	if len(c.SecurityGroups) > 0 {
		props["security_groups"] = c.SecurityGroups
	}
	if c.ImageID != "" {
		props["image_id"] = c.ImageID
	}
	if c.LaunchTime != "" {
		props["launch_time"] = c.LaunchTime
	}
	return props
}
func (c *CloudInstance) ParentRef() *NodeRef      { return nil }
func (c *CloudInstance) RelationshipType() string { return "" }

// CloudFunction represents a serverless function in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The function name
//   - region (required): The region
//
// Relationships:
//   - None (root node)
type CloudFunction struct {
	Name        string // Identifying
	Region      string // Identifying
	Runtime     string
	Handler     string
	Memory      int
	Timeout     int
	Environment map[string]string
	Role        string
	CodeSize    int64
}

func (c *CloudFunction) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudFunction) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
		"region":          c.Region,
	}
}
func (c *CloudFunction) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
		"region":          c.Region,
		"type":            "cloud_function",
	}
	if c.Runtime != "" {
		props["runtime"] = c.Runtime
	}
	if c.Handler != "" {
		props["handler"] = c.Handler
	}
	if c.Memory > 0 {
		props["memory"] = c.Memory
	}
	if c.Timeout > 0 {
		props["timeout"] = c.Timeout
	}
	if len(c.Environment) > 0 {
		props["environment"] = c.Environment
	}
	if c.Role != "" {
		props["role"] = c.Role
	}
	if c.CodeSize > 0 {
		props["code_size"] = c.CodeSize
	}
	return props
}
func (c *CloudFunction) ParentRef() *NodeRef      { return nil }
func (c *CloudFunction) RelationshipType() string { return "" }

// CloudStorage represents a cloud storage bucket in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The bucket/storage name
//   - provider (required): The cloud provider
//
// Relationships:
//   - None (root node)
type CloudStorage struct {
	Name       string // Identifying
	Provider   string // Identifying: "aws", "gcp", "azure"
	Region     string
	Public     bool
	Versioning bool
	Encryption string
	SizeBytes  int64
}

func (c *CloudStorage) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudStorage) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
		"provider":        c.Provider,
	}
}
func (c *CloudStorage) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
		"provider":        c.Provider,
		"type":            "cloud_storage",
	}
	if c.Region != "" {
		props["region"] = c.Region
	}
	props["public"] = c.Public
	props["versioning"] = c.Versioning
	if c.Encryption != "" {
		props["encryption"] = c.Encryption
	}
	if c.SizeBytes > 0 {
		props["size_bytes"] = c.SizeBytes
	}
	return props
}
func (c *CloudStorage) ParentRef() *NodeRef      { return nil }
func (c *CloudStorage) RelationshipType() string { return "" }

// CloudDatabase represents a cloud database instance in the knowledge graph.
//
// Identifying Properties:
//   - id (required): The database identifier
//
// Relationships:
//   - None (root node)
type CloudDatabase struct {
	ID                 string // Identifying
	Name               string
	Engine             string // "postgres", "mysql", "mongodb", etc.
	Version            string
	MultiAZ            bool
	PubliclyAccessible bool
	Endpoint           string
	Port               int
	StorageGB          int
}

func (c *CloudDatabase) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudDatabase) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": c.ID,
	}
}
func (c *CloudDatabase) Properties() map[string]any {
	props := map[string]any{
		"id":   c.ID,
		"type": "cloud_database",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Engine != "" {
		props["engine"] = c.Engine
	}
	if c.Version != "" {
		props["version"] = c.Version
	}
	props["multi_az"] = c.MultiAZ
	props["publicly_accessible"] = c.PubliclyAccessible
	if c.Endpoint != "" {
		props["endpoint"] = c.Endpoint
	}
	if c.Port > 0 {
		props[graphrag.PropPort] = c.Port
	}
	if c.StorageGB > 0 {
		props["storage_gb"] = c.StorageGB
	}
	return props
}
func (c *CloudDatabase) ParentRef() *NodeRef      { return nil }
func (c *CloudDatabase) RelationshipType() string { return "" }

// CloudQueue represents a cloud message queue in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The queue name
//   - provider (required): The cloud provider
//
// Relationships:
//   - None (root node)
type CloudQueue struct {
	Name           string // Identifying
	Provider       string // Identifying: "aws", "gcp", "azure"
	URL            string
	FIFO           bool
	DelaySeconds   int
	MaxMessageSize int
}

func (c *CloudQueue) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudQueue) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
		"provider":        c.Provider,
	}
}
func (c *CloudQueue) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
		"provider":        c.Provider,
		"type":            "cloud_queue",
	}
	if c.URL != "" {
		props[graphrag.PropURL] = c.URL
	}
	props["fifo"] = c.FIFO
	if c.DelaySeconds > 0 {
		props["delay_seconds"] = c.DelaySeconds
	}
	if c.MaxMessageSize > 0 {
		props["max_message_size"] = c.MaxMessageSize
	}
	return props
}
func (c *CloudQueue) ParentRef() *NodeRef      { return nil }
func (c *CloudQueue) RelationshipType() string { return "" }

// CloudAPIGateway represents a cloud API gateway in the knowledge graph.
//
// Identifying Properties:
//   - id (required): The gateway identifier
//
// Relationships:
//   - None (root node)
type CloudAPIGateway struct {
	ID       string // Identifying
	Name     string
	Type     string // "REST", "HTTP", "WebSocket"
	Endpoint string
	Stage    string
	AuthType string
	Routes   []string
}

func (c *CloudAPIGateway) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudAPIGateway) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": c.ID,
	}
}
func (c *CloudAPIGateway) Properties() map[string]any {
	props := map[string]any{
		"id":   c.ID,
		"type": "cloud_api_gateway",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Type != "" {
		props["gateway_type"] = c.Type
	}
	if c.Endpoint != "" {
		props["endpoint"] = c.Endpoint
	}
	if c.Stage != "" {
		props["stage"] = c.Stage
	}
	if c.AuthType != "" {
		props["auth_type"] = c.AuthType
	}
	if len(c.Routes) > 0 {
		props["routes"] = c.Routes
	}
	return props
}
func (c *CloudAPIGateway) ParentRef() *NodeRef      { return nil }
func (c *CloudAPIGateway) RelationshipType() string { return "" }

// CloudCDN represents a cloud CDN distribution in the knowledge graph.
//
// Identifying Properties:
//   - id (required): The CDN distribution identifier
//
// Relationships:
//   - None (root node)
type CloudCDN struct {
	ID         string // Identifying
	DomainName string
	Status     string
	Origins    []string
	Enabled    bool
	PriceClass string
}

func (c *CloudCDN) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudCDN) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": c.ID,
	}
}
func (c *CloudCDN) Properties() map[string]any {
	props := map[string]any{
		"id":   c.ID,
		"type": "cloud_cdn",
	}
	if c.DomainName != "" {
		props["domain_name"] = c.DomainName
	}
	if c.Status != "" {
		props["status"] = c.Status
	}
	if len(c.Origins) > 0 {
		props["origins"] = c.Origins
	}
	props["enabled"] = c.Enabled
	if c.PriceClass != "" {
		props["price_class"] = c.PriceClass
	}
	return props
}
func (c *CloudCDN) ParentRef() *NodeRef      { return nil }
func (c *CloudCDN) RelationshipType() string { return "" }

// CloudDNSZone represents a cloud DNS zone in the knowledge graph.
//
// Identifying Properties:
//   - zone_id (required): The DNS zone identifier
//
// Relationships:
//   - None (root node)
type CloudDNSZone struct {
	ZoneID      string // Identifying
	Name        string
	Type        string // "public", "private"
	RecordCount int
	NameServers []string
}

func (c *CloudDNSZone) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudDNSZone) IdentifyingProperties() map[string]any {
	return map[string]any{
		"zone_id": c.ZoneID,
	}
}
func (c *CloudDNSZone) Properties() map[string]any {
	props := map[string]any{
		"zone_id": c.ZoneID,
		"type":    "cloud_dns_zone",
	}
	if c.Name != "" {
		props[graphrag.PropName] = c.Name
	}
	if c.Type != "" {
		props["zone_type"] = c.Type
	}
	if c.RecordCount > 0 {
		props["record_count"] = c.RecordCount
	}
	if len(c.NameServers) > 0 {
		props["name_servers"] = c.NameServers
	}
	return props
}
func (c *CloudDNSZone) ParentRef() *NodeRef      { return nil }
func (c *CloudDNSZone) RelationshipType() string { return "" }

// CloudCertificate represents a cloud TLS/SSL certificate in the knowledge graph.
//
// Identifying Properties:
//   - arn (required): The certificate ARN/identifier
//
// Relationships:
//   - None (root node)
type CloudCertificate struct {
	Arn        string // Identifying
	DomainName string
	SANs       []string // Subject Alternative Names
	Status     string
	Issuer     string
	NotBefore  string
	NotAfter   string
}

func (c *CloudCertificate) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudCertificate) IdentifyingProperties() map[string]any {
	return map[string]any{
		"arn": c.Arn,
	}
}
func (c *CloudCertificate) Properties() map[string]any {
	props := map[string]any{
		"arn":  c.Arn,
		"type": "cloud_certificate",
	}
	if c.DomainName != "" {
		props["domain_name"] = c.DomainName
	}
	if len(c.SANs) > 0 {
		props["sans"] = c.SANs
	}
	if c.Status != "" {
		props["status"] = c.Status
	}
	if c.Issuer != "" {
		props["issuer"] = c.Issuer
	}
	if c.NotBefore != "" {
		props["not_before"] = c.NotBefore
	}
	if c.NotAfter != "" {
		props["not_after"] = c.NotAfter
	}
	return props
}
func (c *CloudCertificate) ParentRef() *NodeRef      { return nil }
func (c *CloudCertificate) RelationshipType() string { return "" }

// CloudKMSKey represents a cloud KMS encryption key in the knowledge graph.
//
// Identifying Properties:
//   - key_id (required): The KMS key identifier
//
// Relationships:
//   - None (root node)
type CloudKMSKey struct {
	KeyID       string // Identifying
	Arn         string
	Description string
	Enabled     bool
	KeyUsage    string // "ENCRYPT_DECRYPT", "SIGN_VERIFY"
	KeyState    string
}

func (c *CloudKMSKey) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudKMSKey) IdentifyingProperties() map[string]any {
	return map[string]any{
		"key_id": c.KeyID,
	}
}
func (c *CloudKMSKey) Properties() map[string]any {
	props := map[string]any{
		"key_id": c.KeyID,
		"type":   "cloud_kms_key",
	}
	if c.Arn != "" {
		props["arn"] = c.Arn
	}
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	props["enabled"] = c.Enabled
	if c.KeyUsage != "" {
		props["key_usage"] = c.KeyUsage
	}
	if c.KeyState != "" {
		props["key_state"] = c.KeyState
	}
	return props
}
func (c *CloudKMSKey) ParentRef() *NodeRef      { return nil }
func (c *CloudKMSKey) RelationshipType() string { return "" }

// CloudIAMRole represents a cloud IAM role in the knowledge graph.
//
// Identifying Properties:
//   - role_name (required): The role name
//   - account_id (required): The parent account ID
//
// Relationships:
//   - Parent: CloudAccount (via PART_OF relationship)
type CloudIAMRole struct {
	RoleName         string // Identifying
	AccountID        string // Identifying (parent reference)
	Arn              string
	AssumeRolePolicy string
	Policies         []string
}

func (c *CloudIAMRole) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudIAMRole) IdentifyingProperties() map[string]any {
	return map[string]any{
		"role_name":  c.RoleName,
		"account_id": c.AccountID,
	}
}
func (c *CloudIAMRole) Properties() map[string]any {
	props := map[string]any{
		"role_name":  c.RoleName,
		"account_id": c.AccountID,
		"type":       "cloud_iam_role",
	}
	if c.Arn != "" {
		props["arn"] = c.Arn
	}
	if c.AssumeRolePolicy != "" {
		props["assume_role_policy"] = c.AssumeRolePolicy
	}
	if len(c.Policies) > 0 {
		props["policies"] = c.Policies
	}
	return props
}
func (c *CloudIAMRole) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"account_id": c.AccountID, "type": "cloud_account"},
	}
}
func (c *CloudIAMRole) RelationshipType() string { return graphrag.RelTypePartOf }

// CloudIAMPolicy represents a cloud IAM policy in the knowledge graph.
//
// Identifying Properties:
//   - policy_name (required): The policy name
//   - account_id (required): The parent account ID
//
// Relationships:
//   - Parent: CloudAccount (via PART_OF relationship)
type CloudIAMPolicy struct {
	PolicyName string // Identifying
	AccountID  string // Identifying (parent reference)
	Arn        string
	Document   string
	Type       string // "managed", "inline"
}

func (c *CloudIAMPolicy) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudIAMPolicy) IdentifyingProperties() map[string]any {
	return map[string]any{
		"policy_name": c.PolicyName,
		"account_id":  c.AccountID,
	}
}
func (c *CloudIAMPolicy) Properties() map[string]any {
	props := map[string]any{
		"policy_name": c.PolicyName,
		"account_id":  c.AccountID,
		"type":        "cloud_iam_policy",
	}
	if c.Arn != "" {
		props["arn"] = c.Arn
	}
	if c.Document != "" {
		props["document"] = c.Document
	}
	if c.Type != "" {
		props["policy_type"] = c.Type
	}
	return props
}
func (c *CloudIAMPolicy) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"account_id": c.AccountID, "type": "cloud_account"},
	}
}
func (c *CloudIAMPolicy) RelationshipType() string { return graphrag.RelTypePartOf }

// CloudTrail represents a cloud audit trail in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The trail name
//
// Relationships:
//   - None (root node)
type CloudTrail struct {
	Name              string // Identifying
	S3BucketName      string
	IsMultiRegion     bool
	LogFileValidation bool
	Status            string
}

func (c *CloudTrail) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudTrail) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}
func (c *CloudTrail) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
		"type":            "cloud_trail",
	}
	if c.S3BucketName != "" {
		props["s3_bucket_name"] = c.S3BucketName
	}
	props["is_multi_region"] = c.IsMultiRegion
	props["log_file_validation"] = c.LogFileValidation
	if c.Status != "" {
		props["status"] = c.Status
	}
	return props
}
func (c *CloudTrail) ParentRef() *NodeRef      { return nil }
func (c *CloudTrail) RelationshipType() string { return "" }

// CloudMetric represents a cloud monitoring metric in the knowledge graph.
//
// Identifying Properties:
//   - namespace (required): The metric namespace
//   - name (required): The metric name
//
// Relationships:
//   - None (root node)
type CloudMetric struct {
	Namespace  string // Identifying
	Name       string // Identifying
	Dimensions map[string]string
	Statistics []string // "Average", "Sum", "Maximum", etc.
	Unit       string
}

func (c *CloudMetric) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudMetric) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace":       c.Namespace,
		graphrag.PropName: c.Name,
	}
}
func (c *CloudMetric) Properties() map[string]any {
	props := map[string]any{
		"namespace":       c.Namespace,
		graphrag.PropName: c.Name,
		"type":            "cloud_metric",
	}
	if len(c.Dimensions) > 0 {
		props["dimensions"] = c.Dimensions
	}
	if len(c.Statistics) > 0 {
		props["statistics"] = c.Statistics
	}
	if c.Unit != "" {
		props["unit"] = c.Unit
	}
	return props
}
func (c *CloudMetric) ParentRef() *NodeRef      { return nil }
func (c *CloudMetric) RelationshipType() string { return "" }

// CloudAlarm represents a cloud monitoring alarm in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The alarm name
//
// Relationships:
//   - None (root node)
type CloudAlarm struct {
	Name               string // Identifying
	MetricName         string
	Namespace          string
	ComparisonOperator string
	Threshold          float64
	State              string // "OK", "ALARM", "INSUFFICIENT_DATA"
	Actions            []string
}

func (c *CloudAlarm) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudAlarm) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}
func (c *CloudAlarm) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
		"type":            "cloud_alarm",
	}
	if c.MetricName != "" {
		props["metric_name"] = c.MetricName
	}
	if c.Namespace != "" {
		props["namespace"] = c.Namespace
	}
	if c.ComparisonOperator != "" {
		props["comparison_operator"] = c.ComparisonOperator
	}
	if c.Threshold > 0 {
		props["threshold"] = c.Threshold
	}
	if c.State != "" {
		props[graphrag.PropState] = c.State
	}
	if len(c.Actions) > 0 {
		props["actions"] = c.Actions
	}
	return props
}
func (c *CloudAlarm) ParentRef() *NodeRef      { return nil }
func (c *CloudAlarm) RelationshipType() string { return "" }

// CloudRegion represents a cloud provider region in the knowledge graph.
//
// Identifying Properties:
//   - provider (required): The cloud provider
//   - name (required): The region name
//
// Relationships:
//   - None (root node)
type CloudRegion struct {
	Provider string // Identifying: "aws", "gcp", "azure"
	Name     string // Identifying: "us-east-1", "europe-west1", etc.
	Location string // Human-readable location
	Zones    []string
}

func (c *CloudRegion) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (c *CloudRegion) IdentifyingProperties() map[string]any {
	return map[string]any{
		"provider":        c.Provider,
		graphrag.PropName: c.Name,
	}
}
func (c *CloudRegion) Properties() map[string]any {
	props := map[string]any{
		"provider":        c.Provider,
		graphrag.PropName: c.Name,
		"type":            "cloud_region",
	}
	if c.Location != "" {
		props["location"] = c.Location
	}
	if len(c.Zones) > 0 {
		props["zones"] = c.Zones
	}
	return props
}
func (c *CloudRegion) ParentRef() *NodeRef      { return nil }
func (c *CloudRegion) RelationshipType() string { return "" }
