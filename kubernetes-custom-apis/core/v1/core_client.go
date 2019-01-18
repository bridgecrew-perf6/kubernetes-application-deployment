package v1

import (
	"k8s.io/client-go/rest"
)

type RuntimeConfigV1Alpha1Interface interface {
	NewRuntimeConfigs(namespace string) RuntimeConfigInterface
}
type RuntimeConfigV1Alpha1Client struct {
	restClient rest.Interface
}

func (c *RuntimeConfigV1Alpha1Client) NewRuntimeConfigs(namespace, resourceName string) RuntimeConfigInterface {
	return &runtimeConfigclient{
		client:       c.restClient,
		ns:           namespace,
		resourceName: resourceName,
	}
}
