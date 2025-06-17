/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UploadOriginAgent uploads the origin agent.
// nolint:funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginAgent(
	ctx context.Context, gen types.Generation, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {

	// validation.
	if err := gen.Validate(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. invalid generation. err: %v", err)

		return nil, err
	}
	if gen == types.Generation1 {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. generation 1 is not supported")

		return nil, errors.New("generation 1 is not supported")
	}
	if pkgFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, pkgFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkGen2OriginAgentPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to check origin agent package. err: %v", err)

		return nil, err
	}

	pkgName, err := nodepkg.FormatPkgName(
		gen,
		types.ReleaseTypeOriginAgent,
		platform.EmptyPlatform(),
		detail.Version,
	)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package, failed to format package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get temp file. err: %v", err)

		return nil, err
	}

	// upload to upstream.
	if err := m.upstreamOriginAgent.Store(ctx, iface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package, failed to upload to upstream. err: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginAgent.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	info, err := file.Info(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get file info. err: %v", err)

		return nil, err
	}
	detail.FileInfo = info

	// save to db.
	err = m.daoRelease.UpsertMany(ctx, &types.Release{
		Generation: gen,
		Type:       types.ReleaseTypeOriginAgent,
		Version:    detail.Version,
		Platform:   platform.EmptyPlatform(),
		FileName:   info.Name,
		MD5:        info.MD5,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to save to db. err: %v", err)

		return nil, err
	}

	m.logger.InfoCtxf(ctx,
		"uploaded origin agent package to upstream. generation(%d), version(%s), file-name(%s)",
		gen, detail.Version, pkgName)

	return detail, nil
}

// checkGen2OriginAgentPkg check gen2 origin agent package.
func checkGen2OriginAgentPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader. err: %w", err)
	}
	defer func() {
		_ = gzr.Close()
		_ = file.Close()
	}()

	detail := new(types.OriginPkgDetail)
	detail.Platforms = make([]platform.Platform, 0)

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read tar header. err: %w", err)
		}

		switch header.Name {
		case "gse/VERSION":
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read version file. err: %w", err)
			}

			detail.Version = strings.Trim(string(content), "\n\r\t ")

		case "gse/DESCRIPTION":
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read description file. err: %w", err)
			}

			detail.ChangeLogZH = string(content)

		case "gse/DESCRIPTION_EN":
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read description-en file. err: %w", err)
			}

			detail.ChangeLogEN = string(content)

		case "gse/agent_linux_x86_64/bin/gse_agent":
			detail.Platforms = append(detail.Platforms, platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64})

		case "gse/agent_linux_aarch64/bin/gse_agent":
			detail.Platforms = append(detail.Platforms, platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64})

		case "gse/agent_windows_x86_64/bin/gse_agent.exe":
			detail.Platforms = append(detail.Platforms, platform.Platform{OS: criteria.OSWindows, Arch: criteria.CPUArchAmd64})

		case "gse/agent_darwin_x86_64/bin/gse_agent":
			detail.Platforms = append(detail.Platforms, platform.Platform{OS: criteria.OSDarwin, Arch: criteria.CPUArchAmd64})
		}
	}

	if detail.Version == "" {
		return nil, errors.New("version file not found")
	}

	return detail, nil
}
