package crd_runtime

import (
	"reflect"

	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCRD(clientset apiextension.Interface, name, group, version, plural string, object interface{}) error {
	crd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{Name: name},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   reflect.TypeOf(object).Name(),
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
