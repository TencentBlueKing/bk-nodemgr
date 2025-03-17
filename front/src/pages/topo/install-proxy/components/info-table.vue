<template>
  <VxeTable :data="tableData" border class="bg-[#F0F1F5] border-[#DCDEE5]">
    <VxeColgroup align="center">
      <template #header>
        <span class="mr-[5px]">
          {{ $t('topoManager.installProxy.table.hostIp') }}
        </span>
        <span class="mx-[3px] text-[#FF5656]">*</span>
      </template>
      <VxeColumn
        :title="$t('topoManager.installProxy.table.ipv4')"
        field="ipv4"
        width="90">
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.ipv4"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
      <VxeColumn
        :title="$t('topoManager.installProxy.table.ipv6')"
        field="ipv6"
        width="90">
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.ipv6"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
    </VxeColgroup>
    <VxeColgroup :title="$t('topoManager.installProxy.table.hostProperties')" align="center">
      <VxeColumn
        field="os"
        width="120">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.os') }}
          </span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.os"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
    </VxeColgroup>
    <VxeColgroup :title="$t('topoManager.installProxy.table.loginInfo')" align="center">
      <VxeColumn
        field="login_ip"
        width="90">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.loginIp') }}
          </span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.login_ip"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
      <VxeColumn
        field="authentication"
        width="120">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.authenticationMethod') }}
          </span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.authentication"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
      <VxeColumn
        field="password"
        width="130">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.password') }}
          </span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.password"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
    </VxeColgroup>
    <VxeColgroup :title="$t('topoManager.installProxy.table.transmittedInfo')" align="center">
      <VxeColumn
        field="directory"
        width="140">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.tempFileDirectory') }}
          </span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.directory"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
      <VxeColumn
        field="speed_limit"
        :title="$t('topoManager.installProxy.table.speedLimit')"
        width="120">
        <template #default="{ row, $rowIndex, $columnIndex }">
          <ValidateInput
            v-model:value="row.speed_limit"
            required
            :ref="el => setInputRef($rowIndex, $columnIndex, el as InstanceType<typeof ValidateInput>)">
          </ValidateInput>
        </template>
      </VxeColumn>
      <VxeColumn field="zip" width="120">
        <template #header>
          <span class="mr-[5px]">
            {{ $t('topoManager.installProxy.table.dataCompression') }}
          </span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
          <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
        </template>
        <template #default="{ row }">
          <Switcher v-model="row.zip" theme="primary"></Switcher>
        </template>
      </VxeColumn>
    </VxeColgroup>
    <VxeColgroup title="" align="center">
      <VxeColumn :title="$t('table.action')">
        <template #default="{ row, $rowIndex }">
          <div class="flex justify-center">
            <i
              class="nodeman-icon nc-plus text-[16px] mr-[6px] cursor-pointer"
              @click="handleAddRow($rowIndex)"
            ></i>
            <i
              class="nodeman-icon nc-minus text-[16px] cursor-pointer"
              @click="handleDelRow($rowIndex)"
            ></i>
          </div>
        </template>
      </VxeColumn>
    </VxeColgroup>
  </VxeTable>
</template>

<script lang="ts" setup>
import { Message, Switcher } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import type { PropType } from 'vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import ValidateInput from './validate-input.vue';

const tableData = defineModel<Array<typeof initData>>('data', { default: [{
  ipv4: '',
  ipv6: '',
  os: '',
  login_ip: '',
  authentication: '',
  password: '',
  directory: '',
  speed_limit: '',
  zip: false,
}] });

const { t } = useI18n();
const initData = {
  ipv4: '',
  ipv6: '',
  os: '',
  login_ip: '',
  authentication: '',
  password: '',
  directory: '',
  speed_limit: '',
  zip: false,
};

const handleAddRow = (index: number) => {
  tableData.value.splice(index + 1, 0, cloneDeep(initData));
};

const handleDelRow = (index: number) => {
  if (tableData.value.length === 1) {
    Message({
      theme: 'warning',
      message: t('message.warn.keepAtLeast', { x: 1 }),
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

const clearTableValidate = () => {
  for (const row of inputRefs.value) {
    for (const col of row) {
      col.clearValidate();
    }
  }
};

const tableValidate = () => {
  const validate = [];
  for (const row of inputRefs.value) {
    for (const col of row) {
      validate.push(col.validateInput());
    }
  }
  return validate.every(item => item !== false);
};

defineExpose({
  tableValidate,
  clearTableValidate,
});

</script>
