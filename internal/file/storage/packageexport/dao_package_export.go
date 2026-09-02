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

package packageexport

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-export"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) listPackageExport(
	nCtx contextx.IContext, page types.Page, conditions ...*types.PackageExportCondition) (
	[]*types.PackageExport, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	opts, err := convertPackageExportConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	exports, count, err := s.daoPackageExport.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package exports: %w", err)
	}

	return exports, count, nil
}

func convertPackageExportConditionsToOptions(conditions ...*types.PackageExportCondition) (
	[]packageexport.OptFn, error) {

	opts := make([]packageexport.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				packageexport.WithExportID(condition.ExactInclude.ExportID...),
				packageexport.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				packageexport.WithTenantID(condition.ExactInclude.TenantID...),
				packageexport.WithOperator(condition.ExactInclude.Operator...),
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

func (s *Storage) createPackageExport(nCtx contextx.IContext, exportData *types.PackageExport) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}
	if exportData == nil {
		return errors.New("package export is nil")
	}

	if err := s.daoPackageExport.Create(nCtx, exportData); err != nil {
		return fmt.Errorf("failed to create package export: %w", err)
	}

	return nil
}

func (s *Storage) getPackageExport(nCtx contextx.IContext, exportID string) (*types.PackageExport, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}
	if exportID == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	exportData, err := s.daoPackageExport.Get(nCtx, exportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package export: %w", err)
	}

	return exportData, nil
}

func (s *Storage) deletePackageExport(nCtx contextx.IContext, exportID string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}
	if exportID == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if err := s.daoPackageExport.Delete(nCtx, exportID); err != nil {
		return fmt.Errorf("failed to delete package export: %w", err)
	}

	return nil
}
