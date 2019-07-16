package testrunner

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"time"
)

type testInfoKey struct {
	namespace string
	name      string
}

type TestInfo struct {
	*v1alpha1.Test

	lastRunTime time.Duration
	failCount   int
	runCount    int

	//lastSeen is the last time this test existed
	lastSeen time.Time
}

type TestInfoHolder struct {
	tests map[testInfoKey]*TestInfo
}

func NewTestInfoHolder() *TestInfoHolder {
	return &TestInfoHolder{
		tests: make(map[testInfoKey]*TestInfo),
	}
}

func (holder *TestInfoHolder) AddTest(test *v1alpha1.Test) {
	key := testInfoKey{test.Namespace, test.Name}
	if _, ok := holder.tests[key]; !ok {
		holder.tests[key] = &TestInfo{
			Test: test,
		}
	}
}

func (holder *TestInfoHolder) Reconcile(tests []v1alpha1.Test) {
	newTestMap := make(map[testInfoKey]*v1alpha1.Test)
	for _, test := range tests {
		newTestMap[testInfoKey{test.Namespace, test.Name}] = &test
	}
	for key := range holder.tests {
		if _, ok := newTestMap[key]; !ok {
			fmt.Println("Test deleted: "+key.name)
		}
	}
	for key, test := range newTestMap {
		if _, ok := holder.tests[key]; !ok {
			fmt.Println("Test added: "+key.name)
			holder.AddTest(test)
		}
	}
}
