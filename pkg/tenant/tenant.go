/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// Mode tenant mode.
type Mode string

const (
	// ModeSingle single mode.
	ModeSingle Mode = "single"

	// ModeMultiple multiple mode.
	ModeMultiple Mode = "multiple"
)

// Validate validate tenant mode.
func (mode Mode) Validate() error {
	switch mode {
	case ModeSingle, ModeMultiple:
		return nil
	default:
		return fmt.Errorf("failed to validate tenant mode, mode(%s)", mode)
	}
}

// tenantMode tenant mode.
// nolint: gochecknoglobals
var tenantMode = struct {
	mode Mode
	once sync.Once
}{
	mode: ModeSingle,
	once: sync.Once{},
}

// SetMode tenant mode only can be set once.
func SetMode(mode Mode) {
	tenantMode.once.Do(func() {
		tenantMode.mode = mode
	})
}

// GetMode get tenant mode.
func GetMode() Mode {
	return tenantMode.mode
}

const (
	// SingleModeTenantID tenant id for single mode.
	SingleModeTenantID = "default"

	// SystemTenantID tenant id for system in multiple mode.
	SystemTenantID = "system"
)

// ITenantIDProvider tenant id provider.
type ITenantIDProvider interface {
	ListTenantIDs(nCtx contextx.IContext) ([]string, error)
}

// ITenantIDStorage tenant id storage.
// Deprecated: use ITenantIDProvider instead.
type ITenantIDStorage interface {
	GetAllTenantIDs() []string
}

var _ ITenantIDStorage = &singleModeTenantIDStorage{}
var _ ITenantIDProvider = &singleModeTenantIDStorage{}

type singleModeTenantIDStorage struct {
}

// GetAllTenantIDs get all tenant ids.
func (stg *singleModeTenantIDStorage) GetAllTenantIDs() []string {
	return []string{
		SingleModeTenantID,
	}
}

// ListTenantIDs lists all tenant ids.
func (stg *singleModeTenantIDStorage) ListTenantIDs(_ contextx.IContext) ([]string, error) {
	return []string{
		SingleModeTenantID,
	}, nil
}

// nolint: gochecknoglobals
var tenantStorage = struct {
	provider ITenantIDProvider
	sync.Once
}{
	provider: new(singleModeTenantIDStorage),
}

// SetTenantIDProvider sets the tenant id provider only once.
func SetTenantIDProvider(provider ITenantIDProvider) error {
	if provider == nil {
		return errors.New("tenant id provider is nil")
	}

	set := false
	tenantStorage.Once.Do(func() {
		tenantStorage.provider = provider
		set = true
	})
	if !set {
		return errors.New("tenant id provider is already set")
	}

	return nil
}

// ListTenantIDs lists all tenant ids.
func ListTenantIDs(nCtx contextx.IContext) ([]string, error) {
	return tenantStorage.provider.ListTenantIDs(nCtx)
}

// GetAllTenantIDs get all tenant ids.
// Deprecated: use ListTenantIDs instead.
func GetAllTenantIDs() []string {
	tenantIDs, err := ListTenantIDs(contextx.Background())
	if err != nil {
		return nil
	}

	return tenantIDs
}
