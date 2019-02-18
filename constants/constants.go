package constants

var (
	LoggingURL           string
	ServicePort          string
	ClusterAPI           string
	KubernetesEngineURL  string
	EnvironmentEngineURL string
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

	CLUSTER_GET_ENDPOINT                = "/antelope/cluster/{cloud_provider}/status/"
	KUBERNETES_GET_CREDENTIALS_ENDPOINT = "/api/v1/credentials/{envId}"
	KUBERNETES_MASTER_PORT              = "6443"

	EnvironmentEngineEndpoint = "/environments/{envId}"
)
