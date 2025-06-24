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
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	transferPackageTimeoutSec = 600

	transferQueryTickTime = 1 * time.Second

	transferQueryContinuesFailedTimes = 5
)

// ITransfer defines transfer interface.
type ITransfer interface {
	// GetTaskID get task id.
	GetTaskID() string

	// WaitUntilDone wait until done.
	WaitUntilDone(ctx context.Context) (*types.TransferResult, error)
}

// Transfer provides transfer.
type Transfer struct {
	taskID string

	sourceEndpoint *types.Endpoint
	targetEndpoint *types.Endpoint

	gseHandler gse.IHandler
}

// GetTaskID get task id.
func (t *Transfer) GetTaskID() string {
	return t.taskID
}

// WaitUntilDone wait until done.
func (t *Transfer) WaitUntilDone(ctx context.Context) (*types.TransferResult, error) {
	ticker := time.NewTicker(transferQueryTickTime)
	failedCnt := 0
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("context done")
		case <-ticker.C:
			src, dst, err := t.query(ctx)
			if errors.Is(err, errEndpointNotFound) {
				continue
			}

			if err != nil {
				failedCnt++
				if failedCnt > transferQueryContinuesFailedTimes {
					return nil, fmt.Errorf("failed to query file transfer result. source(%s), target(%s): %w",
						t.sourceEndpoint.AgentID, t.targetEndpoint.AgentID, err)
				}

				continue
			}
			failedCnt = 0

			if dst.StatusCode == types.TransferStatusEndDownloading {
				return dst, nil
			}

			if src.StatusCode == types.TransferStatusEndUploading && src.ErrorCode != 0 {
				dst.ErrorCode = src.ErrorCode
				dst.ErrorMessage = src.ErrorMessage

				return dst, nil
			}
		}
	}
}

// query get upload and download result.
// nolint: nonamedreturns
func (t *Transfer) query(ctx context.Context) (src *types.TransferResult, dst *types.TransferResult, err error) {
	var results []*types.TransferResult
	if results, err = t.gseHandler.QueryFileTransmissionResult(
		ctx, t.taskID, t.sourceEndpoint, t.targetEndpoint); err != nil {
		return nil, nil,
			fmt.Errorf("failed to query file transfer result. source(%s), target(%s): %w",
				t.sourceEndpoint.AgentID, t.targetEndpoint.AgentID, err)
	}

	for _, result := range results {
		// check source upload.
		if result.Mode == types.TransferModeUpload &&
			result.Source == *t.sourceEndpoint {

			src = result
		}

		// check target download.
		if result.Mode == types.TransferModeDownload &&
			result.Target == *t.targetEndpoint {

			dst = result
		}
	}

	if src == nil || dst == nil {
		return nil, nil, errEndpointNotFound
	}

	return src, dst, nil
}

var (
	errEndpointNotFound = errors.New("endpoint not found")
)

// TransferPkg transfer package.
func (m *Manager) TransferPkg(ctx context.Context, srcFilePath, dstDir string, dstHost *types.Host) (ITransfer, error) {
	if dstHost == nil {
		return nil, errors.New("destination host is nil")
	}

	sourceEndpoint, err := m.getCurrentGSEEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current endpoint: %w", err)
	}

	if _, err := os.Stat(srcFilePath); err != nil {
		return nil, fmt.Errorf("failed to stat source file(%s): %w", srcFilePath, err)
	}

	// unix user use 'root', windows use 'system'.
	srcUser := "root"
	dstUser := "root"
	if dstHost.Dynamic.NodeOsType == criteria.OSWindows {
		dstUser = "system"
	}

	taskID, err := m.gseHandler.TransferFile(ctx,
		&types.TransferOptions{
			Timeout:   transferPackageTimeoutSec,
			AutoMkdir: true,
		},
		&types.TransferDetail{
			Source: types.TransferSource{
				FileName:  filepath.Base(srcFilePath),
				StoredDir: filepath.Dir(srcFilePath),
				Endpoint: types.EndpointWithAuth{
					Endpoint: *sourceEndpoint,
					User:     srcUser,
				},
			},
			Target: types.TransferTarget{
				StoredDir: dstDir,
				Endpoints: []*types.EndpointWithAuth{
					{
						Endpoint: types.Endpoint{AgentID: dstHost.Dynamic.AgentID},
						User:     dstUser,
					},
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to transfer file via gse handler: %w", err)
	}

	return &Transfer{
		taskID:     taskID,
		gseHandler: m.gseHandler,
	}, nil
}

func (m *Manager) getCurrentGSEEndpoint(ctx context.Context) (*types.Endpoint, error) {
	host, err := m.storageTopo.GetDirectNetworkAreaHostByAnyInnerIP(ctx, m.hostAdvertiseIPV4, m.hostAdvertiseIPV6)
	if err != nil {
		return nil, fmt.Errorf("failed to get host from storage: %w", err)
	}

	containerID, err := m.getCurrentContainerID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current container-id: %w", err)
	}

	return &types.Endpoint{
		AgentID:     host.Dynamic.AgentID,
		ContainerID: containerID,
	}, nil
}

// nolint:mnd
func (m *Manager) getCurrentContainerID() (string, error) {
	if !m.inContainer {
		return "", nil
	}

	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		patterns := []string{
			"docker/",
			"kubepods/",
			"containerd/",
			"crio/",
		}

		for _, pattern := range patterns {
			if idx := strings.Index(line, pattern); idx != -1 {
				idPart := line[idx+len(pattern):]

				if end := strings.Index(idPart, ".scope"); end != -1 {
					return idPart[:end], nil
				}

				if strings.HasPrefix(idPart, "-") {
					idPart = idPart[1:]
				}

				if len(idPart) >= 64 {
					return idPart[:64], nil
				}

				return idPart, nil
			}
		}
	}

	return "", errors.New("container-id not found")
}
