package kubernetes

import (
	"encoding/base64"
	"encoding/json"
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

func CreateRegistrySecret(client *kubernetes.Clientset, name, namespace, username, password, email, registryUrl string) (*v1.Secret, error) {
	dockerConf := map[string]map[string]string{
		registryUrl: {
			"email": email,
			"auth":  base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
		},
	}
	dockerConfMarshaled, _ := json.Marshal(dockerConf)

	data := map[string][]byte{
		".dockercfg": dockerConfMarshaled,
	}
	config := v1.Secret{Type: v1.SecretTypeDockercfg, ObjectMeta: v12.ObjectMeta{Name: name, Namespace: namespace}, Data: data}
	return client.CoreV1().Secrets(namespace).Create(&config)
}
