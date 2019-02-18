package kubernetes

import (
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/utils"
)

//used
type StatefulsetLauncher struct {
	kubeClient     *kubernetes.Clientset
	pod            *PodLauncher
	configMap      *ConfigMap
	kubeService    *ServicesLauncher
	ingressService *IngressControllerManager
	rbac           *RbacController
}

func NewStatefulsetLauncher(c *kubernetes.Clientset) *StatefulsetLauncher {

	this := new(StatefulsetLauncher)
	this.pod = NewPodLauncher(c)
	this.configMap = NewConfigLauncher(c)
	this.kubeService = NewServicesLauncher(c)
	this.ingressService = NewIngressLauncher(c)
	this.rbac = NewRBACLauncher(c)
	return this
}

func (p *StatefulsetLauncher) LaunchStatefulSet(req v1.StatefulSet) (set *v1.StatefulSet, err error) {
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	_, err = CreateNameSpace(p.kubeClient, req.Namespace)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}

	return p.kubeClient.AppsV1().StatefulSets(set.Namespace).Create(&req)
}

func (p *StatefulsetLauncher) PatchStatefulSets(req v1.StatefulSet) (set *v1.StatefulSet, err error) {
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return p.kubeClient.AppsV1().StatefulSets(req.Namespace).Patch(req.Name, kubernetesTypes.StrategicMergePatchType, r)

}
func (p *StatefulsetLauncher) UpdateStatefulSets(req *v1.StatefulSet) (set *v1.StatefulSet, err error) {

	return p.kubeClient.AppsV1().StatefulSets(req.Namespace).Update(req)

}
func (p *StatefulsetLauncher) GetStatefulSet(name, namespace string) (set *v1.StatefulSet, err error) {
	return p.kubeClient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
}
func (p *StatefulsetLauncher) GetAllStatefulSet(namespace string) (set *v1.StatefulSetList, err error) {
	return p.kubeClient.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
}
func (p *StatefulsetLauncher) DeleteStatefulSet(serviceName, namespace string) error {
	return p.kubeClient.AppsV1().StatefulSets(namespace).Delete(serviceName, &metav1.DeleteOptions{})
	//p.DeletePV(namespace,serviceName)
}

//used
func (p *StatefulsetLauncher) UpdateEnvVariableStatefulSet(serviceName, namespace string, env map[string]string, isHighAvailble bool, projectId string) {

}

//used
func (p *StatefulsetLauncher) UpdateCommandStatefulSet(serviceName, namespace string, command string, args []string, isHighAvailble bool) {

}

//used
func (p *StatefulsetLauncher) RollOutUpdateStatefulSet(namespace, serviceName string, imageName, imageTag string) {

}
