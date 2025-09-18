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

// IStorage defines the interface of release storage.
// nolint: interfacebloat
type IStorage interface {
	basestorage.Interface

	IDaoAgent
	IDaoProxy
	IDaoBinTool
	IDaoCert
	IDaoPluginBinTool
	IDaoOfficialPlugin
	IDaoExternalPlugin
}

// IDaoAgent defines the agent interface.
type IDaoAgent interface {
	// GetReleaseAgent gets release agent by generation, type, platform and version.
	GetReleaseAgent(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error)

	// UpsertManyReleaseAgent upsert many release agent.
	UpsertManyReleaseAgent(ctx contextx.IUserContext, releaseAgents []*types.ReleaseAgent) error

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(ctx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoProxy defines the proxy interface.
type IDaoProxy interface {
	// GetReleaseProxy gets release proxy by generation, type, platform and version.
	GetReleaseProxy(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error)

	// UpsertManyReleaseProxy upserts many proxy release.
	UpsertManyReleaseProxy(ctx contextx.IUserContext, releaseProxys []*types.ReleaseProxy) error
}

// IDaoCert defines the cert interface.
type IDaoCert interface {
	// GetReleaseCert gets release cert.
	GetReleaseCert(ctx contextx.IContext) (*types.ReleaseCert, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(ctx contextx.IContext) (bool, error)

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(ctx contextx.IUserContext, cert types.ReleaseCert) error

	// DeleteReleaseCert deletes release cert.
	DeleteReleaseCert(ctx contextx.IContext, fileName string) error
}

// IDaoBinTool defines the bin tool interface.
type IDaoBinTool interface {
	// GetReleaseBinTool gets release bintool.
	GetReleaseBinTool(ctx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error)

	// ExistReleaseBinTool checks if release bintool exists.
	ExistReleaseBinTool(ctx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleaseBinTool upserts release bintool.
	UpsertReleaseBinTool(ctx contextx.IUserContext, bintool types.ReleaseBinTool) error

	// DeleteReleaseBinTool deletes release bintool.
	DeleteReleaseBinTool(ctx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPluginBinTool defines the bin tool interface.
type IDaoPluginBinTool interface {
	// GetReleasePluginBinTool gets release plugin bintool.
	GetReleasePluginBinTool(ctx contextx.IContext, gen types.Generation) (*types.ReleasePluginBinTool, error)

	// ExistReleasePluginBinTool checks if release plugin bintool exists.
	ExistReleasePluginBinTool(ctx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinTool upserts release plugin bintool.
	UpsertReleasePluginBinTool(ctx contextx.IUserContext, pluginBinTool types.ReleasePluginBinTool) error

	// DeleteReleasePluginBinTool deletes release plugin bintool.
	DeleteReleasePluginBinTool(ctx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoOfficialPlugin defines the interface of official plugin.
type IDaoOfficialPlugin interface {
	// GetReleaseOfficialPlugin gets release plugin official plugin.
	GetReleaseOfficialPlugin(ctx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (
		*types.ReleaseOfficialPlugin, error)

	// ExistReleaseOfficialPlugin checks if release official plugin exists.
	ExistReleaseOfficialPlugin(ctx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleaseOfficialPlugin upserts many release official plugin.
	UpsertManyReleaseOfficialPlugin(ctx contextx.IUserContext, releaseOfficialPlugins []*types.ReleaseOfficialPlugin) error
}

// IDaoExternalPlugin defines the interface of external plugin.
type IDaoExternalPlugin interface {
	// GetReleaseExternalPlugin gets release external plugin.
	GetReleaseExternalPlugin(ctx contextx.IContext, pluginName string, gen types.Generation, plat platform.Platform, version string) (
		*types.ReleaseExternalPlugin, error)

	// ExistReleaseExternalPlugin checks if release external plugin exists.
	ExistReleaseExternalPlugin(ctx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleaseExternalPlugin upserts many release external plugin.
	UpsertManyReleaseExternalPlugin(ctx contextx.IUserContext, releaseExternalPlugins []*types.ReleaseExternalPlugin) error
}
