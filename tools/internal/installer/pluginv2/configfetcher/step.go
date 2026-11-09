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

// Package configfetcher fetch configs step.
package configfetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
)

// Step fetch configs step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	CallbackSvrAddr []string

	PluginGroup    string
	PluginName     string
	PluginPkgName  string
	DeployToken    string
	ConfigSavedDir string

	// SelectFetchs set false by default, will fetch all things.
	// set true, then will only fetch the enabled ones following.
	SelectFetchs      bool
	EnableFetchConfig bool
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("plugin-type(%s), plugin-name(%s), plugin-pkg-name(%s), deploy-token(%s)",
		args.PluginGroup, args.PluginName, args.PluginPkgName, args.DeployToken)
}

// NewStep new a step to fetch configs.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to fetch configs.
// nolint: funlen,gocognit
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(pluginv2.StepFetchConfigs, "start to fetch configs. %s", step.args.String())

	gp := gopool.NewPool()

	if !step.args.SelectFetchs || step.args.EnableFetchConfig {
		// fetch plugin config files.
		gp.Go(func() error {
			return retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault()).Do(ctx, func(_ int) error {
				return step.fetchPluginConfig(ctx)
			})
		})
	}

	if err := gp.Wait(); err != nil {
		logger.Infof(pluginv2.StepFetchConfigs, "failed to fetch configs. %s: %v", step.args.String(), err)
		return fmt.Errorf("failed to fetch configs: %w", err)
	}

	logger.Infof(pluginv2.StepFetchConfigs, "successfully fetched all configs")

	return nil
}

const (
	// getPluginConfigPath is the callback api path to fetch all main config files of the plugin deployment.
	getPluginConfigPath = "/api/v3/callback/workflow/plugin/get_main_config"

	// mainConfigFetchTimeout is the timeout for fetching the main config callback api.
	mainConfigFetchTimeout = 60 * time.Second

	// mainConfigFileMode is the permission of the saved main config files, they are copied to the plugin config dir.
	mainConfigFileMode = 0600
)

// getPluginConfigReq is the request body of the main config callback api.
type getPluginConfigReq struct {
	OSType  string `json:"os_type"`
	CPUArch string `json:"cpu_arch"`
	Token   string `json:"token"`
}

// mainConfigsCallbackResp is the response body of the main config callback api.
type mainConfigsCallbackResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		MainConfigs []struct {
			FileName string `json:"file_name"`
			Content  string `json:"content"`
		} `json:"main_configs"`
	} `json:"data"`
}

func (step *Step) fetchPluginConfig(ctx context.Context) error {
	requestBody := getPluginConfigReq{
		OSType:  runtime.GOOS,
		CPUArch: runtime.GOARCH,
		Token:   step.args.DeployToken,
	}

	if len(step.args.CallbackSvrAddr) == 0 {
		return fmt.Errorf("no server addresses provided")
	}

	var lastErr error
	for i, serverAddr := range step.args.CallbackSvrAddr {
		logger.Infof(pluginv2.StepFetchConfigs,
			"attempting to fetch main configs from server, index(%d/%d), url(%s)", i+1, len(step.args.CallbackSvrAddr), serverAddr)

		err := step.fetchMainConfigs(ctx, requestBody, serverAddr)
		if err == nil {
			return nil
		}

		logger.Warnf(pluginv2.StepFetchConfigs, "failed to fetch main configs: %v", err)
		lastErr = err
	}

	return fmt.Errorf("failed to fetch main configs from all %d server(s) %v: %w",
		len(step.args.CallbackSvrAddr), step.args.CallbackSvrAddr, lastErr)
}

// fetchMainConfigs fetches all main config files from one callback server and saves them into the config dir.
func (step *Step) fetchMainConfigs(ctx context.Context, requestBody getPluginConfigReq, baseURL string) error {
	fetchURL, err := url.JoinPath(baseURL, getPluginConfigPath)
	if err != nil {
		logger.Errorf(pluginv2.StepFetchConfigs, "failed to join path(%s, %s): %v", baseURL, getPluginConfigPath, err)

		return fmt.Errorf("failed to join main configs path: %w", err)
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal main configs request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fetchURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create main configs request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: mainConfigFetchTimeout}
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send main configs request. url(%s): %w", fetchURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Errorf(pluginv2.StepFetchConfigs,
			"failed to fetch main configs. url(%s), resp-code(%d), resp-body(%s)", fetchURL, resp.StatusCode, respBody)

		return fmt.Errorf("failed to fetch main configs. url(%s), resp-code(%d)", fetchURL, resp.StatusCode)
	}

	respData := new(mainConfigsCallbackResp)
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return fmt.Errorf("failed to decode main configs response. url(%s): %w", fetchURL, err)
	}

	for _, configFile := range respData.Data.MainConfigs {
		// all main config files are placed in the plugin config dir directly, so the file name must be plain.
		if configFile.FileName == "" || configFile.FileName != filepath.Base(configFile.FileName) || strings.Contains(configFile.FileName, `\`) {
			return fmt.Errorf("invalid main config file name(%s)", configFile.FileName)
		}

		savedPath := filepath.Join(step.args.ConfigSavedDir, configFile.FileName)
		if err := os.WriteFile(savedPath, []byte(configFile.Content), mainConfigFileMode); err != nil {
			return fmt.Errorf("failed to save main config file(%s): %w", savedPath, err)
		}

		logger.Infof(pluginv2.StepFetchConfigs, "saved main config file(%s), size(%d)", savedPath, len(configFile.Content))
	}

	return nil
}
