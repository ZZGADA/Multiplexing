package enum

// TcpState  定义自定义类型
type TcpState int

// 定义常量
const (
	CLOSED TcpState = iota
	LISTEN
	SYN_SENT
	SYN_RECEIVED
	ESTABLISHED
	FIN_WAIT_1
	FIN_WAIT_2
	CLOSE_WAIT
	CLOSING
	LAST_ACK
	TIME_WAIT
)

// 实现 String 方法
func (s TcpState) String() string {
	switch s {
	case CLOSED:
		return "CLOSED"
	case LISTEN:
		return "LISTEN"
	case SYN_SENT:
		return "SYN_SENT"
	case SYN_RECEIVED:
		return "SYN_RECEIVED"
	case ESTABLISHED:
		return "ESTABLISHED"
	case FIN_WAIT_1:
		return "FIN_WAIT_1"
	case FIN_WAIT_2:
		return "FIN_WAIT_2"
	case CLOSE_WAIT:
		return "CLOSE_WAIT"
	case CLOSING:
		return "CLOSING"
	case LAST_ACK:
		return "LAST_ACK"
	case TIME_WAIT:
		return "TIME_WAIT"
	default:
		return "UNKNOWN"
	}
}
