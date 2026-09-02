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

package dpmgr

import (
	"fmt"
	"path"
	"strings"
)

func genPluginNameForSpecifyPluginPkg(pluginPkgName string, deployPolicyID int64, moduleID int64) string {
	return fmt.Sprintf("%s_%d_%d", pluginPkgName, deployPolicyID, moduleID)
}

func genProcessUniqueID(hostID int64, pluginName string) string {
	return fmt.Sprintf("%d_%s", hostID, pluginName)
}

func genDeployPolicySubConfigName(configName string, deployPolicyID int64) string {
	ext := path.Ext(configName)
	baseName := strings.TrimSuffix(configName, ext)

	return fmt.Sprintf("%s_deploy_%d%s", baseName, deployPolicyID, ext)
}

func genDeployPolicySubConfigNameByConfigTemplateName(configTemplateName string, deployPolicyID int64) string {
	ext := path.Ext(configTemplateName)
	baseName := strings.TrimSuffix(configTemplateName, ext)

	return fmt.Sprintf("%s_deploy_%d%s", baseName, deployPolicyID, ext)
}

func genDeployPolicyProcessConfigSet(deployPolicyID int64) string {
	return fmt.Sprintf(processConfigSetDeployPolicyFormat, deployPolicyID)
}

// DSU disjoint set Union.
type DSU[T ~int64] struct {
	father map[T]T
}

// NewDsu new a disjoint set union.
func NewDsu[T ~int64]() *DSU[T] {
	return &DSU[T]{
		father: make(map[T]T),
	}
}

// Find x's father.
func (d *DSU[T]) Find(x T) T {
	if _, ok := d.father[x]; !ok {
		d.father[x] = x
	}

	if d.father[x] != x {
		d.father[x] = d.Find(d.father[x])
	}

	return d.father[x]
}

// Union x and y.
func (d *DSU[T]) Union(x, y T) {
	r1, r2 := d.Find(x), d.Find(y)
	d.father[r2] = r1
}

// Judge whether x and y are in the same set.
func (d *DSU[T]) Judge(x, y T) bool {
	return d.Find(x) == d.Find(y)
}
