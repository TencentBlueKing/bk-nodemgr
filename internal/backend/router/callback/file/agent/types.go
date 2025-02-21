/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import "fmt"

// Access this config defines the access config of gse agent.
type Access struct {
	EnableStaticAccess bool   `json:"enable_static_access"`
	EnableFakeSeed     bool   `json:"enable_fake_seed"`
	ClusterEndpoints   string `json:"cluster_endpoints"`
	DataEndpoints      string `json:"data_endpoints"`
	FileEndpoints      string `json:"file_endpoints"`
}

// Base ths is the base config of gse agent.
type Base struct {
	TLSCaFile          string `json:"tls_ca_file"`
	TLSCertFile        string `json:"tls_cert_file"`
	TLSKeyFile         string `json:"tls_key_file"`
	TLSPasswdFile      string `json:"tls_passwd_file"`
	ProcessorNum       int    `json:"processor_num"`
	ProcessorQueueSize int    `json:"processor_queue_size"`
}

// Proxy this config will be used, when run mode is proxy.
type Proxy struct {
	TLSCaFile     string `json:"tls_ca_file"`
	TLSCertFile   string `json:"tls_cert_file"`
	TLSKeyFile    string `json:"tls_key_file"`
	TLSPasswdFile string `json:"tls_passwd_file"`
	BindIp        string `json:"bind_ip"`
	BindPort      int    `json:"bind_port"`
	ThreadNum     int    `json:"thread_num"`
}

// Task define the task config of gse agent.
type Task struct {
	ProcEventDataID          int `json:"proc_event_data_id"`
	ConcurrenceCount         int `json:"concurrence_count"`
	ScriptFileExpireTimeHour int `json:"script_file_expire_time_hour"`
}

// Data define the data config of gse agent.
type Data struct {
	IPC               string `json:"ipc"`
	EnableCompression bool   `json:"enable_compression"`
}

// File define the file config of gse agent.
type File struct {
	MaxTransferSpeedMbPerSec  int    `json:"max_transfer_speed_mb_per_sec"`
	MaxTransferConcurrentNum  int    `json:"max_transfer_concurrent_num"`
	BtListenInterface         string `json:"bt_listen_interface"`
	BtOutgoingInterface       string `json:"bt_outgoing_interface"`
	BtEnableOutgoingInterface bool   `json:"bt_enable_outgoing_interface"`
}

// Logger define the logger config of gse agent.
type Logger struct {
	Path            string `json:"path"`
	Level           string `json:"level"`
	FilesizeMb      int    `json:"filesize_mb"`
	Filenum         int    `json:"filenum"`
	Rotate          int    `json:"rotate"`
	FlushIntervalMs int    `json:"flush_interval_ms"`
}

// GSEConfig gse agent config.
type GSEConfig struct {
	CMDBConfig `json:",inline"`
	RunMode    RunMode `json:"run_mode"`
	Access     Access  `json:"access"`
	Base       Base    `json:"base"`
	Proxy      Proxy   `json:"proxy"`
	Task       Task    `json:"task"`
	Data       Data    `json:"data"`
	File       File    `json:"file"`
	Logger     Logger  `json:"logger"`

	// ExtraConfigDirectory is the directory of extra config.
	// Deprecated, don't use it.
	ExtraConfigDirectory string `json:"extra_config_directory"`
}

// RunMode the run mode of gse agent.
type RunMode string

const (
	// RunModeAgent agent run mode.
	RunModeAgent RunMode = "agent"

	// RunModeProxy proxy run mode.
	RunModeProxy RunMode = "proxy"
)

// Validate validates the run mode.
func (mode RunMode) Validate() error {
	switch mode {
	case RunModeAgent, RunModeProxy:
		return nil
	default:
		return fmt.Errorf("invalid run mode, mode(%s)", mode)
	}
}

// CMDBConfig cmdb config.
type CMDBConfig struct {
	// CloudID cloud id.
	CloudID int `json:"cloud_id"`

	// ZoneID zone id.
	ZoneID string `json:"zone_id"`

	// CityID city id.
	CityID string `json:"city_id"`
}
