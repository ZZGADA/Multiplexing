package resource

import (
	"Multiplexing_/kubernetes/enum"
)

type TcpResource interface {
	CountTcpTotal(list interface{}, command []string, resource enum.Resource) int
	GetTcpResource(port int, state enum.TcpState, commandStr string) (*Tcp, error)
}

// TcpConnectResource Tcp连接资源
type TcpConnectResource struct {
	UnitNum int
	TcpNum  int
}

// Tcp tcp连接结构体
type Tcp struct {
	Port       int
	State      enum.TcpState
	TcpConnect TcpConnectResource
}

// NewTcpInstance 初始化实例对象
// 有三个默认值
func NewTcpInstance(state enum.TcpState, port int) *Tcp {
	return &Tcp{
		State: state,
		Port:  port,
	}
}
