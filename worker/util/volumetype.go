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

package util

import (
	"encoding/json"
	"fmt"
	"strings"

	dbmodel "github.com/gridworkz/kato/db/model"
	v1 "github.com/gridworkz/kato/worker/appm/types/v1"
	storagev1 "k8s.io/api/storage/v1"
)

var defaultcapacityValidation map[string]interface{}
var defaultAccessMode = []string{"RWO"}
var defaultBackupPolicy = []string{"exclusive"}
var defaultSharePolicy = []string{"exclusive"}

func init() {
	defaultcapacityValidation = make(map[string]interface{})
	defaultcapacityValidation["min"] = 0
	defaultcapacityValidation["required"] = true
	defaultcapacityValidation["max"] = 999999999
}

// TransStorageClass2RBDVolumeType transfer k8s storageclass 2 rbd volumeType
func TransStorageClass2RBDVolumeType(sc *storagev1.StorageClass) *dbmodel.TenantServiceVolumeType {
	if sc.GetName() == v1.KatoStatefuleShareStorageClass {
		return &dbmodel.TenantServiceVolumeType{VolumeType: dbmodel.ShareFileVolumeType.String()}
	}
	if sc.GetName() == v1.KatoStatefuleLocalStorageClass {
		return &dbmodel.TenantServiceVolumeType{VolumeType: dbmodel.LocalVolumeType.String()}
	}
	scbs, _ := json.Marshal(sc)
	cvbs, _ := json.Marshal(defaultcapacityValidation)

	volumeType := &dbmodel.TenantServiceVolumeType{
		VolumeType:         sc.GetName(),
		NameShow:           sc.GetName(),
		CapacityValidation: string(cvbs),
		StorageClassDetail: string(scbs),
		Provisioner:        sc.Provisioner,
		AccessMode:         strings.Join(defaultAccessMode, ","),
		BackupPolicy:       strings.Join(defaultBackupPolicy, ","),
		SharePolicy:        strings.Join(defaultSharePolicy, ","),
		Sort:               999,
		Enable:             true,
	}
	volumeType.ReclaimPolicy = "Retain"
	if sc.ReclaimPolicy != nil {
		volumeType.ReclaimPolicy = fmt.Sprintf("%v", *sc.ReclaimPolicy)
	}
	if sc.Annotations != nil {
		if name, ok := sc.Annotations["rbd_volume_name"]; ok {
			volumeType.NameShow = name
		}
	}
	return volumeType
}

// ValidateVolumeCapacity
func ValidateVolumeCapacity(validation string, capacity int64) error {
	validator := make(map[string]interface{})
	if err := json.Unmarshal([]byte(validation), &validator); err != nil {
		return err
	}

	if min, ok := validator["min"].(int64); ok {
		if capacity < min {
			return fmt.Errorf("volume capacity %v less than min value %v", capacity, min)
		}
	}

	if max, ok := validator["max"].(int64); ok {
		if capacity > max {
			return fmt.Errorf("volume capacity %v more than max value %v", capacity, max)
		}
	}

	return nil
}
