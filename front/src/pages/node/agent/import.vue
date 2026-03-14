<template>
  <div class="setup pt-[24px] pb-[48px]">
    <setup-tip @show-setting="handleShowSetting"></setup-tip>
    <div class="m-[24px]">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.type')"
          required
        >
          <install-type currentNodeType="agent" @change="handleChange"></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <Loading :loading="loading">
            <install-table
              ref="installTableRef"
              v-model:data="formData.info"
              :is-reinstall="true"
              :current-settings="tableSetting"
              :max-height="640"
            ></install-table>
          </Loading>
        </Form.FormItem>
        <Form.FormItem>
          <Button
            text
            theme="primary"
            class="text-[14px]"
            @click="isShow = !isShow"
          >
            <span class="mr-[8.5px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
            <angle-double-down-line
              :class="['text-[14px]', { 'transform rotate-180': isShow }]"
            />
          </Button>
        </Form.FormItem>
        <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.version')" required v-if="isShow">
          <div class="w-[568px]">
            <Table
              :data="systemData"
              :border="true"
              width="568"
              empty-text="$t('platform.nodeMan.installAgentPage.noAvailableVersion')">
              <TableColumn
                field="os"
                :title="$t('platform.nodeMan.installAgentPage.osArch')"
                width="200"
              >
                <template #default="{ row }">
                  {{ row.os?.replace('_', '/') }}
                </template>
              </TableColumn>
              <TableColumn field="version" :title="$t('platform.nodeMan.installAgentPage.packageVersion')" width="368">
                <template #header>
                  <span class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.packageVersion') }}</span>
                  <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                  <Button text @click="handleBatchEditVersion">
                    <i class="nodeman-icon nc-bulk-edit cursor-pointer"></i>
                  </Button>
                </template>
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="(el) => setInputRef(row.os, el)"
                  >
                    <Input
                      :model-value="row.version"
                      :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.select')"
                      @click="handleChooseVersion(row)"
                    />
                  </Validate>
                </template>
              </TableColumn>
            </Table>
          </div>
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
      >
        <span>{{ $t("platform.nodeMan.installAgentPage.button.install") }}</span>
        <span
          class="mx-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
        >
          {{ formData.info.length }}
        </span>
      </Button>
      <Button class="w-[88px]" @click="handleCancel">{{ $t("action.cancel") }}</Button>
    </div>
    <preview
      v-model:is-show="previewData.isShow"
      :data="previewData.data"
      :is-manual="activeInstallType === 'manual'"
    ></preview>
    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :batch="isBatch"
      :release-type="'agent'"
      @confirm="handleConfirmVersion"
    ></choose-version-dialog>
  </div>
</template>
<script lang="ts" setup>
import { Button, Form, Input, Loading, Select, Upload } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, debounce  } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch  } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Preview from './preview.vue';

import type { AgentInstallInfo } from '@/@types/node_agent.d';
import { TopoService } from '@/api/modules/topo';
import { encryptionTool } from '@/common/crypto';
import { scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const { t } = useI18n();
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
  bk_networkarea_id: '',
  bk_networkarea_name: '',
  bk_networkunit_id: '',
  bk_networkunit_name: '',
  bk_biz_id: '',
  bk_host_id: '',
  re_register: false,
  credit: '',
};
const formData = reactive({
  type: '',
  bk_host_name: '',
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

// 系统版本
const systemData = ref([
  {
    os: 'linux_amd64',
    version: '',
  },
  {
    os: 'darwin_amd64',
    version: '',
  },
  {
    os: 'linux_arm64',
    version: '',
  },
  {
    os: 'windows_amd64',
    version: '',
  },
]);

const tableSetting = reactive({
  fields: [
    { title: t('platform.nodeMan.installAgentPage.business'), field: 'bk_biz_id' },
    { title: t('platform.nodeMan.bk_cloud_name'), field: 'bk_networkarea_name' },
    { title: t('platform.nodeMan.bk_cloud_unit'), field: 'bk_networkunit_id' },
    { title: t('platform.nodeMan.inner_ip'), field: 'bk_host_innerip' },
    { title: t('platform.nodeMan.inner_ipv6'), field: 'bk_host_innerip_v6' },
    { title: t('platform.nodeMan.os_type'), field: 'os_type' },
    { title: t('platform.nodeMan.installAgentPage.loginIp'), field: 'login_ip' },
    { title: t('platform.nodeMan.installAgentPage.loginPort'), field: 'login_port' },
    { title: t('platform.nodeMan.installAgentPage.loginUser'), field: 'login_user' },
    { title: t('platform.nodeMan.installAgentPage.loginMode'), field: 'login_mode' },
    { title: t('platform.nodeMan.installAgentPage.passwordKey'), field: 'credit' },
  ],
  checked: [
    'bk_biz_id',
    'bk_networkarea_name',
    'bk_networkunit_id',
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'login_port',
    'login_ip',
    'login_user',
    'login_mode',
    'credit',
  ],
  disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
  size: 'medium',
});

const normalizeNetworkUnitId = (id: unknown) => {
  if (id === '' || id === null || id === undefined) return '';
  return Number(id) === -1 ? '' : String(id);
};

const loading = ref(false);
// 显示侧边栏安装策略
const handleShowPanel = () => {
  showRightPanel.value = true;
};

// 切换安装方式
const handleChange = (value: string) => {
  formRef.value?.clearValidate();
};

const isShowDialog = ref(false);
const dialogData = ref([{
  os: '',
  version: '',
}]);
const isBatch = ref(false);
const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
};
// 批量选择版本
const handleBatchEditVersion = () => {
  isShowDialog.value = true;
  dialogData.value = systemData.value;
  isBatch.value = true;
};
const handleConfirmVersion = (data: any[]) => {
  systemData.value.forEach((sys: { version: string; os: string }) => {
    const find = data.find((item) => sys.os === `${item.os_type}_${item.cup_arch}`);
    if (find) {
      sys.version = find.version;
    }
  });
  isBatch.value = false;
};

const inputRefs = ref<Map<string, InstanceType<typeof Validate>>>(new Map());
const setInputRef = (
  os: string,
  el: InstanceType<typeof Validate> | null,
) => {
  if (el) {
    const key = os;
    inputRefs.value.set(key, el);
  }
};
const systemValidate = async () => {
  const refs = Array.from(inputRefs.value.values());
  const validate = [];
  for (const item of refs) {
    validate.push(item.validate('blur'));
  }
  const result = await Promise.all(validate);
  return result.every(item => item);
};

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
      keyfile: 'login_key_file',
    };
    previewData.data = cloneDeep(formData);
    previewData.data.info.forEach((item: any) => {
      // 获取对应的 key (login_password 或 login_key_file)
      const targetKey = modeMap[item.login_mode];

      if (item.credit) {
        // 密码用 V1 加密，密钥用 V2 加密
        const encryptedValue = item.login_mode === 'keyfile'
          ? encryptionTool.encryptV2Sync(item.credit)
          : encryptionTool.encryptV1Sync(item.credit);

        // 如果加密成功，使用密文；否则使用空字符串
        item[targetKey] = encryptedValue !== false ? encryptedValue : '';
      } else {
        // 如果没有输入值，直接赋值
        item[targetKey] = item.credit;
      }
      // Keep historical semantics: empty network unit should remain -1 instead of 0.
      item.bk_networkunit_id = item.bk_networkunit_id === ''
        ? -1
        : Number(item.bk_networkunit_id);
      delete item.credit;
    });
    if (isShow.value) {
      previewData.data.target_version = systemData.value
        .filter((item: any) => !!item.version)
        .map((item) => {
          const [type, cpu_arch] = item.os.split('_');
          const os_type = type;
          return {
            os_type,
            cpu_arch,
            version: item.version,
          };
        });
    }
  } else {
    scrollToFirstErrorByClassNames();
  }
};
// 显示表格设置
const handleShowSetting = () => {
  installTableRef.value?.showSetting();
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
  encryptionTool.initPublicKey();
  if (footerRef.value) {
    window.addEventListener('resize', debouncedCheck);
    checkIfAtBottom();
  }
  // 使用TopoService.HostList接口进行切片查询获取数据
  if (nodeManageStore.agentEditParams.isCrossPageSelection) {
    // 跨页全选模式：使用HostList接口分页获取所有数据
    const allHosts = [];
    const pageSize = 1000; // 每页大小
    let offset = 0;
    let hasMore = true;
    loading.value = true;

    while (hasMore) {
      const hostListData = await TopoService.HostList({
        page: { offset, limit: pageSize },
        only_count: false,
        exact_include_conditions: nodeManageStore.agentEditParams.queryParams.exact_include_conditions || {},
        exact_exclude_conditions: nodeManageStore.agentEditParams.queryParams.exact_exclude_conditions || {},
        fuzzy_include_conditions: nodeManageStore.agentEditParams.queryParams.fuzzy_include_conditions || {},
      }).catch(() => ({ total: 0, items: [] }));

      if (hostListData.items && hostListData.items.length > 0) {
        allHosts.push(...hostListData.items);
        offset += pageSize;

        // 如果返回的数据少于pageSize，说明没有更多数据了
        if (hostListData.items.length < pageSize) {
          hasMore = false;
        }
      } else {
        hasMore = false;
      }
    }

    formData.info = allHosts.map((host: any) => ({
      ...host.state,
      ...host.info,
      ...host,
      bk_networkunit_id: normalizeNetworkUnitId(host.info.bk_networkunit_id),
      bk_host_innerip: host.info.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6: host.info.bk_host_innerip_v6_list?.join(','),
    }));
    loading.value = false;
  } else {
    // 本页选择模式：使用原有数据
    formData.info = nodeManageStore.agentEditParams.tableData.map(({ info, state, ...rest }) => ({
      target_version: state?.node_version,
      ...rest,
      bk_networkunit_id: normalizeNetworkUnitId(info.bk_networkunit_id),
      bk_host_innerip: info.bk_host_innerip_list?.[0],
      bk_host_innerip_v6: info.bk_host_innerip_v6_list?.[0],
      login_ip: info?.login_ip,
    }));
  }
});
onUnmounted(() => {
  if (footerRef.value) {
    window.removeEventListener('resize', debouncedCheck);
  }
});
</script>
