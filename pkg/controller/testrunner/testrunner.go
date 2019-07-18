package testrunner

import (
	"errors"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

	holder *TestInfoHolder

	running bool
	Stop    chan bool
	errors  chan error
}

func NewTestRunner(config *rest.Config) *TestRunner {
	return &TestRunner{
		ApiExtCli: apiextclient.NewForConfigOrDie(config),
		KubeCli:   kubernetes.NewForConfigOrDie(config),
		TestsCli:  versioned.NewForConfigOrDie(config),
		holder:    NewTestInfoHolder(),
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
			runner.reconcile(tests.Items)
			runner.holder.testStatuses.ForAllOfState(v1alpha1.StatePending, func(test *v1alpha1.Test) {
				err := runner.ProcessTestParallelism(test)
				if err != nil {
					runner.errors <- err
				}
			})
			runner.holder.testStatuses.ForAllOfState(v1alpha1.StateInitializingTestCount, func(test *v1alpha1.Test) {
				err := runner.ProcessTestCount(test)
				if err != nil {
					runner.errors <- err
				}
			})
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}

func (runner *TestRunner) reconcile(tests []v1alpha1.Test) {
	newTestMap := make(map[types.UID]*v1alpha1.Test)
	for _, test := range tests {
		newTestMap[test.UID] = &test
	}
	for key := range runner.holder.tests {
		if _, ok := newTestMap[key]; !ok {
			runner.holder.RemoveTest(key)
		}
	}
	for key, test := range newTestMap {
		if _, ok := runner.holder.tests[key]; !ok {
			_, err := runner.SetState(test, v1alpha1.StatePending, "The test has just been created.")
			if err != nil {
				runner.errors <- err
			}
		} else {
			runner.holder.TrackTest(test)
		}
	}
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
