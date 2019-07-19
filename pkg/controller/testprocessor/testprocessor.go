package testprocessor

import (
	"context"
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

	TestName      string
	TestNamespace string

	ctx       context.Context
	cancel    context.CancelFunc
	currState string

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

func New(
	apiExtCli *apiextclient.Clientset,
	kubeCli *kubernetes.Clientset,
	testsCli *versioned.Clientset,

	test *v1alpha1.Test,
) *TestProcessor {
	proc := &TestProcessor{
		ApiExtCli: apiExtCli,
		KubeCli:   kubeCli,
		TestsCli:  testsCli,

		TestName:      test.Name,
		TestNamespace: test.Namespace,

		knociName:      os.Getenv("MY_POD_NAME"),
		knociNamespace: os.Getenv("MY_POD_NAMESPACE"),
		knociUID:       types.UID(os.Getenv("MY_POD_UID")),

		numTestPodName: "knoci-numtests-" + test.Name,
		testPodName:    "knoci-test-" + test.Name,
		hash:           hashTest(test),
	}
	proc.ctx, proc.cancel = context.WithCancel(context.TODO())

	return proc
}

func (processor *TestProcessor) Process() error {
	if err := processor.ctx.Err(); err != nil {
		return err
	}

	for processor.currState != v1alpha1.StateRunning && processor.currState != v1alpha1.StateFailed {
		err := states[processor.currState].Process(processor)
		if err != nil {
			return fmt.Errorf("failed to process test %s in state %s: %s", processor.TestName, processor.currState, err.Error())
		}
	}
	return nil
}
