package kubernetes

import (
	"encoding/base64"
	"encoding/json"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretsLauncher struct {
	kubeClient *kubernetes.Clientset
}

func NewSecretsLauncher(c *kubernetes.Clientset) *SecretsLauncher {
	this := new(SecretsLauncher)
	this.kubeClient = c
	return this
}
func (p *SecretsLauncher) CreateRegistrySecret(name, namespace, username, password, email, registryUrl string) (*v1.Secret, error) {
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
	return p.kubeClient.CoreV1().Secrets(namespace).Create(&config)
}

func (p *SecretsLauncher) GetRegistrySecret(name, namespace string) (*v1.Secret, error) {
	return p.kubeClient.CoreV1().Secrets(namespace).Get(name, v12.GetOptions{})
}
func (p *SecretsLauncher) DeleteRegistrySecret(name, namespace string) error {
	return p.kubeClient.CoreV1().Secrets(namespace).Delete(name, &v12.DeleteOptions{})
}

////todo: Implement this method
//func (p *SecretsLauncher) UpdateRegistrySecrets(secret *v1.Secret) (secret2 v1.Secret, err error) {
//	return secret, nil
//}
