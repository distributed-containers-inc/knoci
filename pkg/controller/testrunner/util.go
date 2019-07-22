package testrunner

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

func parsePodStartEndTests(pod *corev1.Pod) (start, end int64, err error){
	testStartStr, ok := pod.Labels["knoci-test-start"]
	if !ok {
		return 0, 0, fmt.Errorf("pod was missing a start test, even though it was labeled with knoci-test-name: %s", pod.Name)
	}
	testEndStr, ok := pod.Labels["knoci-test-end"]
	if !ok {
		return 0, 0, fmt.Errorf("pod was missing an end test, even though it was labeled with knoci-test-name: %s", pod.Name)
	}
	testStart, err := strconv.ParseInt(testStartStr, 10, 64)
	if err != nil {
		return 0 ,0, fmt.Errorf("pod had invalid, non-numeric start time (%s): %s", testStartStr, err.Error())
	}
	testEnd, err := strconv.ParseInt(testEndStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("pod had invalid, non-numeric end time (%s): %s", testEndStr, err.Error())
	}
	return testStart, testEnd, nil
}