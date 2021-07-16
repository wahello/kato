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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	katov1alpha1 "github.com/gridworkz/kato/pkg/apis/kato/v1alpha1"
	versioned "github.com/gridworkz/kato/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/gridworkz/kato/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/gridworkz/kato/pkg/generated/listers/kato/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ThirdComponentInformer provides access to a shared informer and lister for
// ThirdComponents.
type ThirdComponentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ThirdComponentLister
}

type thirdComponentInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewThirdComponentInformer constructs a new informer for ThirdComponent type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewThirdComponentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredThirdComponentInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredThirdComponentInformer constructs a new informer for ThirdComponent type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredThirdComponentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KatoV1alpha1().ThirdComponents(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KatoV1alpha1().ThirdComponents(namespace).Watch(context.TODO(), options)
			},
		},
		&katov1alpha1.ThirdComponent{},
		resyncPeriod,
		indexers,
	)
}

func (f *thirdComponentInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredThirdComponentInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *thirdComponentInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&katov1alpha1.ThirdComponent{}, f.defaultInformer)
}

func (f *thirdComponentInformer) Lister() v1alpha1.ThirdComponentLister {
	return v1alpha1.NewThirdComponentLister(f.Informer().GetIndexer())
}
