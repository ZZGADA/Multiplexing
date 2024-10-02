package enum

type SubResourceType string

const Exec SubResourceType = "exec"

func (sr SubResourceType) String() string {
	switch sr {
	case Exec:
		return "exec"
	default:
		return "unknown"
	}
}
