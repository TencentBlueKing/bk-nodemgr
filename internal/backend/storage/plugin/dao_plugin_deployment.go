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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// createPluginDeployment plugin deployment.
func (s *Storage) createPluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if pluginDeployment == nil {
		return errors.New("plugin deployment is nil")
	}

	if err := s.daoPluginDeployment.Create(nCtx, pluginDeployment); err != nil {
		return err
	}

	return nil
}

// getPluginDeploymentInfo get plugin deployment info.
func (s *Storage) getPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	info, err := s.daoPluginDeployment.GetInfo(nCtx, token)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// updatePluginDeploymentInfo update a node deployment info.
func (s *Storage) updatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if pluginDeploymentInfo == nil {
		return errors.New("plugin deployment info is nil")
	}

	if err := s.daoPluginDeployment.UpdateInfo(nCtx, token, pluginDeploymentInfo); err != nil {
		return fmt.Errorf("update plugin deployment info failed: %v", err)
	}

	return nil
}

// getPluginDeploymentPluginConf get plugin deployment plugin conf.
func (s *Storage) getPluginDeploymentPluginConf(ctx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	pluginConf, err := s.daoPluginDeployment.GetPluginConf(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin deployment plugin conf: %v", err)
	}

	return pluginConf, nil
}

// updatePluginDeploymentPluginConf set plugin deployment plugin conf.
func (s *Storage) updatePluginDeploymentPluginConf(ctx contextx.IContext, token string, pluginConf *types.PluginDeploymentPluginConf) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if pluginConf == nil {
		return errors.New("plugin conf is nil")
	}

	if err := s.daoPluginDeployment.UpdatePluginConf(ctx, token, pluginConf); err != nil {
		return fmt.Errorf("failed to set plugin deployment plugin conf: %w", err)
	}

	return nil
}

// getPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config files detail.
func (s *Storage) getPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) ([]*types.PluginConfigDetail, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	configs, err := s.daoPluginDeployment.GetPluginConfConfigFilesDetail(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin deployment config: %v", err)
	}

	return configs, nil
}
