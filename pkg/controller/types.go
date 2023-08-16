package controller

import "time"

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"

	KubernetesLabelAppName      = "app.kubernetes.io/name"
	KubernetesLabelAppInstance  = "app.kubernetes.io/instance"
	KubernetesLabelAppComponent = "app.kubernetes.io/component"
	KubernetesLabelAppPartOf    = "app.kubernetes.io/part-of"
	KubernetesLabelAppManagedBy = "app.kubernetes.io/managed-by"

	DaprReleaseGeneration = "daprs.tools.dapr.io/release.generation"
	DaprReleaseName       = "daprs.tools.dapr.io/release.name"
	DaprReleaseNamespace  = "daprs.tools.dapr.io/release.namespace"

	SyncInterval     = 5 * time.Second
	RetryInterval    = 10 * time.Second
	ConflictInterval = 1 * time.Second
	FinalizerName    = "tools.dapr.io/finalizer"
	FieldManager     = "dapr-release-controller"
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
