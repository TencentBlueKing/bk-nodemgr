/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package basestorage define the Storage basic interface.
package basestorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"go.mongodb.org/mongo-driver/mongo"
)

// Interface define the Storage basic interface.
type Interface interface {
	Start(nCtx contextx.IContext) error
	CheckHealthz() error
	Terminate() error
}

const (
	pingTimeoutDefault = 3 * time.Second
)

// Storage define the Storage basic interface.
type Storage struct {
	// Storage basic fields
	Name      string
	IsRunning bool
	Ctx       context.Context
	Cancel    context.CancelFunc

	Database  *mongo.Database
	Scheduler scheduler.Scheduler

	startFunc     func() error
	checkFunc     func() error
	terminateFunc func()
}

// OptionFunc ...
type OptionFunc func(s *Storage) error

// WithStartFunc ...
func WithStartFunc(startFunc func() error) OptionFunc {
	return func(s *Storage) error {
		s.startFunc = startFunc
		return nil
	}
}

// WithCheckFunc ...
func WithCheckFunc(checkFunc func() error) OptionFunc {
	return func(s *Storage) error {
		s.checkFunc = checkFunc
		return nil
	}
}

// WithTerminateFunc ...
func WithTerminateFunc(terminateFunc func()) OptionFunc {
	return func(s *Storage) error {
		s.terminateFunc = terminateFunc
		return nil
	}
}

// InitStorage initialize the storage.
func InitStorage(s *Storage, opts ...OptionFunc) error {
	if s.Database == nil {
		return errors.New("mongo client is nil")
	}

	s.IsRunning = false

	s.startFunc = func() error {
		return nil
	}

	s.checkFunc = func() error {
		return nil
	}

	s.terminateFunc = func() {
		return
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}

	return nil
}

// Start ...
func (s *Storage) Start(nCtx contextx.IContext) error {
	logger.G.Sys().With("name", s.Name).Info("start storage")

	if s.IsRunning {
		return errors.New("storage already started")
	}

	if s.Database == nil {
		return errors.New("mongo client is nil")
	}

	if s.startFunc == nil {
		return errors.New("start func is nil, need to init base storage")
	}

	if s.checkFunc == nil {
		return errors.New("check func is nil, need to init base storage")
	}

	if s.terminateFunc == nil {
		return errors.New("terminate func is nil, need to init base storage")
	}

	var err error
	s.Ctx, s.Cancel = context.WithCancel(nCtx)
	defer func() {
		if err != nil {
			s.Cancel()
		}
	}()

	s.IsRunning = true

	if s.Scheduler != nil {
		s.Scheduler.Start()
	}

	if err := s.Database.Client().Ping(nCtx, nil); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to ping mongo client")

		return err
	}

	if err := s.startFunc(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to start storage")

		return err
	}

	if err := s.checkFunc(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to check storage health")

		return err
	}

	go func() {
		select {
		case <-s.Ctx.Done():
			{
				if s.Scheduler != nil {
					s.Scheduler.Terminate()
				}

				s.terminateFunc()

				s.IsRunning = false

				logger.G.Sys().With("name", s.Name).Info("terminated storage")
			}
		}
	}()

	logger.G.Sys().With("name", s.Name).Info("started storage")

	return nil
}

// CheckHealthz ...
func (s *Storage) CheckHealthz() error {
	if s.Database == nil {
		return errors.New("mongo client not initialized")
	}

	if s.Ctx == nil {
		return errors.New("context not initialized, need to init base storage")
	}

	ctx, cancel := context.WithDeadline(s.Ctx, time.Now().Add(pingTimeoutDefault))
	defer cancel()

	if err := s.Database.Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping mongo client: %v", err)
	}

	if err := s.checkFunc(); err != nil {
		return err
	}

	logger.G.Sys().With("name", s.Name).Debug("successfully checked healthz of storage")

	return nil
}

// Terminate ...
func (s *Storage) Terminate() error {
	if !s.IsRunning {
		return errors.New("storage already terminated")
	}

	s.Cancel()

	return nil
}
