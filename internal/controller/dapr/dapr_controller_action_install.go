package dapr

import (
	"context"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/helm"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewInstallAction() Action {
	return &InstallAction{
		e: helm.NewEngine(),
		l: ctrl.Log.WithName("action").WithName("install"),
	}
}

type InstallAction struct {
	e *helm.Engine
	l logr.Logger
}

func (u *InstallAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	u.l.Info("configure")
	return b, nil
}

func (u *InstallAction) Apply(ctx context.Context, rc *ReconciliationRequest) error {
	u.l.Info("Apply")
	return nil
}

func (u *InstallAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	u.l.Info("Cleanup")
	return nil
}
