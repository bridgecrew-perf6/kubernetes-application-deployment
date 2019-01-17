package kubernetes

import (
	"encoding/json"
	"k8s.io/api/storage/v1"
	apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (p *StatefulsetLauncher) launchStorageClass(storageClass v1.StorageClass) (*v1.StorageClass, error) {
	json_, _ := json.Marshal(storageClass)
	utils.Info.Println(string(json_))
	return p.kubeClient.StorageV1().StorageClasses().Create(&storageClass)
}
