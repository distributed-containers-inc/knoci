package main

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	"github.com/distributed-containers-inc/knoci/pkg/controller"
	"github.com/distributed-containers-inc/knoci/pkg/controller/testrunner"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	apiextcli := apiextclient.NewForConfigOrDie(config)
	testscli := versioned.NewForConfigOrDie(config)

	err = controller.CreateTestResourceDefinition(apiextcli)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create the custom resource definition: %s", err.Error())
		os.Exit(1)
	}
	err = controller.WaitForCRDReady(func(options metav1.ListOptions) (runtime.Object, error) {
		return testscli.TestingV1alpha1().Tests("default").List(options)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not wait for the Custom Resource Definition to exist: %s", err.Error())
		os.Exit(1)
	}
	fmt.Println("Successfully created the Test resource definition.")

	runner := testrunner.NewTestRunner(config)
	err = runner.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not start the runner: %s", err.Error())
		os.Exit(1)
	}
	err = runner.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
