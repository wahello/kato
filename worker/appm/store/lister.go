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

package store

import (
	crdlisters "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/internalversion"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	autoscalingv2 "k8s.io/client-go/listers/autoscaling/v2beta2"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/listers/extensions/v1beta1"
	storagev1 "k8s.io/client-go/listers/storage/v1"
)

//Lister kube-api client cache
type Lister struct {
	Ingress                 v1beta1.IngressLister
	Service                 corev1.ServiceLister
	Secret                  corev1.SecretLister
	StatefulSet             appsv1.StatefulSetLister
	Deployment              appsv1.DeploymentLister
	Pod                     corev1.PodLister
	ConfigMap               corev1.ConfigMapLister
	Endpoints               corev1.EndpointsLister
	Nodes                   corev1.NodeLister
	StorageClass            storagev1.StorageClassLister
	Claims                  corev1.PersistentVolumeClaimLister
	HorizontalPodAutoscaler autoscalingv2.HorizontalPodAutoscalerLister
	CRD                     crdlisters.CustomResourceDefinitionLister
}
