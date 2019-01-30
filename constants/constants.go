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
	SERVICE_NAME             = "kubernetes-services-deployment"
	LOGGING_ENDPOINT         = "/api/v1/logger"
	LOGGING_LEVEL_INFO       = "info"
	LOGGING_LEVEL_ERROR      = "error"
	LOGGING_LEVEL_WARN       = "warn"
)
