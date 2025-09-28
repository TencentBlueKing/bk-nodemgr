<template>
  <div class="setup pt-[24px] pb-[48px]">
    <setup-tip></setup-tip>
    <div class="m-[24px]">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.type')"
          required
        >
          <install-type :need-type-list="['setup', 'manual']"></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <install-table
            ref="installTableRef"
            v-model:data="formData.info"
            :is-reinstall="true"
          ></install-table>
        </Form.FormItem>
      </Form>
    </div>
    <div
      :class="[
        'h-[48px] w-full flex items-center pl-[174px]',
        { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom },
      ]"
      ref="footerRef"
    >
      <Button
        class="w-[100px] mr-[8px]"
        theme="primary"
        @click="handlePreview"
      >{{ $t("platform.nodeMan.installAgentPage.button.install") }}
      </Button>
      <Button class="w-[88px]" @click="handleCancel">{{ $t("action.cancel") }}</Button>
    </div>
    <preview
      v-model:is-show="previewData.isShow"
      :data="previewData.data"
    ></preview>
  </div>
</template>
<script lang="ts" setup>
import { Button, Form, Input, Message, Select, Upload } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, debounce  } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch  } from 'vue';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Preview from './preview.vue';

import type { AgentInstallInfo } from '@/@types/node_agent.d';
import { TopoService } from '@/api/modules/topo';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const showRightPanel = ref(false);
const initData = {
  login_credit_valid: false,
  bk_addressing: 'static',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  os_type: '',
  login_ip: '',
  login_port: '',
  login_user: '',
  login_mode: 'password',
  login_password: '',
  login_key_file: '',
  bk_networkunit_id: '',
  bk_biz_id: '',
  bk_host_id: '',
  re_register: false,
  prove: '',
};
const formData = reactive({
  type: '',
  business: '',
  cloud: '',
  cloud_unit: '',
  info: [cloneDeep(initData)] as AgentInstallInfo[],
});
const previewData = reactive({
  isShow: false,
  data: null,
});
const rules = {};
const isShow = ref(false);
const isAtBottom = ref(false);
// 安装方式
const activeInstallType = computed(() => mainStore.agentSetupType);

// 显示侧边栏安装策略
const handleShowPanel = () => {
  showRightPanel.value = true;
};
// 显示表格设置
const handleShowSetting = () => {};
const handleCancel = () => {
  router.push({ name: 'agent' });
};
const formRef = ref(null);
const installTableRef = ref(null);
const handlePreview = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isShow.value ? systemValidate() : true,
  ]);
  // 合并多重Promise
  if (Array.isArray(result[2])) {
    result[2] = result[2].every(item => item);
  }
  if (result.every(item => item)) {
    previewData.isShow = true;
    const modeMap = {
      password: 'login_password',
      key: 'login_key_file',
    };
    formData.info.forEach((item) => {
      item[modeMap[item.login_mode]] = item.prove;
    });
    previewData.data = { ...formData };
  }
};

const footerRef = ref<Element | null>(null);
const checkIfAtBottom = () => {
  if (footerRef.value) {
    let { bottom } = footerRef.value.getBoundingClientRect();
    if (isShow.value) {
      bottom += 224;
    } else {
      bottom -= 224;
    }
    isAtBottom.value = bottom >= window.innerHeight;
  }
};
const debouncedCheck = debounce(checkIfAtBottom, 100);
watch(
  () => isShow.value,
  (val: boolean) => {
    checkIfAtBottom();
  },
);
onMounted(async () => {
  if (footerRef.value) {
    window.addEventListener('resize', debouncedCheck);
    checkIfAtBottom();
  }
  formData.info = nodeManageStore.agentEditParams.tableData.map((item: Host) => ({
    ...item,
    target_version: item.state.node_version,
  }));
});
onUnmounted(() => {
  if (footerRef.value) {
    window.removeEventListener('resize', debouncedCheck);
  }
});
</script>
