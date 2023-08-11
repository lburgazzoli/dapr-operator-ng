package dapr

import (
	"context"
	"encoding/json"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/helm"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
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

	for i := range resources {
		u.l.Info("apply", "resource", resources[i].GroupVersionKind().String())
	}

	return nil
}

func (u *UpdateAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	u.l.Info("Cleanup")
	return nil
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
