package controller

import (
	"fmt"
	crdvalidation "github.com/ant31/crd-validation/pkg"
	testingv1alpha1 "github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"time"
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
		ShortNames:         []string{"ts", "te"},
		SpecDefinitionName: "github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1.Test",
		Group:              "knoci.distributedcontainers.com",

		EnableValidation:      true,
		Version:               testingv1alpha1.Version,
		GetOpenAPIDefinitions: testingv1alpha1.GetOpenAPIDefinitions,
	})

	resourceDef.Spec.Subresources.Scale = nil

	return CreateResourceDefinition(clientset, resourceDef)
}

//Attribution: Prometheus Operator
//WaitForCRDReady waits until a CRD specified by a given listfunc exists.
func WaitForCRDReady(listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	err := wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := listFunc(metav1.ListOptions{})
		if err != nil {
			if se, ok := err.(*apierrors.StatusError); ok {
				if se.Status().Code == http.StatusNotFound {
					return false, nil
				}
			}
			return false, errors.Wrap(err, "failed to list CRD")
		}
		return true, nil
	})

	return errors.Wrap(err, fmt.Sprintf("timed out waiting for Custom Resource"))
}