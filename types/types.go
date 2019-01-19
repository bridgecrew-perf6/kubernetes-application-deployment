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
	ClusterInfo *KubernetesClusterInfo   `json:"cluster_info" binding:"required"`
	ServiceData map[string][]interface{} `json:"serivce" binding:"required"`
}
type KubernetesClusterInfo struct {
	Username  string      `json:"username" binding:"required"`
	Password  string      `json:"password" binding:"required"`
	URL       string      `json:"url" binding:"required"`
	TLSConfig interface{} `json:"tls_client_config"`
}

type FilesAttributes struct {
	FileName      string `bson:"file_name" json:"file_name"`
	FilePath      string `bson:"path" json:"path"`
	FileType      string `bson:"file_type" json:"file_type"`
	FileContentes string `bson:"file_contents" json:"file_contents"`
}
