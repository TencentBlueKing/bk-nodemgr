<template>
  <div ref="contentRef">
    <VxeTable :data="tableData" :size="settings.size" :border="true" round>
      <!-- 业务属性 -->
      <VxeColgroup title="业务属性" align="center" v-if="isReinstall">
        <VxeColumn
          field="bk_biz_id"
          title="归属业务"
          :visible="settings.checked.includes('bk_biz_id')"
          :width="150"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.bk_biz_id"
              :rules="rules.bk_biz_id"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Select
                v-model="row.bk_biz_id"
                auto-focus
                filterable
                placeholder="选择业务"
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
            </Validate>
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
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.bk_host_innerip"
              :rules="rules.bk_host_innerip"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input
                v-model.trim="row.bk_host_innerip"
                @change="(val) => handleChangeIPv4(val, row)"
              ></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_host_innerip_v6"
          title="内网 IPv6"
          :visible="settings.checked.includes('bk_host_innerip_v6')"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.bk_host_innerip_v6"
              :rules="rules.bk_host_innerip_v6"
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.bk_host_innerip_v6"></Input>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 主机属性 -->
      <VxeColgroup title="主机属性" align="center">
        <VxeColumn
          field="os_type"
          title="操作系统"
          :visible="settings.checked.includes('os_type')"
          v-if="realeaseType !== 'proxy' || isReinstall"
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
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.os_type"
              :rules="rules.os_type"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Select
                v-model="row.os_type"
                auto-focus
                @change="(val) => handleChangeOsType(val, row)"
              >
                <Select.Option
                  v-for="option in datasourceList"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                >
                </Select.Option>
              </Select>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="export_ip"
          :visible="settings.checked.includes('export_ip')"
          v-if="realeaseType === 'proxy'"
        >
          <template #header>
            <span class="mr-[5px]">出口IP</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.export_ip"
              :rules="rules.login_ip"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.export_ip"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="advertise_ip"
          :visible="settings.checked.includes('advertise_ip')"
          title="服务IP"
          v-if="realeaseType === 'proxy'"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.advertise_ip"
              :rules="rules.login_ip"
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.advertise_ip"></Input>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 登录信息 -->
      <VxeColgroup title="登录信息" align="center" v-if="type !== 'manual'">
        <VxeColumn
          field="login_ip"
          title="登录 IP"
          :visible="settings.checked.includes('login_ip')"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_ip"
              :rules="rules.login_ip"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.login_ip"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_port"
          title="登录端口"
          :visible="settings.checked.includes('login_port')"
          v-if="realeaseType !== 'proxy' || isReinstall"
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
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_port"
              :rules="rules.login_port"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.login_port"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_user"
          title="登录账号"
          :visible="settings.checked.includes('login_user')"
          v-if="realeaseType !== 'proxy' || isReinstall"
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
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_user"
              :rules="rules.login_user"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Input v-model.trim="row.login_user"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_mode"
          width="120"
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
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_mode"
              :rules="rules.login_mode"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <Select
                v-model="row.login_mode"
                auto-focus
                @change="(val) => handleChangeMode(val, row)"
              >
                <Select.Option
                  v-for="option in authenticationTypes"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                >
                </Select.Option>
              </Select>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="credit"
          min-width="80"
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
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Input
              v-if="row.login_mode === 'password_vault'"
              :value="'自动拉取'"
              disabled
            ></Input>
            <Validate
              v-else
              :value="row.credit"
              :rules="rules.credit"
              required
              :ref="(el) => setInputRef($rowIndex, $columnIndex, el)"
            >
              <div
                v-bk-tooltips="{
                  content: '密码有效，点击修改',
                  disabled:
                    !isReinstall || !row.credit || !row.login_credit_valid,
                }"
              >
                <Upload
                  ref="uploader"
                  type="formdata"
                  v-if="row.login_mode === 'keyfile'"
                  :url="url"
                  :size="100"
                  :multiple="false"
                  :limit="1"
                  theme="button"
                  :before-upload="(val) => handleBeforeUpload(val,row)"
                  :custom-request="() => {}"
                ></Upload>
                <Input
                  v-else
                  v-model.trim="row.credit"
                  type="password"
                ></Input>
              </div>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>
      <!-- 登录信息 -->
      <VxeColgroup
        title="开启的服务"
        align="center"
        v-if="realeaseType === 'proxy'"
      >
        <VxeColumn
          width="80"
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
          width="90"
          field="cluster_tunnel"
          title="Agent控制"
          :visible="settings.checked.includes('cluster_tunnel')"
        >
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.cluster_tunnel"></Switcher>
          </template>
        </VxeColumn>
        <VxeColumn
          width="80"
          field="file_tunnel"
          title="文件传输"
          :visible="settings.checked.includes('file_tunnel')"
        >
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.file_tunnel"></Switcher>
          </template>
        </VxeColumn>
        <VxeColumn
          width="80"
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
        <VxeColumn width="80" field="action" title="操作">
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Button text @click="handleAddRow($rowIndex)">
              <i class="nodeman-icon nc-plus"></i>
            </Button>
            <Button
              text
              @click="handleDelRow($rowIndex, rowid, Object.keys(row))"
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
import { cloneDeep, template } from 'lodash';
import { computed, onMounted, reactive, ref } from 'vue';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import Validate from './validate.vue';

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
  data: {
    type: Array,
    default: () => [] as any[],
  },
  realeaseType: {
    type: String,
    default: 'agent',
  },
  isReinstall: {
    type: Boolean,
    default: false,
  },
  currentSettings: {
    type: Object,
    default: {
      fields: [
        { field: 'bk_host_innerip', title: '内网 IPv4' },
        { field: 'bk_host_innerip_v6', title: '内网 IPv6' },
        { field: 'os_type', title: '操作系统' },
        { field: 'login_port', title: '登录端口' },
        { field: 'login_user', title: '登录账号' },
        { field: 'login_ip', title: '登录 IP' },
        { field: 'login_mode', title: '认证方式' },
        { field: 'credit', title: '密码 / 密钥' },
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
    },
  },
});
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
// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();

const mainStore = useMainStore();
const businessList = computed(() => mainStore.businessList);
const type = computed(() => mainStore.agentSetupType);
function getInitData() {
  return cloneDeep(initData);
}

const handleAddRow = (index: number) => {
  if (!(tableData.value instanceof Array)) return;
  tableData.value.splice(index + 1, 0, cloneDeep(initData));
};
const settings = reactive(cloneDeep(props.currentSettings));
const settingChange = (data: {
  checked: string[];
  size: VxeComponentSizeType;
}) => {
  settings.checked = data.checked;
  settings.size = data.size;
};
const datasourceList = ref<{ id: string; name: string }[]>([]);
const authenticationTypes = ref([
  {
    id: 'password',
    name: '密码',
  },
  {
    id: 'keyfile',
    name: '密钥',
  },
  ...(window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH === true
    ? [{ id: 'password_vault', name: window.PROJECT_CONFIG.PASSWORD_VAULT_NAME }]
    : []),
]);
const hostDistinct = ref<TopoHostDistinctRespData | null>();
// 切换认证方式
const handleChangeMode = (newValue: string, row: any) => {
  row.login_mode = newValue;
  row.credit = '';
};
// 登录ip默认回填内网ipv4的值
const handleChangeIPv4 = (val: string, row: any) => {
  if (new RegExp(VALIDATE_REGEX.IPV4).test(val)) {
    row.login_ip = val;
  }
};
// linux登录端口默认为36000
const handleChangeOsType = (val: string, row: any) => {
  if (val === 'linux') {
    row.login_port = '36000';
    row.login_user = 'root';
  }
  if (val === 'windows') {
    row.login_user = 'administrator';
  }
};
// 获取操作系统
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
const handleDelRow = (index: number, rowid: string, fields: string[]) => {
  if (!(tableData.value instanceof Array)) return;
  if (tableData.value.length === 1) {
    Message({
      theme: 'warning',
      message: '至少保留一行',
    });
    return;
  }
  tableData.value.splice(index, 1);
  for (const field of fields) {
    const colKey = `${rowid}-${field}`;
    inputRefs.value?.delete(colKey);
  }
};

// 表头批量操作
const handleBatchEdit = (field: string, value: string | number | Object) => {
  tableData.value?.forEach((item: any) => {
    if (field === 'credit') {
      if (item.login_mode === 'password') {
        item.credit = value.password;
      } else {
        item.credit = value.key;
        item.file = value.file;
      }
    } else {
      item[field] = value;
    }
  });
};

// 上传密钥
const url = location.href;
const handleBeforeUpload = (file: File, row: any) => {
  console.log(file)
  row.file = file;
  const reader = new FileReader();
  reader.onload = (event) => {
    const result = event.target?.result as string;
    if (result) {
      row.credit = result.split(',')[1];
    }
  };
  reader.readAsDataURL(file);
  return true;
};

const inputRefs = ref<Map<string, InstanceType<typeof Validate>>>(new Map());

const setInputRef = (
  rowid: string | number,
  filed: string,
  el: InstanceType<typeof Validate> | null,
) => {
  if (el) {
    const key = `${rowid}-${filed}`;
    inputRefs.value.set(key, el);
  }
};

const tableValidate = async () => {
  const refs = Array.from(inputRefs.value.values());
  const validate = [];
  for (const item of refs) {
    validate.push(item.validate('blur'));
  }
  const result = await Promise.all(validate);
  return result.every(item => item);
};

defineExpose({
  tableValidate,
  // clearTableValidate
});

onMounted(async () => {
  await getHostDistinct();
});
</script>
