// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "go.pinniped.dev/generated/1.17/apis/concierge/authentication/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeJWTAuthenticators implements JWTAuthenticatorInterface
type FakeJWTAuthenticators struct {
	Fake *FakeAuthenticationV1alpha1
	ns   string
}

var jwtauthenticatorsResource = schema.GroupVersionResource{Group: "authentication.concierge.pinniped.dev", Version: "v1alpha1", Resource: "jwtauthenticators"}

var jwtauthenticatorsKind = schema.GroupVersionKind{Group: "authentication.concierge.pinniped.dev", Version: "v1alpha1", Kind: "JWTAuthenticator"}

// Get takes name of the jWTAuthenticator, and returns the corresponding jWTAuthenticator object, and an error if there is any.
func (c *FakeJWTAuthenticators) Get(name string, options v1.GetOptions) (result *v1alpha1.JWTAuthenticator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(jwtauthenticatorsResource, c.ns, name), &v1alpha1.JWTAuthenticator{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.JWTAuthenticator), err
}

// List takes label and field selectors, and returns the list of JWTAuthenticators that match those selectors.
func (c *FakeJWTAuthenticators) List(opts v1.ListOptions) (result *v1alpha1.JWTAuthenticatorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(jwtauthenticatorsResource, jwtauthenticatorsKind, c.ns, opts), &v1alpha1.JWTAuthenticatorList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.JWTAuthenticatorList{ListMeta: obj.(*v1alpha1.JWTAuthenticatorList).ListMeta}
	for _, item := range obj.(*v1alpha1.JWTAuthenticatorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested jWTAuthenticators.
func (c *FakeJWTAuthenticators) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(jwtauthenticatorsResource, c.ns, opts))

}

// Create takes the representation of a jWTAuthenticator and creates it.  Returns the server's representation of the jWTAuthenticator, and an error, if there is any.
func (c *FakeJWTAuthenticators) Create(jWTAuthenticator *v1alpha1.JWTAuthenticator) (result *v1alpha1.JWTAuthenticator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(jwtauthenticatorsResource, c.ns, jWTAuthenticator), &v1alpha1.JWTAuthenticator{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.JWTAuthenticator), err
}

// Update takes the representation of a jWTAuthenticator and updates it. Returns the server's representation of the jWTAuthenticator, and an error, if there is any.
func (c *FakeJWTAuthenticators) Update(jWTAuthenticator *v1alpha1.JWTAuthenticator) (result *v1alpha1.JWTAuthenticator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(jwtauthenticatorsResource, c.ns, jWTAuthenticator), &v1alpha1.JWTAuthenticator{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.JWTAuthenticator), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeJWTAuthenticators) UpdateStatus(jWTAuthenticator *v1alpha1.JWTAuthenticator) (*v1alpha1.JWTAuthenticator, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(jwtauthenticatorsResource, "status", c.ns, jWTAuthenticator), &v1alpha1.JWTAuthenticator{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.JWTAuthenticator), err
}

// Delete takes name of the jWTAuthenticator and deletes it. Returns an error if one occurs.
func (c *FakeJWTAuthenticators) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(jwtauthenticatorsResource, c.ns, name), &v1alpha1.JWTAuthenticator{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeJWTAuthenticators) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(jwtauthenticatorsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.JWTAuthenticatorList{})
	return err
}

// Patch applies the patch and returns the patched jWTAuthenticator.
func (c *FakeJWTAuthenticators) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.JWTAuthenticator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(jwtauthenticatorsResource, c.ns, name, pt, data, subresources...), &v1alpha1.JWTAuthenticator{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.JWTAuthenticator), err
}
