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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// IHandlerSystem exposes system registration operations without HTTP details.
type IHandlerSystem interface {
	SystemExists(ctx contextx.IContext, systemID string) (bool, error)
	CreateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error
	UpdateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error
}

// SystemExists distinguishes an absent system from a failed query.
func (h *Handler) SystemExists(ctx contextx.IContext, systemID string) (bool, error) {
	_, err := h.cli.retrieveSystem(ctx, &RetrieveSystemReq{SystemID: systemID})
	if errors.Is(err, errSystemNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateSystem registers a system with the supplied fields.
func (h *Handler) CreateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error {
	if fields.Name == nil || fields.Clients == nil {
		return fmt.Errorf("name and clients are required to create system %s", systemID)
	}
	_, err := h.cli.createSystem(ctx, &CreateSystemReq{ID: systemID, SystemFields: fields})

	return err
}

// UpdateSystem sends only supplied mutable fields, never the immutable system ID.
func (h *Handler) UpdateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error {
	return h.cli.updateSystem(ctx, &UpdateSystemReq{SystemID: systemID, SystemFields: fields})
}
