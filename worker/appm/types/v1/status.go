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

package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gridworkz/kato/pkg/apis/kato/v1alpha1"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

//IsEmpty is empty
func (a *AppService) IsEmpty() bool {
	empty := len(a.pods) == 0
	if !empty {
		//The remaining pod is at the missing node and is considered closed successfully
		for _, pod := range a.pods {
			if !IsPodNodeLost(pod) {
				return false
			}
		}
		return true
	}
	return empty
}

//IsClosed is closed
func (a *AppService) IsClosed() bool {
	if a.IsCustomComponent() {
		return a.workload == nil
	}
	if a.IsThirdComponent() {
		if a.endpoints == nil || len(a.endpoints) == 0 {
			return true
		}
	} else {
		if a.IsEmpty() && a.statefulset == nil && a.deployment == nil {
			return true
		}
		if a.IsEmpty() && a.statefulset != nil && a.statefulset.ResourceVersion == "" {
			return true
		}
		if a.IsEmpty() && a.deployment != nil && a.deployment.ResourceVersion == "" {
			return true
		}
	}
	return false
}

var (
	//RUNNING if stateful or deployment exist and ready pod number is equal to the service Replicas
	RUNNING = "running"
	//CLOSED if app service is not in store
	CLOSED = "closed"
	//STARTING if stateful or deployment exist and ready pod number is less than service Replicas
	STARTING = "starting"
	//STOPPING if stateful and deployment is nil and pod number is not 0
	STOPPING = "stopping"
	//ABNORMAL if stateful or deployment exist and ready pod number is less than service Replicas and all pod status is Error
	ABNORMAL = "abnormal"
	//SOMEABNORMAL if stateful or deployment exist and ready pod number is less than service Replicas and some pod status is Error
	SOMEABNORMAL = "some_abnormal"
	//UNKNOW indeterminacy status
	UNKNOW = "unknow"
	//UPGRADE if store have more than 1 app service
	UPGRADE = "upgrade"
	//BUILDING app service is building
	BUILDING = "building"
	//BUILDEFAILURE app service is build failure
	BUILDEFAILURE = "build_failure"
	//UNDEPLOY init status
	UNDEPLOY = "undeploy"
)

func conversionThirdComponent(obj runtime.Object) *v1alpha1.ThirdComponent {
	if third, ok := obj.(*v1alpha1.ThirdComponent); ok {
		return third
	}
	if struc, ok := obj.(*unstructured.Unstructured); ok {
		data, _ := struc.MarshalJSON()
		var third v1alpha1.ThirdComponent
		if err := json.Unmarshal(data, &third); err != nil {
			logrus.Errorf("unmarshal object to ThirdComponent failure")
			return nil
		}
		return &third
	}
	return nil
}

//GetServiceStatus get service status
func (a *AppService) GetServiceStatus() string {
	//TODO: support custom component status
	if a.IsCustomComponent() {
		if a.workload != nil {
			switch a.workload.GetObjectKind().GroupVersionKind().Kind {
			case "ThirdComponent":
				third := conversionThirdComponent(a.workload)
				if third != nil {
					switch third.Status.Phase {
					case v1alpha1.ComponentFailed:
						return ABNORMAL
					case v1alpha1.ComponentRunning:
						return RUNNING
					case v1alpha1.ComponentPending:
						return STARTING
					}
				}
				return RUNNING
			default:
				return RUNNING
			}
		}
		return CLOSED
	}
	if a.IsThirdComponent() {
		endpoints := a.GetEndpoints(false)
		if len(endpoints) == 0 {
			return CLOSED
		}
		var readyEndpointSize int
		for _, ed := range endpoints {
			for _, s := range ed.Subsets {
				readyEndpointSize += len(s.Addresses)
			}
		}
		if readyEndpointSize > 0 {
			return RUNNING
		}
		return ABNORMAL
	}
	if a == nil {
		return CLOSED
	}
	if a.IsClosed() {
		return CLOSED
	}
	if a.statefulset == nil && a.deployment == nil && len(a.pods) > 0 {
		return STOPPING
	}
	if (a.statefulset != nil || a.deployment != nil) && len(a.pods) < a.Replicas {
		return STARTING
	}
	if a.statefulset != nil && a.statefulset.Status.ReadyReplicas >= int32(a.Replicas) {
		if a.UpgradeComlete() {
			return RUNNING
		}
		return UPGRADE
	}
	if a.deployment != nil && a.deployment.Status.ReadyReplicas >= int32(a.Replicas) {
		if a.UpgradeComlete() {
			return RUNNING
		}
		return UPGRADE
	}

	if a.deployment != nil && (a.deployment.Status.ReadyReplicas < int32(a.Replicas) && a.deployment.Status.ReadyReplicas != 0) {
		if isHaveTerminatedContainer(a.pods) {
			return SOMEABNORMAL
		}
		if isHaveNormalTerminatedContainer(a.pods) {
			return STOPPING
		}
		return STARTING
	}
	if a.deployment != nil && a.deployment.Status.ReadyReplicas == 0 {
		if isHaveTerminatedContainer(a.pods) {
			return ABNORMAL
		}
		if isHaveNormalTerminatedContainer(a.pods) {
			return STOPPING
		}
		return STARTING
	}
	if a.statefulset != nil && (a.statefulset.Status.ReadyReplicas < int32(a.Replicas) && a.statefulset.Status.ReadyReplicas != 0) {
		if isHaveTerminatedContainer(a.pods) {
			return SOMEABNORMAL
		}
		if isHaveNormalTerminatedContainer(a.pods) {
			return STOPPING
		}
		return STARTING
	}
	if a.statefulset != nil && a.statefulset.Status.ReadyReplicas == 0 {
		if isHaveTerminatedContainer(a.pods) {
			return ABNORMAL
		}
		if isHaveNormalTerminatedContainer(a.pods) {
			return STOPPING
		}
		return STARTING
	}
	return UNKNOW
}

func isHaveTerminatedContainer(pods []*corev1.Pod) bool {
	for _, pod := range pods {
		for _, con := range pod.Status.ContainerStatuses {
			//have Terminated container
			if con.State.Terminated != nil && con.State.Terminated.ExitCode != 0 {
				return true
			}
			if con.LastTerminationState.Terminated != nil {
				return true
			}
		}
	}
	return false
}
func isHaveNormalTerminatedContainer(pods []*corev1.Pod) bool {
	for _, pod := range pods {
		for _, con := range pod.Status.ContainerStatuses {
			//have Terminated container
			if con.State.Terminated != nil && con.State.Terminated.ExitCode == 0 {
				return true
			}
		}
	}
	return false
}

//Ready Whether ready
func (a *AppService) Ready() bool {
	if a.statefulset != nil {
		if a.statefulset.Status.ReadyReplicas >= int32(a.Replicas) {
			return true
		}
	}
	if a.deployment != nil {
		if a.deployment.Status.ReadyReplicas >= int32(a.Replicas) {
			return true
		}
	}
	return false
}

//IsWaitting service status is waitting
//init container init-probe is running
func (a *AppService) IsWaitting() bool {
	var initcontainer []corev1.Container
	if a.statefulset != nil {
		initcontainer = a.statefulset.Spec.Template.Spec.InitContainers
		if len(initcontainer) == 0 {
			return false
		}
	}
	if a.deployment != nil {
		initcontainer = a.deployment.Spec.Template.Spec.InitContainers
		if len(initcontainer) == 0 {
			return false
		}
	}
	var haveProbeInitContainer bool
	for _, init := range initcontainer {
		if init.Image == GetProbeMeshImageName() {
			haveProbeInitContainer = true
			break
		}
	}
	if haveProbeInitContainer {
		if len(a.pods) == 0 {
			return true
		}
		firstPod := a.pods[0]
		for _, initconteir := range firstPod.Status.InitContainerStatuses {
			if initconteir.Image == GetProbeMeshImageName() {
				if initconteir.State.Terminated == nil || initconteir.State.Terminated.ExitCode != 0 {
					return true
				}
			}
		}
	}
	return false
}

//GetReadyReplicas get already ready pod number
func (a *AppService) GetReadyReplicas() int32 {
	if a.statefulset != nil {
		return a.statefulset.Status.ReadyReplicas
	}
	if a.deployment != nil {
		return a.deployment.Status.ReadyReplicas
	}
	return 0
}

//GetRunningVersion get running version
func (a *AppService) GetRunningVersion() string {
	if a.statefulset != nil {
		return a.statefulset.Labels["version"]
	}
	if a.deployment != nil {
		return a.deployment.Labels["version"]
	}
	return ""
}

//UpgradeComlete upgrade comlete
func (a *AppService) UpgradeComlete() bool {
	for _, pod := range a.pods {
		if pod.Labels["version"] != a.DeployVersion {
			return false
		}
	}
	return a.Ready()
}

//AbnormalInfo pod Abnormal info
//Record the container exception exit information in pod.
type AbnormalInfo struct {
	ServiceID     string    `json:"service_id"`
	TenantID      string    `json:"tenant_id"`
	ServiceAlias  string    `json:"service_alias"`
	PodName       string    `json:"pod_name"`
	ContainerName string    `json:"container_name"`
	Reason        string    `json:"reson"`
	Message       string    `json:"message"`
	CreateTime    time.Time `json:"create_time"`
	Count         int       `json:"count"`
}

//Hash get AbnormalInfo hash
func (a AbnormalInfo) Hash() string {
	hash := sha256.New()
	hash.Write([]byte(a.ServiceID + a.ServiceAlias))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func (a AbnormalInfo) String() string {
	return fmt.Sprintf("ServiceID: %s;ServiceAlias: %s;PodName: %s ; ContainerName: %s; Reason: %s; Message: %s",
		a.ServiceID, a.ServiceAlias, a.PodName, a.ContainerName, a.Reason, a.Message)
}
