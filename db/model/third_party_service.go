// Copyright (C) 2021 Gridworkz Co., Ltd.
// KATO, Application Management Platform

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

package model

// Endpoint is a persistent object for table 3rd_party_svc_endpoints.
type Endpoint struct {
	Model
	UUID      string `gorm:"column:uuid;size:32" json:"uuid"`
	ServiceID string `gorm:"column:service_id;size:32;not null" json:"service_id"`
	IP        string `gorm:"column:ip;not null" json:"ip"`
	Port      int    `gorm:"column:port;size:65535" json:"port"`
	//use pointer type, zero values won't be saved into database
	IsOnline *bool `gorm:"column:is_online;default:true" json:"is_online"`
}

// TableName returns table name of Endpoint.
func (Endpoint) TableName() string {
	return "tenant_service_3rd_party_endpoints"
}

// DiscorveryType type of service discovery center.
type DiscorveryType string

// DiscorveryTypeEtcd etcd
var DiscorveryTypeEtcd DiscorveryType = "etcd"

func (d DiscorveryType) String() string {
	return string(d)
}

// ThirdPartySvcDiscoveryCfg s a persistent object for table
// 3rd_party_svc_discovery_cfg. 3rd_party_svc_discovery_cfg contains
// service discovery center configuration for third party service.
type ThirdPartySvcDiscoveryCfg struct {
	Model
	ServiceID string `gorm:"column:service_id;size:32"`
	Type      string `gorm:"column:type"`
	Servers   string `gorm:"column:servers"`
	Key       string `gorm:"key"`
	Username  string `gorm:"username"`
	Password  string `gorm:"password"`
}

// TableName returns table name of ThirdPartySvcDiscoveryCfg.
func (ThirdPartySvcDiscoveryCfg) TableName() string {
	return "tenant_service_3rd_party_discovery_cfg"
}
