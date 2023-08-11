package controller

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"

	KubernetesLabelAppName      = "app.kubernetes.io/name"
	KubernetesLabelAppInstance  = "app.kubernetes.io/instance"
	KubernetesLabelAppComponent = "app.kubernetes.io/component"
	KubernetesLabelAppPartOf    = "app.kubernetes.io/part-of"
	KubernetesLabelAppManagedBy = "app.kubernetes.io/managed-by"

	DaprResourceGeneration = "daprs.dapr.io/resource.generation"
	DaprResourceRef        = "daprs.dapr.io/resource.ref"
)

type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	PprofAddr                     string
	LeaderElectionID              string
	LeaderElectionNamespace       string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
}
