package dao

import "github.com/gridworkz/kato/db/model"

func (t *TenantServicesDaoImpl) ListByAppID(appID string) ([]*model.TenantServices, error) {
	var services []*model.TenantServices
	if err := t.DB.Where("app_id=?", appID).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}
