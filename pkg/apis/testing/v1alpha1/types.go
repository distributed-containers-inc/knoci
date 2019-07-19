package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const Version = "v1alpha1"

// +genclient
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Test struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpec    `json:"spec"`
	Status            *TestStatus `json:"status"`
}

type TestSpec struct {
	Image string `json:"image"`
	// +optional
	Parallelism int64 `json:"parallelism"`
}

const StateInitial = ""
const StateInitializingTestCount = "InitializingTestCount"
const StateRunning = "Running"
const StateSuccess = "Success"
const StateFailed = "Failed"

type TestStatus struct {
	State          string    `json:"state"`
	Reason         string    `json:"reason"`
	NumberOfTests  int64     `json:"numberOfTests"`
	CanParallelize bool      `json:"canParallelize"`
	OwnerUID       types.UID `json:"owner"`
	OwnerName      string    `json:"ownerName"`
	OwnerNamespace string    `json:"ownerNamespace"`
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestList is a list of Test resources
type TestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Test `json:"items"`
}
