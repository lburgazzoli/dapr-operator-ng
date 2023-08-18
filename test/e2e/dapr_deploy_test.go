package e2e

import (
	"testing"

	daprCP "github.com/lburgazzoli/dapr-operator-ng/internal/controller/operator"
	"github.com/rs/xid"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
	. "github.com/onsi/gomega"

	daprAc "github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/applyconfiguration/operator/v1alpha1"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)

	cp := test.Client().DaprCP

	instance, err := cp.Apply(
		test.Ctx(),
		daprAc.DaprControlPlane(daprCP.DaprControlPlaneName, daprCP.DaprControlPlaneNamespace).
			WithSpec(daprAc.DaprControlPlaneSpec().
				WithValues(nil),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.Name(),
		})

	test.Expect(err).
		ToNot(HaveOccurred())

	test.T().Cleanup(func() {
		test.Expect(
			cp.Delete(test.Ctx(), instance.Name, metav1.DeleteOptions{
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

func TestDaprDeployWrongCR(t *testing.T) {
	test := With(t)

	cp := test.Client().DaprCP

	instance, err := cp.Apply(
		test.Ctx(),
		daprAc.DaprControlPlane(xid.New().String(), daprCP.DaprControlPlaneNamespace).
			WithSpec(daprAc.DaprControlPlaneSpec().
				WithValues(nil),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.Name(),
		})

	test.Expect(err).
		ToNot(HaveOccurred())

	test.T().Cleanup(func() {
		test.Expect(
			cp.Delete(test.Ctx(), instance.Name, metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			}),
		).ToNot(HaveOccurred())
	})

	test.Eventually(ControlPlane(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(daprCP.DaprConditionReconcile), Equal(corev1.ConditionFalse)))

}
