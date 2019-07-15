package types

import (
	"time"
)

type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt    time.Time
}

type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

type Status struct {
	Message string `json:"status" example:"deployment is in progress..."`
}

type ServiceRequest struct {
	//ClusterInfo *KubernetesClusterInfo   `json:"kubernetes_credentials"`
	ProjectId   *string                  `json:"project_id" binding:"required"`
	ServiceData map[string][]interface{} `json:"service" binding:"required"`
}

//type FilesAttributes struct {
//	FileName      string `bson:"file_name" json:"file_name"`
//	FilePath      string `bson:"path" json:"path"`
//	FileType      string `bson:"file_type" json:"file_type"`
//	FileContentes string `bson:"file_contents" json:"file_contents"`
//}

type LoggingRequest struct {
	Message     string `json:"message"`
	Id          string `json:"id"`
	Environment string `json:"environment"`
	Service     string `json:"service"`
	Level       string `json:"level"`
}
type ResponseData struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
	Error      error       `json:"error"`
	Status     string      `json:"status"`
}

type SolutionResp struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

type CacheObjectData struct {
	ProjectId                 string
	KubernetesClusterMasterIp string
	KubernetesCredentials     Credentials
}

type Credentials struct {
	Type              string `json:"type"`
	UserName          string `json:"user_name"`
	Password          string `json:"password"`
	BearerToken       string `json:"bearer_token"`
	KubeConfig        string `json:"kube_config"`
	ClientCertificate string `json:"client_certificate"`
	ClientKey         string `json:"client_key"`
	CaCertificate     string `json:"ca_certificate"`
	ClusterURL        string `json:"cluster_url" bson:"cluster_url"`
	ClusterPort       string `json:"cluster_port" bson:"cluster_port"`
}

const (
	KubeconfigCredentialsType        = "kubeconfig"
	BasicCredentialsType             = "basic"
	BearerCredentialsType            = "bearer"
	ClientCeritficateCredentialsType = "client_certificate"
)
