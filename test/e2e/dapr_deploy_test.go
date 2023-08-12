package e2e

import (
	"testing"

	daprAc "github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/applyconfiguration/dapr/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	"github.com/onsi/gomega"
	"github.com/rs/xid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	ns := test.NewTestNamespace()
	dp := test.Client().Dapr.DaprV1alpha1().Daprs(ns.Name)

	instance, err := dp.Apply(
		test.Ctx(),
		daprAc.Dapr(xid.New().String(), ns.Name).
			WithSpec(daprAc.DaprSpec().
				WithValues(nil),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-test",
		})

	test.T().Cleanup(func() {
		test.Expect(
			dp.Delete(test.Ctx(), instance.Name, metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			}),
		).ToNot(gomega.HaveOccurred())
	})

	test.Expect(err).
		ToNot(gomega.HaveOccurred())
}
