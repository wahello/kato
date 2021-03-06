package dao

import (
	"github.com/gridworkz/kato/api/util/bcode"
	"github.com/gridworkz/kato/db/model"
	"github.com/jinzhu/gorm"
)

// AppConfigGroupDaoImpl
type AppConfigGroupDaoImpl struct {
	DB *gorm.DB
}

//AddModel
func (a *AppConfigGroupDaoImpl) AddModel(mo model.Interface) error {
	configReq, _ := mo.(*model.ApplicationConfigGroup)
	var oldApp model.ApplicationConfigGroup
	if err := a.DB.Where("app_id = ? AND config_group_name = ?", configReq.AppID, configReq.ConfigGroupName).Find(&oldApp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return a.DB.Create(configReq).Error
		}
		return err
	}
	return bcode.ErrApplicationConfigGroupExist
}

//UpdateModel
func (a *AppConfigGroupDaoImpl) UpdateModel(mo model.Interface) error {
	updateReq := mo.(*model.ApplicationConfigGroup)
	return a.DB.Model(&model.ApplicationConfigGroup{}).Where("app_id = ? AND config_group_name = ?", updateReq.AppID, updateReq.ConfigGroupName).Update("enable", updateReq.Enable).Error
}

// GetConfigGroupByID
func (a *AppConfigGroupDaoImpl) GetConfigGroupByID(appID, configGroupName string) (*model.ApplicationConfigGroup, error) {
	var oldApp model.ApplicationConfigGroup
	if err := a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Find(&oldApp).Error; err != nil {
		return nil, err
	}
	return &oldApp, nil
}

func (a *AppConfigGroupDaoImpl) ListByServiceID(sid string) ([]*model.ApplicationConfigGroup, error) {
	var groups []*model.ApplicationConfigGroup
	if err := a.DB.Model(model.ApplicationConfigGroup{}).Select("app_config_group.*").Joins("left join app_config_group_service on app_config_group.app_id = app_config_group_service.app_id and app_config_group.config_group_name = app_config_group_service.config_group_name").
		Where("app_config_group_service.service_id = ? and enable = true", sid).Scan(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// GetConfigGroupsByAppID
func (a *AppConfigGroupDaoImpl) GetConfigGroupsByAppID(appID string, page, pageSize int) ([]*model.ApplicationConfigGroup, int64, error) {
	var oldApp []*model.ApplicationConfigGroup
	offset := (page - 1) * pageSize
	db := a.DB.Where("app_id = ?", appID).Order("create_time desc")

	var total int64
	if err := db.Model(&model.ApplicationConfigGroup{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Limit(pageSize).Offset(offset).Find(&oldApp).Error; err != nil {
		return nil, 0, err
	}
	return oldApp, total, nil
}

//DeleteConfigGroup
func (a *AppConfigGroupDaoImpl) DeleteConfigGroup(appID, configGroupName string) error {
	return a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Delete(model.ApplicationConfigGroup{}).Error
}

// AppConfigGroupServiceDaoImpl
type AppConfigGroupServiceDaoImpl struct {
	DB *gorm.DB
}

//AddModel
func (a *AppConfigGroupServiceDaoImpl) AddModel(mo model.Interface) error {
	configReq, _ := mo.(*model.ConfigGroupService)
	var oldApp model.ConfigGroupService
	if err := a.DB.Where("app_id = ? AND config_group_name = ? AND service_id = ?", configReq.AppID, configReq.ConfigGroupName, configReq.ServiceID).Find(&oldApp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return a.DB.Create(configReq).Error
		}
		return err
	}
	return bcode.ErrConfigGroupServiceExist
}

//UpdateModel
func (a *AppConfigGroupServiceDaoImpl) UpdateModel(mo model.Interface) error {
	return nil
}

// GetConfigGroupServicesByID
func (a *AppConfigGroupServiceDaoImpl) GetConfigGroupServicesByID(appID, configGroupName string) ([]*model.ConfigGroupService, error) {
	var oldApp []*model.ConfigGroupService
	if err := a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Find(&oldApp).Error; err != nil {
		return nil, err
	}
	return oldApp, nil
}

//DeleteConfigGroupService
func (a *AppConfigGroupServiceDaoImpl) DeleteConfigGroupService(appID, configGroupName string) error {
	return a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Delete(model.ConfigGroupService{}).Error
}

//DeleteEffectiveServiceByServiceID
func (a *AppConfigGroupServiceDaoImpl) DeleteEffectiveServiceByServiceID(serviceID string) error {
	return a.DB.Where("service_id = ?", serviceID).Delete(model.ConfigGroupService{}).Error
}

// AppConfigGroupItemDaoImpl
type AppConfigGroupItemDaoImpl struct {
	DB *gorm.DB
}

//AddModel
func (a *AppConfigGroupItemDaoImpl) AddModel(mo model.Interface) error {
	configReq, _ := mo.(*model.ConfigGroupItem)
	var oldApp model.ConfigGroupItem
	if err := a.DB.Where("app_id = ? AND config_group_name = ? AND item_key = ?", configReq.AppID, configReq.ConfigGroupName, configReq.ItemKey).Find(&oldApp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return a.DB.Create(configReq).Error
		}
		return err
	}
	return bcode.ErrConfigItemExist
}

//UpdateModel
func (a *AppConfigGroupItemDaoImpl) UpdateModel(mo model.Interface) error {
	updateReq := mo.(*model.ConfigGroupItem)
	return a.DB.Model(&model.ConfigGroupItem{}).
		Where("app_id = ? AND config_group_name = ? AND item_key = ?", updateReq.AppID, updateReq.ConfigGroupName, updateReq.ItemKey).
		Update("item_value", updateReq.ItemValue).Error
}

// GetConfigGroupItemsByID
func (a *AppConfigGroupItemDaoImpl) GetConfigGroupItemsByID(appID, configGroupName string) ([]*model.ConfigGroupItem, error) {
	var oldApp []*model.ConfigGroupItem
	if err := a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Find(&oldApp).Error; err != nil {
		return nil, err
	}
	return oldApp, nil
}

func (a *AppConfigGroupItemDaoImpl) ListByServiceID(sid string) ([]*model.ConfigGroupItem, error) {
	var items []*model.ConfigGroupItem
	if err := a.DB.Model(model.ConfigGroupItem{}).Select("app_config_group_item.*").Joins("left join app_config_group_service on app_config_group_item.app_id = app_config_group_service.app_id and app_config_group_item.config_group_name = app_config_group_service.config_group_name").
		Where("app_config_group_service.service_id = ?", sid).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

//DeleteConfigGroupItem
func (a *AppConfigGroupItemDaoImpl) DeleteConfigGroupItem(appID, configGroupName string) error {
	return a.DB.Where("app_id = ? AND config_group_name = ?", appID, configGroupName).Delete(model.ConfigGroupItem{}).Error
}
