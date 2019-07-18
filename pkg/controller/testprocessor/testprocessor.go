package testprocessor

import (
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
}

type State interface {
	Process(processor *TestProcessor) error
}

var states = map[string]State{
	"": &StateInitial{},
}

func (processor *TestProcessor) Process() error {
	processor.knociName = os.Getenv("MY_POD_NAME")
	processor.knociNamespace = os.Getenv("MY_POD_NAMESPACE")
	processor.knociUID = types.UID(os.Getenv("MY_POD_UID"))
	return states[processor.currState].Process(processor)
}
