/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package configfetcher this package provide config fetch methods.
package configfetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils/downloader"
)

// Step fetch configs step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	CallbackSvrAddr    []string
	NodeRole           types.NodeRole
	DeployToken        string
	Generation         types.Generation
	ConfigSavedDir     string
	CheckListSavedPath string

	// SelectFetchs set false by default, will fetch all things.
	// set true, then will only fetch the enabled ones following.
	SelectFetchs         bool
	EnableFetchConfig    bool
	EnableFetchChecklist bool
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("generation(%d), token(%s), role(%s)",
		int(args.Generation), args.DeployToken, args.NodeRole)
}

// NewStep new a step to fetch configs.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to fetch configs.
// nolint: funlen,gocognit
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepFetchConfigs, "start to fetch configs. %s", step.args.String())

	gp := gopool.NewPool()

	if !step.args.SelectFetchs || step.args.EnableFetchConfig {
		// fetch agent config files.
		gp.Go(func() error {
			return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
				return step.fetchAgentConfig(ctx)
			})
		})

		// fetch proxy config files.
		if step.args.NodeRole == types.NodeRoleProxy {
			gp.Go(func() error {
				return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
					return step.fetchFileProxyConfig(ctx)
				})
			})

			gp.Go(func() error {
				return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
					return step.fetchDataProxyConfig(ctx)
				})
			})
		}
	}

	if !step.args.SelectFetchs || step.args.EnableFetchChecklist {
		// fetch check list.
		gp.Go(func() error {
			return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
				return step.fetchCheckList(ctx)
			})
		})
	}

	if err := gp.Wait(); err != nil {
		logger.Infof(node.StepFetchConfigs, "failed to fetch configs. %s: %v", step.args.String(), err)
		return fmt.Errorf("failed to fetch configs: %w", err)
	}

	logger.Infof(node.StepFetchConfigs, "successfully fetched all configs")

	return nil
}

const (
	maxTime = 300 * time.Second

	// API paths for node installer fetch configs.
	getAgentConfigPath     = "/api/v3/callback/workflow/node_install/get_agent_config"
	getFileProxyConfigPath = "/api/v3/callback/workflow/node_install/get_file_proxy_config"
	getDataProxyConfigPath = "/api/v3/callback/workflow/node_install/get_data_proxy_config"
	getCheckListPath       = "/api/v3/callback/workflow/node_install/get_check_list"
)

// fetchConfigsMultiEndpoint tries multiple server addresses in order until one succeeds.
func (step *Step) fetchConfigsMultiEndpoint(ctx context.Context, reqBody any, serverAddrs []string, subURL, savedPath string) error {
	if len(serverAddrs) == 0 {
		return fmt.Errorf("no server addresses provided")
	}

	var lastErr error
	for i, serverAddr := range serverAddrs {
		logger.Infof(node.StepFetchConfigs, "attempting to fetch from server, index(%d/%d), url(%s)", i+1, len(serverAddrs), serverAddr)
		err := step.fetchConfig(ctx, reqBody, serverAddr, subURL, savedPath)
		if err == nil {
			return nil
		}
		logger.Warnf(node.StepFetchConfigs, "failed to fetch: %v", err)
		lastErr = err
	}

	// All servers failed, return error with server addresses for debugging
	return fmt.Errorf("failed to fetch from all %d server(s) %v: %w", len(serverAddrs), serverAddrs, lastErr)
}

func (step *Step) fetchConfig(ctx context.Context, reqBody any, baseURL, subURL, savedPath string) error {
	fetchURL, err := url.JoinPath(baseURL, subURL)
	if err != nil {
		logger.Errorf(node.StepFetchConfigs, "failed to join path(%s, %s): %v", baseURL, subURL, err)
		return fmt.Errorf("fetch config failed: %v", err)
	}

	config := downloader.Config{
		URL:         fetchURL,
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
	config.ProgressFunc = func(current, total int64) {
		if current == total {
			logger.Infof(node.StepFetchConfigs, "config fetch completed. file(%s)", savedPath)
		}

		if current-lastProgress < total/10 {
			return
		}

		logger.Infof(node.StepFetchConfigs,
			"config fetching. file(%s), progress(%d/%d)", savedPath, current, total)

		lastProgress = current
	}

	d := new(downloader.HTTPDownloader)
	if err := d.Download(ctx, config); err != nil {
		return fmt.Errorf("failed to fetch config. url(%s), file(%s), req-body(%+v): %w",
			fetchURL, savedPath, reqBody, err)
	}

	return nil
}

func (step *Step) fetchAgentConfig(ctx context.Context) error {
	type getAgentConfigReq struct {
		OSType   string `json:"os_type"`
		CPUArch  string `json:"cpu_arch"`
		NodeRole string `json:"node_role"`
		Token    string `json:"token"`
	}

	requestBody := getAgentConfigReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: string(step.args.NodeRole),
		Token:    step.args.DeployToken,
	}

	savedPath := filepath.Join(step.args.ConfigSavedDir, "gse_agent.conf")
	if err := step.fetchConfigsMultiEndpoint(ctx,
		requestBody,
		step.args.CallbackSvrAddr,
		getAgentConfigPath,
		savedPath); err != nil {
		logger.Errorf(node.StepFetchConfigs, "failed to get agent config: %v", err)

		return fmt.Errorf("failed to get agent config: %w", err)
	}

	logger.Infof(node.StepFetchConfigs, "successfully fetched agent-config(%s)", savedPath)

	return nil
}

func (step *Step) fetchFileProxyConfig(ctx context.Context) error {
	type getFileProxyConfReq struct {
		OSType   string `json:"os_type"`
		CPUArch  string `json:"cpu_arch"`
		NodeRole string `json:"node_role"`
		Token    string `json:"token"`
	}

	requestBody := getFileProxyConfReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: string(step.args.NodeRole),
		Token:    step.args.DeployToken,
	}

	savedPath := filepath.Join(step.args.ConfigSavedDir, "gse_file_proxy.conf")
	if err := step.fetchConfigsMultiEndpoint(ctx,
		requestBody,
		step.args.CallbackSvrAddr,
		getFileProxyConfigPath,
		savedPath); err != nil {
		logger.Errorf(node.StepFetchConfigs, "failed to get file proxy config: %v", err)

		return fmt.Errorf("failed to get file proxy config: %w", err)
	}

	logger.Infof(node.StepFetchConfigs, "successfully fetched file-proxy-config(%s)", savedPath)

	return nil
}

func (step *Step) fetchDataProxyConfig(ctx context.Context) error {
	type getDataProxyConfReq struct {
		OSType   string `json:"os_type"`
		CPUArch  string `json:"cpu_arch"`
		NodeRole string `json:"node_role"`
		Token    string `json:"token"`
	}

	requestBody := getDataProxyConfReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: string(step.args.NodeRole),
		Token:    step.args.DeployToken,
	}

	savedPath := filepath.Join(step.args.ConfigSavedDir, "gse_data_proxy.conf")
	if err := step.fetchConfigsMultiEndpoint(ctx,
		requestBody,
		step.args.CallbackSvrAddr,
		getDataProxyConfigPath,
		savedPath); err != nil {
		logger.Errorf(node.StepFetchConfigs, "failed to get data proxy config: %v", err)

		return fmt.Errorf("failed to get data proxy config: %w", err)
	}

	logger.Infof(node.StepFetchConfigs, "successfully fetched data-proxy-config(%s)", savedPath)

	return nil
}

func (step *Step) fetchCheckList(ctx context.Context) error {
	type getCheckListReq struct {
		OSType   string `json:"os_type"`
		CPUArch  string `json:"cpu_arch"`
		NodeRole string `json:"node_role"`
		Token    string `json:"token"`
	}

	requestBody := getCheckListReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: string(step.args.NodeRole),
		Token:    step.args.DeployToken,
	}

	if err := step.fetchConfigsMultiEndpoint(ctx,
		requestBody,
		step.args.CallbackSvrAddr,
		getCheckListPath,
		step.args.CheckListSavedPath); err != nil {
		logger.Errorf(node.StepFetchConfigs, "failed to get check list: %v", err)

		return fmt.Errorf("failed to get check list: %w", err)
	}

	logger.Infof(node.StepFetchConfigs, "successfully fetched check-list(%s)",
		step.args.CheckListSavedPath)

	return nil
}
