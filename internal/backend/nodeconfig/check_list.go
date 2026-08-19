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

package nodeconfig

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// GSEHomeDir GSE home directory.
	GSEHomeDir = "__BK_GSE_HOME_DIR__"

	// GSEProxyBindPort GSE proxy bind port.
	GSEProxyBindPort = "__BK_GSE_PROXY_BIND_PORT__"

	// GSELogFileSizeMB log file size in MB.
	GSELogFileSizeMB = "__BK_GSE_LOG_FILESIZE_MB__"

	// GSELogFileNum log file number.
	GSELogFileNum = "__BK_GSE_LOG_FILENUM__"

	// GSELogPath GSE log path.
	GSELogPath = "__BK_GSE_LOG_PATH__"

	// GSEDataAgentBindPort GSE data agent bind port.
	GSEDataAgentBindPort = "__BK_GSE_DATA_AGENT_BIND_PORT__"

	// GSEDataMetricExporterBindPort GSE data metric exporter bind port.
	GSEDataMetricExporterBindPort = "__BK_GSE_DATA_METRIC_EXPORTER_BIND_PORT__"

	// GSEFileBittorrentBindPort GSE file bittorrent bind port.
	GSEFileBittorrentBindPort = "__BK_GSE_FILE_BITTORRENT_BIND_PORT__"

	// GSEFileBittorrentTrackerBindPort GSE file bittorrent tracker bind port.
	GSEFileBittorrentTrackerBindPort = "__BK_GSE_FILE_BITTORRENT_TRACKER_BIND_PORT__"

	// GSEFileTopologyBindPort GSE file topology bind port.
	GSEFileTopologyBindPort = "__BK_GSE_FILE_TOPOLOGY_BIND_PORT__"

	// GSEFileTopologyThriftBindPort GSE file topology thrift bind port.
	GSEFileTopologyThriftBindPort = "__BK_GSE_FILE_TOPOLOGY_THRIFT_BIND_PORT__"

	// GSEFileMetricExporterBindPort GSE file metric exporter bind port.
	GSEFileMetricExporterBindPort = "__BK_GSE_FILE_METRIC_EXPORTER_BIND_PORT__"
)

// CheckList is the pre-install check list for a node.
type CheckList struct {
	DiskRequires    []DiskRequire   `json:"disk_requires"`
	PortPolicies    []PortPolicy    `json:"port_policies"`
	NetworkPolicies []NetworkPolicy `json:"network_policies"`
}

// DiskRequire describes a disk space requirement.
type DiskRequire struct {
	DemandMB uint64 `json:"demand_mb"`
	DirPath  string `json:"dir_path"`
}

// PortPolicy describes a port availability check.
type PortPolicy struct {
	BindIP  string           `json:"bind_ip"`
	Port    uint64           `json:"port"`
	Network criteria.NetType `json:"network"`
}

// NetworkPolicy describes a network connectivity check.
type NetworkPolicy struct {
	Host    string           `json:"host"`
	Port    uint64           `json:"port"`
	Network criteria.NetType `json:"network"`
}

func calCheckListDiskRequires(nodeConf *types.NodeConf) ([]DiskRequire, error) {
	diskRequiresMap := map[string]uint64{
		GSEHomeDir: 300, // nolint: mnd
	}

	logFileSize, err := conv.ToInt64(nodeConf.PreSetting[GSELogFileSizeMB])
	if err != nil {
		return nil, fmt.Errorf("invalid node conf, key(%s) , value(%v)",
			GSELogFileSizeMB, nodeConf.PreSetting[GSELogFileSizeMB])
	}
	logFileNum, err := conv.ToInt64(nodeConf.PreSetting[GSELogFileNum])
	if err != nil {
		return nil, fmt.Errorf("invalid node conf, key(%s) , value(%v)",
			GSELogFileNum, nodeConf.PreSetting[GSELogFileNum])
	}

	diskRequiresMap[GSELogPath] = uint64(logFileSize) * uint64(logFileNum)

	diskRequires := make([]DiskRequire, 0)
	for key, value := range diskRequiresMap {
		dirPath, ok := nodeConf.PreSetting[key]
		if !ok {
			continue
		}
		dirPathStr, err := conv.ToString(dirPath)
		if err != nil {
			return nil, fmt.Errorf("invalid node conf, key(%s) , value(%v)", key, dirPath)
		}

		diskRequires = append(diskRequires, DiskRequire{
			DemandMB: value,
			DirPath:  dirPathStr,
		})
	}

	return diskRequires, nil
}

func calCheckListPortPolicies(deployInfo *types.DeploymentInfo, _ *types.NodeConf) ([]PortPolicy, error) {
	switch deployInfo.Host.Dynamic.NodeRole {
	case types.NodeRoleProxy:
		return []PortPolicy{}, nil

	case types.NodeRoleAgent:
		return []PortPolicy{}, nil

	default:
		return nil, fmt.Errorf("invalid node role: %s", deployInfo.Host.Dynamic.NodeRole)
	}
}

func calCheckListNetworkPolicies(_ *types.NodeConf) ([]NetworkPolicy, error) {
	// TODO: implement me

	return []NetworkPolicy{}, nil
}

// BuildCheckList assembles a complete pre-install CheckList from deployment info and node config.
func BuildCheckList(deployInfo *types.DeploymentInfo, nodeConf *types.NodeConf) (*CheckList, error) {
	diskRequires, err := calCheckListDiskRequires(nodeConf)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate disk requires: %w", err)
	}

	portPolicies, err := calCheckListPortPolicies(deployInfo, nodeConf)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate port policies: %w", err)
	}

	networkPolicies, err := calCheckListNetworkPolicies(nodeConf)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate network policies: %w", err)
	}

	return &CheckList{
		DiskRequires:    diskRequires,
		PortPolicies:    portPolicies,
		NetworkPolicies: networkPolicies,
	}, nil
}
