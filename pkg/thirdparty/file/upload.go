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

package file

import (
	"fmt"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPkgUploadHandler defines the interface of pkg upload.
type IPkgUploadHandler interface {
	// UploadOriginAgent upload origin agent.
	UploadOriginAgent(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginServer upload origin server.
	UploadOriginServer(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginProxy upload origin proxy.
	UploadOriginProxy(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (*types.OriginPkgDetail, error)

	// UploadOriginCert upload origin cert.
	UploadOriginCert(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginCertPkgDetail, error)

	// UploadOriginBinTool upload origin bintool.
	UploadOriginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
		*types.OriginBinToolPkgDetail, error)

	// UploadOriginPluginV2 upload origin plugin v2.
	UploadOriginPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginPluginV2PkgDetail, error)

	// UploadOriginExternalPluginV2 upload origin external plugin v2.
	UploadOriginExternalPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginExternalPluginV2PkgDetail, error)

	// UploadOriginPluginV3 upload origin plugin v3.
	UploadOriginPluginV3(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginPluginV3PkgDetail, error)

	// UploadOriginPluginBinTool upload origin plugin bin tool.
	UploadOriginPluginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginPluginBinToolPkgDetail, error)
}

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin agent upload", gen)
	}

	tenantID := nCtx.TenantID()

	params := &protoFile.UploadOriginAgentReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginAgent(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   resp.ConvertPlatformsToTypes(),
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin server upload", gen)
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginServerReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginServer(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		Version:   resp.GetVersion(),
		Platforms: resp.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginProxy upload origin proxy.
func (h *handler) UploadOriginProxy(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if gen != types.Generation2 {
		return nil, fmt.Errorf("param generateion(%d) invalid, only generation2 is supported for origin proxy upload", gen)
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginProxyReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginProxy(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   resp.ConvertPlatformsToTypes(),
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (*types.OriginCertPkgDetail, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginCertReq{Overwrite: overwrite}
	resp, err := h.cli.uploadOriginCert(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginCertPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		CertFiles: resp.GetCertFiles(),
	}, nil
}

// UploadOriginBinTool upload origin bintool.
func (h *handler) UploadOriginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, gen types.Generation, overwrite bool) (
	*types.OriginBinToolPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginBinToolReq{
		Generation: int64(gen),
		Overwrite:  overwrite,
	}
	resp, err := h.cli.uploadOriginBinTool(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginBinToolPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:       resp.GetUploadId(),
		Existed:        resp.GetExisted(),
		AgentPlatforms: resp.ConvertAgentPlatformsToTypes(),
		ProxyPlatforms: resp.ConvertProxyPlatformsToTypes(),
	}, nil
}

// UploadOriginPluginV2 upload origin plugin v2.
func (h *handler) UploadOriginPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginV2PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginV2Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginPluginV2(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPluginV2PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:           data.GetUploadId(),
		Existed:            data.GetExisted(),
		PluginPkgName:      data.GetPluginPkgName(),
		Version:            data.GetVersion(),
		Description:        data.GetDescription(),
		DescriptionEn:      data.GetDescriptionEn(),
		Scenario:           data.GetScenario(),
		ScenarioEn:         data.GetScenarioEn(),
		ConfigFile:         data.GetConfigFile(),
		ConfigFormat:       data.GetConfigFormat(),
		LaunchNode:         data.GetLaunchNode(),
		InstallCPULimit:    data.GetInstallCpuLimit(),
		InstallDiskIOLimit: data.GetInstallDiskIoLimit(),
		Platforms:          data.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginExternalPluginV2 upload origin external plugin v2.
func (h *handler) UploadOriginExternalPluginV2(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginExternalPluginV2PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginExternalPluginV2Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginExternalPluginV2(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginExternalPluginV2PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:      data.GetUploadId(),
		Existed:       data.GetExisted(),
		PluginPkgName: data.GetPluginPkgName(),
		Version:       data.GetVersion(),
		Description:   data.GetDescription(),
		DescriptionEn: data.GetDescriptionEn(),
		Scenario:      data.GetScenario(),
		ScenarioEn:    data.GetScenarioEn(),
		ConfigFile:    data.GetConfigFile(),
		ConfigFormat:  data.GetConfigFormat(),
		LaunchNode:    data.GetLaunchNode(),
		Platforms:     data.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginPluginV3 upload origin plugin v3.
func (h *handler) UploadOriginPluginV3(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginV3PkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginV3Req{Overwrite: overwrite}
	data, err := h.cli.uploadOriginPluginV3(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginPluginV3PkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID:           data.GetUploadId(),
		Existed:            data.GetExisted(),
		PluginPkgName:      data.GetPluginPkgName(),
		Version:            data.GetVersion(),
		Description:        data.GetDescription(),
		DescriptionEn:      data.GetDescriptionEn(),
		Scenario:           data.GetScenario(),
		ScenarioEn:         data.GetScenarioEn(),
		LaunchNode:         data.GetLaunchNode(),
		TemplateRenderer:   types.TemplateRendererType(data.GetTemplateRenderer()),
		InstallCPULimit:    data.GetInstallCpuLimit(),
		InstallDiskIOLimit: data.GetInstallDiskIoLimit(),
		Platforms:          data.ConvertPlatformsToTypes(),
	}, nil
}

// UploadOriginPluginBinTool upload origin plugin bin tool.
func (h *handler) UploadOriginPluginBinTool(nCtx contextx.IContext, fileName string, file io.Reader, overwrite bool) (
	*types.OriginPluginBinToolPkgDetail, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	params := &protoFile.UploadOriginPluginBinToolReq{
		Overwrite: overwrite,
	}
	data, err := h.cli.uploadOriginPluginBinTool(nCtx, tenantID, params, fileName, file)
	if err != nil {
		return nil, err
	}

	platsV2 := make([]platfmt.Platform, 0)
	for _, plat := range data.GetV2().GetPlatforms() {
		platsV2 = append(platsV2, platfmt.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	platsV3 := make([]platfmt.Platform, 0)
	for _, plat := range data.GetV3().GetPlatforms() {
		platsV3 = append(platsV3, platfmt.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return &types.OriginPluginBinToolPkgDetail{
		FileInfo: fileiface.FileInfo{
			Name: data.GetName(),
			Size: data.GetSize(),
			MD5:  data.GetMd5(),
		},
		UploadID: data.GetUploadId(),
		Existed:  data.GetExisted(),
		V2: types.OriginPluginBinToolPkgV2Info{
			Platforms: platsV2,
		},
		V3: types.OriginPluginBinToolPkgV3Info{
			Platforms: platsV3,
		},
	}, nil
}
