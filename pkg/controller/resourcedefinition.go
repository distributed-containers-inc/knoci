package controller

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func CreateResourceDefinition(clientset apiextclient.Interface) error {
	resourcedef := &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tests.knoci.distributedcontainers.com",
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group: "tests.knoci.distributedcontainers.com",
			Versions: []apiextv1beta1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha1",
					Served: true,
					Storage: true,
				},
			},
			Scope: apiextv1beta1.NamespaceScoped,
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Plural: "tests",
				Singular: "test",
				Kind: "Test",
				ShortNames: []string{"te", "ts"},
			},
			PreserveUnknownFields: &[]bool{false}[0],
			Validation: &apiextv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextv1beta1.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextv1beta1.JSONSchemaProps{
								"image": {
									Type: "string",
								},
								"parallelization": {
									Type: "integer",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(resourcedef)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
