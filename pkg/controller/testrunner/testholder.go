package testrunner

import (
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type TestInfoHolder struct {
	tests        map[types.UID]*v1alpha1.Test
	testStatuses *InjectiveMultimap

	deletedTests map[types.UID]*v1alpha1.Test
}

func NewTestInfoHolder() *TestInfoHolder {
	return &TestInfoHolder{
		tests:        make(map[types.UID]*v1alpha1.Test),
		testStatuses: NewInjectiveMultimap(),
		deletedTests: make(map[types.UID]*v1alpha1.Test),
	}
}

func (holder *TestInfoHolder) AddTest(test *v1alpha1.Test) {
	holder.tests[test.UID] = test
	holder.UpdateTestStatus(test)
}

func (holder *TestInfoHolder) UpdateTestStatus(test *v1alpha1.Test) {
	holder.testStatuses.Put(test.Status.State, holder.tests[test.UID])
}

func (holder *TestInfoHolder) RemoveTest(uid types.UID) {
	holder.deletedTests[uid] = holder.tests[uid]
	delete(holder.tests, uid)
}
