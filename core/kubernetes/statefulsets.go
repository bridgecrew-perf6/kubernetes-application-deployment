package kubernetes

import (
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/utils"
	"time"
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
	this.kubeClient = c
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
	set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Create(&req)
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Create(&req)
		} else {
			break
		}
	}
	return set, err

}

func (p *StatefulsetLauncher) PatchStatefulSets(req v1.StatefulSet) (set *v1.StatefulSet, err error) {
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Patch(req.Name, kubernetesTypes.StrategicMergePatchType, r)
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Patch(req.Name, kubernetesTypes.StrategicMergePatchType, r)
		} else {
			break
		}
	}
	return set, err
}
func (p *StatefulsetLauncher) UpdateStatefulSets(req *v1.StatefulSet) (set *v1.StatefulSet, err error) {

	set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Update(req)

	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.AppsV1().StatefulSets(req.Namespace).Update(req)
		} else {
			break
		}
	}
	return set, err
}
func (p *StatefulsetLauncher) GetStatefulSet(name, namespace string) (set *v1.StatefulSet, err error) {
	set, err = p.kubeClient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
		} else {
			break
		}
	}
	return set, err
}
func (p *StatefulsetLauncher) GetAllStatefulSet(namespace string) (set *v1.StatefulSetList, err error) {
	set, err = p.kubeClient.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
		} else {
			break
		}
	}
	return set, err
}
func (p *StatefulsetLauncher) DeleteStatefulSet(serviceName, namespace string) error {

	err := p.kubeClient.AppsV1().StatefulSets(namespace).Delete(serviceName, &metav1.DeleteOptions{})
	for err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			err = p.kubeClient.AppsV1().StatefulSets(namespace).Delete(serviceName, &metav1.DeleteOptions{})
		} else {
			break
		}
	}
	return err
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
