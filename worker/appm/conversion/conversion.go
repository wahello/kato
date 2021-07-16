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

package conversion

import (
	"fmt"

	"github.com/gridworkz/kato/api/util/bcode"
	"github.com/gridworkz/kato/db"
	"github.com/gridworkz/kato/db/model"
	"github.com/gridworkz/kato/util"
	v1 "github.com/gridworkz/kato/worker/appm/types/v1"
)

func init() {
	//first conv service source
	RegistConversion("ServiceSource", ServiceSource)
	//step2 conv service base
	RegistConversion("TenantServiceBase", TenantServiceBase)
	// convert config group to env secrets
	RegistConversion("TenantServiceConfigGroup", TenantServiceConfigGroup)
	//step3 conv service pod base info
	RegistConversion("TenantServiceVersion", TenantServiceVersion)
	//step4 conv service plugin
	RegistConversion("TenantServicePlugin", TenantServicePlugin)
	//step5 conv service inner and outer regist
	RegistConversion("TenantServiceRegist", TenantServiceRegist)
	//step6 -
	RegistConversion("TenantServiceAutoscaler", TenantServiceAutoscaler)
	//step7 conv service monitor
	RegistConversion("TenantServiceMonitor", TenantServiceMonitor)
}

//Conversion conversion function
//Any application attribute implementation is similarly injected
type Conversion func(*v1.AppService, db.Manager) error

//CacheConversion conversion cache struct
type CacheConversion struct {
	Name       string
	Conversion Conversion
}

//ConversionList conversion function list
var conversionList []CacheConversion

//RegistConversion regist conversion function list
func RegistConversion(name string, fun Conversion) {
	conversionList = append(conversionList, CacheConversion{Name: name, Conversion: fun})
}

//InitAppService init a app service
func InitAppService(dbmanager db.Manager, serviceID string, configs map[string]string, enableConversionList ...string) (*v1.AppService, error) {
	if configs == nil {
		configs = make(map[string]string)
	}

	appService := &v1.AppService{
		AppServiceBase: v1.AppServiceBase{
			ServiceID:      serviceID,
			ExtensionSet:   configs,
			GovernanceMode: model.GovernanceModeBuildInServiceMesh,
		},
		UpgradePatch: make(map[string][]byte, 2),
	}

	// setup governance mode
	app, err := dbmanager.ApplicationDao().GetByServiceID(serviceID)
	if err != nil && err != bcode.ErrApplicationNotFound {
		return nil, fmt.Errorf("get app based on service id(%s)", serviceID)
	}
	if app != nil {
		appService.AppServiceBase.GovernanceMode = app.GovernanceMode
	}

	for _, c := range conversionList {
		if len(enableConversionList) == 0 || util.StringArrayContains(enableConversionList, c.Name) {
			if err := c.Conversion(appService, dbmanager); err != nil {
				return nil, err
			}
		}
	}
	return appService, nil
}

//InitCacheAppService init cache app service.
//If store manager receives a kube model belonging to a service and the model is not found in a store, one will be will be created
func InitCacheAppService(dbm db.Manager, serviceID, creatorID string) (*v1.AppService, error) {
	appService := &v1.AppService{
		AppServiceBase: v1.AppServiceBase{
			ServiceID:      serviceID,
			CreaterID:      creatorID,
			ExtensionSet:   make(map[string]string),
			GovernanceMode: model.GovernanceModeBuildInServiceMesh,
		},
		UpgradePatch: make(map[string][]byte, 2),
	}

	// setup governance mode
	app, err := dbm.ApplicationDao().GetByServiceID(serviceID)
	if err != nil && err != bcode.ErrApplicationNotFound {
		return nil, fmt.Errorf("get app based on service id (%s)", serviceID)
	}
	if app != nil {
		appService.AppServiceBase.GovernanceMode = app.GovernanceMode
	}

	if err := TenantServiceBase(appService, dbm); err != nil {
		return nil, err
	}
	svc, err := dbm.TenantServiceDao().GetServiceByID(serviceID)
	if err != nil {
		return nil, err
	}
	if svc.Kind == model.ServiceKindThirdParty.String() {
		if err := TenantServiceRegist(appService, dbm); err != nil {
			return nil, err
		}
	}

	return appService, nil
}
