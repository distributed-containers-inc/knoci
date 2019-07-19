package testprocessor

import (
	"bytes"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"os"
)

//TestProcessor : a state machine which takes a test as input, and fully processes it.
type TestProcessor struct {
	ApiExtCli *apiextclient.Clientset
	KubeCli   *kubernetes.Clientset
	TestsCli  *versioned.Clientset

	currState string

	TestName      string
	TestNamespace string

	knociName      string
	knociNamespace string
	knociUID       types.UID

	numTestPodName string
	testPodName    string
	hash           []byte
}

type State interface {
	Process(processor *TestProcessor) error
}

var states = map[string]State{
	v1alpha1.StateInitial:               &StateInitial{},
	v1alpha1.StateInitializingTestCount: &StateInitializingTestCount{},
	v1alpha1.StateRunning:               &StateRunning{},
}

func (processor *TestProcessor) Process() error {
	processor.knociName = os.Getenv("MY_POD_NAME")
	processor.knociNamespace = os.Getenv("MY_POD_NAMESPACE")
	processor.knociUID = types.UID(os.Getenv("MY_POD_UID"))
	processor.numTestPodName = "knoci-numtests-" + processor.TestName
	processor.testPodName = "knoci-test-" + processor.TestName

	newHash, err := processor.hashTest()
	if err != nil {
		return fmt.Errorf("could not hash the tests spec: %s", err.Error())
	}

	if processor.hash != nil && !bytes.Equal(processor.hash, newHash) {
		err = processor.setState(v1alpha1.StateInitializingTestCount, "Test's spec has changed, restarting it")
		if err != nil {
			return err
		}
	}
	processor.hash = newHash

	for processor.currState != v1alpha1.StateRunning && processor.currState != v1alpha1.StateFailed {
		err := states[processor.currState].Process(processor)
		if err != nil {
			return fmt.Errorf("failed to process test %s in state %s: %s", processor.TestName, processor.currState, err.Error())
		}
	}
	return nil
}
