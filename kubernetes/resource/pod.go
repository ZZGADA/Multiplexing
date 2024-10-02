package resource

import (
	"Multiplexing_/kubernetes/enum"
	"Multiplexing_/kubernetes/resource/entity"
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

type Pod struct {
	tcp       *entity.Tcp
	clientSet *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

// PodResource 定义pod的资源情况
type PodResource struct {
	PodNum int `json:"pod_num"`
	TcpSum int `json:"tcp_sum"`
}

func NewPodInstance(namespace string, config *rest.Config, clientSet *kubernetes.Clientset) *Pod {
	return &Pod{namespace: namespace,
		config:    config,
		clientSet: clientSet,
		tcp:       entity.NewTcpInstance(clientSet, config, namespace)}
}

// Get 获取namespace下的pod的列表
func (p *Pod) Get() (*v1.PodList, error) {
	podList, err := p.clientSet.CoreV1().Pods(p.namespace).List(context.TODO(), metav1.ListOptions{})
	return podList, err
}

// GetTcpResource 获取容器的tcp连接数的平均值
func (p *Pod) GetTcpResource(port int, state enum.TcpState, command []string) (*PodResource, error) {
	p.tcp.Port = port
	p.tcp.State = state

	res := &PodResource{}
	if podList, err := p.Get(); err != nil {
		log.Fatalf("获取 %s 空间下的 pod错误,error: %#v\n", p.namespace, err)
		return nil, err
	} else {
		tcpSum := p.tcp.CountTcpTotal(podList.Items, command, enum.Pod)
		res.TcpSum = tcpSum
		res.PodNum = len(podList.Items)
	}
	return res, nil
}
