<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="1200"
    :title="$t('platform.nodeMan.preview.title')"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #default>
      <div class="py-[24px] px-[40px]">
        <div
          class="flex min-h-[56px] bg-[#FFF4E2] border border-[#FFF4E2] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
        >
          <i class="nodeman-icon nc-tips pt-[2px] text-[#FF9C01]"></i>
          <div class="flex-1 text-[12px] text-[#4D4F56]">
            <p>{{ $t("platform.nodeMan.preview.tipTitle") }}</p>
            <p>{{ $t("platform.nodeMan.preview.firstTip") }}</p>
            <p>{{ $t("platform.nodeMan.preview.secondTip") }}</p>
            <p>{{ isZh
              ? '3. 导入CMDB并安装Proxy：节点不在CMDB中，将会自动导入然后安装Proxy。'
              : '3. Import to CMDB & Install Proxy: Node not in CMDB; will be imported automatically.' }}</p>
            <p>{{ isZh
              ? '4. 安装Proxy：节点无异常情况，可以正常安装/重装Proxy。'
              : '4. Install Proxy: Normal status; ready for installation/reinstallation.' }}</p>
          </div>
        </div>
        <div class="flex justify-between mt-[16px]">
          <div class="w-[50%] flex gap-[8px]">
            <SearchSelect
              class="flex-1 bg-[#fff]"
              ref="searchSelect"
              :data="searchSelectData"
              v-model.trim="searchSelectValue"
              :unique-select="true"
              :placeholder="$t('platform.nodeMan.preview.searchPlaceholder')"
            >
            </SearchSelect>
            <copy-ip-dropdown
              :type="'proxy'"
              :disabled="!selection.length"
              :data="tableData"
              :list="[]"
            ></copy-ip-dropdown>
          </div>
          <div class="flex gap-[8px]">
            <Button
              @click="handleAllConfirm"
              :disabled="
                !tableData.length ||
                  !tableData.find((item) => item.category === 'need_confirm')
              "
              v-bk-tooltips="{
                content: $t('platform.nodeMan.preview.processConfirmTip'),
              }"
            >
              {{ $t("platform.nodeMan.preview.button.batchConfirm") }}
            </Button>
            <Button
              @click="handleBatchRemove"
              :disabled="
                !originData.length ||
                  !originData.find((item: any) => item.category === 'error')
              "
            >
              {{ $t("platform.nodeMan.preview.button.batchRemove") }}
            </Button>
          </div>
        </div>
        <Tab
          class="mt-[16px]"
          v-model:active="active"
          type="card"
          :key="tabKey"
        >
          <Tab.TabPanel
            v-for="item in tabs"
            :key="item.name"
            :label="item.label"
            :name="item.name"
          >
            <template #label>
              <div class="flex gap-[5px] items-center">
                <close
                  v-if="item.icon === 'wrong'"
                  width="14px"
                  height="14px"
                  :fill="item.iconColor"
                />
                <i
                  v-else
                  :class="`nodeman-icon nc-${item.icon} text-[14px]`"
                  :style="{ color: item.iconColor }"
                ></i>
                <span class="text-[14px] text-[#313238]">{{ item.label }}</span>
                <div
                  :class="[
                    'rounded-[8px] w-[23px] h-[16px] border text-[12px] leading-[16px] text-center',
                    item.name === active ? 'bg-[#E1ECFF]' : 'bg-[#DCDEE5]'
                  ]"
                >
                  {{ item.count }}
                </div>
              </div>
            </template>
            <template #panel>
              <Table
                :data="tableData"
                :pagination="pagination"
                :empty-text="$t('table.empty')"
                :column-config="{ resizable: true }"
                show-overflow-tooltip
                :tooltip-config="{
                  popupClassName: 'preview-table',
                }"
                :max-height="462"
                :show-settings="isShowSetting"
                :settings="settings"
                :virtual-y-config="{ enabled: true, gt: 20 }"
                @setting-change="handleSettingChange"
                @checkbox-change="handleSelectChange"
                @checkbox-all="handleSelectAllChange"
              >
                <TableColumn
                  type="checkbox"
                  width="80"
                  fixed="left"
                ></TableColumn>
                <TableColumn
                  field="bk_host_innerip"
                  :title="t('platform.nodeMan.inner_ip')"
                  min-width="150"
                  fixed="left"
                ></TableColumn>
                <TableColumn
                  field="bk_host_innerip_v6"
                  :title="t('platform.nodeMan.inner_ipv6')"
                  min-width="150"
                ></TableColumn>
                <TableColumn
                  field="os_type"
                  :title="t('platform.nodeMan.os_type')"
                  min-width="100"
                ></TableColumn>
                <TableColumn
                  field="bk_networkarea_name"
                  :title="t('platform.nodeMan.bk_cloud_name')"
                  min-width="150"
                ></TableColumn>
                <TableColumn
                  field="bk_networkunit_name"
                  :title="$t('platform.nodeMan.bk_cloud_unit')"
                  min-width="150"
                ></TableColumn>
                <TableColumn
                  field="status"
                  :title="$t('platform.nodeMan.preview.status')"
                  min-width="400"
                >
                  <template #default="{ row }">
                    <Spinner v-if="installCheckLoading" class="mr-[8px]" />
                    <div class="flex items-center gap-[4px]" v-else>
                      <close
                        v-if="categoryMap[row.category]?.icon === 'wrong'"
                        width="14px"
                        height="14px"
                        :fill="categoryMap[row.category]?.iconColor"
                      />
                      <i
                        v-else
                        :class="`nodeman-icon nc-${
                          categoryMap[row.category]?.icon
                        } text-[14px]`"
                        :style="{
                          color: categoryMap[row.category]?.iconColor,
                        }"
                      ></i>
                      <OverflowTitle type="tips" class="flex-1 min-w-0">{{ isZh ? row.message_zh : row.message_en }}</OverflowTitle>
                    </div>
                  </template>
                </TableColumn>
                <TableColumn
                  field="action"
                  :title="t('platform.nodeMan.operate')"
                  min-width="140"
                  fixed="right"
                >
                  <template #default="{ row }">
                    <div class="flex gap-[8px]">
                      <Button
                        v-if="row.category === 'need_confirm'"
                        theme="primary"
                        text
                        @click="deel(row)"
                      >{{ $t("action.confirm1") }}</Button>
                      <Button theme="primary" text @click="handleRemove(row)">{{
                        t("platform.nodeMan.preview.button.remove")
                      }}</Button>
                    </div>
                  </template>
                </TableColumn>
              </Table>
            </template>
          </Tab.TabPanel>
        </Tab>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-start gap-[8px] pl-[16px]">
        <Button
          theme="primary"
          @click="handleSetup"
          :disabled="disabledDataNum > 0 || checkFailed"
          :loading="loading"
          v-bk-tooltips="{
            content: disabledDataNum
              ? $t('platform.nodeMan.preview.disabledTip')
              : installCheckLoading
                ? $t('platform.nodeMan.preview.loadingTip') : $t('platform.nodeMan.preview.checkFailedTip'),
            disabled: disabledDataNum === 0 && !checkFailed,
          }"
        >{{ $t('platform.nodeMan.preview.button.setup') }}</Button
        >
        <Button @click="handleBeforeClose">{{ t("action.cancel") }}</Button>
      </div>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import {
  Button,
  InfoBox,
  OverflowTitle,
  SearchSelect,
  Sideslider,
  Tab,
} from 'bkui-vue';
import { Close, Spinner } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { NodeProxyService } from '@/api/modules/node_proxy';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
});

const emit = defineEmits(['close']);

const { t } = useI18n();
const mainStore = useMainStore();
const router = useRouter();
let rowIdSeed = 0;
const genRowID = () => {
  rowIdSeed += 1;
  return `proxy-install-row-${rowIdSeed}`;
};
const originData = ref<any[]>([]);
const tableData = ref<any[]>([]);
const tabKey = ref(Date.now());
const disabledDataNum = ref(0);
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');
const checkFailed = computed(() => tableData.value.some(item => !item.status));
const { pagination } = usePage(tableData);

const categoryMap: Record<string, { icon: string; iconColor: string }> = {
  normal_install: {
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  register_to_cmdb_and_install: {
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  error: {
    icon: 'wrong',
    iconColor: '#EA3636',
  },
  need_confirm: {
    icon: 'danger-fill',
    iconColor: '#FF9C01',
  },
};

function getUniqueChildren(prop: string) {
  const res = Array.from(new Set(originData.value
    .map((item: any) => item[prop])
    .filter((item: any) => item)));
  return res.map((value: any) => {
    const name = String(value);
    return {
      id: value,
      name,
      value,
      text: name,
    };
  });
}

const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  {
    id: 'bk_host_innerip',
    name: t('platform.nodeMan.inner_ip'),
    children: getUniqueChildren('bk_host_innerip'),
  },
  {
    id: 'bk_host_innerip_v6',
    name: t('platform.nodeMan.inner_ipv6'),
    children: getUniqueChildren('bk_host_innerip_v6'),
  },
  {
    id: 'os_type',
    name: t('platform.nodeMan.os_type'),
    children: getUniqueChildren('os_type'),
    multiple: true,
  },
]);

const tabs = computed(() => {
  const countResult = originData.value.reduce(
    (acc: any, item: any) => {
      acc.all += 1;
      if (item.category === 'need_confirm') {
        acc.confirm += 1;
      }
      if (item.category === 'error') {
        acc.error += 1;
      }
      if (item.category === 'register_to_cmdb_and_install') {
        acc.cleanInstallation += 1;
      }
      if (item.category === 'normal_install') {
        acc.normalInstallation += 1;
      }
      return acc;
    },
    {
      all: 0,
      confirm: 0,
      error: 0,
      cleanInstallation: 0,
      normalInstallation: 0,
    },
  );
  disabledDataNum.value = countResult.confirm + countResult.error;
  return [
    {
      label: t('platform.nodeMan.preview.label.all'),
      name: 'all',
      count: countResult.all,
    },
    {
      label: t('platform.nodeMan.preview.label.pendingConfirmation'),
      name: 'need_confirm',
      count: countResult.confirm,
      icon: 'danger-fill',
      iconColor: '#FF9C01',
    },
    {
      label: t('platform.nodeMan.preview.label.error'),
      name: 'error',
      count: countResult.error,
      icon: 'wrong',
      iconColor: '#EA3636',
    },
    {
      label: isZh.value ? '导入CMDB并安装Proxy' : 'Import to CMDB & Install Proxy',
      name: 'register_to_cmdb_and_install',
      count: countResult.cleanInstallation,
      icon: 'check-circle-fill',
      iconColor: '#1CAB88',
    },
    {
      label: isZh.value ? '安装Proxy' : 'Install Proxy',
      name: 'normal_install',
      count: countResult.normalInstallation,
      icon: 'check-circle-fill',
      iconColor: '#1CAB88',
    },
  ];
});

const active = ref('all');
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'bk_host_innerip',
      'bk_host_innerip_v6',
      'bk_networkarea_name',
      'bk_networkunit_name',
      'os_type',
      'status',
      'action',
    ],
    disabled: ['action'],
  },
  'nodeMng-proxy-preview',
);

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
  });
});

const handleRemove = (row: any) => {
  originData.value = originData.value.filter((item: any) => item.__row_id !== row.__row_id);
  tabKey.value = Date.now();
};

const deel = async (row: any) => {
  if (['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(row.status)) {
    row.bk_host_id = row.matched.bk_host_id;
    await installCheck();
  } else {
    row.category = 'normal_install';
    tableData.value = [...tableData.value];
  }
};

const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const handleSelectChange = ({
  checked,
  row,
}: {
  checked: boolean;
  row: any;
}) => {
  row.checked = checked;
};

const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  tableData.value.forEach((item: any) => (item.checked = checked));
};

const handleAllConfirm = async () => {
  const duplicatedItems = tableData.value.filter((item: any) => ['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(item.status));
  if (duplicatedItems.length) {
    duplicatedItems.forEach((item: any) => {
      item.bk_host_id = item.matched.bk_host_id;
    });
    await installCheck();
  }
  tableData.value.forEach((item: any) => {
    if (item.category === 'need_confirm') {
      item.category = 'normal_install';
    }
  });
  tabKey.value = Date.now();
};

const handleBatchRemove = () => {
  originData.value = originData.value.filter((item: any) => item.category !== 'error');
  tabKey.value = Date.now();
};

const loading = ref(false);
const handleSetup = async () => {
  loading.value = true;
  const params = {
    host: originData.value.map((item: any) => {
      const {
        status, category, matched, message_en, message_zh, checked,
        __row_id,
        bk_networkarea_name, bk_networkunit_name,
        dedicated_installer, cluster_tunnel, file_tunnel, data_tunnel,
        ...rest
      } = item;
      return {
        ...rest,
        ...(item.bk_host_id ? { bk_host_id: item.bk_host_id } : {}),
      };
    }),
    target_version: props.data.target_version || [],
    is_manual: props.data.is_manual || false,
    is_offline: props.data.is_offline || false,
  };
  const res = await NodeProxyService.NodeProxyInstall(params).catch(() => ({
    workflow_id: '',
  }));
  loading.value = false;
  if (!res) return;
  if (res.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: res.workflow_id },
      query: {
        active: 'node',
      },
    });
    isShow.value = false;
    emit('close');
  }
};

const installCheckLoading = ref(false);
const installCheck = async () => {
  installCheckLoading.value = true;
  const requestHosts = originData.value.map((item: any) => ({
    ...(item.bk_host_id ? { bk_host_id: item.bk_host_id } : {}),
    bk_biz_id: Number(item.bk_biz_id),
    ...(item.bk_host_innerip ? { bk_host_innerip_list: item.bk_host_innerip.split(';').filter(Boolean) } : {}),
    ...(item.bk_host_innerip_v6 ? { bk_host_innerip_v6_list: item.bk_host_innerip_v6.split(';').filter(Boolean) } : {}),
    bk_networkunit_id: Number(item.bk_networkunit_id),
  }));
  const res = await NodeProxyService.NodeProxyInstallCheck({
    host: requestHosts,
  }).catch(() => ({
    results: [],
  }));
  installCheckLoading.value = false;
  if (res) {
    originData.value = originData.value.map((item: any, index: number) => {
      const find = res.results[index];
      return {
        ...item,
        ...find,
      };
    });
  }
  tableData.value = cloneDeep(originData.value);
};

watch(
  [searchSelectValue],
  () => {
    tableData.value = originData.value.filter((row: any) => searchSelectValue.value.every((searchItem: any) => {
      const { id: searchField, values } = searchItem;
      const searchIds = values?.map((value: { id: string }) => value.id);
      return searchIds.includes(row[searchField]);
    }));
  },
  { immediate: true, deep: true },
);
watch(
  () => tableData.value,
  () => {
    tabKey.value = Date.now();
  },
  { immediate: true },
);
watch(
  [active, originData],
  () => {
    tableData.value = originData.value.filter((item: any) => {
      if (active.value === 'all') {
        return true;
      }
      return item.category === active.value;
    });
  },
  { immediate: true },
);
watch(
  () => isShow,
  async () => {
    if (isShow.value && props.data) {
      originData.value = (props.data.hosts || []).map((item: any) => ({
        ...item,
        checked: false,
        __row_id: genRowID(),
      }));
      await installCheck();
    }
  },
  { immediate: true, deep: true },
);
</script>
<style lang="postcss" scoped>
:deep(.bk-tab-header) {
  background: #f0f1f5;
}
:deep(.bk-tab-content) {
  padding: 0;
}
</style>
<style lang="postcss">
.vxe-table--tooltip-wrapper {
  &.preview-table {
    z-index: 2004 !important;
  }
}
</style>
