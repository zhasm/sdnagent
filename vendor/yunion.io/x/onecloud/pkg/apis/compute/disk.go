// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"time"

	"yunion.io/x/onecloud/pkg/apis"
	"yunion.io/x/onecloud/pkg/apis/billing"
)

type DiskCreateInput struct {
	apis.VirtualResourceCreateInput

	*DiskConfig

	// 此参数仅适用于未指定storage时进行调度到指定区域创建磁盘
	// required: false
	PreferRegion string `json:"prefer_region_id"`

	// 此参数仅适用于未指定storage时进行调度到指定可用区区创建磁盘
	// required: false
	PreferZone string `json:"prefer_zone_id"`

	// swagger:ignore
	PreferWire string `json:"prefer_wire_id"`

	// 此参数仅适用于未指定storage时进行调度到指定可用区区创建磁盘
	// required: false
	PreferHost string `json:"prefer_host_id"`

	// 此参数仅适用于未指定storage时进行调度到指定平台创建磁盘
	// default: kvm
	// enum: kvm, openstack, esxi, aliyun, aws, qcloud, azure, huawei, openstack, ucloud, zstack google, ctyun
	Hypervisor string `json:"hypervisor"`
}

// ToServerCreateInput used by disk schedule
func (req *DiskCreateInput) ToServerCreateInput() *ServerCreateInput {
	input := ServerCreateInput{
		ServerConfigs: &ServerConfigs{
			PreferRegion: req.PreferRegion,
			PreferZone:   req.PreferZone,
			PreferWire:   req.PreferWire,
			PreferHost:   req.PreferHost,
			Hypervisor:   req.Hypervisor,
			Disks:        []*DiskConfig{req.DiskConfig},
			// Project:      req.Project,
			// Domain:       req.Domain,
		},
	}
	input.Name = req.Name
	input.Project = req.Project
	input.ProjectDomain = req.ProjectDomain
	return &input
}

func (req *ServerCreateInput) ToDiskCreateInput() *DiskCreateInput {
	input := DiskCreateInput{
		DiskConfig:   req.Disks[0],
		PreferRegion: req.PreferRegion,
		PreferHost:   req.PreferHost,
		PreferZone:   req.PreferZone,
		PreferWire:   req.PreferWire,
		Hypervisor:   req.Hypervisor,
	}
	input.Name = req.Name
	input.Project = req.Project
	input.ProjectDomain = req.ProjectDomain
	return &input
}

type SnapshotPolicyResourceInput struct {
	// filter disk by snapshotpolicy
	Snapshotpolicy string `json:"snapshotpolicy"`
	// swagger:ignore
	// Deprecated
	// filter disk by snapshotpolicy_id
	SnapshotpolicyId string `json:"snapshotpolicy_id" yunion-deprecated-by:"snapshotpolicy"`
}

type SnapshotPolicyFilterListInput struct {
	SnapshotPolicyResourceInput

	// 以快照策略名称排序
	OrderBySnapshotpolicy string `json:"order_by_snapshotpolicy"`
}

type DiskListInput struct {
	apis.VirtualResourceListInput
	apis.ExternalizedResourceBaseListInput
	billing.BillingResourceListInput
	StorageFilterListInput

	SnapshotPolicyFilterListInput
	ServerFilterListInput

	// filter disk by whether it is being used
	Unused *bool `json:"unused"`

	// swagger:ignore
	// Deprecated
	// filter by disk type
	Type string `json:"type" yunion-deprecated-by:"disk_type"`
	// 过滤指定disk_type的磁盘列表，可能的值为：sys, data, swap. volume
	//
	// | disk_type值 | 说明 |
	// |-------------|----------|
	// | sys         | 虚拟机系统盘    |
	// | data        | 虚拟机数据盘    |
	// | swap        | 虚拟机内存交换盘 |
	// | volume      | 容器volumn盘   |
	//
	DiskType string `json:"disk_type"`

	DiskFormat string `json:"disk_format"`

	DiskSize int `json:"disk_size"`

	AutoDelete *bool `json:"auto_delete"`

	FsFormat string `json:"fs_format"`

	// 镜像
	Template string `json:"template"`
	// swagger:ignore
	// Deprecated
	TemplateId string `json:"template_id" yunion-deprecated-by:"template"`

	// 快照
	Snapshot string `json:"snapshot"`
	// swagger:ignore
	// Deprecated
	SnapshotId string `json:"snapshot_id" yunion-deprecated-by:"snapshot"`
}

type DiskResourceInput struct {
	// 虚拟磁盘（ID或Name）
	Disk string `json:"disk"`
	// swagger:ignore
	// Deprecated
	// filter by disk_id
	DiskId string `json:"disk_id" yunion-deprecated-by:"disk"`
}

type DiskFilterListInputBase struct {
	DiskResourceInput

	// 以磁盘名称排序
	// pattern:asc|desc
	OrderByDisk string `json:"order_by_disk"`
}

type DiskFilterListInput struct {
	StorageFilterListInput

	DiskFilterListInputBase
}

type SimpleGuest struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Status string `json:"status"`
}

type SimpleSnapshotPolicy struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	RepeatWeekdays []int  `json:"repeat_weekdays"`
	TimePoints     []int  `json:"time_points"`
}

type DiskDetails struct {
	apis.VirtualResourceDetails
	StorageResourceInfo

	SDisk

	// 所挂载的虚拟机
	Guests []SimpleGuest `json:"guests"`
	// 所挂载的虚拟机
	Guest string `json:"guest"`
	// 所挂载虚拟机的数量
	GuestCount int `json:"guest_count"`
	// 所挂载虚拟机状态
	GuestStatus string `json:"guest_status"`

	// 自动清理时间
	AutoDeleteAt time.Time `json:"auto_delete_at"`
	// 自动快照策略状态
	SnapshotpolicyStatus string `json:"snapshotpolicy_status,allowempty"`

	// 自动快照策略
	Snapshotpolicies []SimpleSnapshotPolicy `json:"snapshotpolicies"`

	// 手动快照数量
	ManualSnapshotCount int `json:"manual_snapshot_count"`
	// 最多可创建手动快照数量
	MaxManualSnapshotCount int `json:"max_manual_snapshot_count"`
}

type DiskResourceInfoBase struct {
	// 磁盘名称
	Disk string `json:"disk"`
}

type DiskResourceInfo struct {
	DiskResourceInfoBase

	// 存储ID
	StorageId string `json:"storage_id"`

	StorageResourceInfo
}

type DiskSyncstatusInput struct {
}

type DiskUpdateInput struct {
	apis.VirtualResourceBaseUpdateInput

	// 磁盘类型
	DiskType string `json:"disk_type"`
}
