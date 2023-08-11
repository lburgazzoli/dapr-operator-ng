package dapr

import (
	"context"
	"sort"
	"strconv"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/resources"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/defaults"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/helm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewApplyAction() Action {
	return &ApplyAction{
		engine:        helm.NewEngine(),
		l:             ctrl.Log.WithName("action").WithName("apply"),
		subscriptions: make(map[string]struct{}),
	}
}

type ApplyAction struct {
	engine        *helm.Engine
	l             logr.Logger
	subscriptions map[string]struct{}
}

func (a *ApplyAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ApplyAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	items, err := a.engine.Render(rc.Chart, rc.Resource)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	// TODO: this must be ordered by priority/relations
	sort.Slice(items, func(i int, j int) bool {
		return items[i].GroupVersionKind().Kind+":"+items[i].GetName() < items[j].GroupVersionKind().Kind+":"+items[j].GetName()
	})

	for i := range items {
		obj := items[i]

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		resources.Labels(&obj, map[string]string{
			controller.DaprResourceGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			controller.DaprResourceRef:        rc.Resource.Namespace + "-" + rc.Resource.Name,
		})

		if _, ok := dc.(*client.NamespacedResource); ok {
			obj.SetOwnerReferences(resources.OwnerReferences(rc.Resource))
			obj.SetNamespace(rc.Resource.Namespace)

			/*
				r := resources.Ref(&obj)

				if _, ok := a.subscriptions[r]; !ok {
					err = rc.Reconciler.Watch(
						source.Kind(rc.Reconciler.Manager.GetCache(), &obj),
						handler.EnqueueRequestForOwner(
							rc.Reconciler.Manager.GetScheme(),
							rc.Reconciler.Manager.GetRESTMapper(),
							&daprApi.Dapr{},
							handler.OnlyControllerOwner()),
						&predicates.DependentPredicate{},
					)

					if err != nil {
						return err
					}

					a.subscriptions[r] = struct{}{}
				}
			*/
		}

		_, err = dc.Apply(ctx, obj.GetName(), &obj, metav1.ApplyOptions{
			FieldManager: defaults.FieldManager,
		})

		if err != nil {
			return errors.Wrapf(err, "cannot patch object %s", resources.Ref(&obj))
		}

		a.l.Info("apply", "ref", resources.Ref(&obj))
	}

	return nil
}

func (a *ApplyAction) Cleanup(ctx context.Context, rc *ReconciliationRequest) error {
	items, err := a.engine.Render(rc.Chart, rc.Resource)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	for i := range items {
		obj := items[i]

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		// Delete clustered resources
		if _, ok := dc.(*client.ClusteredResource); ok {
			err := dc.Delete(ctx, obj.GetName(), metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			})

			if err != nil {
				return errors.Wrapf(err, "cannot delete object %s", resources.Ref(&obj))
			}

			a.l.Info("delete", "ref", resources.Ref(&obj))
		}
	}

	return nil
}
