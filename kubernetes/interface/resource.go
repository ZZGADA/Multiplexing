package _interface

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
ModulesResource 获取k8s 组件资源
  - 组件资源：
    pod、service、ingress
*/
type ModulesResource interface {
	Get() (metav1.ListInterface, error)
}

type DynamicScalingResource interface {
	ExtendResource(resourceYaml Template, recordExtendResource func(resourceName string), forbiddenExtend func()) bool
	RecallResource(resourceName string) bool
}
