package code

import (
	"Multiplexing_/kubernetes/enum"
	"Multiplexing_/kubernetes/module"
	"bytes"
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"strconv"
	"strings"
	"time"

	"Multiplexing_/kubernetes/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

const configPath = "/Users/tal/.kube/config"
const pods = "pods"
const threshold = 5
const namespace = "backend"

var isCreating = false
var podDynamicScaling = make([]string, 0)

func DynamicStringPod() {
	// 配置Kubernetes客户端
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Fatal(err)
	}
	// 创建Kubernetes核心客户端
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	// -n：显示数字地址和端口号，而不是尝试解析主机名和服务名。这可以加快输出速度，因为不需要进行 DNS 查询。
	// -a：显示所有的网络连接，包括监听（listening）和非监听（non-listening）套接字。
	// -t：显示 TCP 连接。
	// -i: 忽略消协
	command := " netstat -nt | grep -i '18081' | wc -l"

	for {
		// TODO：追加计时器 阈值超过一定时间才动态生成， 追加 3ơ 原则
		podResource := module.NewPodInstance(namespace, config, clientSet)
		podTcpStatus, _ := podResource.GetTcpResource(18081, enum.ESTABLISHED, command)

		if judgeThreadHold(&podTcpStatus.TcpConnect) {
			fmt.Println("负载过高 需用动态扩容 ，准备更新容器")
			isCreating = true
			go startingBackUpDeployment(config)
		} else {
			fmt.Println("资源分配充裕 可以抗下压力")
			go deleteBackUpDeployment(config)
		}
		time.Sleep(5 * time.Second)
	}
}

func judgeThreadHold(podTcpStatus *resource.TcpConnectResource) bool {
	// 计算每个pod的平均tcp连接数
	// 如果超出阈值 同时没有新的容器正在创建的话 那么就要新建容器
	podNum := podTcpStatus.UnitNum
	tcpSum := podTcpStatus.TcpNum
	meanTcpNumEachPod := tcpSum / podNum

	return meanTcpNumEachPod >= threshold && !isCreating
}

// 校验 tcp的连接状态
func checkPodTcpStatus(clientSet *kubernetes.Clientset, config *rest.Config) resource.TcpConnectResource {
	/**
	-
		sh -c 用于执行更加复杂的命令行字符串
		command := []string{"sh", "-c", "ss -t -p | grep 'pid=1' | wc -l"}
		netstat -t -p | grep 18081 | wc -l
	    获取 18081 端口tcp连接数
		-n 显示数字地址而不是尝试解析主机名。
		-t tcp
		-a 所有连接
	*/
	command := []string{"sh", "-c", " netstat -nat | grep -i '18081' | wc -l"}
	podList, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	tcpSum := 0

	for _, pod := range podList.Items {
		req := clientSet.CoreV1().RESTClient().Post().
			Resource(pods).
			Name(pod.Name).
			Namespace(namespace).
			SubResource("exec")

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

		fmt.Println(pod.Name)

		exec, errExec := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if errExec != nil {
			panic(errExec.Error())
		}

		// 从输出流中读取配置
		var stdout, stderr bytes.Buffer
		if err := exec.Stream(remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: &stdout,
			Stderr: &stderr,
		}); err != nil {
			log.Panicf("从pod中读取输出错误,%#v,%#v", err, stderr.String())
		}

		fmt.Println(stdout.String())
		tcpNum, _ := strconv.Atoi(strings.TrimSpace(stdout.String()))
		tcpSum += tcpNum
	}

	return resource.TcpConnectResource{UnitNum: len(podList.Items), TcpNum: tcpSum}
}

// deleteBackUpDeployment 删除创建的负载pod
func deleteBackUpDeployment(config *rest.Config) {
	client, _ := dynamic.NewForConfig(config)
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	fmt.Println("Deleting deployment...")
	for _, deploymentNameNeedToDelete := range podDynamicScaling {
		if err := client.Resource(deploymentRes).Namespace(namespace).Delete(context.TODO(), deploymentNameNeedToDelete, deleteOptions); err != nil {
			panic(err)
		}
	}
	// 清空切片
	podDynamicScaling = make([]string, 0)

	fmt.Println("Deleted deployment.")
}

func startingBackUpDeployment(config *rest.Config) {
	defer func() {
		isCreating = false
	}()
	/**
	-
		schema.GroupVersionResource：这是一个结构体，用于表示 Kubernetes 资源的组、版本和资源类型。
		Group: "apps"：表示资源所属的组是 apps。
		Version: "v1"：表示资源的版本是 v1。
		Resource: "deployments"：表示资源的类型是 deployments。
	*/
	client, _ := dynamic.NewForConfig(config)
	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deploymentName := "mq-utility-bill-deployment-backup" + strconv.Itoa(time.Now().Nanosecond())

	// deployment 的启动对象
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      deploymentName,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app":     "mq-utility-bill",
					"version": "v1",
				},
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "mq-utility-bill",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app":     "mq-utility-bill",
							"version": "v1",
						},
					},

					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "mq-utility-bill",
								"image": "my-utility-bill:1.0.1",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 18081,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")

	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Printf("%#v", err)
		panic(err)
	}

	podDynamicScaling = append(podDynamicScaling, deploymentName)
	fmt.Printf("Created deployment %q.\n", result.GetName())

}
