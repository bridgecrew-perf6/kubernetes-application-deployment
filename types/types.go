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
	Message string `json:"error" example:"status bad request"`
}

type Status struct {
	Message string `json:"status" example:"deployment is in progress..."`
}

type ServiceRequest struct {
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
	Error string `json:"error"`
	Data  string `json:"data"`
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

type LoggingHttpRequest struct {
	RequestId string `json:"request_id"`
	//url of the cloudplex server (e.g. apis.cloudplex.cf)
	Url string `json:"url"`
	//request method (GET/POST/PUT/PATCH/DELETE)
	Method string `json:"method" `
	//request path of backend service
	Path string `json:"path"`
	//request body
	Body string `json:"body"`
	//status code of service
	Status int `json:"status"`
}
type Backendlogging struct {
	//project Id of the project
	ProjectId string `json:"project_id"`
	//resource name like project,network,cluster
	ResourceName string `json:"resource_name"`
	// requester's service name (antelope, msme,raccoon)
	ServiceName string `json:"service_name" binding:"required"`
	//log severity (info,warn,debug,error)
	Severity string `json:"severity" binding:"required"`
	//requested user of platform (haseeb@cloudplex.io)
	UserId string `json:"user_id" binding:"required"`
	//company of the requested user (cloudplex.io)
	Company string `json:"company_id" binding:"required"`
	//message type (stdout,stderr)
	MessageType string `json:"message_type"`
	//Response from actual service when/where log was generated
	Response interface{} `json:"response"`
	//actual message
	Message      interface{} `json:"message" binding:"required" `
	Http_Request struct {
		Request_Id string `json:"request_id"`
		//url of the cloudplex server (e.g. apis.cloudplex.cf)
		Url string `json:"url"`
		//request method (GET/POST/PUT/PATCH/DELETE)
		Method string `json:"method" `
		//request path of backend service
		Path string `json:"path"`
		//request body
		Body string `json:"body"`
		//status code of service
		Status int `json:"status"`
	} `json:"http_request"  binding:"required"`
}

type LoggingRequestFrontend struct {
	Message     string `json:"message"`
	Id          string `json:"id"`
	Environment string `json:"environment"`
	Service     string `json:"service"`
	Level       string `json:"level"`
	//level is info erro
	CompanyId string `json:"company_id"`
	UserId    string `json:"userId"`
	Type      string `json:"type"`
}
