package module

import (
	"Multiplexing_/kubernetes/enum"
	_interface "Multiplexing_/kubernetes/interface"
	"Multiplexing_/kubernetes/resource"
	"Multiplexing_/kubernetes/template"
	"bytes"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"strconv"
	"strings"
)

var ScalingDeployments []string

type Pod struct {
	clientSet     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	config        *rest.Config
	namespace     string
}

func NewPodInstance(namespace string, config *rest.Config, clientSet *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient) *Pod {
	return &Pod{
		namespace:     namespace,
		config:        config,
		clientSet:     clientSet,
		dynamicClient: dynamicClient,
	}
}

// Get 获取namespace下的pod的列表
// 实现ModuleResource
func (p *Pod) Get() (metav1.ListInterface, error) {
	podList, err := p.clientSet.CoreV1().Pods(p.namespace).List(context.TODO(), metav1.ListOptions{})
	return podList, err
}

// GetTcpResource 获取容器的tcp连接数的平均值
func (p *Pod) GetTcpResource(port int, state enum.TcpState, commandStr string) (*resource.Tcp, error) {
	command := []string{"sh", "-c", fmt.Sprintf(commandStr, state, port)}

	res := resource.NewTcpInstance(state, port)
	if podList, err := p.Get(); err != nil {
		log.Fatalf("获取 %s 空间下的 pod错误,error: %#v\n", p.namespace, err)
		return nil, err
	} else {
		pods := podList.(*v1.PodList)
		tcpSum := p.CountTcpTotal(pods.Items, command, enum.Pod)
		res.TcpConnect.TcpNum = tcpSum
		res.TcpConnect.UnitNum = len(pods.Items)
	}
	return res, nil
}

// CountTcpTotal 计算kubernetes组件的tcp连接数量
// 实例TcpResource接口 需要给传入参数进行断言 当前是pod组件就断言为pod list
func (p *Pod) CountTcpTotal(list interface{}, command []string, resource enum.Resource) int {
	pods := list.([]v1.Pod)
	total := 0

	for _, item := range pods {
		req := p.clientSet.CoreV1().RESTClient().Post().
			Resource(resource.String()).
			Name(item.Name).
			Namespace(p.namespace).
			SubResource(enum.Exec.String())

		req.VersionedParams(
			&v1.PodExecOptions{
				Command: command,
				Stdin:   false,
				Stdout:  true,
				Stderr:  true,
				TTY:     false,
			},
			scheme.ParameterCodec,
		)

		log.Println(item.Name)

		exec, errExec := remotecommand.NewSPDYExecutor(p.config, enum.POST.String(), req.URL())
		if errExec != nil {
			log.Panicf("%#v\n", errExec.Error())
		}

		// 从输出流中读取配置
		var stdout, stderr bytes.Buffer
		if err := exec.Stream(remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: &stdout,
			Stderr: &stderr,
		}); err != nil {
			log.Panicf("从pod中读取输出错误,%#v,%#v\n", err, stderr.String())
		}

		log.Println(stdout.String())

		if tcpNum, err := strconv.Atoi(strings.TrimSpace(stdout.String())); err != nil {
			log.Panicf("kubectl 返回结果不为整形,%#v\n", err)
		} else {
			total += tcpNum
		}
	}
	return total
}

// ExtendResource 动态扩容
func (p *Pod) ExtendResource(resourceYaml _interface.Template, recordExtendResource func(resourceName string), forbiddenExtend func()) bool {
	// 在pod资源下面的实例 断言为deployment资源
	deployment := resourceYaml.(*template.Deployment)
	deploymentYaml := deployment.DeploymentYaml

	groupResource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: enum.Deployment.String()}
	targetStructure := &unstructured.Unstructured{Object: deploymentYaml}

	fmt.Printf("Deployment is creating ~ ")
	result, err := p.dynamicClient.Resource(groupResource).Namespace(p.namespace).Create(context.TODO(), targetStructure, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("%#v", err)
		return false
	}
	fmt.Printf("Deployment Created %q \n", result.GetName())

	recordExtendResource(deployment.Name)
	go forbiddenExtend()
	return true
}

func (p *Pod) RecallResource(resourceName string) bool {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	groupResource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: enum.Deployment.String()}

	fmt.Println("Deleting deployment...")

	if err := p.dynamicClient.Resource(groupResource).Namespace(p.namespace).Delete(context.TODO(), resourceName, deleteOptions); err != nil {
		panic(err)
	}

	// 清空切
	fmt.Println("Deleted deployment.")
	return true
}
