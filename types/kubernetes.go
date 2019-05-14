package types

import "k8s.io/api/core/v1"

type KubernetesClusterInfo struct {
	//Username           string      `json:"username" binding:"required"`
	//Password           string      `json:"password" binding:"required"`
	URL                string      `json:"url" binding:"required"`
	TLSConfig          interface{} `json:"tls_client_config"`
	ClusterCredentials Credentials `json:"credentials"`
}
type RegistryRequest struct {
	ProjectId *string    `json:"project_id" binding:"required"`
	Secrets   *v1.Secret `json:"secrets" binding:"required"`
}
type RegistryCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Url      string `json:"url" binding:"required"`
}
