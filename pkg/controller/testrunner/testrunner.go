package testrunner

import (
	"context"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sync"
)

type TestRunner struct {
	kubeCli       *kubernetes.Clientset
	testNamespace string
	testName      string
	testSpec      *v1alpha1.TestSpec
	ctx           context.Context

	NumberOfTests int64
	Parallelize   bool
	SplittingTime int64

	podMutex    sync.Mutex
	livePods    map[types.UID]*corev1.Pod
	pendingPods map[types.UID]*corev1.Pod

	done bool
}

func New(
	kubeCLI *kubernetes.Clientset,
	testNamespace, testName string,
	testSpec *v1alpha1.TestSpec,
	ctx context.Context,
) *TestRunner {
	return &TestRunner{
		kubeCli:       kubeCLI,
		testNamespace: testNamespace,
		testName:      testName,
		testSpec:      testSpec,
		ctx:           ctx,
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
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err := runner.kubeCli.CoreV1().Pods(pod.Namespace).Create(pod)
	return err
}

func (runner *TestRunner) updatePodCache(pod *corev1.Pod) {
	runner.podMutex.Lock()
	defer runner.podMutex.Unlock()

	delete(runner.livePods, pod.UID)
	delete(runner.pendingPods, pod.UID)
	if len(pod.Status.ContainerStatuses) == 1 {
		containerState := pod.Status.ContainerStatuses[0].State
		if containerState.Running != nil {
			runner.livePods[pod.UID] = pod
			return
		}
	}
	switch pod.Status.Phase {
	case corev1.PodRunning: //pod is, e.g., in ImagePullBackoff since no containers have been scheduled
		runner.pendingPods[pod.UID] = pod
	case corev1.PodPending:
		runner.pendingPods[pod.UID] = pod
	}
}

func (runner *TestRunner) watchPods() error {
	watch, err := runner.kubeCli.CoreV1().Pods(runner.testNamespace).Watch(
		metav1.ListOptions{LabelSelector: "knoci-test-name=" + runner.testName})
	if err != nil {
		return err
	}
	defer watch.Stop()
	for !runner.done {
		select {
		case event, ok := <-watch.ResultChan():
			if !ok {
				return fmt.Errorf("watch ended before the testrunner could complete")
			}

			runner.updatePodCache(event.Object.(*corev1.Pod))
		case <-runner.ctx.Done():
			return runner.ctx.Err()
		}
	}
	return nil
}

func (runner *TestRunner) Start() error {
	return nil //TODO
}
