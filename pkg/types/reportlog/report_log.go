/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package reportlog ...
package reportlog

import "time"

// ReportLog the log report from agent.
type ReportLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Step      Step      `json:"step"`
	Log       string    `json:"log"`
	Status    string    `json:"status"`
}

// Step define the step of shell script.
type Step string

const (
	// StepCheckEnv the step of check env.
	StepCheckEnv Step = "check_env"

	// StepDownloadPkg the step of download pkg.
	StepDownloadPkg Step = "download_pkg"

	// StepRemoveAgent the step of remove agent.
	StepRemoveAgent Step = "remove_agent"

	// StepRemoveProxyIfExists the step of remove proxy if exists.
	StepRemoveProxyIfExists Step = "remove_proxy_if_exists"

	// StepSetupAgent the step of setup agent.
	StepSetupAgent Step = "setup_agent"

	// StepCheckDeployResult the step of check deploy result.
	StepCheckDeployResult Step = "check_deploy_result"
)
