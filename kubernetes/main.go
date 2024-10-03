package main

import (
	"Multiplexing_/kubernetes/enum"
	"Multiplexing_/kubernetes/global"
	_interface "Multiplexing_/kubernetes/interface"
	"Multiplexing_/kubernetes/module"
	"Multiplexing_/kubernetes/strategies"
	"Multiplexing_/kubernetes/template"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

var (
	strategy _interface.Strategy
)

// 初始化
func init() {
	var err error

	// homedir.HomeDir()获得home路径
	// 请输入你的.kube/config的绝对路径 或者go run main.go -kubeconfig=
	if home := homedir.HomeDir(); home != "" {
		global.KubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), global.ConfigPath)
	} else {
		global.KubeConfig = flag.String("kubeconfig", "", global.ConfigPath)
	}
	// 解析命令行标志
	flag.Parse()

	// 也可以输入kubectl cluster-info获取主结点的信息
	// 以此获取核心配置、restful客户端、资源
	if global.Config, err = clientcmd.BuildConfigFromFlags("", *global.KubeConfig); err != nil {
		panic(err)
	}
	if global.DynamicClient, err = dynamic.NewForConfig(global.Config); err != nil {
		panic(err)
	}
	if global.ClientSet, err = kubernetes.NewForConfig(global.Config); err != nil {
		panic(err)
	}

	global.DeploymentRes = &schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: enum.Deployment.String()}
	global.DynamicClient, _ = dynamic.NewForConfig(global.Config)
	global.IsCreating = false
	strategy = &strategies.TCPConnectStrategy{}
}

func main() {
	// 后端测试服务的端口
	// 请保证镜像中已经安装了net-tools，如果没有请先安装net-tools
	// -i 表示忽略大小写
	port := 18081
	state := enum.ESTABLISHED
	command := "netstat -nt | grep -i '%s' | grep '%d' | wc -l"
	passTime := 0

	for {
		// TODO：追加 3ơ 原则 用于删除deployment 具体实现有点难写 晚点在写
		podResource := module.NewPodInstance(global.Namespace, global.Config, global.ClientSet, global.DynamicClient)
		podTcpStatus, _ := podResource.GetTcpResource(port, state, command)

		fmt.Printf("%#v\n", podTcpStatus)
		ifNeedExtend := strategy.CheckIfNeedDynamicExtend(podTcpStatus)

		if ifNeedExtend {
			// 判断是否需要扩容
			// 具体资源扩容逻辑和扩容资源的记录分成两个部分 互不影响 使用回调函数
			// 最后清0 重新开始计算
			// 同步创建 可以控制pod的启动状态
			log.Println("准备水平扩容deployment～")
			deployment := template.NewDeploymentInstance(selfDeployment())
			if ifExtend := podResource.ExtendResource(deployment, strategy.ExpandResource, strategy.CountingResourceExtendTime); !ifExtend {
				log.Println("pod扩容失败")
			}

			passTime = 0
		}

		time.Sleep(time.Duration(strategies.TimeSet) * time.Second)
		passTime += strategies.TimeSet

		if passTime%strategies.TimeToRecallResource == 0 {
			// 间隔一段时间后判断是否删除
			log.Println("开始判断是否需要回收资源")
			ifNeedRecall := strategy.CheckIfNeedRecallDeployment()
			if ifNeedRecall {
				log.Println("准备回收pod")
				resource := strategy.RecallResource()
				// 不能异步删除 如果异步 会出现 同时读pod 但是pod删掉了 找不到了 除非读的时候 将异常过滤掉
				if recallResource := podResource.RecallResource(resource); recallResource {
					log.Println("pod回收成功")
				}
			}
		}
	}

}

// selfDeployment 用户自定义的模版 这里做个示例
func selfDeployment() map[string]interface{} {
	deploymentName := "mq-utility-bill-deployment-backup" + strconv.Itoa(time.Now().Nanosecond())
	namespace := "backend"

	object := map[string]interface{}{
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
	}
	return object
}
