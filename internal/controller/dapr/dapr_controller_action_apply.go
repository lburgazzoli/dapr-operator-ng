package dapr

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	daprApi "github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/predicates"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/resources"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/helm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		gvk := obj.GroupVersionKind()
		installOnly := a.installOnly(gvk)

		if rc.Resource.Generation != rc.Resource.Status.ObservedGeneration {
			rc.Reconciler.Event(
				rc.Resource,
				corev1.EventTypeNormal,
				"RenderFullHelmTemplate",
				fmt.Sprintf("Render full Helm template as Dapr spec changed (observedGeneration: %d, generation: %d)",
					rc.Resource.Status.ObservedGeneration,
					rc.Resource.Generation),
			)

			installOnly = false
		}

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		resources.Labels(&obj, map[string]string{
			controller.DaprResourceGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			controller.DaprResourceRef:        rc.Resource.Namespace + "-" + rc.Resource.Name,
		})

		switch dc.(type) {
		// NamespacedResource: in this case, filtering with ownership can be implemented
		// as all the namespaced resources created by this controller have the Dapr CR as
		// an owner
		case *client.NamespacedResource:
			obj.SetOwnerReferences(resources.OwnerReferences(rc.Resource))
			obj.SetNamespace(rc.Resource.Namespace)

			r := gvk.GroupVersion().String() + ":" + gvk.Kind

			if _, ok := a.subscriptions[r]; !ok {
				err = rc.Reconciler.Watch(
					&obj,
					rc.Reconciler.EnqueueRequestForOwner(
						&daprApi.Dapr{},
						handler.OnlyControllerOwner()),
					predicate.And(
						&predicates.HasLabel{
							Name: controller.DaprResourceRef,
						},
						&predicates.DependentPredicate{
							WatchDelete: true,
							WatchUpdate: a.watchForUpdates(gvk),
						},
					),
				)

				if err != nil {
					return err
				}

				a.subscriptions[r] = struct{}{}
			}
		// ClusteredResource: in this case, ownership based filtering is not supported
		// as you cannot have a non namespaced owner. For such reason, the resource for
		// which a reconcile should be triggered can be identified by using a label
		//
		//    daprs.dapr.io/resource.ref = ${namespace}-${name}
		//
		case *client.ClusteredResource:
			r := gvk.GroupVersion().String() + ":" + gvk.Kind

			if _, ok := a.subscriptions[r]; !ok {
				err = rc.Reconciler.Watch(
					&obj,
					rc.Reconciler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object ctrlCli.Object) []reconcile.Request {
						labels := object.GetLabels()
						if labels == nil {
							return nil
						}
						ref := labels[controller.DaprResourceRef]
						if ref == "" {
							return nil
						}

						parts := strings.SplitN(ref, "-", 2)
						if len(parts) != 2 {
							return nil
						}

						return []reconcile.Request{{
							NamespacedName: types.NamespacedName{
								Name:      parts[1],
								Namespace: parts[0],
							},
						}}
					}),
					predicate.And(
						&predicates.HasLabel{
							Name: controller.DaprResourceRef,
						},
						&predicates.DependentPredicate{
							WatchDelete: true,
							WatchUpdate: a.watchForUpdates(gvk),
						},
					),
				)

				if err != nil {
					return err
				}

				a.subscriptions[r] = struct{}{}
			}
		}

		if installOnly {
			old, err := dc.Get(ctx, obj.GetName(), metav1.GetOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "cannot get object %s", resources.Ref(&obj))
				}
			}

			if old != nil {
				// Every time the template is rendered, the helm function genSignedCert kicks in and
				// re-generated certs which causes a number os side effects, like deployments restart
				// etc.
				//
				// As consequence some resources are not meant to be watched and re-created unless the
				// Dapr CR is updated.
				a.l.Info("run",
					"apply", "false",
					"ref", resources.Ref(&obj),
					"reason", "resource marked as install-only")

				continue
			}
		}

		_, err = dc.Apply(ctx, obj.GetName(), &obj, metav1.ApplyOptions{
			FieldManager: controller.FieldManager,
			Force:        true,
		})

		if err != nil {
			return errors.Wrapf(err, "cannot patch object %s", resources.Ref(&obj))
		}

		a.l.Info("run",
			"apply", "true",
			"ref", resources.Ref(&obj))
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

func (a *ApplyAction) watchForUpdates(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return false
	}
	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return false
	}

	return true
}

func (a *ApplyAction) installOnly(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return true
	}
	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return true
	}

	return false
}
