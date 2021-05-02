package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type==\"Ready\")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type MoneroNodeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoneroNodeSetSpec   `json:"spec,omitempty"`
	Status MoneroNodeSetStatus `json:"status,omitempty"`
}

type MoneroNodeSetSpec struct {
	//+kubebuilder:default=1
	Replicas         uint32        `json:"replicas"`
	HardAntiAffinity bool          `json:"hardAntiAffinity,omitempty"`
	Monerod          MonerodConfig `json:"monerod,omityempty"`
}

type MonerodConfig struct {
	Image  string            `json:"image,omitempty"`
	Config map[string]string `json:"config"`
}

type MoneroNodeSetStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

type MoneroNodeSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MoneroNodeSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MoneroNodeSet{}, &MoneroNodeSetList{})
}
