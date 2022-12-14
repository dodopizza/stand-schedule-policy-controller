/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	standschedulesv1 "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeStandSchedulePolicies implements StandSchedulePolicyInterface
type FakeStandSchedulePolicies struct {
	Fake *FakeStandSchedulesV1
}

var standschedulepoliciesResource = schema.GroupVersionResource{Group: "automation.dodois.io", Version: "v1", Resource: "standschedulepolicies"}

var standschedulepoliciesKind = schema.GroupVersionKind{Group: "automation.dodois.io", Version: "v1", Kind: "StandSchedulePolicy"}

// Get takes name of the standSchedulePolicy, and returns the corresponding standSchedulePolicy object, and an error if there is any.
func (c *FakeStandSchedulePolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *standschedulesv1.StandSchedulePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(standschedulepoliciesResource, name), &standschedulesv1.StandSchedulePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*standschedulesv1.StandSchedulePolicy), err
}

// List takes label and field selectors, and returns the list of StandSchedulePolicies that match those selectors.
func (c *FakeStandSchedulePolicies) List(ctx context.Context, opts v1.ListOptions) (result *standschedulesv1.StandSchedulePolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(standschedulepoliciesResource, standschedulepoliciesKind, opts), &standschedulesv1.StandSchedulePolicyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &standschedulesv1.StandSchedulePolicyList{ListMeta: obj.(*standschedulesv1.StandSchedulePolicyList).ListMeta}
	for _, item := range obj.(*standschedulesv1.StandSchedulePolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested standSchedulePolicies.
func (c *FakeStandSchedulePolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(standschedulepoliciesResource, opts))
}

// Create takes the representation of a standSchedulePolicy and creates it.  Returns the server's representation of the standSchedulePolicy, and an error, if there is any.
func (c *FakeStandSchedulePolicies) Create(ctx context.Context, standSchedulePolicy *standschedulesv1.StandSchedulePolicy, opts v1.CreateOptions) (result *standschedulesv1.StandSchedulePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(standschedulepoliciesResource, standSchedulePolicy), &standschedulesv1.StandSchedulePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*standschedulesv1.StandSchedulePolicy), err
}

// Update takes the representation of a standSchedulePolicy and updates it. Returns the server's representation of the standSchedulePolicy, and an error, if there is any.
func (c *FakeStandSchedulePolicies) Update(ctx context.Context, standSchedulePolicy *standschedulesv1.StandSchedulePolicy, opts v1.UpdateOptions) (result *standschedulesv1.StandSchedulePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(standschedulepoliciesResource, standSchedulePolicy), &standschedulesv1.StandSchedulePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*standschedulesv1.StandSchedulePolicy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeStandSchedulePolicies) UpdateStatus(ctx context.Context, standSchedulePolicy *standschedulesv1.StandSchedulePolicy, opts v1.UpdateOptions) (*standschedulesv1.StandSchedulePolicy, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(standschedulepoliciesResource, "status", standSchedulePolicy), &standschedulesv1.StandSchedulePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*standschedulesv1.StandSchedulePolicy), err
}

// Delete takes name of the standSchedulePolicy and deletes it. Returns an error if one occurs.
func (c *FakeStandSchedulePolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(standschedulepoliciesResource, name, opts), &standschedulesv1.StandSchedulePolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStandSchedulePolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(standschedulepoliciesResource, listOpts)

	_, err := c.Fake.Invokes(action, &standschedulesv1.StandSchedulePolicyList{})
	return err
}

// Patch applies the patch and returns the patched standSchedulePolicy.
func (c *FakeStandSchedulePolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *standschedulesv1.StandSchedulePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(standschedulepoliciesResource, name, pt, data, subresources...), &standschedulesv1.StandSchedulePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*standschedulesv1.StandSchedulePolicy), err
}
