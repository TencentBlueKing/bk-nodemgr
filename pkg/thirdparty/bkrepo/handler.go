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

// Package bkrepo provides the handler for bkrepo.
package bkrepo

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler is the interface for bkrepo handler.
type IHandler interface {
	// GetFile returns the file object by specified path.
	GetFile(nCtx contextx.IContext, path string) (fileiface.File, error)

	// GetFileGroup returns the file group object by specified path.
	GetFileGroup(nCtx contextx.IContext, path string) (fileiface.FileGroup, error)

	// EnsureFileGroup get the file group or make it, returns the file group object by specified path.
	EnsureFileGroup(nCtx contextx.IContext, path string) (fileiface.FileGroup, error)
}

// Handler is the handler of bkrepo.
type Handler struct {
	cli *cli
}

// OptionFn ...
type OptionFn func(*Handler)

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli: cli,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// GetFile returns the file object.
func (h *Handler) GetFile(nCtx contextx.IContext, path string) (fileiface.File, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.cli.QueryNodeInfo(nCtx, &QueryNodeInfoReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("get file failed: %w", err)
	}

	if resp.NodeInfo.Folder {
		return nil, errors.New("get file failed, ths path is a folder")
	}

	return &File{
		info:    resp.NodeInfo,
		absDirs: fileiface.ConvertAbsPathToAbsDirs(filepath.Dir(path)),
		handler: h,
	}, nil
}

// GetFileGroup returns the file group object.
func (h *Handler) GetFileGroup(nCtx contextx.IContext, path string) (fileiface.FileGroup, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.cli.QueryNodeInfo(nCtx, &QueryNodeInfoReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("get file failed: %w", err)
	}

	if !resp.NodeInfo.Folder {
		return nil, errors.New("get file failed, ths path is not a folder")
	}

	return &FileGroup{
		info:    resp.NodeInfo,
		absDirs: fileiface.ConvertAbsPathToAbsDirs(path),
		handler: h,
	}, nil
}

// EnsureFileGroup get the file group or make it, returns the file group object by specified path.
func (h *Handler) EnsureFileGroup(nCtx contextx.IContext, path string) (fileiface.FileGroup, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.GetFileGroup(nCtx, path)
	if err == nil {
		return resp, nil
	}

	if !errors.Is(err, errNodeNotFound) {
		return nil, err
	}

	if err = h.cli.MkDir(nCtx, &MkdirReq{Path: path}); err != nil {
		return nil, fmt.Errorf("ensure file group failed: %w", err)
	}

	return h.GetFileGroup(nCtx, path)
}

func (h *Handler) listGroups(nCtx contextx.IContext, path string) ([]fileiface.FileGroup, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	nodes, err := h.listNodes(nCtx, path)
	if err != nil {
		return nil, fmt.Errorf("list groups failed: %w", err)
	}

	groups := make([]fileiface.FileGroup, 0)
	for _, node := range nodes {
		if !node.Folder {
			continue
		}

		group, err := h.GetFileGroup(nCtx, node.FullPath)
		if err != nil {
			return nil, fmt.Errorf("list groups failed: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (h *Handler) listFiles(nCtx contextx.IContext, path string) ([]fileiface.File, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	nodes, err := h.listNodes(nCtx, path)
	if err != nil {
		return nil, fmt.Errorf("list files failed: %w", err)
	}

	files := make([]fileiface.File, 0)
	for _, node := range nodes {
		if node.Folder {
			continue
		}

		file, err := h.GetFile(nCtx, node.FullPath)
		if err != nil {
			return nil, fmt.Errorf("list files failed: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func (h *Handler) listNodes(nCtx contextx.IContext, path string) ([]NodeRecord, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	executor := pageexecutor.NewPageExecutor[NodeRecord](500, time.Minute) // nolint: mnd
	fn := func(nCtx contextx.IContext, p types.Page) ([]NodeRecord, error) {
		result, err := h.cli.ListNode(nCtx, &ListNodeReq{
			Path:     path,
			PageNum:  p.Offset/p.Limit + 1,
			PageSize: p.Limit,
		})
		if err != nil {
			return nil, err
		}

		records := make([]NodeRecord, len(result.Records))
		copy(records, result.Records)

		return records, nil
	}

	records, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("list node failed: %w", err)
	}

	return records.Items, nil
}

func (h *Handler) storeFile(
	nCtx contextx.IContext, fileGroupPath string, info fileiface.FileInfo, file io.ReadCloser, overwrite bool) error {

	if file == nil {
		return errors.New("file reader is nil")
	}

	if nCtx == nil {
		return errInvalidContext
	}
	if err := nCtx.Err(); err != nil {
		return fmt.Errorf("context is done: %w", err)
	}
	if fileGroupPath == "" || info.Name == "" {
		return errEmptyPathOrName
	}

	uploadFilePath, err := url.JoinPath(fileGroupPath, info.Name)
	if err != nil {
		return fmt.Errorf("join path failed: %w", err)
	}

	uploadFileReq := &UploadFileReq{
		Path: uploadFilePath,
		Info: UploadFileInfo{
			Md5:        info.MD5,
			Overwrite:  overwrite,
			ExpireDays: 0,
			Meta: map[string]string{
				"description": info.Description,
			},
		},
		Reader: file,
	}

	for key, value := range info.ExtendFields {
		uploadFileReq.Info.Meta[key] = value
	}

	_, err = h.cli.UploadFile(nCtx, uploadFileReq)
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	return nil
}

func (h *Handler) copyNode(
	nCtx contextx.IContext, srcGroup *FileGroup, srcPath string, destGroup *FileGroup, destPath string, overwrite bool) error {

	if nCtx == nil {
		return errInvalidContext
	}

	if srcGroup.handler != destGroup.handler || srcGroup.handler != h {
		return errors.New("source and destination file groups must use the same handler")
	}
	isDifferentProject := srcGroup.info.ProjectID != destGroup.info.ProjectID
	isDifferentRepo := srcGroup.info.RepoName != destGroup.info.RepoName
	if isDifferentProject || isDifferentRepo {
		return errors.New("source and destination file groups must use the same project and repository")
	}

	srcFullPath, err := resolveNodePath(srcGroup, srcPath, false)
	if err != nil {
		return fmt.Errorf("resolve source node path failed: %w", err)
	}
	destFullPath, err := resolveNodePath(destGroup, destPath, true)
	if err != nil {
		return fmt.Errorf("resolve destination node path failed: %w", err)
	}

	_, err = h.cli.CopyNode(nCtx, &CopyNodeReq{
		SrcProjectID:  h.cli.effectiveProjectID(),
		SrcRepoName:   h.cli.config.RepoName,
		SrcFullPath:   srcFullPath,
		DestProjectID: h.cli.effectiveProjectID(),
		DestRepoName:  h.cli.config.RepoName,
		DestFullPath:  destFullPath,
		Overwrite:     overwrite,
	})
	if err != nil {
		return fmt.Errorf("copy node failed: %w", err)
	}

	return nil
}

func (h *Handler) deleteNode(nCtx contextx.IContext, group *FileGroup, nodePath string) error {
	if nCtx == nil {
		return errInvalidContext
	}

	fullPath, err := resolveNodePath(group, nodePath, false)
	if err != nil {
		return fmt.Errorf("resolve node path failed: %w", err)
	}

	if _, err := h.cli.DeleteNode(nCtx, &DeleteNodeReq{Path: fullPath}); err != nil {
		return fmt.Errorf("delete node failed: %w", err)
	}

	return nil
}

func resolveNodePath(group *FileGroup, relativePath string, allowGroupRoot bool) (string, error) {
	if relativePath == "" {
		return "", errEmptyPathOrName
	}
	if path.IsAbs(relativePath) {
		return "", errors.New("path must be relative")
	}

	cleanedPath := path.Clean(relativePath)
	if cleanedPath == "." {
		if allowGroupRoot {
			return path.Clean(group.info.FullPath), nil
		}

		return "", errors.New("path must identify a node in the file group")
	}
	if cleanedPath == ".." || strings.HasPrefix(cleanedPath, "../") {
		return "", errors.New("path must not escape the file group")
	}

	return path.Join(group.info.FullPath, cleanedPath), nil
}

func (h *Handler) getFileContent(nCtx contextx.IContext, path string) (io.ReadCloser, error) {
	if nCtx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	downloadFileResp, err := h.cli.DownloadFile(nCtx, &DownloadFileReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("download file failed: %w", err)
	}

	return downloadFileResp.Data, nil
}
