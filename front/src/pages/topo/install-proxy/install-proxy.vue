<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.title')"
    width="1290">
    <div class="py-[20px] px-[40px]">
      <!-- 提示 -->
      <div class="flex items-center border-1 border-solid border-[#A3C5FD] bg-[#F0F5FF] h-[32px] w-[1200px] p-[8px]">
        <i class="nodeman-icon nc-tips text-[#3A84FF] mr-[9px]"></i>
        <span class="mr-[8px] text-[#4D4F56] text-[12px]">
          {{ $t('topoManager.installProxy.tips') }}
        </span>
        <Button text theme="primary" class="!text-[12px]">
          {{ $t('topoManager.installProxy.guide') }}
        </Button>
      </div>
      <!-- form -->
      <Form :model="form" class="mt-[24px]">
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.method')"
          property=""
          label-width="90"
          required>
          <SelectItemGroup
            :list="installMethodList"
            @change="handleChange">
          </SelectItemGroup>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.info')"
          property=""
          label-width="90"
          required>
          <InfoTable v-model:data="form.data" ref="infoTableRef"></InfoTable>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.password')"
          property=""
          label-width="90"
          required>
          <Radio.Group v-model="form.saveTime">
            <Radio.Button
              :label="$t('topoManager.installProxy.form.saveTime.oneDay', { x: 1 })">
            </Radio.Button>
            <Radio.Button
              :label="$t('topoManager.installProxy.form.saveTime.longTermSave')">
            </Radio.Button>
          </Radio.Group>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.os')"
          property=""
          label-width="90"
          required>
          <Select class="w-[488px]" v-model="form.os" disabled></Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.port')"
          property=""
          label-width="90"
          required>
          <Input class="w-[488px]" v-model="form.port" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.account')"
          property=""
          label-width="90"
          required>
          <Input class="w-[488px]" v-model="form.account" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.business')"
          property=""
          label-width="90"
          required>
          <Select class="w-[488px]" v-model="form.business"></Select>
        </Form.FormItem>

        <div class="flex mt-[32px] ml-[90px]">
          <Button theme="primary" class="mr-[8px]" @click="handleConfirm">
            {{ $t('action.install') }}
          </Button>
          <Button @click="handleClose">
            {{ $t('action.cancel') }}
          </Button>
        </div>
      </Form>
    </div>
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Form, Input, Radio, Select, Sideslider } from 'bkui-vue';
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import InfoTable from './components/info-table.vue';
import SelectItemGroup from './components/select-item-group.vue';

const isShow = defineModel<boolean>('isShow', { default: false });
const { t } = useI18n();

const form = reactive({
  method: '', // 安装方式
  data: [{
    ipv4: '',
    ipv6: '',
    os: '',
    login_ip: '',
    authentication: '',
    password: '',
    directory: '',
    speed_limit: '',
    zip: false,
  }], // 安装信息
  saveTime: '保存 1 天', // 密钥/密码
  os: 'Linux(64位)', // 操作系统
  port: '', // 登录端口
  account: '', // 登录账号
  business: '', // 归属业务
});

// 安装方式列表
const installMethodList = ref([
  {
    icon: 'nodeman-icon nc-monitor',
    title: t('topoManager.installProxy.installMethodList.remote.title'),
    content: t('topoManager.installProxy.installMethodList.remote.content'),
    value: 0,
  },
  {
    icon: 'nodeman-icon nc-icon-control-fill',
    title: t('topoManager.installProxy.installMethodList.excel.title'),
    content: t('topoManager.installProxy.installMethodList.excel.content'),
    value: 1,
  },
  {
    icon: 'nodeman-icon nc-key',
    title: t('topoManager.installProxy.installMethodList.manual.title'),
    content: t('topoManager.installProxy.installMethodList.manual.content'),
    value: 2,
  },
]);

const handleChange = (values: Array<string | number>) => {
  form.method = values[0] as string;
};

const handleClose = () => {
  isShow.value = false;
};

const infoTableRef = ref();
const handleConfirm = () => {
  console.log(form);
  infoTableRef.value.tableValidate();
};

</script>
