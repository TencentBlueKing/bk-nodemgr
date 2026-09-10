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

package iamv4

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerAction exposes action operations without HTTP details.
type IHandlerAction interface {
	ListActions(ctx contextx.IContext, systemID string) ([]Action, error)
	CreateAction(ctx contextx.IContext, systemID string, action Action) error
	UpdateAction(ctx contextx.IContext, systemID, actionID, name string) error
}

// ListActions retrieves the complete action model or returns an error.
func (h *Handler) ListActions(ctx contextx.IContext, systemID string) ([]Action, error) {
	const pageSize = 100             // IAM's maximum model query page size.
	const timeout = 30 * time.Second // Bound the complete model query, not each page.
	seen := make(map[string]struct{})
	total := 0
	requestID := ""
	executor := pageexecutor.NewPageExecutor[Action](pageSize, timeout)
	fn := func(ctx contextx.IContext, p types.Page) ([]Action, error) {
		if p.Offset > 0 && len(seen) == total {
			// The last full page already completed the model.
			return nil, nil
		}
		req := &ListActionsReq{SystemID: systemID, Page: p.Offset/pageSize + 1, PageSize: pageSize}
		resp, err := h.cli.listActions(ctx, req)
		if err != nil {
			return nil, err
		}
		if err := validateActionPage(req, resp, total, seen); err != nil {
			return nil, err
		}
		total = *resp.Data.Count
		requestID = resp.RequestID

		return resp.Data.Results, nil
	}
	actions, err := executor.Execute(ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("list actions: %w", err)
	}
	if len(actions.Items) != total {
		return nil, fmt.Errorf("list actions: incomplete pagination, got %d of %d (request-id: %s)", len(actions.Items), total, requestID)
	}

	return actions.Items, nil
}

func validateActionPage(
	req *ListActionsReq, resp *BaseBroker[*ListActionsResp], total int, seen map[string]struct{},
) error {

	if resp.Data == nil || resp.Data.Count == nil || *resp.Data.Count < 0 || resp.Data.Results == nil {
		return fmt.Errorf("list actions page %d: incomplete response (request-id: %s)", req.Page, resp.RequestID)
	}
	if req.Page == 1 {
		total = *resp.Data.Count
	}
	if *resp.Data.Count != total || len(resp.Data.Results) > req.PageSize || len(resp.Data.Results) > total-len(seen) {
		return fmt.Errorf("list actions page %d: inconsistent count (request-id: %s)", req.Page, resp.RequestID)
	}
	for _, action := range resp.Data.Results {
		if action.ID == "" || action.Name == "" {
			return fmt.Errorf("list actions page %d: malformed action (request-id: %s)", req.Page, resp.RequestID)
		}
		if _, exists := seen[action.ID]; exists {
			return fmt.Errorf("list actions page %d: duplicate action ID (request-id: %s)", req.Page, resp.RequestID)
		}
		seen[action.ID] = struct{}{}
	}

	return nil
}

// CreateAction registers one action through the batch API and verifies the created ID.
func (h *Handler) CreateAction(ctx contextx.IContext, systemID string, action Action) error {
	resp, err := h.cli.batchCreateAction(ctx, &BatchCreateActionReq{SystemID: systemID, Actions: []Action{action}})
	if err != nil {
		return err
	}
	if len(resp.Data) != 1 || resp.Data[0] != action.ID {
		return fmt.Errorf("create action: response ID does not match; check remote state before rerunning (request-id: %s)",
			resp.RequestID)
	}

	return nil
}

// UpdateAction updates only the action name, preserving its immutable binding.
func (h *Handler) UpdateAction(ctx contextx.IContext, systemID, actionID, name string) error {
	return h.cli.updateAction(ctx, &UpdateActionReq{SystemID: systemID, ActionID: actionID, Name: name})
}
