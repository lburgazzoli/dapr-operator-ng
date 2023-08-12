package predicates

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = DeletedPredicate{}

type DeletedPredicate struct {
	predicate.Funcs
}

func (DeletedPredicate) Create(event.CreateEvent) bool {
	return false
}
func (DeletedPredicate) Update(event.UpdateEvent) bool {
	return false
}
func (DeletedPredicate) Generic(event.GenericEvent) bool {
	return false
}

func (DeletedPredicate) Delete(e event.DeleteEvent) bool {
	o, ok := e.Object.(*unstructured.Unstructured)
	if !ok {
		log.Error(nil, "Unexpected object type", "gvk", e.Object.GetObjectKind().GroupVersionKind().String())
		return false
	}

	log.Info("Reconciling due to dependent resource deletion",
		"name", o.GetName(),
		"namespace", o.GetNamespace(),
		"apiVersion", o.GroupVersionKind().GroupVersion(),
		"kind", o.GroupVersionKind().Kind)

	return true
}
