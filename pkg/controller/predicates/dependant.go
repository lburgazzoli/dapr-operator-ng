package predicates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = DependentPredicate{}

type DependentPredicate struct {
	predicate.Funcs
}

func (DependentPredicate) Delete(e event.DeleteEvent) bool {
	o := e.Object.(*unstructured.Unstructured)

	log.Info("Reconciling due to dependent resource deletion",
		"name", o.GetName(),
		"namespace", o.GetNamespace(),
		"apiVersion", o.GroupVersionKind().GroupVersion(),
		"kind", o.GroupVersionKind().Kind)

	return true
}

func (DependentPredicate) Update(e event.UpdateEvent) bool {
	oldObj := e.ObjectOld.(*unstructured.Unstructured).DeepCopy()
	newObj := e.ObjectNew.(*unstructured.Unstructured).DeepCopy()

	if oldObj.GetResourceVersion() == newObj.GetResourceVersion() {
		return false
	}

	// Update filters out events that change only the dependent resource
	// status. It is not typical for the controller of a primary
	// resource to write to the status of one its dependent resources.
	delete(oldObj.Object, "status")
	delete(newObj.Object, "status")

	oldObj.SetResourceVersion("")
	newObj.SetResourceVersion("")

	oldObj.SetManagedFields(removeTimeFromManagedFields(oldObj.GetManagedFields()))
	newObj.SetManagedFields(removeTimeFromManagedFields(newObj.GetManagedFields()))

	if reflect.DeepEqual(oldObj.Object, newObj.Object) {
		return false
	}

	log.Info("Reconciling due to dependent resource update",
		"name", newObj.GetName(),
		"namespace", newObj.GetNamespace(),
		"apiVersion", newObj.GroupVersionKind().GroupVersion(),
		"kind", newObj.GroupVersionKind().Kind)

	return true
}

func removeTimeFromManagedFields(fields []metav1.ManagedFieldsEntry) []metav1.ManagedFieldsEntry {
	if fields == nil {
		return nil
	}

	newFields := make([]metav1.ManagedFieldsEntry, 0)
	for _, field := range fields {
		field.Time = nil
		newFields = append(newFields, field)
	}

	return newFields
}
