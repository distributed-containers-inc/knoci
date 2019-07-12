package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
type Test struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec TestSpec `json:"spec"`
}

// +k8s:openapi-gen=true
type TestSpec struct {
	Containers []corev1.Container `json:"containers"`
}

