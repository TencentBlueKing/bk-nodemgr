/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bkrepo provides the handler for bkrepo.
package bkrepo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
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
		for idx, node := range result.Records {
			records[idx] = node
		}

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

	if nCtx == nil {
		return errInvalidContext
	}
	if fileGroupPath == "" || info.Name == "" {
		return errEmptyPathOrName
	}
	if file == nil {
		return errors.New("file reader is nil")
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

	defer func() {
		_ = file.Close()
	}()

	_, err = h.cli.UploadFile(nCtx, uploadFileReq)
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	return nil
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

	return io.NopCloser(bytes.NewReader(downloadFileResp.Data)), nil
}
