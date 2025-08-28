/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// IOfficialPlugin defines the interface of official plugin.
type IOfficialPlugin interface {
	// ExistReleaseOfficialPlugin checks if release official plugin exists.
	ExistReleaseOfficialPlugin(ctx contextx.IContext, pluginName string, version string, plat ...platform.Platform) (bool, error)
}

// ExistReleaseOfficialPlugin checks if release plugin exists.
func (s *Storage) ExistReleaseOfficialPlugin(ctx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (bool, error) {
	// TODO: implement me

	return false, nil
}
