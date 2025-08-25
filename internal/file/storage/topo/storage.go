/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides the topo storage interface.
package topo

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the interface of topo storage.
type IStorage interface {
	basestorage.Interface

	// GetDirectNetworkAreaHostByAnyInnerIP get host by any inner ip, v4 or v6 in direct networkarea.
	GetDirectNetworkAreaHostByAnyInnerIP(ctx context.Context, ipv4, ipv6 string) (*types.Host, error)

	// GetHostByID get host by host-id.
	GetHostByID(ctx context.Context, hostID int64) (*types.Host, error)
}

const (
	// StorageName defines the storage name.
	StorageName = "topo"

	// globalNetworkAreaID defines the global network area id.
	globalNetworkAreaID = base.GlobalNetworkAreaID
)

// NewStorage creates a new release storage.
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoHost host.IHandler
}

func (s *Storage) initDao() error {
	s.daoHost = host.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoHost == nil {
		return errors.New("dao host is nil")
	}

	return nil
}

// GetDirectNetworkAreaHostByAnyInnerIP get host by any inner ip, v4 or v6 in direct networkarea.
func (s *Storage) GetDirectNetworkAreaHostByAnyInnerIP(ctx context.Context, ipv4, ipv6 string) (*types.Host, error) {
	if ipv4 != "" {
		hosts, _, err := s.daoHost.List(ctx,
			types.UnlimitedPage(),
			host.WithNetworkAreaID(globalNetworkAreaID),
			host.WithStaticInnerIP(ipv4))
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by ipv4(%s), networkarea(%d): %w",
				ipv4, globalNetworkAreaID, err)
		}

		if len(hosts) == 0 {
			return nil, fmt.Errorf("host not found. ipv4(%s), networkarea(%d)", ipv4, globalNetworkAreaID)
		}

		if len(hosts) > 1 {
			return nil, fmt.Errorf("host not unique. ipv4(%s), networkarea(%d)", ipv4, globalNetworkAreaID)
		}

		return hosts[0], nil
	}

	if ipv6 != "" {
		hosts, _, err := s.daoHost.List(ctx,
			types.UnlimitedPage(),
			host.WithNetworkAreaID(globalNetworkAreaID),
			host.WithStaticInnerIPV6(ipv6))
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by ipv6(%s), networkarea(%d): %w",
				ipv6, globalNetworkAreaID, err)
		}

		if len(hosts) == 0 {
			return nil, fmt.Errorf("host not found. ipv6(%s), networkarea(%d)", ipv6, globalNetworkAreaID)
		}

		if len(hosts) > 1 {
			return nil, fmt.Errorf("host not unique. ipv6(%s), networkarea(%d)", ipv6, globalNetworkAreaID)
		}

		return hosts[0], nil
	}

	return nil, errors.New("ipv4 and ipv6 are all empty")
}

// GetHostByID get host by host-id.
func (s *Storage) GetHostByID(ctx context.Context, hostID int64) (*types.Host, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if hostID < 0 {
		return nil, errors.New("host id should be equal or greater than 0")
	}

	hosts, count, err := s.daoHost.List(ctx, types.Page{Limit: 1}, host.WithHostID(hostID))
	if err != nil {
		return nil, fmt.Errorf("failed to get host by id, host-id(%d), err: %w", hostID, err)
	}

	if count != 1 {
		return nil, fmt.Errorf("failed to get host by id, result count is not 1, host-id(%d), count(%d)",
			hostID, count)
	}

	return hosts[0], nil
}
