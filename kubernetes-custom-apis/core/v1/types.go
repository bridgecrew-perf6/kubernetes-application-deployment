package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RuntimeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              interface{}       `json:"spec"`
	Data              interface{}       `json:"data,omitempty"`
	Type              interface{}       `json:"type,omitempty"`
	Status            interface{}       `json:"status,omitempty"`
	Rules             interface{}       `json:"rules,omitempty"`
	RoleRef           interface{}       `json:"roleRef,omitempty"`
	Subjects          interface{}       `json:"subjects,omitempty"`
	StringData        interface{}       `json:"stringData,omitempty"`
	SCprovisioner     string            `json:"provisioner,omitempty"`
	SCparameters      map[string]string `json:"parameters,omitempty"`
	SCReclaimPolicy   string            `json:"reclaimPolicy,omitempty" protobuf:"bytes,4,opt,name=reclaimPolicy,casttype=k8s.io/api/core/v1.PersistentVolumeReclaimPolicy"`

	SCMountOptions []string `json:"mountOptions,omitempty" protobuf:"bytes,5,opt,name=mountOptions"`

	SCAllowVolumeExpansion bool `json:"allowVolumeExpansion,omitempty" protobuf:"varint,6,opt,name=allowVolumeExpansion"`

	SCVolumeBindingMode string      `json:"volumeBindingMode,omitempty" protobuf:"bytes,7,opt,name=volumeBindingMode"`
	SCAllowedTopologies interface{} `json:"allowedTopologies,omitempty" protobuf:"bytes,8,rep,name=allowedTopologies"`
}
type RuntimeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []RuntimeConfig `json:"items"`
}
