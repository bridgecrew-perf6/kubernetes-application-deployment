package kubernetes

import (
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
func (p *SecretsLauncher) CreateRegistrySecret(req *v1.Secret) (*v1.Secret, error) {

	return p.kubeClient.CoreV1().Secrets(req.Namespace).Create(req)
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
