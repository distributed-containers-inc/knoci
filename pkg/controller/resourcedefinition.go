package controller

import (
	crdvalidation "github.com/ant31/crd-validation/pkg"
	testingv1alpha1 "github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

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
	resourceDef := crdvalidation.NewCustomResourceDefinition(crdvalidation.Config{
		Kind:               "Test",
		Plural:             "tests",
		ShortNames:         []string{"tst", "te"},
		SpecDefinitionName: "github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1.Test",
		Group:              "knoci.distributedcontainers.com",

		EnableValidation:      true,
		Version:               testingv1alpha1.Version,
		GetOpenAPIDefinitions: testingv1alpha1.GetOpenAPIDefinitions,
	})

	return CreateResourceDefinition(clientset, resourceDef)
}
