//go:build integration

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

package release

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	testTenantID        = "tenant_a"
	releaseVersionBeta1 = "v2.1.6-beta.1"
	releaseVersionBeta2 = "v2.1.6-beta.2"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, client := testDatabaseAndClient(t)
	return client
}

func testDatabaseAndClient(t *testing.T) (*mongo.Database, IHandler) {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return db, New(db)
}

func testContext(tenantID string) contextx.IContext {
	return contextx.New(context.Background(), contextx.WithTenantID(tenantID))
}

func countReleaseDocuments(t *testing.T, db *mongo.Database, releaseType types.ReleaseType, tenantID string) int64 {
	t.Helper()

	num, err := db.Collection(TableName(string(releaseType), tenantID)).CountDocuments(context.Background(), bson.D{})
	if err != nil {
		t.Fatalf("CountDocuments() error = %v", err)
	}

	return num
}

func prepareData(t *testing.T, nCtx contextx.IContext) IHandler {
	t.Helper()

	client := testClient(t)
	if err := client.UpsertMany(nCtx, types.ReleaseTypeAgent, releaseFixtures()...); err != nil {
		t.Fatalf("failed to prepare releases: %v", err)
	}

	return client
}

func releaseFixtures() []*types.Release {
	return []*types.Release{
		{
			Name:       "agent-linux-amd64",
			Generation: types.Generation2,
			Type:       types.ReleaseTypeAgent,
			Version:    releaseVersionBeta1,
			Platform: platfmt.Platform{
				OS:   criteria.OSLinux,
				Arch: criteria.CPUArchAmd64,
			},
			Labels:   []string{"test"},
			FileName: "v2.1.6-beta.1-linux-amd64.tgz",
		},
		{
			Name:       "agent-linux-arm64",
			Generation: types.Generation2,
			Type:       types.ReleaseTypeAgent,
			Version:    releaseVersionBeta1,
			Platform: platfmt.Platform{
				OS:   criteria.OSLinux,
				Arch: criteria.CPUArchArm64,
			},
			Labels:   []string{"test"},
			FileName: "v2.1.6-beta.1-linux-arm64.tgz",
		},
		{
			Name:       "agent-windows-amd64",
			Generation: types.Generation2,
			Type:       types.ReleaseTypeAgent,
			Version:    releaseVersionBeta2,
			Platform: platfmt.Platform{
				OS:   criteria.OSWindows,
				Arch: criteria.CPUArchAmd64,
			},
			Labels:   []string{"test"},
			FileName: "v2.1.6-beta.2-windows-amd64.tgz",
			IsHidden: true,
		},
	}
}

// Test_UpsertMany tests the upsert many.
func Test_UpsertMany(t *testing.T) {
	nCtx := testContext(testTenantID)
	client := testClient(t)

	if err := client.UpsertMany(nCtx, types.ReleaseTypeAgent, releaseFixtures()...); err != nil {
		t.Fatalf("UpsertMany() error = %v", err)
	}

	total, err := client.Count(nCtx, types.ReleaseTypeAgent)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}

	if total != 3 {
		t.Fatalf("Count() total = %d, want %d", total, 3)
	}
}

func Test_Get(t *testing.T) {
	nCtx := testContext(testTenantID)
	client := prepareData(t, nCtx)

	got, err := client.Get(
		nCtx,
		types.ReleaseTypeAgent,
		WithGeneration(types.Generation2),
		WithVersion(releaseVersionBeta1),
		WithPlatform(platfmt.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}),
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got == nil {
		t.Fatal("Get() got nil release")
	}

	if got.FileName != "v2.1.6-beta.1-linux-amd64.tgz" {
		t.Fatalf("Get() fileName = %q, want %q", got.FileName, "v2.1.6-beta.1-linux-amd64.tgz")
	}
}

func Test_List(t *testing.T) {
	nCtx := testContext(testTenantID)
	client := prepareData(t, nCtx)

	releases, total, err := client.List(
		nCtx,
		types.ReleaseTypeAgent,
		types.Page{Offset: 0, Limit: 10},
		WithVersion(releaseVersionBeta1),
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("List() gotNum = %d, want %d", len(releases), 2)
	}

	if total != 2 {
		t.Fatalf("List() total = %d, want %d", total, 2)
	}
}

func Test_Delete(t *testing.T) {
	nCtx := testContext(testTenantID)
	client := prepareData(t, nCtx)

	if err := client.Delete(nCtx, types.ReleaseTypeAgent, WithVersion(releaseVersionBeta1)); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	total, err := client.Count(nCtx, types.ReleaseTypeAgent)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}

	if total != 1 {
		t.Fatalf("Count() total = %d, want %d", total, 1)
	}
}

func Test_WithIsHidden(t *testing.T) {
	nCtx := testContext(testTenantID)
	client := prepareData(t, nCtx)

	hiddenCount, err := client.Count(nCtx, types.ReleaseTypeAgent, WithIsHidden(true))
	if err != nil {
		t.Fatalf("Count() hidden error = %v", err)
	}

	if hiddenCount != 1 {
		t.Fatalf("Count() hidden total = %d, want %d", hiddenCount, 1)
	}

	visibleCount, err := client.Count(nCtx, types.ReleaseTypeAgent, WithIsHidden(false))
	if err != nil {
		t.Fatalf("Count() visible error = %v", err)
	}

	if visibleCount != 2 {
		t.Fatalf("Count() visible total = %d, want %d", visibleCount, 2)
	}
}

func Test_TenantIsolation(t *testing.T) {
	db, client := testDatabaseAndClient(t)
	tenantACtx := testContext("tenant_a")
	tenantBCtx := testContext("tenant_b")

	tenantARelease := releaseFixtures()[0]
	tenantARelease.FileName = "tenant-a-agent.tgz"
	tenantBRelease := releaseFixtures()[0]
	tenantBRelease.FileName = "tenant-b-agent.tgz"

	if err := client.UpsertMany(tenantACtx, types.ReleaseTypeAgent, tenantARelease); err != nil {
		t.Fatalf("UpsertMany() tenant_a error = %v", err)
	}

	if err := client.UpsertMany(tenantBCtx, types.ReleaseTypeAgent, tenantBRelease); err != nil {
		t.Fatalf("UpsertMany() tenant_b error = %v", err)
	}

	if countReleaseDocuments(t, db, types.ReleaseTypeAgent, "tenant_a") != 1 {
		t.Fatalf("tenant_a release collection count mismatch")
	}

	if countReleaseDocuments(t, db, types.ReleaseTypeAgent, "tenant_b") != 1 {
		t.Fatalf("tenant_b release collection count mismatch")
	}

	getOpts := []OptFn{
		WithGeneration(types.Generation2),
		WithVersion(releaseVersionBeta1),
		WithPlatform(platfmt.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}),
	}
	tenantAGot, err := client.Get(tenantACtx, types.ReleaseTypeAgent, getOpts...)
	if err != nil {
		t.Fatalf("Get() tenant_a error = %v", err)
	}

	tenantBGot, err := client.Get(tenantBCtx, types.ReleaseTypeAgent, getOpts...)
	if err != nil {
		t.Fatalf("Get() tenant_b error = %v", err)
	}

	if tenantAGot.FileName != "tenant-a-agent.tgz" {
		t.Fatalf("Get() tenant_a fileName = %q, want %q", tenantAGot.FileName, "tenant-a-agent.tgz")
	}

	if tenantBGot.FileName != "tenant-b-agent.tgz" {
		t.Fatalf("Get() tenant_b fileName = %q, want %q", tenantBGot.FileName, "tenant-b-agent.tgz")
	}

	if err := client.Delete(tenantACtx, types.ReleaseTypeAgent, WithVersion(releaseVersionBeta1)); err != nil {
		t.Fatalf("Delete() tenant_a error = %v", err)
	}

	tenantATotal, err := client.Count(tenantACtx, types.ReleaseTypeAgent)
	if err != nil {
		t.Fatalf("Count() tenant_a error = %v", err)
	}

	tenantBTotal, err := client.Count(tenantBCtx, types.ReleaseTypeAgent)
	if err != nil {
		t.Fatalf("Count() tenant_b error = %v", err)
	}

	if tenantATotal != 0 {
		t.Fatalf("Count() tenant_a total = %d, want 0", tenantATotal)
	}

	if tenantBTotal != 1 {
		t.Fatalf("Count() tenant_b total = %d, want 1", tenantBTotal)
	}
}
