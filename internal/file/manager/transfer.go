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
	"regexp"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	transferPackageTimeout = 600 * time.Second

	transferQueryTickTime = 1 * time.Second

	transferQueryContinuesFailedTimes = 5
)

// Transfer provides transfer.
type Transfer struct {
	taskID string

	sourceEndpoint *types.Endpoint
	targetEndpoint *types.Endpoint

	fileInfo iface.FileInfo

	gseHandler gse.IHandler
}

// GetTaskID get task id.
func (t *Transfer) GetTaskID() string {
	return t.taskID
}

// GetFileInfo get file info.
func (t *Transfer) GetFileInfo() iface.FileInfo {
	return t.fileInfo
}

// WaitUntilDone wait until done.
func (t *Transfer) WaitUntilDone(ctx context.Context) (*types.SimpleTransferResult, error) {
	ticker := time.NewTicker(transferQueryTickTime)
	failedCnt := 0
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("context done")
		case <-ticker.C:
			src, dst, err := t.query(ctx)
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
				return types.ConvertTransferResultToSimple(dst), nil
			}

			if src.StatusCode == types.TransferStatusEndUploading && src.ErrorCode != 0 {
				// if upload failed, set the upload error info into simple result.
				result := types.ConvertTransferResultToSimple(dst)
				result.ErrorCode = src.ErrorCode
				result.ErrorMessage = src.ErrorMessage

				return result, nil
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

	return src, dst, nil
}

// QueryTransfer query transfer.
func (m *Manager) QueryTransfer(
	ctx context.Context, taskID string) (*types.SimpleTransferResult, *types.SimpleTransferResult, error) {

	results, err := m.gseHandler.QueryFileTransmissionResult(ctx, taskID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query file transfer result. task-id(%s): %w", taskID, err)
	}

	if len(results) != 2 { // nolint: mnd
		return nil, nil, fmt.Errorf("unexpected transfer result count(%d), results(%+v)", len(results), results)
	}

	sr1 := types.ConvertTransferResultToSimple(results[0])
	sr2 := types.ConvertTransferResultToSimple(results[1])

	if sr1.Mode == types.TransferModeUpload && sr2.Mode == types.TransferModeDownload {
		return sr1, sr2, nil
	}

	if sr2.Mode == types.TransferModeUpload && sr1.Mode == types.TransferModeDownload {
		return sr2, sr1, nil
	}

	return nil, nil, fmt.Errorf("unexpected transfer result: (%v), (%v)", results[0], results[1])
}

// LaunchTransferRelease launch transfer release.
func (m *Manager) LaunchTransferRelease(ctx context.Context,
	gen types.Generation,
	rt types.ReleaseType,
	plat platform.Platform,
	version string,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	file, dir, err := m.EnsureFileToLocal(ctx, gen, rt, plat, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get release file: %w", err)
	}

	info := file.Info()
	fp := filepath.Join(dir, info.Name)

	tf, err := m.transferPkg(ctx, fp, dstDir, dstHost)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer package: %w", err)
	}

	tf.fileInfo = info

	return tf, nil
}

// LaunchTransferInstaller launch transfer installer.
func (m *Manager) LaunchTransferInstaller(ctx context.Context,
	plat platform.Platform,
	dstDir string,
	dstHost *types.Host) (types.ISimpleTransferHandler, error) {

	toolName, err := tool.FormatInstallerName(plat.OS, plat.Arch)
	if err != nil {
		return nil, fmt.Errorf("failed to format installer name: %w", err)
	}

	toolFile, err := m.installerFileGroup.GetFile(ctx, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file, err: %w", err)
	}

	fp := local.GetLocalFileAbsFilePath(toolFile)
	tf, err := m.transferPkg(ctx, fp, dstDir, dstHost)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer package: %w", err)
	}

	tf.fileInfo = toolFile.Info()

	return tf, nil
}

// TransferPkg transfer package.
func (m *Manager) transferPkg(ctx context.Context, srcFilePath, dstDir string, dstHost *types.Host) (*Transfer, error) {
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
			Timeout:   transferPackageTimeout,
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

	agentID := host.Dynamic.AgentID
	if agentID == "" {
		agentID = host.Static.SyncedAgentID
	}

	if agentID == "" {
		return nil, errors.New("agent-id is empty")
	}

	containerID, err := m.getCurrentContainerID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current container-id: %w", err)
	}

	ep := &types.Endpoint{
		AgentID:     agentID,
		ContainerID: containerID,
	}
	m.logger.InfoCtxf(ctx, "got current service endpoint. ipv4(%s), ipv6(%s): %+v",
		m.hostAdvertiseIPV4, m.hostAdvertiseIPV6, ep)

	return ep, nil
}

// nolint:mnd,gocognit
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
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		cgroupPath := parts[2]
		if cgroupPath == "" {
			continue
		}

		lastID := ""
		for _, item := range strings.Split(cgroupPath, "/") {
			// docker format (docker-<id>.scope)
			if strings.HasPrefix(item, "docker-") && strings.HasSuffix(item, ".scope") {
				id := strings.TrimSuffix(strings.TrimPrefix(item, "docker-"), ".scope")
				if isValidContainerID(id) {
					lastID = id

					continue
				}
			}

			// containerd format (cri-containerd-<id>.scope)
			if strings.HasPrefix(item, "cri-containerd-") && strings.HasSuffix(item, ".scope") {
				id := strings.TrimSuffix(strings.TrimPrefix(item, "cri-containerd-"), ".scope")
				if isValidContainerID(id) {
					lastID = id

					continue
				}
			}

			// pure container-id.
			if isValidContainerID(item) {
				lastID = item
			}
		}

		if lastID != "" {
			return lastID, nil
		}
	}

	return "", errors.New("container-id not found")
}

func isValidContainerID(id string) bool {
	matched, _ := regexp.MatchString(`^[0-9a-f]{64}$`, id)
	return matched
}
