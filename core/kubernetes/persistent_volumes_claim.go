package kubernetes

import (
	"encoding/json"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"kubernetes-services-deployment/utils"
)

func (p *StatefulsetLauncher) CreatePersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (*v12.PersistentVolumeClaim, error) {
	if volumeClaim.Namespace == "" {
		volumeClaim.Namespace = "default"
	}

	_, err := CreateNameSpace(p.kubeClient, volumeClaim.Namespace)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}

	utils.Info.Println("creating pvc with name: '" + volumeClaim.Name + "' in namespace: '" + volumeClaim.Namespace + "'")
	return p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Create(&volumeClaim)
}

func (p *StatefulsetLauncher) GetPersistentVolumeClaim(name, namespace string) (*v12.PersistentVolumeClaim, error) {
	return p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

func (p *StatefulsetLauncher) PatchPersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (*v12.PersistentVolumeClaim, error) {
	r, err := json.Marshal(volumeClaim)
	if err != nil {
		return nil, err
	}

	return p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Patch(volumeClaim.Name, kubernetesTypes.StrategicMergePatchType, r)
}

func (p *StatefulsetLauncher) UpdatePersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (*v12.PersistentVolumeClaim, error) {
	return p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Update(&volumeClaim)
}

func (p *StatefulsetLauncher) ListPersistentVolumeClaim(namespace string) (*v12.PersistentVolumeClaimList, error) {
	return p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
}

func (p *StatefulsetLauncher) DeletePersistentVolumeClaim(name, namespace string) error {
	return p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}
