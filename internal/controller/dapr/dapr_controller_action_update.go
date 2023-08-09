package dapr

import (
	"context"
	daprApi "github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewUpdateAction() controller.Action[daprApi.Dapr] {
	return &UpdateAction{}
}

type UpdateAction struct {
}

func (u *UpdateAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (u *UpdateAction) Apply(_ context.Context, _ *controller.ReconciliationRequest[daprApi.Dapr]) error {
	return nil
}

func (u *UpdateAction) Cleanup(_ context.Context, _ *controller.ReconciliationRequest[daprApi.Dapr]) error {
	return nil
}
