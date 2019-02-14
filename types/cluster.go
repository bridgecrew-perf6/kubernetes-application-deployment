package types

import "gopkg.in/mgo.v2/bson"

type AWSCredentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"access_secret"`
	Region    string `json:"region"`
}

type Cluster struct {
	Name            string        `json:"name"`
	ProjectId       string        `json:"project_id"`
	ID              bson.ObjectId `bson:"_id" json:"id"`
	NodePools       []NodePool    `json:"node_pools"`
	CloudProvider   string        `json:"cloud_provider"`
	CloudCredential interface{}   `json:"cloud_credential"`
}

type NodePool struct {
	ID    string `json:"id"`
	Nodes []Node `json:"nodes"`
	Role  string `json:"pool_role"`
}

type Node struct {
	Name               string `json:"name" binding:"required"`
	PublicIP           string `json:"public_ip" binding:"required"`
	PrivateIP          string `json:"private_ip" binding:"required"`
	UserName           string `json:"user_name" binding:"required"`
	Password           string `json:"password" binding:"required"`
	KeyData            string `json:"ssh_key" binding:"required"`
	KeyName            string `json:"key_name"`
	IsPassword         bool   `json:"is_password" binding:"required"`
	Role               string `json:"role" binding:"required"`
	DeployWithPublicIP bool   `json:"deploy_with_public_ip" binding:"required"`
}
