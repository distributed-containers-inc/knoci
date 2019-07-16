package testrunner

import (
	"errors"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
	"time"
)

type TestRunner struct {
	ApiExtCli *apiextclient.Clientset
	KubeCli   *kubernetes.Clientset
	TestsCli  *versioned.Clientset

	tests *TestInfoHolder

	running bool
	Stop    chan bool
	errors 	chan error
}

func NewTestRunner(config *rest.Config) *TestRunner {
	return &TestRunner{
		ApiExtCli: apiextclient.NewForConfigOrDie(config),
		KubeCli:   kubernetes.NewForConfigOrDie(config),
		TestsCli:  versioned.NewForConfigOrDie(config),
		tests:     NewTestInfoHolder(),
		Stop:      make(chan bool),
		errors:    make(chan error),
	}
}

func (runner *TestRunner) Start() error {
	if runner.running {
		return errors.New("this runner is already started")
	}
	runner.running = true
	go func() {
		for {
			select {
			case <-runner.Stop:
				return
			default:
			}

			tests, err := runner.TestsCli.TestingV1alpha1().Tests("").List(metav1.ListOptions{})
			if err != nil {
				runner.errors <- err
				return
			}
			runner.tests.Reconcile(tests.Items)
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}

func (runner *TestRunner) Wait() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-runner.Stop:
	case <-c:
	case err := <-runner.errors:
		if err != nil {
			return err
		}
	}
	return nil
}
