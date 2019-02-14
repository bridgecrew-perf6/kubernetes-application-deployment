package kubernetes

import (
	"k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/utils"
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
	return p.kubeClient.AppsV1().Deployments(req.Namespace).Create(&req)
}
func (cm *Deployments) PatchDeployments() {

}
func (p *Deployments) DeleteDeployments(name, namespace string) error {
	return p.kubeClient.AppsV1().Deployments(namespace).Delete(name, &v12.DeleteOptions{})
}
func (p *Deployments) GetDeployments(name, namespace string) (*v1.Deployment, error) {
	return p.kubeClient.AppsV1().Deployments(namespace).Get(name, v12.GetOptions{})
}
func (p *Deployments) GetAllDeployments(namespace string) (set *v1.DeploymentList, err error) {
	return p.kubeClient.AppsV1().Deployments(namespace).List(v12.ListOptions{})
}
