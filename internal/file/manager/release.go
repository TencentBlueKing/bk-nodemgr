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

package manager

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IReleaseAgent defines agent release metadata operations.
type IReleaseAgent interface {
	// ListReleaseAgent lists agent records by page and conditions.
	ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error)

	// CountReleaseAgent counts agent records by conditions.
	CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseAgent gets an agent release by identity.
	GetReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseAgent, error)

	// DistinctReleaseAgent gets distinct agent fields.
	DistinctReleaseAgent(nCtx contextx.IContext, field types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// EnableReleaseAgent enables an agent release and records the operation.
	EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DisableReleaseAgent disables an agent release and records the operation.
	DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetAsDefaultReleaseAgent sets an agent release as default and records the operation.
	SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// CancelAsDefaultReleaseAgent clears an agent release default and records the operation.
	CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// DeleteReleaseAgent deletes agent release metadata and records the operation.
	DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error

	// SetReleaseAgentLabelsMany updates labels of matching agent releases.
	SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error
}

// IReleaseProxy defines proxy release metadata operations.
type IReleaseProxy interface {
	// ListReleaseProxy lists proxy records by page and conditions.
	ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error)

	// CountReleaseProxy counts proxy records by conditions.
	CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleaseProxy gets a proxy release by identity.
	GetReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseProxy, error)

	// DistinctReleaseProxy gets distinct proxy fields.
	DistinctReleaseProxy(nCtx contextx.IContext, field types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// EnableReleaseProxy enables a proxy release and records the operation.
	EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DisableReleaseProxy disables a proxy release and records the operation.
	DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetAsDefaultReleaseProxy sets a proxy release as default and records the operation.
	SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// CancelAsDefaultReleaseProxy clears a proxy release default and records the operation.
	CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// DeleteReleaseProxy deletes proxy release metadata and records the operation.
	DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error

	// SetReleaseProxyLabelsMany updates labels of matching proxy releases.
	SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error
}

// IReleasePlugin defines plugin release metadata operations.
// nolint:interfacebloat // Plugin release queries and mutations share one domain contract.
type IReleasePlugin interface {
	// ListReleasePlugin lists plugin records by page and conditions.
	ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error)

	// CountReleasePlugin counts plugin records by conditions.
	CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error)

	// GetReleasePlugin gets a plugin release by identity.
	GetReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleasePlugin, error)

	// DistinctReleasePlugin gets distinct plugin fields.
	DistinctReleasePlugin(nCtx contextx.IContext, field types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
		*types.ReleaseDistinctResult, error)

	// GetReleasePluginDefaultVersion gets the default plugin version.
	GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error)

	// DistinctNameReleasePlugin gets distinct plugin release names.
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// EnableReleasePlugin enables a plugin release and records the operation.
	EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DisableReleasePlugin disables a plugin release and records the operation.
	DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetAsDefaultReleasePlugin sets a plugin release as default and records the operation.
	SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelAsDefaultReleasePlugin clears a plugin release default and records the operation.
	CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// DeleteReleasePlugin deletes plugin release metadata and records the operation.
	DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// SetHiddenReleasePlugin hides a plugin release and records the operation.
	SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error

	// CancelHiddenReleasePlugin unhides a plugin release and records the operation.
	CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error
}

// IReleaseCert defines cert release metadata operations.
type IReleaseCert interface {
	// ListReleaseCert lists cert records by page and conditions.
	ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error)

	// DeleteReleaseCert deletes cert release metadata and records the operation.
	DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error
}

// IReleaseBinTool defines bintool release metadata operations.
type IReleaseBinTool interface {
	// ListReleaseBinTool lists bintool records by page and conditions.
	ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error)

	// DeleteReleaseBinTool deletes bintool release metadata and records the operation.
	DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error
}

// IReleasePluginBinTool defines plugin bintool release metadata operations.
type IReleasePluginBinTool interface {
	// ListReleasePluginBinTool lists plugin bintool records by page and conditions.
	ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.ReleasePluginBinTool, int64, error)

	// DistinctNameReleasePluginBinTool gets distinct plugin bintool release names.
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)

	// DeleteReleasePluginBinTool deletes plugin bintool release metadata and records the operation.
	DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error
}

// ListReleaseAgent lists agent releases by page and conditions.
func (m *Manager) ListReleaseAgent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error) {

	return m.storageRelease.ListReleaseAgent(nCtx, page, conditions...)
}

// CountReleaseAgent counts agent releases by conditions.
func (m *Manager) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return m.storageRelease.CountReleaseAgent(nCtx, conditions...)
}

// GetReleaseAgent gets an agent release by identity.
func (m *Manager) GetReleaseAgent(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseAgent, error) {

	return m.storageRelease.GetReleaseAgent(nCtx, gen, plat, version)
}

// DistinctReleaseAgent gets distinct agent release fields.
func (m *Manager) DistinctReleaseAgent(nCtx contextx.IContext, field types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	return m.storageRelease.DistinctReleaseAgent(nCtx, field, conditions...)
}

// EnableReleaseAgent enables an agent release and records the operation.
func (m *Manager) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := m.storageRelease.EnableReleaseAgent(nCtx, key); err != nil {
		return err
	}
	m.recordAgentReleaseEvent(nCtx, key, types.PackageEventTypeEnable)

	return nil
}

// DisableReleaseAgent disables an agent release and records the operation.
func (m *Manager) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := m.storageRelease.DisableReleaseAgent(nCtx, key); err != nil {
		return err
	}
	m.recordAgentReleaseEvent(nCtx, key, types.PackageEventTypeDisable)

	return nil
}

// SetAsDefaultReleaseAgent sets an agent release as default and records the operation.
func (m *Manager) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := m.storageRelease.SetAsDefaultReleaseAgent(nCtx, key); err != nil {
		return err
	}
	m.recordAgentReleaseEvent(nCtx, key, types.PackageEventTypeSetAsDefault)

	return nil
}

// CancelAsDefaultReleaseAgent clears an agent release default and records the operation.
func (m *Manager) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := m.storageRelease.CancelAsDefaultReleaseAgent(nCtx, key); err != nil {
		return err
	}
	m.recordAgentReleaseEvent(nCtx, key, types.PackageEventTypeCancelAsDefault)

	return nil
}

// DeleteReleaseAgent deletes agent release metadata and records the operation.
func (m *Manager) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := m.storageRelease.DeleteReleaseAgent(nCtx, key); err != nil {
		return err
	}
	m.recordAgentReleaseEvent(nCtx, key, types.PackageEventTypeDelete)

	return nil
}

// SetReleaseAgentLabelsMany updates labels of matching agent releases.
func (m *Manager) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return m.storageRelease.SetReleaseAgentLabelsMany(nCtx, labels, conditions...)
}

// ListReleaseProxy lists proxy records by page and conditions.
func (m *Manager) ListReleaseProxy(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error) {

	return m.storageRelease.ListReleaseProxy(nCtx, page, conditions...)
}

// CountReleaseProxy counts proxy records by conditions.
func (m *Manager) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return m.storageRelease.CountReleaseProxy(nCtx, conditions...)
}

// GetReleaseProxy gets a proxy release by identity.
func (m *Manager) GetReleaseProxy(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseProxy, error) {

	return m.storageRelease.GetReleaseProxy(nCtx, gen, plat, version)
}

// DistinctReleaseProxy gets distinct proxy fields.
func (m *Manager) DistinctReleaseProxy(nCtx contextx.IContext, field types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	return m.storageRelease.DistinctReleaseProxy(nCtx, field, conditions...)
}

// EnableReleaseProxy enables a proxy release and records the operation.
func (m *Manager) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := m.storageRelease.EnableReleaseProxy(nCtx, key); err != nil {
		return err
	}
	m.recordProxyReleaseEvent(nCtx, key, types.PackageEventTypeEnable)

	return nil
}

// DisableReleaseProxy disables a proxy release and records the operation.
func (m *Manager) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := m.storageRelease.DisableReleaseProxy(nCtx, key); err != nil {
		return err
	}
	m.recordProxyReleaseEvent(nCtx, key, types.PackageEventTypeDisable)

	return nil
}

// SetAsDefaultReleaseProxy sets a proxy release as default and records the operation.
func (m *Manager) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := m.storageRelease.SetAsDefaultReleaseProxy(nCtx, key); err != nil {
		return err
	}
	m.recordProxyReleaseEvent(nCtx, key, types.PackageEventTypeSetAsDefault)

	return nil
}

// CancelAsDefaultReleaseProxy clears a proxy release default and records the operation.
func (m *Manager) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := m.storageRelease.CancelAsDefaultReleaseProxy(nCtx, key); err != nil {
		return err
	}
	m.recordProxyReleaseEvent(nCtx, key, types.PackageEventTypeCancelAsDefault)

	return nil
}

// DeleteReleaseProxy deletes proxy release metadata and records the operation.
func (m *Manager) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := m.storageRelease.DeleteReleaseProxy(nCtx, key); err != nil {
		return err
	}
	m.recordProxyReleaseEvent(nCtx, key, types.PackageEventTypeDelete)

	return nil
}

// SetReleaseProxyLabelsMany updates labels of matching proxy releases.
func (m *Manager) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return m.storageRelease.SetReleaseProxyLabelsMany(nCtx, labels, conditions...)
}

// ListReleasePlugin lists plugin records by page and conditions.
func (m *Manager) ListReleasePlugin(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error) {

	return m.storageRelease.ListReleasePlugin(nCtx, page, conditions...)
}

// CountReleasePlugin counts plugin records by conditions.
func (m *Manager) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return m.storageRelease.CountReleasePlugin(nCtx, conditions...)
}

// GetReleasePlugin gets a plugin release by identity.
func (m *Manager) GetReleasePlugin(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleasePlugin, error) {

	return m.storageRelease.GetReleasePlugin(nCtx, name, gen, plat, version)
}

// DistinctReleasePlugin gets distinct plugin fields.
func (m *Manager) DistinctReleasePlugin(nCtx contextx.IContext, field types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	return m.storageRelease.DistinctReleasePlugin(nCtx, field, conditions...)
}

// GetReleasePluginDefaultVersion gets the default plugin version.
func (m *Manager) GetReleasePluginDefaultVersion(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error) {

	return m.storageRelease.GetReleasePluginDefaultVersion(nCtx, name, gen, plat)
}

// DistinctNameReleasePlugin gets distinct plugin release names.
func (m *Manager) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return m.storageRelease.DistinctNameReleasePlugin(nCtx, conditions...)
}

// EnableReleasePlugin enables a plugin release and records the operation.
func (m *Manager) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.EnableReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeEnable)

	return nil
}

// DisableReleasePlugin disables a plugin release and records the operation.
func (m *Manager) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.DisableReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeDisable)

	return nil
}

// SetAsDefaultReleasePlugin sets a plugin release as default and records the operation.
func (m *Manager) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.SetAsDefaultReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeSetAsDefault)

	return nil
}

// CancelAsDefaultReleasePlugin clears a plugin release default and records the operation.
func (m *Manager) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.CancelAsDefaultReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeCancelAsDefault)

	return nil
}

// DeleteReleasePlugin deletes plugin release metadata and records the operation.
func (m *Manager) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.DeleteReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeDelete)

	return nil
}

// SetHiddenReleasePlugin hides a plugin release and records the operation.
func (m *Manager) SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.SetHiddenReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeSetHidden)

	return nil
}

// CancelHiddenReleasePlugin unhides a plugin release and records the operation.
func (m *Manager) CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := m.storageRelease.CancelHiddenReleasePlugin(nCtx, key); err != nil {
		return err
	}
	m.recordPluginReleaseEvent(nCtx, key, types.PackageEventTypeCancelHidden)

	return nil
}

// ListReleaseCert lists cert records by page and conditions.
func (m *Manager) ListReleaseCert(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error) {

	return m.storageRelease.ListReleaseCert(nCtx, page, conditions...)
}

// DeleteReleaseCert deletes cert release metadata and records the operation.
func (m *Manager) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	if err := m.storageRelease.DeleteReleaseCert(nCtx, key); err != nil {
		return err
	}
	m.recordUnknownPlatformReleaseEvent(nCtx, types.ReleaseNameCert, types.ReleaseTypeCert, key.Generation)

	return nil
}

// ListReleaseBinTool lists bintool records by page and conditions.
func (m *Manager) ListReleaseBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error) {

	return m.storageRelease.ListReleaseBinTool(nCtx, page, conditions...)
}

// DeleteReleaseBinTool deletes bintool release metadata and records the operation.
func (m *Manager) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	if err := m.storageRelease.DeleteReleaseBinTool(nCtx, key); err != nil {
		return err
	}
	m.recordUnknownPlatformReleaseEvent(nCtx, types.ReleaseNameBinTool, types.ReleaseTypeBinTool, key.Generation)

	return nil
}

// ListReleasePluginBinTool lists plugin bintool records by page and conditions.
func (m *Manager) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {

	return m.storageRelease.ListReleasePluginBinTool(nCtx, page, conditions...)
}

// DistinctNameReleasePluginBinTool gets distinct plugin bintool release names.
func (m *Manager) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return m.storageRelease.DistinctNameReleasePluginBinTool(nCtx, conditions...)
}

// DeleteReleasePluginBinTool deletes plugin bintool release metadata and records the operation.
func (m *Manager) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	if err := m.storageRelease.DeleteReleasePluginBinTool(nCtx, key); err != nil {
		return err
	}
	m.recordUnknownPlatformReleaseEvent(nCtx, key.Name, types.ReleaseTypePluginBinTool, key.Generation)

	return nil
}

func (m *Manager) recordAgentReleaseEvent(nCtx contextx.IContext, key types.ReleaseAgentKey, eventType types.PackageEventType) {
	m.recordPackageEvents(nCtx, releaseEvent(nCtx, types.ReleaseNameAgent, types.ReleaseTypeAgent,
		key.Generation, key.Platform.OS, key.Platform.Arch, key.Version, eventType))
}

func (m *Manager) recordProxyReleaseEvent(nCtx contextx.IContext, key types.ReleaseProxyKey, eventType types.PackageEventType) {
	m.recordPackageEvents(nCtx, releaseEvent(nCtx, types.ReleaseNameProxy, types.ReleaseTypeProxy,
		key.Generation, key.Platform.OS, key.Platform.Arch, key.Version, eventType))
}

func (m *Manager) recordPluginReleaseEvent(nCtx contextx.IContext, key types.ReleasePluginKey, eventType types.PackageEventType) {
	m.recordPackageEvents(nCtx, releaseEvent(nCtx, key.Name, types.ReleaseTypePlugin,
		key.Generation, key.Platform.OS, key.Platform.Arch, key.Version, eventType))
}

func (m *Manager) recordUnknownPlatformReleaseEvent(
	nCtx contextx.IContext, name string, releaseType types.ReleaseType, generation types.Generation) {

	m.recordPackageEvents(nCtx, releaseEvent(nCtx, name, releaseType, generation,
		criteria.OSUnknown, criteria.CPUArchUnknown, "", types.PackageEventTypeDelete))
}

func releaseEvent(nCtx contextx.IContext, name string, releaseType types.ReleaseType, generation types.Generation, osType criteria.OSType,
	cpuArch criteria.CPUArch, version string, eventType types.PackageEventType) *types.PackageEvent {

	return &types.PackageEvent{
		Name:        name,
		EventType:   eventType,
		Generation:  generation,
		ReleaseType: releaseType,
		OSType:      osType,
		CPUArch:     cpuArch,
		Version:     version,
		OperateTime: time.Now(),
		Operator:    nCtx.LoginName(),
	}
}
