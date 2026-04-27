<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.reinstall')"
    width="1200"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <div class="py-[20px] px-[40px]">
      <!-- form -->
      <Form ref="formRef" :model="form" :rules="formRules" class="mt-[24px]">
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.method')"
          property=""
          label-width="90"
          required
        >
          <install-type currentNodeType="proxy" :needTypeList="['setup', 'manual', 'offline']" @change="handleChange"></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.info')"
          property=""
          label-width="90"
          required
        >
          <Loading :loading="loading">
            <install-table
              ref="installTableRef"
              v-model:data="form.info"
              release-type="proxy"
              :is-reinstall="true"
              :current-settings="settings"
              :max-height="520"
            ></install-table>
          </Loading>
        </Form.FormItem>
        <Form.FormItem
          label-width="90">
          <Button
            text
            theme="primary"
            class="text-[14px]"
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
          :label="t('installProxy.installSource')"
          property="proxy_install_origin"
          label-width="90"
          required
        >
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
          label-width="90"
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
                  <p>{{ t('installProxy.relayCallbackPortTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <Input class="w-[488px]" v-model="form.relay_callback_port" />
        </Form.FormItem>
        <Form.FormItem
          v-if="isTargetShow"
          property="relay_download_port"
          label-width="90"
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
                  <p>{{ t('installProxy.relayDownloadPortTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <Input class="w-[488px]" v-model="form.relay_download_port" />
        </Form.FormItem>
        <Form.FormItem :label="t('installProxy.proxyVersion')" label-width="90" required v-if="isTargetShow">
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
      <div class="flex mt-[32px] ml-[90px]">
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
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Cascader, Form, InfoBox, Input, Loading, Message, Popover, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import ProxyPreview from './preview.vue';

import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { encryptionTool } from '@/common/crypto';
import { PACKAGE_GENERATION } from '@/common/const';
import { getDefaultLoginMode, resolveLoginMode, scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';

const isShow = defineModel<boolean>('isShow', { default: false });
const props = defineProps({
  data: {
    type: Array as PropType<Host[]>,
    default: [],
  },
  isCrossPageSelection: {
    type: Boolean,
    default: false,
  },
  params: {
    type: Object,
    default: () => ({}),
  },
});
const { t } = useI18n();
const settings = reactive({
  fields: [
    { field: 'bk_biz_id', title: t('installProxy.business') },
    { field: 'bk_networkarea_name', title: t('platform.nodeMan.bk_cloud_name') },
    { field: 'bk_networkunit_id', title: t('platform.nodeMan.bk_cloud_unit') },
    { field: 'bk_host_innerip', title: t('installProxy.innerIPv4') },
    { field: 'bk_host_innerip_v6', title: t('installProxy.innerIPv6') },
    { field: 'os_type', title: t('installProxy.os') },
    { field: 'cpu_arch', title: t('components.installTable.cpuArch') },
    { field: 'login_port', title: t('installProxy.port') },
    { field: 'login_user', title: t('installProxy.account') },
    { field: 'export_ip', title: t('installProxy.exportIP') },
    { field: 'advertise_ip', title: t('installProxy.serviceIP') },
    { field: 'login_ip', title: t('installProxy.loginIP') },
    { field: 'login_mode', title: t('installProxy.authMethod') },
    { field: 'credit', title: t('installProxy.passwordKey') },
    { field: 'dedicated_installer', title: t('installProxy.installJump') },
    { field: 'cluster_tunnel', title: t('installProxy.agentControl') },
    { field: 'file_tunnel', title: t('installProxy.fileTransfer') },
    { field: 'data_tunnel', title: t('installProxy.dataReport') },
  ],
  checked: [
    'bk_biz_id',
    'bk_networkarea_name',
    'bk_networkunit_id',
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'cpu_arch',
    'login_port',
    'login_ip',
    'login_user',
    'login_mode',
    'credit',
  ],
  disabled: ['os_type', 'cpu_arch', 'login_port', 'login_user', 'login_mode', 'credit', 'bk_networkunit_id'],
  size: 'medium',
});
const initData = {
  login_credit_valid: false,
  credit: '',
  bk_host_id: '',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  bk_networkarea_id: '',
  bk_networkarea_name: '',
  bk_networkunit_id: '',
  bk_networkunit_name: '',
  export_ip: '',
  advertise_ip: '',
  login_ip: '',
  login_mode: getDefaultLoginMode(),
  login_password: '',
  login_key_file: '',
  bk_addressing: 'static',
  bk_biz_id: '',
  os_type: '',
  login_port: window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT,
  login_user: 'root',
  dedicated_installer: true,
  cluster_tunnel: true,
  file_tunnel: true,
  data_tunnel: true,
  proxy_tags: [] as string[],
  cpu_arch: '',
  relay_download_port: '',
  relay_callback_port: '',
};
const form = reactive({
  method: 'setup', // 安装方式
  info: [
    cloneDeep(initData),
  ],
  target_version: [] as TargetVersion[],
  proxy_install_origin: [] as string[],
  relay_download_port: '',
  relay_callback_port: '',
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
  login_port: [
    { required: true, message: t('validate.required'), trigger: 'blur' },
    { validator: portValidator, message: t('installProxy.portMustBeValidRange'), trigger: 'blur' },
  ],
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

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const networkUnitListMap = new Map();
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkunit_id: props.data.map(item => item.info.bk_networkunit_id),
    },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;
  res.items.forEach((item) => {
    networkUnitListMap.set(item.bk_networkunit_id, item.links?.cluster?.bk_networkunit_id);
  });
};

// 查询各单元是否有 proxy（一次查询所有去重单元，排除正在重装的 proxy）
const unitHasProxyMap = ref<Map<number, boolean>>(new Map());
const checkUnitsHasProxy = async (unitIds: number[]) => {
  const uniqueIds = [...new Set(unitIds)].filter(id => !!id);
  if (uniqueIds.length === 0) return;
  const map = new Map<number, boolean>();
  uniqueIds.forEach(id => map.set(id, false));

  // Collect bk_host_id of hosts being reinstalled to exclude from proxy count
  const reinstallHostIds = new Set(
    form.info
      .map((item: any) => item.bk_host_id)
      .filter((id: any) => id !== '' && id !== null && id !== undefined)
      .map((id: any) => Number(id)),
  );

  try {
    const res = await TopoService.HostList({
      page: { offset: 0, limit: 500 },
      only_count: false,
      exact_include_conditions: {
        bk_networkunit_id: uniqueIds,
        node_role: ['proxy'],
        node_status: ['running'],
      },
      fuzzy_include_conditions: {},
    });
    // Filter out proxies being reinstalled — they should not count as available
    const filteredItems = (res.items ?? []).filter(
      (item: any) => !reinstallHostIds.has(Number(item.bk_host_id)),
    );
    const unitsWithProxy = new Set(filteredItems.map((item: any) => Number(item.info.bk_networkunit_id)));
    unitsWithProxy.forEach((unitId) => {
      map.set(unitId, true);
    });
  } catch {
    // 请求失败时所有单元保持无 proxy
  }
  unitHasProxyMap.value = map;
};

// 根据单元是否有 proxy 设置安装源默认值（不展开高级选项；离线无安装源 UI，不展开、不写 cascader）
const setDefaultInstallOrigin = () => {
  const firstUnitId = Number(form.info[0]?.bk_networkunit_id);
  if (!firstUnitId) return;
  if (form.method === 'offline') {
    return;
  }
  const hasProxy = unitHasProxyMap.value.get(firstUnitId) ?? false;
  const upstreamUnitId = networkUnitListMap.get(firstUnitId);
  const hasUpstream = upstreamUnitId !== null && upstreamUnitId !== undefined;
  if (hasProxy) {
    form.proxy_install_origin = ['current'];
  } else if (hasUpstream) {
    form.proxy_install_origin = ['upstream'];
  } else {
    isTargetShow.value = true;
  }
};

// 获取指定单元的默认安装源（不展开时按每行数据单独判断）
const getDefaultOriginForUnit = (unitId: number): string => {
  const hasProxy = unitHasProxyMap.value.get(unitId) ?? false;
  const upstreamUnitId = networkUnitListMap.get(unitId);
  const hasUpstream = upstreamUnitId !== null && upstreamUnitId !== undefined;
  if (hasProxy) {
    return 'current';
  }
  if (hasUpstream) {
    return 'upstream';
  }
  return 'current';
};

// 安装源
const installOriginList = computed(() => ([
  {
    id: 'upstream',
    name: t('installProxy.upstreamUnit'),
  },
  {
    id: 'current',
    name: t('installProxy.currentUnit'),
  },
  {
    id: 'custom',
    name: t('installProxy.custom'),
    children: networkUnitList.value.map(item => ({
      id: String(item.bk_networkunit_id),
      name: `[${item.bk_networkunit_id}] ${item.bk_networkunit_name}`,
    })),
  },
]));

// 版本选择弹窗
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

const originData = ref<any>();
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  // 没有修改，直接关闭
  if (isEqual(form, originData.value)) {
    resolve(true);
    isShow.value = false;
    return;
  }
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
const formRef = ref<InstanceType<typeof Form>>();
const installTableRef = ref<{ tableValidate: () => Promise<boolean>; showSetting: () => void }>();
const inputRefs = ref<Map<string, InstanceType<typeof Validate>>>(new Map());
const setInputRef = (
  os: string,
  el: any,
) => {
  if (el) {
    const key = os;
    inputRefs.value.set(key, el as InstanceType<typeof Validate>);
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

// Fill relay port defaults from proxy list (do not expand; expand only on install when invalid)
const setRelayPortDefaults = () => {
  const ports = form.info
    .map((row: any) => {
      const dlRaw = row.relay_download_port;
      const cbRaw = row.relay_callback_port;
      const dl = dlRaw != null && dlRaw !== '' ? Number(dlRaw) : null;
      const cb = cbRaw != null && cbRaw !== '' ? Number(cbRaw) : null;
      if (dl != null && cb != null && !Number.isNaN(dl) && !Number.isNaN(cb)) return `${dl}-${cb}`;
      return null;
    })
    .filter(Boolean);
  const unique = [...new Set(ports)];
  if (unique.length === 1 && unique[0]) {
    const [dl, cb] = unique[0].split('-').map(Number);
    form.relay_download_port = String(dl);
    form.relay_callback_port = String(cb);
  } else {
    form.relay_download_port = '';
    form.relay_callback_port = '';
  }
};

const handleConfirm = async () => {
  // 如果端口字段为空且高级选项未展开，先展开让表单校验能触发
  const dlRaw = form.relay_download_port?.trim();
  const cbRaw = form.relay_callback_port?.trim();
  if (!dlRaw || !cbRaw) {
    isTargetShow.value = true;
  }

  let result: unknown[];
  try {
    result = await Promise.all([
      formRef.value?.validate().catch(() => false),
      installTableRef.value?.tableValidate().catch(() => false),
      isTargetShow.value ? systemValidate().catch(() => false) : true,
    ]);
  } catch (err) {
    Message({ theme: 'error', message: (err as Error)?.message || t('validate.required') });
    return;
  }
  const resultArr = Array.isArray(result) ? result : [result];
  if (Array.isArray(resultArr[2])) {
    resultArr[2] = (resultArr[2] as boolean[]).every(item => item);
  }
  if (resultArr.every(item => item !== false)) {
    // 表单校验通过后，转换端口值
    const relayDownload = Number(form.relay_download_port);
    const relayCallback = Number(form.relay_callback_port);

    const modeMap: Record<string, string> = {
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
    const hosts = clonedInfo.map((item: any) => {
        const {
          bk_host_id,
          dedicated_installer,
          cluster_tunnel,
          file_tunnel,
          data_tunnel,
          ...rest
        } = item;
        const bkNetworkUnitId = Number(rest.bk_networkunit_id);
        const host = {
          ...rest,
          bk_networkunit_id: bkNetworkUnitId,
          os_type: 'linux',
          login_port: Number(rest.login_port),
          relay_download_port: relayDownload,
          relay_callback_port: relayCallback,
          proxy_install_origin_unit_id: getinstallOriginUnitId(bkNetworkUnitId),
          credit_expired_interval_sec: 7 * 24 * 3600,
          ...(bk_host_id != null && bk_host_id !== '' ? { bk_host_id } : {}),
        };
        return host;
    });
    previewData.value = {
      hosts,
      target_version: form.target_version,
      is_manual: form.method === 'manual',
      is_offline: form.method === 'offline',
    };
    isShowPreview.value = true;
  } else {
    await nextTick();
    scrollToFirstErrorByClassNames();
    Message({
      theme: 'warning',
      message: t('installProxy.validationFailed'),
    });
  }
};
function getinstallOriginUnitId(unit_id: number) {
  if (form.method === 'offline') {
    const o = getDefaultOriginForUnit(unit_id);
    if (o === 'upstream') {
      const id = networkUnitListMap.get(unit_id);
      return id ?? unit_id;
    }
    return unit_id;
  }
  // 如果没有展开高级选项（proxy_install_origin 为空），按每行数据的单元自动判断
  const origin = form.proxy_install_origin.length > 0
    ? form.proxy_install_origin[0]
    : getDefaultOriginForUnit(unit_id);
  let id;
  switch (origin) {
    case 'upstream':
      id = networkUnitListMap.get(unit_id);
      break;
    case 'current':
      id = unit_id;
      break;
    case 'custom':
      id = Number(form.proxy_install_origin[1]);
      break;
    default:
      break;
  }
  return id;
}
// 工具函数
const assign = (data1: any, data2: any, data3?: any) => {
  Object.keys(data1).forEach((key) => {
    data1[key] = data2[key] ?? data3?.[key] ?? data1[key];
  });
};

const normalizeNetworkUnitId = (id: unknown) => {
  if (id === '' || id === null || id === undefined) return '';
  return Number(id) === -1 ? '' : String(id);
};

const handleChange = (value: string) => {
  form.method = value;
  formRef.value?.clearValidate();
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
const loading = ref(false);

watch(() => isShow.value, async () => {
  if (isShow.value && props.data.length) {
    loading.value = true;
    try {
      // 先拉取网络单元，避免 Select 在无选项时回显原始 unit-id
      await getNetworkUnitList();

      // 使用TopoService.HostList接口进行切片查询获取数据
      if (props.isCrossPageSelection) {
        // 跨页全选模式：使用HostList接口分页获取所有数据
        const allHosts = [];
        const pageSize = 1000; // 每页大小
        let offset = 0;
        let hasMore = true;

        while (hasMore) {
          const hostListData = await TopoService.HostList({
            page: { offset, limit: pageSize },
            only_count: false,
            ...props.params,
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

        form.info = allHosts.map((host: any) => {
          const base = {
            ...cloneDeep(initData),
            ...host.state,
            ...host.info,
            ...host,
            bk_networkunit_id: normalizeNetworkUnitId(host.info?.bk_networkunit_id),
            bk_host_innerip: host.info?.bk_host_innerip_list?.join(','),
            bk_host_innerip_v6: host.info?.bk_host_innerip_v6_list?.join(','),
            login_mode: resolveLoginMode(host.info?.login_mode),
            proxy_tags: Array.isArray(host.proxy_tags) ? [...host.proxy_tags] : [],
          };
          // Normalize relay ports (API may return camelCase)
          const dlPort = host.info?.relay_download_port;
          const cbPort = host.info?.relay_callback_port;
          if (dlPort != null && dlPort !== '') base.relay_download_port = String(dlPort);
          if (cbPort != null && cbPort !== '') base.relay_callback_port = String(cbPort);
          return base;
        });
      } else {
        // 本页选择模式：使用原有数据
        form.info = props.data.map((item: Host) => {
          const data = cloneDeep(initData);
          assign(data, item, item.info);
          // Normalize relay ports from API (may return camelCase)
          const dlPort = item.info?.relay_download_port;
          const cbPort = item.info?.relay_callback_port;
          if (dlPort != null && String(dlPort) !== '') data.relay_download_port = String(dlPort);
          if (cbPort != null && String(cbPort) !== '') data.relay_callback_port = String(cbPort);
          data.bk_networkunit_id = normalizeNetworkUnitId(data.bk_networkunit_id);
          data.login_mode = resolveLoginMode(data.login_mode);
          return data;
        });
      }
      // 查询各单元是否有 proxy，并设置安装源默认值
      const unitIds = form.info.map((item: any) => Number(item.bk_networkunit_id)).filter(Boolean);
      await checkUnitsHasProxy(unitIds);
      setDefaultInstallOrigin();
      setRelayPortDefaults();

      await getVersions();
      originData.value = cloneDeep(form);
    } finally {
      loading.value = false;
    }
  }
});
onMounted(() => {
  encryptionTool.initPublicKey();
});
</script>
