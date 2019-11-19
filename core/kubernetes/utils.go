package kubernetes

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func CreateNameSpace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	existingNamespace, err := client.CoreV1().Namespaces().Get(namespace, v12.GetOptions{})
	if err == nil {
		return existingNamespace, nil
	}
	ns, err := client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: v12.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"istio-injection": "enabled",
			},
		},
	})
	for ns == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			ns, err = client.CoreV1().Namespaces().Create(&v1.Namespace{
				ObjectMeta: v12.ObjectMeta{
					Name: namespace,
					Labels: map[string]string{
						"istio-injection": "enabled",
					},
				},
			})
		} else {
			break
		}
	}
	return ns, err
}
