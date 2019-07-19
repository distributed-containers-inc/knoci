package main

import (
	"context"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	"github.com/distributed-containers-inc/knoci/pkg/controller"
	"github.com/distributed-containers-inc/knoci/pkg/controller/testprocessor"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	apiextcli := apiextclient.NewForConfigOrDie(config)
	kubecli := kubernetes.NewForConfigOrDie(config)
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

	if os.Getenv("MY_POD_NAME") == "" || os.Getenv("MY_POD_NAMESPACE") == "" {
		fmt.Fprintln(os.Stderr, "MY_POD_NAME and MY_POD_NAMESPACE must be set. Are you using the provided knoci manifests?")
		os.Exit(1)
	}
	if os.Getenv("MY_POD_UID") == "" {
		fmt.Fprintln(os.Stderr, "MY_POD_UID must be set. Are you using the provided knoci manifest?")
		os.Exit(1)
	}
	watchlist := controller.TestListWatcher{
		TestsCli: testscli,
		AddFunc: func(test *v1alpha1.Test) {
			processor := testprocessor.New(
				apiextcli,
				kubecli,
				testscli,
				test,
			)
			err := processor.Process()
			if err != nil && err != context.Canceled {
				fmt.Fprintf(os.Stderr, "Error while executing test %s in namespace %s: %s\n", test.Name, test.Namespace, err.Error())
			}
		},
		DeleteFunc: func(test *v1alpha1.Test) {
			//TODO processor context cancel
		},
		UpdateFunc: func(oldTest, newTest *v1alpha1.Test) {
			//TODO
		},
	}
	watchlist.Run()
}
