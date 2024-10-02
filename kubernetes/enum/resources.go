package enum

type Resource string

const (
	Pod        Resource = "Pod"
	Service    Resource = "Service"
	Ingress    Resource = "Ingress"
	Deployment Resource = "Deployment"
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
