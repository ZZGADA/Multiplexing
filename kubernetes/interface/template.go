package _interface

type Template interface {
	CreateResourceYaml(selfResourceYaml map[string]interface{})
}
