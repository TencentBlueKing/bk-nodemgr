<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="520"
    :title="$t('topoManager.workUnit.title.updateProxyInfo')"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #default>
      <Form ref="formRef" :model="formData" :rules="rules" class="m-[24px]" :label-width="90" form-type="vertical">
        <Form.FormItem :label="$t('topoManager.installProxy.table.ipv4')" property="bk_host_innerip" required>
          <Input v-model="formData.bk_host_innerip" disabled />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.table.ipv6')" property="bk_host_innerip_v6">
          <Input v-model="formData.bk_host_innerip_v6" disabled />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.table.export_ip')" property="export_ip" required>
          <Input v-model="formData.export_ip" />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.table.advertise_ip')" property="advertise_ip">
          <Input v-model="formData.advertise_ip" />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.table.loginIp')" property="login_ip" required>
          <Input v-model="formData.login_ip" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.table.authenticationMethod')"
          property="credit"
          :required="!formData.login_credit_valid">
          <div class="flex w-full gap-[8px]">
            <Select
              v-model="formData.login_mode"
              auto-focus
              @change="handleChangeMode"
              class="w-1/4"
            >
              <Select.Option
                v-for="option in authenticationTypes"
                :key="option.id"
                :id="option.id"
                :name="option.name"
              >
              </Select.Option>
            </Select>
            <Input
              v-if="formData.login_mode === 'password_vault'"
              :value="'自动拉取'"
              disabled
            ></Input>
            <Input
              v-else
              v-model="formData.credit"
              class="flex-1"
              :placeholder="formData.login_credit_valid
                ? t('components.installTable.creditValid') : t('components.installTable.inputPassword')"
              type="password" />
          </div>
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.form.port')" property="login_port">
          <Input v-model="formData.login_port" />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.form.account')" property="login_user">
          <Input v-model="formData.login_user" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.relayDownloadPort')"
          property="relay_download_port"
          label-width="110"
          required
        >
          <Input class="w-[488px]" v-model="formData.relay_download_port" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.relayCallbackPort')"
          property="relay_callback_port"
          label-width="110"
          required
        >
          <Input class="w-[488px]" v-model="formData.relay_callback_port" />
        </Form.FormItem>
        <div class="flex items-center">
          <div class="flex items-center w-1/4">
            <Checkbox v-model="formData.dedicated_installer" theme="primary"></Checkbox>
            <span class="text-[#63656e] text-[14px] ml-[6px]">
              {{ $t('topoManager.installProxy.table.dedicated_installer') }}
            </span>
          </div>
          <div class="flex items-center w-1/4">
            <Checkbox v-model="formData.cluster_tunnel" theme="primary"></Checkbox>
            <span class="ml-[6px] text-[#63656e] text-[14px]">
              {{ $t('topoManager.installProxy.table.cluster_tunnel') }}
            </span>
          </div>
          <div class="flex items-center w-1/4">
            <Checkbox v-model="formData.file_tunnel" theme="primary"></Checkbox>
            <span class="ml-[6px] text-[#63656e] text-[14px]">
              {{ $t('topoManager.installProxy.table.file_tunnel') }}
            </span>
          </div>
          <div class="flex items-center w-1/4">
            <Checkbox v-model="formData.data_tunnel" theme="primary"></Checkbox>
            <span class="ml-[6px] text-[#63656e] text-[14px]">
              {{ $t('topoManager.installProxy.table.data_tunnel') }}
            </span>
          </div>
        </div>
      </Form>
    </template>
    <template #footer>
      <div class="flex justify-start gap-[8px]">
        <Button theme="primary" @click="handleSave" :loading="loading">{{ $t('action.save') }}</Button>
        <Button @click="handleBeforeClose">{{ $t('action.cancel') }}</Button>
      </div>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import { Button, Checkbox, Form, InfoBox, Input, Message, Select, Sideslider, Switcher } from 'bkui-vue';
import { cloneDeep, isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { VALIDATE_REGEX } from '@/common/const';
import { scrollToFirstErrorByClassNames } from '@/common/util';

interface IValidate {
  validator: Function | RegExp | string;
  message: string;
}
type ValidationRules = Record<string, IValidate[]>;

const isShow = defineModel('isShow', { type: Boolean });
const props = defineProps({
  data: {
    type: Object as PropType<Host | null>,
    default: null,
  },
});
const emit = defineEmits('update');
const proxyTags = ['cluster_tunnel', 'data_tunnel', 'dedicated_installer', 'file_tunnel'];
const { t } = useI18n();
const rules: ValidationRules = {
  bk_host_innerip: [
    { validator: VALIDATE_REGEX.IPV4, message: t('validate.ipv4') },
  ],
  bk_host_innerip_v6: [
    { validator: VALIDATE_REGEX.IPV6, message: t('validate.ipv6') },
  ],
  os_type: [{ validator: (val: string) => val, message: t('validate.os') }],
  login_ip: [
    { validator: VALIDATE_REGEX.IPV4, message: t('validate.loginIp') },
  ],
  login_port: [
    { validator: VALIDATE_REGEX.PORT, message: t('validate.port') },
  ],
  login_user: [{ validator: (val: string) => val, message: t('validate.loginUser') }],
  login_mode: [{ validator: (val: string) => val, message: t('validate.authType') }],
  credit: [{ validator: (val: string) => val, message: t('validate.password') }],
};
const initData = {
  bk_host_id: '',
  login_ip: '',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  login_port: '',
  login_user: '',
  login_mode: 'password',
  login_password: '',
  login_key_file: '',
  credit: '',
  export_ip: '',
  advertise_ip: '',
  dedicated_installer: false,
  cluster_tunnel: false,
  file_tunnel: false,
  data_tunnel: false,
  proxy_tags: [] as string[],
  login_credit_valid: false,
  relay_download_port: '',
  relay_callback_port: '',
};
const isPasswordVaultEnabled = computed(() => window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH === 'true');

const authenticationTypes = ref([
  {
    id: 'password',
    name: t('authenticationType.password'),
  },
  {
    id: 'keyfile',
    name: t('authenticationType.keyfile'),
  },
  ...(isPasswordVaultEnabled.value
    ? [{ id: 'password_vault', name: window.PROJECT_CONFIG.PASSWORD_VAULT_NAME }]
    : []),
]);
const formData = reactive(cloneDeep(initData));

// 切换认证方式
const handleChangeMode = (newValue: string) => {
  formData.login_mode = newValue;
  formData.credit = newValue === 'password_vault' ? '自动拉取' : '';
};

const originData = ref<any>();
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  // 没有修改，直接关闭
  if (isEqual(formData, originData.value)) {
    resolve(true);
    isShow.value = false;
    return;
  }
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
  });
});
const loading = ref(false);
const formRef = ref(null);
const handleSave = async () => {
  const validRes = await formRef.value?.validate().catch(() => false);
  if (!validRes) {
    scrollToFirstErrorByClassNames();
    return;
  };
  loading.value = true;
  const modeMap = {
    password: 'login_password',
    key: 'login_key_file',
  };
  formData[modeMap[formData.login_mode]] = formData.credit;
  formData.proxy_tags = [];
  proxyTags.forEach((key) => {
    formData[key] && formData.proxy_tags.push(key);
  });
  const params = {
    bk_host_id: formData.bk_host_id,
    login_ip: formData.login_ip,
    login_user: formData.login_user,
    login_mode: formData.login_mode,
    login_password: formData.login_password,
    login_key_file: formData.login_key_file,
    proxy_tags: formData.proxy_tags,
    login_port: Number(formData.login_port),
    relay_download_port: Number(formData.relay_download_port),
    relay_callback_port: Number(formData.relay_callback_port),
    export_ip: formData.export_ip || '', // 或者提供一个合适默认值
    advertise_ip: formData.advertise_ip || '',
  };
  const res = await NodeProxyService.NodeProxyUpdate({ host: [params] }).catch(err => false);
  if (res === false) return;
  loading.value = false;
  isShow.value = false;
  Message({
    theme: 'success',
    message: t('message.success.edit'),
  });
  emit('update');
};
watch(() => isShow.value, () => {
  if (isShow.value && props.data) {
    Object.assign(formData, props.data);
    proxyTags.forEach((key) => {
      formData[key] = props.data?.proxy_tags.includes(key);
    });
    formData.credit = formData.login_mode === 'password_vault' ? '自动拉取' : '';
    originData.value = cloneDeep(formData);
  }
  if (!isShow.value) {
    // 重置表单数据
    formData.export_ip = '';
    formData.advertise_ip = '';
    formData.login_ip = '';
    formData.login_user = '';
    formData.login_mode = '';
    formData.login_password = '';
    formData.login_key_file = '';
    formData.proxy_tags = [];
    formData.login_port = '';
    formData.relay_download_port = '';
    formData.relay_callback_port = '';
  }
});
</script>
