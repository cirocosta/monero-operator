package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=monero
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type==\"Ready\")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type MoneroMiningNodeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoneroMiningNodeSetSpec   `json:"spec,omitempty"`
	Status MoneroMiningNodeSetStatus `json:"status,omitempty"`
}

type MoneroMiningNodeSetSpec struct {
	//+kubebuilder:default=1
	Replicas         uint32 `json:"replicas"`
	HardAntiAffinity bool   `json:"hardAntiAffinity,omitempty"`

	Xmrig XmrigConfig `json:"xmrig,omitempty"`
}

type XmrigConfig struct {
	//+kubebuilder:default="index.docker.io/utxobr/xmrig@sha256:a0a231a6fc983885f7fb0ce68fffca027bb2fa032851539901b99ebbfd9140a1"
	Image string   `json:"image,omitempty"`
	Args  []string `json:"args,omitempty"`
}

type MoneroMiningNodeSetStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

type MoneroMiningNodeSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MoneroMiningNodeSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MoneroMiningNodeSet{}, &MoneroMiningNodeSetList{})
}
