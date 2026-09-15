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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/joho/godotenv"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// registrationTestEndpoint returns the etcd endpoint to test against, skipping
// the test when no etcd is configured.
func registrationTestEndpoint(t *testing.T) string {
	t.Helper()

	endpoint := os.Getenv("ETCD_ENDPOINT")
	if endpoint == "" {
		_ = godotenv.Load(".env")
		endpoint = os.Getenv("ETCD_ENDPOINT")
	}

	if endpoint == "" {
		t.Skip("ETCD_ENDPOINT is not set, skipping the registration tests")
	}

	return endpoint
}

// registrationTestPrefix returns a discover prefix unique to this test run, so
// that concurrent runs against a shared etcd do not overwrite each other.
func registrationTestPrefix(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("/registration-test/%s-%d", t.Name(), time.Now().UnixNano())
}

func registrationTestProvider(t *testing.T) (*ProviderEtcd, *clientv3.Client, string) {
	t.Helper()

	endpoint := registrationTestEndpoint(t)
	prefix := registrationTestPrefix(t)

	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{endpoint}, DialTimeout: defaultEtcdDialTimeout})
	if err != nil {
		t.Fatalf("failed to create the assertion client: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })

	provider := NewProviderEtcd(&config.Etcd{Endpoints: []string{endpoint}},
		WithDiscoverPathPrefix(prefix), WithWatch(discover.ServiceNameBackend))
	if err = provider.Start(context.Background()); err != nil {
		t.Fatalf("failed to start the provider: %v", err)
	}
	t.Cleanup(func() { _ = provider.GracefulShutdown() })

	return provider, cli, prefix
}

func registrationTestInstance() discover.Instance {
	return discover.Instance{
		ID:   "registration-test",
		Name: "registration test",
		Endpoints: map[discover.EndpointName]discover.Endpoint{
			discover.EndpointNameBackendBasic: {IPV4: "127.0.0.1", Port: 8000},
		},
	}
}

// instanceKV reads the single instance key under the prefix.
func instanceKV(t *testing.T, cli *clientv3.Client, prefix string) *mvccpb.KeyValue {
	t.Helper()

	resp, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		t.Fatalf("failed to read the instance key: %v", err)
	}

	if len(resp.Kvs) != 1 {
		t.Fatalf("expected exactly one instance key under %s, got %d", prefix, len(resp.Kvs))
	}

	return resp.Kvs[0]
}

// TestRegistrationCostsNoRevisionWhenHealthy guards the property the whole
// registration design is built on: keeping an instance registered relies on
// lease keepalives only, which create no MVCC revision. A registration that
// re-puts itself periodically silently fills up the etcd backend instead.
func TestRegistrationCostsNoRevisionWhenHealthy(t *testing.T) {
	provider, cli, prefix := registrationTestProvider(t)

	if err := provider.Register(discover.ServiceNameBackend, registrationTestInstance()); err != nil {
		t.Fatalf("failed to register the instance: %v", err)
	}

	time.Sleep(2 * time.Second)

	// assert on the instance key itself rather than on the cluster revision,
	// so that unrelated writers on a shared etcd cannot fail the test.
	before := instanceKV(t, cli, prefix)

	// longer than a lease TTL, so the window covers a full round of lease
	// renewals rather than the quiet gap between them.
	time.Sleep((defaultEtcdLeaseTTLSec + 2) * time.Second)

	after := instanceKV(t, cli, prefix)
	if after.ModRevision != before.ModRevision {
		t.Fatalf("a healthy registration re-wrote the instance key: mod revision %d -> %d, want no write",
			before.ModRevision, after.ModRevision)
	}

	instances, err := provider.GetAllService(discover.ServiceNameBackend)
	if err != nil {
		t.Fatalf("failed to get the registered service: %v", err)
	}

	if len(instances) != 1 || instances[0].ID != registrationTestInstance().ID {
		t.Fatalf("the registered instance is not discoverable: %+v", instances)
	}
}

// TestRegistrationHealsAfterLeaseLoss asserts the instance comes back on a new
// lease once the lease backing it is destroyed underneath the process.
func TestRegistrationHealsAfterLeaseLoss(t *testing.T) {
	provider, cli, prefix := registrationTestProvider(t)

	if err := provider.Register(discover.ServiceNameBackend, registrationTestInstance()); err != nil {
		t.Fatalf("failed to register the instance: %v", err)
	}
	time.Sleep(time.Second)

	lostLease := instanceKV(t, cli, prefix).Lease
	if _, err := cli.Revoke(context.Background(), clientv3.LeaseID(lostLease)); err != nil {
		t.Fatalf("failed to revoke the registration lease: %v", err)
	}

	deadline := time.Now().Add((defaultEtcdLeaseTTLSec + 5) * time.Second)
	for time.Now().Before(deadline) {
		resp, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
		if err == nil && len(resp.Kvs) == 1 && resp.Kvs[0].Lease != lostLease {
			return
		}

		time.Sleep(200 * time.Millisecond) // nolint: mnd
	}

	t.Fatal("the instance was not re-registered after its lease was lost")
}

// TestDeregisterIsFinal asserts a deregistered instance is not brought back by
// the keeper that was keeping it registered.
func TestDeregisterIsFinal(t *testing.T) {
	provider, cli, prefix := registrationTestProvider(t)

	instance := registrationTestInstance()
	if err := provider.Register(discover.ServiceNameBackend, instance); err != nil {
		t.Fatalf("failed to register the instance: %v", err)
	}
	time.Sleep(time.Second)

	if err := provider.Deregister(discover.ServiceNameBackend, instance.ID); err != nil {
		t.Fatalf("failed to deregister the instance: %v", err)
	}

	// past a lease TTL: a keeper reacting to the revoked lease would have
	// re-registered the instance by now.
	time.Sleep((defaultEtcdLeaseTTLSec + 2) * time.Second)

	resp, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		t.Fatalf("failed to read the deregistered instance: %v", err)
	}

	if len(resp.Kvs) != 0 {
		t.Fatalf("the deregistered instance was brought back: %d keys", len(resp.Kvs))
	}
}

// TestGracefulShutdownRemovesInstance asserts a shutdown drops the instance
// from discovery at once instead of leaving it visible for a lease TTL.
func TestGracefulShutdownRemovesInstance(t *testing.T) {
	provider, cli, prefix := registrationTestProvider(t)

	if err := provider.Register(discover.ServiceNameBackend, registrationTestInstance()); err != nil {
		t.Fatalf("failed to register the instance: %v", err)
	}
	time.Sleep(time.Second)

	if err := provider.GracefulShutdown(); err != nil {
		t.Fatalf("failed to shut down the provider: %v", err)
	}

	resp, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		t.Fatalf("failed to read the instance after shutdown: %v", err)
	}

	if len(resp.Kvs) != 0 {
		t.Fatalf("the instance is still registered after a graceful shutdown: %d keys", len(resp.Kvs))
	}
}
