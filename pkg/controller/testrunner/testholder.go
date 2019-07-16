package testrunner

import (
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type TestInfo struct {
	*v1alpha1.Test

	lastRunTime time.Duration
	failCount   int
	runCount    int

	//lastSeen is the last time this test existed
	lastSeen time.Time
}

type TestInfoHolder struct {
	tests        map[types.UID]*TestInfo
	testStatuses *InjectiveMultimap

	deletedTests map[types.UID]*TestInfo
}

func NewTestInfoHolder() *TestInfoHolder {
	return &TestInfoHolder{
		tests:        make(map[types.UID]*TestInfo),
		testStatuses: NewInjectiveMultimap(),
		deletedTests: make(map[types.UID]*TestInfo),
	}
}

func (holder *TestInfoHolder) AddTest(test *v1alpha1.Test) {
	holder.tests[test.UID] = &TestInfo{
		Test: test,
	}
	holder.UpdateTestStatus(test)
}

func (holder *TestInfoHolder) UpdateTestStatus(test *v1alpha1.Test) {
	holder.testStatuses.Put(test.Status.State, holder.tests[test.UID])
}

func (holder *TestInfoHolder) RemoveTest(uid types.UID) {
	holder.tests[uid].lastSeen = time.Now()
	holder.deletedTests[uid] = holder.tests[uid]
	delete(holder.tests, uid)
}
