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

package file

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPkgReleaseHandler defines release metadata and state operations.
type IPkgReleaseHandler interface {
	IReleaseAgentHandler
	IReleaseProxyHandler
	IReleasePluginHandler
	IReleaseCertHandler
	IReleaseBinToolHandler
	IReleasePluginBinToolHandler
}

// IReleaseAgentHandler defines agent release operations.
type IReleaseAgentHandler interface {
	// ListReleaseAgent queries agent release metadata.
	ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent queries agent release metadata.
	CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseAgent queries agent release metadata.
	GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error)

	// DistinctReleaseAgent queries agent release metadata.
	DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// EnableReleaseAgent enables a release agent package.
	EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables a release agent package.
	DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets a release agent package as default.
	SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent cancels a release agent package default setting.
	CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DeleteReleaseAgent deletes a release agent package.
	DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetReleaseAgentLabelsMany replaces labels for matching release agent packages.
	SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error
}

// IReleaseProxyHandler defines proxy release operations.
type IReleaseProxyHandler interface {
	// ListReleaseProxy queries proxy release metadata.
	ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy queries proxy release metadata.
	CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseProxy queries proxy release metadata.
	GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error)

	// DistinctReleaseProxy queries proxy release metadata.
	DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// EnableReleaseProxy enables a release proxy package.
	EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables a release proxy package.
	DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultReleaseProxy sets a release proxy package as default.
	SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy cancels a release proxy package default setting.
	CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DeleteReleaseProxy deletes a release proxy package.
	DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetReleaseProxyLabelsMany replaces labels for matching release proxy packages.
	SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error
}

// IReleasePluginHandler defines plugin release operations.
// nolint: interfacebloat // Keep release queries and mutations together by package type.
type IReleasePluginHandler interface {
	// ListReleasePlugin queries plugin release metadata.
	ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin queries plugin release metadata.
	CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleasePlugin queries plugin release metadata.
	GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (*types.ReleasePlugin, error)

	// DistinctReleasePlugin queries plugin release metadata.
	DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// ExistReleasePlugin queries plugin release metadata.
	ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (bool, error)

	// GetReleasePluginDefaultVersion queries plugin release metadata.
	GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error)

	// DistinctNameReleasePlugin queries plugin release metadata.
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// EnableReleasePlugin enables a release plugin package.
	EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables a release plugin package.
	DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetAsDefaultReleasePlugin sets a release plugin package as default.
	SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin cancels a release plugin package default setting.
	CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DeleteReleasePlugin deletes a release plugin package.
	DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetHiddenReleasePlugin hides a release plugin package.
	SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelHiddenReleasePlugin unhides a release plugin package.
	CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error
}

// IReleaseCertHandler defines cert release operations.
type IReleaseCertHandler interface {
	// ListReleaseCert queries cert release metadata.
	ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error)

	// DeleteReleaseCert deletes a release cert package.
	DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error
}

// IReleaseBinToolHandler defines bintool release operations.
type IReleaseBinToolHandler interface {
	// ListReleaseBinTool queries bintool release metadata.
	ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error)

	// DeleteReleaseBinTool deletes a release bin tool package.
	DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error
}

// IReleasePluginBinToolHandler defines pluginbintool release operations.
type IReleasePluginBinToolHandler interface {
	// ListReleasePluginBinTool queries pluginbintool release metadata.
	ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.ReleasePluginBinTool, int64, error)

	// DistinctNameReleasePluginBinTool queries pluginbintool release metadata.
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// DeleteReleasePluginBinTool deletes a release plugin bin tool package.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error
}

// ===============================================================================
// ReleaseAgent Related Interface
// ===============================================================================

// ListReleaseAgent queries agent release metadata.
func (h *handler) ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseAgent, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseAgentListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleaseAgent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleaseAgent queries agent release metadata.
func (h *handler) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleaseAgentCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}

	data, err := h.cli.countReleaseAgent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	return data.ConvertToTypes(), nil
}

// GetReleaseAgent queries agent release metadata.
func (h *handler) GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseAgentGetReq)
	req.ConvertFromTypes(key)
	data, err := h.cli.getReleaseAgent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes()
}

// DistinctReleaseAgent queries agent release metadata.
func (h *handler) DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseAgentDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}

	data, err := h.cli.distinctReleaseAgent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes(), nil
}

// EnableReleaseAgent enables a release agent package.
func (h *handler) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentEnableReq)
	req.ConvertFromTypes(key)

	return h.cli.enableReleaseAgent(nCtx, nCtx.TenantID(), req)
}

// DisableReleaseAgent disables a release agent package.
func (h *handler) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentDisableReq)
	req.ConvertFromTypes(key)

	return h.cli.disableReleaseAgent(nCtx, nCtx.TenantID(), req)
}

// SetAsDefaultReleaseAgent sets a release agent package as default.
func (h *handler) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentSetAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.setAsDefaultReleaseAgent(nCtx, nCtx.TenantID(), req)
}

// CancelAsDefaultReleaseAgent cancels a release agent package default setting.
func (h *handler) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentCancelAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.cancelAsDefaultReleaseAgent(nCtx, nCtx.TenantID(), req)
}

// DeleteReleaseAgent deletes a release agent package.
func (h *handler) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleaseAgent(nCtx, nCtx.TenantID(), req)
}

// SetReleaseAgentLabelsMany replaces labels for matching release agent packages.
func (h *handler) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseAgentSetLabelsManyReq)
	if err := req.ConvertFromTypes(labels, conditions...); err != nil {
		return err
	}

	return h.cli.setReleaseAgentLabelsMany(nCtx, nCtx.TenantID(), req)
}

// ===============================================================================
// ReleaseProxy Related Interface
// ===============================================================================

// ListReleaseProxy queries proxy release metadata.
func (h *handler) ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseProxy, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseProxyListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleaseProxy(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleaseProxy queries proxy release metadata.
func (h *handler) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleaseProxyCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}

	data, err := h.cli.countReleaseProxy(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	return data.ConvertToTypes(), nil
}

// GetReleaseProxy queries proxy release metadata.
func (h *handler) GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseProxyGetReq)
	req.ConvertFromTypes(key)
	data, err := h.cli.getReleaseProxy(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes()
}

// DistinctReleaseProxy queries proxy release metadata.
func (h *handler) DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleaseProxyDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}

	data, err := h.cli.distinctReleaseProxy(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes(), nil
}

// EnableReleaseProxy enables a release proxy package.
func (h *handler) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxyEnableReq)
	req.ConvertFromTypes(key)

	return h.cli.enableReleaseProxy(nCtx, nCtx.TenantID(), req)
}

// DisableReleaseProxy disables a release proxy package.
func (h *handler) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxyDisableReq)
	req.ConvertFromTypes(key)

	return h.cli.disableReleaseProxy(nCtx, nCtx.TenantID(), req)
}

// SetAsDefaultReleaseProxy sets a release proxy package as default.
func (h *handler) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxySetAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.setAsDefaultReleaseProxy(nCtx, nCtx.TenantID(), req)
}

// CancelAsDefaultReleaseProxy cancels a release proxy package default setting.
func (h *handler) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxyCancelAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.cancelAsDefaultReleaseProxy(nCtx, nCtx.TenantID(), req)
}

// DeleteReleaseProxy deletes a release proxy package.
func (h *handler) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxyDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleaseProxy(nCtx, nCtx.TenantID(), req)
}

// SetReleaseProxyLabelsMany replaces labels for matching release proxy packages.
func (h *handler) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseProxySetLabelsManyReq)
	if err := req.ConvertFromTypes(labels, conditions...); err != nil {
		return err
	}

	return h.cli.setReleaseProxyLabelsMany(nCtx, nCtx.TenantID(), req)
}

// ===============================================================================
// ReleasePlugin Related Interface
// ===============================================================================

// ListReleasePlugin queries plugin release metadata.
func (h *handler) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleasePluginListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountReleasePlugin queries plugin release metadata.
func (h *handler) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	req := new(protoFile.ReleasePluginCountReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return 0, err
	}

	data, err := h.cli.countReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	return data.ConvertToTypes(), nil
}

// GetReleasePlugin queries plugin release metadata.
func (h *handler) GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (*types.ReleasePlugin, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginGetReq)
	req.ConvertFromTypes(key)
	data, err := h.cli.getReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes()
}

// DistinctReleasePlugin queries plugin release metadata.
func (h *handler) DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginDistinctReq)
	if err := req.ConvertFromTypes(fields, conditions...); err != nil {
		return nil, err
	}

	data, err := h.cli.distinctReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes(), nil
}

// ExistReleasePlugin queries plugin release metadata.
func (h *handler) ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, err
	}

	req := new(protoFile.ReleasePluginExistReq)
	req.ConvertFromTypes(key)
	data, err := h.cli.existReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return false, err
	}

	return data.ConvertToTypes(), nil
}

// GetReleasePluginDefaultVersion queries plugin release metadata.
func (h *handler) GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return "", err
	}

	req := new(protoFile.ReleasePluginGetDefaultVersionReq)
	req.ConvertFromTypes(name, gen, plat)
	data, err := h.cli.getReleasePluginDefaultVersion(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return "", err
	}

	return data.ConvertToTypes(), nil
}

// DistinctNameReleasePlugin queries plugin release metadata.
func (h *handler) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginDistinctNameReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return nil, err
	}

	data, err := h.cli.distinctNameReleasePlugin(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes(), nil
}

// EnableReleasePlugin enables a release plugin package.
func (h *handler) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginEnableReq)
	req.ConvertFromTypes(key)

	return h.cli.enableReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// DisableReleasePlugin disables a release plugin package.
func (h *handler) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginDisableReq)
	req.ConvertFromTypes(key)

	return h.cli.disableReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// SetAsDefaultReleasePlugin sets a release plugin package as default.
func (h *handler) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginSetAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.setAsDefaultReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// CancelAsDefaultReleasePlugin cancels a release plugin package default setting.
func (h *handler) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginCancelAsDefaultReq)
	req.ConvertFromTypes(key)

	return h.cli.cancelAsDefaultReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// DeleteReleasePlugin deletes a release plugin package.
func (h *handler) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// SetHiddenReleasePlugin hides a release plugin package.
func (h *handler) SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginSetHiddenReq)
	req.ConvertFromTypes(key)

	return h.cli.setHiddenReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// CancelHiddenReleasePlugin unhides a release plugin package.
func (h *handler) CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginCancelHiddenReq)
	req.ConvertFromTypes(key)

	return h.cli.cancelHiddenReleasePlugin(nCtx, nCtx.TenantID(), req)
}

// ===============================================================================
// ReleaseCert Related Interface
// ===============================================================================

// ListReleaseCert queries cert release metadata.
func (h *handler) ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseCert, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseCertListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleaseCert(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// DeleteReleaseCert deletes a release cert package.
func (h *handler) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseCertDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleaseCert(nCtx, nCtx.TenantID(), req)
}

// ===============================================================================
// ReleaseBinTool Related Interface
// ===============================================================================

// ListReleaseBinTool queries bintool release metadata.
func (h *handler) ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseBinTool, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleaseBinToolListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleaseBinTool(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// DeleteReleaseBinTool deletes a release bin tool package.
func (h *handler) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleaseBinToolDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleaseBinTool(nCtx, nCtx.TenantID(), req)
}

// ===============================================================================
// ReleasePluginBinTool Related Interface
// ===============================================================================

// ListReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePluginBinTool, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	req := new(protoFile.ReleasePluginBinToolListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}

	data, err := h.cli.listReleasePluginBinTool(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}

	total, items, err := data.ConvertToTypes()
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// DistinctNameReleasePluginBinTool queries plugin_bintool release metadata.
func (h *handler) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	req := new(protoFile.ReleasePluginBinToolDistinctNameReq)
	if err := req.ConvertFromTypes(conditions...); err != nil {
		return nil, err
	}

	data, err := h.cli.distinctNameReleasePluginBinTool(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return data.ConvertToTypes(), nil
}

// DeleteReleasePluginBinTool deletes a release plugin bin tool package.
func (h *handler) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	req := new(protoFile.ReleasePluginBinToolDeleteReq)
	req.ConvertFromTypes(key)

	return h.cli.deleteReleasePluginBinTool(nCtx, nCtx.TenantID(), req)
}
