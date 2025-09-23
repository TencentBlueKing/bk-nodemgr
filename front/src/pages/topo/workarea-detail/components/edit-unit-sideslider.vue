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
        <Form.FormItem :label="$t('topoManager.installProxy.table.authenticationMethod')" property="prove" required>
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
            <Input v-model="formData.prove" class="flex-1" type="password" />
          </div>
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.form.port')" property="login_port">
          <Input v-model="formData.login_port" />
        </Form.FormItem>
        <Form.FormItem :label="$t('topoManager.installProxy.form.account')" property="login_user">
          <Input v-model="formData.login_user" />
        </Form.FormItem>
        <div class="flex items-center">
          <Form.FormItem :label="$t('topoManager.installProxy.table.dedicated_installer')" class="w-1/2">
            <Switcher v-model="formData.dedicated_installer" theme="primary"></Switcher>
          </Form.FormItem>
          <Form.FormItem :label="$t('topoManager.installProxy.table.cluster_tunnel')">
            <Switcher v-model="formData.cluster_tunnel" theme="primary"></Switcher>
          </Form.FormItem>
        </div>
        <div class="flex items-center">
          <Form.FormItem :label="$t('topoManager.installProxy.table.file_tunnel')" class="w-1/2">
            <Switcher v-model="formData.file_tunnel" theme="primary"></Switcher>
          </Form.FormItem>
          <Form.FormItem :label="$t('topoManager.installProxy.table.data_tunnel')">
            <Switcher v-model="formData.data_tunnel" theme="primary"></Switcher>
          </Form.FormItem>
        </div>
      </Form>
    </template>
    <template #footer>
      <div class="flex justify-start gap-[8px]">
        <Button theme="primary" @click="handleSave" :loading="loading">保存</Button>
        <Button @click="handleBeforeClose">取消</Button>
      </div>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Message, Select, Sideslider, Switcher } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import type { PropType } from 'vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { VALIDATE_REGEX } from '@/common/const';
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
  prove: [{ validator: (val: string) => val, message: '请输入密码 / 密钥' }],
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
  prove: '',
  export_ip: '',
  advertise_ip: '',
  dedicated_installer: false,
  cluster_tunnel: false,
  file_tunnel: false,
  data_tunnel: false,
  proxy_tags: [] as string[],
};
const authenticationTypes = ref([
  {
    id: 'password',
    name: '密码',
  },
  {
    id: 'key',
    name: '密钥',
  },
]);
const formData = reactive(cloneDeep(initData));
// 切换认证方式
const handleChangeMode = (newValue: string) => {
  formData.login_mode = newValue;
  formData.prove = '';
};
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  InfoBox({
    title: '确认关闭?',
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
  if (!validRes) return;
  loading.value = true;
  const modeMap = {
    password: 'login_password',
    key: 'login_key_file',
  };
  formData[modeMap[formData.login_mode]] = formData.prove;
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
  }
});
</script>
