package testrunner

import (
	"errors"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	"github.com/distributed-containers-inc/knoci/pkg/controller/states"
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

	tests        map[types.UID]*v1alpha1.Test
	testStatuses *InjectiveMultimap

	running bool
	Stop    chan bool
	errors  chan error
}

func NewTestRunner(config *rest.Config) *TestRunner {
	return &TestRunner{
		ApiExtCli: apiextclient.NewForConfigOrDie(config),
		KubeCli:   kubernetes.NewForConfigOrDie(config),
		TestsCli:  versioned.NewForConfigOrDie(config),

		tests:        make(map[types.UID]*v1alpha1.Test),
		testStatuses: NewInjectiveMultimap(),

		Stop:   make(chan bool),
		errors: make(chan error),
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
			runner.testStatuses.ForAllOfState(v1alpha1.StatePending, func(test *v1alpha1.Test) {
				err := states.ProcessTestParallelism(runner, test)
				if err != nil {
					runner.errors <- err //TODO use a logger instead of crashing2
				}
			})
			runner.testStatuses.ForAllOfState(v1alpha1.StateInitializingTestCount, func(test *v1alpha1.Test) {
				err := states.ProcessTestCount(runner, test)
				if err != nil {
					runner.errors <- err
				}
			})
			runner.testStatuses.ForAllOfState(v1alpha1.StateRunnable, func(test *v1alpha1.Test) {
				err := states.ProcessRunnable(runner, test)
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
	for key := range runner.tests {
		if _, ok := newTestMap[key]; !ok {
			runner.RemoveTest(key)
		}
	}
	for key, test := range newTestMap {
		if _, ok := runner.tests[key]; !ok {
			_, err := runner.SetState(test, v1alpha1.StatePending, "The test has just been created.")
			if err != nil {
				runner.errors <- err
			}
		} else {
			runner.TrackTest(test)
		}
	}
}

func (runner *TestRunner) TrackTest(test *v1alpha1.Test) {
	runner.tests[test.UID] = test
	runner.testStatuses.Put(test.Status.State, runner.tests[test.UID])
}

func (runner *TestRunner) RemoveTest(uid types.UID) {
	delete(runner.tests, uid)
}

func (runner *TestRunner) SetState(test *v1alpha1.Test, state, reason string) (*v1alpha1.Test, error) {
	currTest, err := runner.TestsCli.TestingV1alpha1().Tests(test.Namespace).Get(test.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get latest version for %s/%s: %s", test.Namespace, test.Name, err.Error())
	}

	test = currTest

	if test.Status == nil {
		test.Status = &v1alpha1.TestStatus{}
	}

	test.Status.State = state
	test.Status.Reason = reason
	currTest, err = runner.TestsCli.TestingV1alpha1().Tests(test.Namespace).UpdateStatus(test)
	if err != nil {
		return nil, fmt.Errorf("failure while changing state from Pending -> StateInitializingTestCount for %s/%s: %s", test.Namespace, test.Name, err.Error())
	}
	runner.TrackTest(test)
	return currTest, nil
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
