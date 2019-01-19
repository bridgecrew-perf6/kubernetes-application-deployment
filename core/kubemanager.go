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
	var errs []string
	for kubeType, data := range req.ServiceData {
		utils.Info.Println(len(data))
		if len(data) == 0 {
			continue
		}
		switch kubeType {
		case constants.KubernetesStatefulSets:
			err = DeployStatefulSets(client, data)
		case constants.KubernetesService:
			err = DeployKubernetesService(client, data)
		case constants.KubernetesConfigMaps:
			err = DeployKubernetesConfigMap(client, data)
		case constants.KubernetesDeployment:
			err = DeployKubernetesDeployment(client, data)
		default:
			//for now default case is for istio and knative
			utils.Info.Println(kubeType)
			err = DeployCRDS(client, &config, kubeType, data)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ";")
		return errors.New(finalErr)

	}
	return nil
}
func DeployStatefulSets(client *kubernetes.Clientset, data []interface{}) error {
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		_, err = statefulset.LaunchStatefulSet(req[i])
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("kubernetes statefulsets deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes statefulsets deployed successfully")
		}
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
}

func DeployKubernetesService(client *kubernetes.Clientset, data []interface{}) error {
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		_, err = svc.LaunchService(&req[i])
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

func DeployKubernetesConfigMap(client *kubernetes.Clientset, data []interface{}) error {
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
		raw, _ := json.Marshal(req[i])
		utils.Info.Println(string(raw))
		_, err := svc.CreateConfigMap(req[i])
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

func DeployKubernetesDeployment(client *kubernetes.Clientset, data []interface{}) error {
	var errs []string
	depObj := appKubernetes.NewDeploymentLauncher(client)
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
		_, err = depObj.CreateDeployments(req[i])
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

func DeployCRDS(client *kubernetes.Clientset, config *rest.Config, key string, data []interface{}) error {
	var errs []string
	raw, err := json.Marshal(data)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	runtimeConfig := []v1alpha.RuntimeConfig{}
	err = json.Unmarshal(raw, &runtimeConfig)
	if err != nil {
		utils.Error.Println(err)
		return err
	}
	utils.Info.Println(len(runtimeConfig))
	for i := range runtimeConfig {
		raw, _ := json.Marshal(runtimeConfig[i])
		utils.Info.Println(string(raw))
		groupInfo := strings.Split(runtimeConfig[i].APIVersion, "/")
		if len(groupInfo) != 2 {
			utils.Error.Println("apiVersion " + runtimeConfig[i].APIVersion + "is wrong")
			errs = append(errs, "apiVersion "+runtimeConfig[i].APIVersion+"is wrong")
			continue
		}
		rest.InClusterConfig()
		schemaDef := schema.GroupVersion{Group: groupInfo[0], Version: groupInfo[1]}
		alphaClient, err := v1alpha.NewClient(config, schemaDef)
		if err != nil {
			utils.Error.Println(err)
			errs = append(errs, err.Error())
			continue
		}

		//kind to crdplural  for example kind=VirtualService and plural=virtualservices
		crdPlural := utils.Pluralize(strings.ToLower(runtimeConfig[i].Kind))

		namespace := ""
		if runtimeConfig[i].Namespace == "" {
			namespace = "default"
		} else {
			namespace = runtimeConfig[i].Namespace
		}
		utils.Info.Println(crdPlural, namespace)

		data, err := alphaClient.NewRuntimeConfigs(namespace, crdPlural).Create(raw)
		if err != nil {
			errs = append(errs, err.Error())
			utils.Error.Println("kubernetes crd deployed failed. Error: ", err)
		} else {
			utils.Info.Println("kubernetes crd deployed successfully")
		}
		dd, _ := json.Marshal(data)
		utils.Info.Println(string(dd))
	}
	if len(errs) >= 1 {
		finalErr := strings.Join(errs, ",")
		return errors.New(finalErr)
	}
	return nil
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
