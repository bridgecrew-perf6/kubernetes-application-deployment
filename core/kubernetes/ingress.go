package kubernetes

import "k8s.io/client-go/kubernetes"

type IngressControllerManager struct {
	kubeClient *kubernetes.Clientset
}

func NewIngressLauncher(c *kubernetes.Clientset) *IngressControllerManager {
	this := new(IngressControllerManager)
	this.kubeClient = c
	return this
}

/*func (ic *IngressControllerManager) LaunchIngress(service types.Service) {



}

func (ic *IngressControllerManager) PatchIngress(serviceName, nameSpace string, ingressRules []map[string]string) string {


}

func (ic *IngressControllerManager) CreateIngressRule(service types.Service) k8ext.Ingress {


}

*/
