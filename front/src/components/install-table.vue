<template>
    <VxeTable
      :data="tableData"
      border
      round>
      <!-- 主机 IP -->
      <VxeColgroup align="center">
        <template #header>
          <span class="mr-[5px]">主机 IP</span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
        </template>
        <VxeColumn
          field="ipv4"
          title="内网 IPv4">
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.ipv4"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
        <VxeColumn
          field="ipv6"
          title="内网 IPv6">
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.ipv6"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 主机属性 -->
      <VxeColgroup title="主机属性" align="center">
        <VxeColumn
          field="os"
          title="操作系统">
          <template #header>
            <span class="mr-[5px]">操作系统</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.os"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 登录信息 -->
      <VxeColgroup title="登录信息" align="center" v-if="type !== 'manual'">
        <VxeColumn
          field="login_ip">
          <template #header>
            <span class="mr-[5px]">登录 IP</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.login_ip"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_port"
          title="登录端口">
          <template #header>
            <span class="mr-[5px]">登录端口</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.login_port"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
        <VxeColumn
          field="login_user"
          title="登录账号">
          <template #header>
            <span class="mr-[5px]">登录账号</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.os"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
        <VxeColumn
          field="authentication"
          width="120">
          <template #header>
            <span class="mr-[5px]">认证方式</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.authentication"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
        <VxeColumn
          field="password"
          width="130">
          <template #header>
            <span class="mr-[5px]">密码 / 密钥</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
          </template>
          <template #default="{ row, $rowIndex, $columnIndex }">
            <ValidateInput
              v-model:value="row.password"
              required
              :ref="el => setInputRef($rowIndex, $columnIndex, el)">
            </ValidateInput>
          </template>
        </VxeColumn>
      </VxeColgroup>
      <VxeColgroup>
        <template #header>
          <Button text><i class="nodeman-icon nc-setting"></i></Button>
          <Button text style="margin-left: 8px;"><i class="nodeman-icon nc-icon-full-screen"></i></Button>
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
</template>

<script lang="ts" setup>
import { VxeTable, VxeColumn, VxeColgroup } from '@blueking/vxe-table';
import { Input, Button, Message } from 'bkui-vue';
import { computed, ref, reactive } from 'vue';
import { useMainStore } from '@/stores/main';
import { VALIDATE_REGEX } from '@/common/const';
import { cloneDeep } from 'lodash';
import ValidateInput from './validate-input.vue';

const props = defineProps({
  data: {
    type: Array,
    default: () => [],
  }
});

const mainStore = useMainStore();
const type = computed(() => mainStore.agentSetupType);
const tableData = ref(cloneDeep(props.data));
const handleAddRow = (index: number) => {
  tableData.value.splice(index + 1, 0, cloneDeep(initData));
};

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
  if (validate.every(item => item !== false)) {
    console.log('校验通过');
  } else {
    console.log('校验失败');
  }
}
defineExpose({ validate, clearAllValidate });
</script>