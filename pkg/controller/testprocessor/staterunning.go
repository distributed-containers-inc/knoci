package testprocessor

import (
	"bytes"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StateRunning struct{}

func (s *StateRunning) DeletePod(processor *TestProcessor) error {
	err := processor.KubeCli.CoreV1().Pods(processor.TestNamespace).Delete(
		processor.testPodName,
		&metav1.DeleteOptions{GracePeriodSeconds: &[]int64{0}[0]},
	)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (s *StateRunning) CreatePod(processor *TestProcessor) error {
	test, err := processor.getTest()
	if err != nil {
		return fmt.Errorf("could not get the test: %s", err.Error())
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      processor.testPodName,
			Namespace: processor.TestNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: test.Spec.Image,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err = processor.KubeCli.CoreV1().Pods(pod.Namespace).Create(pod)
	return err
}

func (s *StateRunning) Process(processor *TestProcessor) error {
	err := s.DeletePod(processor)
	if err != nil {
		return fmt.Errorf("failure while deleting the test pod: %s", err.Error())
	}
	err = s.CreatePod(processor)
	if err != nil {
		return fmt.Errorf("failure while creating test pod: %s", err.Error())
	}
	watch, err := processor.KubeCli.CoreV1().Pods(processor.TestNamespace).Watch(
		metav1.ListOptions{FieldSelector: "metadata.name=" + processor.testPodName})
	if err != nil {
		return fmt.Errorf("could not watch test pod: %s", err.Error())
	}
	defer watch.Stop()

	for event := range watch.ResultChan() {
		pod := event.Object.(*corev1.Pod)

		newHash, err := processor.hashTest()
		if err != nil {
			return fmt.Errorf("could not hash the tests spec: %s", err.Error())
		}
		if !bytes.Equal(newHash, processor.hash) {
			err := s.DeletePod(processor)
			if err != nil  {
				return fmt.Errorf("could not delete existing test pod named %s", processor.numTestPodName)
			}
			err = processor.setState(v1alpha1.StateInitial, "the spec changed while the test was running")
			if err != nil {
				return err
			}
			return processor.Process()
		}

		detailedMessage := pod.Status.Reason
		if len(pod.Status.ContainerStatuses) == 1 {
			containerState := pod.Status.ContainerStatuses[0].State
			if containerState.Terminated != nil {
				detailedMessage = containerState.Terminated.Message
			} else if containerState.Waiting != nil {
				detailedMessage = containerState.Waiting.Message
			}
		}
		err = nil
		switch pod.Status.Phase {
		case corev1.PodPending:
			err = processor.setState(v1alpha1.StateRunning, "pod is pending: "+detailedMessage)
		case corev1.PodSucceeded:
			watch.Stop()
			err = processor.setState(v1alpha1.StateSuccess, "all tests succeeded")
		case corev1.PodFailed:
			watch.Stop()
			err = processor.setState(v1alpha1.StateSuccess, fmt.Sprintf("tests failed, see logs of pod %s in namespace %s for details", processor.testPodName, processor.TestNamespace))
		case corev1.PodRunning:
			err = processor.setState(v1alpha1.StateRunning, "pod is running: "+detailedMessage)
		case corev1.PodUnknown:
			watch.Stop()
			err = processor.setState(v1alpha1.StateFailed, "could not get pod information: "+detailedMessage)
			if err != nil {
				return err
			}
			return fmt.Errorf("could not reach pod %s in namespace %s: %s", processor.testPodName, processor.TestNamespace, detailedMessage)
		}
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("watch ended before we could figure out pod status for pod %s in namespace %s", processor.testPodName, processor.TestNamespace)
}
