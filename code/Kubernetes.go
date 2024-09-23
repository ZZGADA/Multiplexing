package code

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

const namespace = "backend"

func Kubernetes() {
	// 配置Kubernetes客户端
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/tal/.kube/config")
	if err != nil {
		log.Fatal(err)
	}
	// 创建Kubernetes核心客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	getPod(clientset)
	getSvc(clientset)
	getIngress(clientset)

}

/**
 * 通过CoreV1 与kubernetes进行操作  .Pod .Service
 */

func getPod(clientSet *kubernetes.Clientset) {
	// 获取Pod列表
	// CoreV1 返回一个可以操作Kubernetes核心V1版本API资源的接口。在这个例子中，它返回的是一个可以操作Pod资源的接口。
	// List 这是一个方法，用于列出Pod资源。
	// context 这是一个上下文参数，用于控制请求的生命周期。
	// 这是一个选项参数，用于指定列出资源时的过滤条件。在这个例子中，使用了空的ListOptions，表示不进行任何过滤，获取所有的Pod。
	pods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// 解析并处理获取到的Pod信息
	fmt.Println("Pod列表：")
	for _, pod := range pods.Items {
		fmt.Printf("名称：%s，状态：%s，创建时间：%s\n", pod.Name, pod.Status.Phase, pod.CreationTimestamp)
	}
}

func getSvc(clientSet *kubernetes.Clientset) {
	// 获取svc列表
	list, err := clientSet.CoreV1().Services(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("error,%#v", err)
	}

	fmt.Println("Service列表 & 端点信息：")
	for _, svc := range list.Items {
		fmt.Printf("名称：%s，状态：%s，创建时间：%s\n", svc.Name, svc.Status.String(), svc.CreationTimestamp)

		// 获取服务的端点信息
		endpoints, err := clientSet.CoreV1().Endpoints(namespace).Get(context.TODO(), "my-service-cluster", v1.GetOptions{})
		if err != nil {
			log.Fatalf("Error getting endpoints: %s", err.Error())
		}

		// 打印端点信息
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				podName := address.TargetRef.Name
				pod, err := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
				if err != nil {
					log.Printf("Error getting pod %s: %s", podName, err.Error())
					continue
				}

				// 获取Pod的运行状态
				podStatus := pod.Status.Phase

				for _, port := range subset.Ports {
					fmt.Printf("Pod Name: %s, IP: %s, Port: %d, Status: %s\n", podName, address.IP, port.Port, podStatus)
				}
			}
		}

	}
}

func getIngress(client *kubernetes.Clientset) {
	ingresses, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return
	}

	fmt.Println("Ingress")
	for _, ingress := range ingresses.Items {
		fmt.Printf("Name: %s\n", ingress.Name)
		fmt.Printf("Namespace: %s\n", ingress.Namespace)
		fmt.Printf("Annotations: %v\n", ingress.Annotations)
		fmt.Printf("Labels: %v\n", ingress.Labels)
		fmt.Printf("Rules:\n")
		for _, rule := range ingress.Spec.Rules {
			fmt.Printf("  Host: %s\n", rule.Host)
			for _, path := range rule.HTTP.Paths {
				fmt.Printf("    Path: %s\n", path.Path)
				fmt.Printf("    Backend: %s:%d\n", path.Backend.Service.Name, path.Backend.Service.Port.Number)
			}
		}
		fmt.Println()
	}

}
