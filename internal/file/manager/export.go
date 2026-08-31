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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IExport defines plugin package export manager methods.
type IExport interface {
	// ExportDownloadOriginPluginPackage downloads an exportable origin plugin package.
	ExportDownloadOriginPluginPackage(nCtx contextx.IContext, token string) (fileiface.File, error)

	// ExportPrepareOriginPluginPackage prepares an exportable origin plugin package.
	ExportPrepareOriginPluginPackage(nCtx contextx.IContext, param types.ExportPrepareParam) (fileiface.FileInfo, string, error)
}

// ExportDownloadOriginPluginPackage downloads an exportable origin plugin package.
func (m *Manager) ExportDownloadOriginPluginPackage(nCtx contextx.IContext, token string) (fileiface.File, error) {
	if token == "" {
		return nil, errors.New("export download token is required")
	}

	if m.exportCrypter == nil {
		return nil, errors.New("export download crypter is not initialized")
	}

	ciphertext, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode export download token: %w", err)
	}
	plaintext, err := m.exportCrypter.Decrypt(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt export download token: %w", err)
	}

	payload := types.ExportDownloadTokenPayload{}
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal export download token: %w", err)
	}
	if err := payload.Validate(); err != nil {
		return nil, fmt.Errorf("invalid export download token payload: %w", err)
	}
	if payload.ExpiresAt <= time.Now().Unix() {
		return nil, errors.New("export download token is expired")
	}

	tenantCtx := contextx.From(nCtx, contextx.WithTenantID(payload.TenantID))
	file, err := m.upstreamExport.GetFile(tenantCtx, payload.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get export package: %w", err)
	}

	return file, nil
}

type exportOriginPackage struct {
	uploadID string
	file     fileiface.File
}

const (
	// exportPackageFormat defines the export package format.
	exportPackageFormat = "%s-%s_%s.tgz"
)

// ExportPrepareOriginPluginPackage prepares an exportable origin plugin package.
func (m *Manager) ExportPrepareOriginPluginPackage(nCtx contextx.IContext, param types.ExportPrepareParam) (fileiface.FileInfo, string, error) {
	originPackages, err := m.collectOriginPackages(nCtx, param.UploadOriginPkgType, param.UploadIDs)
	if err != nil {
		return fileiface.FileInfo{}, "", err
	}

	platformOwners := make(map[string]string)
	for _, originPackage := range originPackages {
		platforms, err := m.validateOriginPackage(nCtx, param.UploadOriginPkgType, originPackage.file,
			param.PluginPkgName, param.PluginPkgVersion)
		if err != nil {
			return fileiface.FileInfo{}, "", fmt.Errorf("failed to validate origin package, upload-id(%s): %w", originPackage.uploadID, err)
		}

		for _, plat := range platforms {
			platform := plat.String()
			owner, ok := platformOwners[platform]
			if !ok {
				platformOwners[platform] = originPackage.uploadID
				continue
			}

			return fileiface.FileInfo{}, "", fmt.Errorf("conflict origin upload package platform, upload-id(%s), owner-upload-id(%s), platform(%s)",
				originPackage.uploadID, owner, platform)
		}
	}

	exportFileName, err := formatExportOriginPluginPackageName(param.PluginPkgName, param.PluginPkgVersion, param.FileNameSuffix)
	if err != nil {
		return fileiface.FileInfo{}, "", err
	}

	if len(originPackages) == 1 {
		return m.storeSingleExportOriginPackage(nCtx, originPackages[0].file, exportFileName)
	}

	return m.storeMergedExportOriginPackage(nCtx, param.UploadOriginPkgType, param.PluginPkgName, originPackages, exportFileName)
}

func formatExportOriginPluginPackageName(name string, version string, suffix string) (string, error) {
	if name == "" {
		return "", errors.New("format export origin plugin package name failed, pkg name is empty")
	}

	if version == "" {
		return "", errors.New("format plugin pkg name failed, version is empty")
	}

	return fmt.Sprintf(exportPackageFormat, name, version, suffix), nil
}

func (m *Manager) collectOriginPackages(nCtx contextx.IContext, uploadType types.UploadCategory, uploadIDs []string) ([]exportOriginPackage, error) {
	uploadIDs = conv.SliceUnique(uploadIDs)
	originPackages := make([]exportOriginPackage, 0, len(uploadIDs))
	for _, uploadID := range uploadIDs {
		var (
			up   *types.Upload
			file fileiface.File
			err  error
		)

		switch uploadType {
		case types.UploadCategoryOriginPluginV2:
			up, err = m.storageUpload.GetPluginV2Upload(nCtx, uploadID)
		case types.UploadCategoryOriginExternalPluginV2:
			up, err = m.storageUpload.GetExternalPluginV2Upload(nCtx, uploadID)
		case types.UploadCategoryOriginPluginV3:
			up, err = m.storageUpload.GetPluginV3Upload(nCtx, uploadID)
		default:
			return nil, fmt.Errorf("unsupported upload origin pkg type: %s", uploadType)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get origin plugin upload, upload-id(%s), type(%s): %w", uploadID, uploadType, err)
		}

		if up.Category != uploadType {
			return nil, fmt.Errorf("upload category mismatch, upload-id(%s), expected(%s), actual(%s)", uploadID, uploadType, up.Category)
		}

		switch uploadType {
		case types.UploadCategoryOriginPluginV2:
			file, err = m.upstreamOriginPluginV2.GetFile(nCtx, up.SavedName)
		case types.UploadCategoryOriginExternalPluginV2:
			file, err = m.upstreamOriginExternalPluginV2.GetFile(nCtx, up.SavedName)
		case types.UploadCategoryOriginPluginV3:
			file, err = m.upstreamOriginPluginV3.GetFile(nCtx, up.SavedName)
		default:
			return nil, fmt.Errorf("unsupported upload origin pkg type: %s", uploadType)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get origin plugin file, upload-id(%s): %w", uploadID, err)
		}

		originPackages = append(originPackages, exportOriginPackage{uploadID: uploadID, file: file})
	}

	return originPackages, nil
}

func (m *Manager) validateOriginPackage(
	nCtx contextx.IContext, uploadType types.UploadCategory, file fileiface.File, pkgName string,
	pkgVersion string) ([]platfmt.Platform, error) {

	content, err := file.Content(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get origin package content: %w", err)
	}

	switch uploadType {
	case types.UploadCategoryOriginPluginV2:
		detail, err := checkOriginPluginV2Pkg(content)
		if err != nil {
			return nil, err
		}
		if err := validateExportOriginPackageDetail(detail.PluginPkgName, detail.Version, pkgName, pkgVersion); err != nil {
			return nil, err
		}

		return detail.Platforms, nil
	case types.UploadCategoryOriginExternalPluginV2:
		detail, err := checkOriginExternalPluginPkg(content)
		if err != nil {
			return nil, err
		}
		if err := validateExportOriginPackageDetail(detail.PluginPkgName, detail.Version, pkgName, pkgVersion); err != nil {
			return nil, err
		}

		return detail.Platforms, nil
	case types.UploadCategoryOriginPluginV3:
		detail, err := checkOriginPluginV3Pkg(content)
		if err != nil {
			return nil, err
		}
		if err := validateExportOriginPackageDetail(detail.PluginPkgName, detail.Version, pkgName, pkgVersion); err != nil {
			return nil, err
		}

		return detail.Platforms, nil
	default:
		return nil, fmt.Errorf("unsupported upload origin pkg type: %s", uploadType)
	}
}

func validateExportOriginPackageDetail(actualName, actualVersion, expectedName, expectedVersion string) error {
	if actualName != expectedName {
		return fmt.Errorf("plugin package name mismatch, expected(%s), actual(%s)", expectedName, actualName)
	}

	if actualVersion != expectedVersion {
		return fmt.Errorf("plugin package version mismatch, expected(%s), actual(%s)", expectedVersion, actualVersion)
	}

	return nil
}

func (m *Manager) storeSingleExportOriginPackage(nCtx contextx.IContext, originFile fileiface.File, exportFileName string) (
	fileiface.FileInfo, string, error) {

	content, err := originFile.Content(nCtx)
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to get origin package content: %w", err)
	}

	if err := m.upstreamExport.Store(nCtx, fileiface.FileInfo{Name: exportFileName}, content, true); err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to store export package: %w", err)
	}

	exportedFile, err := m.upstreamExport.GetFile(nCtx, exportFileName)
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to get export package: %w", err)
	}

	return exportedFile.Info(), m.exportAddress, nil
}

func (m *Manager) storeMergedExportOriginPackage(
	nCtx contextx.IContext, uploadType types.UploadCategory, pluginName string, originPackages []exportOriginPackage, exportFileName string) (
	fileiface.FileInfo, string, error) {

	sources, err := buildOriginUploadTgzMergeSources(nCtx, originPackages)
	if err != nil {
		return fileiface.FileInfo{}, "", err
	}

	rules, err := buildOriginUploadTgzMergeRules(uploadType, pluginName)
	if err != nil {
		for _, source := range sources {
			if closeErr := source.sourceFile.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}

		return fileiface.FileInfo{}, "", err
	}

	var mergedContent bytes.Buffer
	if err := mergeOriginUploadTgz(&mergedContent, sources, rules); err != nil {
		return fileiface.FileInfo{}, "", err
	}

	tmpFile, err := tmp.NewTempFile(io.NopCloser(bytes.NewReader(mergedContent.Bytes())), "merged")
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to create merged export package: %w", err)
	}
	defer func() {
		if cleanErr := tmpFile.CleanUp(); cleanErr != nil {
			errMsg := cleanErr.Error()
			logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(cleanErr).Error("failed to clean temp file")
		}
	}()

	mergedFile, err := local.NewLocalFile(tmpFile.Path())
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to open merged export package: %w", err)
	}

	mergedContentReader, err := mergedFile.Content(nCtx)
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to read merged export package: %w", err)
	}

	if err := m.upstreamExport.Store(nCtx, fileiface.FileInfo{Name: exportFileName}, mergedContentReader, true); err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to store export package: %w", err)
	}

	exportedFile, err := m.upstreamExport.GetFile(nCtx, exportFileName)
	if err != nil {
		return fileiface.FileInfo{}, "", fmt.Errorf("failed to get export package: %w", err)
	}

	return exportedFile.Info(), m.exportAddress, nil
}

func buildOriginUploadTgzMergeSources(nCtx contextx.IContext, originPackages []exportOriginPackage) ([]tgzMergeSource, error) {
	var err error
	sources := make([]tgzMergeSource, 0, len(originPackages))
	defer func() {
		if err != nil {
			for _, source := range sources {
				if errClose := source.sourceFile.Close(); errClose != nil {
					err = errors.Join(err, errClose)
				}
			}
		}
	}()

	for _, originPackage := range originPackages {
		content, err := originPackage.file.Content(nCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to get origin package content, upload-id(%s): %w", originPackage.uploadID, err)
		}

		sources = append(sources, tgzMergeSource{
			name:       originPackage.uploadID,
			sourceFile: content,
		})
	}

	return sources, nil
}

func buildOriginUploadTgzMergeRules(uploadType types.UploadCategory, pluginName string) ([]originUploadTgzMergeRule, error) {
	switch uploadType {
	case types.UploadCategoryOriginPluginV2:
		return buildOriginPluginV2TgzMergeRules(), nil
	case types.UploadCategoryOriginExternalPluginV2:
		return buildOriginExternalPluginV2TgzMergeRules(), nil
	case types.UploadCategoryOriginPluginV3:
		return buildOriginPluginV3TgzMergeRules(pluginName), nil
	default:
		return nil, fmt.Errorf("unsupported upload origin pkg type: %s", uploadType)
	}
}

type tgzMergeSource struct {
	name       string
	sourceFile io.ReadCloser
}

type originUploadTgzMergeRule struct {
	sourcePathSegment int
	sourcePathPrefix  string
	sourcePath        string
	targetPath        func(string) string
	writeEntry        func(*tar.Writer, *tar.Header, io.Reader, string, string) error
	finalize          func(*tar.Writer) error
}

func buildOriginPluginV2TgzMergeRules() []originUploadTgzMergeRule {
	return []originUploadTgzMergeRule{
		{
			sourcePathSegment: 0,
			sourcePathPrefix:  originalPluginPkgV2DirNamePlatPrefix,
			targetPath:        func(name string) string { return name },
			writeEntry:        writeOriginUploadTgzEntry,
		},
	}
}

func buildOriginExternalPluginV2TgzMergeRules() []originUploadTgzMergeRule {
	return []originUploadTgzMergeRule{
		{
			sourcePathSegment: 0,
			sourcePathPrefix:  originalExternalPluginDirNamePlatPrefix,
			targetPath:        func(name string) string { return name },
			writeEntry:        writeOriginUploadTgzEntry,
		},
	}
}

func buildOriginPluginV3TgzMergeRules(pluginName string) []originUploadTgzMergeRule {
	rootWritten := false
	var projectFile []byte
	var projectFileMode int64

	return []originUploadTgzMergeRule{
		{
			sourcePathSegment: 1,
			sourcePathPrefix:  originalPluginPkgV3DirNamePlatPrefix,
			targetPath:        func(name string) string { return name },
			writeEntry: func(tarWriter *tar.Writer, header *tar.Header, source io.Reader, sourceName, targetName string) error {
				if !rootWritten {
					if err := tarWriter.WriteHeader(&tar.Header{
						Name:     pluginName,
						Mode:     tgzModeDir,
						ModTime:  time.Now(),
						Typeflag: tar.TypeDir,
					}); err != nil {
						return fmt.Errorf("failed to write origin upload tgz root directory, path(%s): %w", pluginName, err)
					}
					rootWritten = true
				}

				return writeOriginUploadTgzEntry(tarWriter, header, source, sourceName, targetName)
			},
		},
		{
			sourcePath: path.Join(pluginName, originalPluginPkgV3FileNameProject),
			targetPath: func(name string) string { return name },
			writeEntry: func(_ *tar.Writer, header *tar.Header, source io.Reader, sourceName, targetName string) error {
				content, mode, err := readOriginUploadTgzProjectFile(sourceName, targetName, header, source)
				if err != nil {
					return err
				}
				projectFile = content
				projectFileMode = mode

				return nil
			},
			finalize: func(tarWriter *tar.Writer) error {
				if projectFile == nil {
					return nil
				}

				name := path.Join(pluginName, originalPluginPkgV3FileNameProject)
				if err := tarWriter.WriteHeader(&tar.Header{
					Name:     name,
					Mode:     projectFileMode,
					ModTime:  time.Now(),
					Typeflag: tar.TypeReg,
					Size:     int64(len(projectFile)),
				}); err != nil {
					return fmt.Errorf("failed to write origin upload tgz project file header, path(%s): %w", name, err)
				}

				if _, err := tarWriter.Write(projectFile); err != nil {
					return fmt.Errorf("failed to write origin upload tgz project file, path(%s): %w", name, err)
				}

				return nil
			},
		},
	}
}

// mergeOriginUploadTgz takes responsibility for closing all source files and writers.
func mergeOriginUploadTgz(targetFile io.Writer, sources []tgzMergeSource, rules []originUploadTgzMergeRule) (err error) {
	defer func() {
		for _, source := range sources {
			if errClose := source.sourceFile.Close(); errClose != nil {
				err = errors.Join(err, errClose)
			}
		}
	}()

	gzipWriter := gzip.NewWriter(targetFile)
	defer func() {
		if errClose := gzipWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if errClose := tarWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	for _, source := range sources {
		if err := writeOriginUploadTgzSource(tarWriter, source, rules); err != nil {
			return err
		}
	}

	for _, rule := range rules {
		if rule.finalize == nil {
			continue
		}
		if err := rule.finalize(tarWriter); err != nil {
			return err
		}
	}

	return nil
}

func writeOriginUploadTgzSource(tarWriter *tar.Writer, source tgzMergeSource, rules []originUploadTgzMergeRule) (err error) {
	gzipReader, err := gzip.NewReader(source.sourceFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader, source(%s): %w", source.name, err)
	}
	defer func() {
		if errClose := gzipReader.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header, source(%s): %w", source.name, err)
		}

		name, err := cleanTgzPath(header.Name)
		if err != nil {
			return fmt.Errorf("invalid tar path, source(%s), path(%s): %w", source.name, header.Name, err)
		}

		pathSegments := strings.Split(name, "/")
		for _, rule := range rules {
			if !matchOriginUploadTgzMergeRule(rule, name, pathSegments) {
				continue
			}

			targetName := rule.targetPath(name)
			if err := rule.writeEntry(tarWriter, header, tarReader, source.name, targetName); err != nil {
				return err
			}

			break
		}
	}

	return nil
}

func matchOriginUploadTgzMergeRule(rule originUploadTgzMergeRule, name string, pathSegments []string) bool {
	if rule.sourcePath != "" {
		return name == rule.sourcePath
	}
	if rule.sourcePathSegment >= len(pathSegments) {
		return false
	}

	return strings.HasPrefix(pathSegments[rule.sourcePathSegment], rule.sourcePathPrefix)
}

func writeOriginUploadTgzEntry(tarWriter *tar.Writer, header *tar.Header, source io.Reader, sourceName string, targetName string) error {
	switch header.Typeflag {
	case tar.TypeDir, tar.TypeReg:
		if err := tarWriter.WriteHeader(&tar.Header{
			Name:     targetName,
			Mode:     header.Mode,
			ModTime:  time.Now(),
			Typeflag: header.Typeflag,
			Size:     header.Size,
		}); err != nil {
			return fmt.Errorf("failed to write origin upload tgz header, source(%s), path(%s): %w", sourceName, targetName, err)
		}

		if header.Typeflag != tar.TypeReg {
			return nil
		}

		if _, err := io.Copy(tarWriter, source); err != nil {
			return fmt.Errorf("failed to write origin upload tgz file, source(%s), path(%s): %w", sourceName, targetName, err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported origin upload tgz entry type, source(%s), path(%s), type(%d)", sourceName, targetName, header.Typeflag)
	}
}

func readOriginUploadTgzProjectFile(sourceName string, name string, header *tar.Header, tarReader io.Reader) ([]byte, int64, error) {
	if header.Typeflag != tar.TypeReg {
		return nil, 0, fmt.Errorf("invalid origin upload tgz project file type, source(%s), path(%s), type(%d)", sourceName, name, header.Typeflag)
	}

	content, err := io.ReadAll(tarReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read origin upload tgz project file, source(%s), path(%s): %w", sourceName, name, err)
	}

	return content, header.Mode, nil
}

func cleanTgzPath(name string) (string, error) {
	name = strings.TrimPrefix(name, "./")
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == "" {
		return "", errors.New("path is empty")
	}

	if path.IsAbs(cleaned) || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") || cleaned == ".." {
		return "", errors.New("path escapes archive root")
	}

	return cleaned, nil
}
