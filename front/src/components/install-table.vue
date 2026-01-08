<template>
  <div ref="contentRef">
    <VxeTable
      ref="xTableRef"
      :data="tableData"
      :size="settings.size"
      :border="true"
      :max-height="maxHeight"
      :scroll-y="{ enabled: true, gt: 20 }"
      :row-config="{ isHover: true, useKey: true }"
      round
    >
      <!-- 业务属性 -->
      <VxeColgroup title="业务属性" align="center" v-if="isReinstall">
        <VxeColumn
          field="bk_biz_id"
          title="归属业务"
          :visible="settings.checked.includes('bk_biz_id')"
          :min-width="150"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_biz_id')">
              <Select
                v-model="row.bk_biz_id"
                auto-focus
                filterable
                :disabled="true"
                @change="clearError(rowIndex, 'bk_biz_id')"
                @toggle="
                  (val) =>
                    !val &&
                    handleFieldBlur(rowIndex, 'bk_biz_id', row.bk_biz_id)
                "
              >
                <Select.Option
                  v-for="item in businessList"
                  :key="item.bk_biz_id"
                  :name="item.bk_biz_name"
                  :id="item.bk_biz_id"
                >
                  [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
                </Select.Option>
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 拓扑属性 -->
      <VxeColgroup title="拓扑属性" align="center" v-if="isReinstall">
        <VxeColumn
          field="bk_networkarea_name"
          title="管控区域"
          :visible="settings.checked.includes('bk_networkarea_name')"
          :min-width="150"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkarea_name')">
              <Input
                v-model.trim="row.bk_networkarea_name"
                :disabled="true"
                @change="
                  (val) => {
                    handleChangeIPv4(val, row, rowIndex);
                    clearError(rowIndex, 'bk_networkarea_name');
                  }
                "
                @blur="
                  handleFieldBlur(
                    rowIndex,
                    'bk_networkarea_name',
                    row.bk_networkarea_name
                  )
                "
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_networkunit_id"
          title="管控单元"
          :min-width="150"
          :visible="settings.checked.includes('bk_networkunit_id')"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <Select
                v-model="row.bk_networkunit_id"
                auto-focus
                filterable
                @change="
                  (val) => {
                    clearError(rowIndex, 'bk_networkunit_id');
                    handleNetworkUnitChange(val, row, rowIndex);
                  }
                "
                @toggle="
                  (val) =>
                    !val &&
                    handleFieldBlur(rowIndex, 'bk_networkunit_id', row.bk_networkunit_id)
                "
              >
                <Select.Option
                  v-for="option in getNetworkUnitsByAreaId(row.bk_networkarea_id)"
                  :key="option.bk_networkunit_id"
                  :id="String(option.bk_networkunit_id)"
                  :name="option.bk_networkunit_name"
                >
                  [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
                </Select.Option>
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 主机 IP -->
      <VxeColgroup align="center">
        <template #header>
          <span class="mr-[5px]">主机 IP</span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
        </template>
        <VxeColumn
          field="bk_host_innerip"
          title="内网 IPv4"
          :visible="settings.checked.includes('bk_host_innerip')"
          :min-width="150"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip')">
              <Input
                v-model.trim="row.bk_host_innerip"
                :disabled="isReinstall"
                @change="
                  (val) => {
                    handleChangeIPv4(val, row, rowIndex);
                    clearError(rowIndex, 'bk_host_innerip');
                  }
                "
                @blur="
                  handleFieldBlur(
                    rowIndex,
                    'bk_host_innerip',
                    row.bk_host_innerip
                  )
                "
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_host_innerip_v6"
          title="内网 IPv6"
          :min-width="150"
          :visible="settings.checked.includes('bk_host_innerip_v6')"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip_v6')">
              <Input
                :disabled="isReinstall"
                v-model.trim="row.bk_host_innerip_v6"
                @change="clearError(rowIndex, 'bk_host_innerip_v6')"
                @blur="
                  handleFieldBlur(
                    rowIndex,
                    'bk_host_innerip_v6',
                    row.bk_host_innerip_v6
                  )
                "
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 主机属性 -->
      <VxeColgroup title="主机属性" align="center">
        <VxeColumn
          field="os_type"
          title="操作系统"
          :min-width="120"
          :visible="settings.checked.includes('os_type')"
          v-if="releaseType !== 'proxy' || isReinstall"
        >
          <template #header>
            <span class="mr-[5px]">操作系统</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="'批量编辑操作系统'"
              type="select"
              :options="datasourceList"
              @confirm="(value) => handleBatchEdit('os_type', value)"
            >
            </BatchEdit>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'os_type')">
              <Select
                v-model="row.os_type"
                auto-focus
                @change="
                  (val) => {
                    handleChangeOsType(val, row, rowIndex);
                    clearError(rowIndex, 'os_type');
                  }
                "
                @toggle="
                  (val) =>
                    !val && handleFieldBlur(rowIndex, 'os_type', row.os_type)
                "
              >
                <Select.Option
                  v-for="option in datasourceList"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                >
                </Select.Option>
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="export_ip"
          :min-width="150"
          :visible="settings.checked.includes('export_ip')"
          v-if="releaseType === 'proxy'"
        >
          <template #header>
            <span class="mr-[5px]">出口IP</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'export_ip')">
              <Input
                v-model.trim="row.export_ip"
                @change="clearError(rowIndex, 'export_ip')"
                @blur="handleFieldBlur(rowIndex, 'export_ip', row.export_ip)"
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="advertise_ip"
          title="服务IP"
          :min-width="150"
          v-if="releaseType === 'proxy'"
          :visible="settings.checked.includes('advertise_ip')"
        >
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'advertise_ip')">
              <Input
                v-model.trim="row.advertise_ip"
                @change="clearError(rowIndex, 'advertise_ip')"
                @blur="
                  handleFieldBlur(rowIndex, 'advertise_ip', row.advertise_ip)
                "
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 登录信息 -->
      <VxeColgroup title="登录信息" align="center" v-if="type !== 'manual'">
        <VxeColumn
          field="login_ip"
          title="登录 IP"
          :min-width="150"
          :visible="settings.checked.includes('login_ip')"
        >
          <template #header>
            <span class="mr-[5px]">登录 IP</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_ip')">
              <Input
                v-model.trim="row.login_ip"
                @change="clearError(rowIndex, 'login_ip')"
                @blur="handleFieldBlur(rowIndex, 'login_ip', row.login_ip)"
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_port"
          title="登录端口"
          :min-width="120"
          :visible="settings.checked.includes('login_port')"
          v-if="releaseType !== 'proxy' || isReinstall"
        >
          <template #header>
            <span class="mr-[5px]">登录端口</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="'批量编辑登录端口'"
              type="input"
              @confirm="(value) => handleBatchEdit('login_port', value)"
            >
            </BatchEdit>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_port')">
              <Input
                v-model.trim="row.login_port"
                @change="() => {
                  clearError(rowIndex, 'login_port')
                  console.log(row.login_port)
                }"
                @blur="handleFieldBlur(rowIndex, 'login_port', row.login_port)"
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_user"
          title="登录账号"
          :min-width="150"
          :visible="settings.checked.includes('login_user')"
          v-if="releaseType !== 'proxy' || isReinstall"
        >
          <template #header>
            <span class="mr-[5px]">登录账号</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="'批量编辑登录账号'"
              type="input"
              @confirm="(value) => handleBatchEdit('login_user', value)"
            >
            </BatchEdit>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_user')">
              <Input
                v-model.trim="row.login_user"
                @change="clearError(rowIndex, 'login_user')"
                @blur="handleFieldBlur(rowIndex, 'login_user', row.login_user)"
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_mode"
          :min-width="120"
          :visible="settings.checked.includes('login_mode')"
        >
          <template #header>
            <span class="mr-[5px]">认证方式</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="'批量编辑认证方式'"
              type="select"
              :options="authenticationTypes"
              @confirm="(value) => handleBatchEdit('login_mode', value)"
            >
            </BatchEdit>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_mode')">
              <Select
                v-model="row.login_mode"
                auto-focus
                @change="
                  (val) => {
                    handleChangeMode(val, row, rowIndex);
                    clearError(rowIndex, 'login_mode');
                  }
                "
                @toggle="
                  (val) =>
                    !val &&
                    handleFieldBlur(rowIndex, 'login_mode', row.login_mode)
                "
              >
                <Select.Option
                  v-for="option in authenticationTypes"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                >
                </Select.Option>
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="credit"
          :min-width="170"
          :visible="settings.checked.includes('credit')"
        >
          <template #header>
            <div class="flex">
              <span class="mr-[5px]">密码 / 密钥</span>
              <span class="mx-[3px] text-[#FF5656]">*</span>
              <BatchEdit
                :title="'批量编辑密码/密钥'"
                type="credit"
                @confirm="(value) => handleBatchEdit('credit', value)"
              >
              </BatchEdit>
            </div>
          </template>
          <template #default="{ row, rowIndex }">
            <Input
              v-if="row.login_mode === 'password_vault'"
              :value="'自动拉取'"
              disabled
            ></Input>
            <ValidateCell v-else :error="getError(rowIndex, 'credit')">
              <Upload
                ref="uploader"
                type="formdata"
                v-if="row.login_mode === 'keyfile'"
                :url="url"
                :size="100"
                :multiple="false"
                :limit="1"
                theme="button"
                :before-upload="(val) => handleBeforeUpload(val, row)"
                :custom-request="() => {}"
                @change="clearError(rowIndex, 'credit')"
              ></Upload>
              <Input
                v-else
                v-model.trim="row.credit"
                :placeholder="
                  row.login_credit_valid ? '密码有效，点击修改' : '请输入密码'
                "
                type="password"
                @change="clearError(rowIndex, 'credit')"
                @blur="handleFieldBlur(rowIndex, 'credit', row.credit)"
              ></Input>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 开启的服务 -->
      <VxeColgroup
        title="开启的服务"
        align="center"
        v-if="releaseType === 'proxy'"
      >
        <VxeColumn
          :min-width="90"
          field="dedicated_installer"
          title="安装跳板"
          :visible="settings.checked.includes('dedicated_installer')"
        >
          <template #default="{ row }">
            <Switcher
              theme="primary"
              v-model="row.dedicated_installer"
            ></Switcher>
          </template>
        </VxeColumn>
        <VxeColumn
          :min-width="90"
          field="cluster_tunnel"
          title="Agent控制"
          :visible="settings.checked.includes('cluster_tunnel')"
        >
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.cluster_tunnel"></Switcher>
          </template>
        </VxeColumn>
        <VxeColumn
          :min-width="90"
          field="file_tunnel"
          title="文件传输"
          :visible="settings.checked.includes('file_tunnel')"
        >
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.file_tunnel"></Switcher>
          </template>
        </VxeColumn>
        <VxeColumn
          :min-width="90"
          field="data_tunnel"
          title="数据上报"
          :visible="settings.checked.includes('data_tunnel')"
        >
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.data_tunnel"></Switcher>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <VxeColgroup>
        <template #header>
          <Button text style="margin-right: 8px">
            <Settings
              ref="settingRef"
              :settings="settings"
              @setting-change="settingChange"
            ></Settings>
          </Button>
          <Button text v-if="isFullscreen" @click="switchFullScreen">
            <i class="nodeman-icon nc-icon-un-full-screen"></i>
          </Button>
          <Button text v-else @click="switchFullScreen">
            <i class="nodeman-icon nc-icon-full-screen"></i>
          </Button>
        </template>
        <VxeColumn :min-width="80" field="action" title="操作">
          <template #default="{ rowIndex }">
            <Button
              :disabled="isReinstall"
              text
              @click="handleAddRow(rowIndex)">
              <i class="nodeman-icon nc-plus"></i>
            </Button>
            <Button
              text
              :disabled="isReinstall"
              @click="handleDelRow(rowIndex)"
              style="margin-left: 8px"
            ><i class="nodeman-icon nc-minus"></i
            ></Button>
          </template>
        </VxeColumn>
      </VxeColgroup>
      <template #empty>
        <slot></slot>
      </template>
    </VxeTable>
  </div>
</template>

<script lang="ts" setup>
import { Button, Input, Message, Select, Switcher, Upload } from 'bkui-vue';
import { cloneDeep, groupBy } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

// 引入轻量级组件 ValidateCell
import ValidateCell from './validateCell.vue';

import type { TopoHostDistinctRespData } from '@/@types/topo.d';
import { TopoService } from '@/api/modules/topo';
import { VALIDATE_REGEX } from '@/common/const';
import BatchEdit from '@/components/batch-edit.vue';
import useFullScreen from '@/composables/use-fullscreen';
import { useMainStore } from '@/stores/main';

type VxeComponentSizeType = 'small' | 'medium' | 'large';
interface IValidate {
  validator: Function | RegExp | string;
  message: string;
}
type ValidationRules = Record<string, IValidate[]>;

const tableData = defineModel<Array<ReturnType<typeof getInitData>>>('data');
const props = defineProps({
  data: { type: Array, default: () => [] as any[] },
  maxHeight: { type: Number, default: 300 },
  releaseType: { type: String, default: 'agent' },
  isReinstall: { type: Boolean, default: false },
  currentSettings: {
    type: Object,
    default: () => ({
      fields: [
        { title: '内网 IPv4', field: 'bk_host_innerip' },
        { title: '内网 IPv6', field: 'bk_host_innerip_v6' },
        { title: '操作系统', field: 'os_type' },
        { title: '登录 IP', field: 'login_ip' },
        { title: '登录端口', field: 'login_port' },
        { title: '登录账号', field: 'login_user' },
        { title: '认证方式', field: 'login_mode' },
        { title: '密码 / 密钥', field: 'credit' },
      ],
      checked: [
        'bk_host_innerip',
        'bk_host_innerip_v6',
        'os_type',
        'login_port',
        'login_ip',
        'login_user',
        'login_mode',
        'credit',
      ],
      disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
      size: 'medium' as VxeComponentSizeType,
    }),
  },
});
// const manualSetting = {
//   fields: [
//     { title: '内网 IPv4', field: 'bk_host_innerip' },
//     { title: '内网 IPv6', field: 'bk_host_innerip_v6' },
//     { title: '操作系统', field: 'os_type' },
//   ],
//   checked: ['bk_host_innerip', 'bk_host_innerip_v6', 'os_type'],
//   disabled: ['os_type'],
//   size: 'medium' as VxeComponentSizeType,
// };
// const autoSetting  = {
//   fields: [
//     { title: '内网 IPv4', field: 'bk_host_innerip' },
//     { title: '内网 IPv6', field: 'bk_host_innerip_v6' },
//     { title: '操作系统', field: 'os_type' },
//     { title: '登录 IP', field: 'login_ip' },
//     { title: '登录端口', field: 'login_port' },
//     { title: '登录账号', field: 'login_user' },
//     { title: '认证方式', field: 'login_mode' },
//     { title: '密码 / 密钥', field: 'credit' },
//   ],
//   checked: ['bk_host_innerip', 'bk_host_innerip_v6', 'os_type', 'login_port', 'login_ip', 'login_user', 'login_mode', 'credit'],
//   disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
//   size: 'medium' as VxeComponentSizeType,
// };

const initData = {
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  os_type: '',
  login_ip: '',
  login_port: '',
  login_user: '',
  login_mode: 'password',
  login_password: '',
  login_key_file: '',
  bk_addressing: 'static',
  bk_networkunit_id: '',
  bk_biz_id: '',
  bk_host_id: '',
  re_register: false,
  credit: '',
  export_ip: '',
  advertise_ip: '',
  dedicated_installer: true,
  cluster_tunnel: true,
  file_tunnel: true,
  data_tunnel: true,
  proxy_tags: [],
};

const rules: ValidationRules = {
  bk_host_innerip: [
    { validator: VALIDATE_REGEX.IPV4, message: '请输入正确的内网 IPv4' },
  ],
  bk_host_innerip_v6: [
    { validator: VALIDATE_REGEX.IPV6, message: '请输入正确的内网 IPv6' },
  ],
  os_type: [{ validator: (val: string) => val, message: '请输入操作系统' }],
  login_ip: [
    { validator: VALIDATE_REGEX.IPV4, message: '请输入正确的登录 IP' },
  ],
  login_port: [
    { validator: VALIDATE_REGEX.PORT, message: '请输入正确的登录端口' },
  ],
  login_user: [{ validator: (val: string) => val, message: '请输入登录账号' }],
  login_mode: [{ validator: (val: string) => val, message: '请输入认证方式' }],
  credit: [{ validator: (val: string) => val, message: '请输入密码 / 密钥' }],
};

const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const xTableRef = ref();
const mainStore = useMainStore();
const businessList = computed(() => mainStore.businessList);
const type = computed(() => mainStore.agentSetupType);

// --- 错误状态管理 ---
const errorMap = reactive<Record<number, Record<string, string>>>({});

const getError = (rowIndex: number, field: string) => errorMap[rowIndex]?.[field] || '';
const setError = (rowIndex: number, field: string, msg: string) => {
  if (!errorMap[rowIndex]) errorMap[rowIndex] = {};
  errorMap[rowIndex][field] = msg;
};
const clearError = (rowIndex: number, field: string) => {
  if (errorMap[rowIndex]) delete errorMap[rowIndex][field];
};
// 清除所有错误标记的方法
const clearAllErrors = () => {
  for (const key in errorMap) delete errorMap[key];
};
function getInitData() {
  return cloneDeep(initData);
}

// --- 错误位移算法 ---
const shiftErrors = (index: number, offset: number) => {
  const newMap: Record<number, Record<string, string>> = {};
  Object.keys(errorMap).forEach((keyStr) => {
    const k = Number(keyStr);
    if (offset === 1) {
      // 新增
      if (k <= index) newMap[k] = errorMap[k];
      else newMap[k + 1] = errorMap[k];
    } else if (offset === -1) {
      // 删除
      if (k < index) newMap[k] = errorMap[k];
      else if (k > index) newMap[k - 1] = errorMap[k];
    }
  });
  for (const k in errorMap) delete errorMap[k];
  Object.assign(errorMap, newMap);
};

const handleAddRow = (index: number) => {
  if (!Array.isArray(tableData.value)) return;
  tableData.value.splice(index + 1, 0, cloneDeep(initData));
  shiftErrors(index, 1);
};

const handleDelRow = (index: number) => {
  if (!Array.isArray(tableData.value)) return;
  if (tableData.value.length === 1) return Message({ theme: 'warning', message: '至少保留一行' });
  tableData.value.splice(index, 1);
  shiftErrors(index, -1);
};

const settings = reactive(cloneDeep(props.currentSettings));
const settingChange = (data: any) => {
  settings.checked = data.checked;
  settings.size = data.size;
};

const datasourceList = ref<{ id: string; name: string }[]>([]);
const authenticationTypes = ref([
  { id: 'password', name: '密码' },
  { id: 'keyfile', name: '密钥' },
  ...(window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH === 'true'
    ? [
      {
        id: 'password_vault',
        name: window.PROJECT_CONFIG.PASSWORD_VAULT_NAME,
      },
    ]
    : []),
]);
const hostDistinct = ref<TopoHostDistinctRespData | null>();

// --- 业务逻辑 ---
const handleChangeMode = (val: string, row: any, rowIndex: number) => {
  row.login_mode = val;
  row.credit = '';
  clearError(rowIndex, 'credit');
};
const handleChangeIPv4 = (val: string, row: any, rowIndex: number) => {
  if (new RegExp(VALIDATE_REGEX.IPV4).test(val)) {
    row.login_ip = val;
    handleFieldBlur(rowIndex, 'login_ip', val);
  };
};
const handleChangeOsType = (val: string, row: any, rowIndex: number) => {
  if (val === 'linux') {
    row.login_port = '36000';
    row.login_user = 'root';
    handleFieldBlur(rowIndex, 'login_port', '36000');
    handleFieldBlur(rowIndex, 'login_user', 'root');
  }
  if (val === 'windows') {
    row.login_user = 'administrator';
    handleFieldBlur(rowIndex, 'login_user', 'administrator');
  }
};

const getHostDistinct = async () => {
  const res = await TopoService.HostDistinct({}).catch(() => null);
  if (res) {
    hostDistinct.value = res;
    datasourceList.value = res.os_type.map(item => ({
      id: item,
      name: item,
    }));
  }
};

const handleBatchEdit = (field: string, value: any) => {
  tableData.value?.forEach((item: any, index: number) => {
    if (field === 'credit') {
      if (item.login_mode === 'password') item.credit = value.password;
      else {
        item.credit = value.key;
        item.file = value.file;
      }
    } else {
      item[field] = value;
    }

    handleFieldBlur(index, field, value);

    // 如果存在联动（例如修改 OS 会影响 Port），可以在这里加特判
    if (field === 'os_type') {
      // 复用之前的联动逻辑
      if (value === 'linux') {
        item.login_port = '36000';
        item.login_user = 'root';
        handleFieldBlur(index, 'login_port', '36000');
        handleFieldBlur(index, 'login_user', 'root');
      } else if (value === 'windows') {
        item.login_user = 'administrator';
        handleFieldBlur(index, 'login_user', 'administrator');
      }
    }
  });
};

const url = location.href;
const handleBeforeUpload = (file: File, row: any) => {
  row.file = file;
  const reader = new FileReader();
  reader.onload = (e) => {
    const res = e.target?.result as string;
    if (res) row.credit = res.split(',')[1];
  };
  reader.readAsDataURL(file);
  return true;
};

// --- 数据校验逻辑 ---
const validateItemData = (value: any, rulesArr: IValidate[]) => {
  if (!rulesArr || !rulesArr.length) return true;
  if (!value && value !== 0) return false;
  for (const rule of rulesArr) {
    const { validator } = rule;
    let isValid = true;
    if (typeof validator === 'function') isValid = validator(value);
    else if (validator instanceof RegExp) isValid = validator.test(value);
    else if (typeof validator === 'string') isValid = new RegExp(validator).test(value);
    if (!isValid) return false;
  }
  return true;
};

// 【新增】单个字段失焦校验
// 模拟旧组件的 validate('blur') 行为
// 【修正】单个字段失焦校验
const handleFieldBlur = (rowIndex: number, field: string, value: any) => {
  // 先假设校验通过，清除错误（或者在下面的逻辑分支中显式清除）
  // 推荐策略：只在发现错误时 setError，否则 clearError。

  // 1. 必填检查
  const requiredFields = [
    'bk_host_innerip',
    'os_type',
    'login_ip',
    'login_port',
    'login_user',
    'login_mode',
    'export_ip',
  ];
  if (props.isReinstall) requiredFields.push('bk_biz_id');

  if (requiredFields.includes(field) && !value && value !== 0) {
    setError(rowIndex, field, '必填项');
    return;
  }

  // 密码特殊必填处理
  if (field === 'credit') {
    const row = tableData.value![rowIndex];
    if (
      row.login_mode !== 'password_vault' &&
      !row.login_credit_valid &&
      !value
    ) {
      setError(rowIndex, field, '必填项');
      return;
    }
  }

  // 2. 规则校验
  const fieldRules = rules[field];
  if (fieldRules) {
    if (!validateItemData(value, fieldRules)) {
      setError(rowIndex, field, fieldRules[0].message);
    } else {
      // 【关键修复】校验通过，必须清除错误！
      clearError(rowIndex, field);
    }
  } else {
    // 没有规则且通过了必填检查 -> 清除错误
    clearError(rowIndex, field);
  }
};

// --- 全局校验 ---
const tableValidate = async () => {
  const data = tableData.value;
  if (!Array.isArray(data) || !data.length) return true;

  for (const key in errorMap) delete errorMap[key];

  let isValid = true;
  let firstErrorRowIndex = -1;

  for (let i = 0; i < data.length; i++) {
    const row = data[i];
    let rowValid = true;

    // 2.1 业务属性
    if (
      props.isReinstall
      && settings.checked.includes('bk_biz_id')
      && !row.bk_biz_id
    ) {
      setError(i, 'bk_biz_id', '必填项');
      rowValid = false;
    }
    // 2.2 IP
    if (settings.checked.includes('bk_host_innerip')) {
      if (!row.bk_host_innerip) {
        setError(i, 'bk_host_innerip', '必填项');
        rowValid = false;
      } else if (
        !validateItemData(row.bk_host_innerip, rules.bk_host_innerip)
      ) {
        setError(i, 'bk_host_innerip', rules.bk_host_innerip[0].message);
        rowValid = false;
      }
    }
    if (
      settings.checked.includes('bk_host_innerip_v6')
      && row.bk_host_innerip_v6
    ) {
      if (!validateItemData(row.bk_host_innerip_v6, rules.bk_host_innerip_v6)) {
        setError(i, 'bk_host_innerip_v6', rules.bk_host_innerip_v6[0].message);
        rowValid = false;
      }
    }
    // 2.3 OS
    if (
      (props.releaseType !== 'proxy' || props.isReinstall)
      && settings.checked.includes('os_type')
      && !row.os_type
    ) {
      setError(i, 'os_type', '必填项');
      rowValid = false;
    }
    // 2.4 Proxy IP
    if (props.releaseType === 'proxy') {
      if (settings.checked.includes('export_ip')) {
        if (!row.export_ip) {
          setError(i, 'export_ip', '必填项');
          rowValid = false;
        } else if (!validateItemData(row.export_ip, rules.login_ip)) {
          setError(i, 'export_ip', rules.login_ip[0].message);
          rowValid = false;
        }
      }
      if (settings.checked.includes('advertise_ip') && row.advertise_ip) {
        if (!validateItemData(row.advertise_ip, rules.login_ip)) {
          setError(i, 'advertise_ip', rules.login_ip[0].message);
          rowValid = false;
        }
      }
    }
    // 2.5 Login Info
    if (type.value !== 'manual') {
      if (settings.checked.includes('login_ip')) {
        if (!row.login_ip) {
          setError(i, 'login_ip', '必填项');
          rowValid = false;
        } else if (!validateItemData(row.login_ip, rules.login_ip)) {
          setError(i, 'login_ip', rules.login_ip[0].message);
          rowValid = false;
        }
      }
      if (props.releaseType !== 'proxy' || props.isReinstall) {
        if (settings.checked.includes('login_port')) {
          if (!row.login_port) {
            setError(i, 'login_port', '必填项');
            rowValid = false;
          } else if (!validateItemData(row.login_port, rules.login_port)) {
            setError(i, 'login_port', rules.login_port[0].message);
            rowValid = false;
          }
        }
        if (settings.checked.includes('login_user') && !row.login_user) {
          setError(i, 'login_user', '必填项');
          rowValid = false;
        }
      }
      if (settings.checked.includes('login_mode') && !row.login_mode) {
        setError(i, 'login_mode', '必填项');
        rowValid = false;
      }
      if (
        settings.checked.includes('credit')
        && row.login_mode !== 'password_vault'
      ) {
        if (!row.login_credit_valid && !row.credit) {
          setError(i, 'credit', '必填项');
          rowValid = false;
        }
      }
    }

    if (row.login_mode === 'password_vault' && window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH !== 'true') {
      setError(i, 'login_mode', '密码库功能未开启');
      rowValid = false;
    }

    if (!rowValid) {
      isValid = false;
      if (firstErrorRowIndex === -1) firstErrorRowIndex = i;
    }
  }

  // 3. 定位到错误行
  if (!isValid && firstErrorRowIndex !== -1 && xTableRef.value) {
    await xTableRef.value.scrollToRow(data[firstErrorRowIndex]);
  }

  return isValid;
};

// 管控单元下拉列表获取
const networkUnitList = ref<any[]>([]);
// 分组映射：{bk_networkarea_id: [网络单元对象数组]}
const networkUnitGroupMap = ref<Record<number, any[]>>({});

const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: tableData.value?.map((item: any) => Number(item.bk_networkarea_id)) || [],
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;

  // 使用Lodash的groupBy函数进行分组
  networkUnitGroupMap.value = groupBy(res.items, 'bk_networkarea_id');
};

// 根据网络区域ID获取对应的网络单元列表
const getNetworkUnitsByAreaId = (bkNetworkAreaId: number) => {
  return networkUnitGroupMap.value[bkNetworkAreaId] || [];
};

// 获取所有可用的网络区域ID列表
const getNetworkAreaIds = () => {
  return Object.keys(networkUnitGroupMap.value).map(id => Number(id));
};

// 处理管控单元变更，获取对应的名称
const handleNetworkUnitChange = (val: string, row: any, rowIndex: number) => {
  if (!val) return;

  // 在所有网络单元中查找对应的名称
  const networkUnit = networkUnitList.value.find(
    (unit: any) => String(unit.bk_networkunit_id) === val
  );

  if (networkUnit) {
    row.bk_networkunit_name = networkUnit.bk_networkunit_name;
  }
};

const settingRef = ref();
const showSetting = () => settingRef.value?.showSetting();

defineExpose({ tableValidate, showSetting });

// 监听mainStore的变化来更新settings
// watch(() => mainStore.agentSetupType, (newType: string) => {
//   if (newType === 'manual') {
//     settings.fields = manualSetting.fields;
//     settings.checked = manualSetting.checked;
//     settings.size = manualSetting.size;
//   } else {
//     settings.fields = autoSetting.fields;
//     settings.checked = autoSetting.checked;
//     settings.size = autoSetting.size;
//   }
// }, { immediate: true });

onMounted(async () => {
  await getHostDistinct();
  if (props.isReinstall) {
    await getNetworkUnitList();
  }
});

watch(() => type.value, () => {
  clearAllErrors();
}, { immediate: true });
</script>
<style lang="postcss" scoped>
::v-deep(.vxe-body--column) {
  height: 56px !important;
}
</style>
