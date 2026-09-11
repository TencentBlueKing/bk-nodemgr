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

// IStorage defines the interface of release storage.
// nolint: interfacebloat
type IStorage interface {
	basestorage.Interface

	IDaoAgent
	IDaoProxy
	IDaoCert
	IDaoBinTool
	IDaoPluginBinTool
	IDaoPlugin
}

// IDaoAgent defines the agent interface.
// nolint: interfacebloat // Keep release queries and mutations together by package type.
type IDaoAgent interface {
	// ListReleaseAgent lists agent releases by page and conditions.
	ListReleaseAgent(nCtx contextx.IContext, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts agent releases by conditions.
	CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// DistinctReleaseAgent gets distinct agent release fields.
	DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField,
		conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)

	// SetReleaseAgentLabelsMany updates labels for matching releases.
	SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error

	// DeleteReleaseAgent deletes release metadata identified by its key.
	DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// EnableReleaseAgent enables an agent release.
	EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables an agent release.
	DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets an agent release as default.
	SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent cancels an agent release default.
	CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// GetReleaseAgent gets release agent by generation, type, platform and version.
	GetReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error)

	// UpsertManyReleaseAgent upsert many release agent.
	UpsertManyReleaseAgent(nCtx contextx.IContext, releaseAgents []*types.ReleaseAgent) error

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoProxy defines the proxy interface.
// nolint: interfacebloat // Keep release queries and mutations together by package type.
type IDaoProxy interface {
	// ListReleaseProxy lists proxy releases by page and conditions.
	ListReleaseProxy(nCtx contextx.IContext, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts proxy releases by conditions.
	CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// DistinctReleaseProxy gets distinct proxy release fields.
	DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField,
		conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)

	// SetReleaseProxyLabelsMany updates labels for matching releases.
	SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error

	// DeleteReleaseProxy deletes release metadata identified by its key.
	DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// EnableReleaseProxy enables a proxy release.
	EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables a proxy release.
	DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultReleaseProxy sets a proxy release as default.
	SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy cancels a proxy release default.
	CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// GetReleaseProxy gets release proxy by generation, type, platform and version.
	GetReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error)

	// UpsertManyReleaseProxy upserts many proxy release.
	UpsertManyReleaseProxy(nCtx contextx.IContext, releaseProxys []*types.ReleaseProxy) error

	// ExistReleaseProxy checks if release proxy exists.
	ExistReleaseProxy(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoCert defines the cert interface.
type IDaoCert interface {
	// ListReleaseCert lists cert releases by page and conditions.
	ListReleaseCert(nCtx contextx.IContext, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error)

	// DeleteReleaseCert deletes release metadata identified by its key.
	DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error

	// GetReleaseCert gets release cert.
	GetReleaseCert(nCtx contextx.IContext) (*types.ReleaseCert, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(nCtx contextx.IContext) (bool, error)

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(nCtx contextx.IContext, cert types.ReleaseCert) error
}

// IDaoBinTool defines the bin tool interface.
type IDaoBinTool interface {
	// ListReleaseBinTool lists bintool releases by page and conditions.
	ListReleaseBinTool(nCtx contextx.IContext, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error)

	// DeleteReleaseBinTool deletes release metadata identified by its key.
	DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error

	// GetReleaseBinTool gets release bintool.
	GetReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error)

	// ExistReleaseBinTool checks if release bintool exists.
	ExistReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleaseBinTool upserts release bintool.
	UpsertReleaseBinTool(nCtx contextx.IContext, bintool types.ReleaseBinTool) error
}

// IDaoPluginBinTool defines the plugin bin tool interface.
type IDaoPluginBinTool interface {
	// ListReleasePluginBinTool lists plugin bintool releases by page and conditions.
	ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
		conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error)

	// DeleteReleasePluginBinTool deletes release metadata identified by its key.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error
	// DistinctNameReleasePluginBinTool gets distinct plugin bintool release names.
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// GetReleasePluginBinTool gets release plugin bintool.
	GetReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*types.ReleasePluginBinTool, error)

	// ExistReleasePluginBinTool checks if release plugin bintool exists.
	ExistReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinTool upserts release plugin bintool.
	UpsertReleasePluginBinTool(nCtx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error
}

// IDaoPlugin defines the interface of plugin.
// nolint: interfacebloat // Keep release queries and mutations together by package type.
type IDaoPlugin interface {
	// EnableReleasePlugin enables a plugin release.
	EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables a plugin release.
	DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetAsDefaultReleasePlugin sets a plugin release as default.
	SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin cancels a plugin release default.
	CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetHiddenReleasePlugin hides a plugin release.
	SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelHiddenReleasePlugin unhides a plugin release.
	CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DeleteReleasePlugin deletes release metadata identified by its key.
	DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// ListReleasePlugin lists plugin releases by page and conditions.
	ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error)
	// CountReleasePlugin counts plugin releases by conditions.
	CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// DistinctReleasePlugin gets distinct plugin release fields.
	DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField,
		conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)

	// GetReleasePluginDefaultVersion gets the default plugin version.
	GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform) (string, error)
	// DistinctNameReleasePlugin gets distinct plugin release names.
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// GetReleasePlugin gets release plugin.
	GetReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (*types.ReleasePlugin, error)

	// ExistReleasePlugin checks if release plugin exists.
	ExistReleasePlugin(nCtx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleasePlugin upserts many release plugin.
	UpsertManyReleasePlugin(nCtx contextx.IContext, releasePlugins []*types.ReleasePlugin) error
}
