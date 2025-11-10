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
	IDaoPluginBinToolV2
	IDaoPluginBinToolV3
	IDaoPlugin
}

// IDaoAgent defines the agent interface.
type IDaoAgent interface {
	// GetReleaseAgent gets release agent by generation, type, platform and version.
	GetReleaseAgent(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error)

	// UpsertManyReleaseAgent upsert many release agent.
	UpsertManyReleaseAgent(ctx contextx.IContext, releaseAgents []*types.ReleaseAgent) error

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(ctx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error)
}

// IDaoProxy defines the proxy interface.
type IDaoProxy interface {
	// GetReleaseProxy gets release proxy by generation, type, platform and version.
	GetReleaseProxy(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error)

	// UpsertManyReleaseProxy upserts many proxy release.
	UpsertManyReleaseProxy(ctx contextx.IContext, releaseProxys []*types.ReleaseProxy) error
}

// IDaoCert defines the cert interface.
type IDaoCert interface {
	// GetReleaseCert gets release cert.
	GetReleaseCert(ctx contextx.IContext) (*types.ReleaseCert, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(ctx contextx.IContext) (bool, error)

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(ctx contextx.IContext, cert types.ReleaseCert) error

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
	UpsertReleaseBinTool(ctx contextx.IContext, bintool types.ReleaseBinTool) error

	// DeleteReleaseBinTool deletes release bintool.
	DeleteReleaseBinTool(ctx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPluginBinToolV2 defines the plugin bin tool v2 interface.
type IDaoPluginBinToolV2 interface {
	// GetReleasePluginBinToolV2 gets release plugin bintool v2.
	GetReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation) (*types.ReleasePluginBinToolV2, error)

	// ExistReleasePluginBinToolV2 checks if release plugin bintool v2 exists.
	ExistReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinToolV2 upserts release plugin bintool v2.
	UpsertReleasePluginBinToolV2(ctx contextx.IContext, pluginBinTool types.ReleasePluginBinToolV2) error

	// DeleteReleasePluginBinToolV2 deletes release plugin bintool v2.
	DeleteReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPluginBinToolV3 defines the plugin bin tool v3 interface.
type IDaoPluginBinToolV3 interface {
	// GetReleasePluginBinToolV3 gets release plugin bintool v3.
	GetReleasePluginBinToolV3(ctx contextx.IContext, gen types.Generation) (*types.ReleasePluginBinToolV3, error)

	// ExistReleasePluginBinToolV3 checks if release plugin bintool v3 exists.
	ExistReleasePluginBinToolV3(ctx contextx.IContext, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinToolV3 upserts release plugin bintool v3.
	UpsertReleasePluginBinToolV3(ctx contextx.IContext, pluginBinTool types.ReleasePluginBinToolV3) error

	// DeleteReleasePluginBinToolV3 deletes release plugin bintool v3.
	DeleteReleasePluginBinToolV3(ctx contextx.IContext, gen types.Generation, fileName string) error
}

// IDaoPlugin defines the interface of plugin.
type IDaoPlugin interface {
	// GetReleasePlugin gets release plugin.
	GetReleasePlugin(ctx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (
		*types.ReleasePlugin, error)

	// ExistReleasePlugin checks if release plugin exists.
	ExistReleasePlugin(ctx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)

	// UpsertManyReleasePlugin upserts many release plugin.
	UpsertManyReleasePlugin(ctx contextx.IContext, releasePlugins []*types.ReleasePlugin) error
}
