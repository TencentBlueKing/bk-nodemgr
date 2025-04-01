<template>
  <VxeTable :data="tableData" border class="bg-[#F0F1F5] border-[#DCDEE5]">
    <!-- 主机 IP -->
    <CustomColGroup
      :title="$t('topoManager.installProxy.table.hostIp')"
      :required="true"
    >
      <template #default="{ required: GroupRequired }">
        <!-- 内网 IPv4 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.ipv4')"
          field="ipv4"
          :width="90"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
        </CustomCol>

        <!-- 内网 IPv6 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.ipv6')"
          field="ipv6"
          :width="90"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
        </CustomCol>
      </template>
    </CustomColGroup>

    <!-- 主机属性 -->
    <CustomColGroup :title="$t('topoManager.installProxy.table.hostProperties')">
      <template #default="{ required: GroupRequired }">
        <!-- 操作系统 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.os')"
          field="os"
          :width="120"
          :required="true"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
        </CustomCol>
      </template>
    </CustomColGroup>

    <!-- 登录信息 -->
    <CustomColGroup
      :title="$t('topoManager.installProxy.table.loginInfo')"
    >
      <template #default="{ required: GroupRequired }">
        <!-- 登录 IP -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.loginIp')"
          field="login_ip"
          :width="90"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
          <template #edit="{ field }">
            <BatchEdit
              :title="$t('topoManager.installProxy.batchEdit.loginIp')"
              type="input"
              @confirm="(value) => handleBatchEdit(field, value)"
            >
            </BatchEdit>
          </template>
        </CustomCol>

        <!-- 认证方式 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.authenticationMethod')"
          field="authentication"
          :width="120"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
        </CustomCol>

        <!-- 密码 / 密钥 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.password')"
          field="password"
          :required="true"
          :width="130"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
          <template #edit="{ field }">
            <BatchEdit
              :title="$t('topoManager.installProxy.batchEdit.password')"
              type="input"
              @confirm="(value) => handleBatchEdit(field, value)"
            >
            </BatchEdit>
          </template>
        </CustomCol>
      </template>
    </CustomColGroup>

    <!-- 传输信息 -->
    <CustomColGroup
      :title="$t('topoManager.installProxy.table.transmittedInfo')"
    >
      <template #default="{ required: GroupRequired }">
        <!-- 临时文件目录 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.tempFileDirectory')"
          field="directory"
          :required="true"
          :width="140"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
          <template #edit="{ field }">
            <BatchEdit
              :title="$t('topoManager.installProxy.batchEdit.directory')"
              type="input"
              @confirm="(value) => handleBatchEdit(field, value)"
            >
            </BatchEdit>
          </template>
        </CustomCol>

        <!-- 传输限速 M/s -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.speedLimit')"
          field="speed_limit"
          :width="120"
        >
          <template #default="{ row, field, required, rowid }">
            <ValidateInput
              v-model:value="row[field]"
              :ref="(el) => setInputRef(rowid, field, el as InstanceType<typeof ValidateInput>)"
              :required="GroupRequired || required">
            </ValidateInput>
          </template>
        </CustomCol>

        <!-- 数据压缩 -->
        <CustomCol
          :title="$t('topoManager.installProxy.table.dataCompression')"
          field="zip"
          :required="true"
          :width="120"
        >
          <template #default="{ row, field }">
            <Switcher v-model="row[field]" theme="primary"></Switcher>
          </template>
          <template #edit="{ field }">
            <BatchEdit
              :title="$t('topoManager.installProxy.batchEdit.zip')"
              type="switcher"
              @confirm="(value) => handleBatchEdit(field, value)"
            >
            </BatchEdit>
          </template>
        </CustomCol>
      </template>
    </CustomColGroup>
    <VxeColgroup title="" align="center">
      <VxeColumn :title="$t('table.action')">
        <template #default="{ $rowIndex, rowid, row }">
          <div class="flex justify-center">
            <i
              class="nodeman-icon nc-plus text-[16px] mr-[6px] cursor-pointer"
              @click="handleAddRow($rowIndex)">
            </i>
            <i
              class="nodeman-icon nc-minus text-[16px] cursor-pointer"
              @click="handleDelRow($rowIndex, rowid, Object.keys(row))">
            </i>
          </div>
        </template>
      </VxeColumn>
    </VxeColgroup>
  </VxeTable>
</template>

<script lang="ts" setup>
import { Message, Switcher } from 'bkui-vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import BatchEdit from './batch-edit.vue';
import CustomCol from './custom-col.vue';
import CustomColGroup from './custom-col-group.vue';
import ValidateInput from './validate-input.vue';

const tableData = defineModel<Array<ReturnType<typeof getInitData>>>('data');

const { t } = useI18n();
function getInitData() {
  return {
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
};

const handleAddRow = (index: number) => {
  if (!(tableData.value instanceof Array)) return;

  tableData.value.splice(index + 1, 0, getInitData());
};

const handleDelRow = (index: number, rowid: string, fields: string[]) => {
  if (!(tableData.value instanceof Array)) return;

  if (tableData.value.length === 1) {
    Message({
      theme: 'warning',
      message: t('message.warn.keepAtLeast', { x: 1 }),
    });
    return;
  }
  tableData.value.splice(index, 1);

  for (const field of fields) {
    const colKey = `${rowid}-${field}`;
    inputRefs.value?.delete(colKey);
  }
};

const handleBatchEdit = (field: string, value: string | number) => {
  for (const data of tableData.value) {
    data[field] = value;
  }
};

const inputRefs = ref<Map<string, InstanceType<typeof ValidateInput>>>(new Map());

const setInputRef = (
  rowid: string,
  filed: string,
  el: InstanceType<typeof ValidateInput> | null,
) => {
  if (el) {
    const key = `${rowid}-${filed}`;
    inputRefs.value.set(key, el);
  }
};

const clearTableValidate = () => Array
  .from(inputRefs.value.values())
  .forEach(col => col.clearValidate());

const tableValidate = () => {
  const refs = Array.from(inputRefs.value.values());
  const validate = [];
  for (const item of refs) {
    validate.push(item.validateInput());
  }
  return validate.every(item => item);
};

defineExpose({
  tableValidate,
  clearTableValidate,
});

</script>
