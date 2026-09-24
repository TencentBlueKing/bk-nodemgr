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

package manager

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	originalCertFileName = "cert-all.tgz"
	releaseCertFileName  = "cert.tgz"
)

// ICert defines the interface for cert.
type ICert interface {
	IReleaseCert

	// UploadOriginCert uploads the origin cert.
	UploadOriginCert(nCtx contextx.IContext, certFile io.ReadCloser) (*types.OriginCertPkgDetail, error)

	// PublishReleaseCert generates release cert by upload-id and is-shared.
	PublishReleaseCert(nCtx contextx.IContext, uploadID string, isShared bool) error

	// EnsureCertToLocal ensure cert to local.
	EnsureCertToLocal(nCtx contextx.IContext, gen types.Generation) (fileiface.File, string, error)
}

// UploadOriginCert uploads origin cert.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting),golint/funcret(need reassign before return).
func (m *Manager) UploadOriginCert(nCtx contextx.IContext, certFile io.ReadCloser) (_ *types.OriginCertPkgDetail, retErr error) {
	if certFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin cert package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, certFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package. failed to get temp file")

		return nil, err
	}
	defer func() {
		if errClose := checkingFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	detail, err := checkOriginCertPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error(
			"failed to upload origin cert package. failed to check origin cert package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package. failed to get temp file")

		return nil, err
	}
	defer func() {
		if errClose := uploadingFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	pkgName := m.wrapOriginPackageName(originalCertFileName)

	// upload to upstream.
	if err := m.upstreamOriginCert.Store(nCtx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginCert.GetFile(nCtx, pkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseCert(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	uploadID, err := m.storageUpload.CreateCertUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginCert,
		SavedName: pkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin cert package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// record upload event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeCert, types.ReleaseNameCert, types.ReleaseVersionCert,
		[]platfmt.Platform{platfmt.UnknownPlatform()})

	logger.G.Biz(nCtx).With("filename", pkgName).Info("uploaded origin cert package to upstream")

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

// Certificate files must be either at the archive root or under cert/, never both.
// nolint: mnd
func checkOriginCertPkg(file io.ReadCloser) (*types.OriginCertPkgDetail, error) {
	detail := new(types.OriginCertPkgDetail)
	certFileNames := originCertFileNames()
	var hasRootFiles, hasDirFiles bool
	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{".*"},
			callback: func(paths []string, _ io.Reader) error {
				if len(paths) > 2 {
					return nil
				}
				if len(paths) == 2 && paths[0] != certDirNameRoot {
					return nil
				}

				fileName := paths[len(paths)-1]
				if !slices.Contains(certFileNames, fileName) {
					return nil
				}

				hasRootFiles = hasRootFiles || len(paths) == 1
				hasDirFiles = hasDirFiles || len(paths) == 2
				if hasRootFiles && hasDirFiles {
					return errors.New("cert package must not mix certificate files at root and under cert/")
				}

				detail.CertFiles = append(detail.CertFiles, fileName)

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	return detail, nil
}

const (
	certDirNameRoot            = "cert"
	certFileNameCaCrt          = "gseca.crt"
	certFileNameCertEncryptKey = "cert_encrypt.key"
	certFileNameAgentCrt       = "gse_agent.crt"
	certFileNameAgentKey       = "gse_agent.key"
	certFileNameServerCrt      = "gse_server.crt"
	certFileNameServerKey      = "gse_server.key"
	certFileNameAPIClientCrt   = "gse_api_client.crt"
	certFileNameAPIClientKey   = "gse_api_client.key"
)

func originCertFileNames() []string {
	return []string{
		certFileNameCaCrt,
		certFileNameAgentCrt,
		certFileNameAgentKey,
		certFileNameServerCrt,
		certFileNameServerKey,
		certFileNameAPIClientCrt,
		certFileNameAPIClientKey,
		certFileNameCertEncryptKey,
	}
}

// PublishReleaseCert generates release cert by upload-id and is-shared.
// nolint: funlen
func (m *Manager) PublishReleaseCert(nCtx contextx.IContext, uploadID string, isShared bool) (retErr error) {
	up, err := m.storageUpload.GetCertUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release cert, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginCert {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release cert, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginCert.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release cert, failed to get file")

		return err
	}

	// get origin content.
	content, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release cert, failed to get content")

		return err
	}
	defer func() {
		if errClose := content.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	// generate release file.
	generatedFile, err := m.generateCertPkg(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release cert, failed to generate cert pkg")

		return err
	}
	defer func() {
		if errClose := generatedFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	// upload to upstream.
	if err = m.upstreamReleaseCert.Store(
		nCtx, fileiface.FileInfo{Name: releaseCertFileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release cert, failed to upload to upstream")

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleaseCert.GetFile(nCtx, releaseCertFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release cert, failed to get release file")

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	certInfo := &types.ReleaseCert{
		Release: types.Release{
			Name:         releaseCertFileName,
			Generation:   types.Generation2,
			Type:         types.ReleaseTypeCert,
			Version:      types.ReleaseVersionCert,
			Platform:     platfmt.UnknownPlatform(),
			Labels:       nil,
			FileName:     releaseInfo.Name,
			MD5:          releaseInfo.MD5,
			Enabled:      true,
			AsDefault:    true,
			IsShared:     isShared && nCtx.TenantID() == tenant.SystemTenantID,
			IsSynced:     false,
			UpdatedAt:    time.Now(),
			Operator:     nCtx.BKUsername(),
			AdditionInfo: nil,
		},
	}

	// upsert release cert.
	if err = m.storageRelease.UpsertReleaseCert(nCtx, *certInfo); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release cert, failed to upsert release cert")

		return err
	}

	// record cert event.
	m.recordPublishEvent(nCtx, &certInfo.Release)

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release cert")

	return nil
}

// EnsureCertToLocal ensure cert to local.
func (m *Manager) EnsureCertToLocal(nCtx contextx.IContext, _ types.Generation) (fileiface.File, string, error) {
	// get cert from storage.
	cert, err := m.storageRelease.GetReleaseCert(nCtx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get release cert: %w", err)
	}

	return m.ensureReleaseToLocal(nCtx, cert.Release)
}

// nolint: nonamedreturns // Deferred cleanup must clear content when closing a file fails.
func (m *Manager) generateCertPkg(nCtx contextx.IContext, sourceFile io.ReadCloser) (content io.ReadCloser, err error) {
	defer func() {
		// Target cleanup may fail after the output stream is opened.
		if err != nil && content != nil {
			if errClose := content.Close(); errClose != nil {
				err = errors.Join(err, errClose)
			}
			content = nil
		}
	}()

	tempFileName, err := m.createTempFile(nCtx)
	if err != nil {
		return nil, err
	}

	targetFile, err := m.openTempFile(nCtx, tempFileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if errClose := targetFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	if err = generateTgz(targetFile,
		[]tgzWriteRuleDir{
			{targetFilePath: []string{certDirNameRoot}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules:  buildCertTgzWriteRules(),
		}},
	); err != nil {
		return nil, err
	}

	file, err := m.tempFileGroup.GetFile(nCtx, tempFileName)
	if err != nil {
		return nil, err
	}

	return file.Content(nCtx)
}

func buildCertTgzWriteRules() []tgzWriteRuleFile {
	certFileNames := originCertFileNames()
	rules := make([]tgzWriteRuleFile, 0)

	for _, certFileName := range certFileNames {
		rules = append(rules, tgzWriteRuleFile{
			sourceFilePath: []string{certFileName},
			targetFilePath: []string{certDirNameRoot, certFileName},
			targetFileMode: tgzModeFile,
		})
		rules = append(rules, tgzWriteRuleFile{
			sourceFilePath: []string{certDirNameRoot, certFileName},
			targetFilePath: []string{certDirNameRoot, certFileName},
			targetFileMode: tgzModeFile,
		})
	}

	return rules
}
