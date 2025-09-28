/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package filedownloader

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils/downloader"
)

// Step download files step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	DownloadSvrAddr string
	CallbackSvrAddr string

	PluginType     types.PluginType
	PluginName     string
	DeployToken    string
	PkgVersion     string
	PkgSavedPath   string
	ConfigSavedDir string

	// SelectDownloads set false by default, will download all things.
	// set true, then will only download the enabled ones following.
	SelectDownloads              bool
	EnableDownloadConfig         bool
	EnableDownloadReleasePackage bool
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("plugin-type(%s), plugin-name(%s), deploy-token(%s), pkg-version(%s)",
		args.PluginType, args.PluginName, args.DeployToken, args.PkgVersion)
}

// NewStep new a step to download package.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to download files.
// nolint: funlen,gocognit
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(plugin.StepDownloadFiles, "start to download files. %s", step.args.String())

	gp := gopool.NewPool()

	if !step.args.SelectDownloads || step.args.EnableDownloadConfig {
		// download plugin config files.
		gp.Go(func() error {
			return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
				return step.downloadPluginConfig(ctx)
			})
		})
	}

	if !step.args.SelectDownloads || step.args.EnableDownloadReleasePackage {
		// download release packages.
		gp.Go(func() error {
			return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
				return step.downloadReleasePackage(ctx)
			})
		})
	}

	if err := gp.Wait(); err != nil {
		logger.Infof(plugin.StepDownloadFiles, "failed to download files. %s: %v", step.args.String(), err)
		return fmt.Errorf("failed to download files: %w", err)
	}

	logger.Infof(plugin.StepDownloadFiles, "successfully downloaded all files")

	return nil
}

const (
	maxTime = 300 * time.Second
)

func (step *Step) downloadFile(ctx context.Context, reqBody any, baseURL, subURL, savedPath string) error {
	downloadURL, err := url.JoinPath(baseURL, subURL)
	if err != nil {
		logger.Errorf(plugin.StepDownloadFiles, "failed to join path(%s, %s): %v", baseURL, subURL, err)
		return fmt.Errorf("download file failed: %v", err)
	}

	downloadConfig := downloader.Config{
		URL:         downloadURL,
		Method:      http.MethodPost,
		RequestBody: reqBody,
		DestPath:    savedPath,
		Timeout:     maxTime,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	lastProgress := int64(0)

	// start progress report.
	downloadConfig.ProgressFunc = func(current, total int64) {
		if current == total {
			logger.Infof(plugin.StepDownloadFiles, "file downloading completed. file(%s)", savedPath)
		}

		if current-lastProgress < total/10 {
			return
		}

		logger.Infof(plugin.StepDownloadFiles,
			"file downloading. file(%s), progress(%d/%d)", savedPath, current, total)

		lastProgress = current
	}

	d := new(downloader.HTTPDownloader)
	if err := d.Download(ctx, downloadConfig); err != nil {
		return fmt.Errorf("failed to download file. url(%s), file(%s), req-body(%+v): %w",
			downloadURL, savedPath, reqBody, err)
	}

	return nil
}

func pluginConfName(pluginName string) string {
	return fmt.Sprintf("%s.conf", pluginName)
}

func (step *Step) downloadPluginConfig(ctx context.Context) error {
	type getPluginConfigReq struct {
		OSType  string `json:"os_type"`
		CPUArch string `json:"cpu_arch"`
		Token   string `json:"token"`
	}

	requestBody := getPluginConfigReq{
		OSType:  runtime.GOOS,
		CPUArch: runtime.GOARCH,
		Token:   step.args.DeployToken,
	}

	savedPath := filepath.Join(step.args.ConfigSavedDir, pluginConfName(step.args.PluginName))
	if err := step.downloadFile(ctx,
		requestBody,
		step.args.CallbackSvrAddr,
		"/callback/workflow/plugin/get_main_config",
		savedPath); err != nil {
		logger.Errorf(plugin.StepDownloadFiles, "failed to get config: %v", err)

		return fmt.Errorf("failed to get plugin config: %w", err)
	}

	logger.Infof(plugin.StepDownloadFiles, "successfully downloaded plugin-config(%s)", savedPath)

	return nil
}

func (step *Step) downloadReleasePackage(ctx context.Context) error {
	type getReleasePackageReq struct {
		OSType     string `json:"os_type"`
		CPUArch    string `json:"cpu_arch"`
		Version    string `json:"version"`
		PluginName string `json:"plugin_name"`
		PluginType string `json:"plugin_type"`
	}

	requestBody := getReleasePackageReq{
		OSType:     runtime.GOOS,
		CPUArch:    runtime.GOARCH,
		PluginName: step.args.PluginName,
		PluginType: string(step.args.PluginType),
		Version:    step.args.PkgVersion,
	}

	if err := step.downloadFile(ctx,
		requestBody,
		step.args.DownloadSvrAddr,
		"/download/plugin",
		step.args.PkgSavedPath); err != nil {
		logger.Errorf(plugin.StepDownloadFiles, "failed to get release package: %v", err)

		return fmt.Errorf("failed to get release package: %w", err)
	}

	logger.Infof(plugin.StepDownloadFiles, "successfully downloaded release package(%s)",
		step.args.PkgSavedPath)

	return nil
}
