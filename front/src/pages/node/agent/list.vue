<template>
  <div class="p-[24px]">
    <!-- agnet操作及搜索 -->
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[8px]">
        <Dropdown
          theme="light"
          trigger="click"
          placement="bottom-start"
          :popover-options="{
            clickContentAutoHide: true,
          }"
        >
          <Button class="w-[130px]" theme="primary" @click="handleInstall">{{
            $t("platform.nodeMan.installAgent")
          }}</Button>
          <template #content>
            <Dropdown.DropdownMenu ext-cls="dropDown-menu">
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in agentInstallType"
                :key="item.id"
                @click="triggerHandler('setup', item.id)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
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
                @click="handleOperate(item.id, selection, true)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <copy-ip-dropdown
          :type="'agent'"
          :disabled="!selection.length"
          :data="tableData"
          :list="[]"
        ></copy-ip-dropdown>
      </div>
      <div class="flex gap-[8px]">
        <!-- <Cascader
          class="w-[250px]"
          is-remote
          clearable
          v-model="topo"
          :list="topoBizFilterList"
          id-key="bk_biz_id"
          name-key="bk_biz_name"
          :remote-method="topoRemotehandler"
          ref="topoSelect"
          :placeholder="$t('platform.nodeMan.bussinessTopology')"
        /> -->
        <SearchSelect
          class="w-[480px] z-99"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('platform.nodeMan.agentSearchPlaceholder')"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <bk-loading
      :title="$t('table.loading')"
      :loading="loading"
      class="w-full overflow-auto"
    >
      <Table
        class="filterTable"
        :data="tableData"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
        <TableColumn
          field="bk_host_id"
          title="Host ID"
          :min-width="100"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip"
          :title="t('platform.nodeMan.inner_ip')"
          :min-width="150"
          fixed="left"
        >
          <template #default="{ row }">
            {{ row.bk_host_innerip_list?.join(',') }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_host_innerip_v6"
          :title="t('platform.nodeMan.inner_ipv6')"
          :min-width="150"
        >
          <template #default="{ row }">
            {{ row.bk_host_innerip_v6_list?.join(',') }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_agent_id"
          :title="t('platform.nodeMan.agentId')"
          :min-width="320"
        ></TableColumn>
        <TableColumn
          field="bk_networkarea_id"
          :title="t('platform.nodeMan.bk_cloud_name')"
          :filter="filterOptionSource.bk_networkarea_id"
          :min-width="120"
          show-overflow
        >
          <template #default="{ row }">
            {{ row.bk_networkarea_name }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="bk_networkunit_id"
          :title="t('platform.nodeMan.bk_cloud_unit')"
          :filter="filterOptionSource.bk_networkunit_id"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ row.bk_networkunit_name }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="os_type"
          :title="t('platform.nodeMan.os_type')"
          :filter="filterOptionSource.os_type"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          show-overflow
          field="node_version"
          :title="t('platform.nodeMan.agent_version')"
          :filter="filterOptionSource.node_version"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          field="node_status"
          :title="t('platform.nodeMan.status')"
          :min-width="150"
          :filter="filterOptionSource.node_status"
        >
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <span
                :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"
              ></span>
              <span>{{ row.node_status }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.node_status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="t('platform.nodeMan.operate')"
          :min-width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <Button
              theme="primary"
              text
              ext-cls="reinstall"
              @click="handleOperate('reinstall', [row])"
            >
              {{ $t("platform.nodeMan.agentStatus.button.reinstall") }}
            </Button>

            <Dropdown
              theme="light"
              trigger="click"
              :popover-options="{
                clickContentAutoHide: true,
              }">
              <Button class="ml-[15px]" text>
                <span class="nodeman-icon nc-more"></span>
              </Button>
              <template #content>
                <Dropdown.DropdownMenu>
                  <Dropdown.DropdownItem
                    v-for="item in operate"
                    :key="item.id"
                    v-show="getOperateShow(row, item)"
                    @click="handleOperate(item.id, [row])"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                </Dropdown.DropdownMenu>
              </template>
            </Dropdown>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
    <choose-version-dialog
      v-model:is-show="chooseVersionData.isShow"
      :title="chooseVersionData.title"
      :data="chooseVersionData.data"
      :batch="chooseVersionData.batch"
      @confirm="handleUpgrade"
    >
    </choose-version-dialog>
    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      @confirm="operateJob"
    ></operate-dialog>
  </div>
</template>
<script setup lang="ts">
import { Button, Cascader, Checkbox, Dropdown, InfoBox, Input, SearchSelect } from 'bkui-vue';
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { TopoHostDistinctRespData } from '@/@types/topo';
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import { NodeAgentService } from '@/api/modules/node_agent';
import { TopoService } from '@/api/modules/topo';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const tableData = ref<Host[]>([]);
const agentList = ref<Host[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const chooseVersionData = reactive({
  title: '',
  isShow: false,
  data: null,
  batch: false,
});
const operateDialogIsShow = ref(false);
const operateDialogData = {
  type: '',
  title: '',
  subTitle: '',
};
const searchSelectData = computed(() => [
  {
    id: 'bk_host_innerip',
    name: t('platform.nodeMan.inner_ip'),
    multiple: true,
  },
  {
    id: 'bk_host_innerip_v6',
    name: t('platform.nodeMan.inner_ipv6'),
    multiple: true,
  },
  { id: 'bk_agent_id', name: 'Agent ID', multiple: true },
  {
    id: 'bk_networkarea_id',
    name: '管控区域',
    children: getUniqueChildrenFrom('bk_networkarea_id', networkAreaListMap),
    multiple: true,
  },
  {
    id: 'bk_networkunit_id',
    name: '管控单元',
    children: getUniqueChildrenFrom('bk_networkunit_id', networkUnitListMap),
    multiple: true,
  },
  {
    id: 'os_type',
    name: '操作系统',
    children: getUniqueChildrenFrom('os_type'),
    multiple: true,
  },
  {
    id: 'node_version',
    name: 'Agent 版本',
    children: getUniqueChildrenFrom('node_version'),
    multiple: true,
  },
  {
    id: 'node_status',
    name: 'Agent 状态',
    children: getUniqueChildrenFrom('node_status'),
    multiple: true,
  },
]);
// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_agent_id',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'os_type',
    'node_version',
    'node_status',
    'action',
  ],
  disabled: ['action'],
}, 'nodeMng-agent');

const filterOptionSource: Record<string, FilterOption> = reactive({
  bk_networkarea_id: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  bk_networkunit_id: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  os_type: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  node_version: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  node_status: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
});
// 批量操作
const operate = [
  {
    id: 'reinstall',
    name: '重装',
    disabled: false,
    show: true,
  },
  {
    id: 'upgrade',
    name: '升级/回退',
    disabled: false,
    show: true,
  },
  {
    id: 'restart',
    name: '重启',
    disabled: false,
    show: true,
  },
  {
    id: 'uninstall',
    name: '卸载',
    disabled: false,
    show: true,
  },
];
// 安装方式
const agentInstallType = [
  {
    id: 'setup',
    name: '普通远程安装',
  },
  {
    id: 'import',
    name: 'Excel 导入远程安装',
  },
  {
    id: 'manual',
    name: '手动安装',
  },
];

// 后端分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const loading = ref(false);
// 拓扑级联选择器的选值
const topo = ref([]);
const topoBizFilterList = computed(() => mainStore.businessList);
const topoRemotehandler = () => {};
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  // 给筛选器添加选中值
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id]) {
      filterOptionSource[item.id].checked = item.values.map((item: any) => item.id);
    }
  });
};

const networkAreaListMap = new Map<number, string | number>([[-1, -1]]);
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
  res.items.forEach(item => {
    networkAreaListMap.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
};
// 管控单元下拉列表获取
const networkUnitListMap = new Map<number, string | number>([[-1, -1]]);
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
  res.items.forEach(item => {
    networkUnitListMap.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
};
const hostDistinct = ref<TopoHostDistinctRespData | null>();
// 筛选
const getHostDistinct = async () => {
  const params = {
    exact_include_conditions: {
      node_role: ['agent', 'blank'],
      bk_biz_id: mainStore.selectedBusinessId,
    },
  };
  const res = await TopoService.HostDistinct(params).catch(() => null);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      const curUniqueValues = res[key] || [];
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = curUniqueValues
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text;
            switch (key) {
              case 'bk_networkarea_id':
                text = networkAreaListMap.get(Number(value));
                break;
              case 'bk_networkunit_id':
                text = networkUnitListMap.get(Number(value));
                break;
              default:
                text = value;
                break;
            }
            return {
              text,
              value,
            };
          });
      }
    });
  }
};
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: t(field),
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'bk_networkarea_id':
            name = networkAreaListMap.get(Number(item));
            break;
          case 'bk_networkunit_id':
            name = networkUnitListMap.get(Number(item));
            break;
          default:
            name = item;
            break;
        }
        return {
          id: item,
          name,
        };
      }),
    });
  }
};
// 分页操作
const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  await getAgentList();
};
const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getAgentList();
};

const dropdownShow = ref(false);
const handleInstall = () => {
  if (selection.value.length) {
    dropdownShow.value = false;
    triggerHandler('reinstall');
  } else {
    dropdownShow.value = !dropdownShow.value;
  }
};

// 搜索
function getUniqueChildrenFrom <K extends keyof TopoHostDistinctRespData>(
  prop: K,
  keyMap?: Record<string, any>,
) {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap && keyMap.get(value) ? keyMap.get(value) : String(value),
    }));
};

const triggerHandler = (type: string, setupType = 'setup') => {
  switch (type) {
    // 重启 重装 重载配置 卸载 升级
    case 'restart':
    case 'reinstall':
    case 'uninstall':
    case 'upgrade':
      handleOperate(type, selection.value, true);
      break;
    case 'setup':
      router.push({ name: 'agentSetup' });
      mainStore.updateAgentSetupType(setupType);
      break;
  }
};
/**
 * 当前操作项是否显示
 */
const getOperateShow = (row: Host, config: any) => {
  if (config.id === 'reinstall') {
    return false;
  }
  return config.show;
};

// 操作
const handleOperate = (type: string, data: Host[], batch = false) => {
  let jobType = '';

  switch (type) {
    // 重启
    case 'restart':
      handleOperatetHost(data, batch, 'restart');
      break;
    // 重装
    case 'reinstall':
      jobType = 'reinstall';
      break;
    // 卸载
    case 'uninstall':
      handleOperatetHost(data, batch, 'uninstall');
      break;
    // 升级
    case 'upgrade':
      handleOperatetHost(data, batch, 'upgrade');
      break;
  }
  if (!jobType) return;

  router.push({ name: 'agentEdit' });
  const params = {
    tableData: data.map((item: any) => ({
      ...item,
    })),
    type: jobType,
  };
  nodeManageStore.updateAgentEditRowData(params);
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

const operateData = ref<Host[]>();
// 重启
const operateJob = async (extraData: any = {}) => {
  loading.value = true;
  const params = {
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      force: extraData.isForce,
      graceful_restart_timeout_sec: extraData.time,
    })),
  };
  let result;
  if (extraData.isReconfig) {
    result = await NodeAgentService.NodeAgentReconfig(params).catch(() => ({
      workflow_id: '',
    }));
  } else {
    result = await NodeAgentService.NodeAgentRestart(params).catch(() => ({
      workflow_id: '',
    }));
  }
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};
// 升级回退
const handleUpgrade = async (osVersion: any[]) => {
  loading.value = true;
  const params = {
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      target_version: osVersion[0].version,
    })),
  };
  const result = await NodeAgentService.NodeAgentUpgrade(params).catch(() => ({
    workflow_id: '',
  }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};

// 卸载
const handleUninstall = async () => {
  loading.value = true;
  const result = await NodeAgentService.NodeAgentUninstall({
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
    })),
  }).catch(() => ({
    workflow_id: '',
  }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};

/**
 * Agent操作
 * @param {String} type 操作类型
 * @param {Array} data agent数据
 * @param {Boolean} batch 是否是批量操作
 */

/**
 * @param {Array} data
 */
const handleOperatetHost = async (
  data: Host[],
  batch: boolean,
  operateType: string,
) => {
  const titleObj = {
    firstIp: data[0].info.bk_host_innerip_list.join(','),
    num: data.length,
  };
  let type = '';
  switch (operateType) {
    // 重启
    case 'restart':
      type = '重启';
      break;
    // 升级
    case 'upgrade':
      type = '升级/回退';
      break;
    case 'uninstall':
      type = '卸载';
      break;
  }
  operateData.value = data;
  if (operateType === 'upgrade') {
    chooseVersionData.title = 'Agent 升级/回退';
    chooseVersionData.isShow = true;
    chooseVersionData.data = data;
    chooseVersionData.batch = batch;
  } else if (operateType === 'restart') {
    operateDialogIsShow.value = true;
    operateDialogData.type = operateType;
    operateDialogData.title = batch ? `请确认是否批量${type}` : `请确认是否${type}`;
    operateDialogData.subTitle = batch
      ? `${type} ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
      : `${type} ${titleObj.firstIp} 的Agent`;
  } else if (operateType === 'uninstall') {
    InfoBox({
      title: batch ? `请确认是否批量${type}` : `请确认是否${type}`,
      subTitle: batch
        ? `${type} ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
        : `${type} ${titleObj.firstIp} 的Agent`,
      onConfirm: () => {
        handleUninstall();
      },
    });
  }
};

const fuzzyKeys = new Set([
  'bk_host_innerip',
  'bk_host_innerip_v6',
  'bk_host_name',
  'dept_name',
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      bk_biz_id: mainStore.selectedBusinessId,
      node_role: ['agent', 'blank'],
    } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getAgentList = async () => {
  loading.value = true;
  const res = await TopoService.HostList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;
  tableData.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
  }));
  agentList.value = tableData.value;
  loading.value = false;
};
watch(
  searchSelectValue,
  async () => {
    await getAgentList();
  },
  { deep: true },
);
watch(route, async () => {
  if (route.query.os_type) {
    searchSelectValue.value.push(...[
      {
        id: 'os_type',
        name: '操作系统',
        values: [{
          id: route.query.os_type,
          name: route.query.os_type,
        }],
      },
      {
        id: 'cpu_arch',
        name: '架构',
        values: [{
          id: route.query.cpu_arch,
          name: route.query.cpu_arch,
        }],
      },
      {
        id: 'node_version',
        name: 'Agent版本',
        values: [{
          id: route.query.node_version,
          name: route.query.node_version,
        }],
      },
    ]);
    await getAgentList();
  }
}, { immediate: true, deep: true });
watch(() => mainStore.selectedBusinessId, async () => {
  await getAgentList();
}, { immediate: true });
onMounted(async () => {
  await getNetworkAreaList();
  await getNetworkUnitList();
  await getHostDistinct();
});
</script>
<style lang="postcss" scoped>
.dropDown-menu {
  .bk-dropdown-item {
    font-size: 14px;
  }
}
.bk-cascader-wrapper {
  width: 250px;

  .bk-cascader-wrapper:not(:last-of-type) {
    margin-bottom: 20px;
  }
}

.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 14px;
  height: 14px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
  vertical-align: middle;
}
.nc-running {
  &::before {
    background: #3fc06d;
    border-color: #e5f6ea;
  }
}
.nc-terminated {
  &::before {
    border-color: #ffe6e6;
    background: #ea3636;
  }
}
.nc-unknown {
  &::before {
    border-color: #f0f1f5;
    background: #b2b5bd;
  }
}
</style>
