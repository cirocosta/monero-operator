package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type==\"Ready\")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type MoneroNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoneroNetworkSpec   `json:"spec,omitempty"`
	Status MoneroNetworkStatus `json:"status,omitempty"`
}

type MoneroNetworkSpec struct {
	//+kubebuilder:default=3
	Replicas uint32                `json:"replicas"`
	Template MoneroNetworkTemplate `json:"template"`
}

type MoneroNetworkTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MoneroNodeSetSpec `json:"spec"`
}

type MoneroNetworkStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

type MoneroNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MoneroNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MoneroNetwork{}, &MoneroNetworkList{})
}
