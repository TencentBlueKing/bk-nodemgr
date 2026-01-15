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
            <p>{{ $t("platform.nodeMan.preview.thirdTip") }}</p>
            <p>{{ $t("platform.nodeMan.preview.fourthTip") }}</p>
          </div>
        </div>
        <div class="flex justify-between mt-[16px]">
          <div class="w-[50%] flex gap-[8px]">
            <SearchSelect
              class="flex-1"
              ref="searchSelect"
              :data="searchSelectData"
              v-model.trim="searchSelectValue"
              :unique-select="true"
              :placeholder="'IPV4、IPV6、操作系统、主机名'"
            >
            </SearchSelect>
            <copy-ip-dropdown
              :type="'agent'"
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
                content: '处理待确认为正常安装',
              }"
            >
              {{ $t("platform.nodeMan.preview.button.batchConfirm") }}
            </Button>
            <Button
              @click="handleBatchRemove"
              :disabled="
                !selection.length ||
                  !selection.find((item) => item.category === 'error')
              "
            >
              {{ $t("platform.nodeMan.preview.button.batchRemove") }}
            </Button>
            <Dropdown
              theme="light"
              trigger="click"
              :popover-options="{
                clickContentAutoHide: true,
              }">
              <Button :disabled="!selection.length">
                <span>{{ $t("platform.nodeMan.batchOperate") }}</span>
                <i
                  class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
                ></i>
              </Button>
              <template #content>
                <Dropdown.DropdownMenu>
                  <Dropdown.DropdownItem
                    v-for="item in operate"
                    :key="item.id"
                    @click="handleOperate(item.id)"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                </Dropdown.DropdownMenu>
              </template>
            </Dropdown>
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
                  title="管控单元"
                  min-width="150"
                ></TableColumn>
                <TableColumn
                  field="status"
                  title="状态"
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
                      <p>{{ isZh ? row.message_zh : row.message_en }}</p>
                      <Button
                        v-show="row.category === 'need_confirm'"
                        theme="primary"
                        text
                        @click="deel(row)"
                      >
                        确认
                      </Button>
                    </div>
                  </template>
                </TableColumn>
                <TableColumn
                  field="action"
                  :title="t('platform.nodeMan.operate')"
                  min-width="100"
                  fixed="right"
                >
                  <template #default="{ row }">
                    <Button theme="primary" text @click="handleRemove(row)">{{
                      t("platform.nodeMan.preview.button.remove")
                    }}</Button>
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
              ? '有节点待确认或错误，不可安装'
              : installCheckLoading ? '校验安装节点中...' : '校验失败，不可安装',
            disabled: disabledDataNum === 0 && !checkFailed,
          }"
        >执行安装全部节点</Button
        >
        <Button @click="handleBeforeClose">{{ t("action.cancel") }}</Button>
      </div>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import {
  Button,
  Dropdown,
  InfoBox,
  PopConfirm,
  Radio,
  SearchSelect,
  Sideslider,
  Tab,
} from 'bkui-vue';
import { Close, Spinner } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { AgentInstallInfo } from '@/@types/node_agent.d';
import { NodeAgentService } from '@/api/modules/node_agent';
import { TopoService } from '@/api/modules/topo';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps({
  data: {
    type: Object,
    default: () => {},
  },
  isManual: {
    type: Boolean,
    default: false,
  },
});

const { t } = useI18n();
const mainStore = useMainStore();
const router = useRouter();
const originData = ref<AgentInstallInfo[]>([]);
const tableData = ref<AgentInstallInfo[]>([]);
const tabKey = ref(Date.now());
const disabledDataNum = ref(0);
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');
const checkFailed = computed(() => tableData.value.some(item => !item.status));
const { pagination } = usePage(tableData);

// 批量
const operate = ref([
  {
    id: 'comfirm',
    match: 'need_confirm',
    name: '确认',
  },
  {
    id: 'remove',
    match: 'error',
    name: '移除',
  }
]);
const handleOperate = async (id: string) => {
  if (id === 'comfirm') {
    const findItem = selection.value.filter((item: any) => ['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(item.status));
    const otherItemIps = selection.value
      .filter((item: any) => item.status !== 'duplicated_inner_ip' && item.status !== 'duplicated_inner_ipv6')
      .map((item: any) => item.bk_host_innerip);
    if (findItem) {
      findItem.forEach((item: any) => {
        item.bk_host_id = item.matched.bk_host_id;
      });
      await installCheck();
    }
    tableData.value.forEach((item: any) => {
      if (otherItemIps.includes(item.bk_host_innerip) && item.category === 'need_confirm') {
        item.category = 'normal_install';
      }
    });
  } else if (id === 'remove') {
    originData.value = originData.value.filter(item => !item.checked);
  }
  tabKey.value = Date.now();
};
// 搜索
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
  {
    id: 'bk_host_name',
    name: '主机名',
    children: getUniqueChildren('bk_host_name'),
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
      label: t('platform.nodeMan.preview.label.cleanInstallation'),
      name: 'register_to_cmdb_and_install',
      count: countResult.cleanInstallation,
      icon: 'check-circle-fill',
      iconColor: '#1CAB88',
    },
    {
      label: t('platform.nodeMan.preview.label.normalInstallation'),
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
      'bk_agent_id',
      'bk_host_name',
      'bk_networkarea_name',
      'bk_networkarea_id',
      'bk_networkunit_name',
      'os_type',
      'node_version',
      'status',
      'action',
    ],
    disabled: ['action'],
  },
  'nodeMng-preview',
);
// const statusMap = {
//   register_to_cmdb_and_install: {
//     text: '可执行：全新安装并导入 CMDB',
//     icon: 'check-circle-fill',
//     iconColor: '#1CAB88',
//   },
//   normal_install: {
//     text: '可执行：正常安装',
//     icon: 'check-circle-fill',
//     iconColor: '#1CAB88',
//   },
//   exist_proxy: {
//     text: '该主机已安装 proxy，若继续安装该 Proxy 将会被覆盖',
//     icon: 'wrong',
//     iconColor: '#EA3636',
//   },
//   exist_agent: {
//     text: '该 IP 主机在当前管控区域已安装 Agent，无法重复安装',
//     icon: 'wrong',
//     iconColor: '#EA3636',
//   },
//   not_exist_relay: {
//     text: '不存在可用于安装agent的中继节点',
//     icon: 'wrong',
//     iconColor: '#EA3636',
//   },
//   conflict_ip: {
//     text: '当前业务下可能已存在您希望安装的相同 IP 主机',
//     icon: 'danger-fill',
//     iconColor: '#FF9C01',
//   },
//   duplicate_dynamic_ip: {
//     text: '动态寻址下已存在相同 IP 的安装记录',
//     icon: 'danger-fill',
//     iconColor: '#FF9C01',
//   },
// };
const categoryMap = {
  normal_install: {
    // text: '可执行：正常安装',
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  register_to_cmdb_and_install: {
    // text: '可执行：全新安装并导入 CMDB',
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  error: {
    // text: '错误',
    icon: 'wrong',
    iconColor: '#EA3636',
  },
  need_confirm: {
    // text: '待确认',
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

const handleRemove = (row: AgentInstallInfo) => {
  originData.value = originData.value.filter((item: any) => item.bk_host_innerip !== row.bk_host_innerip);
  tabKey.value = Date.now();
};

// 处理待确认
const deel = async (row: any) => {
  if (['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(row.status)) {
    row.bk_host_id = row.matched.bk_host_id;
    await installCheck();
  } else {
    row.category = 'normal_install';
    tableData.value = [...tableData.value];
  }
};
// 表格勾选
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

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  tableData.value.forEach((item: any) => (item.checked = checked));
};
const handleAllConfirm = async () => {
  const findItem = tableData.value.filter((item: any) => ['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(item.status));
  if (findItem) {
    findItem.forEach((item: any) => {
      item.bk_host_id = item.matched.bk_host_id;
    });
    await installCheck();
  }
  tableData.value.forEach((item: any) => {
    if (['duplicated_inner_ip', 'duplicated_inner_ipv6'].includes(item.status)) {
      item.category = 'normal_install';
    }
  });
  tabKey.value = Date.now();
};
const handleBatchRemove = () => {
  originData.value = originData.value.filter((item: any) => !(item.category == 'error' && item.checked));
  tabKey.value = Date.now();
};

const loading = ref(false);
const handleSetup = async () => {
  loading.value = true;
  // 过滤掉每条数据中的duplicate_host_ids和elig_status字段
  const filteredData = tableData.value.map((item: any) => {
    const { pending_host_ids, status, bk_networkunit_id, ...rest } = item;
    return {
      ...rest,
      bk_networkunit_id: Number(bk_networkunit_id),
    };
  });
  const res = await NodeAgentService.NodeAgentInstall({
    info: filteredData,
    target_version: props.data.target_version,
    disable_default_target_version: props.data.disable_default_target_version,
    is_manual: props.isManual,
  }).catch(() => ({
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
  }
};
// 定义排序优先级
const priority = {
  conflict_ip: 0,
  duplicate_dynamic_ip: 0,
  exist_proxy: 1,
  exist_agent: 1,
  not_exist_relay: 1,
  register_to_cmdb_and_install: 2,
  normal_install: 3,
};

// 排序函数
function sortByEligStatus(arr: any[]) {
  // 复制原数组避免修改原数组
  return [...arr].sort((a, b) => {
    // 获取当前元素的优先级，默认最低
    const priorityA = priority[a.status] ?? 6;
    const priorityB = priority[b.status] ?? 6;
    // 按优先级升序排列（数值越小优先级越高）
    return priorityA - priorityB;
  });
}

// 安装检查
const installCheckLoading = ref(false);
const installCheck = async () => {
  installCheckLoading.value = true;
  const res = await NodeAgentService.NodeAgentInstallCheck({
    host: originData.value.map((item: any) => ({
      ...(item.bk_host_id ? { bk_host_id: item.bk_host_id } : {}),
      bk_biz_id: Number(item.bk_biz_id),
      ...(item.bk_host_innerip ? { bk_host_innerip_list: item.bk_host_innerip.split(';') } : {}),
      ...(item.bk_host_innerip_v6 ? { bk_host_innerip_v6_list: item.bk_host_innerip_v6.split(';') } : {}),
      bk_networkunit_id: Number(item.bk_networkunit_id),
    })),
  }).catch(() => ({
    results: [],
  }));
  installCheckLoading.value = false;
  if (res) {
    originData.value = originData.value.map((item: any, originIndex: number) => {
      const find = res.results.find((result: any, index: number) =>
        result.matched?.bk_host_id === item.bk_host_id
          || item.bk_host_innerip.includes(result.matched?.bk_host_innerip_list[0])
          || item.bk_host_innerip_v6.includes(result.matched?.bk_host_innerip_v6_list[0])
          || originIndex === index);
      return {
        ...item,
        ...find,
      };
    });
  }
  tableData.value = cloneDeep(originData.value);
};

const deelTabelData = ref<any[]>([]);
const getAgentList = async (row: any) => {
  const res = await TopoService.HostList({
    page: { limit: 500, offset: 0 },
    exact_include_conditions: {
      bk_host_id: [...row.pending_host_ids],
    },
  }).catch(err => ({
    total: 0,
    items: [],
  }));
  deelTabelData.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
  }));
};
function isEmpty(str: string | number | undefined | null) {
  return str === undefined || str === null || str === '';
}
// 前端过滤数据
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
    // 每次tableData变化时，重新生成一个key，以强制刷新tab
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
      originData.value = props.data.info.map(item => ({
        ...item,
        login_port: Number(item.login_port),
        bk_biz_id: isEmpty(item.bk_biz_id)
          ? Number(props.data.bk_biz_id)
          : item.bk_biz_id,
        bk_networkunit_id: isEmpty(item.bk_networkunit_id)
          ? Number(props.data.bk_networkunit_id)
          : item.bk_networkunit_id,
        bk_host_id: Number(item.bk_host_id),
        bk_networkarea_name:
          item.bk_networkarea_name || props.data.bk_networkarea_name,
        bk_networkunit_name:
          item.bk_networkunit_name || props.data.bk_networkunit_name,
        bk_addressing: 'static',
        checked: false,
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
