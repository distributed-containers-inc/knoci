package controller

import (
	"github.com/distributed-containers-inc/knoci/pkg/api"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BoolPointer(val bool) *bool {
	return &val
}

func NewResourceDefinition(
	specSchema apiextv1beta1.JSONSchemaProps,
	names apiextv1beta1.CustomResourceDefinitionNames,
	) *apiextv1beta1.CustomResourceDefinition {
	return &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: names.Plural+"."+api.Group, //i.e., tests.knoci.distributedcontainers.com
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group: api.Group,
			Names: names,
			Versions: []apiextv1beta1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Scope: apiextv1beta1.NamespaceScoped,
			PreserveUnknownFields: BoolPointer(false),
			Validation: &apiextv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextv1beta1.JSONSchemaProps{
						"spec": specSchema,
					},
				},
			},
		},
	}
}

func CreateResourceDefinition(
	clientset apiextclient.Interface,
	resourcedef *apiextv1beta1.CustomResourceDefinition,
) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(resourcedef)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func CreateTestResourceDefinition(clientset apiextclient.Interface) error {
	schema := apiextv1beta1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"image": {
				Type: "string",
			},
			"parallelization": {
				Type: "integer",
			},
		},
		Required: []string{"image"},
	}

	names := apiextv1beta1.CustomResourceDefinitionNames{
		Plural:     "tests",
		Singular:   "test",
		Kind:       "Test",
		ShortNames: []string{"te", "ts"},
	}

	resourceDef := NewResourceDefinition(schema, names)
	return CreateResourceDefinition(clientset, resourceDef)
}
