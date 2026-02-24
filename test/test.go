/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package test provides common configuration and utilities for integration tests.
package test

import "flag"

// Test flags are initialized in InitFlags.
// nolint:gochecknoglobals
var (
	// TestFlagBackendServer backend test flag.
	TestFlagBackendServer string
	// TestFlagApplicationServer application test flag.
	TestFlagApplicationServer string
	// TestFlagFileServer file test flag.
	TestFlagFileServer string

	// TestFlagDataDir path to test data directory for precheck.
	TestFlagDataDir string
)

// InitFlags initializes test command line flags.
func InitFlags() {
	flag.StringVar(&TestFlagBackendServer, "backend", "127.0.0.1:28100", "backend host and port")
	flag.StringVar(&TestFlagApplicationServer, "application", "127.0.0.1:28000", "application host and port")
	flag.StringVar(&TestFlagFileServer, "file", "127.0.0.1:28200", "file host and port")

	flag.StringVar(&TestFlagDataDir, "data-dir", "", "path to test data directory")
}
