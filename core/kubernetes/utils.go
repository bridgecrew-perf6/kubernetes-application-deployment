package kubernetes

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

func CreateNameSpace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	existingNamespace, err := client.CoreV1().Namespaces().Get(namespace, v12.GetOptions{})
	if err == nil {
		return existingNamespace, nil
	}
	nsTemp := v1.Namespace{ObjectMeta: v12.ObjectMeta{
		Name: namespace,
		Labels: map[string]string{
			"istio-injection": "enabled",
		},
	}}
	ns, err := client.CoreV1().Namespaces().Create(&nsTemp)
	for ns == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			ns, err = client.CoreV1().Namespaces().Create(&nsTemp)
		} else {
			break
		}
	}
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ns, nil
	}
	return ns, err
}
