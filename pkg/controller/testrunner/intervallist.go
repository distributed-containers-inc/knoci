package testrunner

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strconv"
)

//An IntervalList is a partitioning of (1...global end) into [start, end] pairs
type IntervalList []struct{start, end int64}

func (list *IntervalList) append(start, end int64) {
	*list = append(*list, struct{start, end int64}{start, end})
}

func (list IntervalList) sort() {
	sort.Slice(list, func(i, j int) bool {
		return list[i].end < list[j].end
	})
}

func (list IntervalList) validGivenEnd(end int64) bool {
	var expectedStart int64 = 1
	for _, interval := range list {
		if interval.start != expectedStart {
			return false
		}
		expectedStart = interval.end+1
	}
	return expectedStart == end+1
}

func (runner *TestRunner) LoadPreviousIntervals() (IntervalList, error) {
	podList, err := runner.kubeCli.CoreV1().Pods(runner.testNamespace).List(metav1.ListOptions{
		LabelSelector: "knoci-test-name="+runner.testName,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading pods: %s", err.Error())
	}
	intervalList := new(IntervalList)
	for _, pod := range podList.Items {
		startTimeStr, ok := pod.Labels["knoci-test-start"]
		if !ok {
			return nil, fmt.Errorf("pod was missing a start test, even though it was labeled with knoci-test-name: %s", pod.Name)
		}
		endTimeStr, ok := pod.Labels["knoci-test-end"]
		if !ok {
			return nil, fmt.Errorf("pod was missing an end test, even though it was labeled with knoci-test-name: %s", pod.Name)
		}
		startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("pod had invalid, non-numeric start time (%s): %s", startTimeStr, err.Error())
		}
		endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("pod had invalid, non-numeric end time (%s): %s", endTimeStr, err.Error())
		}
		intervalList.append(startTime, endTime)
	}
	intervalList.sort()
	if !intervalList.validGivenEnd(runner.NumberOfTests) {
		return nil, fmt.Errorf("the number of pods has changed")
	}
}