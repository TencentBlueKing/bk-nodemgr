/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	valueGroupKeyFileMandatoryTCP = "VALUE_GROUP_FILE_MANDATORY_TCP"
	valueGroupKeyProxyLoggerLevel = "VALUE_GROUP_PROXY_LOGGER_LEVEL"
	valueGroupKeyProxyLoggerPath  = "VALUE_GROUP_PROXY_LOGGER_PATH"
)

var valueGroupKeyMap = map[string]struct{}{
	valueGroupKeyFileMandatoryTCP: {},
	valueGroupKeyProxyLoggerLevel: {},
	valueGroupKeyProxyLoggerPath:  {},
}

func isValueGroupKey(key string) bool {
	_, ok := valueGroupKeyMap[key]
	return ok
}

func getConfigTemplate(nodeRole types.NodeRole) []types.ConfigPolicyTemplateBlock {
	switch nodeRole {
	case types.NodeRoleAgent:
		return agentConfig
	case types.NodeRoleProxy:
		return proxyConfig
	default:
		return nil
	}
}

func findTemplateItem(blocks []types.ConfigPolicyTemplateBlock, blockID, itemID string) (types.ConfigPolicyTemplateItem, bool) {
	for _, block := range blocks {
		if block.ID == blockID {
			for _, item := range block.Items {
				if item.ID == itemID {
					return item, true
				}
			}
		}
	}

	return types.ConfigPolicyTemplateItem{}, false
}

var agentConfig = []types.ConfigPolicyTemplateBlock{
	{
		ID:      "base_config",
		TitleEN: "Base config",
		TitleZH: "基础配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:        "enable_fake_seed",
				NameEN:    "Non-fixed MAC address mode",
				NameZH:    "不固定MAC地址模式",
				RemarkEN:  "Unfixed MAC addresses may pose security risks. Please choose carefully.",
				RemarkZH:  "不固定MAC地址可能存在安全风险, 请谨慎选择.",
				Key:       "agent.access.enable_fake_seed",
				Type:      types.ConfigPolicyTemplateTypeBool,
				ValueBool: false,
			},
			{
				ID:          "logger_level",
				NameEN:      "Log level",
				NameZH:      "日志等级",
				RemarkEN:    "The lower the log level, the more logs are printed.",
				RemarkZH:    "日志等级越低打印的日志数量越多.",
				Key:         "agent.logger.level",
				Type:        types.ConfigPolicyTemplateTypeStringSelect,
				ValueString: "INFO",
				ValueStringSelect: []string{
					"DEBUG",
					"INFO",
					"WARN",
					"ERROR",
				},
			},
		},
	},
	{
		ID:      "script_config",
		TitleEN: "Script task config",
		TitleZH: "脚本任务配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:       "concurrent",
				NameEN:   "Maximum concurrent execution",
				NameZH:   "最大同时执行数量",
				RemarkEN: "If the number exceeds this limit, it will be queued and executed after the currently executing task is completed. In early versions, tasks exceeding the specified number would be discarded directly.",
				RemarkZH: "超过该数量会排队, 等待正在执行的任务完成后再执行. 早期版本中, 超过数量的任务会被直接放弃.",
				Key:      "agent.task.concurrence_count",
				Type:     types.ConfigPolicyTemplateTypeInt,
				ValueInt: 100,
			},
		},
	},
	{
		ID:      "data_config",
		TitleEN: "Data report config",
		TitleZH: "数据上报配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:        "enable_data_compression",
				NameEN:    "Enable data compression",
				NameZH:    "开启数据压缩",
				RemarkEN:  "Enabling data compression can reduce traffic consumption on the reporting link, but it will increase CPU consumption.",
				RemarkZH:  "开启数据压缩可以减少上报链路的流量消耗, 但会额外增加CPU消耗.",
				Key:       "agent.data.enable_compression",
				Type:      types.ConfigPolicyTemplateTypeBool,
				ValueBool: false,
			},
		},
	},
	{
		ID:      "file_config",
		TitleEN: "File task config",
		TitleZH: "文件任务配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:       "max_speed",
				NameEN:   "Maximum transfer speed(MBytes/s)",
				NameZH:   "最大传输速度(MBytes/s)",
				RemarkEN: "Uploads and downloads are calculated separately.",
				RemarkZH: "上传和下载各自独立计算.",
				Key:      "agent.file.max_transfer_speed_mb_per_sec",
				Type:     types.ConfigPolicyTemplateTypeInt,
				ValueInt: 100,
			},
			{
				ID:       "concurrent",
				NameEN:   "Maximum concurrent execution",
				NameZH:   "最大同时执行数量",
				RemarkEN: "If the number exceeds this limit, it will be queued and executed after the currently executing task is completed.",
				RemarkZH: "超过该数量会排队, 等待正在执行的任务完成后再执行.",
				Key:      "agent.file.max_transfer_concurrent_num",
				Type:     types.ConfigPolicyTemplateTypeInt,
				ValueInt: 10,
			},
			{
				ID:        "forbid_utp",
				NameEN:    "Mandatory use of TCP transmission",
				NameZH:    "强制使用TCP传输",
				RemarkEN:  "Disabling uTP transmission in the BT protocol will result in more stable transmission. However, this may cause packet loss in cloud CVMs on Windows, so please use with caution.",
				RemarkZH:  "禁用BT协议中的uTP传输, 传输效果会更加稳定. 但在Windows上可能会导致云CVM入包Drop, 请谨慎使用.",
				Key:       valueGroupKeyFileMandatoryTCP,
				Type:      types.ConfigPolicyTemplateTypeBool,
				ValueBool: false,
				ValueGroupFixed: map[string]any{
					"agent.file.bt_enable_outgoing_utp": false,
					"agent.file.bt_enable_incoming_utp": false,
					"agent.file.bt_enable_outgoing_tcp": true,
					"agent.file.bt_enable_incoming_tcp": true,
				},
			},
			{
				ID:        "disable_listen_sockets",
				NameEN:    "Enable silent mode",
				NameZH:    "开启无监听模式",
				RemarkEN:  "The agent will not listen to any file service ports and will use one-way outbound TCP connections for file transfers.",
				RemarkZH:  "agent将不会监听任何文件服务端口, 使用单向对外的TCP链接进行文件传输.",
				Key:       "agent.file.disable_listen_sockets",
				Type:      types.ConfigPolicyTemplateTypeBool,
				ValueBool: false,
			},
		},
	},
}

var proxyConfig = []types.ConfigPolicyTemplateBlock{
	{
		ID:      "base_config",
		TitleEN: "Base config",
		TitleZH: "基础配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:          "all_logger_level",
				NameEN:      "Log level",
				NameZH:      "日志等级",
				RemarkEN:    "The lower the log level, the more logs are printed.",
				RemarkZH:    "日志等级越低打印的日志数量越多.",
				Key:         valueGroupKeyProxyLoggerLevel,
				Type:        types.ConfigPolicyTemplateTypeStringSelect,
				ValueString: "INFO",
				ValueStringSelect: []string{
					"DEBUG",
					"INFO",
					"WARN",
					"ERROR",
				},
				ValueGroupAssigned: map[string]any{
					"agent.logger.level": "",
					"file.logger.level":  "",
					"data.logger.level":  "",
				},
			},
			{
				ID:          "all_logger_path",
				NameEN:      "Log Path",
				NameZH:      "日志存储路径",
				RemarkEN:    "",
				RemarkZH:    "",
				Key:         valueGroupKeyProxyLoggerLevel,
				Type:        types.ConfigPolicyTemplateTypeString,
				ValueString: "",
				ValueGroupAssigned: map[string]any{
					"agent.logger.path": "",
					"file.logger.path":  "",
					"data.logger.path":  "",
				},
			},
		},
	},
	{
		ID:      "file_config",
		TitleEN: "File task config",
		TitleZH: "文件任务配置",
		Items: []types.ConfigPolicyTemplateItem{
			{
				ID:          "file_cache_dir",
				NameEN:      "Cache file storage path",
				NameZH:      "缓存文件存储路径",
				RemarkEN:    "All file transfers through this proxy are cached in this directory. The cache is automatically cleared, but if the disk where the directory is located has insufficient space, it will affect file transfers.",
				RemarkZH:    "经过该Proxy的文件传输都会缓存在这个目录下, 缓存会自动清理, 若目录所在磁盘可用空间不足会影响文件传输.",
				Key:         "file.cache.dirs",
				Type:        types.ConfigPolicyTemplateTypeString,
				ValueString: "",
			},
		},
	},
}
