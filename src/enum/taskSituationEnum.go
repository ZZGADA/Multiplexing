package enum

type Task int64

const (
	TaskSingleFileUpload Task = 1
	TaskSynFileEs        Task = 2
)

func (task Task) ToInt64() int64 {
	switch task {
	case TaskSingleFileUpload:
		return 1
	case TaskSynFileEs:
		return 2
	default:
		return 0
	}
}
