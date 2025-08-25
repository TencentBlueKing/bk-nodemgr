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
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler is the interface for bkrepo handler.
type IHandler interface {
	// GetFile returns the file object by specified path.
	GetFile(ctx context.Context, path string) (fileiface.File, error)

	// GetFileGroup returns the file group object by specified path.
	GetFileGroup(ctx context.Context, path string) (fileiface.FileGroup, error)

	// EnsureFileGroup get the file group or make it, returns the file group object by specified path.
	EnsureFileGroup(ctx context.Context, path string) (fileiface.FileGroup, error)
}

// Handler is the handler of bkrepo.
type Handler struct {
	cli    *cli
	logger logger.ILogger
}

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.ILogger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:    cli,
		logger: logger.LoggerDefault{},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// GetFile returns the file object.
func (h *Handler) GetFile(ctx context.Context, path string) (fileiface.File, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.cli.QueryNodeInfo(ctx, &QueryNodeInfoReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("get file failed, err: %w", err)
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
func (h *Handler) GetFileGroup(ctx context.Context, path string) (fileiface.FileGroup, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.cli.QueryNodeInfo(ctx, &QueryNodeInfoReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("get file failed, err: %w", err)
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
func (h *Handler) EnsureFileGroup(ctx context.Context, path string) (fileiface.FileGroup, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	resp, err := h.GetFileGroup(ctx, path)
	if err == nil {
		return resp, nil
	}

	if !errors.Is(err, errNodeNotFound) {
		return nil, err
	}

	if err = h.cli.MkDir(ctx, &MkdirReq{Path: path}); err != nil {
		return nil, fmt.Errorf("ensure file group failed, err: %w", err)
	}

	return h.GetFileGroup(ctx, path)
}

func (h *Handler) listGroups(ctx context.Context, path string) ([]fileiface.FileGroup, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	nodes, err := h.listNodes(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list groups failed, err: %w", err)
	}

	groups := make([]fileiface.FileGroup, 0)
	for _, node := range nodes {
		if !node.Folder {
			continue
		}

		group, err := h.GetFileGroup(ctx, node.FullPath)
		if err != nil {
			return nil, fmt.Errorf("list groups failed, err: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (h *Handler) listFiles(ctx context.Context, path string) ([]fileiface.File, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	nodes, err := h.listNodes(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list files failed, err: %w", err)
	}

	files := make([]fileiface.File, 0)
	for _, node := range nodes {
		if node.Folder {
			continue
		}

		file, err := h.GetFile(ctx, node.FullPath)
		if err != nil {
			return nil, fmt.Errorf("list files failed, err: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func (h *Handler) listNodes(ctx context.Context, path string) ([]NodeRecord, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	executor := runtime.NewPageExecutor[NodeRecord](500, time.Minute) // nolint: mnd
	fn := func(ctx context.Context, p types.Page) ([]NodeRecord, error) {
		result, err := h.cli.ListNode(ctx, &ListNodeReq{
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

	records, err := executor.Execute(ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	return records.Items, nil
}

func (h *Handler) storeFile(
	ctx context.Context, fileGroupPath string, info fileiface.FileInfo, file io.ReadCloser, overwrite bool) error {

	if ctx == nil {
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
		return fmt.Errorf("join path failed, err: %w", err)
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

	_, err = h.cli.UploadFile(ctx, uploadFileReq)
	if err != nil {
		return fmt.Errorf("upload file failed, err: %w", err)
	}

	return nil
}

func (h *Handler) getFileContent(ctx context.Context, path string) (io.ReadCloser, error) {
	if ctx == nil {
		return nil, errInvalidContext
	}
	if path == "" {
		return nil, errEmptyPathOrName
	}

	downloadFileResp, err := h.cli.DownloadFile(ctx, &DownloadFileReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("download file failed, err: %w", err)
	}

	return io.NopCloser(bytes.NewReader(downloadFileResp.Data)), nil
}
