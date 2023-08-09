package client

import (
	route "github.com/openshift/client-go/route/clientset/versioned"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

var scaleConverter = scale.NewScaleConverter()
var codecs = serializer.NewCodecFactory(scaleConverter.Scheme())

type Client struct {
	ctrl.Client
	kubernetes.Interface

	Discovery discovery.DiscoveryInterface
	Route     route.Interface

	dynamic *dynamic.DynamicClient
	scheme  *runtime.Scheme
	config  *rest.Config
	rest    rest.Interface
	mapper  *restmapper.DeferredDiscoveryRESTMapper
}

func NewClient(cfg *rest.Config, scheme *runtime.Scheme, cc ctrl.Client) (*Client, error) {

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	restClient, err := NewRESTClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := Client{
		Client:    cc,
		Interface: kubeClient,
		Discovery: discoveryClient,
		dynamic:   dynClient,
		mapper:    restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)),
		scheme:    scheme,
		config:    cfg,
		rest:      restClient,
	}

	io, err := IsOpenShift(discoveryClient)
	if err != nil {
		return nil, err
	}

	if io {
		routeClient, err := route.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}

		c.Route = routeClient
	}

	return &c, nil
}

func NewRESTClientForConfig(config *rest.Config) (*rest.RESTClient, error) {
	cfg := rest.CopyConfig(config)
	// so that the RESTClientFor doesn't complain
	cfg.GroupVersion = &schema.GroupVersion{}
	cfg.NegotiatedSerializer = codecs.WithoutConversion()
	if len(cfg.UserAgent) == 0 {
		cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return rest.RESTClientFor(cfg)
}

// IsOpenShift returns true if we are connected to a OpenShift cluster.
func (c *Client) IsOpenShift() (bool, error) {
	if c.Discovery == nil {
		return false, nil
	}

	return IsOpenShift(c.Discovery)
}

func (c *Client) Dynamic(gvk *schema.GroupVersionKind, obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	var dr dynamic.ResourceInterface

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = c.dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		dr = c.dynamic.Resource(mapping.Resource)
	}

	return dr, nil
}

func IsOpenShift(d discovery.DiscoveryInterface) (bool, error) {
	_, err := d.ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err != nil && k8serrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
