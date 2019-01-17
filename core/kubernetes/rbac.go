package kubernetes

import "k8s.io/client-go/kubernetes"

type RbacController struct {
	kubeClient *kubernetes.Clientset
}

func NewRBACLauncher(c *kubernetes.Clientset) *RbacController {
	this := new(RbacController)
	this.kubeClient = c
	return this
}

/*func (cm *RbacController) LaunchSideCarServiceAccount(serv types.Service) {


}

func (cm *RbacController) LaunchSideCarClusterRoleBinding(serv types.Service) {


}*/
