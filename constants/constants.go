package constants

import (
	"github.com/patrickmn/go-cache"
)

var (
	LoggingURL           string
	ServicePort          string
	ServiceGRPCPort      string = "8094"
	ClusterAPI           string
	KubernetesEngineURL  string
	EnvironmentEngineURL string
	VaultURL             string
	CacheObj             *cache.Cache
	RbacURL              string
	WoodpeckerURL        string
)

type RequestType string

const (
	IstioServicePostEndpoint            = "/istio-mesh/deploy"
	IstioServicePutEndpoint             = ""
	KubernetesStatefulSets              = "statefulset"
	KubernetesService                   = "kubernetes-service"
	KubernetesConfigMaps                = "configmap"
	KubernetesDeployment                = "deployment"
	KubernetesPersistentVolumeClaims    = "persistent-volume-claims"
	KubernetesStorageClasses            = "storage-classes"
	KubernetesDaemonSet                 = "daemonsets"
	KubernetesCronJobs                  = "cronjobs"
	IstioComponent                      = "istio-component"
	SERVICE_NAME                        = "kubernetes-services-deployment"
	LOGGING_ENDPOINT                    = "/api/v1/logger"
	LOGGING_LEVEL_INFO                  = "info"
	LOGGING_LEVEL_ERROR                 = "error"
	LOGGING_LEVEL_WARN                  = "warn"
	LOGGING_LEVEL_DEBUG                 = "debug"
	CLUSTER_GET_ENDPOINT                = "/antelope/cluster/{cloud_provider}/status/"
	KUBERNETES_GET_CREDENTIALS_ENDPOINT = "/kube/api/v1/credentials/{envId}"
	KUBERNETES_MASTER_PORT              = "6443"

	ProjectEngineEndpoint = "/raccoon/projects/{project_id}"
	VaultEndpoint         = "/robin/api/v1/project/{project_id}/{cloud_provider}/credentials"
	AWS                   = "AWS"
	AZURE                 = "AZURE"

	BACKEND_LOGGING_ENDPOINT  = "/elephant/api/v1/backend/logging"
	FRONTEND_LOGGING_ENDPOINT = "/elephant/api/v1/frontend/logging"
	Rbac_Token_Info           = "/security/api/rbac/token/info"

	POST   RequestType = "post"
	GET    RequestType = "get"
	PATCH  RequestType = "patch"
	PUT    RequestType = "put"
	DELETE RequestType = "delete"
	LIST   RequestType = "list"
)
