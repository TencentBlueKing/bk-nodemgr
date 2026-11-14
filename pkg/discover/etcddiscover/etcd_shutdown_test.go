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

package etcddiscover

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type shutdownTestLease struct {
	clientv3.Lease
	nextID atomic.Int64
	revoke func(context.Context, clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)
}

func (lease *shutdownTestLease) Grant(ctx context.Context, _ int64) (*clientv3.LeaseGrantResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &clientv3.LeaseGrantResponse{ID: clientv3.LeaseID(lease.nextID.Add(1))}, nil
}

func (lease *shutdownTestLease) KeepAlive(ctx context.Context, _ clientv3.LeaseID) (
	<-chan *clientv3.LeaseKeepAliveResponse, error,
) {
	responses := make(chan *clientv3.LeaseKeepAliveResponse)
	go func() {
		<-ctx.Done()
		close(responses)
	}()

	return responses, nil
}

func (lease *shutdownTestLease) Revoke(ctx context.Context, id clientv3.LeaseID) (
	*clientv3.LeaseRevokeResponse, error,
) {
	return lease.revoke(ctx, id)
}

func (lease *shutdownTestLease) Close() error {
	return nil
}

type shutdownTestKV struct {
	clientv3.KV
	put func(context.Context) (*clientv3.PutResponse, error)
}

func (kv *shutdownTestKV) Put(ctx context.Context, _, _ string, _ ...clientv3.OpOption) (
	*clientv3.PutResponse, error,
) {
	return kv.put(ctx)
}

func shutdownTestProvider(t *testing.T, lease *shutdownTestLease) *ProviderEtcd {
	t.Helper()

	// A real connection object exercises Client.Close; all RPCs used by these
	// tests are replaced through the SDK's existing Lease and KV interfaces.
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:1"}})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	cli.Lease = lease
	cli.KV = &shutdownTestKV{put: func(context.Context) (*clientv3.PutResponse, error) {
		return &clientv3.PutResponse{}, nil
	}}
	t.Cleanup(func() { _ = cli.Close() })

	provider := NewProviderEtcd(&config.Etcd{})
	provider.etcdClient = cli
	provider.ctx, provider.cancel = context.WithCancel(context.Background())
	t.Cleanup(provider.cancel)

	return provider
}

func TestProviderEtcdShutdownBeforeStart(t *testing.T) {
	provider := NewProviderEtcd(&config.Etcd{})
	if err := provider.GracefulShutdown(); err != nil {
		if err.Error() != discover.ErrDiscoverNotStarted().Error() {
			t.Fatalf("unexpected error: %v", err)
		}

		return
	}

	t.Fatal("shutdown before Start must fail")
}

func TestProviderEtcdShutdownAggregatesRevokeErrors(t *testing.T) {
	firstErr := errors.New("first revoke failed")
	secondErr := errors.New("second revoke failed")
	var revoked atomic.Int64
	provider := shutdownTestProvider(t, &shutdownTestLease{
		revoke: func(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
			revoked.Add(1)
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("cleanup context already canceled: %w", err)
			}
			if _, ok := ctx.Deadline(); !ok {
				return nil, errors.New("cleanup context has no deadline")
			}
			if id == 1 {
				return nil, firstErr
			}

			return nil, secondErr
		},
	})
	first := registrationTestInstance()
	second := first
	second.ID = "second"
	if err := provider.Register(discover.ServiceNameBackend, first, second); err != nil {
		t.Fatal(err)
	}

	err := provider.GracefulShutdown()
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("shutdown did not preserve both revoke errors: %v", err)
	}
	if revoked.Load() != 2 || len(provider.takeAllRegistrations()) != 0 {
		t.Fatal("shutdown did not attempt every registration")
	}
	if !errors.Is(provider.etcdClient.Ctx().Err(), context.Canceled) {
		t.Fatal("client was not closed after revoke failures")
	}
}

func TestProviderEtcdShutdownIgnoresExpiredLease(t *testing.T) {
	provider := shutdownTestProvider(t, &shutdownTestLease{
		revoke: func(context.Context, clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
			return nil, rpctypes.ErrLeaseNotFound
		},
	})
	if err := provider.Register(discover.ServiceNameBackend, registrationTestInstance()); err != nil {
		t.Fatal(err)
	}
	if err := provider.GracefulShutdown(); err != nil {
		t.Fatalf("an already expired lease should not fail shutdown: %v", err)
	}
}

func TestProviderEtcdShutdownWaitsForInFlightRegister(t *testing.T) {
	var revoked atomic.Int64
	provider := shutdownTestProvider(t, &shutdownTestLease{
		revoke: func(context.Context, clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
			revoked.Add(1)

			return &clientv3.LeaseRevokeResponse{}, nil
		},
	})
	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	provider.etcdClient.KV = &shutdownTestKV{put: func(ctx context.Context) (*clientv3.PutResponse, error) {
		close(entered)
		<-ctx.Done()
		<-release

		// A successful RPC may finish concurrently with cancellation.
		return &clientv3.PutResponse{}, nil
	}}
	registered := make(chan error, 1)
	go func() {
		registered <- provider.Register(discover.ServiceNameBackend, registrationTestInstance())
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("registration did not enter Put")
	}

	stopped := make(chan error, 1)
	go func() { stopped <- provider.GracefulShutdown() }()
	select {
	case <-provider.ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not cancel in-flight operations")
	}
	select {
	case err := <-stopped:
		t.Fatalf("shutdown returned before the in-flight registration finished: %v", err)
	default:
	}

	instance := registrationTestInstance()
	for name, operation := range map[string]func() error{
		"register":   func() error { return provider.Register(discover.ServiceNameBackend, instance) },
		"update":     func() error { return provider.Update(discover.ServiceNameBackend, instance) },
		"deregister": func() error { return provider.Deregister(discover.ServiceNameBackend, instance.ID) },
		"shutdown":   provider.GracefulShutdown,
	} {
		if err := operation(); !errors.Is(err, errProviderShuttingDown) {
			t.Fatalf("%s during shutdown returned %v", name, err)
		}
	}
	// Unblock Put without closing the channel twice in failure cleanup.
	release <- struct{}{}
	if err := <-registered; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not finish after registration")
	}
	if revoked.Load() != 1 || len(provider.takeAllRegistrations()) != 0 {
		t.Fatal("in-flight registration escaped shutdown cleanup")
	}
}

func TestProviderEtcdShutdownPreservesCloseError(t *testing.T) {
	provider := NewProviderEtcd(&config.Etcd{})
	provider.etcdClient = clientv3.NewCtxClient(context.Background())
	provider.ctx, provider.cancel = context.WithCancel(context.Background())
	t.Cleanup(provider.cancel)

	// A context-only client returns its canceled context from Close.
	if err := provider.GracefulShutdown(); !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown did not propagate the client close error: %v", err)
	}
}
