// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MoneroNodeSet) DeepCopyInto(out *MoneroNodeSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MoneroNodeSet.
func (in *MoneroNodeSet) DeepCopy() *MoneroNodeSet {
	if in == nil {
		return nil
	}
	out := new(MoneroNodeSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MoneroNodeSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MoneroNodeSetList) DeepCopyInto(out *MoneroNodeSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MoneroNodeSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MoneroNodeSetList.
func (in *MoneroNodeSetList) DeepCopy() *MoneroNodeSetList {
	if in == nil {
		return nil
	}
	out := new(MoneroNodeSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MoneroNodeSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MoneroNodeSetSpec) DeepCopyInto(out *MoneroNodeSetSpec) {
	*out = *in
	in.Monerod.DeepCopyInto(&out.Monerod)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MoneroNodeSetSpec.
func (in *MoneroNodeSetSpec) DeepCopy() *MoneroNodeSetSpec {
	if in == nil {
		return nil
	}
	out := new(MoneroNodeSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MoneroNodeSetStatus) DeepCopyInto(out *MoneroNodeSetStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MoneroNodeSetStatus.
func (in *MoneroNodeSetStatus) DeepCopy() *MoneroNodeSetStatus {
	if in == nil {
		return nil
	}
	out := new(MoneroNodeSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonerodConfig) DeepCopyInto(out *MonerodConfig) {
	*out = *in
	in.Config.DeepCopyInto(&out.Config)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonerodConfig.
func (in *MonerodConfig) DeepCopy() *MonerodConfig {
	if in == nil {
		return nil
	}
	out := new(MonerodConfig)
	in.DeepCopyInto(out)
	return out
}
