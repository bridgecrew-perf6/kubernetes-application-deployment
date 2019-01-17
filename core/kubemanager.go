package core

import (
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubernetes-services-deployment/constants"
	appKubernetes "kubernetes-services-deployment/core/kubernetes"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"strings"
)

func StartServiceDeployment(req *types.ServiceRequest) error {
	if req.ClusterInfo == nil {
		return errors.New("cluster configuration not found in request")
	}
	client, err := kubernetes.NewForConfig(
		&rest.Config{Host: req.ClusterInfo.URL,
			Username:        req.ClusterInfo.Username,
			Password:        req.ClusterInfo.Password,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true}})
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

func DeployCustomResourceDefinition(client *kubernetes.Clientset, key, data interface{}) {

}
