package enum

type IsDeleted int32

const (
	Deleted     IsDeleted = 1
	NoneDeleted IsDeleted = 0
)

func (isDeleted IsDeleted) ToInt32() int32 {
	switch isDeleted {
	case NoneDeleted:
		return int32(0)
	case Deleted:
		return int32(1)
	default:
		return int32(0)
	}
}
