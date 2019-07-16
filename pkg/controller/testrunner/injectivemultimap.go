package testrunner

import (
	"k8s.io/apimachinery/pkg/types"
	"sync"
)

type TestInfoSet map[types.UID]*TestInfo

func (set TestInfoSet) delete(testInfo *TestInfo) {
	delete(set, testInfo.UID)
}

func (set TestInfoSet) add(testInfo *TestInfo) {
	(set)[testInfo.UID] = testInfo
}

type InjectiveMultimap struct {
	mutex sync.Mutex

	buckets         map[string]TestInfoSet
	elementStateMap map[types.UID]string
}

func NewInjectiveMultimap() *InjectiveMultimap {
	return &InjectiveMultimap{
		buckets:         make(map[string]TestInfoSet),
		elementStateMap: make(map[types.UID]string),
	}
}

func (m *InjectiveMultimap) Put(state string, testInfo *TestInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if currState, ok := m.elementStateMap[testInfo.UID]; ok {
		m.buckets[currState].delete(testInfo)
		delete(m.elementStateMap, testInfo.UID)
	}

	if _, ok := m.buckets[testInfo.Status.State]; !ok {
		m.buckets[testInfo.Status.State] = make(map[types.UID]*TestInfo)
	}
	m.buckets[testInfo.Status.State].add(testInfo)
	m.elementStateMap[testInfo.UID] = testInfo.Status.State
}

func (m *InjectiveMultimap) Delete(testInfo *TestInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if currState, ok := m.elementStateMap[testInfo.UID]; ok {
		m.buckets[currState].delete(testInfo)
		delete(m.elementStateMap, testInfo.UID)
	}
}

func (m *InjectiveMultimap) ForAllOfState(state string, fn func(*TestInfo)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if keys, ok := m.buckets[state]; ok {
		for _, test := range keys {
			fn(test)
		}
	}
}