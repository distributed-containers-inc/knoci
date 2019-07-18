package testprocessor

import (
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (processor *TestProcessor) getTest() (*v1alpha1.Test, error) {
	return processor.TestsCli.TestingV1alpha1().Tests(processor.TestNamespace).Get(processor.TestName, metav1.GetOptions{})
}

func (processor *TestProcessor) updateTestStatus(test *v1alpha1.Test) (*v1alpha1.Test, error) {
	return processor.TestsCli.TestingV1alpha1().Tests(processor.TestNamespace).UpdateStatus(test)
}