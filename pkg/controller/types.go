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

	DaprResourceGeneration = "daprs.dapr.io/resource.generation"
	DaprResourceName       = "daprs.dapr.io/resource.name"
	DaprResourceNamespace  = "daprs.dapr.io/resource.namespace"

	SyncInterval     = 5 * time.Second
	RetryInterval    = 10 * time.Second
	ConflictInterval = 1 * time.Second
	FinalizerName    = "dapr.io/finalizer"
	FieldManager     = "dapr-controller"
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
