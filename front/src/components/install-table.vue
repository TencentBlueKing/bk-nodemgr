<template>
  <div ref="contentRef">
    <VxeTable
      :data="tableData"
      :size="settings.size"
      :border="true"
      round>
      <!-- 主机 IP -->
      <VxeColgroup align="center">
        <template #header>
          <span class="mr-[5px]">主机 IP</span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
        </template>
        <VxeColumn
          field="ipv4"
          title="内网 IPv4"
          :visible="settings.checked.includes('ipv4')"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.ipv4"
              :rules="rules.ipv4"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.ipv4"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="ipv6"
          title="内网 IPv6"
          :visible="settings.checked.includes('ipv6')"
        >
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.ipv6"
              :rules="rules.ipv6"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.ipv6"></Input>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 主机属性 -->
      <VxeColgroup title="主机属性" align="center">
        <VxeColumn
          field="os"
          title="操作系统"
          :visible="settings.checked.includes('os')"
        >
          <template #header>
            <span class="mr-[5px]">操作系统</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.os"
              :rules="rules.os"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Select
                v-model="row.os"
                :list="datasourceList"
                auto-focus
                filterable
                @select="handleSelect">
              </Select>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 登录信息 -->
      <VxeColgroup title="登录信息" align="center" v-if="type !== 'manual'">
        <VxeColumn
          field="login_ip"
          :visible="settings.checked.includes('login_ip')"
        >
          <template #header>
            <span class="mr-[5px]">登录 IP</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_ip"
              :rules="rules.login_ip"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.login_ip"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_port"
          title="登录端口"
          :visible="settings.checked.includes('login_port')"
        >
          <template #header>
            <span class="mr-[5px]">登录端口</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_port"
              :rules="rules.login_port"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.login_port"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_user"
          title="登录账号"
          :visible="settings.checked.includes('login_user')"
        >
          <template #header>
            <span class="mr-[5px]">登录账号</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.login_user" 
              :rules="rules.login_user"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.login_user"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="authentication"
          width="120"
          :visible="settings.checked.includes('authentication')"
        >
          <template #header>
            <span class="mr-[5px]">认证方式</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.authentication"
              :rules="rules.authentication"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.authentication"></Input>
            </Validate>
          </template>
        </VxeColumn>
        <VxeColumn
          field="password"
          width="130"
          :visible="settings.checked.includes('password')"
        >
          <template #header>
            <span class="mr-[5px]">密码 / 密钥</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Validate
              :value="row.password"
              :rules="rules.password"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
              <Input v-model="row.password"></Input>
            </Validate>
          </template>
        </VxeColumn>
      </VxeColgroup>
      <VxeColgroup>
        <template #header>
          <Button text style="margin-right: 8px;">
            <Settings :settings="settings" @setting-change="settingChange"></Settings>
          </Button>
          <Button text v-if="isFullscreen" @click="switchFullScreen"><i class="nodeman-icon nc-icon-un-full-screen"></i></Button>
          <Button text v-else @click="switchFullScreen"><i class="nodeman-icon nc-icon-full-screen"></i></Button>
        </template>
        <VxeColumn field="action" title="操作">
          <template #default="{ row, $rowIndex, $columnIndex }">
            <Button text @click="handleAddRow($rowIndex)"><i class="nodeman-icon nc-plus"></i></Button>
            <Button text @click="handleDelRow($rowIndex)" style="margin-left: 8px;"><i class="nodeman-icon nc-minus"></i></Button>
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
type VxeComponentSizeType = 'small' | 'medium' | 'large';
interface IValidate {
  validator: Function | RegExp | string;
  message: string;
}
type ValidationRules = Record<string, IValidate[]>;
import { VxeTable, VxeColumn, VxeColgroup } from '@blueking/vxe-table';
import { Input, Button, Message, Select } from 'bkui-vue';
import { computed, ref, reactive } from 'vue';
import { useMainStore } from '@/stores/main';
import { VALIDATE_REGEX } from '@/common/const';
import { cloneDeep, set } from 'lodash';
import ValidateInput from './validate-input.vue';
import useFullScreen from '@/composables/use-fullscreen';
import { TopoService } from '@/api/modules/topo';

const props = defineProps({
  data: {
    type: Array,
    default: () => [],
  }
});
const initData = {
  ipv4: '',
  ipv6: '',
  os: '',
  login_ip: '',
  login_port: '',
  login_user: '',
  authentication: '',
  password: '',
};
const rules: ValidationRules = {
  ipv4: [
    // { validator: (val: string) => val, message: '请输入内网 IPv4'},
    { validator: VALIDATE_REGEX.IPV4, message: '请输入正确的内网 IPv4'},
  ],
  ipv6: [
    // { validator: (val: string) => val, message: '请输入内网 IPv6'},
    { validator: VALIDATE_REGEX.IPV6, message: '请输入正确的内网 IPv6'},
  ],
  os: [
    { validator: (val: string) => val, message: '请输入操作系统'},
  ],
  login_ip: [ 
    // { validator: (val: string) => val, message: '请输入登录 IP'},
    { validator: VALIDATE_REGEX.IPV4, message: '请输入正确的登录 IP'},
  ],
  login_port: [
    // { validator: (val: number) => val, message: '请输入登录端口'},
    { validator: VALIDATE_REGEX.PORT, message: '请输入正确的登录端口'},
  ],
  login_user: [
    { validator: (val: string) => val, message: '请输入登录账号'},
  ],
  authentication: [
    { validator: (val: string) => val, message: '请输入认证方式'},
  ],
  password: [
    { validator: (val: string) => val, message: '请输入密码 / 密钥'},
  ],
}
// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const mainStore = useMainStore();
const type = computed(() => mainStore.agentSetupType);
const tableData = ref(cloneDeep(props.data));
const handleAddRow = (index: number) => {
  tableData.value.splice(index + 1, 0, cloneDeep(initData));
};
const settings = reactive({
  fields: [
    {field: 'ipv4', title: '内网 IPv4'},
    {field: 'ipv6', title: '内网 IPv6'},
    {field: 'os', title: '操作系统'},
    {field: 'login_ip', title: '登录 IP'},
    {field: 'login_port', title: '登录端口'},
    {field: 'login_user', title: '登录账号'},
    {field: 'authentication', title: '认证方式'},
    {field: 'password', title: '密码 / 密钥'}
  ],
  checked: ['ipv4', 'ipv6', 'os', 'login_ip', 'login_port', 'login_user', 'authentication', 'password'],
  disabled: ['os', 'login_port', 'login_user', 'authentication', 'password'],
  size: 'medium' as VxeComponentSizeType
});
const settingChange = (data: {checked: string[], size: VxeComponentSizeType}) => {
  settings.checked = data.checked;
  settings.size = data.size;
  
}
const handleDelRow = (index: number) => {
  if (tableData.value.length === 1) {
    Message({
      theme: 'warning',
      message: '至少保留一行',
    });
    return;
  }
  tableData.value.splice(index, 1);
  inputRefs.value.splice(index, 1);
};
const inputRefs = ref<any>([[]]);

const setInputRef = (
  rowIndex: number,
  columnIndex: number,
  el: InstanceType<typeof ValidateInput> | null,
) => {
  if (el) {
    if (!inputRefs.value[rowIndex]) {
      inputRefs.value[rowIndex] = [];
    }
    inputRefs.value[rowIndex][columnIndex] = el;
  }
};

const clearAllValidate = () => {
  for (const row of inputRefs.value) {
    for (const col of row) {
      col.clearValidate();
    }
  }
};

const validate = async () => {
  const validate = [];
  for (const row of inputRefs.value) {
    for (const col of row) {
      validate.push(col.validateInput());
    }
  }
  try {
    const results = await Promise.all(validate);
    return results.every(result => result !== false);
  } catch (error) {
    console.error('Validation failed at some input:', error);
    return false;
  }
}
defineExpose({ validate, clearAllValidate });
</script>