//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureResource) DeepCopyInto(out *AzureResource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureResource.
func (in *AzureResource) DeepCopy() *AzureResource {
	if in == nil {
		return nil
	}
	out := new(AzureResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatusCondition) DeepCopyInto(out *StatusCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatusCondition.
func (in *StatusCondition) DeepCopy() *StatusCondition {
	if in == nil {
		return nil
	}
	out := new(StatusCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcesSpec) DeepCopyInto(out *ResourcesSpec) {
	*out = *in
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = make([]AzureResource, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcesSpec.
func (in *ResourcesSpec) DeepCopy() *ResourcesSpec {
	if in == nil {
		return nil
	}
	out := new(ResourcesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScheduleSpec) DeepCopyInto(out *ScheduleSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScheduleSpec.
func (in *ScheduleSpec) DeepCopy() *ScheduleSpec {
	if in == nil {
		return nil
	}
	out := new(ScheduleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StandSchedulePolicy) DeepCopyInto(out *StandSchedulePolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StandSchedulePolicy.
func (in *StandSchedulePolicy) DeepCopy() *StandSchedulePolicy {
	if in == nil {
		return nil
	}
	out := new(StandSchedulePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StandSchedulePolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StandSchedulePolicyList) DeepCopyInto(out *StandSchedulePolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StandSchedulePolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StandSchedulePolicyList.
func (in *StandSchedulePolicyList) DeepCopy() *StandSchedulePolicyList {
	if in == nil {
		return nil
	}
	out := new(StandSchedulePolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StandSchedulePolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StandSchedulePolicySpec) DeepCopyInto(out *StandSchedulePolicySpec) {
	*out = *in
	out.Schedule = in.Schedule
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StandSchedulePolicySpec.
func (in *StandSchedulePolicySpec) DeepCopy() *StandSchedulePolicySpec {
	if in == nil {
		return nil
	}
	out := new(StandSchedulePolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StandSchedulePolicyStatus) DeepCopyInto(out *StandSchedulePolicyStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]StatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StandSchedulePolicyStatus.
func (in *StandSchedulePolicyStatus) DeepCopy() *StandSchedulePolicyStatus {
	if in == nil {
		return nil
	}
	out := new(StandSchedulePolicyStatus)
	in.DeepCopyInto(out)
	return out
}
