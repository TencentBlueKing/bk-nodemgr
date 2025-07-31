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
	"errors"
	"io"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	releaseCertFileName = "cert.tgz"
)

// UploadOriginCert uploads origin cert.
// nolint:funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginCert(ctx context.Context, certFileName string, certFile io.ReadCloser) (
	*types.OriginCertPkgDetail, error) {

	if certFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, certFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkOriginCertPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload origin cert package. failed to check origin cert package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package. failed to get temp file. err: %v", err)

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(certFileName)

	// upload to upstream.
	if err := m.upstreamOriginCert.Store(ctx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package, failed to store to upstream: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginCert.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseCert(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginCert,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin cert package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx, "uploaded origin cert package to upstream. file-name(%s)", pkgName)

	return detail, nil
}

/**
 * cert files contains:
 * gseca.crt
 * gse_agent.crt
 * gse_agent.key
 * gse_server.crt
 * gse_server.key
 * cert_encrypt.key
 * others can be ignored.
 */
func checkOriginCertPkg(file io.ReadCloser) (*types.OriginCertPkgDetail, error) {
	detail := new(types.OriginCertPkgDetail)
	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{"gseca.crt"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gseca.crt")

				return nil
			},
		},
		{
			filePath: []string{"gse_agent.crt"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_agent.crt")

				return nil
			},
		},
		{
			filePath: []string{"gse_agent.key"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_agent.key")

				return nil
			},
		},
		{
			filePath: []string{"gse_server.crt"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_server.crt")

				return nil
			},
		},
		{
			filePath: []string{"gse_server.key"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_server.key")

				return nil
			},
		},
		{
			filePath: []string{"gse_api_client.crt"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_api_client.crt")

				return nil
			},
		},
		{
			filePath: []string{"gse_api_client.key"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "gse_api_client.key")

				return nil
			},
		},
		{
			filePath: []string{"cert_encrypt.key"},
			callback: func(_ []string, _ io.Reader) error {
				detail.CertFiles = append(detail.CertFiles, "cert_encrypt.key")

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	return detail, nil
}

// PublishReleaseCert generates release cert by upload-id.
func (m *Manager) PublishReleaseCert(ctx context.Context, uploadID string) error {
	up, err := m.storageUpload.GetUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginCert {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginCert.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	content, err := file.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// generate release file.
	generatedFile, err := m.generateCertPkg(ctx, content)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to generate cert pkg. err: %v", err)

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleaseCert.Store(
		ctx, fileiface.FileInfo{Name: releaseCertFileName}, generatedFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to upload to upstream. err: %v", err)

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleaseCert.GetFile(ctx, releaseCertFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to get release file. err: %v", err)

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	// upsert release cert.
	if err = m.storageRelease.UpsertReleaseCert(ctx, types.ReleaseCert{
		FileName: releaseInfo.Name,
		MD5:      releaseInfo.MD5,
	}); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to upsert release cert: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release cert. file(%s), md5(%s)", releaseInfo.Name, releaseInfo.MD5)

	return nil
}

func (m *Manager) generateCertPkg(ctx context.Context, sourceFile io.ReadCloser) (io.ReadCloser, error) {
	tempFileName, err := m.createTempFile(ctx)
	if err != nil {
		return nil, err
	}

	targetFile, err := m.openTempFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	if err = generateTgz(targetFile,
		[]tgzWriteRuleDir{
			{targetFilePath: []string{"cert"}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{"gseca.crt"},
					targetFilePath: []string{"cert", "gseca.crt"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_agent.crt"},
					targetFilePath: []string{"cert", "gse_agent.crt"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_agent.key"},
					targetFilePath: []string{"cert", "gse_agent.key"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_server.crt"},
					targetFilePath: []string{"cert", "gse_server.crt"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_server.key"},
					targetFilePath: []string{"cert", "gse_server.key"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_api_client.crt"},
					targetFilePath: []string{"cert", "gse_api_client.crt"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"gse_api_client.key"},
					targetFilePath: []string{"cert", "gse_api_client.key"},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{"cert_encrypt.key"},
					targetFilePath: []string{"cert", "cert_encrypt.key"},
					targetFileMode: tgzModeFile,
				},
			},
		}},
	); err != nil {
		return nil, err
	}

	file, err := m.tempFileGroup.GetFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	return file.Content(ctx)
}
