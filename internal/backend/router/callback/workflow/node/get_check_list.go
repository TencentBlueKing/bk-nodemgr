/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package node ...
package node

import (
	"fmt"
	"net/http"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetCheckList get gse node check list.
func (h *handler) GetCheckList(gCtx *gin.Context) {
	req := new(protoBackend.GetCheckListReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.GetNodeDeploymentNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf := new(CheckList)
	conf.DiskRequires, err = h.calCheckListDiskRequires(nodeConf)
	if err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf.PortPolicies, err = h.calCheckListPortPolicies(nodeConf)
	if err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf.NetworkPolicies, err = h.calCheckListNetworkPolicies(nodeConf)
	if err != nil {
		h.logger.Errorf("get gse node check list failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.JSON(http.StatusOK, conf)

	return
}

const (
	// GSELogFileSizeMB log file size in MB.
	GSELogFileSizeMB = "__BK_GSE_LOG_FILESIZE_MB__"

	// GSELogFileNum log file number.
	GSELogFileNum = "__BK_GSE_LOG_FILENUM__"
)

func (h *handler) calCheckListDiskRequires(nodeConf *types.NodeConf) ([]DiskRequire, error) {
	diskRequiresMap := map[string]uint64{
		"__BK_GSE_HOME_DIR__": 300, // nolint: mnd
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

	diskRequiresMap["__BK_GSE_LOG_PATH__"] = uint64(logFileSize) * uint64(logFileNum)

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

func (h *handler) calCheckListPortPolicies(nodeConf *types.NodeConf) ([]PortPolicy, error) {
	portPoliciesMap := map[string]criteria.NetType{
		// "__BK_GSE_DATA_AGENT_BIND_PORT__":              criteria.NetTypeTCP,
		// "__BK_GSE_DATA_METRIC_EXPORTER_BIND_PORT__":    criteria.NetTypeTCP,
		// "__BK_GSE_FILE_BITTORRENT_BIND_PORT__":         criteria.NetTypeTCP,
		// "__BK_GSE_FILE_BITTORRENT_TRACKER_BIND_PORT__": criteria.NetTypeTCP,
		// "__BK_GSE_FILE_TOPOLOGY_BIND_PORT__":           criteria.NetTypeTCP,
		// "__BK_GSE_FILE_TOPOLOGY_THRIFT_BIND_PORT__":    criteria.NetTypeTCP,
		// "__BK_GSE_FILE_METRIC_EXPORTER_BIND_PORT__":    criteria.NetTypeTCP,
		"__BK_GSE_PROXY_BIND_PORT__": criteria.NetTypeTCP,
	}

	portPolicies := make([]PortPolicy, 0)
	for key, netType := range portPoliciesMap {
		port, ok := nodeConf.PreSetting[key]
		if !ok {
			continue
		}
		portNum, err := conv.ToInt64(port)
		if err != nil {
			return nil, fmt.Errorf("invalid node conf, key(%s) , value(%v)", key, port)
		}

		portPolicies = append(portPolicies, PortPolicy{
			Port:    uint64(portNum),
			Network: netType,
		})
	}

	return portPolicies, nil
}

func (h *handler) calCheckListNetworkPolicies(nodeConf *types.NodeConf) ([]NetworkPolicy, error) {
	// TODO: implement me

	return nil, nil
}

// CheckList check list.
type CheckList struct {
	// DiskRequires disk require.
	DiskRequires []DiskRequire `json:"disk_requires"`
	// PortPolicies port policy.
	PortPolicies []PortPolicy `json:"port_policies"`
	// NetworkPolicies network policy.
	NetworkPolicies []NetworkPolicy `json:"network_policies"`
}

// DiskRequire disk require.
type DiskRequire struct {
	DemandMB uint64 `json:"demand_mb"`
	DirPath  string `json:"dir_path"`
}

// PortPolicy port policy.
type PortPolicy struct {
	// policy port
	Port uint64 `json:"port"`

	// Network
	Network criteria.NetType `json:"network"`
}

// NetworkPolicy network policy.
type NetworkPolicy struct {
	Host    string           `json:"host"`
	Port    uint64           `json:"port"`
	Network criteria.NetType `json:"network"`
}
