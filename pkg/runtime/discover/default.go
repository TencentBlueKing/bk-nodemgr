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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// ProviderDefault this defines the default service discovery provider.
type ProviderDefault struct {
	mutex      sync.RWMutex
	services   map[string]map[string]Instance
	watchChans map[string][]chan []Instance
	done       chan struct{}
	selector   Selector
}

// NewProviderDefault creates a new default provider.
func NewProviderDefault(selector Selector) *ProviderDefault {
	if selector == nil {
		selector = &RandomSelector{}
	}

	return &ProviderDefault{
		services:   make(map[string]map[string]Instance),
		watchChans: make(map[string][]chan []Instance),
		done:       make(chan struct{}),
		selector:   selector,
	}
}

// GetAllService get all specific service instances.
func (p *ProviderDefault) GetAllService(_ context.Context, serviceName string) ([]Instance, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	instanceMap, ok := p.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound()
	}

	instances := conv.MapToSlice(instanceMap)

	return instances, nil
}

// GetService get one with the selector.
func (p *ProviderDefault) GetService(ctx context.Context, serviceName string, selector Selector) (Instance, error) {
	instances, err := p.GetAllService(ctx, serviceName)
	if err != nil {
		return Instance{}, err
	}

	if selector == nil {
		selector = &RandomSelector{}
	}

	return selector.Select(instances)
}

// Watch watches the changes of a service.
// this will return a channel that can get the all latest instances of the service.
func (p *ProviderDefault) Watch(ctx context.Context, serviceName string) (<-chan []Instance, error) {
	p.mutex.Lock()

	ch := make(chan []Instance, 1)

	if _, exists := p.watchChans[serviceName]; !exists {
		p.watchChans[serviceName] = make([]chan []Instance, 0)
	}
	p.watchChans[serviceName] = append(p.watchChans[serviceName], ch)

	instances, err := p.getAllServiceNoLock(serviceName)
	p.mutex.Unlock()

	if err == nil && len(instances) > 0 {
		select {
		case ch <- instances:
		default:
		}
	}

	go func() {
		select {
		case <-ctx.Done():
			p.removeWatchChan(serviceName, ch)
		case <-p.done:
			close(ch)
		}
	}()

	return ch, nil
}

func (p *ProviderDefault) removeWatchChan(serviceName string, ch chan []Instance) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	chans, exists := p.watchChans[serviceName]
	if !exists {
		return
	}

	for i, watchCh := range chans {
		if watchCh == ch {
			p.watchChans[serviceName] = append(chans[:i], chans[i+1:]...)
			close(ch)

			break
		}
	}
}

// Register registers a service instance.
func (p *ProviderDefault) Register(_ context.Context, serviceName string, instances ...Instance) error {
	for _, instance := range instances {
		if err := instance.Validate(); err != nil {
			return err
		}
	}

	p.mutex.Lock()

	if _, exists := p.services[serviceName]; !exists {
		p.services[serviceName] = make(map[string]Instance)
	}

	for _, instance := range instances {
		p.services[serviceName][instance.ID] = instance
	}

	newInstances, _ := p.getAllServiceNoLock(serviceName)
	p.notifyWatchers(serviceName, newInstances)
	p.mutex.Unlock()

	return nil
}

func (p *ProviderDefault) getAllServiceNoLock(serviceName string) ([]Instance, error) {
	instanceMap, ok := p.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound()
	}

	instances := conv.MapToSlice(instanceMap)

	return instances, nil
}

// Deregister deregister a service instance.
func (p *ProviderDefault) Deregister(_ context.Context, serviceName string, instanceID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.services[serviceName]; !exists {
		return ErrNotRegistered()
	}

	delete(p.services[serviceName], instanceID)

	instances, _ := p.getAllServiceNoLock(serviceName)

	p.notifyWatchers(serviceName, instances)

	return nil
}

func (p *ProviderDefault) notifyWatchers(serviceName string, instances []Instance) {
	chans, exists := p.watchChans[serviceName]
	if !exists || len(chans) == 0 {
		return
	}

	for _, ch := range chans {
		select {
		case ch <- instances:
		default:
		}
	}
}

// Close closes the provider.
func (p *ProviderDefault) Close() error {
	close(p.done)

	return nil
}
