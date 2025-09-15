/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// create plugin deployment.
func (s *Storage) create(ctx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if pluginDeployment == nil {
		return errors.New("plugin deployment is nil")
	}

	if err := s.daoPluginDeployment.Create(ctx, pluginDeployment); err != nil {
		return err
	}

	return nil
}

// getInfo get plugin deployment info.
func (s *Storage) getInfo(ctx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	info, err := s.daoPluginDeployment.GetInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// updateInfo update a node deployment info.
func (s *Storage) updateInfo(ctx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if pluginDeploymentInfo == nil {
		return errors.New("plugin deployment info is nil")
	}

	if err := s.daoPluginDeployment.UpdateInfo(ctx, token, pluginDeploymentInfo); err != nil {
		return fmt.Errorf("update plugin deployment info failed, err: %v", err)
	}

	return nil
}
