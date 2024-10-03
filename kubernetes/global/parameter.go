package global

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	Namespace  = "backend"
	ConfigPath = "/Users/tal/.kube/config"
)

var (
	KubeConfig    *string
	Config        *restclient.Config
	IsCreating    bool
	DynamicClient *dynamic.DynamicClient
	ClientSet     *kubernetes.Clientset
	DeploymentRes *schema.GroupVersionResource
	command       = []string{"sh", "-c", " netstat -nat | grep -i '18081' | wc -l"}
)
