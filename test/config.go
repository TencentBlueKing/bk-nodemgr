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

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultEnvFile = "env.yaml"

// nolint:gochecknoglobals
var (
	// TestFlagEnvFile path to the test environment config file.
	TestFlagEnvFile string

	// Env is the test environment configuration loaded from env.yaml.
	// It is populated by ParseFlagsAndLoadEnv and is safe to read after that call.
	Env *EnvConfig
)

// ParseFlagsAndLoadEnv parses CLI flags and loads env.yaml from the specified file.
func ParseFlagsAndLoadEnv() {
	flag.StringVar(&TestFlagEnvFile, "env-file", defaultEnvFile, "path to test environment file")
	flag.Parse()

	env, err := LoadEnvConfig(TestFlagEnvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load test env config: %v\n", err)
		os.Exit(1)
	}

	Env = env
}

const (
	defaultBackendBasicEndpoint     = "127.0.0.1:28102"
	defaultApplicationBasicEndpoint = "127.0.0.1:28002"
	defaultFileBasicEndpoint        = "127.0.0.1:28202"
	defaultFileDownloadEndpoint     = "127.0.0.1:28203"
)

// EnvConfig holds all test environment configuration loaded from env.yaml.
type EnvConfig struct {
	DataDir  string   `yaml:"dataDir"`
	Servers  Servers  `yaml:"servers"`
	Precheck Precheck `yaml:"precheck"`
}

// Servers holds the endpoints of each service under test.
type Servers struct {
	BackendBasicEndpoint     string `yaml:"backendBasic"`
	ApplicationBasicEndpoint string `yaml:"applicationBasic"`
	FileBasicEndpoint        string `yaml:"fileBasic"`
	FileDownloadEndpoint     string `yaml:"fileDownload"`
}

// Precheck groups parameters needed by precheck test cases.
type Precheck struct {
	File PrecheckFile `yaml:"file"`
}

// PrecheckFile holds file paths for the init-packages precheck tests.
type PrecheckFile struct {
	OriginCertPath         string `yaml:"originCertPath"`
	OriginBinToolPath      string `yaml:"originBinToolPath"`
	OriginAgentPath        string `yaml:"originAgentPath"`
	OriginPluginBinToolPath string `yaml:"originPluginBinToolPath"`
	OriginPluginPath       string `yaml:"originPluginPath"`
}

const defaultDataDir = "data"

// defaultEnvConfig returns the default environment configuration.
func defaultEnvConfig() *EnvConfig {
	return &EnvConfig{
		DataDir: defaultDataDir,
		Servers: Servers{
			BackendBasicEndpoint:     defaultBackendBasicEndpoint,
			ApplicationBasicEndpoint: defaultApplicationBasicEndpoint,
			FileBasicEndpoint:        defaultFileBasicEndpoint,
			FileDownloadEndpoint:     defaultFileDownloadEndpoint,
		},
	}
}

// LoadEnvConfig loads environment configuration from the specified file.
func LoadEnvConfig(envFile string) (*EnvConfig, error) {
	if envFile == "" {
		return defaultEnvConfig(), nil
	}

	path := filepath.Clean(envFile)

	data, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read env config from %s: %w", path, err)
	}

	env := defaultEnvConfig()
	if err := yaml.Unmarshal(data, env); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	return env, nil
}
