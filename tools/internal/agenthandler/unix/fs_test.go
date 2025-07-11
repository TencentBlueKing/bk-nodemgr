//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package unix

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/joho/godotenv"
)

var (
	rootAbsDir   string
	agentPkgPath string
)

func testClient(t *testing.T) agenthandler.IAgentFSHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	rootAbsDir = os.Getenv("ROOT_ABS_DIR")
	agentPkgPath = os.Getenv("AGENT_PKG_PATH")

	return NewAgentHandler(rootAbsDir)
}

// Test_CheckIntegrity tests CheckIntegrity.
func Test_CheckIntegrity(t *testing.T) {
	handler := testClient(t)

	if err := handler.Init(context.Background()); err != nil {
		t.Fatalf("failed to init fs: %v", err)
	}

	fileAgent, err := os.Create(filepath.Join(rootAbsDir, "agent", "bin", "gse_agent"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_ = fileAgent.Close()

	fileConf, err := os.Create(filepath.Join(rootAbsDir, "agent", "etc", "gse_agent.conf"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_ = fileConf.Close()

	if err := handler.CheckIntegrity(context.Background()); err != nil {
		t.Fatalf("failed to check integrity: %v", err)
	}
}

// Test_Init tests Init.
func Test_Init(t *testing.T) {
	handler := testClient(t)

	if err := handler.Init(context.Background()); err != nil {
		t.Fatalf("failed to init fs: %v", err)
	}
}

// Test_Backup tests Backup.
func Test_Backup(t *testing.T) {
	handler := testClient(t)

	if err := handler.Backup(context.Background()); err != nil {
		t.Fatalf("failed to backup fs: %v", err)
	}
}

// Test_Purge tests Purge.
func Test_Purge(t *testing.T) {
	handler := testClient(t)

	if err := handler.Purge(context.Background()); err != nil {
		t.Fatalf("failed to purge fs: %v", err)
	}
}

// Test_CopyConfigDir tests CopyConfigDir.
func Test_CopyConfigDir(t *testing.T) {
	handler := testClient(t)

	if err := handler.Init(context.Background()); err != nil {
		t.Fatalf("failed to init fs: %v", err)
	}

	tmpSource := filepath.Join(rootAbsDir, "tmp_cert")
	if err := os.MkdirAll(tmpSource, os.ModePerm); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	tmp1, err := os.Create(filepath.Join(tmpSource, "tmp1.conf"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_ = tmp1.Close()

	tmp2, err := os.Create(filepath.Join(tmpSource, "tmp2.conf"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_ = tmp2.Close()

	tmp3, err := os.Create(filepath.Join(tmpSource, "tmp3.conf"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_ = tmp3.Close()

	if err := handler.CopyConfigDir(context.Background(), tmpSource); err != nil {
		t.Fatalf("failed to copy config dir: %v", err)
	}
}

// Test_UnpackReleasePackage tests UnpackReleasePackage.
func Test_UnpackReleasePackage(t *testing.T) {
	handler := testClient(t)

	if err := handler.UnpackReleasePackage(context.Background(), agentPkgPath, false); err != nil {
		t.Fatalf("failed to unpack release package: %v", err)
	}

	if err := handler.UnpackReleasePackage(context.Background(), agentPkgPath, true); err != nil {
		t.Fatalf("failed to unpack release package: %v", err)
	}
}

// Test_Clean tests Clean.
func Test_Clean(t *testing.T) {
	handler := testClient(t)

	if err := handler.Clean(context.Background()); err != nil {
		t.Fatalf("failed to clean fs: %v", err)
	}
}
