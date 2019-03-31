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
	KubernetesCredentials     struct {
		UserName string
		Password string
	}
}
