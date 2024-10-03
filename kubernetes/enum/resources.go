package enum

type Resource string

const (
	Pod        Resource = "pods"
	Service    Resource = "services"
	Ingress    Resource = "ingresses"
	Deployment Resource = "deployments"
)

func (r Resource) String() string {
	switch r {
	case Pod:
		return "pods"
	case Service:
		return "services"
	case Ingress:
		return "ingresses"
	case Deployment:
		return "deployments"
	default:
		return "unknown"
	}
}
