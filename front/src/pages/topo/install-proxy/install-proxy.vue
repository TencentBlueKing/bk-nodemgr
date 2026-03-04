<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.title')"
    width="1200"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <div class="py-[20px] px-[40px]">
      <!-- form -->
      <Form ref="formRef" :model="form" class="mt-[24px]">
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.method')"
          property="method"
          label-width="110"
          required
        >
          <install-type currentNodeType="proxy" @change="handleChange"></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.info')"
          property=""
          label-width="110"
          required
        >
          <!-- <template #label>
            <div class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.info') }}</div>
            <Button text theme="primary" @click="handleExcelImport">
              {{ $t('platform.nodeMan.installAgentPage.excelImport') }}
            </Button>
          </template> -->
          <install-table
            ref="installTableRef"
            v-model:data="form.info"
            release-type="proxy"
            :current-settings="settings"
            :max-height="520"
          >
            <UploadExcel @upload="handleUpload" v-if="form.method === '1'"></UploadExcel>
          </install-table>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.password')"
          property="saveTime"
          label-width="110"
          required
        >
          <Radio.Group v-model="form.saveTime">
            <Radio.Button
              :label="1"
            >
              {{ $t('topoManager.installProxy.form.saveTime.oneDay', { x: 1 }) }}
            </Radio.Button>
            <Radio.Button
              :label="7"
            >
              {{ $t('topoManager.installProxy.form.saveTime.oneDay', { x: 7 }) }}
            </Radio.Button>
            <Radio.Button
              :label="365"
            >
              {{ $t('topoManager.installProxy.form.saveTime.oneDay', { x: 365 }) }}
            </Radio.Button>
          </Radio.Group>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.os')"
          property="os_type"
          label-width="110"
          required
        >
          <Select class="w-[488px]" v-model="form.os_type" disabled></Select>
        </Form.FormItem>
        <Form.FormItem
          property="login_port"
          label-width="110"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="240"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('topoManager.installProxy.form.port') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.portTooltipDesc') }}</p>
                  <div class="mt-[8px]">
                    <p>{{ $t('components.installTable.portTooltipLinuxLabel') }}<span class="text-[#FF9C01]">{{ $t('components.installTable.portTooltipLinuxPort') }}</span></p>
                  </div>
                </div>
              </template>
            </Popover>
          </template>
          <Input class="w-[488px]" v-model="form.login_port" />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.installProxy.form.account')"
          property="login_user"
          label-width="110"
          required
        >
          <Input class="w-[488px]" v-model="form.login_user" />
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
          <Select
            class="w-[488px]"
            v-model="form.bk_biz_id"
            auto-focus
            filterable
            :filter-option="filterOption"
            :show-selected-icon="false"
            :placeholder="t('installProxy.selectBusiness')"
            @toggle="handleBizToggle"
          >
            <Select.Option
              v-for="item in sortedBusinessList"
              :key="item.bk_biz_id"
              :name="item.bk_biz_name"
              :id="item.bk_biz_id"
            >
              <div class="w-full flex items-center biz-select-option overflow-hidden">
                <Button
                  class="mr-[8px] w-[18px] shrink-0"
                  text
                  @click.native.stop="handleCollect(item.bk_biz_id)"
                >
                  <i
                    class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
                    v-if="collectList.includes(item.bk_biz_id)">
                  </i>
                  <i
                    class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
                    v-else>
                  </i>
                </Button>
                <div class="truncate">
                  [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
                </div>
              </div>
            </Select.Option>
          </Select>
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
          <Select
            class="w-[488px]"
            v-model="form.bk_networkunit_id"
            auto-focus
            filterable
            :disabled="!form.bk_networkarea_id"
          >
            <Select.Option
              v-for="option in areaUnitlist"
              :key="option.bk_networkunit_id"
              :id="String(option.bk_networkunit_id)"
              :name="option.bk_networkunit_name"
              :disabled="option.is_direct"
              v-bk-tooltips="{
                content: $t('topoManager.installProxy.form.tip'),
                disabled: !option.is_direct,
                boundary: 'parent',
                placement: 'left'
              }"
            >
              [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
            </Select.Option>
          </Select>
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
          v-if="isTargetShow"
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
        <div class="flex mt-[32px] ml-[90px] gap-[8px]">
          <!-- <Button
            v-if="excelImportData.length && form.info.length === 0"
            class="w-[100px]"
            theme="primary"
            @click="handleImport">
            {{ $t('platform.nodeMan.installAgentPage.excelImport') }}
          </Button> -->
          <Button
            theme="primary"
            class="w-[120px]"
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
              class="mx-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
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
      </Form>
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
      <UploadExcel ref="uploadExcelRef" @upload="handleUpload"></UploadExcel>
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
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Cascader, Dialog, Form, InfoBox, Input, Message, Popover, Radio, Select, Sideslider } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import ProxyPreview from './preview.vue';
import SelectItemGroup from './components/select-item-group.vue';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { encryptionTool } from '@/common/crypto';
import { scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';
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
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const mainStore = useMainStore();
const initData = {
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
};
const form = reactive({
  method: 'setup', // 安装方式
  info: [cloneDeep(initData)], // 安装信息
  saveTime: 1, // 密钥/密码保存时间
  os_type: 'linux', // 操作系统
  login_port: '36000', // 登录端口
  login_user: 'root', // 登录账号
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
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'export_ip',
    'login_ip',
    'login_mode',
    'credit',
  ],
  disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
  size: 'medium',
});
const isTargetShow = ref(false);
const businessList = computed(() => mainStore.businessList);
const collectList = ref<number[]>([]);
const selectedBizIds = computed(() => {
  if (!form.bk_biz_id) return [];
  const value = Number(form.bk_biz_id);
  return Number.isNaN(value) ? [] : [value];
});
const sortedBusinessList = computed(() => {
  const selectedSet = new Set(selectedBizIds.value);
  const list = [...businessList.value];
  return list.sort((a, b) => {
    const aIsCollected = collectList.value.includes(a.bk_biz_id);
    const bIsCollected = collectList.value.includes(b.bk_biz_id);
    const aIsSelected = selectedSet.has(a.bk_biz_id);
    const bIsSelected = selectedSet.has(b.bk_biz_id);

    if (aIsSelected && !bIsSelected) {
      return -1;
    }
    if (!aIsSelected && bIsSelected) {
      return 1;
    }
    if (aIsCollected && !bIsCollected) {
      return -1;
    }
    if (!aIsCollected && bIsCollected) {
      return 1;
    }
    return a.bk_biz_id - b.bk_biz_id;
  });
});
const filterOption = (input: any, options: { id: number, name: string }) => {
  const inputStr = String(input).trim();
  if (!inputStr) return false;

  const keywords = inputStr.split(/[\s,;]+/).filter(keyword => keyword.trim());
  if (keywords.length === 0) return false;

  const nameMatch = keywords.some(keyword => {
    const safeKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const nameRegex = new RegExp(safeKeyword, 'i');
    return options.name?.match(nameRegex);
  });
  const idMatch = keywords.some(keyword => {
    const keywordStr = String(keyword).trim();
    return keywordStr === String(options.id);
  });
  return nameMatch || idMatch;
};
const handleBizToggle = () => {
  // 侧边栏行为对齐：下拉展开/收起时重新计算排序
  sortedBusinessList.value;
};
const handleCollect = (val: number) => {
  const list = [...collectList.value];
  if (list.includes(val)) {
    collectList.value = list.filter(item => item !== val);
  } else {
    collectList.value = [...list, val];
  }
  localStorage.setItem('collect', JSON.stringify(collectList.value));
};
const initCollectList = () => {
  const collectsJson = localStorage.getItem('collect');
  if (collectsJson) {
    const collects = JSON.parse(collectsJson) as unknown;
    if (Array.isArray(collects)) {
      collectList.value = collects.map(item => Number(item)).filter(item => Number.isFinite(item));
    }
  }
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

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: [],
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

// eslint-disable-next-line max-len
const areaUnitlist = computed(() => networkUnitList.value.filter((item: NetworkUnit) => [Number(route.params.workarea), Number(form.bk_networkarea_id)].includes(item.bk_networkarea_id)));
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
const handleChange = (value: string) => {
  form.method = value;
  formRef.value?.clearValidate();
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
      key: 'login_key_file',
    };
    form.info.forEach((item: any) => {
      // 获取对应的 key (login_password 或 login_key_file)
      const targetKey = modeMap[item.login_mode];

      if (item.credit) {
        // 同步加密
        const encryptedValue = encryptionTool.encryptSync(item.credit);

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
    const proxy_install_origin_unit_id = form.proxy_install_origin[0] === 'custom'
      ? Number(form.proxy_install_origin[1])
      : installOriginList.value.find(item => item.id === form.proxy_install_origin[0])?.bk_networkunit_id;
    const hosts = form.info.map((item: any) => {
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
        proxy_install_origin_unit_id,
        credit_expired_interval_sec: form.saveTime * 24 * 3600,
        login_port: Number(form.login_port),
        relay_download_port: Number(form.relay_download_port),
        relay_callback_port: Number(form.relay_callback_port),
        bk_networkunit_id: props.bk_networkunit_id || Number(form.bk_networkunit_id),
        bk_networkarea_name: form.bk_networkarea_name,
        ...(bk_host_id !== null && bk_host_id !== '' ? { bk_host_id } : {}),
      };
    });
    previewData.value = {
      hosts,
      target_version: form.target_version,
      is_manual: form.method === 'manual',
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
  form.info = [...form.info, ...excelImportData.value];
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleExcelImportCancel = () => {
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleUpload = (data: any) => {
  excelImportData.value = data.info;
};
onMounted(() => {
  initCollectList();
});

const handleSetpBack = () => {
  form.info = [];
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

watch(() => isShow.value, async () => {
  if (isShow.value) {
    await getVersions();
  } else {
    formRef.value?.clearValidate();
    // 重置数据
    Object.assign(form, {
      method: 'setup', // 安装方式
      info: [cloneDeep(initData)], // 安装信息
      saveTime: 1, // 密钥/密码保存时间
      os_type: 'linux', // 操作系统
      login_port: '36000', // 登录端口
      login_user: 'root', // 登录账号
      bk_biz_id: '', // 归属业务
      bk_networkarea_id: '', // 管控区域
      bk_networkarea_name: '',
      bk_networkunit_id: '', // 管控单元
      target_version: [] as TargetVersion[],
      proxy_install_origin: [],
    });
  }
  originData.value = cloneDeep(form);
});
watch(
  () => form.bk_networkarea_id,
  async () => {
    form.bk_networkunit_id = '';
    await getNetworkUnitList();
  },
);
onMounted(() => {
  encryptionTool.initPublicKey();
});
</script>
<style lang="postcss" scoped>
.biz-select-option {
  &:hover {
    .nc-not-favorited {
      display: inline;
    }
  }
}
</style>
