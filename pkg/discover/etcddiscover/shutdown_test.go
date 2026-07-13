/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package etcddiscover

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type fakeEtcdClient struct {
	mutex sync.Mutex

	deleteResponses map[string]*clientv3.DeleteResponse
	deleteErrors    map[string]error
	revokeErrors    map[clientv3.LeaseID]error

	deletedKeys        []string
	deleteOptionCounts []int
	revokedLeaseIDs    []clientv3.LeaseID
	closed             bool

	grantStarted chan struct{}
	grantOnce    sync.Once
}

func (client *fakeEtcdClient) Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return new(clientv3.GetResponse), nil
}

func (client *fakeEtcdClient) Put(ctx context.Context, _, _ string, _ ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return new(clientv3.PutResponse), ctx.Err()
}

func (client *fakeEtcdClient) Delete(_ context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.deletedKeys = append(client.deletedKeys, key)
	client.deleteOptionCounts = append(client.deleteOptionCounts, len(opts))
	if err := client.deleteErrors[key]; err != nil {
		return nil, err
	}
	if resp := client.deleteResponses[key]; resp != nil {
		return resp, nil
	}

	return new(clientv3.DeleteResponse), nil
}

func (client *fakeEtcdClient) Watch(context.Context, string, ...clientv3.OpOption) clientv3.WatchChan {
	return nil
}

func (client *fakeEtcdClient) Grant(ctx context.Context, _ int64) (*clientv3.LeaseGrantResponse, error) {
	client.grantOnce.Do(func() {
		if client.grantStarted != nil {
			close(client.grantStarted)
		}
	})
	<-ctx.Done()

	return nil, ctx.Err()
}

func (client *fakeEtcdClient) Revoke(_ context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.revokedLeaseIDs = append(client.revokedLeaseIDs, id)
	if err := client.revokeErrors[id]; err != nil {
		return nil, err
	}

	return new(clientv3.LeaseRevokeResponse), nil
}

func (client *fakeEtcdClient) KeepAlive(context.Context, clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ch := make(chan *clientv3.LeaseKeepAliveResponse)
	close(ch)

	return ch, nil
}

func (client *fakeEtcdClient) Close() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.closed = true

	return nil
}

func (client *fakeEtcdClient) isClosed() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	return client.closed
}

func (client *fakeEtcdClient) calls() ([]string, []int, []clientv3.LeaseID) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	return append([]string(nil), client.deletedKeys...),
		append([]int(nil), client.deleteOptionCounts...),
		append([]clientv3.LeaseID(nil), client.revokedLeaseIDs...)
}

func newShutdownTestProvider(client etcdClient) *ProviderEtcd {
	provider := NewProviderEtcd(nil)
	provider.etcdClient = client
	provider.ctx, provider.cancel = context.WithCancel(context.Background())
	provider.keeperCtx, provider.keeperCancel = context.WithCancel(provider.ctx)

	return provider
}

func TestProviderEtcdGracefulShutdownRejectsRegistryOperations(t *testing.T) {
	client := &fakeEtcdClient{}
	provider := newShutdownTestProvider(client)

	provider.startRegisterKeeper(discover.ServiceNameBackend, "missing")
	if err := provider.gracefulShutdown(time.Second, time.Second); err != nil {
		t.Fatalf("gracefulShutdown() error = %v", err)
	}

	instance := discover.Instance{ID: "instance", Name: "instance"}
	operations := []struct {
		name string
		call func() error
	}{
		{name: "register", call: func() error { return provider.Register(discover.ServiceNameBackend, instance) }},
		{name: "update", call: func() error { return provider.Update(discover.ServiceNameBackend, instance) }},
		{name: "deregister", call: func() error { return provider.Deregister(discover.ServiceNameBackend, instance.ID) }},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.call(); !errors.Is(err, errProviderShuttingDown) {
				t.Fatalf("operation error = %v, want %v", err, errProviderShuttingDown)
			}
		})
	}
}

func TestProviderEtcdGracefulShutdownCancelsBlockedOperation(t *testing.T) {
	client := &fakeEtcdClient{grantStarted: make(chan struct{})}
	provider := newShutdownTestProvider(client)
	registerDone := make(chan error, 1)
	go func() {
		registerDone <- provider.Register(discover.ServiceNameBackend, discover.Instance{ID: "instance", Name: "instance"})
	}()

	<-client.grantStarted
	err := provider.gracefulShutdown(time.Millisecond, time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("gracefulShutdown() error = %v, want deadline exceeded", err)
	}
	if registerErr := <-registerDone; !errors.Is(registerErr, context.Canceled) {
		t.Fatalf("Register() error = %v, want context canceled", registerErr)
	}
	if !client.isClosed() {
		t.Fatal("etcd client was not closed")
	}
}

func TestProviderEtcdGracefulShutdownDeletesExactKeysAndRevokesLeases(t *testing.T) {
	const (
		firstRecordedLease clientv3.LeaseID = 11
		firstActualLease   clientv3.LeaseID = 12
		secondLease        clientv3.LeaseID = 21
	)

	firstKey := filepath.Join(defaultEtcdPrefix, string(discover.ServiceNameBackend), "first")
	secondKey := filepath.Join(defaultEtcdPrefix, string(discover.ServiceNameBackend), "second")
	client := &fakeEtcdClient{
		deleteResponses: map[string]*clientv3.DeleteResponse{
			firstKey:  {PrevKvs: []*mvccpb.KeyValue{{Key: []byte(firstKey), Lease: int64(firstActualLease)}}},
			secondKey: {PrevKvs: []*mvccpb.KeyValue{{Key: []byte(secondKey), Lease: int64(secondLease)}}},
		},
		revokeErrors: map[clientv3.LeaseID]error{firstRecordedLease: errors.New("revoke failed")},
	}
	provider := newShutdownTestProvider(client)
	holder := provider.getLocalInstanceHolder(discover.ServiceNameBackend)
	first := discover.Instance{ID: "first", Name: "first"}
	first.SetMeta(metaKeyLeaseID, int64(firstRecordedLease))
	holder.upsert(first)
	second := discover.Instance{ID: "second", Name: "second"}
	second.SetMeta(metaKeyLeaseID, int64(secondLease))
	holder.upsert(second)

	err := provider.gracefulShutdown(time.Second, time.Second)
	if err == nil {
		t.Fatal("gracefulShutdown() error = nil, want revoke error")
	}
	if !client.isClosed() {
		t.Fatal("etcd client was not closed")
	}
	deletedKeys, deleteOptionCounts, revokedLeaseIDs := client.calls()
	if len(deletedKeys) != 2 {
		t.Fatalf("deleted keys = %v, want two exact keys", deletedKeys)
	}
	for _, optionCount := range deleteOptionCounts {
		if optionCount != 1 {
			t.Fatalf("delete option count = %d, want WithPrevKV option", optionCount)
		}
	}

	wantLeases := map[clientv3.LeaseID]bool{
		firstRecordedLease: false,
		firstActualLease:   false,
		secondLease:        false,
	}
	for _, leaseID := range revokedLeaseIDs {
		if _, ok := wantLeases[leaseID]; !ok {
			t.Fatalf("unexpected revoked lease %d", leaseID)
		}
		if wantLeases[leaseID] {
			t.Fatalf("lease %d was revoked more than once", leaseID)
		}
		wantLeases[leaseID] = true
	}
	for leaseID, revoked := range wantLeases {
		if !revoked {
			t.Fatalf("lease %d was not revoked", leaseID)
		}
	}
}

func TestProviderEtcdGracefulShutdownIgnoresRevokedLease(t *testing.T) {
	const leaseID clientv3.LeaseID = 31

	client := &fakeEtcdClient{
		revokeErrors: map[clientv3.LeaseID]error{leaseID: rpctypes.ErrLeaseNotFound},
	}
	provider := newShutdownTestProvider(client)
	instance := discover.Instance{ID: "instance", Name: "instance"}
	instance.SetMeta(metaKeyLeaseID, int64(leaseID))
	provider.getLocalInstanceHolder(discover.ServiceNameBackend).upsert(instance)

	if err := provider.GracefulShutdown(); err != nil {
		t.Fatalf("GracefulShutdown() error = %v", err)
	}
	if !client.isClosed() {
		t.Fatal("etcd client was not closed")
	}
}
