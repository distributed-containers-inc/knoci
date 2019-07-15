package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const Version = "v1alpha1"

// +genclient
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/kubernetes/runtime.Object,k8s.io/kubernetes/runtime.List
// +k8s:openapi-gen=true
type Test struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpec   `json:"spec"`
	Status            TestStatus `json:"status"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
type TestSpec struct {
	Image string `json:"image"`
	// +optional
	Parallelism int `json:"parallelism"`
}

type TestStatus struct {
	State string `json:"state"`
}
