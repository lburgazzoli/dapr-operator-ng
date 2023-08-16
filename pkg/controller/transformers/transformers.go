package transformers

import (
	"context"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"k8s.io/apimachinery/pkg/types"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func LabelsToRequest(_ context.Context, object ctrlCli.Object) []reconcile.Request {
	labels := object.GetLabels()
	if labels == nil {
		return nil
	}
	name := labels[controller.DaprReleaseName]
	if name == "" {
		return nil
	}
	namespace := labels[controller.DaprReleaseNamespace]
	if namespace == "" {
		return nil
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}}
}
