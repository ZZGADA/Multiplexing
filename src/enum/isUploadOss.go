package enum

type IsUploadOss int32

const (
	UploadOss     IsUploadOss = 1
	NoneUploadOss IsUploadOss = 0
)

func (isUploadOss IsUploadOss) ToInt32() int32 {
	switch isUploadOss {
	case NoneUploadOss:
		return int32(0)
	case UploadOss:
		return int32(1)
	default:
		return int32(0)
	}
}
