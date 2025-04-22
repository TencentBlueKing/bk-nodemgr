/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"io"
	"net/url"
)

// IHandler is the interface for handler.
type IHandler interface {
	GetFile(ctx context.Context, path string) (iface.File, error)
	GetFileGroup(ctx context.Context, path string) (iface.FileGroup, error)
}

// Handler is the handler of bkrepo.
type Handler struct {
	cli    *cli
	logger logger.Logger
}

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.Logger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *client.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
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
func (h *Handler) GetFile(ctx context.Context, path string) (iface.File, error) {
	req := &QueryNodeInfoReq{
		ProjectId: "nodemgr-test",
		RepoName:  "nodemgr-pkg",
		Path:      path,
	}

	resp, err := h.cli.QueryNodeInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get file failed, err: %w", err)
	}

	if resp.NodeInfo.Folder {
		return nil, errors.New("get file failed, ths path is a folder")
	}

	file := &File{
		infoFn: func() (iface.FileInfo, error) {
			queryNodeInfoResp, err := h.cli.QueryNodeInfo(ctx, req)
			if err != nil {
				return iface.FileInfo{}, fmt.Errorf("query node info failed, err: %w", err)
			}

			if queryNodeInfoResp.NodeInfo.Folder {
				return iface.FileInfo{}, errors.New("query node info failed, ths path is a folder")
			}

			info := iface.FileInfo{
				Name: queryNodeInfoResp.NodeInfo.Name,
				Size: int64(queryNodeInfoResp.NodeInfo.Size),
				Md5:  queryNodeInfoResp.NodeInfo.Md5,
			}

			return info, nil
		},
		contentFn: func() (io.ReadCloser, error) {
			downloadFileReq := &DownloadFileReq{
				ProjectID: "nodemgr-test",
				RepoName:  "nodemgr-pkg",
				Path:      resp.NodeInfo.FullPath,
			}

			downloadFileResp, err := h.cli.DownloadFile(ctx, downloadFileReq)
			if err != nil {
				return nil, fmt.Errorf("download file failed, err: %w", err)
			}

			return io.NopCloser(bytes.NewReader(downloadFileResp.Data)), nil
		},
		fullPath: resp.NodeInfo.FullPath,
	}

	return file, nil
}

// GetFileGroup returns the file group object.
func (h *Handler) GetFileGroup(ctx context.Context, path string) (iface.FileGroup, error) {
	req := &QueryNodeInfoReq{
		ProjectId: "nodemgr-test",
		RepoName:  "nodemgr-pkg",
		Path:      path,
	}

	resp, err := h.cli.QueryNodeInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get file failed, err: %w", err)
	}

	if !resp.NodeInfo.Folder {
		return nil, errors.New("get file failed, ths path is not a folder")
	}

	group := &FileGroup{
		name: resp.NodeInfo.Name,
	}

	group.refreshFileMapFn = func() error {
		_, err, _ := group.sg.Do("refreshFileMapFn", func() (interface{}, error) {

			return nil, errors.New("implement me")
		})
		if err != nil {
			return err
		}

		return nil
	}

	group.refreshSubGroupMapFn = func() error {
		_, err, _ := group.sg.Do("refreshSubGroupMapFn", func() (interface{}, error) {

			return nil, errors.New("implement me")
		})
		if err != nil {
			return err
		}

		return nil
	}

	group.storeFn = func(ctx context.Context, info iface.FileInfo, reader io.ReadCloser, overwrite bool) error {
		uploadFilePath, err := url.JoinPath(resp.NodeInfo.FullPath, info.Name)
		if err != nil {
			return fmt.Errorf("join path failed, err: %w", err)
		}

		uploadFileReq := &UploadFileReq{
			ProjectId: "nodemgr-test",
			RepoName:  "nodemgr-pkg",
			Path:      uploadFilePath,
			Info: UploadFileInfo{
				Md5:        info.Md5,
				Overwrite:  true,
				ExpireDays: 0,
				Meta: map[string]string{
					"description": info.Description,
				},
			},
			Data: make([]byte, 0),
		}

		for key, value := range info.ExtendFields {
			uploadFileReq.Info.Meta[key] = value
		}

		_, err = h.cli.UploadFile(ctx, uploadFileReq)
		if err != nil {
			return fmt.Errorf("upload file failed, err: %w", err)
		}

		if err := group.refreshFileMapFn(); err != nil {
			return fmt.Errorf("refresh file map failed, err: %w", err)
		}

		return nil
	}

	return group, nil
}
