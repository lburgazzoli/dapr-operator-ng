package support

import (
	"errors"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	daprClient "github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/clientset/versioned"
)

type Client struct {
	kubernetes.Interface

	Dapr      daprClient.Interface
	Discovery discovery.DiscoveryInterface

	//nolint:unused
	scheme *runtime.Scheme
	config *rest.Config
}

func newClient() (*Client, error) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		home := homedir.HomeDir()
		if home != "" {
			kc = filepath.Join(home, ".kube", "config")
		}
	}

	if kc == "" {
		return nil, errors.New("unable to determine KUBECONFIG")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dClient, err := daprClient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := Client{
		Interface: kubeClient,
		Discovery: discoveryClient,
		Dapr:      dClient,
		config:    cfg,
	}

	return &c, nil
}
