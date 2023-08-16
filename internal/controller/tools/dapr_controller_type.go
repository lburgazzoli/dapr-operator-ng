package tools

import (
	"context"

	daprApi "github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	// HelmChartsDir is the relative directory.
	HelmChartsDir = "helm-charts/dapr"
)

type HelmOptions struct {
	ChartsDir string
}

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	Reconciler  *Reconciler
	ClusterType controller.ClusterType
	Resource    *daprApi.Dapr
	Chart       *chart.Chart
	Overrides   map[string]interface{}
}

type Action interface {
	Configure(context.Context, *client.Client, *builder.Builder) (*builder.Builder, error)
	Run(context.Context, *ReconciliationRequest) error
	Cleanup(context.Context, *ReconciliationRequest) error
}
