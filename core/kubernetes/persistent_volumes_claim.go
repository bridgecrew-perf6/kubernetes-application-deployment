package kubernetes

import (
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"strconv"
)

func (p *StatefulsetLauncher) CreatePersistentVolumeClaim(serviceName, namespace string, volume types.ExternalVolume) (*v12.PersistentVolumeClaim, error) {
	objMeta := v1.ObjectMeta{
		Name: serviceName,
		Annotations: map[string]string{
			"volume.beta.kubernetes.io/storage-class": serviceName,
			"DeleteOnTerm": strconv.FormatBool(volume.Delete_on_termination),
		},
	}
	utils.Info.Println(objMeta)

	size, _ := resource.ParseQuantity(strconv.Itoa(volume.VolumeSize) + "Gi")

	storageClassName := serviceName

	volumeClaim := v12.PersistentVolumeClaim{
		ObjectMeta: objMeta,
		Spec: v12.PersistentVolumeClaimSpec{
			AccessModes:      []v12.PersistentVolumeAccessMode{v12.ReadWriteOnce},
			StorageClassName: &storageClassName,
			Resources:        v12.ResourceRequirements{Requests: v12.ResourceList{v12.ResourceStorage: size}},
		},
	}
	return p.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(&volumeClaim)
}

func (p *StatefulsetLauncher) CreatePersistentVolumeClaimExisting(serviceName string, volume types.ExternalVolume) {

}

//used
func (p *StatefulsetLauncher) DeletePVC(namespace, serviceName string) {

}
