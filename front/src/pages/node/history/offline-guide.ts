/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// OfflineGuideStep defines a single step in the offline install guide.
export interface OfflineGuideStep {
  // key is a stable identifier for the step.
  key: string;
  // name_zh is the Chinese step title.
  name_zh: string;
  // name_en is the English step title.
  name_en: string;
  // desc_zh is the Chinese step description shown below the title.
  desc_zh?: string;
  // desc_en is the English step description shown below the title.
  desc_en?: string;
  // has_result_input indicates that this step should show a result paste area and submit button.
  has_result_input?: boolean;
}

// offlineGuideSteps defines the ordered guide steps for offline proxy installation.
// The last step contains `has_result_input: true` which triggers the result submission UI.
export const offlineGuideSteps: OfflineGuideStep[] = [
  {
    key: 'download',
    name_zh: '下载离线安装包',
    name_en: 'Download offline install package',
    desc_zh: '点击下方按钮下载当前主机的离线安装包（每台主机的安装包唯一，包含专属配置）',
    desc_en: 'Click the button below to download the offline install package for this host (each host has a unique package with dedicated configuration)',
  },
  {
    key: 'transfer',
    name_zh: '将安装包传输至目标机器',
    name_en: 'Transfer the package to the target host',
    desc_zh: '自行将 .tar.gz 文件传输到目标机器上，并解压',
    desc_en: 'Transfer the .tar.gz file to the target host and extract it',
  },
  {
    key: 'execute',
    name_zh: '以 root 身份执行安装脚本',
    name_en: 'Execute the install script as root',
    desc_zh: '以 root 身份执行以下命令：',
    desc_en: 'Run the following commands as root:',
  },
  {
    key: 'submit',
    name_zh: '填写安装结果并提交',
    name_en: 'Fill in the install result and submit',
    desc_zh: '脚本执行成功后会输出 "--- installer.data.json ---" 并打印一段 JSON 数据，将其复制粘贴到下方输入框并点击"提交结果"',
    desc_en: 'After the script finishes successfully, it prints "--- installer.data.json ---" followed by a JSON snippet. Copy and paste it into the box below, then click "Submit result"',
    has_result_input: true,
  },
];
