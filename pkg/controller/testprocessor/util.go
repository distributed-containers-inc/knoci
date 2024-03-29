package testprocessor

import (
	"crypto/sha256"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (processor *TestProcessor) getTest() (*v1alpha1.Test, error) {
	return processor.TestsCli.TestingV1alpha1().Tests(processor.TestNamespace).Get(processor.TestName, metav1.GetOptions{})
}

func (processor *TestProcessor) updateTestStatus(test *v1alpha1.Test) (*v1alpha1.Test, error) {
	return processor.TestsCli.TestingV1alpha1().Tests(processor.TestNamespace).UpdateStatus(test)
}

func (processor *TestProcessor) setState(newState, reason string) error {
	test, err := processor.getTest()
	if err != nil {
		return err
	}
	if test.Status == nil {
		test.Status = &v1alpha1.TestStatus{}
	}
	if test.Status.State == newState {
		return nil
	}
	test.Status.State = newState
	test.Status.Reason = reason
	_, err = processor.updateTestStatus(test)
	processor.currState = test.Status.State
	return err
}

func HashTest(test *v1alpha1.Test) []byte {
	//todo should use image manifest instead of just string representation
	//so that two tagged images with the same manifest do not constitute a change
	hash := sha256.New()
	testSpecStr := fmt.Sprintf("%+v", test.Spec)
	hash.Write([]byte(testSpecStr))
	return hash.Sum(nil)
}