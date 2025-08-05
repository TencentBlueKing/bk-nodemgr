/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager ...
package manager

import (
	"sync"

	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
)

// EventDispatcher dispatches events.
type EventDispatcher interface {
	// Dispatch dispatches the event.
	Dispatch(eventType protoRelay.ServerPushEventType, payload []byte)

	// RegisterHandler registers the handler for the event.
	RegisterHandler(eventType protoRelay.ServerPushEventType, handler HandlerFunc)
}

// HandlerFunc is a function that handles the event.
type HandlerFunc func([]byte)

type defaultEventDispatcher struct {
	handlers map[protoRelay.ServerPushEventType]HandlerFunc
	mux      sync.RWMutex
}

// NewDefaultEventDispatcher returns a default event dispatcher.
func NewDefaultEventDispatcher() EventDispatcher {
	return &defaultEventDispatcher{
		handlers: make(map[protoRelay.ServerPushEventType]HandlerFunc),
	}
}

// Dispatch dispatches the event.
func (d *defaultEventDispatcher) Dispatch(eventType protoRelay.ServerPushEventType, payload []byte) {
	d.mux.RLock()
	handler, exists := d.handlers[eventType]
	d.mux.RUnlock()

	if exists && handler != nil {
		go handler(payload)
	}
}

// RegisterHandler registers the handler for the event.
func (d *defaultEventDispatcher) RegisterHandler(eventType protoRelay.ServerPushEventType, handler HandlerFunc) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.handlers[eventType] = handler
}
