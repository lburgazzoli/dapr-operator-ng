package e2e

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	"github.com/rs/xid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
	. "github.com/onsi/gomega"

	daprApi "github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1"
	daprAc "github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/applyconfiguration/tools/v1alpha1"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	ns := test.NewTestNamespace()
	dp := test.Client().Dapr.ToolsV1alpha1().Daprs(ns.Name)

	instance, err := dp.Apply(
		test.Ctx(),
		daprAc.Dapr(xid.New().String(), ns.Name).
			// This should not be needed but for some reasons, the
			// applyconfiguration-gen tool sets a wrong APIVersion
			// for the Dapr type (tools/v1alpha1 instead of the one
			// with the domain tools.dapr.io/v1alpha1).
			//
			// TODO: figure out why
			WithAPIVersion(daprApi.GroupVersion.String()).
			WithSpec(daprAc.DaprSpec().
				WithValues(nil),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-test",
		})

	test.Expect(err).
		ToNot(HaveOccurred())

	test.T().Cleanup(func() {
		test.Expect(
			dp.Delete(test.Ctx(), instance.Name, metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			}),
		).ToNot(HaveOccurred())
	})

	test.Eventually(Deployment(test, instance, "dapr-operator"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, instance, "dapr-sentry"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, instance, "dapr-sidecar-injector"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

}
