<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.title')"
    width="1490"
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
            :current-settings="settings"
          ></install-table>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.password')"
          property=""
          label-width="90"
          required
        >
          <Radio.Group v-model="form.saveTime">
            <Radio.Button
              :label="
                $t('topoManager.installProxy.form.saveTime.oneDay', { x: 1 })
              "
            >
            </Radio.Button>
            <Radio.Button
              :label="$t('topoManager.installProxy.form.saveTime.longTermSave')"
            >
            </Radio.Button>
          </Radio.Group>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.os')"
          property=""
          label-width="90"
          required
        >
          <Select class="w-[488px]" v-model="form.os_type" disabled></Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.port')"
          property=""
          label-width="90"
          required
        >
          <Input class="w-[488px]" v-model="form.login_port" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.account')"
          property=""
          label-width="90"
          required
        >
          <Input class="w-[488px]" v-model="form.login_user" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.business')"
          property=""
          label-width="90"
          required
        >
          <Select
            class="w-[488px]"
            v-model="form.bk_biz_id"
            auto-focus
            filterable
            placeholder="选择业务"
          >
            <Select.Option
              v-for="item in businessList"
              :key="item.bk_biz_id"
              :name="item.bk_biz_name"
              :id="item.bk_biz_id"
            >
              [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          v-if="bk_networkunit_id === null"
          :label="$t('platform.nodeMan.bk_cloud_name')"
          property="bk_networkarea_id"
          label-width="90"
          required
        >
          <Select
            class="w-[488px]"
            v-model="form.bk_networkarea_id"
            auto-focus
            filterable
            @select="handleSelect"
          >
            <Select.Option
              v-for="option in networkAreaList"
              :key="option.bk_networkarea_id"
              :id="String(option.bk_networkarea_id)"
              :name="option.bk_networkarea_name"
            >
              {{ option.bk_networkarea_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          v-if="bk_networkunit_id === null"
          :label="$t('platform.nodeMan.bk_cloud_unit')"
          property="bk_networkunit_id"
          label-width="90"
          required
        >
          <Select
            class="w-[488px]"
            v-model="form.bk_networkunit_id"
            auto-focus
            filterable
            :disabled="!form.bk_networkarea_id"
          >
            <Select.Option
              v-for="option in networkUnitList"
              :key="option.bk_networkarea_id"
              :id="String(option.bk_networkunit_id)"
              :name="option.bk_networkunit_name"
            >
              {{ option.bk_networkunit_name }}
            </Select.Option>
          </Select>
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
        <Form.FormItem :label="$t('Proxy 版本')" label-width="90" required v-if="isTargetShow">
          <div class="w-[488px]">
            <Table :data="systemData" :border="true">
              <TableColumn
                field="displayName"
                :title="$t('操作系统/架构')"
                width="200"
              ></TableColumn>
              <TableColumn field="version" :title="$t('版本')" width="288">
                <!-- <template #header>
                  <span>{{ $t('版本') }}</span>
                  <Button text @click="handleChooseVersion(row)">
                    <i class="nodeman-icon nc-edit text-[18px] cursor-pointer"></i>
                  </Button>
                </template> -->
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="`${row.os}_ref`"
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
        <div class="flex mt-[32px] ml-[90px]">
          <Button
            theme="primary"
            class="mr-[8px] w-[120px]"
            @click="handleConfirm"
          >
            <span>
              {{ $t("action.install") }}
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
      </Form>
    </div>
    <chooseVersionDialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :release-type="'proxy'"
      @confirm="handleComfirmVerion"
    ></chooseVersionDialog>
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Message, Radio, Select, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import type { PropType } from 'vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import SelectItemGroup from './components/select-item-group.vue';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import { useMainStore } from '@/stores/main';

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
const router = useRouter();
const { t } = useI18n();
const mainStore = useMainStore();

const form = reactive({
  method: '0', // 安装方式
  info: [
    {
      bk_host_id: '',
      bk_host_innerip: '',
      bk_host_innerip_v6: '',
      export_ip: '',
      advertise_ip: '',
      login_ip: '',
      login_mode: 'password',
      login_password: '',
      login_key_file: '',
      bk_addressing: 'static',
      dedicated_installer: true,
      cluster_tunnel: true,
      file_tunnel: true,
      data_tunnel: true,
      proxy_tags: [] as string[],
    },
  ], // 安装信息
  saveTime: '保存 1 天', // 密钥/密码
  os_type: 'Linux', // 操作系统
  login_port: '36000', // 登录端口
  login_user: 'root', // 登录账号
  bk_biz_id: '', // 归属业务
  bk_networkarea_id: '', // 管控区域
  bk_networkarea_name: '',
  bk_networkunit_id: '', // 管控单元
  target_version: [] as TargetVersion[],
});
const settings = reactive({
  fields: [
    { field: 'bk_host_innerip', title: '内网 IPv4' },
    { field: 'bk_host_innerip_v6', title: '内网 IPv6' },
    { field: 'export_ip', title: '出口IP' },
    { field: 'advertise_ip', title: '服务IP' },
    { field: 'login_ip', title: '登录 IP' },
    { field: 'login_mode', title: '认证方式' },
    { field: 'prove', title: '密码 / 密钥' },
    { field: 'dedicated_installer', title: '安装跳板' },
    { field: 'cluster_tunnel', title: 'Agent控制' },
    { field: 'file_tunnel', title: '文件传输' },
    { field: 'data_tunnel', title: '数据上报' },
  ],
  checked: [
    'bk_host_innerip',
    'export_ip',
    'login_ip',
    'login_mode',
    'prove',
  ],
  disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'prove'],
  size: 'medium',
});
const isTargetShow = ref(false);
const businessList = computed(() => mainStore.businessList);
const systemData = ref([
  {
    displayName: 'linux/amd64',
    os: 'Linux_amd64',
    cpu_arch: 'amd64',
    os_type: 'linux',
    version: '默认',
  },
  {
    displayName: 'linux/arm64',
    os: 'Linux_arm64',
    cpu_arch: 'arm64',
    os_type: 'linux',
    version: '默认',
  },
]);
// 安装方式列表
const installMethodList = ref([
  {
    icon: 'nodeman-icon nc-remote-install',
    title: t('topoManager.installProxy.installMethodList.remote.title'),
    content: t('topoManager.installProxy.installMethodList.remote.content'),
    value: 0,
  },
  {
    icon: 'nodeman-icon nc-excel-2',
    title: t('topoManager.installProxy.installMethodList.excel.title'),
    content: t('topoManager.installProxy.installMethodList.excel.content'),
    value: 1,
  },
  {
    icon: 'nodeman-icon nc-custom-install',
    title: t('topoManager.installProxy.installMethodList.manual.title'),
    content: t('topoManager.installProxy.installMethodList.manual.content'),
    value: 2,
  },
]);
const networkAreaList = ref<NetworkArea[]>([]);
// 管控区域下拉列表获取
const getNetworkAreaList = async () => {
  const res = await TopoService.NetworkAreaList({
    page: {
      limit: 0,
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  networkAreaList.value = res.items;
};

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: [Number(form.bk_networkarea_id)],
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;
};
// 选择管控区域
const handleSelect = (newValue: string) => {
  form.bk_networkarea_name = networkAreaList.value?.find((item: any) => String(item.bk_networkarea_id) === newValue)?.bk_networkarea_name || '';
};

const isShowDialog = ref(false);
const dialogData = ref({
  os: '',
  version: '',
});

const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = row;
};
const handleComfirmVerion = (val: string) => {
  val && (dialogData.value.version = val);
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
const Linux_amd64_ref = ref();
const Linux_arm64_ref = ref();
const proxy_tags = ['dedicated_installer', 'cluster_tunnel', 'file_tunnel', 'data_tunnel'];
const handleConfirm = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isTargetShow.value
      ? Promise.all([
        Linux_amd64_ref.value?.validate('blur').catch(() => false),
        Linux_arm64_ref.value?.validate('blur').catch(() => false),
      ])
      : true,
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
      item[modeMap[item.login_mode]] = item.prove;
      Object.keys(item).forEach((key: string) => {
        if (proxy_tags.includes(key) && item[key] && !item.proxy_tags.includes(key)) {
          item.proxy_tags.push(key);
        }
      });
    });
    if (isTargetShow.value) {
      form.target_version = systemData.value
        .filter((item: any) => item.version !== '默认')
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
          bk_biz_id: form.bk_biz_id,
          login_user: form.login_user,
          login_port: Number(form.login_port),
          bk_networkunit_id: props.bk_networkunit_id || Number(form.bk_networkunit_id),
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
      message: 'proxy安装成功！',
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
watch(() => isShow.value, () => {
  if (isShow.value) {
    Object.assign(form, props.data);
  }
});
watch(
  () => form.bk_networkarea_id,
  async () => {
    await getNetworkUnitList();
  },
);
onMounted(async () => {
  await getNetworkAreaList();
});
</script>
