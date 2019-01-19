package kubernetes

import (
	"encoding/json"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type ServicesLauncher struct {
	kubeClient *kubernetes.Clientset
}

func NewServicesLauncher(c *kubernetes.Clientset) *ServicesLauncher {
	this := new(ServicesLauncher)
	this.kubeClient = c
	return this
}
func (p *ServicesLauncher) LaunchService(req *v1.Service) (*v1.Service, error) {
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	return p.kubeClient.CoreV1().Services(req.Namespace).Create(req)
}

func (p *ServicesLauncher) UpdateService(req *v1.Service) (*v1.Service, error) {
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return p.kubeClient.CoreV1().Services(req.Namespace).Patch(req.Name, kubernetesTypes.MergePatchType, r)
}
func (p *ServicesLauncher) DeleteServices(namespace string, serviceName string) error {
	return p.kubeClient.CoreV1().Services(namespace).Delete(serviceName, &metav1.DeleteOptions{})
}
func (p *ServicesLauncher) GetService(name, namespace string) (*v1.Service, error) {
	return p.kubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}
