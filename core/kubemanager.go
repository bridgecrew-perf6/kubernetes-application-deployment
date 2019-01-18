package core

import (
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubernetes-services-deployment/constants"
	appKubernetes "kubernetes-services-deployment/core/kubernetes"
	v1alpha "kubernetes-services-deployment/kubernetes-custom-apis/core/v1"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"strings"
)

func StartServiceDeployment(req *types.ServiceRequest) error {
	if req.ClusterInfo == nil {
		return errors.New("cluster configuration not found in request")
	}
	config := rest.Config{Host: req.ClusterInfo.URL,
		Username:        req.ClusterInfo.Username,
		Password:        req.ClusterInfo.Password,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true}}
	client, err := kubernetes.NewForConfig(&config)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	for kubeType, data := range req.ServiceData {
		switch kubeType {
		case constants.KubernetesStatefulSets:
			err = DeployStatefulSets(client, data)
		case constants.KubernetesService:
			err = DeployKubernetesService(client, data)
		case constants.KubernetesConfigMaps:
			err = DeployKubernetesConfigMap(client, data)
		case constants.KubernetesDeployment:
		default:
			//for now default case is for istio and knative
			DeployCRDS(client, &config, kubeType, data)
		}
	}
	return err
}
func DeployStatefulSets(client *kubernetes.Clientset, data interface{}) error {
	var errs []string
	statefulset := appKubernetes.NewStatefulsetLauncher(client)
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
		statefulset.LaunchStatefulSet(req[i])
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}

func DeployKubernetesService(client *kubernetes.Clientset, data interface{}) error {
	var errs []string
	svc := appKubernetes.NewServicesLauncher(client)
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
		svc.LaunchService(&req[i])
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}

func DeployKubernetesConfigMap(client *kubernetes.Clientset, data interface{}) error {
	var errs []string
	svc := appKubernetes.NewConfigLauncher(client)
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
		_, err := svc.CreateConfigMap(req[i])
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}

func DeployCRDS(client *kubernetes.Clientset, config *rest.Config, key, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	var istiojsonData map[string]interface{}
	err = json.Unmarshal(raw, &istiojsonData)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	apiversion, err := findKey(istiojsonData, "apiVersion")
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	kind, err := findKey(istiojsonData, "kind")
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	groupInfo := strings.Split(apiversion, "/")
	if len(groupInfo) != 2 {
		utils.Error.Println("apiVersion " + apiversion + "is wrong")
		return errors.New("apiVersion " + apiversion + "is wrong")
	}
	rest.InClusterConfig()
	schemaDef := schema.GroupVersion{Group: groupInfo[0], Version: groupInfo[1]}
	alphaClient, err := v1alpha.NewClient(config, schemaDef)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	runtimeConfig := v1alpha.RuntimeConfig{}
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	utils.Info.Println(runtimeConfig)

	//kind to crdplural  for example kind=VirtualService and plural=virtualservices
	crdPlural := strings.ToLower(kind) + "s"
	_, err = alphaClient.NewRuntimeConfigs("", crdPlural).Create(&runtimeConfig)
	return err
}

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
