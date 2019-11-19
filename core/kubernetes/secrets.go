package kubernetes

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubernetes-services-deployment/utils"
	"time"
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

	if req.Namespace == "" {
		req.Namespace = "default"
	}
	_, err := CreateNameSpace(p.kubeClient, req.Namespace)
	if err != nil {
		utils.Error.Println(err)
		return nil, err
	}
	dep, err := p.kubeClient.CoreV1().Secrets(req.Namespace).Create(req)
	for dep == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			dep, err = p.kubeClient.CoreV1().Secrets(req.Namespace).Create(req)
		} else {
			break
		}
	}
	return dep, err

}

func (p *SecretsLauncher) GetRegistrySecret(name, namespace string) (*v1.Secret, error) {

	dep, err := p.kubeClient.CoreV1().Secrets(namespace).Get(name, v12.GetOptions{})
	for dep == nil && err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			dep, err = p.kubeClient.CoreV1().Secrets(namespace).Get(name, v12.GetOptions{})
		} else {
			break
		}
	}
	return dep, err
}
func (p *SecretsLauncher) DeleteRegistrySecret(name, namespace string) error {

	err := p.kubeClient.CoreV1().Secrets(namespace).Delete(name, &v12.DeleteOptions{})
	for err != nil {
		if err.Error() == "" {
			time.Sleep(1 * time.Second)
			err = p.kubeClient.CoreV1().Secrets(namespace).Delete(name, &v12.DeleteOptions{})
		} else {
			break
		}
	}
	return err
}

////todo: Implement this method
//func (p *SecretsLauncher) UpdateRegistrySecrets(secret *v1.Secret) (secret2 v1.Secret, err error) {
//	return secret, nil
//}
