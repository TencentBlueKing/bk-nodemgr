/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getNodeDeploymentNodeConf get gse node conf.
func (s *Storage) getNodeDeploymentNodeConf(nCtx contextx.IContext, token string) (*types.NodeConf, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	nodeConf, err := s.daoNodeDeployment.GetNodeDeploymentNodeConf(nCtx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get node conf: %v", err)
	}

	return nodeConf, nil
}

// getNodeDeploymentInfo get node deployment info.
func (s *Storage) getNodeDeploymentInfo(nCtx contextx.IContext, token string) (*types.DeploymentInfo, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	deployInfo, err := s.daoNodeDeployment.GetNodeDeploymentInfo(nCtx, token)
	if err != nil {
		return nil, err
	}

	return deployInfo, nil
}

// listNodeDeployment list node deployments.
func (s *Storage) listNodeDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeDeploymentCondition) (
	[]*types.NodeDeployment, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	opts, err := convertNodeDeploymentConditionToOptions(conditions)
	if err != nil {
		return nil, 0, err
	}

	nodeDeployments, total, err := s.daoNodeDeployment.ListNodeDeployment(nCtx, page, opts...)
	if err != nil {
		return nil, 0, err
	}

	return nodeDeployments, total, nil
}

// seNodeDeploymenttNodeConf set gse node conf.
func (s *Storage) seNodeDeploymenttNodeConf(nCtx contextx.IContext, token string, conf *types.NodeConf) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if conf == nil {
		return errors.New("node conf is nil")
	}

	if err := s.daoNodeDeployment.SetNodeDeploymentNodeConf(nCtx, token, conf); err != nil {
		return err
	}

	return nil
}

// updateNodeDeploymentInfo update node deployment info.
func (s *Storage) updateNodeDeploymentInfo(nCtx contextx.IContext, token string, info *types.DeploymentInfo) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if info == nil {
		return errors.New("node deployment info is nil")
	}

	if err := s.daoNodeDeployment.UpdateNodeDeploymentInfo(nCtx, token, info); err != nil {
		return err
	}

	return nil
}

// createNodeDeployment create a node deployment.
func (s *Storage) createNodeDeployment(nCtx contextx.IContext, nodeDeployment *types.NodeDeployment) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if nodeDeployment == nil {
		return errors.New("node deployment is nil")
	}

	if err := s.daoNodeDeployment.CreateNodeDeployment(nCtx, nodeDeployment); err != nil {
		return fmt.Errorf("failed to createNodeDeployment node deployment: %v", err)
	}

	return nil
}

// convertNodeDeploymentConditionToOptions convert node deployment condition to options.
func convertNodeDeploymentConditionToOptions(conditions []*types.NodeDeploymentCondition) (
	[]daoNodeDeployment.OptFn, error) {

	opts := make([]daoNodeDeployment.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoNodeDeployment.WithToken(condition.ExactInclude.Token...),
				daoNodeDeployment.WithInfoInnerIP(condition.ExactInclude.InnerIP...),
				daoNodeDeployment.WithInfoInnerIPV6(condition.ExactInclude.InnerIPv6...),
				daoNodeDeployment.WithInfoBizID(condition.ExactInclude.BizID...),
				daoNodeDeployment.WithInfoNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				daoNodeDeployment.WithInfoNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				daoNodeDeployment.WithInfoNodeVersion(condition.ExactInclude.NodeVersion...),
			)
		}

		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			return nil, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
