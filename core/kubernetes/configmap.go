package kubernetes

import (
	"encoding/json"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"time"
)

type ConfigMap struct {
	kubeClient *kubernetes.Clientset
}

func NewConfigLauncher(c *kubernetes.Clientset) *ConfigMap {
	this := new(ConfigMap)
	this.kubeClient = c
	return this
}
func (cm *ConfigMap) LaunchSideCarConfigMap() {

}
func (cm *ConfigMap) CreateConfigMap(configMap v1.ConfigMap) (cmp *v1.ConfigMap, err error) {
	cmp, err = cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(&configMap)
	for cmp == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			cmp, err = cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(&configMap)
		} else {
			break
		}
	}
	return cmp, err
}
func (cm *ConfigMap) PatchConfigMap(configMap v1.ConfigMap) (cmp *v1.ConfigMap, err error) {
	r, err := json.Marshal(configMap)
	if err != nil {
		return nil, err
	}
	cmp, err = cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Patch(configMap.Name, kubernetesTypes.StrategicMergePatchType, r)
	for cmp == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			cmp, err = cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Patch(configMap.Name, kubernetesTypes.StrategicMergePatchType, r)
		} else {
			break
		}
	}
	return
}
func (cm *ConfigMap) UpdateConfigMap(configMap *v1.ConfigMap) (*v1.ConfigMap, error) {

	cmp, err := cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Update(configMap)
	for cmp == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			cmp, err = cm.kubeClient.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Update(configMap)
		} else {
			break
		}
	}
	return cmp, err
}
func (cm *ConfigMap) DeleteConfigMap(name, namespace string) error {

	err := cm.kubeClient.CoreV1().ConfigMaps(namespace).Delete(name, &v12.DeleteOptions{})
	for err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			err = cm.kubeClient.CoreV1().ConfigMaps(namespace).Delete(name, &v12.DeleteOptions{})
		} else {
			break
		}
	}
	return err
}
func (cm *ConfigMap) GetConfigMap(name, namespace string) (cmp *v1.ConfigMap, err error) {
	cmp, err = cm.kubeClient.CoreV1().ConfigMaps(namespace).Get(name, v12.GetOptions{})
	for cmp == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			cmp, err = cm.kubeClient.CoreV1().ConfigMaps(namespace).Get(name, v12.GetOptions{})
		} else {
			break
		}
	}
	return cmp, err
}
func (cm *ConfigMap) GetAllConfigMap(namespace string) (*v1.ConfigMapList, error) {

	set, err := cm.kubeClient.CoreV1().ConfigMaps(namespace).List(v12.ListOptions{})
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = cm.kubeClient.CoreV1().ConfigMaps(namespace).List(v12.ListOptions{})
		} else {
			break
		}
	}
	return set, err
}
