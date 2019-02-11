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
	ClusterInfo *KubernetesClusterInfo   `json:"kubernetes_credentials" binding:"required"`
	ServiceData map[string][]interface{} `json:"service" binding:"required"`
}
type KubernetesClusterInfo struct {
	Username  string      `json:"username" binding:"required"`
	Password  string      `json:"password" binding:"required"`
	URL       string      `json:"url" binding:"required"`
	TLSConfig interface{} `json:"tls_client_config"`
}
type RegistryRequest struct {
	ClusterInfo         *KubernetesClusterInfo `json:"cluster_info" binding:"required"`
	Name                string                 `json:"name"  binding:"required"`
	Namespace           string                 `json:"namespace"  binding:"required"`
	RegistryCredentials `json:"registry_credentials"`
}
type RegistryCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Url      string `json:"url" binding:"required"`
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
