package testprocessor

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

func (processor *TestProcessor) CheckTestOwnedByUs() bool {
	test, err := processor.getTest()
	if err != nil {
		return false
	}
	return test.Status != nil && test.Status.OwnerUID == processor.knociUID
}

func (processor *TestProcessor) CheckOwnerAlive() (bool, error) {
	test, err := processor.getTest()
	if err != nil {
		return false, err
	}
	if test.Status == nil {
		return false, nil //it's a new test
	}
	ownerName := test.Status.OwnerName
	ownerNamespace := test.Status.OwnerNamespace
	ownerPod, err := processor.KubeCli.CoreV1().Pods(ownerNamespace).Get(ownerName, metav1.GetOptions{})
	if err == nil {
		return true, nil
	} else if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return ownerPod.Status.Phase == corev1.PodRunning, nil
}

func (processor *TestProcessor) AtomicOwn() bool {
	test, err := processor.getTest()
	if err != nil {
		return false
	}
	if test.Status == nil {
		test.Status = &v1alpha1.TestStatus{}
	}
	test.Status.OwnerUID = processor.knociUID
	test.Status.OwnerNamespace = processor.knociNamespace
	test.Status.OwnerName = processor.knociName

	_, err = processor.updateTestStatus(test)
	return err == nil
}

type StateInitial struct{}

func (s *StateInitial) Process(processor *TestProcessor) error {
	for i := 0; i < 3; i++ {
		if err := processor.ctx.Err(); err != nil {
			return err
		}

		if !processor.CheckTestOwnedByUs() {
			alive, err := processor.CheckOwnerAlive()
			if err != nil {
				return fmt.Errorf("could not check if tests owner is alive: %s", err.Error())
			}
			if alive {
				return nil
			}
			if processor.AtomicOwn() {
				break
			}
		}
		time.Sleep(time.Millisecond * time.Duration(rand.IntnRange(100, 500)))
	}
	if !processor.CheckTestOwnedByUs() {
		return fmt.Errorf("test could not be owned by us and its owner was not alive")
	}
	return processor.setState(v1alpha1.StateInitializingTestCount, "Test is being processed by "+processor.knociNamespace + "/"+processor.knociName)
}
