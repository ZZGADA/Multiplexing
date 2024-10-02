package entity

import (
	"Multiplexing_/kubernetes/enum"
	"bytes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"strconv"
	"strings"
)

type TcpResource interface {
	CountTcpTotal(list interface{}, command []string, resource enum.Resource) int
}

// Tcp tcp连接结构体
type Tcp struct {
	Port  int
	State enum.TcpState

	clientSet *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

// NewTcpInstance 初始化实例对象
// 有三个默认值
func NewTcpInstance(clientSet *kubernetes.Clientset, config *rest.Config, namespace string) *Tcp {
	return &Tcp{
		clientSet: clientSet,
		config:    config,
		namespace: namespace,
		State:     enum.ESTABLISHED,
		Port:      8080,
	}
}

// CountTcpTotal 计算kubernetes组件的tcp连接数量
// 实例TcpResource接口 需要给传入参数进行断言
func (tcp *Tcp) CountTcpTotal(list interface{}, command []string, resource enum.Resource) int {
	switch resource {
	case enum.Pod:
		pods := list.([]v1.Pod)
		return tcp.getTcpTotalFromPod(pods, command, resource)
	default:
		return 0
	}
}

func (tcp *Tcp) getTcpTotalFromPod(list []v1.Pod, command []string, resource enum.Resource) int {
	total := 0

	for _, item := range list {
		req := tcp.clientSet.CoreV1().RESTClient().Post().
			Resource(resource.String()).
			Name(item.Name).
			Namespace(tcp.namespace).
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

		exec, errExec := remotecommand.NewSPDYExecutor(tcp.config, enum.POST.String(), req.URL())
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
			log.Fatalf("kubectl 返回结果不为整形,%#v\n", err)
		} else {
			total += tcpNum
		}
	}
	return total
}
