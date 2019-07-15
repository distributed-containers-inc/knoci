package main

import (
  "fmt"
  "github.com/distributed-containers-inc/knoci/pkg/client/versioned"
  "github.com/distributed-containers-inc/knoci/pkg/controller"
  "k8s.io/apimachinery/pkg/runtime"
  "os"
  "time"

  "k8s.io/apimachinery/pkg/api/errors"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func main() {
  config, err := rest.InClusterConfig()
  if err != nil {
    panic(err.Error())
  }

  // creates the clientset
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err.Error())
  }

  apiextcli := apiextclient.New(clientset.RESTClient())
  testscli := versioned.New(clientset.RESTClient())

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

  for {
    pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
    if err != nil {
      panic(err.Error())
    }
    fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

    // Examples for error handling:
    // - Use helper functions like e.g. errors.IsNotFound()
    // - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
    _, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
    if errors.IsNotFound(err) {
      fmt.Printf("Pod not found\n")
    } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
      fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
    } else if err != nil {
      panic(err.Error())
    } else {
      fmt.Printf("Found pod\n")
    }

    time.Sleep(10 * time.Second)
  }
}
