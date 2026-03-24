/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// TAG_NEED_OFFLINE_MANUAL_INSTALL is the tag emitted by the WaitOfflineManualInstall action.
// It must match the Go constant `TagNeedOfflineManualInstall` in pkg/workflow/action/definition.go.
export const TAG_NEED_OFFLINE_MANUAL_INSTALL = 'need_offline_manual_install';

// STEP_KEY_WAIT_OFFLINE_MANUAL_INSTALL is the oper_inst_logs map key for that action.
// It must match `ActionNameWaitOfflineManualInstall` in
// internal/backend/manager/workflowdef/node/action_wait_offline_manual_install.go.
export const STEP_KEY_WAIT_OFFLINE_MANUAL_INSTALL = 'wait_offline_manual_install';

// True when action tags indicate an offline install waiting step.
export const isOfflineGuideStep = (tags: string[]): boolean => (
  Array.isArray(tags) && tags.includes(TAG_NEED_OFFLINE_MANUAL_INSTALL)
);

// OFFLINE_PACKAGE_DOWNLOAD_PATH is the application-layer POST path for streaming the offline
// install package (body: { operation_id }).
export const OFFLINE_PACKAGE_DOWNLOAD_PATH = '/api/v3/node/workflow/operation/offline/download';
