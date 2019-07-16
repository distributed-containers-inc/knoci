package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const Version = "v1alpha1"

// +genclient
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Test struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpec   `json:"spec"`
	Status            *TestStatus `json:"status"`
}

type TestSpec struct {
	Image string `json:"image"`
	// +optional
	Parallelism int64 `json:"parallelism"`
}

const StatePending = "Pending"
const StateInitializingTestCount = "InitializingTestCount"
const StateRunning = "Running"
const StateSucceeded = "Success"
const StateFailed = "Failed"
type TestStatus struct {
	State string `json:"state"`
	NumberOfTests int64 `json:"numberOfTests"`
	CanParallelize bool `json:"canParallelize"`
}

// +k8s:deepcopy-gen=true

// TestList is a list of Test resources
type TestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Test `json:"items"`
}
