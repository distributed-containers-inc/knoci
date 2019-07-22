package testrunner

import (
	"context"
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sync"
	"time"
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

	watchEstablished sync.Cond

	podMutex      sync.Mutex
	runningPods   map[types.UID]*corev1.Pod
	pendingPods   map[types.UID]*corev1.Pod
	completedPods map[types.UID]*corev1.Pod

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

func (runner *TestRunner) createPod(startTest, endTest *int64) error {
	podName := fmt.Sprintf("knoci-test-%s", runner.testName)
	labels := map[string]string{
		"knoci-test-name": runner.testName,
	}
	var args []string = nil
	if startTest == nil {
		labels["knoci-test-start"] = "1"
		labels["knoci-test-end"] = "1"
		podName = fmt.Sprintf("knoci-test-%s", runner.testName)
	} else {
		startTestStr := fmt.Sprintf("%d", *startTest)
		endTestStr := fmt.Sprintf("%d", *endTest)
		labels["knoci-test-start"] = startTestStr
		labels["knoci-test-end"] = endTestStr
		args = []string{startTestStr, endTestStr}
		podName = fmt.Sprintf("knoci-test-%s-%d-%d", runner.testName, *startTest, *endTest)
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

	delete(runner.runningPods, pod.UID)
	delete(runner.pendingPods, pod.UID)
	if len(pod.Status.ContainerStatuses) == 1 {
		containerState := pod.Status.ContainerStatuses[0].State
		if containerState.Running != nil {
			runner.runningPods[pod.UID] = pod
			return
		}
	}
	switch pod.Status.Phase {
	case corev1.PodRunning: //pod is, e.g., in ImagePullBackoff since no containers have been scheduled
		runner.pendingPods[pod.UID] = pod
	case corev1.PodPending:
		runner.pendingPods[pod.UID] = pod
	case corev1.PodFailed:
		runner.completedPods[pod.UID] = pod
	case corev1.PodSucceeded:
		runner.completedPods[pod.UID] = pod
	}
}

func (runner *TestRunner) watchPods(ctx context.Context) error {
	watch, err := runner.kubeCli.CoreV1().Pods(runner.testNamespace).Watch(
		metav1.ListOptions{LabelSelector: "knoci-test-name=" + runner.testName})
	if err != nil {
		return err
	}
	runner.watchEstablished.Broadcast()
	defer watch.Stop()
	for runner.podsAreRunning() {
		select {
		case event, ok := <-watch.ResultChan():
			if !ok {
				return fmt.Errorf("watch ended before the testrunner could complete")
			}

			runner.updatePodCache(event.Object.(*corev1.Pod))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (runner *TestRunner) podsAreRunning() bool {
	runner.podMutex.Lock()
	defer runner.podMutex.Unlock()
	return len(runner.pendingPods)+len(runner.runningPods) > 0
}

func (runner *TestRunner) startInitialPods(oldIntervals *IntervalList) error {
	runner.watchEstablished.Wait()
	//todo use old intervals
	if runner.Parallelize {
		return runner.createPod(&[]int64{1}[0], &runner.NumberOfTests)
	} else {
		return runner.createPod(nil, nil)
	}
}

func (runner *TestRunner) tryToSplit(ctx context.Context) error {
	for runner.podsAreRunning() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		runner.podMutex.Lock()
		for _, pod := range runner.runningPods {
			podStart, podEnd, err := parsePodStartEndTests(pod)
			if err != nil {
				return err
			}
			podStartedAt := pod.Status.ContainerStatuses[0].State.Running.StartedAt
			if podStart != podEnd && time.Now().After(podStartedAt.Add(time.Second*time.Duration(runner.SplittingTime))) {
				//todo split pod
			}
		}
		runner.podMutex.Unlock()
		time.Sleep(time.Second)
	}
	return nil
}

func (runner *TestRunner) Run() error {
	var oldIntervals *IntervalList
	var err error

	if runner.Parallelize {
		oldIntervals, err = runner.LoadPreviousIntervals()
		if err != nil {
			klog.Infof("Not parallelizing test %s, could not load previous intervals: %s", runner.testName, err.Error())
			oldIntervals = nil
		}
	}
	err = runner.deletePods()
	if err != nil {
		return fmt.Errorf("could not delete old pods for test %s: %s", runner.testName, err.Error())
	}

	eg, ctx := errgroup.WithContext(runner.ctx)
	eg.Go(func() error {return runner.watchPods(ctx)})
	eg.Go(func() error {return runner.tryToSplit(ctx)})
	eg.Go(func() error {return runner.startInitialPods(oldIntervals)})

	err = eg.Wait()
	if err != context.Canceled {
		return err
	}

	completedIntervals := IntervalList{}
	for _, pod := range runner.completedPods {
		startTime, endTime, err := parsePodStartEndTests(pod)
		if err != nil {
			return fmt.Errorf("error while checking completed pods: %s", err.Error())
		}
		completedIntervals.append(startTime, endTime)
	}
	completedIntervals.sort()
	if completedIntervals.validGivenEnd(runner.NumberOfTests) {
		return nil
	}
	return fmt.Errorf("some pods did not complete successfully, try 'kubectl get po -l knoci-test-name=%s'", runner.testName)
}
