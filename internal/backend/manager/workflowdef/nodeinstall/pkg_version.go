/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

func autoSelectVersion(ctx *action.InstanceContext, osType criteria.OSType,
	cpuArch criteria.CPUArch) (string, error) {

	return "", errors.New("not implemented")
}

func checkVersionAvailability(ctx *action.InstanceContext, osType criteria.OSType,
	cpuArch criteria.CPUArch, version string) error {

	// TODO: implement me

	return nil
}
