/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dapr

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	daprvApi "github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
)

func NewReconciler(manager ctrl.Manager, o HelmOptions) (*Reconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		return nil, err
	}

	rec := Reconciler{}
	rec.l = ctrl.Log.WithName("controller")
	rec.Client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = controller.ClusterTypeVanilla

	isOpenshift, err := c.IsOpenShift()
	if err != nil {
		return nil, err
	}
	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	rec.actions = append(rec.actions, NewUpdateAction())

	hc, err := loader.Load(o.ChartsDir)
	if err != nil {
		return nil, err
	}

	rec.c = hc
	if rec.c.Values == nil {
		rec.c.Values = make(map[string]interface{})
	}

	rec.c.Values["dapr_operator.runAsNonRoot"] = "true"
	rec.c.Values["dapr_placement.runAsNonRoot"] = "true"
	rec.c.Values["dapr_sentry.runAsNonRoot"] = "true"
	rec.c.Values["dapr_dashboard.runAsNonRoot"] = "true"

	return &rec, nil
}

//+kubebuilder:rbac:groups=dapr.dapr.io,resources=daprs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dapr.dapr.io,resources=daprs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dapr.dapr.io,resources=daprs/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=*
//+kubebuilder:rbac:groups="",resources=secrets,verbs=*
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions/finalizers,verbs=*

type Reconciler struct {
	*client.Client

	Scheme      *runtime.Scheme
	ClusterType controller.ClusterType
	actions     []Action
	l           logr.Logger
	c           *chart.Chart
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)

	c = c.For(&daprvApi.Dapr{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))

	for i := range r.actions {
		b, err := r.actions[i].Configure(ctx, r.Client, c)
		if err != nil {
			return err
		}

		c = b
	}

	return c.Complete(r)
}
