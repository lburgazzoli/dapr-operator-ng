package e2e

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	daprCP "github.com/lburgazzoli/dapr-operator-ng/internal/controller/operator"
	"github.com/rs/xid"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
	. "github.com/onsi/gomega"

	daprAc "github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/applyconfiguration/operator/v1alpha1"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)

	instance := test.NewDaprControlPlane(daprAc.DaprControlPlaneSpec().
		WithValues(nil),
	)

	test.Eventually(Deployment(test, instance, "dapr-operator"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, instance, "dapr-sentry"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, instance, "dapr-sidecar-injector"), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

}

func TestDaprDeployWrongCR(t *testing.T) {
	test := With(t)

	instance := test.NewNamespacedNameDaprControlPlane(
		types.NamespacedName{
			Name:      xid.New().String(),
			Namespace: daprCP.DaprControlPlaneNamespace,
		},
		daprAc.DaprControlPlaneSpec().
			WithValues(nil),
	)

	test.Eventually(ControlPlane(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(daprCP.DaprConditionReconcile), Equal(corev1.ConditionFalse)))
	test.Eventually(ControlPlane(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionReason(daprCP.DaprConditionReconcile), Equal(daprCP.DaprConditionReasonUnsupportedConfiguration)))

}
