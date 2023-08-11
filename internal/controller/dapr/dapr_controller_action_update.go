package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/apply"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/helm"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chartutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewUpdateAction() Action {

	return &UpdateAction{
		e: helm.NewEngine(),
		l: ctrl.Log.WithName("action").WithName("update"),
	}
}

type UpdateAction struct {
	e *helm.Engine
	l logr.Logger
}

func (u *UpdateAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	u.l.Info("configure")
	return b, nil
}

func (u *UpdateAction) Apply(ctx context.Context, rc *ReconciliationRequest) error {
	vals, err := u.values(ctx, rc)
	if err != nil {
		return errors.Wrap(err, "cannot render vals")
	}

	resources, err := u.e.Render(rc.Chart, vals)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	// TODO: this must be ordered by priority/relations
	sort.Slice(resources, func(i int, j int) bool {
		return resources[i].GroupVersionKind().Kind+":"+resources[i].GetName() < resources[j].GroupVersionKind().Kind+":"+resources[j].GetName()
	})

	for i := range resources {
		obj := resources[i]

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		apply.Labels(&obj, map[string]string{
			"daprs.dapr.io/created-by.generation": strconv.FormatInt(rc.Resource.Generation, 10),
			"daprs.dapr.io/created-by.namespace":  rc.Resource.Namespace,
			"daprs.dapr.io/created-by.name":       rc.Resource.Name,
		})

		if _, ok := dc.(*client.NamespacedResource); ok {
			obj.SetOwnerReferences(apply.OwnerReferences(rc.Resource))
			obj.SetNamespace(rc.Resource.Namespace)
		}

		_, err = dc.Apply(ctx, obj.GetName(), &obj, metav1.ApplyOptions{
			FieldManager: defaults.FieldManager,
		})

		if err != nil {
			return errors.Wrapf(err, "cannot patch object %s", u.ref(&obj))
		}

		u.l.Info("apply", "ref", u.ref(&obj))
	}

	return nil
}

func (u *UpdateAction) Cleanup(ctx context.Context, rc *ReconciliationRequest) error {
	vals, err := u.values(ctx, rc)
	if err != nil {
		return errors.Wrap(err, "cannot render vals")
	}

	resources, err := u.e.Render(rc.Chart, vals)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	for i := range resources {
		obj := resources[i]

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
				return errors.Wrapf(err, "cannot delete object %s", u.ref(&obj))
			}

			u.l.Info("delete", "ref", u.ref(&obj))
		}
	}

	return nil
}

func (u *UpdateAction) ref(obj *unstructured.Unstructured) string {
	name := obj.GetName()
	if name == "" {
		name = obj.GetNamespace() + ":" + obj.GetName()
	}

	return fmt.Sprintf(
		"%s:%s:%s",
		obj.GroupVersionKind().Kind,
		obj.GroupVersionKind().GroupVersion().String(),
		name,
	)
}

func (u *UpdateAction) values(_ context.Context, rc *ReconciliationRequest) (chartutil.Values, error) {
	values := make(map[string]interface{})

	if rc.Resource.Spec.Values != nil {
		if err := json.Unmarshal(rc.Resource.Spec.Values.RawMessage, &values); err != nil {
			return chartutil.Values{}, errors.Wrap(err, "unable to decode values")
		}
	}

	if err := chartutil.ProcessDependencies(rc.Chart, values); err != nil {
		return chartutil.Values{}, errors.Wrap(err, "cannot process dependencies")
	}

	return chartutil.ToRenderValues(
		rc.Chart,
		values,
		chartutil.ReleaseOptions{
			Name:      rc.Resource.Name,
			Namespace: rc.Resource.Namespace,
			Revision:  int(rc.Resource.Generation),
			IsInstall: false,
			IsUpgrade: true,
		},
		nil)
}
