package _interface

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
ResourceInterface 获取k8s 组件资源
  - 组件资源：
    pod、service、ingress
*/
type ResourceInterface interface {
	Get() (metav1.ListInterface, error)
}
