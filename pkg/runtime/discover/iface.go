/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Provider this defines the interface of a complete service discovery provider.
type Provider interface {
	Discover
	Watcher
	Registry
	Close()
}

// NewInstance creates a new instance.
func NewInstance(name, address string, meta map[string]string) Instance {
	inst := Instance{
		ID:      fmt.Sprintf("%s-%s-%s", name, address, uuid.NewString()),
		Name:    name,
		Address: address,
		Meta:    meta,
	}

	return inst
}

// Instance this defines a service instance.
type Instance struct {
	ID      string
	Name    string
	Address string
	Meta    map[string]string
}

// Validate validates the instance.
func (instance *Instance) Validate() error {
	if instance.ID == "" {
		return ErrInvalidInstanceID()
	}
	if instance.Name == "" {
		return ErrInvalidInstanceName()
	}
	if instance.Address == "" {
		return ErrInvalidInstanceAddress()
	}

	return nil
}

// Discover this defines the interface of service discovery.
type Discover interface {
	// GetAllService get all specific service instances.
	GetAllService(ctx context.Context, serviceName string) ([]Instance, error)

	// GetService get one with the selector.
	GetService(ctx context.Context, serviceName string, selector Selector) (Instance, error)
}

// Registry this defines the interface of service registry.
type Registry interface {
	// Register registers a service instance.
	Register(ctx context.Context, serviceName string, instance Instance) error

	// Deregister deregisters a service instance.
	Deregister(ctx context.Context, serviceName string, instanceID string) error
}

// Watcher this defines the interface of service watcher.
type Watcher interface {
	// Watch watches the changes of a service.
	// this will return a channel that can get the all latest instances of the service.
	Watch(ctx context.Context, serviceName string) (<-chan []Instance, error)
}

// Selector this defines the interface of a service instance selector.
type Selector interface {
	// Select selects a service instance from the given instances.
	Select(instances []Instance) (Instance, error)
}
