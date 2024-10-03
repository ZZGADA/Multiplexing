package _interface

type Strategy interface {
	RecallResource() string
	ExpandResource(recordExtendResource string)
	CountingResourceExtendTime()
	CheckIfNeedDynamicExtend(parameter interface{}) bool
	CheckIfNeedRecallDeployment() bool
}
