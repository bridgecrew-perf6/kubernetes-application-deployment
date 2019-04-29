package kubernetes

import (
	"encoding/json"
	"k8s.io/api/storage/v1"
	apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
)

func (p *StatefulsetLauncher) createAWSStorageClass(serviceName, zones string, volume types.ExternalVolume) v1.StorageClass {

	objMeta := apimachinery.ObjectMeta{Name: serviceName}
	provisioner := "kubernetes.io/aws-ebs"
	// TODO add support for all io1 volume type and iops parameter
	//parameters := map[string]string{"type": volume.VolumeType, "zone": zones}
	parameters := map[string]string{"type": "gp2", "zone": zones}

	if volume.Encryption {
		parameters["encrypted"] = "true"
	}

	sClass := v1.StorageClass{
		ObjectMeta:  objMeta,
		Provisioner: provisioner,
		Parameters:  parameters,
	}
	sClass.TypeMeta.APIVersion = "storage.k8s.io/v1"
	sClass.TypeMeta.Kind = "StorageClass"

	return sClass

}
func (p *StatefulsetLauncher) createGCPStorageClass(serviceName string, volume types.ExternalVolume) v1.StorageClass {

	objMeta := apimachinery.ObjectMeta{Name: serviceName}
	provisioner := "kubernetes.io/gce-pd"
	// TODO add support for all io1 volume type and iops parameter
	//parameters := map[string]string{"type": volume.VolumeType, "zone": zones}
	parameters := map[string]string{"type": volume.VolumeType, "replication-type": "none"}

	sClass := v1.StorageClass{
		ObjectMeta:  objMeta,
		Provisioner: provisioner,
		Parameters:  parameters,
	}
	sClass.TypeMeta.APIVersion = "storage.k8s.io/v1"
	sClass.TypeMeta.Kind = "StorageClass"

	return sClass

}
func (p *StatefulsetLauncher) createAZUREStorageClass(serviceName string, volume types.ExternalVolume) v1.StorageClass {

	objMeta := apimachinery.ObjectMeta{Name: serviceName}
	provisioner := "kubernetes.io/azure-disk"
	// TODO add support for all io1 volume type and iops parameter

	parameters := map[string]string{"storageaccounttype": "Standard_LRS", "kind": "Managed"}

	sClass := v1.StorageClass{
		ObjectMeta:  objMeta,
		Provisioner: provisioner,
		Parameters:  parameters,
	}
	sClass.TypeMeta.APIVersion = "storage.k8s.io/v1"
	sClass.TypeMeta.Kind = "StorageClass"

	return sClass

}
func (p *StatefulsetLauncher) LaunchStorageClass(storageClass v1.StorageClass) (*v1.StorageClass, error) {
	utils.Info.Println("creating storage-class with name: '" + storageClass.Name + "'")
	return p.kubeClient.StorageV1().StorageClasses().Create(&storageClass)
}

func (p *StatefulsetLauncher) GetStorageClass(name string) (*v1.StorageClass, error) {
	return p.kubeClient.StorageV1().StorageClasses().Get(name, metav1.GetOptions{})
}

func (p *StatefulsetLauncher) PatchStorageClass(storageClass v1.StorageClass) (*v1.StorageClass, error) {
	r, err := json.Marshal(storageClass)
	if err != nil {
		return nil, err
	}

	return p.kubeClient.StorageV1().StorageClasses().Patch(storageClass.Name, kubernetesTypes.StrategicMergePatchType, r)
}

func (p *StatefulsetLauncher) UpdateStorageClass(storageClass v1.StorageClass) (*v1.StorageClass, error) {
	return p.kubeClient.StorageV1().StorageClasses().Update(&storageClass)
}

func (p *StatefulsetLauncher) ListStorageClass() (*v1.StorageClassList, error) {
	return p.kubeClient.StorageV1().StorageClasses().List(metav1.ListOptions{})
}

func (p *StatefulsetLauncher) DeleteStorageClass(name string) error {
	return p.kubeClient.StorageV1().StorageClasses().Delete(name, &metav1.DeleteOptions{})
}
