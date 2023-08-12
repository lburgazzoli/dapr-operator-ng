package e2e

import (
	daprApi "github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	"github.com/onsi/gomega"
	"github.com/rs/xid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	ns := test.NewTestNamespace()
	dp := test.Client().Dapr.DaprV1alpha1().Daprs(ns.Name)

	instance, err := dp.Create(
		test.Ctx(),
		&daprApi.Dapr{
			ObjectMeta: metav1.ObjectMeta{
				Name:      xid.New().String(),
				Namespace: ns.Name,
			},
		},
		metav1.CreateOptions{})

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
