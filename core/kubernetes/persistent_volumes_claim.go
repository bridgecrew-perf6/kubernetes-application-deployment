package kubernetes

import (
	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/utils"
	"encoding/json"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"time"
)

func (p *StatefulsetLauncher) CreatePersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (pvc *v12.PersistentVolumeClaim, err error) {
	if volumeClaim.Namespace == "" {
		volumeClaim.Namespace = "default"
	}

	_, err = CreateNameSpace(p.kubeClient, volumeClaim.Namespace)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}

	utils.Info.Println("creating pvc with name: '" + volumeClaim.Name + "' in namespace: '" + volumeClaim.Namespace + "'")
	pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Create(&volumeClaim)
	for pvc == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Create(&volumeClaim)
		} else {
			break
		}
	}
	return pvc, err
}

func (p *StatefulsetLauncher) GetPersistentVolumeClaim(name, namespace string) (pvc *v12.PersistentVolumeClaim, err error) {
	pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
	for pvc == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		} else {
			break
		}
	}
	return pvc, err
}

func (p *StatefulsetLauncher) PatchPersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (pvc *v12.PersistentVolumeClaim, err error) {
	r, err := json.Marshal(volumeClaim)
	if err != nil {
		return nil, err
	}
	pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Patch(volumeClaim.Name, kubernetesTypes.StrategicMergePatchType, r)
	for pvc == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Patch(volumeClaim.Name, kubernetesTypes.StrategicMergePatchType, r)
		} else {
			break
		}
	}
	return pvc, err
}

func (p *StatefulsetLauncher) UpdatePersistentVolumeClaim(volumeClaim v12.PersistentVolumeClaim) (*v12.PersistentVolumeClaim, error) {

	pvc, err := p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Update(&volumeClaim)
	for pvc == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			pvc, err = p.kubeClient.CoreV1().PersistentVolumeClaims(volumeClaim.Namespace).Update(&volumeClaim)
		} else {
			break
		}
	}
	return pvc, err
}

func (p *StatefulsetLauncher) ListPersistentVolumeClaim(namespace string) (*v12.PersistentVolumeClaimList, error) {

	set, err := p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	for set == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			set, err = p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
		} else {
			break
		}
	}
	return set, err

}

func (p *StatefulsetLauncher) DeletePersistentVolumeClaim(name, namespace string) error {

	err := p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
	for err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			err = p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
		} else {
			break
		}
	}
	return err
}
