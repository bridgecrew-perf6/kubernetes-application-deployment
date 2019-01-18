package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type runtimeConfigclient struct {
	client       rest.Interface
	ns           string
	resourceName string
}

type RuntimeConfigInterface interface {
	Create(obj *RuntimeConfig) (*RuntimeConfig, error)
	Update(obj *RuntimeConfig) (*RuntimeConfig, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*RuntimeConfig, error)
}

func (c *runtimeConfigclient) Create(obj *RuntimeConfig) (*RuntimeConfig, error) {
	result := &RuntimeConfig{}
	err := c.client.Post().
		Namespace(c.ns).Resource(c.resourceName).
		Body(obj).Do().Into(result)
	return result, err
}

func (c *runtimeConfigclient) Update(obj *RuntimeConfig) (*RuntimeConfig, error) {
	result := &RuntimeConfig{}
	err := c.client.Put().
		Namespace(c.ns).Resource(c.resourceName).
		Body(obj).Do().Into(result)
	return result, err
}

func (c *runtimeConfigclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource(c.resourceName).
		Name(name).Body(options).Do().
		Error()
}

func (c *runtimeConfigclient) Get(name string) (*RuntimeConfig, error) {
	result := &RuntimeConfig{}
	err := c.client.Get().
		Namespace(c.ns).Resource(c.resourceName).
		Name(name).Do().Into(result)
	return result, err
}
