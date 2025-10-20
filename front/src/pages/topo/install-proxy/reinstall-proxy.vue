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
          <SelectItemGroup :list="installMethodList" @change="handleChange">
          </SelectItemGroup>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.info')"
          property=""
          label-width="90"
          required
        >
          <install-table
            ref="installTableRef"
            v-model:data="form.info"
            realease-type="proxy"
            :is-reinstall="true"
            :method="form.method"
            :current-settings="settings"
          ></install-table>
        </Form.FormItem>
        <Form.FormItem
          label-width="90">
          <Button
            text
            theme="primary"
            class="text-[14px]"
            @click="isTargetShow = !isTargetShow"
          >
            <span class="mr-[8.5px]">高级选项</span>
            <angle-double-down-line
              :class="{ 'transform rotate-180': isTargetShow }"
            />
          </Button>
        </Form.FormItem>
        <Form.FormItem
          v-if="isTargetShow"
          :label="'安装源'"
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
        <Form.FormItem :label="$t('Proxy 版本')" label-width="90" required v-if="isTargetShow">
          <div class="w-[488px]">
            <Table :data="systemData" :border="true">
              <TableColumn
                field="displayName"
                :title="$t('操作系统/架构')"
                width="200"
              ></TableColumn>
              <TableColumn field="version" :title="$t('版本')" width="288">
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="(el) => setInputRef(row.os, el)"
                  >
                    <Input
                      :model-value="row.version"
                      :placeholder="$t('请选择')"
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
import { Button, Cascader, Form, InfoBox, Input, Message, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import type { PropType } from 'vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import SelectItemGroup from './components/select-item-group.vue';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import Validate from '@/components/validate.vue';

const isShow = defineModel<boolean>('isShow', { default: false });
const props = defineProps({
  data: {
    type: Array as PropType<Host[]>,
    default: [],
  },
});
const router = useRouter();
const { t } = useI18n();
const settings = reactive({
  fields: [
    { field: 'bk_biz_id', title: '归属业务' },
    { field: 'bk_host_innerip', title: '内网 IPv4' },
    { field: 'bk_host_innerip_v6', title: '内网 IPv6' },
    { field: 'os_type', title: '操作系统' },
    { field: 'login_port', title: '登录端口' },
    { field: 'login_user', title: '登录账号' },
    { field: 'export_ip', title: '出口IP' },
    { field: 'advertise_ip', title: '服务IP' },
    { field: 'login_ip', title: '登录 IP' },
    { field: 'login_mode', title: '认证方式' },
    { field: 'credit', title: '密码 / 密钥' },
    { field: 'dedicated_installer', title: '安装跳板' },
    { field: 'cluster_tunnel', title: 'Agent控制' },
    { field: 'file_tunnel', title: '文件传输' },
    { field: 'data_tunnel', title: '数据上报' },
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
  method: '0', // 安装方式
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
    os: 'Linux_amd64',
    cpu_arch: 'amd64',
    os_type: 'linux',
    version: '自动',
  },
  {
    displayName: 'linux/arm64',
    os: 'Linux_arm64',
    cpu_arch: 'arm64',
    os_type: 'linux',
    version: '自动',
  },
]);
// 安装方式列表
const installMethodList = ref([
  {
    icon: 'nodeman-icon nc-remote-install',
    title: t('topoManager.installProxy.installMethodList.remote.title'),
    content: t('topoManager.installProxy.installMethodList.remote.content'),
    value: '0',
  },
  {
    icon: 'nodeman-icon nc-custom-install',
    title: t('topoManager.installProxy.installMethodList.manual.title'),
    content: t('topoManager.installProxy.installMethodList.manual.content'),
    value: '2',
  },
]);
const isShowDialog = ref(false);
const dialogData = ref([{
  os: '',
  version: '',
}]);

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
  res.items.forEach(item => {
    networkUnitListMap.set(item.bk_networkunit_id, item.links.cluster.bk_networkunit_id);
  });
};

// 安装源
const installOriginList = computed(() => ([
  {
    id: 'upstream',
    name: '上级管控单元',
  },
  {
    id: 'current',
    name: '当前管控单元',
  },
  {
    id: 'custom',
    name: '自定义',
    children: networkUnitList.value.map(item => ({
      id: String(item.bk_networkunit_id),
      name: `[${item.bk_networkunit_id}] ${item.bk_networkunit_name}`,
    })),
  },
]));

const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
};
const handleComfirmVerion = (data: any) => {
  data.version && (dialogData.value[0].version = data.version === 'auto' ? '自动' : data.version);
};
const handleChange = (values: Array<string | number>) => {
  form.method = values[0] as string;
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
        .filter((item: any) => item.version !== '自动')
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
    };
    const res = await NodeProxyService.NodeProxyInstall(params).catch((err) => {
      console.log(err);
    });
    if (!res) return;
    Message({
      theme: 'success',
      message: 'proxy重装成功！',
    });
    isShow.value = false;
    if (res.workflow_id) {
      router.push({
        name: 'taskDetail',
        params: { taskId: res.workflow_id },
      });
    }
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

watch(() => isShow.value, async () => {
  if (isShow.value && props.data.length) {
    form.info = props.data.map((item: Host) => {
      const data = cloneDeep(initData);
      assign(data, item, item.info);
      return data;
    });
    await getNetworkUnitList();
  }
});
</script>
