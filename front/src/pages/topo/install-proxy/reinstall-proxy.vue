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
      <Form ref="formRef" :model="form" class="mt-[24px]">
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.method')"
          property=""
          label-width="90"
          required
        >
          <install-type currentNodeType="proxy" @change="handleChange"></install-type>
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
          v-if="isTargetShow"
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
    </div>
    <template #footer>
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
            {{ $t("action.reinstallProxy") }}
          </span>
          <span
            class="mx-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
          >
            {{ form.info.length }}
          </span>
        </Button>
        <Button @click="handleBeforeClose">
          {{ $t("action.cancel") }}
        </Button>
      </div>
    </template>
    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :release-type="'proxy'"
      @confirm="handleComfirmVerion"
    ></choose-version-dialog>
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Cascader, Form, InfoBox, Input, Loading, Message, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import SelectItemGroup from './components/select-item-group.vue';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { scrollToFirstErrorByClassNames } from '@/common/util';
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
const router = useRouter();
const { t } = useI18n();
const settings = reactive({
  fields: [
    { field: 'bk_biz_id', title: t('installProxy.business') },
    { field: 'bk_host_innerip', title: t('installProxy.innerIPv4') },
    { field: 'bk_host_innerip_v6', title: t('installProxy.innerIPv6') },
    { field: 'os_type', title: t('installProxy.os') },
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
const initData = {
  login_credit_valid: false,
  credit: '',
  bk_host_id: '',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  bk_networkunit_id: '',
  export_ip: '',
  advertise_ip: '',
  login_ip: '',
  login_mode: 'password',
  login_password: '',
  login_key_file: '',
  bk_addressing: 'static',
  bk_biz_id: '',
  os_type: '',
  login_port: '36000',
  login_user: 'root',
  dedicated_installer: true,
  cluster_tunnel: true,
  file_tunnel: true,
  data_tunnel: true,
  proxy_tags: [] as string[],
};
const form = reactive({
  method: 'setup', // 安装方式
  info: [
    cloneDeep(initData),
  ],
  target_version: [] as TargetVersion[],
  proxy_install_origin: [],
});
const isTargetShow = ref(false);
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
    console.log(err);
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
const proxy_tags = ['dedicated_installer', 'cluster_tunnel', 'file_tunnel', 'data_tunnel'];
const handleConfirm = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isTargetShow.value ? systemValidate() : true,
  ]);
  // 合并多重Promise
  if (Array.isArray(result[2])) {
    result[2] = result[2].every(item => item);
  }
  if (result.every(item => item)) {
    const modeMap = {
      password: 'login_password',
      key: 'login_key_file',
    };
    form.info.forEach((item: any) => {
      item[modeMap[item.login_mode]] = item.credit;
      Object.keys(item).forEach((key: string) => {
        if (proxy_tags.includes(key) && item[key] && !item.proxy_tags.includes(key)) {
          item.proxy_tags.push(key);
        }
      });
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
    const params = {
      host: form.info.map((item: any) => {
        const {
          bk_host_id,
          dedicated_installer,
          cluster_tunnel,
          file_tunnel,
          data_tunnel,
          ...rest
        } = item;
        return {
          ...rest,
          os_type: 'linux',
          login_port: Number(rest.login_port),
          proxy_install_origin_unit_id: getinstallOriginUnitId(rest.bk_networkunit_id),
          ...(bk_host_id !== null && bk_host_id !== '' ? { bk_host_id } : {}),
        };
      }),
      target_version: form.target_version,
      is_manual: form.method === 'manual',
    };
    const res = await NodeProxyService.NodeProxyInstall(params).catch((err) => {
      console.log(err);
    });
    if (!res) return;
    Message({
      theme: 'success',
      message: t('installProxy.reinstallInitiated'),
    });
    isShow.value = false;
    if (res.workflow_id) {
      router.push({
        name: 'taskDetail',
        params: { taskId: res.workflow_id },
        query: {
          active: 'node',
        },
      });
    }
  } else {
    scrollToFirstErrorByClassNames();
  }
};
function getinstallOriginUnitId(unit_id: Number) {
  let id;
  switch (form.proxy_install_origin[0]) {
    case 'upstream':
      id = networkUnitListMap.get(unit_id);
      break;
    case 'current':
      id = unit_id;
      break;
    case 'custom':
      id = Number(form.proxy_install_origin[1]);
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

const handleChange = (value: string) => {
  form.method = value;
  formRef.value?.clearValidate();
};

// 获取版本，用来检查是否有对应架构的包版本去安装
const getVersions = async () => {
  const res = await PackageService.ListReleaseProxy({
    page: { limit: 500, offset: 0 },
    generation: 2,
    exact_include_conditions: {
      release_type: ['proxy'],
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const osMap: any = {};
  res.items.forEach((item) => {
    const key = `${item.release.os_type}_${item.release.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        enableVersions: [],
      };
    }
    if (item.release.enabled) {
      osMap[key].enableVersions.push({
        version: item.release.version,
        os_type: item.release.os_type,
        cpu_arch: item.release.cpu_arch,
      });
    }
  });
  systemData.value = systemData.value.filter(item => !!osMap[item.os]?.enableVersions.length);
};
const loading = ref(false);

watch(() => isShow.value, async () => {
  if (isShow.value && props.data.length) {
    // 使用TopoService.HostList接口进行切片查询获取数据
    if (props.isCrossPageSelection) {
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

      form.info = allHosts.map((host: any) => ({
        ...host.state,
        ...host.info,
        ...host,
        bk_host_innerip: host.info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6: host.info.bk_host_innerip_v6_list?.join(','),
      }));
      loading.value = false;
    } else {
      // 本页选择模式：使用原有数据
      form.info = props.data.map((item: Host) => {
        const data = cloneDeep(initData);
        assign(data, item, item.info);
        return data;
      });
    }
    await getVersions();
    await getNetworkUnitList();
    originData.value = cloneDeep(form);
  }
});
</script>
