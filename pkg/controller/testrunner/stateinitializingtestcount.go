package testrunner

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func setTestCount(runner *TestRunner, test *v1alpha1.Test, count int64) error {
	test, err := runner.TestsCli.TestingV1alpha1().Tests(test.Namespace).Get(test.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	test.Status.NumberOfTests = count
	_, err = runner.TestsCli.TestingV1alpha1().Tests(test.Namespace).UpdateStatus(test)
	return err
}

func readPodLogs(runner *TestRunner, namespace, name string) (string, error) {
	req := runner.KubeCli.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{})
	data, err := req.DoRaw()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (runner *TestRunner) ProcessTestCount(test *v1alpha1.Test) error {
	podName := "knoci-numtestget-" + test.Name
	pod, err := runner.KubeCli.CoreV1().Pods(test.Namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pod: %s", err.Error())
	}
	switch pod.Status.Phase {
	case corev1.PodPending:
		detailedMessage := pod.Status.Reason
		if len(pod.Status.ContainerStatuses) == 1 {
			if status := pod.Status.ContainerStatuses[0].State.Waiting; status != nil {
				detailedMessage = status.Message
			}
		}
		if detailedMessage == "" {
			_, err = runner.SetState(test, v1alpha1.StateInitializingTestCount, "Pod "+podName+" is pending.")
		} else {
			_, err = runner.SetState(test, v1alpha1.StateInitializingTestCount, "Pod "+podName+" is pending: "+detailedMessage)
		}
	case corev1.PodRunning:
		_, err = runner.SetState(test, v1alpha1.StateInitializingTestCount, "Pod "+podName+" is running")
	case corev1.PodSucceeded:
		logs, err := readPodLogs(runner, test.Namespace, podName)
		if err != nil {
			return err
		}
		testCount, err := strconv.ParseInt(logs, 10, 64)
		if err != nil {
			testCount = 0
			_, err = runner.SetState(test, v1alpha1.StateRunnable, "Pod "+podName+" did not return an integer from /num_test, running without concurrency.")
			if err != nil {
				return err
			}
		} else {
			_, err = runner.SetState(test, v1alpha1.StateRunnable, fmt.Sprintf("Test %s can run %d tests in parallel.", test.Name, testCount))
			if err != nil {
				return err
			}
		}
		err = setTestCount(runner, test, testCount)
	case corev1.PodFailed:
		err = setTestCount(runner, test, 0)
		if err != nil {
			return err
		}
		detailedMessage := pod.Status.Reason
		if len(pod.Status.ContainerStatuses) == 1 {
			if status := pod.Status.ContainerStatuses[0].State.Terminated; status != nil {
				detailedMessage = status.Message
			}
		}
		if detailedMessage == "" {
			_, err = runner.SetState(test, v1alpha1.StateRunnable, "Pod "+podName+" does not implement /num_test, running without concurrency.")
		} else {
			_, err = runner.SetState(test, v1alpha1.StateRunnable, "Pod "+podName+" does not implement /num_test, running without concurrency: "+detailedMessage)
		}
	}
	return err
}
