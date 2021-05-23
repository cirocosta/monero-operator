package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=monero
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type==\"Ready\")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type MoneroNodeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoneroNodeSetSpec   `json:"spec,omitempty"`
	Status MoneroNodeSetStatus `json:"status,omitempty"`
}

func (self *MoneroNodeSet) ApplyDefaults() {
	self.Spec.ApplyDefaults()
}

type MoneroNodeSetSpec struct {
	Replicas         uint32               `json:"replicas,omitempty"`
	HardAntiAffinity bool                 `json:"hardAntiAffinity,omitempty"`
	DiskSize         string               `json:"diskSize,omitempty"`
	Service          MoneroNodeSetService `json:"service,omitempty"`
	StorageClass     string               `json:"storageClass,omitempty"`
	Tor              MoneroTorConfig      `json:"tor,omitempty"`

	Monerod MonerodConfig `json:"monerod,omitempty"`
}

type MoneroNodeSetService struct {
	Type string `json:"type"`
}

func (self *MoneroNodeSetSpec) ApplyDefaults() {
	if self.DiskSize == "" {
		self.DiskSize = "50Gi"
	}

	if self.Replicas == 0 {
		self.Replicas = 1
	}

	self.Monerod.ApplyDefaults()
}

type MoneroTorConfig struct {
	Enabled   bool                        `json:"enabled,omitempty"`
	SecretRef corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

type MonerodConfig struct {
	//+kubebuilder:default=""
	Image string   `json:"image,omitempty"`
	Args  []string `json:"args,omitempty"`
}

const (
	DefaultMonerodImage = "index.docker.io/utxobr/monerod@sha256:19ba5793c00375e7115469de9c14fcad928df5867c76ab5de099e83f646e175d"
)

func (self *MonerodConfig) ApplyDefaults() {
	if self.Image == "" {
		self.Image = DefaultMonerodImage
	}
}

type MoneroNodeSetStatus struct {
	Conditions []metav1.Condition  `json:"conditions,omitempty"`
	Tor        MoneroNodeStatusTor `json:"tor,omitempty"`
}

type MoneroNodeStatusTor struct {
	Address string `json:"address,omitempty"`
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
