package kubernetes

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	corecs "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// import auth plugins to make oidc auth work
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	standscs "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/clientset/versioned"
)

type (
	client struct {
		config        *rest.Config
		coreClient    corecs.Interface
		standscClient standscs.Interface
	}
	Interface interface {
		CoreClient() corecs.Interface
		StandSchedulesClient() standscs.Interface
	}
	Config struct {
		AccessType string `env-required:"true" json:"access_type" env:"KUBE_ACCESS_TYPE"`
	}
)

func NewForAccessType(accessType string) (Interface, error) {
	switch accessType {
	case "internal":
		return NewForInternalCluster()
	case "external":
		return NewForExternalCluster()
	default:
		return nil, errors.New("invalid access type specified")
	}
}

func NewForInternalCluster() (Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return NewForConfig(cfg)
}

func NewForExternalCluster() (Interface, error) {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return NewForConfig(cfg)
}

func NewForTest(env string) (Interface, error) {
	kubeconfig := os.Getenv(env)
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return NewForConfig(cfg)
}

func NewPluginClient(clientGetter genericclioptions.RESTClientGetter) (Interface, error) {
	restConfig, err := clientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	restConfig.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	if restConfig.APIPath == "" {
		restConfig.APIPath = "/api"
	}
	if restConfig.NegotiatedSerializer == nil {
		restConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	}
	if len(restConfig.UserAgent) == 0 {
		restConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	return NewForConfig(restConfig)
}

func NewForConfig(cfg *rest.Config) (*client, error) {
	kc, err := corecs.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	standsc, err := standscs.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &client{
		config:        cfg,
		coreClient:    kc,
		standscClient: standsc,
	}, nil
}

func (c *client) CoreClient() corecs.Interface {
	return c.coreClient
}

func (c *client) StandSchedulesClient() standscs.Interface {
	return c.standscClient
}
