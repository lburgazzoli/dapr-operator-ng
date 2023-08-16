package predicates

import (
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func DependantWithLabels(watchUpdate bool, watchDelete bool) predicate.Predicate {
	return predicate.And(
		&HasLabel{
			Name: controller.DaprReleaseName,
		},
		&HasLabel{
			Name: controller.DaprReleaseNamespace,
		},
		&DependentPredicate{
			WatchUpdate: watchUpdate,
			WatchDelete: watchDelete,
		},
	)
}
