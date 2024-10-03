package template

import (
	"log"
)

type Deployment struct {
	DeploymentYaml map[string]interface{}
	Name           string
}

func NewDeploymentInstance(selfDefineResource map[string]interface{}) *Deployment {
	deployment := &Deployment{}
	deployment.CreateResourceYaml(selfDefineResource)
	return deployment
}

// lazyLoadDeployment 懒加载返回默认值和需要的key
func (deployment *Deployment) lazyLoadDeployment() map[string]interface{} {
	return map[string]interface{}{
		APIVersion: nil,
		Kind:       nil,
		Metadata: map[string]interface{}{
			Name:      nil,
			Namespace: nil,
			Labels: map[string]interface{}{
				App:     nil,
				Version: nil,
			},
		},
		Spec: map[string]interface{}{
			Replicas: nil,
			Selector: map[string]interface{}{
				MatchLabels: map[string]interface{}{
					App: nil,
				},
			},
			Template: map[string]interface{}{
				Metadata: map[string]interface{}{
					Labels: map[string]interface{}{
						App:     nil,
						Version: nil,
					},
				},

				Spec: map[string]interface{}{
					Containers: []map[string]interface{}{
						{
							Name:  nil,
							Image: nil,
							Ports: []map[string]interface{}{
								{
									Name:          nil,
									Protocol:      nil,
									ContainerPort: nil,
								},
							},
						},
					},
				},
			},
		},
	}
}

// CreateResourceYaml 自定义的Deployment模版 创建kubectl api需要的yaml格式文件
func (deployment *Deployment) CreateResourceYaml(selfResourceYaml map[string]interface{}) {
	deployment.DeploymentYaml = deployment.lazyLoadDeployment()
	deployment.setMapIntoDeploymentTemplate(selfResourceYaml, deployment.DeploymentYaml)
	deployment.Name = deployment.DeploymentYaml[Metadata].(map[string]interface{})[Name].(string)
}

func (deployment *Deployment) setMapIntoDeploymentTemplate(selfResourceYaml map[string]interface{}, deploymentTemplate map[string]interface{}) {
	var valueSelfDefine interface{}
	var exist bool
	var ok bool
	for key, value := range deploymentTemplate {
		if valueSelfDefine, exist = selfResourceYaml[key]; !exist {
			log.Panicf("资源类型不匹配,key不存在，必须输入key：%s", key)
		}

		// deploymentTemplate 的value为空的话 那么deployment 传入的value就为一个string
		if value == nil {
			deploymentTemplate[key] = valueSelfDefine
			continue
		}

		// // 如果无法断言为map[string]interface{} 表示可以断言为 []map[string]interface{}
		if _, ok = value.(map[string]interface{}); !ok {
			deployment.setSliceIntoDeploymentTemplate(selfResourceYaml[key].([]map[string]interface{}), value.([]map[string]interface{}))
		} else {
			deployment.setMapIntoDeploymentTemplate(selfResourceYaml[key].(map[string]interface{}), value.(map[string]interface{}))
		}
	}
}

// setSliceIntoDeploymentTemplate 递归遍历slice
// 按照k8s的设计理念 一个pod就一个容器  但是容器开放的port可能是多个 索性就port和container都允许为多个
func (deployment *Deployment) setSliceIntoDeploymentTemplate(selfResourceYaml []map[string]interface{}, deploymentTemplate []map[string]interface{}) {
	for i := 0; i < len(selfResourceYaml); i++ {
		if i == 0 {
			deploymentTemplate[i][Name] = selfResourceYaml[i][Name]
			deploymentTemplate[i][Image] = selfResourceYaml[i][Image]

			portsTemplates := deploymentTemplate[i][Ports].([]map[string]interface{})
			ports := selfResourceYaml[i][Ports].([]map[string]interface{})

			for j := 0; j < len(portsTemplates); j++ {
				if j == 0 {
					portsTemplates[j][Name] = ports[j][Name]
					portsTemplates[j][Protocol] = ports[j][Protocol]
					portsTemplates[j][ContainerPort] = ports[j][ContainerPort]
				} else {
					portsTemplates = append(portsTemplates, ports[j])
				}
			}

		} else {
			deploymentTemplate = append(deploymentTemplate, selfResourceYaml[i])
		}
	}
}
