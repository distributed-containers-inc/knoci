package testrunner

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (runner *TestRunner) SetState(test *v1alpha1.Test, state, reason string) (*v1alpha1.Test, error){
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
	runner.holder.TrackTest(test)
	return currTest, nil
}
