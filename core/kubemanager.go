package core

import (
	"github.com/gedex/inflector"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kubernetes-services-deployment/constants"
	appKubernetes "kubernetes-services-deployment/core/kubernetes"
	v1alpha "kubernetes-services-deployment/kubernetes-custom-apis/core/v1"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"strings"
	"time"
)

type KubernetesClient struct {
	Config     *rest.Config
	Client     *kubernetes.Clientset
	Namespaces map[string]bool
	context    *Context
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
	data, ok := constants.CacheObj.Get(*projectId)
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
		data := types.CacheObjectData{
			ProjectId:                 *projectId,
			KubernetesClusterMasterIp: kubernetesClusterIp,
			KubernetesCredentials:     credentials,
		}
		constants.CacheObj.Set(*projectId, data, cache.DefaultExpiration)
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
		return responses, errors.New("invalid request while starting deployment")
	}
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return responses, err
	}
	var errs []string
	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		switch kubeType {
		case constants.KubernetesStatefulSets:
			respTemp, err = c.deployStatefulSets(data)
		case constants.KubernetesService:
			respTemp, err = c.deployKubernetesService(data)
		case constants.KubernetesConfigMaps:
			respTemp, err = c.deployKubernetesConfigMap(data)
		case constants.KubernetesDeployment:
			respTemp, err = c.deployKubernetesDeployment(data)
		case constants.KubernetesPersistentVolumeClaims:
			respTemp, err = c.deployKubernetesPVC(data)
		case constants.KubernetesStorageClasses:
			respTemp, err = c.deployKubernetesStorageClasses(data)
		default:
			//for now default case is for istio and knative
			respTemp, err = c.deployCRDS(kubeType, data)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
		responses[kubeType] = respTemp
	}
	r, _ := json.Marshal(responses)
	cpContext.SendBackendLogs(string(r), constants.LOGGING_LEVEL_DEBUG)
	utils.Info.Println(string(r))
	return responses, nil
}
func GetServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return responses, err
	}
	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		switch kubeType {
		case constants.KubernetesStatefulSets:
			respTemp, err = c.getStatefulSets(data)
		case constants.KubernetesService:
			respTemp, err = c.getKubernetesService(data)
		case constants.KubernetesConfigMaps:
			respTemp, err = c.getKubernetesConfigMap(data)
		case constants.KubernetesDeployment:
			respTemp, err = c.getKubernetesDeployment(data)
		case constants.KubernetesPersistentVolumeClaims:
			respTemp, err = c.getKubernetesPVC(data)
		case constants.KubernetesStorageClasses:
			respTemp, err = c.getKubernetesStorageClass(data)
		default:
			//for now default case is for istio and knative
			respTemp, err = c.getCRDS(kubeType, data)

		}
		if err != nil {
			errs = append(errs, err.Error())
		}
		responses[kubeType] = respTemp

	}
	r, _ := json.Marshal(responses)
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	utils.Info.Println(string(r))
	return responses, nil
}
func ListServiceDeployment(cpContext *Context, req *types.ServiceRequest) (responses map[string]interface{}, err error) {
	responses = make(map[string]interface{})
	if req == nil {
		return responses, errors.New("invalid request while starting deployment")
	}
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return responses, err
	}
	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {
		var respTemp interface{}
		if len(data) == 0 {
			continue
		}
		//for now default case is for istio and knative
		respTemp, err = c.listCRDS(kubeType, data)
		if err != nil {
			errs = append(errs, err.Error())
		}
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
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
		return responses, err
	}
	cpContext.SendBackendLogs(req.ServiceData, constants.LOGGING_LEVEL_DEBUG)
	var errs []string
	for kubeType, data := range req.ServiceData {

		utils.Info.Println(len(data))
		if len(data) == 0 {
			continue
		}
		switch kubeType {
		case constants.KubernetesStatefulSets:
			err = c.deleteStatefulSets(data)
		case constants.KubernetesService:
			err = c.deleteKubernetesService(data)
		case constants.KubernetesConfigMaps:
			err = c.deleteKubernetesConfigMap(data)
		case constants.KubernetesDeployment:
			err = c.deleteKubernetesDeployment(data)
		default:
			//for now default case is for istio and knative
			utils.Info.Println(kubeType)
			err = c.deleteCRDS(kubeType, data)

		}
		if err != nil {
			errs = append(errs, err.Error())
		}
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
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
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
		switch kubeType {
		case constants.KubernetesStatefulSets:
			respTemp, err = c.patchStatefulSets(data)
		case constants.KubernetesService:
			respTemp, err = c.patchKubernetesService(data)
		case constants.KubernetesConfigMaps:
			respTemp, err = c.patchKubernetesConfigMap(data)
		case constants.KubernetesDeployment:
			respTemp, err = c.patchKubernetesDeployment(data)
		default:
			//for now default case is for istio and knative
			utils.Info.Println(kubeType)
			respTemp, err = c.patchCRDS(kubeType, data)

		}
		if err != nil {
			errs = append(errs, err.Error())
		}
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
	c, err := GetKubernetesClient(cpContext, req.ProjectId)
	if err != nil {
		utils.Error.Println(err)
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
		switch kubeType {
		case constants.KubernetesStatefulSets:
			respTemp, err = c.putStatefulSets(data)
		case constants.KubernetesService:
			respTemp, err = c.putKubernetesService(data)
		case constants.KubernetesConfigMaps:
			respTemp, err = c.putKubernetesConfigMap(data)
		case constants.KubernetesDeployment:
			respTemp, err = c.putKubernetesDeployment(data)
		default:
			//for now default case is for istio and knative
			utils.Info.Println(kubeType)
			respTemp, err = c.putCRDS(kubeType, data)

		}
		if err != nil {
			errs = append(errs, err.Error())
		} else {

			responses[kubeType] = respTemp
		}
	}
	cpContext.SendBackendLogs(responses, constants.LOGGING_LEVEL_DEBUG)
	return responses, nil
}

func (c *KubernetesClient) deployStatefulSets(data []interface{}) (resp []interface{}, err error) {
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
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
		} else {
			tempResp, err := statefulset.LaunchStatefulSet(req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes statefulsets deployed failed. Error: ", err)
			} else {

				responseObj.Data = tempResp
				utils.Info.Println("kubernetes statefulsets deployed successfully")
			}
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
func (c *KubernetesClient) deployKubernetesService(data []interface{}) (resp []interface{}, err error) {
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
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
		} else {
			tempResp, err := svc.LaunchService(&req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes service deployed failed. Error: ", err)
			} else {
				responseObj.Data = tempResp
				utils.Info.Println("kubernetes service deployed successfully")
			}
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
func (c *KubernetesClient) deployKubernetesConfigMap(data []interface{}) (resp []interface{}, err error) {
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
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
		} else {
			tempResp, err := svc.CreateConfigMap(req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes configmap deployed failed. Error: ", err)
			} else {
				responseObj.Data = tempResp
				utils.Info.Println("kubernetes configmap deployed successfully")
			}
		}
		resp = append(resp, responseObj)
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}
func (c *KubernetesClient) deployKubernetesDeployment(data []interface{}) (resp []interface{}, err error) {
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
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
		} else {
			tempResp, err := depObj.CreateDeployments(req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes deployment deployed failed. Error: ", err)
			} else {
				responseObj.Data = tempResp
				utils.Info.Println("kubernetes deployment deployed successfully")
			}
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
func (c *KubernetesClient) deployKubernetesPVC(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewStatefulsetLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
		} else {
			tempResp, err := depObj.CreatePersistentVolumeClaim(req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes pvc deployment failed. Error: ", err)
			} else {
				responseObj.Data = tempResp
				utils.Info.Println("kubernetes pvc deployed successfully")
			}
		}
		resp = append(resp, responseObj)

	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}

	return resp, nil
}
func (c *KubernetesClient) deployKubernetesStorageClasses(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewStorageLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println("request payload", string(raw))
		c.Namespaces[req[i].Namespace] = true
		_, err := appKubernetes.CreateNameSpace(c.Client, req[i].Namespace)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
		} else {
			tempResp, err := depObj.LaunchStorageClass(req[i])
			if err != nil {
				errs = append(errs, err.Error())
				responseObj.Error = err.Error()
				utils.Error.Println("kubernetes storage class deployment failed. Error: ", err)
			} else {
				responseObj.Data = tempResp
				utils.Info.Println("kubernetes storage class deployed successfully")
			}
		}
		resp = append(resp, responseObj)

	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}

	return resp, nil
}
func (c *KubernetesClient) deployCRDS(key string, data []interface{}) (resp []interface{}, err error) {
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
		responseObj, _ := c.crdManager(runtimeConfig[i], "post")
		resp = append(resp, responseObj)

	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return resp, errors.New(finalErr)
	}
	return resp, nil
}

func (c *KubernetesClient) getStatefulSets(data []interface{}) (resp []interface{}, err error) {
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
		respTemp, err := statefulset.GetStatefulSet(req[i].Name, req[i].Namespace)
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
func (c *KubernetesClient) getKubernetesService(data []interface{}) (resp []interface{}, err error) {
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
		respTemp, err := svc.GetService(req[i].Name, req[i].Namespace)
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
func (c *KubernetesClient) getKubernetesConfigMap(data []interface{}) (resp []interface{}, err error) {
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
		respTemp, err := svc.GetConfigMap(req[i].Name, req[i].Namespace)
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
func (c *KubernetesClient) getKubernetesDeployment(data []interface{}) (resp []interface{}, err error) {
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
		respTemp, err := depObj.GetDeployments(req[i].Name, req[i].Namespace)
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
}
func (c *KubernetesClient) getKubernetesPVC(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewStatefulsetLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := depObj.GetPersistentVolumeClaim(req[i].Name, req[i].Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes pvc deployment failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes pvc deployed successfully")
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
func (c *KubernetesClient) getKubernetesStorageClass(data []interface{}) (resp []interface{}, err error) {
	var errs []string
	depObj := appKubernetes.NewStorageLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		respTemp, err := depObj.GetStorageClass(req[i].Name)
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes storage-class deployment failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes storage-class deployed successfully")
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
func (c *KubernetesClient) getCRDS(key string, data []interface{}) (resp []interface{}, err error) {
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
		rest.InClusterConfig()
		responseObj, _ := c.crdManager(runtimeConfig[i], "get")

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
func (c *KubernetesClient) listCRDS(key string, data []interface{}) (resp []interface{}, err error) {
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
		rest.InClusterConfig()
		responseObj, _ := c.crdManager(runtimeConfig[i], "list")

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
func (c *KubernetesClient) getCRDClient(apiVersion string) (*v1alpha.RuntimeConfigV1Alpha1Client, error) {
	groupInfo := strings.Split(apiVersion, "/")
	if len(groupInfo) == 0 {
		utils.Error.Println("apiVersion " + apiVersion + " is wrong")
		return nil, errors.New("apiVersion " + apiVersion + " is wrong")

	}
	groupName := ""
	groupVersion := ""
	apiPath := ""
	if len(groupInfo) == 1 {
		groupName = ""
		groupVersion = groupInfo[0]
		apiPath = "/api"
	} else {
		groupName = groupInfo[0]
		groupVersion = groupInfo[1]
		apiPath = "/apis"
	}
	schemaDef := schema.GroupVersion{Group: groupName, Version: groupVersion}
	alphaClient, err := v1alpha.NewClient(c.Config, schemaDef, apiPath)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}
	return alphaClient, nil
}

func (c *KubernetesClient) deleteStatefulSets(data []interface{}) error {
	var errs []string
	statefulset := appKubernetes.NewStatefulsetLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		err = statefulset.DeleteStatefulSet(req[i].Name, req[i].Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			responseObj.Error = err.Error()
			utils.Error.Println("kubernetes statefulsets deletion failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes statefulsets deleted successfully")
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (c *KubernetesClient) deleteKubernetesService(data []interface{}) error {
	var errs []string
	svc := appKubernetes.NewServicesLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		err = svc.DeleteServices(req[i].Name, req[i].Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("kubernetes service deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes service deployed successfully")
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (c *KubernetesClient) deleteKubernetesConfigMap(data []interface{}) error {
	var errs []string
	svc := appKubernetes.NewConfigLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		err := svc.DeleteConfigMap(req[i].Name, req[i].Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("kubernetes configmap deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes configmap deployed successfully")
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}
func (c *KubernetesClient) deleteKubernetesDeployment(data []interface{}) error {
	var errs []string
	depObj := appKubernetes.NewDeploymentLauncher(c.Client)
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		err = depObj.DeleteDeployments(req[i].Name, req[i].Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("kubernetes deployment deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes deployment deployed successfully")
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}

	return nil
}
func (c *KubernetesClient) deleteCRDS(key string, data []interface{}) (err error) {
	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	var runtimeConfig []v1alpha.RuntimeConfig
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {
		rest.InClusterConfig()
		res, _ := c.crdManager(runtimeConfig[i], "delete")
		/*//kind to crdplural  for example kind=VirtualService and plural=virtualservices
		crdPlural := utils.Pluralize(strings.ToLower(runtimeConfig[i].Kind))
		namespace := ""
		if runtimeConfig[i].Namespace == "" {
			namespace = "default"
		} else {
			namespace = runtimeConfig[i].Namespace
		}
		alphaClient, err := c.getCRDClient(runtimeConfig[i].APIVersion)
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("failed to fetch data. Error: ", err)
		} else {
			err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Delete(runtimeConfig[i].Name, &v13.DeleteOptions{})
			if err != nil {
				errs = append(errs, err.Error())
				utils.Error.Println("failed to fetch data. Error: ", err)
			}
		}*/
		if res.Error != "" {
			errs = append(errs, res.Error)
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		err = errors.New(finalErr)
	}
	return err
}

func (c *KubernetesClient) patchStatefulSets(data []interface{}) (resp []interface{}, err error) {
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
		respTemp, err := statefulset.PatchStatefulSets(req[i])
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
func (c *KubernetesClient) patchKubernetesService(data []interface{}) (resp []interface{}, err error) {
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
func (c *KubernetesClient) patchKubernetesConfigMap(data []interface{}) (resp []interface{}, err error) {
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
func (c *KubernetesClient) patchKubernetesDeployment(data []interface{}) (resp []interface{}, err error) {
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
}
func (c *KubernetesClient) patchCRDS(key string, data []interface{}) (resp []interface{}, err error) {
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
		responseObj, _ := c.crdManager(runtimeConfig[i], "patch")
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
		data, err := alphaClient.NewRuntimeConfigs(namespace, crdPlural).Patch(runtimeConfig[i].Name, kubernetesTypes.MergePatchType, raw)
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

func (c *KubernetesClient) putStatefulSets(data []interface{}) (resp []interface{}, err error) {
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
}
func (c *KubernetesClient) putCRDS(key string, data []interface{}) (resp []interface{}, err error) {
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
		responseObj, _ := c.crdManager(runtimeConfig[i], "put")
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
func (c *KubernetesClient) GetKubernetesServiceExternalIp(namespace, name string) (string, error) {
	serviceObj := appKubernetes.NewServicesLauncher(c.Client)
	svc, err := serviceObj.GetService(name, namespace)
	if err != nil {
		utils.Error.Println(err)
		return "", err
	}
	externalIp := ""
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP == "" {
			externalIp = ingress.Hostname
		} else {
			externalIp = ingress.IP
		}
	}
	return externalIp, nil
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

func (c *KubernetesClient) crdManager(runtimeConfig interface{}, method string) (responseObj types.SolutionResp, err error) {

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
	rest.InClusterConfig()
	if runtimeObj.Kind == "" || runtimeObj.APIVersion == "" {
		utils.Error.Println("Kind/APIVersion is empty")
		responseObj.Error = "Kind/APIVersion is empty"
		return responseObj, errors.New("Kind/APIVersion is empty")
	}
	//kind to crdplural  for example kind=VirtualService and plural=virtualservices
	crdPlural := inflector.Pluralize(strings.ToLower(runtimeObj.Kind))

	namespace := runtimeObj.Namespace

	utils.Info.Println(crdPlural, namespace)

	c.Namespaces[namespace] = true
	if namespace != "" {
		_, err = appKubernetes.CreateNameSpace(c.Client, namespace)
		if err != nil && !errors2.IsAlreadyExists(err) {
			utils.Error.Println(err)
			responseObj.Error = err.Error()
			return responseObj, err
		}
	}
	alphaClient, err := c.getCRDClient(runtimeObj.APIVersion)
	if err != nil {
		responseObj.Error = err.Error()
		utils.Error.Println("kubernetes crd deployed failed. Error: ", err)
		return responseObj, err
	} else {
		var data interface{}
		var err error
		switch method {
		case "post":
			for data == nil && err == nil {
				time.Sleep(1 * time.Second)
				data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Create(raw)
			}
		case "get":
			for data == nil && err == nil {
				time.Sleep(1 * time.Second)
				data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Get(runtimeConfig.(v1alpha.RuntimeConfig).Name)
			}
		case "put":
			for data == nil && err == nil {
				time.Sleep(1 * time.Second)
				data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Update(runtimeConfig)
			}
		case "patch":
			for data == nil && err == nil {
				time.Sleep(1 * time.Second)
				data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Patch(runtimeConfig.(v1alpha.RuntimeConfig).Name, kubernetesTypes.MergePatchType, raw)
			}
		case "delete":
			err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).Delete(runtimeConfig.(v1alpha.RuntimeConfig).Name, &v13.DeleteOptions{})
		case "list":
			for data == nil && err == nil {
				time.Sleep(1 * time.Second)
				data, err = alphaClient.NewRuntimeConfigs(namespace, crdPlural).List(v13.ListOptions{})
			}
		}
		if err != nil {
			responseObj.Error = err.Error()
			return responseObj, err
		} else {
			responseObj.Data = data

			dd, _ := json.Marshal(data)
			utils.Info.Println(string(dd))
		}
	}

	raw, _ = json.Marshal(responseObj)
	utils.Info.Println("response payload", string(raw))
	return responseObj, nil
}
