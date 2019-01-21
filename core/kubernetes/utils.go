package kubernetes

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNameSpace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	existingNamespace, err := client.CoreV1().Namespaces().Get(namespace, v12.GetOptions{})
	if err == nil {
		return existingNamespace, nil
	}
	return client.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: v12.ObjectMeta{Name: namespace}})
}
