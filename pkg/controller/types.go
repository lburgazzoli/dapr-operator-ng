package controller

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"

	DaprReleaseGeneration = "daprs.tools.dapr.io/release.generation"
	DaprReleaseName       = "daprs.tools.dapr.io/release.name"
	DaprReleaseNamespace  = "daprs.tools.dapr.io/release.namespace"

	FinalizerName = "tools.dapr.io/finalizer"
	FieldManager  = "dapr-release-controller"
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
