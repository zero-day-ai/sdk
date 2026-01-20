package domain

// DiscoveryResult is a standardized container for tool and agent output.
// It provides typed slices for each canonical domain type, plus a flexible
// Custom slice for agent-defined types.
//
// Tools should populate the appropriate slices with discovered assets,
// and the GraphRAG loader will automatically create nodes and relationships
// in the knowledge graph.
//
// Example (Network Reconnaissance Tool):
//
//	result := &DiscoveryResult{
//	    Hosts: []*Host{
//	        {IP: "192.168.1.1", Hostname: "gateway", State: "up"},
//	        {IP: "192.168.1.10", Hostname: "web-server", State: "up"},
//	    },
//	    Ports: []*Port{
//	        {HostID: "192.168.1.10", Number: 80, Protocol: "tcp", State: "open"},
//	        {HostID: "192.168.1.10", Number: 443, Protocol: "tcp", State: "open"},
//	    },
//	    Services: []*Service{
//	        {PortID: "192.168.1.10:80:tcp", Name: "http", Version: "nginx/1.18.0"},
//	        {PortID: "192.168.1.10:443:tcp", Name: "https", Version: "nginx/1.18.0"},
//	    },
//	}
//
// Example (Custom Kubernetes Agent):
//
//	result := &DiscoveryResult{
//	    Custom: []GraphNode{
//	        NewCustomEntity("k8s", "pod").
//	            WithIDProps(map[string]any{"namespace": "default", "name": "web-01"}),
//	        NewCustomEntity("k8s", "service").
//	            WithIDProps(map[string]any{"namespace": "default", "name": "web-svc"}),
//	    },
//	}
type DiscoveryResult struct {
	// Hosts contains discovered network hosts (IP addresses).
	Hosts []*Host `json:"hosts,omitempty"`

	// Ports contains discovered network ports on hosts.
	Ports []*Port `json:"ports,omitempty"`

	// Services contains discovered services running on ports.
	Services []*Service `json:"services,omitempty"`

	// Endpoints contains discovered web endpoints or URLs.
	Endpoints []*Endpoint `json:"endpoints,omitempty"`

	// Domains contains discovered root domain names.
	Domains []*Domain `json:"domains,omitempty"`

	// Subdomains contains discovered subdomains under root domains.
	Subdomains []*Subdomain `json:"subdomains,omitempty"`

	// Technologies contains discovered technologies, frameworks, or software.
	Technologies []*Technology `json:"technologies,omitempty"`

	// Certificates contains discovered TLS/SSL certificates.
	Certificates []*Certificate `json:"certificates,omitempty"`

	// CloudAssets contains discovered cloud infrastructure resources.
	CloudAssets []*CloudAsset `json:"cloud_assets,omitempty"`

	// APIs contains discovered web API services.
	APIs []*API `json:"apis,omitempty"`

	// Network contains discovered network infrastructure (15 types).
	DNSRecords        []*DNSRecord        `json:"dns_records,omitempty"`
	Firewalls         []*Firewall         `json:"firewalls,omitempty"`
	FirewallRules     []*FirewallRule     `json:"firewall_rules,omitempty"`
	Routers           []*Router           `json:"routers,omitempty"`
	Routes            []*Route            `json:"routes,omitempty"`
	LoadBalancers     []*LoadBalancer     `json:"load_balancers,omitempty"`
	Proxies           []*Proxy            `json:"proxies,omitempty"`
	VPNs              []*VPN              `json:"vpns,omitempty"`
	Networks          []*Network          `json:"networks,omitempty"`
	VLANs             []*VLAN             `json:"vlans,omitempty"`
	NetworkInterfaces []*NetworkInterface `json:"network_interfaces,omitempty"`
	NetworkZones      []*NetworkZone      `json:"network_zones,omitempty"`
	NetworkACLs       []*NetworkACL       `json:"network_acls,omitempty"`
	NATGateways       []*NATGateway       `json:"nat_gateways,omitempty"`
	BGPPeers          []*BGPPeer          `json:"bgp_peers,omitempty"`

	// Web/API contains discovered web and API components (16 types).
	APIEndpoints     []*APIEndpoint     `json:"api_endpoints,omitempty"`
	Parameters       []*Parameter       `json:"parameters,omitempty"`
	Headers          []*Header          `json:"headers,omitempty"`
	Cookies          []*Cookie          `json:"cookies,omitempty"`
	Forms            []*Form            `json:"forms,omitempty"`
	FormFields       []*FormField       `json:"form_fields,omitempty"`
	WebSockets       []*WebSocket       `json:"web_sockets,omitempty"`
	GraphQLSchemas   []*GraphQLSchema   `json:"graphql_schemas,omitempty"`
	GraphQLQueries   []*GraphQLQuery    `json:"graphql_queries,omitempty"`
	GraphQLMutations []*GraphQLMutation `json:"graphql_mutations,omitempty"`
	RESTResources    []*RESTResource    `json:"rest_resources,omitempty"`
	CORSPolicies     []*CORSPolicy      `json:"cors_policies,omitempty"`
	RateLimits       []*RateLimit       `json:"rate_limits,omitempty"`
	RequestBodies    []*RequestBody     `json:"request_bodies,omitempty"`
	Responses        []*Response        `json:"responses,omitempty"`
	StatusCodes      []*StatusCode      `json:"status_codes,omitempty"`

	// Identity contains discovered identity and access resources (16 types).
	Users             []*User             `json:"users,omitempty"`
	Groups            []*Group            `json:"groups,omitempty"`
	Roles             []*Role             `json:"roles,omitempty"`
	Permissions       []*Permission       `json:"permissions,omitempty"`
	Policies          []*Policy           `json:"policies,omitempty"`
	Credentials       []*Credential       `json:"credentials,omitempty"`
	APIKeys           []*APIKey           `json:"api_keys,omitempty"`
	Tokens            []*Token            `json:"tokens,omitempty"`
	OAuthClients      []*OAuthClient      `json:"oauth_clients,omitempty"`
	OAuthScopes       []*OAuthScope       `json:"oauth_scopes,omitempty"`
	SAMLProviders     []*SAMLProvider     `json:"saml_providers,omitempty"`
	IdentityProviders []*IdentityProvider `json:"identity_providers,omitempty"`
	ServiceAccounts   []*ServiceAccount   `json:"service_accounts,omitempty"`
	Sessions          []*Session          `json:"sessions,omitempty"`
	AccessKeys        []*AccessKey        `json:"access_keys,omitempty"`
	MFADevices        []*MFADevice        `json:"mfa_devices,omitempty"`

	// AI/LLM contains discovered AI and LLM resources (16 types).
	LLMs               []*LLM               `json:"llms,omitempty"`
	LLMDeployments     []*LLMDeployment     `json:"llm_deployments,omitempty"`
	Prompts            []*Prompt            `json:"prompts,omitempty"`
	SystemPrompts      []*SystemPrompt      `json:"system_prompts,omitempty"`
	Guardrails         []*Guardrail         `json:"guardrails,omitempty"`
	ContentFilters     []*ContentFilter     `json:"content_filters,omitempty"`
	LLMResponses       []*LLMResponse       `json:"llm_responses,omitempty"`
	TokenUsages        []*TokenUsage        `json:"token_usages,omitempty"`
	EmbeddingModels    []*EmbeddingModel    `json:"embedding_models,omitempty"`
	FineTunes          []*FineTune          `json:"fine_tunes,omitempty"`
	ModelRegistries    []*ModelRegistry     `json:"model_registries,omitempty"`
	ModelVersions      []*ModelVersion      `json:"model_versions,omitempty"`
	InferenceEndpoints []*InferenceEndpoint `json:"inference_endpoints,omitempty"`
	BatchJobs          []*BatchJob          `json:"batch_jobs,omitempty"`
	TrainingRuns       []*TrainingRun       `json:"training_runs,omitempty"`
	Datasets           []*Dataset           `json:"datasets,omitempty"`

	// AI Agent contains discovered AI agent resources (15 types).
	AIAgents       []*AIAgent       `json:"ai_agents,omitempty"`
	AgentConfigs   []*AgentConfig   `json:"agent_configs,omitempty"`
	AgentMemories  []*AgentMemory   `json:"agent_memories,omitempty"`
	AgentTools     []*AgentTool     `json:"agent_tools,omitempty"`
	Chains         []*Chain         `json:"chains,omitempty"`
	Workflows      []*Workflow      `json:"workflows,omitempty"`
	Crews          []*Crew          `json:"crews,omitempty"`
	AgentTasks     []*AgentTask     `json:"agent_tasks,omitempty"`
	AgentRoles     []*AgentRole     `json:"agent_roles,omitempty"`
	ToolCalls      []*ToolCall      `json:"tool_calls,omitempty"`
	ReasoningSteps []*ReasoningStep `json:"reasoning_steps,omitempty"`
	MemoryEntries  []*MemoryEntry   `json:"memory_entries,omitempty"`
	AgentLoops     []*AgentLoop     `json:"agent_loops,omitempty"`
	PlanningSteps  []*PlanningStep  `json:"planning_steps,omitempty"`
	AgentArtifacts []*AgentArtifact `json:"agent_artifacts,omitempty"`

	// MCP contains discovered Model Context Protocol resources (9 types).
	MCPServers      []*MCPServer     `json:"mcp_servers,omitempty"`
	MCPTools        []*MCPTool       `json:"mcp_tools,omitempty"`
	MCPResources    []*MCPResource   `json:"mcp_resources,omitempty"`
	MCPPrompts      []*MCPPrompt     `json:"mcp_prompts,omitempty"`
	MCPClients      []*MCPClient     `json:"mcp_clients,omitempty"`
	MCPTransports   []*MCPTransport  `json:"mcp_transports,omitempty"`
	MCPCapabilities []*MCPCapability `json:"mcp_capabilities,omitempty"`
	MCPSamplings    []*MCPSampling   `json:"mcp_samplings,omitempty"`
	MCPRoots        []*MCPRoots      `json:"mcp_roots,omitempty"`

	// RAG contains discovered RAG and knowledge base resources (11 types).
	VectorStores       []*VectorStore      `json:"vector_stores,omitempty"`
	VectorIndexes      []*VectorIndex      `json:"vector_indexes,omitempty"`
	Documents          []*Document         `json:"documents,omitempty"`
	DocumentChunks     []*DocumentChunk    `json:"document_chunks,omitempty"`
	KnowledgeBases     []*KnowledgeBase    `json:"knowledge_bases,omitempty"`
	Retrievers         []*Retriever        `json:"retrievers,omitempty"`
	RAGPipelines       []*RAGPipeline      `json:"rag_pipelines,omitempty"`
	Embeddings         []*Embedding        `json:"embeddings,omitempty"`
	Rerankers          []*Reranker         `json:"rerankers,omitempty"`
	ChunkingStrategies []*ChunkingStrategy `json:"chunking_strategies,omitempty"`
	RetrievalResults   []*RetrievalResult  `json:"retrieval_results,omitempty"`

	// Data contains discovered data storage and processing resources (16 types).
	Databases        []*Database        `json:"databases,omitempty"`
	Tables           []*Table           `json:"tables,omitempty"`
	Columns          []*Column          `json:"columns,omitempty"`
	Indexes          []*Index           `json:"indexes,omitempty"`
	Views            []*View            `json:"views,omitempty"`
	StoredProcedures []*StoredProcedure `json:"stored_procedures,omitempty"`
	Triggers         []*Trigger         `json:"triggers,omitempty"`
	Files            []*File            `json:"files,omitempty"`
	StorageBuckets   []*StorageBucket   `json:"storage_buckets,omitempty"`
	Objects          []*Object          `json:"objects,omitempty"`
	Queues           []*Queue           `json:"queues,omitempty"`
	Topics           []*Topic           `json:"topics,omitempty"`
	Streams          []*Stream          `json:"streams,omitempty"`
	Caches           []*Cache           `json:"caches,omitempty"`
	Schemas          []*Schema          `json:"schemas,omitempty"`
	DataPipelines    []*DataPipeline    `json:"data_pipelines,omitempty"`

	// Container contains discovered container resources (4 types).
	Containers          []*Container         `json:"containers,omitempty"`
	ContainerImages     []*ContainerImage    `json:"container_images,omitempty"`
	ContainerRegistries []*ContainerRegistry `json:"container_registries,omitempty"`
	Dockerfiles         []*Dockerfile        `json:"dockerfiles,omitempty"`

	// Kubernetes contains discovered Kubernetes resources (22 types).
	K8sClusters            []*K8sCluster            `json:"k8s_clusters,omitempty"`
	K8sNamespaces          []*K8sNamespace          `json:"k8s_namespaces,omitempty"`
	K8sPods                []*K8sPod                `json:"k8s_pods,omitempty"`
	K8sDeployments         []*K8sDeployment         `json:"k8s_deployments,omitempty"`
	K8sServices            []*K8sService            `json:"k8s_services,omitempty"`
	K8sIngresses           []*K8sIngress            `json:"k8s_ingresses,omitempty"`
	K8sConfigMaps          []*K8sConfigMap          `json:"k8s_config_maps,omitempty"`
	K8sSecrets             []*K8sSecret             `json:"k8s_secrets,omitempty"`
	K8sPVCs                []*K8sPVC                `json:"k8s_pvcs,omitempty"`
	K8sPVs                 []*K8sPV                 `json:"k8s_pvs,omitempty"`
	K8sStatefulSets        []*K8sStatefulSet        `json:"k8s_stateful_sets,omitempty"`
	K8sDaemonSets          []*K8sDaemonSet          `json:"k8s_daemon_sets,omitempty"`
	K8sJobs                []*K8sJob                `json:"k8s_jobs,omitempty"`
	K8sCronJobs            []*K8sCronJob            `json:"k8s_cron_jobs,omitempty"`
	K8sServiceAccounts     []*K8sServiceAccount     `json:"k8s_service_accounts,omitempty"`
	K8sRoles               []*K8sRole               `json:"k8s_roles,omitempty"`
	K8sRoleBindings        []*K8sRoleBinding        `json:"k8s_role_bindings,omitempty"`
	K8sClusterRoles        []*K8sClusterRole        `json:"k8s_cluster_roles,omitempty"`
	K8sClusterRoleBindings []*K8sClusterRoleBinding `json:"k8s_cluster_role_bindings,omitempty"`
	K8sNetworkPolicies     []*K8sNetworkPolicy      `json:"k8s_network_policies,omitempty"`
	K8sLimitRanges         []*K8sLimitRange         `json:"k8s_limit_ranges,omitempty"`
	K8sResourceQuotas      []*K8sResourceQuota      `json:"k8s_resource_quotas,omitempty"`

	// Cloud contains discovered cloud resources (20 types).
	CloudAccounts       []*CloudAccount       `json:"cloud_accounts,omitempty"`
	CloudVPCs           []*CloudVPC           `json:"cloud_vpcs,omitempty"`
	CloudSubnets        []*CloudSubnet        `json:"cloud_subnets,omitempty"`
	CloudSecurityGroups []*CloudSecurityGroup `json:"cloud_security_groups,omitempty"`
	CloudInstances      []*CloudInstance      `json:"cloud_instances,omitempty"`
	CloudFunctions      []*CloudFunction      `json:"cloud_functions,omitempty"`
	CloudStorages       []*CloudStorage       `json:"cloud_storages,omitempty"`
	CloudDatabases      []*CloudDatabase      `json:"cloud_databases,omitempty"`
	CloudQueues         []*CloudQueue         `json:"cloud_queues,omitempty"`
	CloudAPIGateways    []*CloudAPIGateway    `json:"cloud_api_gateways,omitempty"`
	CloudCDNs           []*CloudCDN           `json:"cloud_cdns,omitempty"`
	CloudDNSZones       []*CloudDNSZone       `json:"cloud_dns_zones,omitempty"`
	CloudCertificates   []*CloudCertificate   `json:"cloud_certificates,omitempty"`
	CloudKMSKeys        []*CloudKMSKey        `json:"cloud_kms_keys,omitempty"`
	CloudIAMRoles       []*CloudIAMRole       `json:"cloud_iam_roles,omitempty"`
	CloudIAMPolicies    []*CloudIAMPolicy     `json:"cloud_iam_policies,omitempty"`
	CloudTrails         []*CloudTrail         `json:"cloud_trails,omitempty"`
	CloudMetrics        []*CloudMetric        `json:"cloud_metrics,omitempty"`
	CloudAlarms         []*CloudAlarm         `json:"cloud_alarms,omitempty"`
	CloudRegions        []*CloudRegion        `json:"cloud_regions,omitempty"`

	// Attack contains discovered MITRE ATT&CK techniques (2 types).
	Tactics    []*Tactic    `json:"tactics,omitempty"`
	Techniques []*Technique `json:"techniques,omitempty"`

	// Findings contains discovered security findings.
	Findings []*Finding `json:"findings,omitempty"`

	// Custom contains custom agent-defined graph nodes.
	// Use this for domain-specific types like "k8s:pod" or "aws:security_group".
	Custom []GraphNode `json:"custom,omitempty"`
}

// AllNodes returns all discovered nodes as a flattened slice of GraphNode interfaces.
// This is the primary method used by the GraphRAG loader to process all discoveries.
//
// Nodes are returned in dependency order:
//  1. Cloud foundation (accounts, regions, VPCs)
//  2. Network infrastructure (networks, firewalls, routers)
//  3. K8s foundation (clusters, namespaces)
//  4. Compute resources (hosts, cloud instances, containers)
//  5. K8s workloads (pods, deployments, services)
//  6. Network details (ports, DNS, load balancers)
//  7. Services and endpoints
//  8. Web/API resources (endpoints, parameters, forms)
//  9. AI/LLM infrastructure (models, deployments, embeddings)
//
// 10. AI agents and workflows
// 11. MCP servers and resources
// 12. RAG systems (vector stores, documents, pipelines)
// 13. Data resources (databases, tables, queues)
// 14. Identity and access (users, groups, roles)
// 15. Security findings and attack techniques
//
// This ordering ensures parent nodes are created before child nodes reference them.
func (d *DiscoveryResult) AllNodes() []GraphNode {
	var nodes []GraphNode

	// 1. Cloud foundation - accounts, regions, VPCs first
	for _, account := range d.CloudAccounts {
		nodes = append(nodes, account)
	}
	for _, region := range d.CloudRegions {
		nodes = append(nodes, region)
	}
	for _, vpc := range d.CloudVPCs {
		nodes = append(nodes, vpc)
	}
	for _, subnet := range d.CloudSubnets {
		nodes = append(nodes, subnet)
	}
	for _, sg := range d.CloudSecurityGroups {
		nodes = append(nodes, sg)
	}

	// 2. Network infrastructure - networks, zones, firewalls, routers
	for _, network := range d.Networks {
		nodes = append(nodes, network)
	}
	for _, zone := range d.NetworkZones {
		nodes = append(nodes, zone)
	}
	for _, vlan := range d.VLANs {
		nodes = append(nodes, vlan)
	}
	for _, firewall := range d.Firewalls {
		nodes = append(nodes, firewall)
	}
	for _, rule := range d.FirewallRules {
		nodes = append(nodes, rule)
	}
	for _, acl := range d.NetworkACLs {
		nodes = append(nodes, acl)
	}
	for _, router := range d.Routers {
		nodes = append(nodes, router)
	}
	for _, route := range d.Routes {
		nodes = append(nodes, route)
	}
	for _, nat := range d.NATGateways {
		nodes = append(nodes, nat)
	}
	for _, bgp := range d.BGPPeers {
		nodes = append(nodes, bgp)
	}

	// 3. K8s foundation - clusters, namespaces
	for _, cluster := range d.K8sClusters {
		nodes = append(nodes, cluster)
	}
	for _, ns := range d.K8sNamespaces {
		nodes = append(nodes, ns)
	}

	// 4. Compute resources - hosts, cloud instances, containers
	for _, host := range d.Hosts {
		nodes = append(nodes, host)
	}
	for _, instance := range d.CloudInstances {
		nodes = append(nodes, instance)
	}
	for _, registry := range d.ContainerRegistries {
		nodes = append(nodes, registry)
	}
	for _, image := range d.ContainerImages {
		nodes = append(nodes, image)
	}
	for _, dockerfile := range d.Dockerfiles {
		nodes = append(nodes, dockerfile)
	}
	for _, container := range d.Containers {
		nodes = append(nodes, container)
	}

	// 5. K8s workloads - pods, deployments, statefulsets, etc.
	for _, pod := range d.K8sPods {
		nodes = append(nodes, pod)
	}
	for _, deployment := range d.K8sDeployments {
		nodes = append(nodes, deployment)
	}
	for _, ss := range d.K8sStatefulSets {
		nodes = append(nodes, ss)
	}
	for _, ds := range d.K8sDaemonSets {
		nodes = append(nodes, ds)
	}
	for _, job := range d.K8sJobs {
		nodes = append(nodes, job)
	}
	for _, cron := range d.K8sCronJobs {
		nodes = append(nodes, cron)
	}
	for _, svc := range d.K8sServices {
		nodes = append(nodes, svc)
	}
	for _, ingress := range d.K8sIngresses {
		nodes = append(nodes, ingress)
	}

	// 6. Network details - interfaces, DNS, load balancers, proxies
	for _, iface := range d.NetworkInterfaces {
		nodes = append(nodes, iface)
	}
	for _, dns := range d.DNSRecords {
		nodes = append(nodes, dns)
	}
	for _, lb := range d.LoadBalancers {
		nodes = append(nodes, lb)
	}
	for _, proxy := range d.Proxies {
		nodes = append(nodes, proxy)
	}
	for _, vpn := range d.VPNs {
		nodes = append(nodes, vpn)
	}
	for _, port := range d.Ports {
		nodes = append(nodes, port)
	}

	// 7. Services and domains
	for _, service := range d.Services {
		nodes = append(nodes, service)
	}
	for _, domain := range d.Domains {
		nodes = append(nodes, domain)
	}
	for _, subdomain := range d.Subdomains {
		nodes = append(nodes, subdomain)
	}
	for _, tech := range d.Technologies {
		nodes = append(nodes, tech)
	}
	for _, cert := range d.Certificates {
		nodes = append(nodes, cert)
	}
	for _, cloud := range d.CloudAssets {
		nodes = append(nodes, cloud)
	}

	// 8. Web/API resources - APIs, endpoints, parameters, forms
	for _, api := range d.APIs {
		nodes = append(nodes, api)
	}
	for _, endpoint := range d.Endpoints {
		nodes = append(nodes, endpoint)
	}
	for _, apiEndpoint := range d.APIEndpoints {
		nodes = append(nodes, apiEndpoint)
	}
	for _, param := range d.Parameters {
		nodes = append(nodes, param)
	}
	for _, header := range d.Headers {
		nodes = append(nodes, header)
	}
	for _, cookie := range d.Cookies {
		nodes = append(nodes, cookie)
	}
	for _, form := range d.Forms {
		nodes = append(nodes, form)
	}
	for _, field := range d.FormFields {
		nodes = append(nodes, field)
	}
	for _, ws := range d.WebSockets {
		nodes = append(nodes, ws)
	}
	for _, schema := range d.GraphQLSchemas {
		nodes = append(nodes, schema)
	}
	for _, query := range d.GraphQLQueries {
		nodes = append(nodes, query)
	}
	for _, mutation := range d.GraphQLMutations {
		nodes = append(nodes, mutation)
	}
	for _, resource := range d.RESTResources {
		nodes = append(nodes, resource)
	}
	for _, cors := range d.CORSPolicies {
		nodes = append(nodes, cors)
	}
	for _, limit := range d.RateLimits {
		nodes = append(nodes, limit)
	}
	for _, body := range d.RequestBodies {
		nodes = append(nodes, body)
	}
	for _, resp := range d.Responses {
		nodes = append(nodes, resp)
	}
	for _, status := range d.StatusCodes {
		nodes = append(nodes, status)
	}

	// 9. Cloud services - functions, storage, databases, API gateways
	for _, fn := range d.CloudFunctions {
		nodes = append(nodes, fn)
	}
	for _, storage := range d.CloudStorages {
		nodes = append(nodes, storage)
	}
	for _, db := range d.CloudDatabases {
		nodes = append(nodes, db)
	}
	for _, queue := range d.CloudQueues {
		nodes = append(nodes, queue)
	}
	for _, gw := range d.CloudAPIGateways {
		nodes = append(nodes, gw)
	}
	for _, cdn := range d.CloudCDNs {
		nodes = append(nodes, cdn)
	}
	for _, dnsZone := range d.CloudDNSZones {
		nodes = append(nodes, dnsZone)
	}
	for _, cloudCert := range d.CloudCertificates {
		nodes = append(nodes, cloudCert)
	}
	for _, kms := range d.CloudKMSKeys {
		nodes = append(nodes, kms)
	}
	for _, trail := range d.CloudTrails {
		nodes = append(nodes, trail)
	}
	for _, metric := range d.CloudMetrics {
		nodes = append(nodes, metric)
	}
	for _, alarm := range d.CloudAlarms {
		nodes = append(nodes, alarm)
	}

	// 10. AI/LLM infrastructure - registries, models, deployments
	for _, registry := range d.ModelRegistries {
		nodes = append(nodes, registry)
	}
	for _, llm := range d.LLMs {
		nodes = append(nodes, llm)
	}
	for _, version := range d.ModelVersions {
		nodes = append(nodes, version)
	}
	for _, deployment := range d.LLMDeployments {
		nodes = append(nodes, deployment)
	}
	for _, endpoint := range d.InferenceEndpoints {
		nodes = append(nodes, endpoint)
	}
	for _, embedding := range d.EmbeddingModels {
		nodes = append(nodes, embedding)
	}
	for _, dataset := range d.Datasets {
		nodes = append(nodes, dataset)
	}
	for _, finetune := range d.FineTunes {
		nodes = append(nodes, finetune)
	}
	for _, training := range d.TrainingRuns {
		nodes = append(nodes, training)
	}
	for _, batch := range d.BatchJobs {
		nodes = append(nodes, batch)
	}
	for _, prompt := range d.Prompts {
		nodes = append(nodes, prompt)
	}
	for _, sysPrompt := range d.SystemPrompts {
		nodes = append(nodes, sysPrompt)
	}
	for _, guardrail := range d.Guardrails {
		nodes = append(nodes, guardrail)
	}
	for _, filter := range d.ContentFilters {
		nodes = append(nodes, filter)
	}
	for _, resp := range d.LLMResponses {
		nodes = append(nodes, resp)
	}
	for _, usage := range d.TokenUsages {
		nodes = append(nodes, usage)
	}

	// 11. AI agents - agents, configs, tools, workflows
	for _, agent := range d.AIAgents {
		nodes = append(nodes, agent)
	}
	for _, config := range d.AgentConfigs {
		nodes = append(nodes, config)
	}
	for _, tool := range d.AgentTools {
		nodes = append(nodes, tool)
	}
	for _, chain := range d.Chains {
		nodes = append(nodes, chain)
	}
	for _, workflow := range d.Workflows {
		nodes = append(nodes, workflow)
	}
	for _, crew := range d.Crews {
		nodes = append(nodes, crew)
	}
	for _, task := range d.AgentTasks {
		nodes = append(nodes, task)
	}
	for _, role := range d.AgentRoles {
		nodes = append(nodes, role)
	}
	for _, call := range d.ToolCalls {
		nodes = append(nodes, call)
	}
	for _, step := range d.ReasoningSteps {
		nodes = append(nodes, step)
	}
	for _, plan := range d.PlanningSteps {
		nodes = append(nodes, plan)
	}
	for _, loop := range d.AgentLoops {
		nodes = append(nodes, loop)
	}
	for _, memory := range d.AgentMemories {
		nodes = append(nodes, memory)
	}
	for _, entry := range d.MemoryEntries {
		nodes = append(nodes, entry)
	}
	for _, artifact := range d.AgentArtifacts {
		nodes = append(nodes, artifact)
	}

	// 12. MCP servers and resources
	for _, server := range d.MCPServers {
		nodes = append(nodes, server)
	}
	for _, client := range d.MCPClients {
		nodes = append(nodes, client)
	}
	for _, transport := range d.MCPTransports {
		nodes = append(nodes, transport)
	}
	for _, cap := range d.MCPCapabilities {
		nodes = append(nodes, cap)
	}
	for _, tool := range d.MCPTools {
		nodes = append(nodes, tool)
	}
	for _, resource := range d.MCPResources {
		nodes = append(nodes, resource)
	}
	for _, prompt := range d.MCPPrompts {
		nodes = append(nodes, prompt)
	}
	for _, sampling := range d.MCPSamplings {
		nodes = append(nodes, sampling)
	}
	for _, root := range d.MCPRoots {
		nodes = append(nodes, root)
	}

	// 13. RAG systems - vector stores, documents, pipelines
	for _, store := range d.VectorStores {
		nodes = append(nodes, store)
	}
	for _, index := range d.VectorIndexes {
		nodes = append(nodes, index)
	}
	for _, kb := range d.KnowledgeBases {
		nodes = append(nodes, kb)
	}
	for _, doc := range d.Documents {
		nodes = append(nodes, doc)
	}
	for _, chunk := range d.DocumentChunks {
		nodes = append(nodes, chunk)
	}
	for _, strategy := range d.ChunkingStrategies {
		nodes = append(nodes, strategy)
	}
	for _, embedding := range d.Embeddings {
		nodes = append(nodes, embedding)
	}
	for _, retriever := range d.Retrievers {
		nodes = append(nodes, retriever)
	}
	for _, reranker := range d.Rerankers {
		nodes = append(nodes, reranker)
	}
	for _, pipeline := range d.RAGPipelines {
		nodes = append(nodes, pipeline)
	}
	for _, result := range d.RetrievalResults {
		nodes = append(nodes, result)
	}

	// 14. Data resources - databases, tables, storage, queues
	for _, db := range d.Databases {
		nodes = append(nodes, db)
	}
	for _, schema := range d.Schemas {
		nodes = append(nodes, schema)
	}
	for _, table := range d.Tables {
		nodes = append(nodes, table)
	}
	for _, col := range d.Columns {
		nodes = append(nodes, col)
	}
	for _, idx := range d.Indexes {
		nodes = append(nodes, idx)
	}
	for _, view := range d.Views {
		nodes = append(nodes, view)
	}
	for _, sp := range d.StoredProcedures {
		nodes = append(nodes, sp)
	}
	for _, trigger := range d.Triggers {
		nodes = append(nodes, trigger)
	}
	for _, bucket := range d.StorageBuckets {
		nodes = append(nodes, bucket)
	}
	for _, obj := range d.Objects {
		nodes = append(nodes, obj)
	}
	for _, file := range d.Files {
		nodes = append(nodes, file)
	}
	for _, queue := range d.Queues {
		nodes = append(nodes, queue)
	}
	for _, topic := range d.Topics {
		nodes = append(nodes, topic)
	}
	for _, stream := range d.Streams {
		nodes = append(nodes, stream)
	}
	for _, cache := range d.Caches {
		nodes = append(nodes, cache)
	}
	for _, pipeline := range d.DataPipelines {
		nodes = append(nodes, pipeline)
	}

	// 15. K8s resources - configs, secrets, RBAC, policies
	for _, cm := range d.K8sConfigMaps {
		nodes = append(nodes, cm)
	}
	for _, secret := range d.K8sSecrets {
		nodes = append(nodes, secret)
	}
	for _, pvc := range d.K8sPVCs {
		nodes = append(nodes, pvc)
	}
	for _, pv := range d.K8sPVs {
		nodes = append(nodes, pv)
	}
	for _, sa := range d.K8sServiceAccounts {
		nodes = append(nodes, sa)
	}
	for _, role := range d.K8sRoles {
		nodes = append(nodes, role)
	}
	for _, rb := range d.K8sRoleBindings {
		nodes = append(nodes, rb)
	}
	for _, cr := range d.K8sClusterRoles {
		nodes = append(nodes, cr)
	}
	for _, crb := range d.K8sClusterRoleBindings {
		nodes = append(nodes, crb)
	}
	for _, np := range d.K8sNetworkPolicies {
		nodes = append(nodes, np)
	}
	for _, lr := range d.K8sLimitRanges {
		nodes = append(nodes, lr)
	}
	for _, quota := range d.K8sResourceQuotas {
		nodes = append(nodes, quota)
	}

	// 16. Identity and access - users, groups, roles, credentials
	for _, idp := range d.IdentityProviders {
		nodes = append(nodes, idp)
	}
	for _, saml := range d.SAMLProviders {
		nodes = append(nodes, saml)
	}
	for _, user := range d.Users {
		nodes = append(nodes, user)
	}
	for _, group := range d.Groups {
		nodes = append(nodes, group)
	}
	for _, sa := range d.ServiceAccounts {
		nodes = append(nodes, sa)
	}
	for _, role := range d.Roles {
		nodes = append(nodes, role)
	}
	for _, perm := range d.Permissions {
		nodes = append(nodes, perm)
	}
	for _, policy := range d.Policies {
		nodes = append(nodes, policy)
	}
	for _, cloudRole := range d.CloudIAMRoles {
		nodes = append(nodes, cloudRole)
	}
	for _, cloudPolicy := range d.CloudIAMPolicies {
		nodes = append(nodes, cloudPolicy)
	}
	for _, cred := range d.Credentials {
		nodes = append(nodes, cred)
	}
	for _, key := range d.APIKeys {
		nodes = append(nodes, key)
	}
	for _, token := range d.Tokens {
		nodes = append(nodes, token)
	}
	for _, accessKey := range d.AccessKeys {
		nodes = append(nodes, accessKey)
	}
	for _, mfa := range d.MFADevices {
		nodes = append(nodes, mfa)
	}
	for _, client := range d.OAuthClients {
		nodes = append(nodes, client)
	}
	for _, scope := range d.OAuthScopes {
		nodes = append(nodes, scope)
	}
	for _, session := range d.Sessions {
		nodes = append(nodes, session)
	}

	// 17. Security findings and attack techniques
	for _, finding := range d.Findings {
		nodes = append(nodes, finding)
	}
	for _, tactic := range d.Tactics {
		nodes = append(nodes, tactic)
	}
	for _, technique := range d.Techniques {
		nodes = append(nodes, technique)
	}

	// 18. Custom nodes (preserve order)
	nodes = append(nodes, d.Custom...)

	return nodes
}

// IsEmpty returns true if the discovery result contains no nodes.
func (d *DiscoveryResult) IsEmpty() bool {
	return len(d.Hosts) == 0 &&
		len(d.Ports) == 0 &&
		len(d.Services) == 0 &&
		len(d.Endpoints) == 0 &&
		len(d.Domains) == 0 &&
		len(d.Subdomains) == 0 &&
		len(d.Technologies) == 0 &&
		len(d.Certificates) == 0 &&
		len(d.CloudAssets) == 0 &&
		len(d.APIs) == 0 &&
		// Network (15)
		len(d.DNSRecords) == 0 &&
		len(d.Firewalls) == 0 &&
		len(d.FirewallRules) == 0 &&
		len(d.Routers) == 0 &&
		len(d.Routes) == 0 &&
		len(d.LoadBalancers) == 0 &&
		len(d.Proxies) == 0 &&
		len(d.VPNs) == 0 &&
		len(d.Networks) == 0 &&
		len(d.VLANs) == 0 &&
		len(d.NetworkInterfaces) == 0 &&
		len(d.NetworkZones) == 0 &&
		len(d.NetworkACLs) == 0 &&
		len(d.NATGateways) == 0 &&
		len(d.BGPPeers) == 0 &&
		// Web/API (16)
		len(d.APIEndpoints) == 0 &&
		len(d.Parameters) == 0 &&
		len(d.Headers) == 0 &&
		len(d.Cookies) == 0 &&
		len(d.Forms) == 0 &&
		len(d.FormFields) == 0 &&
		len(d.WebSockets) == 0 &&
		len(d.GraphQLSchemas) == 0 &&
		len(d.GraphQLQueries) == 0 &&
		len(d.GraphQLMutations) == 0 &&
		len(d.RESTResources) == 0 &&
		len(d.CORSPolicies) == 0 &&
		len(d.RateLimits) == 0 &&
		len(d.RequestBodies) == 0 &&
		len(d.Responses) == 0 &&
		len(d.StatusCodes) == 0 &&
		// Identity (16)
		len(d.Users) == 0 &&
		len(d.Groups) == 0 &&
		len(d.Roles) == 0 &&
		len(d.Permissions) == 0 &&
		len(d.Policies) == 0 &&
		len(d.Credentials) == 0 &&
		len(d.APIKeys) == 0 &&
		len(d.Tokens) == 0 &&
		len(d.OAuthClients) == 0 &&
		len(d.OAuthScopes) == 0 &&
		len(d.SAMLProviders) == 0 &&
		len(d.IdentityProviders) == 0 &&
		len(d.ServiceAccounts) == 0 &&
		len(d.Sessions) == 0 &&
		len(d.AccessKeys) == 0 &&
		len(d.MFADevices) == 0 &&
		// AI/LLM (16)
		len(d.LLMs) == 0 &&
		len(d.LLMDeployments) == 0 &&
		len(d.Prompts) == 0 &&
		len(d.SystemPrompts) == 0 &&
		len(d.Guardrails) == 0 &&
		len(d.ContentFilters) == 0 &&
		len(d.LLMResponses) == 0 &&
		len(d.TokenUsages) == 0 &&
		len(d.EmbeddingModels) == 0 &&
		len(d.FineTunes) == 0 &&
		len(d.ModelRegistries) == 0 &&
		len(d.ModelVersions) == 0 &&
		len(d.InferenceEndpoints) == 0 &&
		len(d.BatchJobs) == 0 &&
		len(d.TrainingRuns) == 0 &&
		len(d.Datasets) == 0 &&
		// AI Agent (15)
		len(d.AIAgents) == 0 &&
		len(d.AgentConfigs) == 0 &&
		len(d.AgentMemories) == 0 &&
		len(d.AgentTools) == 0 &&
		len(d.Chains) == 0 &&
		len(d.Workflows) == 0 &&
		len(d.Crews) == 0 &&
		len(d.AgentTasks) == 0 &&
		len(d.AgentRoles) == 0 &&
		len(d.ToolCalls) == 0 &&
		len(d.ReasoningSteps) == 0 &&
		len(d.MemoryEntries) == 0 &&
		len(d.AgentLoops) == 0 &&
		len(d.PlanningSteps) == 0 &&
		len(d.AgentArtifacts) == 0 &&
		// MCP (9)
		len(d.MCPServers) == 0 &&
		len(d.MCPTools) == 0 &&
		len(d.MCPResources) == 0 &&
		len(d.MCPPrompts) == 0 &&
		len(d.MCPClients) == 0 &&
		len(d.MCPTransports) == 0 &&
		len(d.MCPCapabilities) == 0 &&
		len(d.MCPSamplings) == 0 &&
		len(d.MCPRoots) == 0 &&
		// RAG (11)
		len(d.VectorStores) == 0 &&
		len(d.VectorIndexes) == 0 &&
		len(d.Documents) == 0 &&
		len(d.DocumentChunks) == 0 &&
		len(d.KnowledgeBases) == 0 &&
		len(d.Retrievers) == 0 &&
		len(d.RAGPipelines) == 0 &&
		len(d.Embeddings) == 0 &&
		len(d.Rerankers) == 0 &&
		len(d.ChunkingStrategies) == 0 &&
		len(d.RetrievalResults) == 0 &&
		// Data (16)
		len(d.Databases) == 0 &&
		len(d.Tables) == 0 &&
		len(d.Columns) == 0 &&
		len(d.Indexes) == 0 &&
		len(d.Views) == 0 &&
		len(d.StoredProcedures) == 0 &&
		len(d.Triggers) == 0 &&
		len(d.Files) == 0 &&
		len(d.StorageBuckets) == 0 &&
		len(d.Objects) == 0 &&
		len(d.Queues) == 0 &&
		len(d.Topics) == 0 &&
		len(d.Streams) == 0 &&
		len(d.Caches) == 0 &&
		len(d.Schemas) == 0 &&
		len(d.DataPipelines) == 0 &&
		// Container (4)
		len(d.Containers) == 0 &&
		len(d.ContainerImages) == 0 &&
		len(d.ContainerRegistries) == 0 &&
		len(d.Dockerfiles) == 0 &&
		// Kubernetes (22)
		len(d.K8sClusters) == 0 &&
		len(d.K8sNamespaces) == 0 &&
		len(d.K8sPods) == 0 &&
		len(d.K8sDeployments) == 0 &&
		len(d.K8sServices) == 0 &&
		len(d.K8sIngresses) == 0 &&
		len(d.K8sConfigMaps) == 0 &&
		len(d.K8sSecrets) == 0 &&
		len(d.K8sPVCs) == 0 &&
		len(d.K8sPVs) == 0 &&
		len(d.K8sStatefulSets) == 0 &&
		len(d.K8sDaemonSets) == 0 &&
		len(d.K8sJobs) == 0 &&
		len(d.K8sCronJobs) == 0 &&
		len(d.K8sServiceAccounts) == 0 &&
		len(d.K8sRoles) == 0 &&
		len(d.K8sRoleBindings) == 0 &&
		len(d.K8sClusterRoles) == 0 &&
		len(d.K8sClusterRoleBindings) == 0 &&
		len(d.K8sNetworkPolicies) == 0 &&
		len(d.K8sLimitRanges) == 0 &&
		len(d.K8sResourceQuotas) == 0 &&
		// Cloud (20)
		len(d.CloudAccounts) == 0 &&
		len(d.CloudVPCs) == 0 &&
		len(d.CloudSubnets) == 0 &&
		len(d.CloudSecurityGroups) == 0 &&
		len(d.CloudInstances) == 0 &&
		len(d.CloudFunctions) == 0 &&
		len(d.CloudStorages) == 0 &&
		len(d.CloudDatabases) == 0 &&
		len(d.CloudQueues) == 0 &&
		len(d.CloudAPIGateways) == 0 &&
		len(d.CloudCDNs) == 0 &&
		len(d.CloudDNSZones) == 0 &&
		len(d.CloudCertificates) == 0 &&
		len(d.CloudKMSKeys) == 0 &&
		len(d.CloudIAMRoles) == 0 &&
		len(d.CloudIAMPolicies) == 0 &&
		len(d.CloudTrails) == 0 &&
		len(d.CloudMetrics) == 0 &&
		len(d.CloudAlarms) == 0 &&
		len(d.CloudRegions) == 0 &&
		// Attack (2)
		len(d.Tactics) == 0 &&
		len(d.Techniques) == 0 &&
		// Findings
		len(d.Findings) == 0 &&
		// Custom
		len(d.Custom) == 0
}

// NodeCount returns the total number of nodes in this discovery result.
func (d *DiscoveryResult) NodeCount() int {
	return len(d.Hosts) +
		len(d.Ports) +
		len(d.Services) +
		len(d.Endpoints) +
		len(d.Domains) +
		len(d.Subdomains) +
		len(d.Technologies) +
		len(d.Certificates) +
		len(d.CloudAssets) +
		len(d.APIs) +
		// Network (15)
		len(d.DNSRecords) +
		len(d.Firewalls) +
		len(d.FirewallRules) +
		len(d.Routers) +
		len(d.Routes) +
		len(d.LoadBalancers) +
		len(d.Proxies) +
		len(d.VPNs) +
		len(d.Networks) +
		len(d.VLANs) +
		len(d.NetworkInterfaces) +
		len(d.NetworkZones) +
		len(d.NetworkACLs) +
		len(d.NATGateways) +
		len(d.BGPPeers) +
		// Web/API (16)
		len(d.APIEndpoints) +
		len(d.Parameters) +
		len(d.Headers) +
		len(d.Cookies) +
		len(d.Forms) +
		len(d.FormFields) +
		len(d.WebSockets) +
		len(d.GraphQLSchemas) +
		len(d.GraphQLQueries) +
		len(d.GraphQLMutations) +
		len(d.RESTResources) +
		len(d.CORSPolicies) +
		len(d.RateLimits) +
		len(d.RequestBodies) +
		len(d.Responses) +
		len(d.StatusCodes) +
		// Identity (16)
		len(d.Users) +
		len(d.Groups) +
		len(d.Roles) +
		len(d.Permissions) +
		len(d.Policies) +
		len(d.Credentials) +
		len(d.APIKeys) +
		len(d.Tokens) +
		len(d.OAuthClients) +
		len(d.OAuthScopes) +
		len(d.SAMLProviders) +
		len(d.IdentityProviders) +
		len(d.ServiceAccounts) +
		len(d.Sessions) +
		len(d.AccessKeys) +
		len(d.MFADevices) +
		// AI/LLM (16)
		len(d.LLMs) +
		len(d.LLMDeployments) +
		len(d.Prompts) +
		len(d.SystemPrompts) +
		len(d.Guardrails) +
		len(d.ContentFilters) +
		len(d.LLMResponses) +
		len(d.TokenUsages) +
		len(d.EmbeddingModels) +
		len(d.FineTunes) +
		len(d.ModelRegistries) +
		len(d.ModelVersions) +
		len(d.InferenceEndpoints) +
		len(d.BatchJobs) +
		len(d.TrainingRuns) +
		len(d.Datasets) +
		// AI Agent (15)
		len(d.AIAgents) +
		len(d.AgentConfigs) +
		len(d.AgentMemories) +
		len(d.AgentTools) +
		len(d.Chains) +
		len(d.Workflows) +
		len(d.Crews) +
		len(d.AgentTasks) +
		len(d.AgentRoles) +
		len(d.ToolCalls) +
		len(d.ReasoningSteps) +
		len(d.MemoryEntries) +
		len(d.AgentLoops) +
		len(d.PlanningSteps) +
		len(d.AgentArtifacts) +
		// MCP (9)
		len(d.MCPServers) +
		len(d.MCPTools) +
		len(d.MCPResources) +
		len(d.MCPPrompts) +
		len(d.MCPClients) +
		len(d.MCPTransports) +
		len(d.MCPCapabilities) +
		len(d.MCPSamplings) +
		len(d.MCPRoots) +
		// RAG (11)
		len(d.VectorStores) +
		len(d.VectorIndexes) +
		len(d.Documents) +
		len(d.DocumentChunks) +
		len(d.KnowledgeBases) +
		len(d.Retrievers) +
		len(d.RAGPipelines) +
		len(d.Embeddings) +
		len(d.Rerankers) +
		len(d.ChunkingStrategies) +
		len(d.RetrievalResults) +
		// Data (16)
		len(d.Databases) +
		len(d.Tables) +
		len(d.Columns) +
		len(d.Indexes) +
		len(d.Views) +
		len(d.StoredProcedures) +
		len(d.Triggers) +
		len(d.Files) +
		len(d.StorageBuckets) +
		len(d.Objects) +
		len(d.Queues) +
		len(d.Topics) +
		len(d.Streams) +
		len(d.Caches) +
		len(d.Schemas) +
		len(d.DataPipelines) +
		// Container (4)
		len(d.Containers) +
		len(d.ContainerImages) +
		len(d.ContainerRegistries) +
		len(d.Dockerfiles) +
		// Kubernetes (22)
		len(d.K8sClusters) +
		len(d.K8sNamespaces) +
		len(d.K8sPods) +
		len(d.K8sDeployments) +
		len(d.K8sServices) +
		len(d.K8sIngresses) +
		len(d.K8sConfigMaps) +
		len(d.K8sSecrets) +
		len(d.K8sPVCs) +
		len(d.K8sPVs) +
		len(d.K8sStatefulSets) +
		len(d.K8sDaemonSets) +
		len(d.K8sJobs) +
		len(d.K8sCronJobs) +
		len(d.K8sServiceAccounts) +
		len(d.K8sRoles) +
		len(d.K8sRoleBindings) +
		len(d.K8sClusterRoles) +
		len(d.K8sClusterRoleBindings) +
		len(d.K8sNetworkPolicies) +
		len(d.K8sLimitRanges) +
		len(d.K8sResourceQuotas) +
		// Cloud (20)
		len(d.CloudAccounts) +
		len(d.CloudVPCs) +
		len(d.CloudSubnets) +
		len(d.CloudSecurityGroups) +
		len(d.CloudInstances) +
		len(d.CloudFunctions) +
		len(d.CloudStorages) +
		len(d.CloudDatabases) +
		len(d.CloudQueues) +
		len(d.CloudAPIGateways) +
		len(d.CloudCDNs) +
		len(d.CloudDNSZones) +
		len(d.CloudCertificates) +
		len(d.CloudKMSKeys) +
		len(d.CloudIAMRoles) +
		len(d.CloudIAMPolicies) +
		len(d.CloudTrails) +
		len(d.CloudMetrics) +
		len(d.CloudAlarms) +
		len(d.CloudRegions) +
		// Attack (2)
		len(d.Tactics) +
		len(d.Techniques) +
		// Findings
		len(d.Findings) +
		// Custom
		len(d.Custom)
}

// NewDiscoveryResult creates an empty DiscoveryResult with all slices initialized.
func NewDiscoveryResult() *DiscoveryResult {
	return &DiscoveryResult{
		Hosts:        make([]*Host, 0),
		Ports:        make([]*Port, 0),
		Services:     make([]*Service, 0),
		Endpoints:    make([]*Endpoint, 0),
		Domains:      make([]*Domain, 0),
		Subdomains:   make([]*Subdomain, 0),
		Technologies: make([]*Technology, 0),
		Certificates: make([]*Certificate, 0),
		CloudAssets:  make([]*CloudAsset, 0),
		APIs:         make([]*API, 0),
		Custom:       make([]GraphNode, 0),
	}
}
