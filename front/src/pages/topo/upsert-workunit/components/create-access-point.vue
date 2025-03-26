<template>
  <Form :model="form" ref="formRef" :rules="rules">
    <!-- 接入点名称 -->
    <Form.FormItem
      :label="$t('topoManager.workUnit.form.accessPointName')"
      property="accesspoint_name"
      label-width="120">
      <Input v-model="form.accesspoint_name" class="w-[283px]" />
    </Form.FormItem>
    <!-- Cluster -->
    <Form.FormItem
      label="Cluster"
      property="cluster"
      label-width="120">
      <InputGroup v-model:values="form.endpoints.cluster" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
    <!-- File -->
    <Form.FormItem
      label="File"
      property="file"
      label-width="120">
      <InputGroup v-model:values="form.endpoints.file" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
    <!-- Data -->
    <Form.FormItem
      label="Data"
      property="data"
      label-width="120"
      class="mb-0">
      <InputGroup v-model:values="form.endpoints.data" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
  </Form>
</template>

<script lang="ts" setup>
import { Form, Input } from 'bkui-vue';
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import InputGroup from './input-group.vue';

const form = defineModel<Omit<AccessPoint, 'accesspoint_id' | 'tenant_id'>>('form', { required: true });

const { t } = useI18n();

const inputPlaceholder = t('topoManager.workUnit.form.input.placeholder');
const rules = reactive({
  accesspoint_name: [
    {
      required: true,
      trigger: 'blur',
    },
  ],
  cluster: [
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.cluster.every(item => item !== ''),
    },
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.cluster.every(item => isInputValid(item)),
      message: t('topoManager.workUnit.form.input.validate'),
    },
  ],
  file: [
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.file.every(item => item !== ''),
    },
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.file.every(item => isInputValid(item)),
      message: t('topoManager.workUnit.form.input.validate'),
    },
  ],
  data: [
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.data.every(item => item !== ''),
    },
    {
      trigger: 'blur',
      validator: () => form.value.endpoints.data.every(item => isInputValid(item)),
      message: t('topoManager.workUnit.form.input.validate'),
    },
  ],
});

const formRef = ref();
const validateForm = async () => {
  const validate = await formRef.value?.validate().catch(() => false);
  return validate;
};

const isInputValid = (value: string) => {
  if (!value) return false;
  if (!/^\d{1,3}(\.\d{1,3}){3}:\d+$/.test(value)) {
    return false;
  }

  const validateIP = (ip: string) => ip.split('.').every((segment) => {
    const num = parseInt(segment, 10);
    return num >= 0 && num <= 255 && segment === num.toString();
  });

  const validatePort = (port: string) => {
    const portNum = parseInt(port, 10);
    return portNum >= 0 && portNum <= 65535;
  };

  const [ip, port] = value.split(':');
  return validateIP(ip) && validatePort(port);
};

defineExpose({
  validateForm,
});

</script>
