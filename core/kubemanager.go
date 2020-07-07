package core

import (
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gedex/inflector"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"io"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"strconv"
	"sync"

	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/constants"
	appKubernetes "bitbucket.org/cloudplex-devs/kubernetes-services-deployment/core/kubernetes"
	v1alpha "bitbucket.org/cloudplex-devs/kubernetes-services-deployment/kubernetes-custom-apis/core/v1"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/types"
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

type KubernetesClient struct {
	Config     *rest.Config
	Client     *kubernetes.Clientset
	Namespaces map[string]bool
	context    *Context
}

type AgentConnection struct {
	connection  *grpc.ClientConn
	agentCtx    context.Context
	agentClient agent_api.AgentServerClient
	cpCtx       *Context
	projectId   string
	companyId   string
	Mux         sync.Mutex
}

func RetryAgentConn(agent *AgentConnection) error {
	err := agent.connection.Close()
	if err != nil {
		utils.Error.Println("error while closing connection :", err.Error())
	}
	count := 0
	flag := true
	for flag && count < 5 {
		conn, err := GetGrpcAgentConnection()
		if err != nil {
			count++
		} else {
			agent.connection = conn.connection
			agent.InitializeAgentClient(agent.projectId, agent.companyId)
			flag = false
		}

		time.Sleep(time.Second * 5)
	}

	if count == 5 {
		utils.Error.Println(errors.New("connection cant be established"))
		return errors.New("connection cant be established")
	}
	return nil
}

func GetGrpcAgentConnection() (*AgentConnection, error) {
	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	conn, err := grpc.Dial(constants.WoodpeckerURL, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		utils.Error.Println("error while connecting with agent :", err)
		return &AgentConnection{}, err
	}

	return &AgentConnection{connection: conn}, nil
}

func (agent *AgentConnection) InitializeAgentClient(projectId, companyId string) error {
	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	agent.projectId = projectId
	agent.companyId = companyId
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 200*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)
	return nil
}

func createKubernetesClient(req *types.KubernetesClusterInfo) (config *rest.Config, client *kubernetes.Clientset, err error) {
	utils.Info.Println("kubernetes api authentication mechanism:", req.ClusterCredentials.Type)
	switch strings.ToLower(req.ClusterCredentials.Type) {
	case types.BasicCredentialsType:
		config = &rest.Config{
			Host:            req.URL,
			Username:        req.ClusterCredentials.UserName,
			Password:        req.ClusterCredentials.Password,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		}
	case types.KubeconfigCredentialsType:
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(req.ClusterCredentials.KubeConfig))
		if err != nil {
			utils.Info.Println(err)
			return nil, nil, err
		}

	case types.BearerCredentialsType:
		if req.ClusterCredentials.BearerToken != "" {
			config = &rest.Config{
				Host:            req.URL,
				BearerToken:     req.ClusterCredentials.BearerToken,
				TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			}
		} else {
			errorStr := "no bearer token found in cluster credentials"
			utils.Info.Println(errorStr)
			return nil, nil, errors.New(errorStr)
		}

	case types.ClientCeritficateCredentialsType:

		if req.ClusterCredentials.ClientCertificate != "" && req.ClusterCredentials.ClientKey != "" {
			config = &rest.Config{
				Host:            req.URL,
				TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			}
			config.TLSClientConfig.CertData = []byte(req.ClusterCredentials.ClientCertificate)
			config.TLSClientConfig.KeyData = []byte(req.ClusterCredentials.ClientKey)
		} else {
			errorStr := "no client cert/key found in cluster credentials"
			utils.Info.Println(errorStr)
			return nil, nil, errors.New(errorStr)
		}
	}

	if req.ClusterCredentials.CaCertificate != "" {
		config.TLSClientConfig.Insecure = false
		config.TLSClientConfig.CAData = []byte(req.ClusterCredentials.CaCertificate)
	}

	client, err = kubernetes.NewForConfig(config)
	return config, client, err
}
func GetKubernetesClient(c *Context, projectId *string) (kubeClient KubernetesClient, err error) {
	kubernetesClusterIp := ""
	kubernetesClusterPort := constants.KUBERNETES_MASTER_PORT
	credentials := types.Credentials{}
	companyId, isCompayId := c.Keys["companyId"].(string)
	cacheId := *projectId
	ok := false
	var data interface{}
	if isCompayId {
		cacheId = cacheId + companyId
		data, ok = constants.CacheObj.Get(cacheId)
	}
	if ok {
		kubernetesData, ok1 := data.(types.CacheObjectData)
		if !ok1 {

		} else {
			kubernetesClusterIp = kubernetesData.KubernetesClusterMasterIp
			credentials = kubernetesData.KubernetesCredentials
		}
	} else {
		project, err := GetProject(c, projectId)
		if err != nil {
			return kubeClient, err
		}
		publicIp, privateIp, err := GetClusterMaster(c, *projectId, project.Data.Cloud, project.Data.CredentialsProfileId)
		if publicIp == "" {
			kubernetesClusterIp = privateIp
		} else {
			kubernetesClusterIp = publicIp
		}
		if err != nil {
			return kubeClient, err
		}
		credentials, err = GetKubernetesCredentials(c, *projectId)

		if kubernetesClusterIp == "" {
			kubernetesClusterIp = credentials.ClusterURL
			kubernetesClusterPort = credentials.ClusterPort
		}

		if err != nil {
			return kubeClient, err
		}
		if isCompayId {
			data := types.CacheObjectData{
				ProjectId:                 cacheId,
				KubernetesClusterMasterIp: kubernetesClusterIp,
				KubernetesCredentials:     credentials,
			}
			constants.CacheObj.Set(cacheId, data, cache.DefaultExpiration)

		}
	}
	kubernetesClusterObj := types.KubernetesClusterInfo{URL: kubernetesClusterIp + ":" + kubernetesClusterPort + "/", ClusterCredentials: credentials}
	config, client, err := createKubernetesClient(&kubernetesClusterObj)
	if err != nil {
		return kubeClient, err
	}
	return KubernetesClient{Config: config, Client: client, Namespaces: make(map[string]bool)}, nil
}
func StartServiceDeployment(req *types.ServiceRequest, cpContext *Context) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting depleoyment")
	}

	var errs []string
	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		return responses, err
	}

	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.deployCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		}
		cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)
		responses[kubeType] = respTemp

	}
	//r, _ := json.Marshal(responses)
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	utils.Info.Println(responses)
	return responses, nil
}
func GetServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		return responses, err
	}

	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.getCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		}
		cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)
		responses[kubeType] = respTemp

	}
	//r, _ := json.Marshal(responses)
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	utils.Info.Println(responses)
	return responses, nil
}
func ListServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
	}

	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.listCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		}
		cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)
		responses[kubeType] = respTemp
	}
	r, _ := json.Marshal(responses)
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	utils.Info.Println(string(r))
	return responses, nil
}
func DeleteServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
	}

	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		utils.Info.Println(len(data))
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.deleteCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		}
		cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)
		responses[kubeType] = respTemp
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ";")
		return nil, errors.New(finalErr)

	}
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	return responses, nil
}
func PatchServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {

	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
	}
	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		utils.Info.Println(len(data))
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.patchCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		}
		cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)
		responses[kubeType] = respTemp

	}
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	return responses, nil
}
func PutServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}

	agent, err := GetGrpcAgentConnection()
	if err != nil {
		utils.Error.Println(err)
		return responses, err
	}
	defer agent.connection.Close()

	err = agent.InitializeAgentClient(*req.ProjectId, cpContext.GetString("company_id"))
	if err != nil {
		return responses, err
	}

	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		utils.Info.Println(len(data))
		if len(data) == 0 {
			continue
		}
		respTemp, err = agent.putCRDS(kubeType, data, *req.ProjectId, cpContext.GetString("company_id"))
		if err != nil {
			cpContext.SendFrontendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR)
			errs = append(errs, err.Error())
		} else {
			cpContext.SendFrontendLogs(respTemp, constants.LOGGING_LEVEL_INFO)

			responses[kubeType] = respTemp
		}
	}
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	return responses, nil
}

/*func (agent *AgentConnection) deployStatefulSets(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.StatefulSet{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) deployKubernetesService(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.Service{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) deployKubernetesConfigMap(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.ConfigMap{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) deployKubernetesDeployment(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {

	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.Deployment{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil

}
func (agent *AgentConnection) deployKubernetesPVC(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {

	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.PersistentVolumeClaim{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) deployKubernetesStorageClasses(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []storage.StorageClass{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	for i := range req {
		var responseObj types.SolutionResp

		if req[i].Namespace != "" {
			_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", "ns", req[i].Namespace},
			})
			if err != nil {
				response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", req[i].Namespace},
				})
				if err != nil {
					errs = append(errs, err.Error())
					responseObj.Error = err.Error()
					utils.Error.Println(err)
					return resp, err
				}
				utils.Info.Println(response.Stdout)
			}
		}

		raw, err := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println(err)
				return resp, err
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
			return resp, err
		} else {
			responseObj.Data = kubectlResp.Stdout
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
		raw, _ = json.Marshal(responseObj)
		utils.Info.Println("response payload", string(raw))
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}*/
func (agent *AgentConnection) deployCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []interface{}
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}

	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {

		responseObj, _ := agent.crdManager(runtimeConfig[i], "post")
		resp = append(resp, responseObj)

	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}

/*func (agent *AgentConnection) getStatefulSets(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.StatefulSet{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) getKubernetesService(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.Service{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) getKubernetesConfigMap(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.ConfigMap{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) getKubernetesDeployment(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.Deployment{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) getKubernetesPVC(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.PersistentVolumeClaim{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) getKubernetesStorageClass(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []storage.StorageClass{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}*/
func (agent *AgentConnection) getCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range runtimeConfig {
		//rest.InClusterConfig()

		responseObj, _ := agent.crdManager(runtimeConfig[i], "get")
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return resp, err
}
func (agent *AgentConnection) listCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range runtimeConfig {
		//rest.InClusterConfig()

		responseObj, _ := agent.crdManager(runtimeConfig[i], "list")

		/*
			//kind to crdplural  for example kind=VirtualService and plural=virtualservices
			crdPlural := utils.Pluralize(strings.ToLower(runtimeConfig[i].Kind))
			namespace := ""
			if runtimeConfig[i].Namespace == "" {
				namespace = "default"
			} else {
				namespace = runtimeConfig[i].Namespace
			}
			alphaClient, err := c.getCRDClient(runtimeConfig[i].APIVersion)
			if err != nil {

			}
			var responseObj types.SolutionResp
			data, err := alphaClient.NewRuntimeConfigs(namespace, crdPlural).Get(runtimeConfig[i].Name)
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("failed to fetch data. Error: ", err)
			} else {
				dd, _ := json.Marshal(data)
				responseObj.Data = data
				utils.Info.Println(string(dd))
			}*/

		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return resp, err
}

//func (c *KubernetesClient) getCRDClient(apiVersion string) (*v1alpha.RuntimeConfigV1Alpha1Client, error) {
//	groupInfo := strings.Split(apiVersion, "/")
//	if len(groupInfo) == 0 {
//		utils.Error.Println("apiVersion " + apiVersion + " is wrong")
//		return nil, errors.New("apiVersion " + apiVersion + " is wrong")
//
//	}
//	groupName := ""
//	groupVersion := ""
//	apiPath := ""
//	if len(groupInfo) == 1 {
//		groupName = ""
//		groupVersion = groupInfo[0]
//		apiPath = "/api"
//	} else {
//		groupName = groupInfo[0]
//		groupVersion = groupInfo[1]
//		apiPath = "/apis"
//	}
//	schemaDef := schema.GroupVersion{Group: groupName, Version: groupVersion}
//	alphaClient, err := v1alpha.NewClient(c.Config, schemaDef, apiPath)
//	if err != nil {
//		utils.Error.Println(err)
//		return nil, err
//	}
//	return alphaClient, nil
//}

/*func (agent *AgentConnection) deleteStatefulSets(data []interface{}, projectId string, companyId string) error {

	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	req := []v1.StatefulSet{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	for i := range req {
		var responseObj types.SolutionResp
		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", req[i].Kind, req[i].Name, "-n", req[i].Namespace},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (agent *AgentConnection) deleteKubernetesService(data []interface{}, projectId string, companyId string) error {
	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	req := []v12.Service{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	for i := range req {
		var responseObj types.SolutionResp
		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", req[i].Kind, req[i].Name, "-n", req[i].Namespace},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (agent *AgentConnection) deleteKubernetesConfigMap(data []interface{}, projectId string, companyId string) error {
	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	req := []v12.ConfigMap{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	for i := range req {
		var responseObj types.SolutionResp
		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", req[i].Kind, req[i].Name, "-n", req[i].Namespace},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (agent *AgentConnection) deleteKubernetesDeployment(data []interface{}, projectId string, companyId string) error {
	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	req := []v1.Deployment{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	for i := range req {
		var responseObj types.SolutionResp
		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", req[i].Kind, req[i].Name, "-n", req[i].Namespace},
		})
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}*/
func (agent *AgentConnection) deleteCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {
		responseObj, _ := agent.crdManager(runtimeConfig[i], "delete")
		if responseObj.Error != "" {
			errs = append(errs, responseObj.Error)
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return resp, err
}

/*func (agent *AgentConnection) patchStatefulSets(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.StatefulSet{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		_, err = agent.CreateFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "~/" + req[i].Name + ".json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				utils.Error.Println(err)
			}
			utils.Info.Println(feature.Stdout)
		}

		_, err = agent.DeleteFile(req[i].Name, string(raw))
		if err != nil {
			utils.Error.Println(err)
			return data, err
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println(err)
			return data, err
		} else {
			responseObj.Data = kubectlResp.Stdout
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) patchKubernetesService(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	var errs []string
	svc := appKubernetes.NewServicesLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.Service{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := svc.PatchService(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes service deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes service deployed successfully")

			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) patchKubernetesConfigMap(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	var errs []string
	svc := appKubernetes.NewConfigLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.ConfigMap{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := svc.PatchConfigMap(req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes configmap deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes configmap deployed successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (agent *AgentConnection) patchKubernetesDeployment(data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewDeploymentLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.Deployment{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := depObj.PatchDeployments(req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes deployment deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes deployment deployed successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}

	return resp, nil
}*/
func (agent *AgentConnection) patchCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {
		responseObj, _ := agent.crdManager(runtimeConfig[i], "patch")
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return resp, err
}

/*func (c *KubernetesClient) putStatefulSets(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	statefulset := appKubernetes.NewStatefulsetLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.StatefulSet{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := statefulset.UpdateStatefulSets(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("fail to get kubernetes statefulset. Error: ", err)
		} else {
			utils.Info.Println("kubernetes statefulset fetched successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (c *KubernetesClient) putKubernetesService(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	svc := appKubernetes.NewServicesLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.Service{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := svc.UpdateService(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes service deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes service deployed successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (c *KubernetesClient) putKubernetesConfigMap(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	svc := appKubernetes.NewConfigLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v12.ConfigMap{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := svc.UpdateConfigMap(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes configmap deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes configmap deployed successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (c *KubernetesClient) putKubernetesDeployment(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewDeploymentLauncher(c.Client)
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	req := []v1.Deployment{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	for i := range req {
		var responseObj types.SolutionResp
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := depObj.UpdateDeployments(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes deployment deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes deployment deployed successfully")
			responseObj.Data = respTemp
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}

	return resp, nil
}*/
func (agent *AgentConnection) putCRDS(key string, data []interface{}, projectId string, companyId string) (resp []interface{}, err error) {
	if projectId == "" || companyId == "" {
		return resp, errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)

	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return resp, err
	}
	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {
		responseObj, _ := agent.crdManager(runtimeConfig[i], "put")
		/*var responseObj types.SolutionResp
		raw, err := json.Marshal(runtimeConfig[i])
		utils.Info.Println(string(raw))
		runtimeObj := v1alpha.RuntimeConfig{}
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}
		err = json.Unmarshal(raw, &runtimeObj)
		if err != nil {
			utils.Error.Println(err)
			return resp, err
		}
		rest.InClusterConfig()
		//kind to crdplural  for example kind=VirtualService and plural=virtualservices
		crdPlural := utils.Pluralize(strings.ToLower(runtimeConfig[i].Kind))
		namespace := ""
		if runtimeConfig[i].Namespace == "" {
			namespace = "default"
		} else {
			namespace = runtimeConfig[i].Namespace
		}
		alphaClient, err := c.getCRDClient(runtimeConfig[i].APIVersion)
		if err != nil {

		}
		data, err := alphaClient.NewRuntimeConfigs(namespace, crdPlural).Update(runtimeConfig[i])
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("failed to fetch data. Error: ", err)
		} else {
			utils.Info.Println("")
			dd, _ := json.Marshal(data)

			responseObj.Data = data
			utils.Info.Println(string(dd))
		}*/
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return resp, err
}

func (c *KubernetesClient) CreateDockerRegistryCredentials(req *types.RegistryRequest) (*v12.Secret, error) {

	if req.Secrets.Namespace == "" {
		req.Secrets.Namespace = "default"
	}
	secrets := appKubernetes.NewSecretsLauncher(c.Client)
	return secrets.CreateRegistrySecret(req.Secrets)
}
func (c *KubernetesClient) GetDockerRegistryCredentials(name, namespace string) (*v12.Secret, error) {
	if namespace == "" {
		namespace = "default"
	}
	secrets := appKubernetes.NewSecretsLauncher(c.Client)
	return secrets.GetRegistrySecret(name, namespace)
}
func (c *KubernetesClient) DeleteDockerRegistryCredentials(name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	secrets := appKubernetes.NewSecretsLauncher(c.Client)
	return secrets.DeleteRegistrySecret(name, namespace)
}

func (c *KubernetesClient) ListStatefulSets(namespace string) (*v1.StatefulSetList, error) {
	statefulsetObj := appKubernetes.NewStatefulsetLauncher(c.Client)
	return statefulsetObj.GetAllStatefulSet(namespace)
}
func (c *KubernetesClient) GetStatefulSet(name, namespace string) (*v1.StatefulSet, error) {
	statefulsetObj := appKubernetes.NewStatefulsetLauncher(c.Client)
	return statefulsetObj.GetStatefulSet(name, namespace)
}
func (c *KubernetesClient) DeleteStatefulSet(name, namespace string) error {
	statefulsetObj := appKubernetes.NewStatefulsetLauncher(c.Client)
	return statefulsetObj.DeleteStatefulSet(name, namespace)
}

func (c *KubernetesClient) ListDeployments(namespace string) (*v1.DeploymentList, error) {
	deploymentObj := appKubernetes.NewDeploymentLauncher(c.Client)
	return deploymentObj.GetAllDeployments(namespace)
}
func (c *KubernetesClient) GetDeployment(name, namespace string) (*v1.Deployment, error) {
	deploymentObj := appKubernetes.NewDeploymentLauncher(c.Client)
	return deploymentObj.GetDeployments(name, namespace)
}
func (c *KubernetesClient) DeleteDeployment(name, namespace string) error {
	deploymentObj := appKubernetes.NewDeploymentLauncher(c.Client)
	return deploymentObj.DeleteDeployments(name, namespace)
}

func (c *KubernetesClient) ListKubernetesServices(namespace string) (*v12.ServiceList, error) {

	serviceObj := appKubernetes.NewServicesLauncher(c.Client)
	return serviceObj.GetAllServices(namespace)
}
func (c *KubernetesClient) GetKubernetesService(namespace, name string) (*v12.Service, error) {
	serviceObj := appKubernetes.NewServicesLauncher(c.Client)
	return serviceObj.GetService(name, namespace)
}
func (agent *AgentConnection) GetKubernetesServiceExternalIp(namespace, name string) (string, error) {

	resp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubeclt",
		Args:    []string{"get", "svc", name, "-n", namespace, "-o", "json"},
	})
	if err != nil {
		err = RetryAgentConn(agent)
		if err != nil {
			utils.Error.Print(err)
			return "", err
		}

		resp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubeclt",
			Args:    []string{"get", "svc", name, "-n", namespace, "-o", "json"},
		})
		if err != nil {
			utils.Error.Println("getting ingress external IP", err.Error())
			return "", err
		}
	}

	defer agent.connection.Close()

	var ingress v12.Service
	b := []byte(resp.Stdout[0])
	err = json.Unmarshal(b, &ingress)
	if err != nil {
		return "", err
	}

	externalIp := ""
	for _, ingress := range ingress.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			externalIp = ingress.IP
			break
		} else if ingress.Hostname != "" {
			externalIp = ingress.Hostname
			break
		}
	}

	return externalIp, nil
}

func (agent *AgentConnection) GetKubernetesHealth() (types.HealthObject, error) {

	var result types.HealthObject
	resp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"get", "nodes", "-o", "json"},
	})
	if err != nil {
		err = RetryAgentConn(agent)
		if err != nil {
			utils.Error.Print(err)
			return result, err
		}

		resp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", "nodes", "-o", "json"},
		})
		if err != nil {
			utils.Error.Println("getting nodes : ", err.Error())
			return result, err
		}
	}

	defer agent.connection.Close()

	var nodeList v12.NodeList
	b := []byte(resp.Stdout[0])
	err = json.Unmarshal(b, &nodeList)
	if err != nil {
		return result, err
	}

	var summary types.ClusterSummary
	for _, node := range nodeList.Items {
		//Calling Describe node
		resp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"describe", "node", node.Name},
		})
		if err != nil {
			err = RetryAgentConn(agent)
			if err != nil {
				utils.Error.Print(err)
				return result, err
			}

			resp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"describe", "nodes"},
			})
			if err != nil {
				utils.Error.Println("describing nodes : ", err.Error())
				return result, err
			}
		}

		defer agent.connection.Close()

		payload := resp.Stdout[0]
		scanner := bufio.NewScanner(strings.NewReader(payload))
		var temp types.ClusterNodes
		temp.Name = node.Name
		temp.CreationTime = node.CreationTimestamp.String()

		for scanner.Scan() {
			//fmt.Println(".........")
			temp_str := scanner.Text()
			if strings.Contains(temp_str, "Non-terminated Pods:") {
				temp.UsedResources.Pods, err = strconv.Atoi(strings.Split(strings.Split(temp_str, "(")[1], " ")[0])
				summary.UsedResources.Pods = summary.UsedResources.Pods + temp.UsedResources.Pods
			}
			if strings.Contains(temp_str, "cpu") && strings.Contains(temp_str, "%") {
				temp.UsedResources.CPUPercentage, err = strconv.Atoi(strings.Split(strings.Split(temp_str, "(")[1], "%")[0]) //CPU
				_ = err
				summary.UsedResources.CPUPercentage = summary.UsedResources.CPUPercentage + temp.UsedResources.CPUPercentage

			}
			if strings.Contains(temp_str, "memory") && strings.Contains(temp_str, "%") {
				temp.UsedResources.MemoryPercentage, err = strconv.Atoi(strings.Split(strings.Split(temp_str, "(")[1], "%")[0]) //Memory
				_ = err
				summary.UsedResources.MemoryPercentage = summary.UsedResources.MemoryPercentage + temp.UsedResources.MemoryPercentage

			}

			if strings.Contains(temp_str, "ephemeral-storage") && strings.Contains(temp_str, "%") {
				temp.UsedResources.StoragePercentage, err = strconv.Atoi(strings.Split(strings.Split(temp_str, "(")[1], "%")[0]) //Storage
				_ = err
				summary.UsedResources.StoragePercentage = summary.UsedResources.StoragePercentage + temp.UsedResources.StoragePercentage
			}

			if strings.Contains(temp_str, "ProviderID:") && len(strings.Split(temp_str, ":")) > 1 {
				temp.ProviderId = strings.TrimSpace(strings.Split(temp_str, ":")[1]) //Provider
			}
			if strings.Contains(temp_str, "Hostname:") && len(strings.Split(temp_str, ":")) == 2 {
				temp.NodeHostname = strings.TrimSpace(strings.Split(temp_str, ":")[1]) //Hostname
			}
			if strings.Contains(temp_str, "InternalIP:") && len(strings.Split(temp_str, ":")) == 2 {
				temp.NodeInternalIp = strings.TrimSpace(strings.Split(temp_str, ":")[1]) //InternalIP
			}
			if strings.Contains(temp_str, "ExternalIP:") && len(strings.Split(temp_str, ":")) == 2 {
				temp.NodeExternalIp = strings.TrimSpace(strings.Split(temp_str, ":")[1]) //External IP
			}
		}

		_ = node

		//Capacity
		if node.Status.Capacity.Cpu() != nil {
			temp.Capacity.CPU = node.Status.Capacity.Cpu().Value()
			summary.Capacity.CPU += temp.Capacity.CPU
		}

		if node.Status.Capacity.Memory() != nil {
			temp.Capacity.Memory = node.Status.Capacity.Memory().Value()
			summary.Capacity.Memory += temp.Capacity.Memory

		}

		if node.Status.Capacity.StorageEphemeral() != nil {
			temp.Capacity.Storage = node.Status.Capacity.StorageEphemeral().Value()
			summary.Capacity.Storage += temp.Capacity.Storage

		}

		if node.Status.Capacity.Pods() != nil {
			temp.Capacity.Pods = node.Status.Capacity.Pods().Value()
			summary.Capacity.Pods += temp.Capacity.Pods
		}
		//Allocatable
		if node.Status.Allocatable.Cpu() != nil {
			temp.Allocatable.CPU = node.Status.Allocatable.Cpu().Value()
			summary.Allocatable.CPU += temp.Allocatable.CPU

		}

		if node.Status.Allocatable.Memory() != nil {
			temp.Allocatable.Memory = node.Status.Allocatable.Memory().Value()
			summary.Allocatable.Memory += temp.Allocatable.Memory

		}

		if node.Status.Allocatable.StorageEphemeral() != nil {
			temp.Allocatable.Storage = node.Status.Allocatable.StorageEphemeral().Value()
			summary.Allocatable.Storage += temp.Allocatable.Storage

		}

		if node.Status.Allocatable.Pods() != nil {
			temp.Allocatable.Pods = node.Status.Allocatable.Pods().Value()
			summary.Allocatable.Pods += temp.Allocatable.Pods
		}
		result.ClusterNodes = append(result.ClusterNodes, temp)

	}

	if len(nodeList.Items) != 0 {
		summary.UsedResources.StoragePercentage /= len(nodeList.Items)
		summary.UsedResources.MemoryPercentage /= len(nodeList.Items)
		summary.UsedResources.CPUPercentage /= len(nodeList.Items)
	}

	result.ClusterSummary = summary
	return result, nil
}
func (c *KubernetesClient) DeleteKubernetesService(name, namespace string) error {

	serviceObj := appKubernetes.NewServicesLauncher(c.Client)
	return serviceObj.DeleteServices(name, namespace)
}

func (c *KubernetesClient) ListConfigMaps(namespace string) (*v12.ConfigMapList, error) {

	configMapsObj := appKubernetes.NewConfigLauncher(c.Client)
	return configMapsObj.GetAllConfigMap(namespace)
}
func (c *KubernetesClient) GetConfigMap(name, namespace string) (*v12.ConfigMap, error) {
	configMapsObj := appKubernetes.NewConfigLauncher(c.Client)
	return configMapsObj.GetConfigMap(name, namespace)
}
func (c *KubernetesClient) DeleteConfigMap(name, namespace string) error {

	configMapsObj := appKubernetes.NewConfigLauncher(c.Client)
	return configMapsObj.DeleteConfigMap(name, namespace)
}

/*func (c *KubernetesClient) ListPersistentVolumes(namespace string) (*v1.DeploymentList, error) {

	deploymentObj := appKubernetes.NewDeploymentLauncher(c.Client)
	return deploymentObj.GetAllDeployments(namespace)
}
func (c *KubernetesClient) GetPersistentVolume(namespace, name string) (*v1.Deployment, error) {
	statefulsetObj := appKubernetes.NewDeploymentLauncher(c.Client)
	return statefulsetObj.GetDeployments(namespace, name)
}
func (c *KubernetesClient) DeletePersistentVolume(namespace, name string) error {

	statefulsetObj := appKubernetes.NewStatefulsetLauncher(c.Client)
	return statefulsetObj.DeleteStatefulSet(namespace, name)
}*/

func findKey(istiojsonData map[string]interface{}, key string) (string, error) {
	keyData, ok := istiojsonData[key]
	if !ok {
		utils.Error.Println(key + " is missing in JSON")
		return "", errors.New(key + " is missing in JSON")
	}
	data, ok := keyData.(string)
	if !ok {
		utils.Error.Println(key + " type is not string")
		return "", errors.New(key + " type is not string")
	}
	return data, nil
}

func (agent *AgentConnection) crdManager(runtimeConfig interface{}, method string) (responseObj types.SolutionResp, err error) {

	raw, err := json.Marshal(runtimeConfig)
	utils.Info.Println(string(raw))
	runtimeObj := v1alpha.RuntimeConfig{}
	if err != nil {
		utils.Error.Println(err)
		responseObj.Error = err.Error()
		return responseObj, err
	}
	err = json.Unmarshal(raw, &runtimeObj)
	if err != nil {
		utils.Error.Println(err)
		responseObj.Error = err.Error()
		return responseObj, err
	}

	if runtimeObj.Kind == "" || runtimeObj.APIVersion == "" {
		utils.Error.Println("Kind/APIVersion is empty")
		responseObj.Error = "Kind/APIVersion is empty"
		return responseObj, errors.New("Kind/APIVersion is empty")
	}

	//kind to crdplural  for example kind=VirtualService and plural=virtualservices
	crdPlural := inflector.Pluralize(strings.ToLower(runtimeObj.Kind))

	namespace := runtimeObj.Namespace

	utils.Info.Println(crdPlural, namespace)

	if namespace != "" {

		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", "ns", namespace},
		})
		if err != nil {
			response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"create", "ns", namespace},
			})
			if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || !strings.Contains(err.Error(), "already exists")) {
				err = RetryAgentConn(agent)
				if err != nil {
					return responseObj, err
				}
				response, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
					Command: "kubectl",
					Args:    []string{"create", "ns", namespace},
				})
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					utils.Error.Println(namespace+" namespace creation failed", err)
					responseObj.Error = err.Error()
					return responseObj, err
				} else if response != nil {
					utils.Info.Println(response.Stdout)
				}
			} else if err != nil && !strings.Contains(err.Error(), "already exists") {
				utils.Error.Println(namespace+" namespace creation failed", err)
				responseObj.Error = err.Error()
				return responseObj, err
			}
		}

		_, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"label", "ns", namespace, "istio-injection=enabled"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || !strings.Contains(err.Error(), "already has a value (enabled)")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"label", "ns", namespace, "istio-injection=enabled"},
			})
			if err != nil && !strings.Contains(err.Error(), "already has a value (enabled)") {
				utils.Error.Println(namespace+" label attachment failed", err)
				responseObj.Error = err.Error()
				return responseObj, err
			}

		} else if err != nil && !strings.Contains(err.Error(), "already has a value (enabled)") {
			utils.Error.Println(namespace+" label attachment failed", err)
			responseObj.Error = err.Error()
			return responseObj, err
		}

	}

	//var data interface{}
	var data2 string
	var podErrors []string
	switch method {
	case "post":

		if runtimeObj.Kind == "Certificate" || runtimeObj.Kind == "ClusterIssuer" {
			agent.InstallCertManager()
		}

		name := fmt.Sprintf("%s-%s", runtimeObj.Name, runtimeObj.Kind)
		_, err = agent.CreateFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.CreateFile(name, string(raw))
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
			}
		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
		}

		flag := true
		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			flag = false
			kubectlStreamResp, err = agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"create", "-f", "/tmp/" + name + ".json"},
			})
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl stream :", err)
			}

			for {
				feature, err := kubectlStreamResp.Recv()
				if err == io.EOF || err == nil {
					break
				}
				if err != nil {
					//responseObj.Error = err.Error() + "service :" + name
					utils.Error.Println("kubectl stream reading :", err)
					break
				} else {
					utils.Info.Println(feature.Stdout, feature.Stderr)
				}
			}

		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
			utils.Error.Println("kubectl stream :", err)
		}
		for flag {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF || err == nil {
				break
			}
			if err != nil {
				//responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl stream reading :", err)
				break
			} else {
				utils.Info.Println(feature.Stdout, feature.Stderr)
			}
		}

		//_, err = agent.DeleteFile(name, string(raw))
		//if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
		//	err = RetryAgentConn(agent)
		//	if err != nil {
		//		return responseObj, err
		//	}
		//
		//	_, err = agent.DeleteFile(name, string(raw))
		//	if err != nil {
		//		responseObj.Error = err.Error()
		//	}
		//} else if err != nil {
		//	responseObj.Error = err.Error()
		//}

		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
		}

	case "get":
		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}
			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error() + "service :" + runtimeObj.Kind + "-" + runtimeObj.Name
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + runtimeObj.Kind + "-" + runtimeObj.Name
			utils.Error.Println(err)
		} else {
			podStatus, err := agent.GetPodStatus(runtimeObj.Name, runtimeObj.Namespace)
			if err != nil {
				return responseObj, err
			}
			var reasons []string
			for _, value := range podStatus.ContainerStatuses {
				if value.State.Waiting != nil {
					reasons = append(reasons, value.State.Waiting.Message)
				}
			}
			for _, value := range reasons {
				podErrors = append(podErrors, value)
			}
			data2 = kubectlResp.Stdout[0]
		}
	case "put":
		if runtimeObj.Kind == "Certificate" || runtimeObj.Kind == "ClusterIssuer" {
			agent.InstallCertManager()
		}

		name := fmt.Sprintf("%s-%s", runtimeObj.Name, runtimeObj.Kind)
		_, err = agent.CreateFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.CreateFile(name, string(raw))
			if err != nil {
				responseObj.Error = err.Error()
			}
		} else if err != nil {
			responseObj.Error = err.Error()
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlStreamResp, err = agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream :", err)
			}

		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl stream :", err)
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF || err == nil {
				break
			}
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream reading :", err)
				break
			}
			utils.Info.Println(feature.Stdout, feature.Stderr)
		}

		//_, err = agent.DeleteFile(name, string(raw))
		//if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
		//	err = RetryAgentConn(agent)
		//	if err != nil {
		//		return responseObj, err
		//	}
		//
		//	_, err = agent.DeleteFile(name, string(raw))
		//	if err != nil {
		//		responseObj.Error = err.Error()
		//	}
		//} else if err != nil {
		//	responseObj.Error = err.Error()
		//}

		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
		}

	case "patch":
		if runtimeObj.Kind == "Certificate" || runtimeObj.Kind == "ClusterIssuer" {
			agent.InstallCertManager()
		}

		name := fmt.Sprintf("%s-%s", runtimeObj.Name, runtimeObj.Kind)
		_, err = agent.DeleteFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}
			_, _ = agent.DeleteFile(name, string(raw))
		}
		_, err = agent.CreateFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.CreateFile(name, string(raw))
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
			}
		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlStreamResp, err = agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
			})
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl stream :", err)
			}

		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
			utils.Error.Println("kubectl stream :", err)
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				//responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl stream reading :", err)
				break
			}
			utils.Info.Println(feature.Stdout, feature.Stderr)
		}

		//_, err = agent.DeleteFile(name, string(raw))
		//if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
		//	err = RetryAgentConn(agent)
		//	if err != nil {
		//		return responseObj, err
		//	}
		//
		//	_, err = agent.DeleteFile(name, string(raw))
		//	if err != nil {
		//		responseObj.Error = err.Error()
		//	}
		//} else if err != nil {
		//	responseObj.Error = err.Error()
		//}

		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "no registered agent with")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error() + "service :" + name
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error() + "service :" + name
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
			podStatus, err := agent.GetPodStatus(runtimeObj.Name, runtimeObj.Namespace)
			if err != nil {
				return responseObj, err
			}
			var reasons []string
			for _, value := range podStatus.ContainerStatuses {
				if value.State.Waiting != nil {
					reasons = append(reasons, value.State.Waiting.Message)
				}
			}
			fmt.Println(reasons)
		}
	case "delete":
		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubeResponse, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || !strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"delete", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace},
			})
			if err != nil && !strings.Contains(err.Error(), "not found") {
				responseObj.Error = err.Error()
				utils.Error.Println(err)
			}

		} else if err != nil && !strings.Contains(err.Error(), "not found") {
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else if kubeResponse != nil {
			data2 = kubeResponse.Stdout[0]
		}
	case "list":
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			data2 = kubectlResp.Stdout[0]
		}
	}

	if err != nil {
		responseObj.Error = err.Error()
		return responseObj, err
	} else {
		responseObj.Data = data2
		responseObj.PodErrors = podErrors
		utils.Info.Println(data2)
		utils.Info.Println(responseObj)
	}

	utils.Info.Println("response payload", responseObj)
	return responseObj, nil
}

/*
case constants.GET:
			data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Get(runtimeConfig.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string))
			for data == nil && err != nil {
				if err.Error() == "" {
					time.Sleep(1 * time.Second)
					data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Get(runtimeConfig.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string))
				} else {
					break
				}
			}
		case constants.PUT:
			data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Get(runtimeConfig.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string))
			for data == nil && err != nil {
				if err.Error() == "" {
					time.Sleep(1 * time.Second)
					data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Update(runtimeConfig)
				} else {
					break
				}
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
		}

	case "get":
		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}
			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			data2 = kubectlResp.Stdout[0]
		}
	case "put":
		name := fmt.Sprintf("%s-%s", runtimeObj.Name, runtimeObj.Kind)
		_, err = agent.CreateFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.CreateFile(name, string(raw))
			if err != nil {
				responseObj.Error = err.Error()
			}
		} else if err != nil {
			responseObj.Error = err.Error()
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlStreamResp, err = agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream :", err)
			}

		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl stream :", err)
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF || err == nil {
				break
			}
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream reading :", err)
				break
			}
			utils.Info.Println(feature.Stdout, feature.Stderr)
		}

		//_, err = agent.DeleteFile(name, string(raw))
		//if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
		//	err = RetryAgentConn(agent)
		//	if err != nil {
		//		return responseObj, err
		//	}
		//
		//	_, err = agent.DeleteFile(name, string(raw))
		//	if err != nil {
		//		responseObj.Error = err.Error()
		//	}
		//} else if err != nil {
		//	responseObj.Error = err.Error()
		//}

		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
		}

	case "patch":

		name := fmt.Sprintf("%s-%s", runtimeObj.Name, runtimeObj.Kind)
		_, err = agent.CreateFile(name, string(raw))
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.CreateFile(name, string(raw))
			if err != nil {
				responseObj.Error = err.Error()
			}
		} else if err != nil {
			responseObj.Error = err.Error()
		}

		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "transport is closing")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlStreamResp, err = agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"apply", "-f", "/tmp/" + name + ".json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream :", err)
			}

		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl stream :", err)
		}
		for {
			feature, err := kubectlStreamResp.Recv()
			if err == io.EOF || err == nil {
				break
			}
			if err != nil {
				//responseObj.Error = err.Error()
				utils.Error.Println("kubectl stream reading :", err)
				break
			}
			utils.Info.Println(feature.Stdout, feature.Stderr)
		}

		//_, err = agent.DeleteFile(name, string(raw))
		//if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
		//	err = RetryAgentConn(agent)
		//	if err != nil {
		//		return responseObj, err
		//	}
		//
		//	_, err = agent.DeleteFile(name, string(raw))
		//	if err != nil {
		//		responseObj.Error = err.Error()
		//	}
		//} else if err != nil {
		//	responseObj.Error = err.Error()
		//}

		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"get", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace, "-o", "json"},
			})
			if err != nil {
				responseObj.Error = err.Error()
				utils.Error.Println("kubectl :", err)
			} else {
				fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
				data2 = kubectlResp.Stdout[0]
			}
		} else if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println("kubectl :", err)
		} else {
			fmt.Println(kubectlResp.Stdout, kubectlResp.Stderr, "haroon")
			data2 = kubectlResp.Stdout[0]
		}
	case "delete":
		if strings.Contains(runtimeObj.APIVersion, "serving.knative") {
			runtimeObj.Kind = "ksvc"
		}

		_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"delete", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace},
		})
		if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || !strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "transport is closing") || strings.Contains(err.Error(), "upstream request timeout")) {
			err = RetryAgentConn(agent)
			if err != nil {
				return responseObj, err
			}

			_, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
				Command: "kubectl",
				Args:    []string{"delete", runtimeObj.Kind, runtimeObj.Name, "-n", runtimeObj.Namespace},
			})
			if err != nil && !strings.Contains(err.Error(), "not found") {
				responseObj.Error = err.Error()
				utils.Error.Println(err)
			}

		} else if err != nil && !strings.Contains(err.Error(), "not found") {
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		}
	case "list":
		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
			Command: "kubectl",
			Args:    []string{"get", runtimeObj.Kind, "-n", runtimeObj.Namespace, "-o", "json"},
		})
		if err != nil {
			responseObj.Error = err.Error()
			utils.Error.Println(err)
		} else {
			data2 = kubectlResp.Stdout[0]
		}
	}

			dd, _ := json.Marshal(data)
			utils.Info.Println(string(dd))
		}

*/
//func (agent *AgentConnection) AgentCrdManager(method constants.RequestType, data []interface{}, projectId string, companyId string, kubeType string) (resp []interface{}, err error) {
//
//	if projectId == "" || companyId == ""{
//		return
//	}
//	md := metadata.Pairs(
//		"name", *GetSha256(&projectId, &companyId),
//	)
//	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
//	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
//	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)
//
//	var req interface{}
//	switch kubeType {
//	case constants.KubernetesStatefulSets:
//		raw, err := json.Marshal(data)
//		if err != nil {
//			utils.Error.Println(err)
//			return resp, err
//		}
//		req = []v1.StatefulSet{}
//		err = json.Unmarshal(raw, &req)
//		if err != nil {
//			utils.Error.Println(err)
//			return resp, err
//		}
//
//	case constants.KubernetesService:
//
//	case constants.KubernetesConfigMaps:
//
//	case constants.KubernetesDeployment:
//
//	case constants.KubernetesPersistentVolumeClaims:
//
//	case constants.KubernetesStorageClasses:
//
//	default:
//		//for now default case is for istio and knative
//		respTemp, err = c.deployCRDS(kubeType, data)
//	}
//	switch method {
//	case constants.POST:
//		for i := range req {
//			var responseObj types.SolutionResp
//
//			if req[i].Namespace != "" {
//				_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//					Command: "kubectl",
//					Args:    []string{"get", "ns", req[i].Namespace},
//				})
//				if err != nil {
//					response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//						Command: "kubectl",
//						Args:    []string{"create", "ns", req[i].Namespace},
//					})
//					if err != nil {
//						errs = append(errs, err.Error())
//						responseObj.Error = err.Error()
//						utils.Error.Println(err)
//						return resp, err
//					}
//					utils.Info.Println(response.Stdout)
//				}
//			}
//
//			raw, err := json.Marshal(req[i])
//			_, err = agent.CreateFile(req[i].Name, string(raw))
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			}
//
//			kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//				Command: "kubectl",
//				Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
//			})
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			}
//			for {
//				feature, err := kubectlStreamResp.Recv()
//				if err == io.EOF {
//					break
//				}
//				if err != nil {
//					errs = append(errs, err.Error())
//					responseObj.Error = err.Error()
//					utils.Error.Println(err)
//					return resp, err
//				}
//				utils.Info.Println(feature.Stdout)
//			}
//
//			_, err = agent.DeleteFile(req[i].Name, string(raw))
//			if err != nil {
//				utils.Error.Println(err)
//				return resp, err
//			}
//
//			kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//				Command: "kubectl",
//				Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
//			})
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			} else {
//				responseObj.Data = kubectlResp.Stdout
//				utils.Info.Println("kubernetes statefulsets deployed successfully")
//			}
//			raw, _ = json.Marshal(responseObj)
//			utils.Info.Println("response payload", string(raw))
//			resp = append(resp, responseObj)
//		}
//	case constants.GET:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//	case constants.DELETE:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"delete", kind, name, "-n", namespace},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//	case constants.PATCH:
//
//		_, err = agent.CreateFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"apply", "-f", "~/" + name + ".json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		for {
//			feature, err := kubectlStreamResp.Recv()
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				utils.Error.Println(err)
//			}
//			utils.Info.Println(feature.Stdout)
//		}
//
//		_, err = agent.DeleteFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//
//	case constants.PUT:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"delete", kind, name, "-n", namespace},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//		}
//		utils.Info.Println(kubectlResp)
//
//		_, err = agent.CreateFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"create", "-f", "~/" + name + ".json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		for {
//			feature, err := kubectlStreamResp.Recv()
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				utils.Error.Println(err)
//			}
//			utils.Info.Println(feature.Stdout)
//		}
//
//		_, err = agent.DeleteFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//
//	}
//
//	return data, nil
//}

func (agent *AgentConnection) CreateFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.CreateFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "/tmp/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(name+".json file creation failed:", err)
		return response, err
	}
	utils.Info.Println(name, " ", response) //status : successfully created all file
	return response, err
}

func (agent *AgentConnection) InstallCertManager() {
	_ = RetryAgentConn(agent)
	//checking whether cert manager already installed or not.
	_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
		Command: "kubectl",
		Args:    []string{"get", "ns", "cert-manager"},
	})
	if err != nil {
		url := constants.KubernetesEngineURL + strings.Replace(constants.KUBERNETES_GET_CREDENTIALS_ENDPOINT, "{envId}", agent.projectId, -1)
		_, err = utils.Post(url, nil, map[string]string{
			"Content-Type":         "application/json",
			constants.AuthTokenKey: agent.cpCtx.Keys[constants.AuthTokenKey].(string),
		})
		if err != nil {
			utils.Error.Printf("Error while installing cert manager for project %s", agent.projectId)
			return
		}
	}
	return
}

func (agent *AgentConnection) DeleteFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.DeleteFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "/tmp/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(name+".json file deletion failed:", err)
		return response, err
	}
	utils.Info.Println(response) //status:"successfully deleted all files"
	return response, err
}

func GetAgentID(projectId, companyId *string) *string {
	base := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s+%s", *projectId, *companyId)))
	return &base
}

//func (agent *AgentConnection) AgentCrdManager(method constants.RequestType, data []interface{}, projectId string, companyId string, kubeType string) (resp []interface{}, err error) {
//
//	if projectId == "" || companyId == ""{
//		return
//	}
//	md := metadata.Pairs(
//		"name", *GetSha256(&projectId, &companyId),
//	)
//	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
//	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
//	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)
//
//	var req interface{}
//	switch kubeType {
//	case constants.KubernetesStatefulSets:
//		raw, err := json.Marshal(data)
//		if err != nil {
//			utils.Error.Println(err)
//			return resp, err
//		}
//		req = []v1.StatefulSet{}
//		err = json.Unmarshal(raw, &req)
//		if err != nil {
//			utils.Error.Println(err)
//			return resp, err
//		}
//
//	case constants.KubernetesService:
//
//	case constants.KubernetesConfigMaps:
//
//	case constants.KubernetesDeployment:
//
//	case constants.KubernetesPersistentVolumeClaims:
//
//	case constants.KubernetesStorageClasses:
//
//	default:
//		//for now default case is for istio and knative
//		respTemp, err = c.deployCRDS(kubeType, data)
//	}
//	switch method {
//	case constants.POST:
//		for i := range req {
//			var responseObj types.SolutionResp
//
//			if req[i].Namespace != "" {
//				_, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//					Command: "kubectl",
//					Args:    []string{"get", "ns", req[i].Namespace},
//				})
//				if err != nil {
//					response, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//						Command: "kubectl",
//						Args:    []string{"create", "ns", req[i].Namespace},
//					})
//					if err != nil {
//						errs = append(errs, err.Error())
//						responseObj.Error = err.Error()
//						utils.Error.Println(err)
//						return resp, err
//					}
//					utils.Info.Println(response.Stdout)
//				}
//			}
//
//			raw, err := json.Marshal(req[i])
//			_, err = agent.CreateFile(req[i].Name, string(raw))
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			}
//
//			kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//				Command: "kubectl",
//				Args:    []string{"create", "-f", "~/" + req[i].Name + ".json"},
//			})
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			}
//			for {
//				feature, err := kubectlStreamResp.Recv()
//				if err == io.EOF {
//					break
//				}
//				if err != nil {
//					errs = append(errs, err.Error())
//					responseObj.Error = err.Error()
//					utils.Error.Println(err)
//					return resp, err
//				}
//				utils.Info.Println(feature.Stdout)
//			}
//
//			_, err = agent.DeleteFile(req[i].Name, string(raw))
//			if err != nil {
//				utils.Error.Println(err)
//				return resp, err
//			}
//
//			kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//				Command: "kubectl",
//				Args:    []string{"get", req[i].Kind, req[i].Name, "-n", req[i].Namespace, "-o", "json"},
//			})
//			if err != nil {
//				errs = append(errs, err.Error())
//				responseObj.Error = err.Error()
//				utils.Error.Println(err)
//				return resp, err
//			} else {
//				responseObj.Data = kubectlResp.Stdout
//				utils.Info.Println("kubernetes statefulsets deployed successfully")
//			}
//			raw, _ = json.Marshal(responseObj)
//			utils.Info.Println("response payload", string(raw))
//			resp = append(resp, responseObj)
//		}
//	case constants.GET:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//	case constants.DELETE:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"delete", kind, name, "-n", namespace},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//	case constants.PATCH:
//
//		_, err = agent.CreateFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"apply", "-f", "~/" + name + ".json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		for {
//			feature, err := kubectlStreamResp.Recv()
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				utils.Error.Println(err)
//			}
//			utils.Info.Println(feature.Stdout)
//		}
//
//		_, err = agent.DeleteFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//
//	case constants.PUT:
//		kubectlResp, err := agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"delete", kind, name, "-n", namespace},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//		}
//		utils.Info.Println(kubectlResp)
//
//		_, err = agent.CreateFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlStreamResp, err := agent.agentClient.ExecKubectlStream(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"create", "-f", "~/" + name + ".json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//		for {
//			feature, err := kubectlStreamResp.Recv()
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				utils.Error.Println(err)
//			}
//			utils.Info.Println(feature.Stdout)
//		}
//
//		_, err = agent.DeleteFile(name, string(request.Service))
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		}
//
//		kubectlResp, err = agent.agentClient.ExecKubectl(agent.agentCtx, &agent_api.ExecKubectlRequest{
//			Command: "kubectl",
//			Args:    []string{"get", kind, name, "-n", namespace, "-o", "json"},
//		})
//		if err != nil {
//			utils.Error.Println(err)
//			return data, err
//		} else {
//			data, _ = json.Marshal(kubectlResp.Stdout)
//		}
//
//	}
//
//	return data, nil
//}

/*func (agent *AgentConnection) CreateFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.CreateFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "/tmp/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(name+".json file creation failed:", err)
		return response, err
	}
	utils.Info.Println(name, " ", response) //status : successfully created all file
	return response, err
}

func (agent *AgentConnection) DeleteFile(name, data string) (response *agent_api.FileResponse, err error) {
	response, err = agent.agentClient.DeleteFile(agent.agentCtx, &agent_api.CreateFileRequest{
		Name: name,
		Files: []*agent_api.File{
			{
				Name: name + ".json",
				Data: data,
				Path: "/tmp/",
			},
		},
	})
	if err != nil {
		utils.Error.Println(name+".json file deletion failed:", err)
		return response, err
	}
	utils.Info.Println(response) //status:"successfully deleted all files"
	return response, err
}

func GetAgentID(projectId, companyId *string) *string {
	base := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s+%s", *projectId, *companyId)))
	return &base
}
*/
