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
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	originServerFileName = "gse_origin_server-unknown.tgz"
)

// UploadOriginServer uploads the origin server.
// nolint:funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginServer(ctx context.Context, pkgFileContent io.ReadCloser) (*types.OriginPkgDetail, error) {
	if pkgFileContent == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tmpOrignFile, err := tmp.NewTempFileWithSpecialName(pkgFileContent, originServerFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := tmpOrignFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get temp file content."+
			" err: %v", err)

		return nil, err
	}

	detail, err := checkGSE2OriginServerPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to check origin server package. err: %v", err)

		return nil, err
	}

	// origin server package only have one platform.
	if len(detail.Platforms) == 0 {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get platform")

		return nil, errors.New("failed to get platform")
	}
	plat := detail.Platforms[0]

	pkgName, err := nodepkg.FormatPkgName(
		types.Generation2,
		types.ReleaseTypeOriginServer,
		plat,
		detail.Version,
	)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package, failed to format package. err: %v", err)

		return nil, err
	}
	pkgName = m.wrapOriginPackageName(pkgName)

	uploadingFileContent, err := tmpOrignFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get temp file. err: %v", err)

		return nil, err
	}

	// upload to upstream.
	if err := m.upstreamOriginServer.Store(ctx, iface.FileInfo{Name: pkgName}, uploadingFileContent, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package, failed to upload to upstream. err: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginServer.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// create the upload record.
	uploadID, err := m.storageUpload.CreateUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginServer,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx,
		"uploaded origin server package to upstream. platform(%s), version(%s), file-name(%s)",
		plat, detail.Version, pkgName)

	return detail, nil
}

// checkGSE2OriginServerPkg check gse2 origin server package.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/gocyclo,cyclop (this function should be complex).
func checkGSE2OriginServerPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	detail := new(types.OriginPkgDetail)
	var seenFile, seenData bool
	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, "VERSION"},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read version file. err: %w", err)
				}

				detail.Version = strings.Trim(string(content), "\n\r\t ")

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "server", "bin", "gse_file"},
			callback: func(_ []string, r io.Reader) error {
				seenFile = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check server binary platform. err: %w", err)
				}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "server", "bin", "gse_data"},
			callback: func(_ []string, r io.Reader) error {
				seenData = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check server binary platform. err: %w", err)
				}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin server package. err: %w", err)
	}

	if !seenFile || !seenData || detail.Version == "" {
		return nil, fmt.Errorf("invalid origin server package. gse-file(%t) gse-data(%t) version(%s) platform(%v)",
			seenFile, seenData, detail.Version, detail.Platforms)
	}

	return detail, nil
}

const (
	elfMagic   = "\x7FELF"
	machineAMD = 0x3E // EM_X86_64
	machineARM = 0xB7 // EM_AARCH64
)

func checkServerBinaryPlatform(r io.Reader) (*platform.Platform, error) {
	// read the ELF header.
	buf := make([]byte, 20) // nolint:mnd
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	// check ELF magic number.
	if string(buf[:4]) != elfMagic { // nolint:mnd
		return nil, errors.New("not an ELF file")
	}

	// check ELFCLASS64
	if buf[4] != 2 { // nolint:mnd
		return nil, errors.New("not a 64-bit ELF file")
	}

	// check ELFDATA2LSB
	if buf[5] != 1 { // nolint:mnd
		return nil, errors.New("not little-endian ELF file")
	}

	// check the machine type
	machine := binary.LittleEndian.Uint16(buf[18:20]) // nolint:mnd

	switch machine {
	case machineAMD:
		return &platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}, nil
	case machineARM:
		return &platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64}, nil
	default:
		return nil, fmt.Errorf("unsupported architecture: 0x%X", machine)
	}
}
