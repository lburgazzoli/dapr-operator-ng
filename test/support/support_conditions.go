package support

import (
	"github.com/lburgazzoli/dapr-operator-ng/api/operator/v1alpha1"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/conditions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type conditionType interface {
	~string
}

func ConditionStatus[T conditionType](conditionType T) func(any) corev1.ConditionStatus {
	return func(object any) corev1.ConditionStatus {
		switch o := object.(type) {
		case conditions.Getter:
			if c := conditions.Get(o, conditions.ConditionType(conditionType)); c != nil {
				return corev1.ConditionStatus(c.Status)
			}
		case *appsv1.Deployment:
			if o != nil {
				for i := range o.Status.Conditions {
					if string(o.Status.Conditions[i].Type) == string(conditionType) {
						return o.Status.Conditions[i].Status
					}
				}
			}
		case *v1alpha1.DaprControlPlane:
			if o != nil {
				for i := range o.Status.Conditions {
					if o.Status.Conditions[i].Type == string(conditionType) {
						return corev1.ConditionStatus(o.Status.Conditions[i].Status)
					}
				}
			}
		}

		return corev1.ConditionUnknown
	}
}

func ConditionReason[T conditionType](conditionType T) func(any) string {
	return func(object any) string {
		switch o := object.(type) {
		case conditions.Getter:
			if c := conditions.Get(o, conditions.ConditionType(conditionType)); c != nil {
				return c.Reason
			}
		case *appsv1.Deployment:
			if o != nil {
				for i := range o.Status.Conditions {
					if string(o.Status.Conditions[i].Type) == string(conditionType) {
						return o.Status.Conditions[i].Reason
					}
				}
			}
		case *v1alpha1.DaprControlPlane:
			if o != nil {
				for i := range o.Status.Conditions {
					if o.Status.Conditions[i].Type == string(conditionType) {
						return o.Status.Conditions[i].Reason
					}
				}
			}
		}

		return ""
	}
}
