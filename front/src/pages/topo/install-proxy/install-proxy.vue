<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.title')"
    width="1200"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <div class="py-[20px] px-[40px]">
      <!-- 安装策略提示条 -->
      <div
        class="flex items-center min-h-[32px] bg-[#F0F8FF] border border-[#C5DAFF] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
      >
        <i class="nodeman-icon nc-tips text-[#3A84FF]"></i>
        <div class="flex-1 flex items-center text-[#4d4f56] text-[12px] leading-[20px]">
          <span class="mr-[5px]">{{ $t('topoManager.installProxy.tips') }}</span>
          <Button
            text
            theme="primary"
            @click="handleShowStrategy"
          >
            {{ $t('topoManager.installProxy.guide') }}
          </Button>
        </div>
      </div>
      <!-- form -->
      <Form ref="formRef" :model="form" :rules="formRules" class="mt-[24px]">
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.method')"
          property="method"
          label-width="110"
          required
        >
          <install-type currentNodeType="proxy" :needTypeList="['setup', 'manual', 'offline']" @change="handleChange"></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.info')"
          property=""
          label-width="110"
          required
        >
          <template #label>
            <div class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.info') }}</div>
            <Button text theme="primary" @click="handleExcelImport">
              {{ $t('platform.nodeMan.installAgentPage.excelImport') }}
            </Button>
          </template>
          <install-table
            ref="installTableRef"
            v-model:data="form.info"
            release-type="proxy"
            :current-settings="settings"
            :max-height="520"
          >
            <UploadExcel type="proxy" @upload="handleUpload" v-if="form.method === '1'"></UploadExcel>
          </install-table>
        </Form.FormItem>
        <Form.FormItem
          property="bk_biz_id"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('topoManager.installProxy.form.business') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.installBusinessTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <BizSelect
            select-class="w-[488px]"
            v-model="form.bk_biz_id"
            action="proxy_operate"
            :placeholder="t('installProxy.selectBusiness')"
          />
        </Form.FormItem>
        <Form.FormItem
          v-if="bk_networkunit_id === null"
          property="bk_networkarea_id"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.bk_cloud_name') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <AreaSelector class="w-[488px]" :multiple="false" @change="handleSingleChange" />
        </Form.FormItem>
        <Form.FormItem
          v-if="bk_networkunit_id === null"
          property="bk_networkunit_id"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.bk_cloud_unit') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudUnitTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <UnitSelector
            class="w-[488px]"
            v-model="form.bk_networkunit_id"
            :disabled="!form.bk_networkarea_id"
            :disable-direct="true"
            :direct-tip="$t('topoManager.installProxy.form.tip')"
            :options="areaUnitlist"
            @change="handleUnitSelectorChange"
          />
        </Form.FormItem>

        <Form.FormItem
          label-width="110">
          <Button
            text
            theme="primary"
            @click="isTargetShow = !isTargetShow"
          >
            <span class="mr-[8.5px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
            <angle-double-down-line
              :class="['text-[14px]', { 'transform rotate-180': isTargetShow }]"
            />
          </Button>
        </Form.FormItem>
        <Form.FormItem
          v-if="isTargetShow && form.method !== 'offline'"
          property="proxy_install_origin"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('installProxy.installSource') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('installProxy.installSourceTooltipDesc') }}</p>
                  <p class="mt-[8px]">{{ $t('installProxy.installSourceTooltipUpstream') }}</p>
                  <p class="mt-[8px]">{{ $t('installProxy.installSourceTooltipCurrent') }}</p>
                  <p class="mt-[8px]">{{ $t('installProxy.installSourceTooltipCustom') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <Cascader
            v-model="form.proxy_install_origin"
            :list="installOriginList"
            class="w-[488px]"
            trigger="click"
          ></Cascader>
        </Form.FormItem>
        <Form.FormItem
          v-if="isTargetShow"
          property="relay_callback_port"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('topoManager.installProxy.form.relayCallbackPort') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('installProxy.relayCallbackPortTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <Input class="w-[488px]" v-model="form.relay_callback_port" />
        </Form.FormItem>
        <Form.FormItem
          v-if="isTargetShow"
          property="relay_download_port"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('topoManager.installProxy.form.relayDownloadPort') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('installProxy.relayDownloadPortTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <Input class="w-[488px]" v-model="form.relay_download_port" />
        </Form.FormItem>
        <Form.FormItem label-width="110" required v-if="isTargetShow">
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('installProxy.proxyVersion') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('installProxy.proxyVersionTooltipDesc') }}</p>
                  <p class="mt-[8px]">{{ $t('installProxy.proxyVersionTooltipDefault') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <div class="w-[488px]">
            <Table :data="systemData" :border="true" :empty-text="t('installProxy.noAvailableVersion')">
              <TableColumn
                field="displayName"
                :title="t('installProxy.osArch')"
                width="200"
              ></TableColumn>
              <TableColumn field="version" :title="t('installProxy.version')" width="288">
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="(el) => setInputRef(row.os, el)"
                  >
                    <Input
                      :model-value="row.version"
                      :placeholder="t('installProxy.pleaseSelect')"
                      @click="handleChooseVersion(row)"
                    />
                  </Validate>
                </template>
              </TableColumn>
            </Table>
          </div>
        </Form.FormItem>
      </Form>
      <div class="flex mt-[32px] ml-[110px]">
        <!-- <Button
          v-if="excelImportData.length && form.info.length === 0"
          class="w-[100px]"
          theme="primary"
          @click="handleImport">
          {{ $t('platform.nodeMan.installAgentPage.excelImport') }}
        </Button> -->
        <Button
          theme="primary"
          class="mr-[8px] w-[120px]"
          :disabled="systemData.length === 0"
          v-bk-tooltips="{
            content: t('installProxy.noAvailableVersionTip'),
            disabled: systemData.length > 0
          }"
          @click="handleConfirm"
        >
          <span>
            {{ $t("topoManager.installProxy.button.install") }}
          </span>
          <span
            class="ml-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
          >
            {{ form.info.length }}
          </span>
        </Button>
        <!-- <Button
          v-if="excelImportData.length && form.method === '1' && form.info.length > 0"
          class="w-[88px]"
          @click="handleSetpBack">
          {{ '上一步' }}
        </Button> -->
        <Button @click="handleBeforeClose">
          {{ $t("action.cancel") }}
        </Button>
      </div>
    </div>
    <proxy-preview
      v-model:is-show="isShowPreview"
      :data="previewData"
      @close="isShow = false"
    ></proxy-preview>
    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :release-type="'proxy'"
      @confirm="handleComfirmVerion"
    ></choose-version-dialog>
    <Dialog
      :is-show="isShowExcelImport"
      :width="1048"
      :title="t('installProxy.excelImport')"
      @closed="handleExcelImportCancel">
      <UploadExcel type="proxy" ref="uploadExcelRef" @upload="handleUpload"></UploadExcel>
      <template #footer>
        <Button
          :disabled="!uploadExcelRef?.curFile?.data"
          theme="primary"
          class="mr-[8px]"
          @click="handleExcelImportConfirm">
          {{ $t("action.confirm") }}
        </Button>
        <Button class="" @click="handleExcelImportCancel">{{ $t("action.cancel") }}</Button>
      </template>
    </Dialog>
    <install-strategy-sideslider
      v-model:is-show="isShowStrategy"
      :table-data="form.info"
      :network-unit-id="props.bk_networkunit_id || form.bk_networkunit_id"
      :area-name="strategyAreaName"
    />
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Cascader, Dialog, Form, InfoBox, Input, Message, Popover, Select, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import type { PropType } from 'vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import InstallStrategySideslider from './components/install-strategy-sideslider.vue';
import ProxyPreview from './preview.vue';
import SelectItemGroup from './components/select-item-group.vue';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { encryptionTool } from '@/common/crypto';
import { PACKAGE_GENERATION } from '@/common/const';
import { getDefaultLoginMode, scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';
import BizSelect from '@/components/biz-select.vue';
import { useMainStore } from '@/stores/main';
import { useTopoStore } from '@/stores/topo';
import { useWorkareaStore } from '@/stores/workarea';

const isShow = defineModel<boolean>('isShow', { default: false });
const props = defineProps({
  bk_networkunit_id: {
    type: Number,
    default: null,
  },
  data: {
    type: Array as PropType<Host[]>,
    default: [],
  },
});
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const mainStore = useMainStore();
const topoStore = useTopoStore();
const workareaStore = useWorkareaStore();
const initData = {
  bk_host_id: '',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  os_type: 'linux',
  login_port: window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT,
  login_user: 'root',
  export_ip: '',
  advertise_ip: '',
  login_ip: '',
  login_mode: getDefaultLoginMode(),
  login_password: '',
  login_key_file: '',
  bk_addressing: 'static',
  dedicated_installer: true,
  install_pre_ordered_plugins: true,
  cluster_tunnel: true,
  file_tunnel: true,
  data_tunnel: true,
  proxy_tags: [] as string[],
  cpu_arch: '',
};
const form = reactive({
  method: 'setup', // 安装方式
  info: [cloneDeep(initData)], // 安装信息
  bk_biz_id: '', // 归属业务
  bk_networkarea_id: '', // 管控区域
  bk_networkarea_name: '',
  bk_networkunit_id: '', // 管控单元
  target_version: [] as TargetVersion[],
  proxy_install_origin: [],
  relay_download_port: '28303',
  relay_callback_port: '28302',
});
const settings = reactive({
  fields: [
    { field: 'bk_host_innerip', title: t('installProxy.innerIPv4') },
    { field: 'bk_host_innerip_v6', title: t('installProxy.innerIPv6') },
    { field: 'os_type', title: t('installProxy.os') },
    { field: 'bk_addressing', title: t('components.installTable.addressingMode') },
    { field: 'cpu_arch', title: t('components.installTable.cpuArch') },
    { field: 'login_port', title: t('installProxy.port') },
    { field: 'login_user', title: t('installProxy.account') },
    { field: 'export_ip', title: t('installProxy.exportIP') },
    { field: 'advertise_ip', title: t('installProxy.serviceIP') },
    { field: 'login_ip', title: t('installProxy.loginIP') },
    { field: 'login_mode', title: t('installProxy.authMethod') },
    { field: 'credit', title: t('installProxy.passwordKey') },
    { field: 'install_pre_ordered_plugins', title: t('components.installTable.installPreOrderedPlugins') },
    { field: 're_register', title: t('components.installTable.reRegisterAgentId') },
    { field: 'dedicated_installer', title: t('installProxy.installJump') },
    { field: 'cluster_tunnel', title: t('installProxy.agentControl') },
    { field: 'file_tunnel', title: t('installProxy.fileTransfer') },
    { field: 'data_tunnel', title: t('installProxy.dataReport') },
  ],
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'bk_addressing',
    'cpu_arch',
    'login_port',
    'login_user',
    'export_ip',
    'login_ip',
    'login_mode',
    'credit',
  ],
  disabled: ['os_type', 'cpu_arch', 'login_port', 'login_user', 'login_mode', 'credit'],
  size: 'medium',
});
const isTargetShow = ref(false);

// 端口校验规则
const portValidator = (value: string) => {
  if (!value || value.trim() === '') {
    return false;
  }
  const port = Number(value);
  if (Number.isNaN(port) || !Number.isInteger(port)) {
    return false;
  }
  if (port <= 0 || port > 65535) {
    return false;
  }
  return true;
};

const formRules = {
  relay_callback_port: [
    { required: true, message: t('validate.required'), trigger: 'blur' },
    { validator: portValidator, message: t('installProxy.portMustBeValidRange'), trigger: 'blur' },
  ],
  relay_download_port: [
    { required: true, message: t('validate.required'), trigger: 'blur' },
    { validator: portValidator, message: t('installProxy.portMustBeValidRange'), trigger: 'blur' },
  ],
};
const systemData = ref([
  {
    displayName: 'linux/amd64',
    os: 'linux_amd64',
    cpu_arch: 'amd64',
    os_type: 'linux',
    version: '',
  },
  {
    displayName: 'linux/arm64',
    os: 'linux_arm64',
    cpu_arch: 'arm64',
    os_type: 'linux',
    version: '',
  },
]);

const handleSingleChange = (id: string, rows: any[]) => {
  // 单选通常用于表单赋值
  form.bk_networkarea_id = id;
  form.bk_networkunit_id = '';
  form.bk_networkarea_name = rows[0].bk_networkarea_name;
};

const handleUnitSelectorChange = (_id: string | number, row: any) => {
  form.bk_networkunit_name = row?.bk_networkunit_name || '';
};

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: [],
    },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;
};

// eslint-disable-next-line max-len
const areaUnitlist = computed(() => networkUnitList.value.filter((item: NetworkUnit) => [Number(route.params.workarea), Number(form.bk_networkarea_id)].includes(item.bk_networkarea_id)));

// 查询指定单元是否有 proxy
const currentUnitHasProxy = ref(false);
const checkUnitHasProxy = async (unitId: number) => {
  if (!unitId) {
    currentUnitHasProxy.value = false;
    return;
  }
  try {
    const res = await TopoService.HostList({
      page: { offset: 0, limit: 1 },
      only_count: true,
      exact_include_conditions: {
        bk_networkunit_id: [unitId],
        node_role: ['proxy'],
        node_status: ['running'],
      },
      fuzzy_include_conditions: {},
    });
    currentUnitHasProxy.value = (res.total ?? 0) > 0;
  } catch {
    currentUnitHasProxy.value = false;
  }
};

// 根据当前单元是否有 proxy 设置安装源默认值（离线不展示安装源，仍刷新 currentUnitHasProxy 供默认 origin）
const setDefaultInstallOrigin = async () => {
  const unitId = props.bk_networkunit_id || Number(form.bk_networkunit_id);
  if (!unitId) return;
  await checkUnitHasProxy(unitId);
  if (form.method === 'offline') {
    return;
  }
  // eslint-disable-next-line max-len
  const unit = areaUnitlist.value.find((item: NetworkUnit) => [props.bk_networkunit_id, Number(form.bk_networkunit_id)].includes(item.bk_networkunit_id));
  const hasUpstream = unit?.links?.cluster?.bk_networkunit_id !== null && unit?.links?.cluster?.bk_networkunit_id !== undefined;
  if (currentUnitHasProxy.value) {
    form.proxy_install_origin = ['current'];
  } else if (hasUpstream) {
    form.proxy_install_origin = ['upstream'];
  } else {
    isTargetShow.value = true;
  }
};

// 离线安装无安装源 UI 时，与自动选择 current/upstream 一致
const getDefaultProxyInstallOriginUnitIdForNewInstall = (): number => {
  const unitId = props.bk_networkunit_id || Number(form.bk_networkunit_id);
  if (!Number.isFinite(unitId) || unitId <= 0) return 0;
  // eslint-disable-next-line max-len
  const unit = areaUnitlist.value.find((item: NetworkUnit) => [props.bk_networkunit_id, Number(form.bk_networkunit_id)].includes(item.bk_networkunit_id));
  const upstreamId = unit?.links?.cluster?.bk_networkunit_id;
  const hasUpstream = upstreamId !== null && upstreamId !== undefined;
  if (currentUnitHasProxy.value) return unitId;
  if (hasUpstream) return Number(upstreamId);
  return unitId;
};

// 安装源
const installOriginList = computed(() => {
  // eslint-disable-next-line max-len
  const unit = areaUnitlist.value.find((item: NetworkUnit) => [props.bk_networkunit_id, Number(form.bk_networkunit_id)].includes(item.bk_networkunit_id));
  let list;
  if (unit?.links?.cluster?.bk_networkunit_id !== null) {
    list = [
      {
        id: 'upstream',
        name: t('installProxy.upstreamUnit'),
        bk_networkunit_id: unit?.links?.cluster?.bk_networkunit_id,
      },
      {
        id: 'current',
        name: t('installProxy.currentUnit'),
        bk_networkunit_id: Number(props.bk_networkunit_id || form.bk_networkunit_id),
      },
      {
        id: 'custom',
        name: t('installProxy.custom'),
        children: areaUnitlist.value.map((item: NetworkUnit) => ({
          id: String(item.bk_networkunit_id),
          name: `[${item.bk_networkunit_id}] ${item.bk_networkunit_name}`,
        })),
      },
    ];
  } else {
    list = [
      {
        id: 'current',
        name: t('installProxy.currentUnit'),
        bk_networkunit_id: Number(props.bk_networkunit_id || form.bk_networkunit_id),
      },
      {
        id: 'custom',
        name: t('installProxy.custom'),
        children: areaUnitlist.value.map((item: NetworkUnit) => ({
          id: String(item.bk_networkunit_id),
          name: `[${item.bk_networkunit_id}] ${item.bk_networkunit_name}`,
        })),
      },
    ];
  }
  return list;
});
const isShowDialog = ref(false);
const dialogData = ref([{
  os: '',
  version: '',
}]);

const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
};
const handleComfirmVerion = (data: any[]) => {
  dialogData.value[0].version = data[0]?.version;
};
const handleChange = async (value: string) => {
  form.method = value;
  formRef.value?.clearValidate();
  if (value === 'offline') {
    const uid = props.bk_networkunit_id || Number(form.bk_networkunit_id);
    if (uid) await checkUnitHasProxy(uid);
  }
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  InfoBox({
    title: t('installProxy.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
  });
});

// 安装策略侧边栏
const isShowStrategy = ref(false);
const handleShowStrategy = () => {
  isShowStrategy.value = true;
};
const strategyAreaName = computed(() => {
  if (form.bk_networkarea_name) return form.bk_networkarea_name;
  const unitId = props.bk_networkunit_id || Number(form.bk_networkunit_id);
  if (!unitId) return '';
  const unit = areaUnitlist.value.find((u: NetworkUnit) => u.bk_networkunit_id === unitId);
  if (unit?.bk_networkarea_id != null) {
    const area = workareaStore.allWorkareaList.get(unit.bk_networkarea_id);
    return area?.bk_networkarea_name || '';
  }
  return '';
});

const formRef = ref(null);
const installTableRef = ref(null);
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
const isShowPreview = ref(false);
const previewData = ref<any>({});
const proxy_tags = ['dedicated_installer', 'cluster_tunnel', 'file_tunnel', 'data_tunnel'];
const handleConfirm = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isTargetShow.value ? systemValidate() : true,
  ]);
  if (Array.isArray(result[2])) {
    result[2] = result[2].every(item => item);
  }
  if (result.every(item => item)) {
    const modeMap = {
      password: 'login_password',
      keyfile: 'login_key_file',
    };
    // Deep clone to avoid mutating original form.info — preserve credit on API failure
    const clonedInfo = cloneDeep(form.info);
    clonedInfo.forEach((item: any) => {
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
      Object.keys(item).forEach((key: string) => {
        if (proxy_tags.includes(key) && item[key] && !item.proxy_tags?.includes(key)) {
          item.proxy_tags?.push(key);
        }
      });
      delete item.credit;
    });
    if (isTargetShow.value) {
      form.target_version = systemData.value
        .filter((item: any) => !!item.version)
        .map((item: any) => ({
          os_type: item.os_type,
          cpu_arch: item.cpu_arch,
          version: item.version,
        }));
    }
    const unitId = props.bk_networkunit_id || Number(form.bk_networkunit_id);
    const proxy_install_origin_unit_id = form.method === 'offline'
      ? getDefaultProxyInstallOriginUnitIdForNewInstall()
      : (form.proxy_install_origin[0] === 'custom'
        ? Number(form.proxy_install_origin[1])
        : installOriginList.value.find(item => item.id === form.proxy_install_origin[0])?.bk_networkunit_id ?? unitId);
    const unitName = areaUnitlist.value.find(
      (u: NetworkUnit) => u.bk_networkunit_id === unitId,
    )?.bk_networkunit_name ?? '';
    const hosts = clonedInfo.map((item: any) => {
      const {
        bk_host_id,
        dedicated_installer,
        cluster_tunnel,
        file_tunnel,
        data_tunnel,
        login_port: rowLoginPort,
        login_user: rowLoginUser,
        ...rest
      } = item;
      return {
        ...rest,
        os_type: rest.os_type || 'linux',
        bk_biz_id: form.bk_biz_id,
        proxy_install_origin_unit_id,
        relay_download_port: Number(form.relay_download_port),
        relay_callback_port: Number(form.relay_callback_port),
        bk_networkunit_id: unitId,
        bk_networkunit_name: unitName,
        bk_networkarea_name: form.bk_networkarea_name,
        ...(bk_host_id !== null && bk_host_id !== '' ? { bk_host_id } : {}),
        ...(form.method !== 'manual' && form.method !== 'offline' ? {
          login_user: rowLoginUser,
          login_port: Number(rowLoginPort),
          credit_expired_interval_sec: 7 * 24 * 3600,
        } : {}),
        ...(form.method === 'offline' ? (() => {
          const os = rest.os_type || 'linux';
          const defaultUser = os === 'windows' ? 'administrator' : 'root';
          const defaultPort = os === 'windows'
            ? window.PROJECT_CONFIG.WINDOWS_WMI_PORT_DEFAULT
            : window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
          const portNum = rowLoginPort !== '' && rowLoginPort != null
            ? Number(rowLoginPort)
            : NaN;
          return {
            login_user: (rowLoginUser && String(rowLoginUser).trim()) ? rowLoginUser : defaultUser,
            login_port: Number.isFinite(portNum) && portNum > 0 ? portNum : defaultPort,
          };
        })() : {}),
      };
    });
    previewData.value = {
      hosts,
      target_version: form.target_version,
      is_manual: form.method === 'manual',
      is_offline: form.method === 'offline',
    };
    isShowPreview.value = true;
  } else {
    scrollToFirstErrorByClassNames();
    Message({
      theme: 'warning',
      message: t('installProxy.validationFailed'),
    });
  }
};

// excel 导入弹窗
const isShowExcelImport = ref(false);
const excelImportData = ref([]);
const uploadExcelRef = ref();
const handleExcelImport = () => {
  isShowExcelImport.value = true;
};
const handleExcelImportConfirm = () => {
  // 导入数据覆盖当前表格：过滤掉空行（无 ipv4 且无 ipv6 的行），布尔字段缺省补 true
  const boolDefaults = { dedicated_installer: true, cluster_tunnel: true, file_tunnel: true, data_tunnel: true };
  const newRows = (excelImportData.value as any[]).filter((row: any) => {
    return row.bk_host_innerip || row.bk_host_innerip_v6;
  }).map((row: any) => {
    for (const key of Object.keys(boolDefaults)) {
      if (row[key] == null) row[key] = boolDefaults[key as keyof typeof boolDefaults];
    }
    return row;
  });

  // 直接使用导入数据覆盖当前表格
  form.info = newRows.length > 0 ? newRows : [cloneDeep(initData)];
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleExcelImportCancel = () => {
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleUpload = (data: any) => {
  excelImportData.value = data?.info || [];
};
onMounted(() => {
  encryptionTool.initPublicKey();
});

const handleSetpBack = () => {
  form.info = [];
};
// 获取版本，用来检查是否有对应架构的包版本去安装
const getVersions = async () => {
  const res = await PackageService.ListReleaseProxyBrief({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      release_type: ['proxy'],
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const osMap: any = {};
  res.items.forEach((item) => {
    const key = `${item.os_type}_${item.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        enableVersions: [],
      };
    }
    if (item.enabled) {
      osMap[key].enableVersions.push({
        version: item.version,
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
      });
    }
  });
  systemData.value = systemData.value.filter(item => !!osMap[item.os]?.enableVersions.length);
};

watch(() => isShow.value, async () => {
  if (isShow.value) {
    await getVersions();
    // 如果从 props 传入了单元 id，初始化时设置安装源默认值
    if (props.bk_networkunit_id) {
      await setDefaultInstallOrigin();
    }
    // 预加载管控单元详情（含上游接入点信息），供安装策略侧边栏使用
    const unitId = props.bk_networkunit_id || Number(form.bk_networkunit_id);
    if (unitId) {
      topoStore.handleFetchNetworkUnitDetail(unitId);
    }
  } else {
    formRef.value?.clearValidate();
    // 重置数据
    Object.assign(form, {
      method: 'setup', // 安装方式
      info: [cloneDeep(initData)], // 安装信息
      bk_biz_id: '', // 归属业务
      bk_networkarea_id: '', // 管控区域
      bk_networkarea_name: '',
      bk_networkunit_id: '', // 管控单元
      target_version: [] as TargetVersion[],
      proxy_install_origin: [],
    });
  }
});
watch(
  () => form.bk_networkarea_id,
  async () => {
    form.bk_networkunit_id = '';
    await getNetworkUnitList();
  },
);

// props 传入的管控单元 id 变化时预加载详情（供安装策略侧边栏使用）
watch(
  () => props.bk_networkunit_id,
  (val) => {
    if (val) {
      topoStore.handleFetchNetworkUnitDetail(val);
    }
  },
  { immediate: true },
);
// 当用户选择管控单元变化时，重新设置安装源默认值
watch(
  () => form.bk_networkunit_id,
  async (newVal) => {
    if (newVal) {
      await setDefaultInstallOrigin();
    }
  },
);
onMounted(() => {
  encryptionTool.initPublicKey();
});
</script>
