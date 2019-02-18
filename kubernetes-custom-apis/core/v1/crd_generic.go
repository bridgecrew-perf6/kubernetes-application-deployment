package v1

import (
	"encoding/json"
	"fmt"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type runtimeConfigclient struct {
	client       rest.Interface
	ns           string
	resourceName string
}

type RuntimeConfigInterface interface {
	Create(obj interface{}) (interface{}, error)
	Update(obj interface{}) (interface{}, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (interface{}, error)
	Patch(name string, pt kubernetesTypes.PatchType, data []byte, subresources ...string) (interface{}, error)
}

func (c *runtimeConfigclient) Create(obj interface{}) (interface{}, error) {
	result := &RuntimeConfig{}
	raw_data, err := c.client.Post().
		Namespace(c.ns).Resource(c.resourceName).
		Body(obj).Do().Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (c *runtimeConfigclient) Update(obj interface{}) (interface{}, error) {
	result := &RuntimeConfig{}
	raw_data, err := c.client.Put().
		Namespace(c.ns).
		Resource(c.resourceName).
		Body(obj).
		Do().
		Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
func (c *runtimeConfigclient) Patch(name string, pt kubernetesTypes.PatchType, data []byte, subresources ...string) (interface{}, error) {
	result := &RuntimeConfig{}
	raw_data, err := c.client.Patch(pt).
		Namespace(c.ns).
		Resource(c.resourceName).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
func (c *runtimeConfigclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource(c.resourceName).
		Name(name).Body(options).Do().
		Error()
}

func (c *runtimeConfigclient) Get(name string) (interface{}, error) {
	result := &RuntimeConfig{}
	raw_data, err := c.client.Get().
		Namespace(c.ns).Resource(c.resourceName).
		Name(name).Do().Raw()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(raw_data))
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
