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

package pluginrequirement

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-requirement"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createPluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if err := s.daoPluginRequirement.Create(nCtx, record); err != nil {
		return fmt.Errorf("failed to create plugin requirement: %w", err)
	}

	return nil
}

func (s *Storage) updatePluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if err := s.daoPluginRequirement.Update(nCtx, record); err != nil {
		return fmt.Errorf("failed to update plugin requirement: %w", err)
	}

	return nil
}

func (s *Storage) updatePluginRequirementScope(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error {
	if record == nil {
		return errors.New("plugin requirement file record is nil")
	}
	if err := s.daoPluginRequirement.UpdateScope(nCtx, record); err != nil {
		return fmt.Errorf("failed to update plugin requirement scope: %w", err)
	}

	return nil
}

func (s *Storage) listPluginRequirements(nCtx contextx.IContext, page types.Page,
	condition *types.PluginRequirementFileCondition) ([]*types.PluginRequirementFileRecord, int64, error) {

	var opts []pluginrequirement.OptFn
	if condition != nil && condition.ExactInclude != nil {
		exact := condition.ExactInclude
		opts = append(opts,
			pluginrequirement.WithID(exact.ID...),
			pluginrequirement.WithName(exact.Name...),
			pluginrequirement.WithOS(exact.OS...),
			pluginrequirement.WithArch(exact.Arch...),
			pluginrequirement.WithVersion(exact.Version...))
	}
	records, total, err := s.daoPluginRequirement.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plugin requirements: %w", err)
	}

	return records, total, nil
}

func (s *Storage) getPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error) {
	if id == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}
	record, err := s.daoPluginRequirement.Get(nCtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement: %w", err)
	}

	return record, nil
}
