//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package uninstallnode

// Step uninstall agent step.
type Step struct {
	setupDirPath string
	binDirPath   string
	gseCtlPath   string
	deployEnv    string
}

// StepArgs define args for step.
type StepArgs struct {
	SetupDirPath string
	BinDirPath   string
	GseCtlPath   string
	DeployEnv    string
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	step := &Step{
		setupDirPath: args.SetupDirPath,
		binDirPath:   args.BinDirPath,
		gseCtlPath:   args.GseCtlPath,
		deployEnv:    args.DeployEnv,
	}

	return step
}
