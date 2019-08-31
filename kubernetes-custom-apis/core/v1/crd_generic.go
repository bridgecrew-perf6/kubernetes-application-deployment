package v1

import (
	"encoding/json"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"time"
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
	List(opts meta_v1.ListOptions) (interface{}, error)
	Patch(name string, pt kubernetesTypes.PatchType, data []byte, subresources ...string) (interface{}, error)
}

func (c *runtimeConfigclient) Create(obj interface{}) (interface{}, error) {
	result := &RuntimeConfig{}
	request := c.client.Post().
		Resource(c.resourceName).
		Body(obj)
	if c.ns != "" {
		request.Namespace(c.ns)
	}
	resultTemp := request.Do()
	raw_data, err := resultTemp.Raw()
	if err != nil {
		return nil, resultTemp.Error()
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (c *runtimeConfigclient) Update(obj interface{}) (interface{}, error) {
	result := &RuntimeConfig{}
	request := c.client.Put().
		Namespace(c.ns).
		Resource(c.resourceName).
		Body(obj)
	//if c.ns != "" {
	//	request.Namespace(c.ns)
	//}
	resultTemp := request.Do()
	raw_data, err := resultTemp.Raw()
	if err != nil {
		return nil, resultTemp.Error()
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
func (c *runtimeConfigclient) Patch(name string, pt kubernetesTypes.PatchType, data []byte, subresources ...string) (interface{}, error) {
	result := &RuntimeConfig{}
	request := c.client.Patch(pt).
		Namespace(c.ns).
		Resource(c.resourceName).
		SubResource(subresources...).
		Name(name).
		Body(data)

	//if c.ns != "" {
	//	request.Namespace(c.ns)
	//}
	resultTemp := request.Do()
	raw_data, err := resultTemp.Raw()
	if err != nil {
		return nil, resultTemp.Error()
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
func (c *runtimeConfigclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	request := c.client.Delete().
		Namespace(c.ns).Resource(c.resourceName).
		Name(name).Body(options)
	//if c.ns != "" {
	//	request.Namespace(c.ns)
	//}
	return request.Do().Error()
}

func (c *runtimeConfigclient) Get(name string) (interface{}, error) {
	result := &RuntimeConfig{}
	request := c.client.Get().
		Resource(c.resourceName).
		Name(name)
	//if c.ns != "" {
	//	request.Namespace(c.ns)
	//}
	resultTemp := request.Do()
	raw_data, err := resultTemp.Raw()
	if err != nil {
		return nil, resultTemp.Error()
	}
	err = json.Unmarshal(raw_data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (c *runtimeConfigclient) List(opts meta_v1.ListOptions) (interface{}, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	var result interface{}
	request := c.client.Get().
		Resource(c.resourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout)
	//if c.ns != "" {
	//	request.Namespace(c.ns)
	//}
	resultTemp := request.Do()
	raw_data, err := resultTemp.Raw()
	if err != nil {
		return nil, resultTemp.Error()
	}
	err = json.Unmarshal(raw_data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
