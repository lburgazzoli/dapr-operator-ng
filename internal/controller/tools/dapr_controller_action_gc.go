package tools

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/lburgazzoli/dapr-operator-ng/pkg/resources"

	"github.com/go-logr/logr"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller"
	"github.com/lburgazzoli/dapr-operator-ng/pkg/controller/client"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	authorization "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGCAction() Action {
	return &GCAction{
		limiter:         rate.NewLimiter(rate.Every(time.Minute), 1),
		collectableGVKs: make(map[schema.GroupVersionKind]struct{}),
		l:               ctrl.Log.WithName("action").WithName("gc"),
	}
}

type GCAction struct {
	l               logr.Logger
	lock            sync.Mutex
	limiter         *rate.Limiter
	collectableGVKs map[schema.GroupVersionKind]struct{}
}

func (a *GCAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *GCAction) Run(ctx context.Context, rc *ReconciliationRequest) error {

	return a.gc(ctx, rc)
}

func (a *GCAction) Cleanup(ctx context.Context, rc *ReconciliationRequest) error {
	return a.gc(ctx, rc)
}

func (a *GCAction) gc(ctx context.Context, rc *ReconciliationRequest) error {
	a.l.Info("run")

	deletableGVKs, err := a.getDeletableTypes(ctx, rc)
	if err != nil {
		return fmt.Errorf("cannot discover GVK types: %w", err)
	}

	namespace, err := labels.NewRequirement(
		controller.DaprResourceNamespace,
		selection.Equals,
		[]string{rc.Resource.Namespace})

	if err != nil {
		return errors.Wrap(err, "cannot determine ref requirement")
	}

	name, err := labels.NewRequirement(
		controller.DaprResourceName,
		selection.Equals,
		[]string{rc.Resource.Name})

	if err != nil {
		return errors.Wrap(err, "cannot determine ref requirement")
	}

	generation, err := labels.NewRequirement(
		controller.DaprResourceGeneration,
		selection.LessThan,
		[]string{strconv.FormatInt(rc.Resource.Status.ObservedGeneration, 10)})

	if err != nil {
		return errors.Wrap(err, "cannot determine generation requirement")
	}

	selector := labels.NewSelector().
		Add(*namespace).
		Add(*name).
		Add(*generation)

	return a.deleteEachOf(ctx, rc, deletableGVKs, selector)
}

func (a *GCAction) deleteEachOf(ctx context.Context, rc *ReconciliationRequest, deletableGVKs map[schema.GroupVersionKind]struct{}, selector labels.Selector) error {
	for GVK := range deletableGVKs {
		items := unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"apiVersion": GVK.GroupVersion().String(),
				"kind":       GVK.Kind,
			},
		}
		options := []ctrlCli.ListOption{
			ctrlCli.MatchingLabelsSelector{Selector: selector},
		}

		if err := rc.Client.List(ctx, &items, options...); err != nil {
			if !k8serrors.IsNotFound(err) {
				return errors.Wrap(err, "cannot list child resources")
			}
			continue
		}

		for i := range items.Items {
			resource := items.Items[i]

			if !a.canBeDeleted(ctx, rc, resource) {
				continue
			}

			a.l.Info("deleting", "ref", resources.Ref(&resource))

			err := rc.Client.Delete(ctx, &resource, ctrlCli.PropagationPolicy(metav1.DeletePropagationForeground))
			if err != nil {
				// The resource may have already been deleted
				if !k8serrors.IsNotFound(err) {
					continue
				}

				return errors.Wrapf(
					err,
					"cannot delete resources gvk:%s, namespace: %s, name: %s",
					resource.GroupVersionKind().String(),
					resource.GetNamespace(),
					resource.GetName())
			}

			a.l.Info("deleted", "ref", resources.Ref(&resource))
		}
	}

	return nil
}

func (a *GCAction) canBeDeleted(_ context.Context, _ *ReconciliationRequest, _ unstructured.Unstructured) bool {
	return true
}

func (a *GCAction) getDeletableTypes(ctx context.Context, rc *ReconciliationRequest) (map[schema.GroupVersionKind]struct{}, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	// Rate limit to avoid Discovery and SelfSubjectRulesReview requests at every reconciliation.
	if !a.limiter.Allow() {
		// Return the cached set of garbage collectable GVKs.
		return a.collectableGVKs, nil
	}

	// We rely on the discovery API to retrieve all the resources GVK,
	// that results in an unbounded set that can impact garbage collection latency when scaling up.
	items, err := rc.Client.Discovery.ServerPreferredNamespacedResources()

	// Swallow group discovery errors, e.g., Knative serving exposes
	// an aggregated API for custom.metrics.k8s.io that requires special
	// authentication scheme while discovering preferred resources.
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, err
	}

	// We only take types that support the "delete" verb,
	// to prevents from performing queries that we know are going to return "MethodNotAllowed".
	apiResourceLists := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"delete"}}, items)

	// Retrieve the permissions granted to the operator service account.
	// We assume the operator has only to garbage collect the resources it has created.
	ssrr := &authorization.SelfSubjectRulesReview{
		Spec: authorization.SelfSubjectRulesReviewSpec{
			Namespace: rc.Namespace,
		},
	}
	ssrr, err = rc.Client.AuthorizationV1().SelfSubjectRulesReviews().Create(ctx, ssrr, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	GVKs := make(map[schema.GroupVersionKind]struct{})
	for _, res := range apiResourceLists {
		for i := range res.APIResources {
			resourceGroup := res.APIResources[i].Group
			if resourceGroup == "" {
				// Empty implies the group of the containing resource list should be used
				gv, err := schema.ParseGroupVersion(res.GroupVersion)
				if err != nil {
					return nil, err
				}
				resourceGroup = gv.Group
			}
		rule:
			for _, rule := range ssrr.Status.ResourceRules {
				if !slices.Contains(rule.Verbs, "delete") && !slices.Contains(rule.Verbs, "*") {
					continue
				}

				for _, ruleGroup := range rule.APIGroups {
					for _, ruleResource := range rule.Resources {
						if (resourceGroup == ruleGroup || ruleGroup == "*") && (res.APIResources[i].Name == ruleResource || ruleResource == "*") {
							GVK := schema.FromAPIVersionAndKind(res.GroupVersion, res.APIResources[i].Kind)
							GVKs[GVK] = struct{}{}
							break rule
						}
					}
				}
			}
		}
	}

	a.collectableGVKs = GVKs

	return a.collectableGVKs, nil
}
