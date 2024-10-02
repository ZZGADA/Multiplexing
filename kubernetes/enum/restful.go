package enum

type Restful int64

const (
	GET Restful = iota
	POST
	PUT
	DELETE
)

func (restful Restful) String() string {
	switch restful {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}
