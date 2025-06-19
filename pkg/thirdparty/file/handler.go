/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package file

import (
	"context"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler is interface for backend file handler.
type IHandler interface {
	// UploadOriginAgent upload origin agent.
	UploadOriginAgent(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginPkgDetail, error)

	// UploadOriginServer upload origin server.
	UploadOriginServer(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginPkgDetail, error)

	// UploadOriginCert upload origin cert.
	UploadOriginCert(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginCertPkgDetail, error)

	// UploadOriginBinTool upload origin bintool.
	UploadOriginBinTool(ctx context.Context, fileName string, file io.Reader) (
		*types.OriginBinToolPkgDetail, error)

	// PublishReleaseAgent publish release agent.
	PublishReleaseAgent(ctx context.Context, uploadID string) error

	// PublishReleaseProxy publish release server.
	PublishReleaseProxy(ctx context.Context, uploadID string) error

	// PublishReleaseCert publish release cert.
	PublishReleaseCert(ctx context.Context, uploadID string) error

	// PublishReleaseBinTool publish release bintool.
	PublishReleaseBinTool(ctx context.Context, uploadID string) error
}

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *client.Capability, conf *Config) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginPkgDetail, error) {

	resp, err := h.cli.uploadOriginAgent(
		ctx, &protoFile.UploadOriginAgentReq{Generation: int64(types.Generation2)}, fileName, file)
	if err != nil {
		return nil, err
	}

	plats := make([]platform.Platform, 0)
	for _, plat := range resp.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   plat.GetOsType(),
			Arch: plat.GetCpuArch(),
		})
	}

	return &types.OriginPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:    resp.GetUploadId(),
		Existed:     resp.GetExisted(),
		Version:     resp.GetVersion(),
		Platforms:   plats,
		ChangeLogEN: resp.GetChangelogEn(),
		ChangeLogZH: resp.GetChangelogZh(),
	}, nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginPkgDetail, error) {

	resp, err := h.cli.uploadOriginServer(
		ctx, &protoFile.UploadOriginServerReq{Generation: int64(types.Generation2)}, fileName, file)
	if err != nil {
		return nil, err
	}

	plats := make([]platform.Platform, 0)
	for _, plat := range resp.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   plat.GetOsType(),
			Arch: plat.GetCpuArch(),
		})
	}

	return &types.OriginPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:  resp.GetUploadId(),
		Existed:   resp.GetExisted(),
		Version:   resp.GetVersion(),
		Platforms: plats,
	}, nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginCertPkgDetail, error) {

	resp, err := h.cli.uploadOriginCert(ctx, &protoFile.UploadOriginCertReq{}, fileName, file)
	if err != nil {
		return nil, err
	}

	return &types.OriginCertPkgDetail{
		FileInfo: iface.FileInfo{
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
func (h *handler) UploadOriginBinTool(ctx context.Context, fileName string, file io.Reader) (
	*types.OriginBinToolPkgDetail, error) {

	resp, err := h.cli.uploadOriginBinTool(ctx, &protoFile.UploadOriginBinToolReq{}, fileName, file)
	if err != nil {
		return nil, err
	}

	agentPlats := make([]platform.Platform, 0)
	for _, plat := range resp.GetAgentPlatforms() {
		agentPlats = append(agentPlats, platform.Platform{
			OS:   plat.GetOsType(),
			Arch: plat.GetCpuArch(),
		})
	}
	proxyPlats := make([]platform.Platform, 0)
	for _, plat := range resp.GetProxyPlatforms() {
		proxyPlats = append(proxyPlats, platform.Platform{
			OS:   plat.GetOsType(),
			Arch: plat.GetCpuArch(),
		})
	}

	return &types.OriginBinToolPkgDetail{
		FileInfo: iface.FileInfo{
			Name: resp.GetName(),
			Size: resp.GetSize(),
			MD5:  resp.GetMd5(),
		},
		UploadID:       resp.GetUploadId(),
		Existed:        resp.GetExisted(),
		AgentPlatforms: agentPlats,
		ProxyPlatforms: proxyPlats,
	}, nil
}

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseAgent(ctx, &protoFile.PublishReleaseAgentReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseProxy publish release proxy.
func (h *handler) PublishReleaseProxy(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseProxy(ctx, &protoFile.PublishReleaseProxyReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseCert(ctx, &protoFile.PublishReleaseCertReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(ctx context.Context, uploadID string) error {
	_, err := h.cli.publishReleaseBinTool(ctx, &protoFile.PublishReleaseBinToolReq{
		UploadId: uploadID,
	})
	if err != nil {
		return err
	}

	return nil
}
