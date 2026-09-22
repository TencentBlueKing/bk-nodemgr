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

package plugindeployment

const (
	// FieldKeyToken the token field key.
	FieldKeyToken = "data.token"

	// FieldKeyInfo the info field key.
	FieldKeyInfo = "data.info"

	// FieldKeyPluginConf the plugin conf field key.
	FieldKeyPluginConf = "data.plugin_config"

	// FieldKeyPluginConfSet the plugin_config.set field key.
	FieldKeyPluginConfSet = "data.plugin_config.set"

	// FieldKeyPluginConfConfigFilesDetail the plugin_config.config_files_detail field key.
	FieldKeyPluginConfConfigFilesDetail = "data.plugin_config.config_files_detail"

	// FieldKeyPluginConfRemoveAllConfigs the plugin_config.remove_all_configs field key.
	FieldKeyPluginConfRemoveAllConfigs = "data.plugin_config.remove_all_configs"

	// FieldKeyExpireAt the expire_at field key.
	FieldKeyExpireAt = "data.expire_at"

	// FieldKeyHostID the host id field key.
	FieldKeyHostID = "data.info.process.host_id"

	// FieldKeyPluginName the plugin name field key.
	FieldKeyPluginName = "data.info.process.name"

	// FieldKeyProcessLastSyncAt the process last sync time field key.
	FieldKeyProcessLastSyncAt = "data.info.process.info.last_sync_at"

	// FieldKeyPluginVersion the plugin version field key.
	FieldKeyPluginVersion = "data.info.install_options.version"
)
