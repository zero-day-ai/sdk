package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// K8sCluster represents a Kubernetes cluster in the knowledge graph.
//
// Identifying Properties:
//   - name (required): The cluster name
//
// Relationships:
//   - None (root node)
//   - Children: K8sNamespace, K8sPV, K8sClusterRole, K8sClusterRoleBinding
type K8sCluster struct {
	Name       string // Identifying
	APIServer  string
	Version    string
	Provider   string // "eks", "gke", "aks", "openshift", "vanilla"
	Region     string
	NodeCount  int
	Kubeconfig string
}

func (k *K8sCluster) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sCluster) IdentifyingProperties() map[string]any {
	return map[string]any{graphrag.PropName: k.Name}
}
func (k *K8sCluster) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: k.Name,
		"type":            "k8s_cluster",
	}
	if k.APIServer != "" {
		props["api_server"] = k.APIServer
	}
	if k.Version != "" {
		props["version"] = k.Version
	}
	if k.Provider != "" {
		props["provider"] = k.Provider
	}
	if k.Region != "" {
		props["region"] = k.Region
	}
	if k.NodeCount > 0 {
		props["node_count"] = k.NodeCount
	}
	if k.Kubeconfig != "" {
		props["kubeconfig"] = k.Kubeconfig
	}
	return props
}
func (k *K8sCluster) ParentRef() *NodeRef      { return nil }
func (k *K8sCluster) RelationshipType() string { return "" }

// K8sNamespace represents a Kubernetes namespace in the knowledge graph.
//
// Identifying Properties:
//   - cluster_id (required): The parent cluster name
//   - name (required): The namespace name
//
// Relationships:
//   - Parent: K8sCluster (via PART_OF relationship)
type K8sNamespace struct {
	ClusterID   string // Identifying (parent reference)
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Status      string // "Active", "Terminating"
}

func (k *K8sNamespace) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sNamespace) IdentifyingProperties() map[string]any {
	return map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sNamespace) Properties() map[string]any {
	props := map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
		"type":            "k8s_namespace",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Status != "" {
		props["status"] = k.Status
	}
	return props
}
func (k *K8sNamespace) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{graphrag.PropName: k.ClusterID, "type": "k8s_cluster"},
	}
}
func (k *K8sNamespace) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sPod represents a Kubernetes pod in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The pod name
//
// Relationships:
//   - Parent: K8sNamespace (via PART_OF relationship)
type K8sPod struct {
	NamespaceID string // Identifying (parent reference)
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Status      string // "Running", "Pending", "Failed", "Succeeded"
	NodeName    string
	PodIP       string
	HostIP      string
	Containers  []string
}

func (k *K8sPod) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sPod) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sPod) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_pod",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Status != "" {
		props["status"] = k.Status
	}
	if k.NodeName != "" {
		props["node_name"] = k.NodeName
	}
	if k.PodIP != "" {
		props["pod_ip"] = k.PodIP
	}
	if k.HostIP != "" {
		props["host_ip"] = k.HostIP
	}
	if len(k.Containers) > 0 {
		props["containers"] = k.Containers
	}
	return props
}
func (k *K8sPod) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sPod) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sDeployment represents a Kubernetes deployment in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The deployment name
type K8sDeployment struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Replicas    int32
	Selector    map[string]string
	Strategy    string // "RollingUpdate", "Recreate"
}

func (k *K8sDeployment) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sDeployment) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sDeployment) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_deployment",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Replicas > 0 {
		props["replicas"] = k.Replicas
	}
	if len(k.Selector) > 0 {
		props["selector"] = k.Selector
	}
	if k.Strategy != "" {
		props["strategy"] = k.Strategy
	}
	return props
}
func (k *K8sDeployment) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sDeployment) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sService represents a Kubernetes service in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The service name
type K8sService struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Type        string // "ClusterIP", "NodePort", "LoadBalancer", "ExternalName"
	ClusterIP   string
	Ports       []string // "80/TCP", "443/TCP"
	Selector    map[string]string
}

func (k *K8sService) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sService) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sService) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_service",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Type != "" {
		props["service_type"] = k.Type
	}
	if k.ClusterIP != "" {
		props["cluster_ip"] = k.ClusterIP
	}
	if len(k.Ports) > 0 {
		props["ports"] = k.Ports
	}
	if len(k.Selector) > 0 {
		props["selector"] = k.Selector
	}
	return props
}
func (k *K8sService) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sService) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sIngress represents a Kubernetes ingress in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The ingress name
type K8sIngress struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Hosts       []string
	TLS         []string // TLS secret names
}

func (k *K8sIngress) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sIngress) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sIngress) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_ingress",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Hosts) > 0 {
		props["hosts"] = k.Hosts
	}
	if len(k.TLS) > 0 {
		props["tls"] = k.TLS
	}
	return props
}
func (k *K8sIngress) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sIngress) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sConfigMap represents a Kubernetes configmap in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The configmap name
type K8sConfigMap struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Data        map[string]string
}

func (k *K8sConfigMap) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sConfigMap) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sConfigMap) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_configmap",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Data) > 0 {
		props["data_keys"] = getKeys(k.Data) // Store keys only, not full data
	}
	return props
}
func (k *K8sConfigMap) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sConfigMap) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sSecret represents a Kubernetes secret in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The secret name
type K8sSecret struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Type        string   // "Opaque", "kubernetes.io/tls", etc.
	DataKeys    []string // Store keys only, never actual secret values
}

func (k *K8sSecret) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sSecret) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sSecret) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_secret",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Type != "" {
		props["secret_type"] = k.Type
	}
	if len(k.DataKeys) > 0 {
		props["data_keys"] = k.DataKeys
	}
	return props
}
func (k *K8sSecret) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sSecret) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sPVC represents a Kubernetes PersistentVolumeClaim in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The PVC name
type K8sPVC struct {
	NamespaceID  string // Identifying
	Name         string // Identifying
	Labels       map[string]string
	Annotations  map[string]string
	StorageClass string
	Size         string // "10Gi", "100Gi"
	Status       string // "Bound", "Pending", "Lost"
	VolumeName   string
}

func (k *K8sPVC) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sPVC) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sPVC) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_pvc",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.StorageClass != "" {
		props["storage_class"] = k.StorageClass
	}
	if k.Size != "" {
		props["size"] = k.Size
	}
	if k.Status != "" {
		props["status"] = k.Status
	}
	if k.VolumeName != "" {
		props["volume_name"] = k.VolumeName
	}
	return props
}
func (k *K8sPVC) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sPVC) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sPV represents a Kubernetes PersistentVolume in the knowledge graph.
//
// Identifying Properties:
//   - cluster_id (required): The parent cluster name
//   - name (required): The PV name
type K8sPV struct {
	ClusterID     string // Identifying
	Name          string // Identifying
	Labels        map[string]string
	Annotations   map[string]string
	StorageClass  string
	Size          string // "10Gi", "100Gi"
	Status        string // "Available", "Bound", "Released", "Failed"
	ReclaimPolicy string // "Retain", "Delete", "Recycle"
}

func (k *K8sPV) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sPV) IdentifyingProperties() map[string]any {
	return map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sPV) Properties() map[string]any {
	props := map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
		"type":            "k8s_pv",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.StorageClass != "" {
		props["storage_class"] = k.StorageClass
	}
	if k.Size != "" {
		props["size"] = k.Size
	}
	if k.Status != "" {
		props["status"] = k.Status
	}
	if k.ReclaimPolicy != "" {
		props["reclaim_policy"] = k.ReclaimPolicy
	}
	return props
}
func (k *K8sPV) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{graphrag.PropName: k.ClusterID, "type": "k8s_cluster"},
	}
}
func (k *K8sPV) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sStatefulSet represents a Kubernetes statefulset in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The statefulset name
type K8sStatefulSet struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Replicas    int32
	Selector    map[string]string
	ServiceName string
}

func (k *K8sStatefulSet) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sStatefulSet) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sStatefulSet) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_statefulset",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Replicas > 0 {
		props["replicas"] = k.Replicas
	}
	if len(k.Selector) > 0 {
		props["selector"] = k.Selector
	}
	if k.ServiceName != "" {
		props["service_name"] = k.ServiceName
	}
	return props
}
func (k *K8sStatefulSet) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sStatefulSet) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sDaemonSet represents a Kubernetes daemonset in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The daemonset name
type K8sDaemonSet struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Selector    map[string]string
}

func (k *K8sDaemonSet) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sDaemonSet) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sDaemonSet) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_daemonset",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Selector) > 0 {
		props["selector"] = k.Selector
	}
	return props
}
func (k *K8sDaemonSet) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sDaemonSet) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sJob represents a Kubernetes job in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The job name
type K8sJob struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Completions int32
	Parallelism int32
	Status      string // "Complete", "Failed", "Active"
}

func (k *K8sJob) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sJob) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sJob) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_job",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Completions > 0 {
		props["completions"] = k.Completions
	}
	if k.Parallelism > 0 {
		props["parallelism"] = k.Parallelism
	}
	if k.Status != "" {
		props["status"] = k.Status
	}
	return props
}
func (k *K8sJob) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sJob) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sCronJob represents a Kubernetes cronjob in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The cronjob name
type K8sCronJob struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Schedule    string // Cron schedule expression
	Suspend     bool
}

func (k *K8sCronJob) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sCronJob) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sCronJob) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_cronjob",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.Schedule != "" {
		props["schedule"] = k.Schedule
	}
	props["suspend"] = k.Suspend
	return props
}
func (k *K8sCronJob) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sCronJob) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sServiceAccount represents a Kubernetes service account in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The service account name
type K8sServiceAccount struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Secrets     []string
}

func (k *K8sServiceAccount) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sServiceAccount) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sServiceAccount) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_serviceaccount",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Secrets) > 0 {
		props["secrets"] = k.Secrets
	}
	return props
}
func (k *K8sServiceAccount) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sServiceAccount) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sRole represents a Kubernetes role in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The role name
type K8sRole struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Rules       []string // Simplified representation of policy rules
}

func (k *K8sRole) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sRole) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sRole) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_role",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Rules) > 0 {
		props["rules"] = k.Rules
	}
	return props
}
func (k *K8sRole) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sRole) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sRoleBinding represents a Kubernetes role binding in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The rolebinding name
type K8sRoleBinding struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	RoleRef     string   // Name of the role
	Subjects    []string // ServiceAccounts, Users, Groups
}

func (k *K8sRoleBinding) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sRoleBinding) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sRoleBinding) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_rolebinding",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.RoleRef != "" {
		props["role_ref"] = k.RoleRef
	}
	if len(k.Subjects) > 0 {
		props["subjects"] = k.Subjects
	}
	return props
}
func (k *K8sRoleBinding) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sRoleBinding) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sClusterRole represents a Kubernetes cluster role in the knowledge graph.
//
// Identifying Properties:
//   - cluster_id (required): The parent cluster name
//   - name (required): The cluster role name
type K8sClusterRole struct {
	ClusterID   string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Rules       []string
}

func (k *K8sClusterRole) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sClusterRole) IdentifyingProperties() map[string]any {
	return map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sClusterRole) Properties() map[string]any {
	props := map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
		"type":            "k8s_clusterrole",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Rules) > 0 {
		props["rules"] = k.Rules
	}
	return props
}
func (k *K8sClusterRole) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{graphrag.PropName: k.ClusterID, "type": "k8s_cluster"},
	}
}
func (k *K8sClusterRole) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sClusterRoleBinding represents a Kubernetes cluster role binding in the knowledge graph.
//
// Identifying Properties:
//   - cluster_id (required): The parent cluster name
//   - name (required): The cluster rolebinding name
type K8sClusterRoleBinding struct {
	ClusterID   string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	RoleRef     string
	Subjects    []string
}

func (k *K8sClusterRoleBinding) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sClusterRoleBinding) IdentifyingProperties() map[string]any {
	return map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sClusterRoleBinding) Properties() map[string]any {
	props := map[string]any{
		"cluster_id":      k.ClusterID,
		graphrag.PropName: k.Name,
		"type":            "k8s_clusterrolebinding",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if k.RoleRef != "" {
		props["role_ref"] = k.RoleRef
	}
	if len(k.Subjects) > 0 {
		props["subjects"] = k.Subjects
	}
	return props
}
func (k *K8sClusterRoleBinding) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{graphrag.PropName: k.ClusterID, "type": "k8s_cluster"},
	}
}
func (k *K8sClusterRoleBinding) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sNetworkPolicy represents a Kubernetes network policy in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The network policy name
type K8sNetworkPolicy struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	PodSelector map[string]string
	PolicyTypes []string // "Ingress", "Egress"
}

func (k *K8sNetworkPolicy) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sNetworkPolicy) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sNetworkPolicy) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_networkpolicy",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.PodSelector) > 0 {
		props["pod_selector"] = k.PodSelector
	}
	if len(k.PolicyTypes) > 0 {
		props["policy_types"] = k.PolicyTypes
	}
	return props
}
func (k *K8sNetworkPolicy) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sNetworkPolicy) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sLimitRange represents a Kubernetes limit range in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The limit range name
type K8sLimitRange struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Limits      []string // Simplified representation of limit specifications
}

func (k *K8sLimitRange) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sLimitRange) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sLimitRange) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_limitrange",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Limits) > 0 {
		props["limits"] = k.Limits
	}
	return props
}
func (k *K8sLimitRange) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sLimitRange) RelationshipType() string { return graphrag.RelTypePartOf }

// K8sResourceQuota represents a Kubernetes resource quota in the knowledge graph.
//
// Identifying Properties:
//   - namespace_id (required): The parent namespace name
//   - name (required): The resource quota name
type K8sResourceQuota struct {
	NamespaceID string // Identifying
	Name        string // Identifying
	Labels      map[string]string
	Annotations map[string]string
	Hard        map[string]string // Resource limits
	Used        map[string]string // Current usage
}

func (k *K8sResourceQuota) NodeType() string { return graphrag.NodeTypeCloudAsset }
func (k *K8sResourceQuota) IdentifyingProperties() map[string]any {
	return map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
	}
}
func (k *K8sResourceQuota) Properties() map[string]any {
	props := map[string]any{
		"namespace_id":    k.NamespaceID,
		graphrag.PropName: k.Name,
		"type":            "k8s_resourcequota",
	}
	if len(k.Labels) > 0 {
		props["labels"] = k.Labels
	}
	if len(k.Annotations) > 0 {
		props["annotations"] = k.Annotations
	}
	if len(k.Hard) > 0 {
		props["hard"] = k.Hard
	}
	if len(k.Used) > 0 {
		props["used"] = k.Used
	}
	return props
}
func (k *K8sResourceQuota) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType:   graphrag.NodeTypeCloudAsset,
		Properties: map[string]any{"namespace_id": k.NamespaceID, graphrag.PropName: k.NamespaceID, "type": "k8s_namespace"},
	}
}
func (k *K8sResourceQuota) RelationshipType() string { return graphrag.RelTypePartOf }

// Helper function to extract keys from a map
func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
