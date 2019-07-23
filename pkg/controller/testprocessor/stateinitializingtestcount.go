package testprocessor

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

type StateInitializingTestCount struct{}

func readPodLogs(processor *TestProcessor) (string, error) {
	req := processor.KubeCli.CoreV1().Pods(processor.TestNamespace).GetLogs(processor.numTestPodName, &corev1.PodLogOptions{})
	data, err := req.DoRaw()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (s *StateInitializingTestCount) setNumTests(processor *TestProcessor, numTests int64) error {
	test, err := processor.getTest()
	if err != nil {
		return fmt.Errorf("could not get the test: %s", err.Error())
	}
	test.Status.NumberOfTests = numTests
	_, err = processor.updateTestStatus(test)
	return err
}

func (s *StateInitializingTestCount) processFailedPodCount(processor *TestProcessor) error {
	err := s.setNumTests(processor, 0)
	if err != nil {
		return err
	}
	err = processor.setState(v1alpha1.StateRunning, "Pod "+processor.numTestPodName+" does not implement /num_test (which should return the number of tests it will run as a single integer to stdout), running without concurrency")
	if err != nil {
		return err
	}
	return s.setNumTests(processor, 0)
}

func (s *StateInitializingTestCount) processPodCount(processor *TestProcessor) error {
	logs, err := readPodLogs(processor)
	if err != nil {
		return err
	}
	testCount, err := strconv.ParseInt(logs, 10, 64)
	if err != nil {
		return processor.setState(v1alpha1.StateFailed, "Pod "+processor.numTestPodName+" did not return an integer ("+logs+") from /num_test")
	}
	err = processor.setState(v1alpha1.StateRunning, fmt.Sprintf("Test %s can run %d tests in parallel.", processor.TestName, testCount))
	if err != nil {
		return err
	}
	return s.setNumTests(processor, testCount)
}

func (s *StateInitializingTestCount) DeletePod(processor *TestProcessor) error {
	err := processor.KubeCli.CoreV1().Pods(processor.TestNamespace).Delete(
		processor.numTestPodName,
		&metav1.DeleteOptions{GracePeriodSeconds: &[]int64{0}[0]},
	)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (s *StateInitializingTestCount) CreatePod(processor *TestProcessor) error {
	test, err := processor.getTest()
	if err != nil {
		return fmt.Errorf("could not get the test: %s", err.Error())
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "corev1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      processor.numTestPodName,
			Namespace: processor.TestNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "numtestget",
					Image:   test.Spec.Image,
					Command: []string{"/num_tests"},
					Args:    []string{},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err = processor.KubeCli.CoreV1().Pods(processor.TestNamespace).Create(pod)
	return err
}

func (s *StateInitializingTestCount) Process(processor *TestProcessor) error {
	err := s.DeletePod(processor)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("could not delete existing 'numtestget' named %s: %s", processor.numTestPodName, err.Error())
	}

	err = s.CreatePod(processor)
	if err != nil {
		return fmt.Errorf("could not create a 'numtestget' pod: %s", err.Error())
	}

	watch, err := processor.KubeCli.CoreV1().Pods(processor.TestNamespace).Watch(
		metav1.ListOptions{FieldSelector: "metadata.name=" + processor.numTestPodName})

	if err != nil {
		return fmt.Errorf("could not watch 'numtestget' pod: %s", err.Error())
	}
	defer watch.Stop()
	for {
		select {
		case event, ok := <-watch.ResultChan():
			if !ok {
				return fmt.Errorf("watch ended before we could figure out pod status for pod %s in namespace %s", processor.numTestPodName, processor.TestNamespace)
			}

			pod := event.Object.(*corev1.Pod)
			detailedMessage := pod.Status.Reason
			if len(pod.Status.ContainerStatuses) == 1 {
				containerState := pod.Status.ContainerStatuses[0].State
				if containerState.Terminated != nil {
					detailedMessage = containerState.Terminated.Message
				} else if containerState.Waiting != nil {
					detailedMessage = containerState.Waiting.Message
				}
			}
			switch pod.Status.Phase {
			case corev1.PodPending:
				err = processor.setState(v1alpha1.StateInitializingTestCount, "pod is pending: "+detailedMessage)
				if err != nil {
					return err
				}
			case corev1.PodRunning:
				err = processor.setState(v1alpha1.StateInitializingTestCount, "pod is running: "+detailedMessage)
				if err != nil {
					return err
				}
			case corev1.PodSucceeded:
				return s.processPodCount(processor)
			case corev1.PodFailed:
				return s.processFailedPodCount(processor)
			case corev1.PodUnknown:
				return processor.setState(v1alpha1.StateFailed, "could not get pod information: "+detailedMessage)
			}
		case <-processor.ctx.Done():
			return processor.ctx.Err()
		}
	}
}
