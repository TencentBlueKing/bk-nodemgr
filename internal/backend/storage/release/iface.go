/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface for release storage.
type IStorage interface {
	basestorage.Interface

	IRelease
	IAgent
	IProxy
	IPlugin
	IPackageEvent
}

// IRelease define the release interface.
// nolint:interfacebloat
type IRelease interface {
	// GetRelease gets release by generation, release type, platform and version.
	GetRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
		name string, plat platform.Platform, version string) (*types.Release, error)

	// ListRelease lists release by page and conditions.
	ListRelease(nCtx contextx.IContext, releaseType types.ReleaseType, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.Release, int64, error)

	// DistinctRelease distincts release by conditions.
	DistinctRelease(nCtx contextx.IContext, releaseType types.ReleaseType, distinctField types.ReleaseDistinctField,
		conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)

	// CountRelease counts release by conditions.
	CountRelease(
		nCtx contextx.IContext, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition) (int64, error)

	// SetReleaseLabels sets release labels.
	SetReleaseLabels(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
		plat platform.Platform, version string, labels []string) error

	// SetReleaseLabelsMany sets many release labels.
	SetReleaseLabelsMany(nCtx contextx.IContext, releaseType types.ReleaseType, gens []types.Generation,
		plats []platform.Platform, versions []string, labels []string) error

	// EnableRelease enables release active by generation, release type, platform and version.
	EnableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType, plat platform.Platform,
		version string) error

	// DisableRelease disables release disactive by generation, release type, platform and version.
	DisableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType, plat platform.Platform,
		version string) error

	// SetAsDefaultRelease sets the release as default.
	SetAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
		plat platform.Platform, version string) error

	// CancelAsDefaultRelease cancels the release as default.
	CancelAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
		plat platform.Platform, version string) error

	// 	DeleteRelease deletes release by generation, release type, platform and version.
	DeleteRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType, plat platform.Platform,
		version string) error
}

// IProxy define the proxy interface.
type IProxy interface {
	// GetReleaseProxy gets release by generation, release type, platform and version.
	GetReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (
		*types.ReleaseProxy, error)
}

// IAgent define the agent interface.
type IAgent interface {
	// GetReleaseAgent gets release by generation, release type, platform and version.
	GetReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (
		*types.ReleaseAgent, error)
}

// IPlugin define the plugin interface.
type IPlugin interface {
	// GetReleasePlugin gets release by generation, release type, platform and version.
	GetReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (*types.ReleasePlugin, error)
}

// IPackageEvent define the package event interface.
type IPackageEvent interface {
	// CountPackageEvent counts package events by conditions.
	CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error)

	// ListPackageEvent lists package events by page and conditions.
	ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
		[]*types.PackageEvent, int64, error)

	// CreateManyPackageEvent creates multiple package events.
	CreateManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error

	// DistinctPackageEvent distincts package event fields.
	DistinctPackageEvent(
		nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
		*types.PackageEventDistinctResult, error)
}
