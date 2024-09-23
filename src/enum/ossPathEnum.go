package enum

type OssPath string

const (
	OssPathDefault OssPath = ""
)

func (ossPath OssPath) ToString() string {
	switch ossPath {
	case OssPathDefault:
		return ""
	default:
		return ""
	}
}
