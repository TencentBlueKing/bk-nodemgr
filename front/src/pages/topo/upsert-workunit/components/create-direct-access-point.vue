<template>
  <div>
    <!-- cluster -->
    <Form.FormItem
      label="cluster"
      property="cluster"
      :rules="clusterRule"
      label-width="120">
      <InputGroup v-model:values="data.cluster" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
    <!-- file -->
    <Form.FormItem
      label="file"
      property="file"
      :rules="fileRule"
      label-width="120">
      <InputGroup v-model:values="data.file" :placeholder="$t('topoManager.workUnit.form.input.directFilePlaceholder')"></InputGroup>
    </Form.FormItem>
    <!-- data -->
    <Form.FormItem
      label="data"
      property="data"
      :rules="dataRule"
      label-width="120"
      class="mb-0">
      <InputGroup v-model:values="data.data" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
  </div>
</template>

<script lang="ts" setup>
import { Form, Input } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import InputGroup from './input-group.vue';

const data = defineModel<Endpoints>('data', {required: true });

const { t } = useI18n();

const inputPlaceholder = t('topoManager.workUnit.form.input.placeholder');

// cluster校验规则
const clusterRule = [{
  trigger: 'blur',
  validator: () => data.value.cluster.every((item: string) => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.cluster.every((item: string) => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

// file校验规则
const fileRule = [{
  trigger: 'blur',
  validator: () => data.value.file.every((item: string) => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.file.every((item: string) => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

// data校验规则
const dataRule = [{
  trigger: 'blur',
  validator: () => data.value.data.every((item: string) => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.data.every((item: string) => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

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

</script>
