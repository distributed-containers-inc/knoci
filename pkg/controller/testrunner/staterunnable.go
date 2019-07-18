package testrunner

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (runner *TestRunner) ProcessRunnable(test *v1alpha1.Test) error {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "knoci-test-" + test.Name,
			Namespace: test.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "test",
					Image:   test.Spec.Image,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err := runner.KubeCli.CoreV1().Pods(pod.Namespace).Create(pod)
	if errors.IsAlreadyExists(err) {
		currPod, err := runner.KubeCli.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failure while querying existing test pod: %s", err.Error())
		}
		if currPod.CreationTimestamp.After(time.Now().Add(-30 * time.Second)) {
			return nil  //a recent pod has been created, give it time
		}
		err = runner.KubeCli.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{GracePeriodSeconds: &[]int64{0}[0]})
		if err != nil {
			return fmt.Errorf("failure while deleting existing test pod: %s", err.Error())
		}
		pod, err = runner.KubeCli.CoreV1().Pods(pod.Namespace).Create(pod)
		if err != nil {
			return fmt.Errorf("failure while creating new test pod after deleting old one: %s", err.Error())
		}
	} else if err != nil {
		return fmt.Errorf("failure while creating test pod: %s", err.Error())
	}

	_, err = runner.SetState(test, v1alpha1.StateRunning, "Pod has been created to run tests")
	return err
}
