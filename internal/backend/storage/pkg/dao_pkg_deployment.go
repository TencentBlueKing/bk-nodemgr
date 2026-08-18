/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	packagedeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createPackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if deployment == nil {
		return errors.New("package deployment is nil")
	}

	return s.daoPackageDeployment.CreatePackageDeployment(nCtx, deployment)
}

func (s *Storage) listPackageDeployment(nCtx contextx.IContext, page types.Page, opts ...packagedeployment.OptFn) (
	[]*types.PackageDeployment, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	return s.daoPackageDeployment.ListPackageDeployment(nCtx, page, opts...)
}

func (s *Storage) getPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	return s.daoPackageDeployment.GetPackageDeploymentInfo(nCtx, token)
}

func (s *Storage) updatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if info == nil {
		return errors.New("package deployment info is nil")
	}

	return s.daoPackageDeployment.UpdatePackageDeploymentInfo(nCtx, token, info)
}
