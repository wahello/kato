// KATO, Application Management Platform
// Copyright (C) 2021 Gridworkz Co., Ltd.

// Permission is hereby granted, free of charge, to any person obtaining a copy of this 
// software and associated documentation files (the "Software"), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons 
// to whom the Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all copies or
// substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
// PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
// FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/gridworkz/kato/pkg/apis/kato/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeHelmApps implements HelmAppInterface
type FakeHelmApps struct {
	Fake *FakeKatoV1alpha1
	ns   string
}

var helmappsResource = schema.GroupVersionResource{Group: "kato.io", Version: "v1alpha1", Resource: "helmapps"}

var helmappsKind = schema.GroupVersionKind{Group: "kato.io", Version: "v1alpha1", Kind: "HelmApp"}

// Get takes name of the helmApp, and returns the corresponding helmApp object, and an error if there is any.
func (c *FakeHelmApps) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.HelmApp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(helmappsResource, c.ns, name), &v1alpha1.HelmApp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmApp), err
}

// List takes label and field selectors, and returns the list of HelmApps that match those selectors.
func (c *FakeHelmApps) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.HelmAppList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(helmappsResource, helmappsKind, c.ns, opts), &v1alpha1.HelmAppList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HelmAppList{ListMeta: obj.(*v1alpha1.HelmAppList).ListMeta}
	for _, item := range obj.(*v1alpha1.HelmAppList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested helmApps.
func (c *FakeHelmApps) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(helmappsResource, c.ns, opts))

}

// Create takes the representation of a helmApp and creates it.  Returns the server's representation of the helmApp, and an error, if there is any.
func (c *FakeHelmApps) Create(ctx context.Context, helmApp *v1alpha1.HelmApp, opts v1.CreateOptions) (result *v1alpha1.HelmApp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(helmappsResource, c.ns, helmApp), &v1alpha1.HelmApp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmApp), err
}

// Update takes the representation of a helmApp and updates it. Returns the server's representation of the helmApp, and an error, if there is any.
func (c *FakeHelmApps) Update(ctx context.Context, helmApp *v1alpha1.HelmApp, opts v1.UpdateOptions) (result *v1alpha1.HelmApp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(helmappsResource, c.ns, helmApp), &v1alpha1.HelmApp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmApp), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeHelmApps) UpdateStatus(ctx context.Context, helmApp *v1alpha1.HelmApp, opts v1.UpdateOptions) (*v1alpha1.HelmApp, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(helmappsResource, "status", c.ns, helmApp), &v1alpha1.HelmApp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmApp), err
}

// Delete takes name of the helmApp and deletes it. Returns an error if one occurs.
func (c *FakeHelmApps) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(helmappsResource, c.ns, name), &v1alpha1.HelmApp{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHelmApps) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(helmappsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.HelmAppList{})
	return err
}

// Patch applies the patch and returns the patched helmApp.
func (c *FakeHelmApps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.HelmApp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(helmappsResource, c.ns, name, pt, data, subresources...), &v1alpha1.HelmApp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmApp), err
}
