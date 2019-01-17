package crd_runtime

import meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type RuntimeConfig struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               interface{} `json:"spec"`
	Status             interface{} `json:"status,omitempty"`
}

type RuntimeConfigList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata,omitempty"`
	Items            []RuntimeConfig `json:"items"`
}
