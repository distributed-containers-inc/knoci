package testrunner

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type TestRunner struct {
	kubeCli       *kubernetes.Clientset
	testNamespace string
	testName      string
	testSpec      *v1alpha1.TestSpec

	NumberOfTests int64
	Parallelize   bool
	SplittingTime int64
}

func New(
	kubeCLI *kubernetes.Clientset,
	testNamespace, testName string,
	testSpec *v1alpha1.TestSpec,
) *TestRunner {
	return &TestRunner{
		kubeCli:       kubeCLI,
		testNamespace: testNamespace,
		testName:      testName,
		testSpec:      testSpec,
	}
}

func (runner *TestRunner) deletePods() error {
	return runner.kubeCli.CoreV1().Pods(runner.testNamespace).DeleteCollection(
		&metav1.DeleteOptions{
			GracePeriodSeconds: &[]int64{0}[0],
		},
		metav1.ListOptions{
			LabelSelector: "knoci-test-name=" + runner.testName,
		},
	)
}

func (runner *TestRunner) createPod(startTest, endTest *int) error {
	podName := fmt.Sprintf("knoci-test-%s", runner.testName)
	labels := map[string]string{
		"knoci-test-name": runner.testName,
	}
	var args []string = nil
	if startTest != nil {
		podName = fmt.Sprintf("knoci-test-%s-%d-%d", runner.testName, *startTest, *endTest)
		startTestStr := fmt.Sprintf("%d", *startTest)
		endTestStr := fmt.Sprintf("%d", *endTest)
		labels["knoci-test-start"] = startTestStr
		labels["knoci-test-end"] = endTestStr
		args = []string{startTestStr, endTestStr}
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: runner.testNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: runner.testSpec.Image,
					Args:  args,
				},
			},
		},
	}

	_, err := runner.kubeCli.CoreV1().Pods(pod.Namespace).Create(pod)
	return err
}

func (runner *TestRunner) Run() error {
	return nil //TODO
}
