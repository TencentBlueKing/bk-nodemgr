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
type IDaoAgent interface {
	// GetReleaseAgent gets release agent by generation, type, platform and version.
	GetReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error)

	// UpsertManyReleaseAgent upsert many release agent.
	UpsertManyReleaseAgent(nCtx contextx.IContext, releaseAgents []*types.ReleaseAgent) error

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoProxy defines the proxy interface.
type IDaoProxy interface {
	// GetReleaseProxy gets release proxy by generation, type, platform and version.
	GetReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error)

	// UpsertManyReleaseProxy upserts many proxy release.
	UpsertManyReleaseProxy(nCtx contextx.IContext, releaseProxys []*types.ReleaseProxy) error

	// ExistReleaseProxy checks if release proxy exists.
	ExistReleaseProxy(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoCert defines the cert interface.
type IDaoCert interface {
	// GetReleaseCert gets release cert.
	GetReleaseCert(nCtx contextx.IContext) (*types.ReleaseCert, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(nCtx contextx.IContext) (bool, error)

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(nCtx contextx.IContext, cert types.ReleaseCert) error

	// DeleteReleaseCert deletes release cert.
	DeleteReleaseCert(nCtx contextx.IContext, fileName string) error
}

// IDaoBinTool defines the bin tool interface.
type IDaoBinTool interface {
	// GetReleaseBinTool gets release bintool.
	GetReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error)

	// ExistReleaseBinTool checks if release bintool exists.
	ExistReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleaseBinTool upserts release bintool.
	UpsertReleaseBinTool(nCtx contextx.IContext, bintool types.ReleaseBinTool) error

	// DeleteReleaseBinTool deletes release bintool.
	DeleteReleaseBinTool(nCtx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPluginBinTool defines the plugin bin tool interface.
type IDaoPluginBinTool interface {
	// GetReleasePluginBinTool gets release plugin bintool.
	GetReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*types.ReleasePluginBinTool, error)

	// ExistReleasePluginBinTool checks if release plugin bintool exists.
	ExistReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinTool upserts release plugin bintool.
	UpsertReleasePluginBinTool(nCtx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error

	// DeleteReleasePluginBinTool deletes release plugin bintool.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPlugin defines the interface of plugin.
type IDaoPlugin interface {
	// ListReleasePlugin lists plugin releases by page and conditions.
	ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error)

	// GetReleasePlugin gets release plugin.
	GetReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (*types.ReleasePlugin, error)

	// ExistReleasePlugin checks if release plugin exists.
	ExistReleasePlugin(nCtx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleasePlugin upserts many release plugin.
	UpsertManyReleasePlugin(nCtx contextx.IContext, releasePlugins []*types.ReleasePlugin) error
}
