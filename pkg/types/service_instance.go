/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package types

import "time"

// ServiceInstance defines the service instance.
type ServiceInstance struct {
	ID                int64
	Name              string
	Labels            map[string]string
	Processes         map[string]ServiceInstanceProcess
	BizID             int64
	HostID            int64
	ModuleID          int64
	ServiceTemplateID int64
	ServiceCategoryID int64
}

// ServiceInstanceProcess defines the process rendered from a service instance.
type ServiceInstanceProcess struct {
	AutoStart       bool
	BizID           int64
	FuncName        string
	ProcessID       int64
	ProcessName     string
	StartParamRegex string
	SupplierAccount string
	CreateTime      time.Time
	LastTime        time.Time
	Description     string
	FaceStopCmd     string
	PidFile         string
	Priority        int64
	ProcNum         int64
	ReloadCmd       string
	RestartCmd      string
	StartCmd        string
	StopCmd         string
	Timeout         int64
	User            string
	WorkPath        string
	CreateAt        string
	CreateBy        string
	UpdateAt        string
	UpdateBy        string
	BindInfo        []ServiceInstanceProcessBindInfo
}

// ServiceInstanceProcessBindInfo defines the bind info of a service instance process.
type ServiceInstanceProcessBindInfo struct {
	Enable        bool
	IP            string
	Port          string
	Protocol      string
	TemplateRowID int64
}
