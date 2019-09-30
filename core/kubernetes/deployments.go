package kubernetes

import (
	"encoding/json"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/utils"
	"time"
)

type Deployments struct {
	kubeClient *kubernetes.Clientset
}

func NewDeploymentLauncher(c *kubernetes.Clientset) *Deployments {
	this := new(Deployments)
	this.kubeClient = c
	return this
}

func (p *Deployments) CreateDeployments(req v1.Deployment) (dep *v1.Deployment, err error) {
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	_, err = CreateNameSpace(p.kubeClient, req.Namespace)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}
	for dep == nil && err == nil {
		time.Sleep(1 * time.Second)
		dep, err = p.kubeClient.AppsV1().Deployments(req.Namespace).Create(&req)
	}
	return dep, err
}
func (p *Deployments) PatchDeployments(req v1.Deployment) (dep *v1.Deployment, err error) {
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	for dep == nil && err == nil {
		time.Sleep(1 * time.Second)
		dep, err = p.kubeClient.AppsV1().Deployments(req.Namespace).Patch(req.Name, kubernetesTypes.StrategicMergePatchType, r)
	}
	return dep, err
}
func (p *Deployments) UpdateDeployments(req *v1.Deployment) (*v1.Deployment, error) {

	return p.kubeClient.AppsV1().Deployments(req.Namespace).Update(req)
}
func (p *Deployments) DeleteDeployments(name, namespace string) error {

	return p.kubeClient.AppsV1().Deployments(namespace).Delete(name, &v12.DeleteOptions{})
}
func (p *Deployments) GetDeployments(name, namespace string) (dep *v1.Deployment, err error) {
	for dep == nil && err == nil {
		time.Sleep(1 * time.Second)
		dep, err = p.kubeClient.AppsV1().Deployments(namespace).Get(name, v12.GetOptions{})
	}
	return dep, err
}
func (p *Deployments) GetAllDeployments(namespace string) (set *v1.DeploymentList, err error) {
	return p.kubeClient.AppsV1().Deployments(namespace).List(v12.ListOptions{})
}
