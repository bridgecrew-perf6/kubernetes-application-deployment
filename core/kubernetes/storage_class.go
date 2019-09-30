package kubernetes

import (
	"encoding/json"
	"k8s.io/api/storage/v1"
	apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/types"
	"kubernetes-services-deployment/utils"
	"time"
)

type StorageClass struct {
	kubeClient *kubernetes.Clientset
}

func NewStorageLauncher(c *kubernetes.Clientset) *StorageClass {
	this := new(StorageClass)
	this.kubeClient = c
	return this
}
func (p *StorageClass) createAWSStorageClass(serviceName, zones string, volume types.ExternalVolume) v1.StorageClass {

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
func (p *StorageClass) createGCPStorageClass(serviceName string, volume types.ExternalVolume) v1.StorageClass {

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
func (p *StorageClass) createAZUREStorageClass(serviceName string, volume types.ExternalVolume) v1.StorageClass {

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
func (p *StorageClass) LaunchStorageClass(storageClass v1.StorageClass) (ss *v1.StorageClass, err error) {
	utils.Info.Println("creating storage-class with name: '" + storageClass.Name + "'")
	for ss == nil && err == nil {
		time.Sleep(1 * time.Second)
		ss, err = p.kubeClient.StorageV1().StorageClasses().Create(&storageClass)
	}
	return ss, err
}

func (p *StorageClass) GetStorageClass(name string) (ss *v1.StorageClass, err error) {
	for ss == nil && err == nil {
		time.Sleep(1 * time.Second)
		ss, err = p.kubeClient.StorageV1().StorageClasses().Get(name, metav1.GetOptions{})
	}
	return ss, err
}

func (p *StorageClass) PatchStorageClass(storageClass v1.StorageClass) (ss *v1.StorageClass, err error) {
	r, err := json.Marshal(storageClass)
	if err != nil {
		return nil, err
	}
	for ss == nil && err == nil {
		time.Sleep(1 * time.Second)
		ss, err = p.kubeClient.StorageV1().StorageClasses().Patch(storageClass.Name, kubernetesTypes.StrategicMergePatchType, r)
	}
	return ss, err
}

func (p *StorageClass) UpdateStorageClass(storageClass v1.StorageClass) (*v1.StorageClass, error) {
	return p.kubeClient.StorageV1().StorageClasses().Update(&storageClass)
}

func (p *StorageClass) ListStorageClass() (*v1.StorageClassList, error) {
	return p.kubeClient.StorageV1().StorageClasses().List(metav1.ListOptions{})
}

func (p *StorageClass) DeleteStorageClass(name string) error {
	return p.kubeClient.StorageV1().StorageClasses().Delete(name, &metav1.DeleteOptions{})
}
