package constants

var (
	LoggingURL       string
	IstioEngineURL   string
	KnativeEngineURL string
	ServicePort      string
)

const (
	IstioServicePostEndpoint = "/istio-mesh/deploy"
	IstioServicePutEndpoint  = ""
	KubernetesStatefulSets   = "statefulset"
	KubernetesService        = "kubernetes-service"
	KubernetesConfigMaps     = "configmap"
	KubernetesDeployment     = "deployment"
	KubernetesDaemonSet      = "daemonsets"
	KubernetesCronJobs       = "cronjobs"
	IstioComponent           = "istio-component"
)
