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

	IAgent
	IProxy
	IPlugin
	ICert
	IBinTool
	IPluginBinTool
	IPackageEvent
}

// IProxy define the proxy interface.
// nolint: interfacebloat
type IProxy interface {
	// ListReleaseProxy lists proxy releases by page and conditions.
	ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts proxy releases by conditions.
	CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseProxy gets proxy release.
	GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error)

	// DistinctReleaseProxy gets proxy releases distinct.
	DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// DeleteReleaseProxy deletes proxy release.
	DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetReleaseProxyLabelsMany sets many proxy releases labels.
	SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error

	// EnableReleaseProxy enables proxy release active.
	EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables proxy release disactive.
	DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultRelease sets the proxy release as default.
	SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy cancels the proxy release as default.
	CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error
}

// IAgent define the agent interface.
// nolint: interfacebloat
type IAgent interface {
	// ListReleaseAgent lists agent releases by page and conditions.
	ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts agent releases by conditions.
	CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseAgent gets agent release.
	GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error)

	// DistinctReleaseAgent gets agent releases distinct.
	DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// DeleteReleaseAgent deletes agent release.
	DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetReleaseAgentLabelsMany sets many agent release labels.
	SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error

	// EnableReleaseAgent enables agent release active.
	EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables agent release disactive.
	DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets the agent release as default.
	SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent cancels the agent release as default.
	CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error
}

// IPlugin define the plugin interface.
// nolint:interfacebloat
type IPlugin interface {
	// ListReleasePlugin lists plugin releases by page and conditions.
	ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin counts plugin release by conditions.
	CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// DistinctReleasePlugin gets plugin releases distinct.
	DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// GetReleasePlugin gets plugin release.
	GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (*types.ReleasePlugin, error)

	// 	DeleteReleasePlugin deletes plugin release.
	DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// EnableReleasePlugin enables plugin release active.
	EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables plugin release disactive.
	DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetHiddenReleasePlugin sets the plugin release hidden state.
	SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey, isHidden bool) error

	// SetSharedReleasePlugin sets the plugin release shared state.
	SetSharedReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey, isShared bool) error

	// SetSyncedReleasePlugin sets the plugin release synced state.
	SetSyncedReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey, isSynced bool) error

	// SetAsDefaultReleasePlugin sets the plugin release as default.
	SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin cancels the plugin release as default.
	CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// ExistReleasePlugin exist plugin release.
	ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (bool, error)

	// GetReleasePluginDefaultVersion gets release plugin default version by name, generation and platform.
	GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform) (string, error)

	// DistinctNameReleasePlugin gets distinct plugin release names.
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
}

// ICert define the cert interface.
type ICert interface {
	// ListReleaseCert lists cert releases by page and conditions.
	ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error)

	// CountReleaseCert counts cert release by conditions.
	CountReleaseCert(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseCert gets cert release.
	GetReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) (*types.ReleaseCert, error)

	// DeleteReleaseCert deletes cert release.
	DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error
}

// IBinTool define the bintool interface.
type IBinTool interface {
	// ListReleaseBinTool lists bintool releases by page and conditions.
	ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error)

	// CountReleaseBinTool counts bintool release by conditions.
	CountReleaseBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseBinTool gets bintool release.
	GetReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) (*types.ReleaseBinTool, error)

	// DeleteReleaseBinTool deletes bintool release.
	DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error
}

// IPluginBinTool define the plugin bintool interface.
type IPluginBinTool interface {
	// ListReleasePluginBinTool lists plugin bintool release by page and conditions.
	ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.ReleasePluginBinTool, int64, error)

	// CountReleasePluginBinTool counts plugin bintool release by conditions.
	CountReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleasePluginBinTool gets plugin bintool release.
	GetReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) (*types.ReleasePluginBinTool, error)

	// DeleteReleasePluginBinTool deletes plugin bintool release.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error

	// DistinctNameReleasePluginBinTool gets distinct plugin bintool release names.
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
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
