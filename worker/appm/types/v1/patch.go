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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gridworkz/kato/util"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

//SetUpgradePatch create and set upgrade patch for deployment and statefulset
func (a *AppService) SetUpgradePatch(new *AppService) error {
	if a.statefulset != nil && new.statefulset != nil {
		// If the controller originally had a startup sequence, then the startup sequence needs to be updated
		if isContainsBootSequence(a.statefulset.Spec.Template.Spec.InitContainers) &&
			!isContainsBootSequence(new.statefulset.Spec.Template.Spec.InitContainers) && new.BootSeqContainer != nil {
			new.statefulset.Spec.Template.Spec.InitContainers = append(new.statefulset.Spec.Template.Spec.InitContainers, *new.BootSeqContainer)
		}
		statefulsetPatch, err := getStatefulsetModifiedConfiguration(a.statefulset, new.statefulset)
		if err != nil {
			return err
		}
		if len(statefulsetPatch) == 0 {
			return fmt.Errorf("no upgrade")
		}
		logrus.Debugf("stateful patch %s", string(statefulsetPatch))
		new.UpgradePatch["statefulset"] = statefulsetPatch
	}
	if a.deployment != nil && new.deployment != nil {
		// If the controller originally had a startup sequence, then the startup sequence needs to be updated
		if isContainsBootSequence(a.deployment.Spec.Template.Spec.InitContainers) &&
			!isContainsBootSequence(new.deployment.Spec.Template.Spec.InitContainers) && new.BootSeqContainer != nil {
			new.deployment.Spec.Template.Spec.InitContainers = append(new.deployment.Spec.Template.Spec.InitContainers, *new.BootSeqContainer)
		}
		deploymentPatch, err := getDeploymentModifiedConfiguration(a.deployment, new.deployment)
		if err != nil {
			return err
		}
		if len(deploymentPatch) == 0 {
			return fmt.Errorf("no upgrade")
		}
		new.UpgradePatch["deployment"] = deploymentPatch
	}
	//update cache app service base info by new app service
	a.AppServiceBase = new.AppServiceBase
	return nil
}

//EncodeNode
type EncodeNode struct {
	body  []byte
	value []byte
	Field map[string]EncodeNode
}

//UnmarshalJSON custom yaml decoder
func (e *EncodeNode) UnmarshalJSON(code []byte) error {
	e.body = code
	if len(code) < 1 {
		return nil
	}
	if code[0] != '{' {
		e.value = code
		return nil
	}
	var fields = make(map[string]EncodeNode)
	if err := json.Unmarshal(code, &fields); err != nil {
		return err
	}
	e.Field = fields
	return nil
}

//MarshalJSON custom marshal json
func (e *EncodeNode) MarshalJSON() ([]byte, error) {
	if e.value != nil {
		return e.value, nil
	}
	if e.Field != nil {
		var buffer = bytes.NewBufferString("{")
		count := 0
		length := len(e.Field)
		for k, v := range e.Field {
			buffer.WriteString(fmt.Sprintf("\"%s\":", k))
			value, err := v.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buffer.Write(value)
			count++
			if count < length {
				buffer.WriteString(",")
			}
		}
		buffer.WriteByte('}')
		return buffer.Bytes(), nil
	}
	return nil, fmt.Errorf("marshal error")
}

//Contrast Compare value
func (e *EncodeNode) Contrast(endpoint *EncodeNode) bool {
	return util.BytesSliceEqual(e.value, endpoint.value)
}

//GetChange get change fields
func (e *EncodeNode) GetChange(endpoint *EncodeNode) *EncodeNode {
	if util.BytesSliceEqual(e.body, endpoint.body) {
		return nil
	}
	return getChange(*e, *endpoint)
}

func getChange(old, new EncodeNode) *EncodeNode {
	var result EncodeNode
	if util.BytesSliceEqual(old.body, new.body) {
		return nil
	}
	if old.Field == nil && new.Field == nil {
		if !util.BytesSliceEqual(old.value, new.value) {
			result.value = new.value
			return &result
		}
	}
	for k, v := range new.Field {
		if result.Field == nil {
			result.Field = make(map[string]EncodeNode)
		}
		if value := getChange(old.Field[k], v); value != nil {
			result.Field[k] = *value
		}
	}
	return &result
}

//stateful label can not be patched
func getStatefulsetModifiedConfiguration(old, new *v1.StatefulSet) ([]byte, error) {
	old.Status = new.Status
	oldNeed := getAllowFields(old)
	newNeed := getAllowFields(new)
	return getchange(oldNeed, newNeed)
}

// updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden.
func getAllowFields(s *v1.StatefulSet) *v1.StatefulSet {
	return &v1.StatefulSet{
		Spec: v1.StatefulSetSpec{
			Replicas:       s.Spec.Replicas,
			Template:       s.Spec.Template,
			UpdateStrategy: s.Spec.UpdateStrategy,
		},
	}
}

func getDeploymentModifiedConfiguration(old, new *v1.Deployment) ([]byte, error) {
	old.Status = new.Status
	return getchange(old, new)
}

func getchange(old, new interface{}) ([]byte, error) {
	oldbuffer := bytes.NewBuffer(nil)
	newbuffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(oldbuffer).Encode(old)
	if err != nil {
		return nil, fmt.Errorf("encode old body error %s", err.Error())
	}
	err = json.NewEncoder(newbuffer).Encode(new)
	if err != nil {
		return nil, fmt.Errorf("encode new body error %s", err.Error())
	}
	var en EncodeNode
	if err := json.NewDecoder(oldbuffer).Decode(&en); err != nil {
		return nil, err
	}
	var ennew EncodeNode
	if err := json.NewDecoder(newbuffer).Decode(&ennew); err != nil {
		return nil, err
	}
	change := en.GetChange(&ennew)
	changebody, err := json.Marshal(change)
	if err != nil {
		return nil, err
	}
	return changebody, nil
}

func isContainsBootSequence(initContainers []corev1.Container) bool {
	for _, initContainer := range initContainers {
		if strings.Contains(initContainer.Name, "probe-mesh-") {
			return true
		}
	}
	return false
}
