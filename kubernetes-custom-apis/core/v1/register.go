package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type SchemaGroupTemp struct {
	SchemeGroupVersion schema.GroupVersion
}

func (t *SchemaGroupTemp) AddKnownTypes(scheme *runtime.Scheme) error {

	scheme.AddKnownTypes(t.SchemeGroupVersion,
		&RuntimeConfig{},
		&RuntimeConfigList{},
	)

	metav1.AddToGroupVersion(scheme, t.SchemeGroupVersion)
	return nil
}
func NewClient(cfg *rest.Config, schemeGroupVersion schema.GroupVersion) (*RuntimeConfigV1Alpha1Client, error) {
	scheme := runtime.NewScheme()
	t := SchemaGroupTemp{SchemeGroupVersion: schemeGroupVersion}
	SchemeBuilder := runtime.NewSchemeBuilder(t.AddKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.GroupVersion = &schemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &RuntimeConfigV1Alpha1Client{restClient: client}, nil
}
