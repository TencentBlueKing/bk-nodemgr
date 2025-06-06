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
	"encoding/binary"
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

// UploadOriginServer uploads the origin server.
// nolint:funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginServer(
	ctx context.Context, gen types.Generation, pkgFile io.ReadCloser) (iface.FileInfo, error) {

	// validation.
	if err := gen.Validate(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. invalid generation. err: %v", err)

		return iface.FileInfo{}, err
	}
	if gen == types.Generation1 {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. generation 1 is not supported")

		return iface.FileInfo{}, errors.New("generation 1 is not supported")
	}
	if pkgFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. file is nil")

		return iface.FileInfo{}, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, pkgFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to save temp file. err: %v", err)

		return iface.FileInfo{}, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get temp file. err: %v", err)

		return iface.FileInfo{}, err
	}

	detail, err := checkGen2OriginServerPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to check origin server package. err: %v", err)

		return iface.FileInfo{}, err
	}

	pkgName, err := nodepkg.FormatPkgName(
		gen,
		types.ReleaseTypeOriginServer,
		detail.Platform,
		detail.Version,
	)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package, failed to format package. err: %v", err)

		return iface.FileInfo{}, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get temp file. err: %v", err)

		return iface.FileInfo{}, err
	}

	// upload to upstream.
	if err := m.upstreamOriginServer.Store(ctx, iface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package, failed to upload to upstream. err: %v", err)

		return iface.FileInfo{}, err
	}

	// get file.
	file, err := m.upstreamOriginServer.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get file from upstream. err: %v", err)

		return iface.FileInfo{}, err
	}

	// get info.
	info, err := file.Info(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to get file info. err: %v", err)

		return iface.FileInfo{}, err
	}

	// save to db.
	err = m.daoRelease.UpsertMany(ctx, &types.Release{
		Generation: gen,
		Type:       types.ReleaseTypeOriginServer,
		Version:    detail.Version,
		Platform:   detail.Platform,
		FileName:   info.Name,
		MD5:        info.MD5,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin server package. failed to save to db. err: %v", err)

		return iface.FileInfo{}, err
	}

	m.logger.InfoCtxf(ctx,
		"uploaded origin server package to upstream. generation(%d), platform(%s), version(%s), file-name(%s)",
		gen, detail.Platform, detail.Version, pkgName)

	return info, nil
}

// checkGen2OriginServerPkg check gen2 origin server package.
func checkGen2OriginServerPkg(file io.ReadCloser) (*OriginServerPkgDetail, error) {
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader. err: %w", err)
	}
	defer func() {
		_ = gzr.Close()
		_ = file.Close()
	}()

	var seenFile, seenData bool
	detail := new(OriginServerPkgDetail)

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

		case "gse/server/bin/gse_file":
			seenFile = true
			plat, err := checkServerBinaryPlatform(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to check server binary platform. err: %w", err)
			}
			detail.Platform = *plat

		case "gse/server/bin/gse_data":
			seenData = true
			plat, err := checkServerBinaryPlatform(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to check server binary platform. err: %w", err)
			}
			detail.Platform = *plat
		}

		if seenFile && seenData && detail.Version != "" && detail.Platform.Validate() {
			return detail, nil
		}
	}

	return nil, fmt.Errorf("invalid origin server package. gse-file(%t) gse-data(%t) version(%s) platform(%s)",
		seenFile, seenData, detail.Version, detail.Platform)
}

// OriginServerPkgDetail origin server pkg detail.
type OriginServerPkgDetail struct {
	Version  string
	Platform platform.Platform
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
