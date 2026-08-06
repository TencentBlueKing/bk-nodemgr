// Tencent is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
// Copyright (C) 2017 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT

package workflow

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createPackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow is nil")
	}

	if err := workflow.Type.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.type: %w", err)
	}

	if err := workflow.Status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.status: %w", err)
	}

	if err := s.daoPackageWorkflow.Create(nCtx, workflow); err != nil {
		return fmt.Errorf("failed to create package workflow: %w", err)
	}

	return nil
}
