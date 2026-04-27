<template>
  <page-header :title="'route.agentStatus'" :back="Object.keys(route.query).length > 0"></page-header>
  <div class="p-[24px]">
    <!-- agnet操作及搜索 -->
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[8px]">
        <!-- <Dropdown
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
        </Dropdown> -->
        <!-- 主按钮（安装/重装）：始终可点击，无需权限校验 -->
        <Button
          class="w-[130px]"
          theme="primary"
          :class="{ 'btn-reinstall': hasSelection }"
          @click="handlePrimaryButtonClick">
          {{ primaryButtonLabel }}
        </Button>
        <!-- 批量操作：无权限时置灰 + hover 带锁 + 点击申请权限 -->
        <Dropdown
          v-if="hasOperateAuth"
          theme="light"
          trigger="click"
          :popover-options="{
            clickContentAutoHide: true,
          }">
          <Button :disabled="!selection.length" :loading="crossPageSelectLoading">
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
                :disabled="item.disabled"
                :class="{ 'operate-item-disabled': item.disabled }"
                v-bk-tooltips="{
                  content: item.tooltip,
                  disabled: !item.disabled,
                }"
                @click.stop="!item.disabled && handleOperate(item.id, selection, true)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <span
          v-else
          class="inline-flex items-center auth-lock-wrapper"
          @click="handleAuthClick"
          @mouseenter="authLockMouseEnter($event, false)"
          @mousemove="authLockMouseMove($event, false)"
          @mouseleave="authLockMouseLeave()"
        >
          <Button class="auth-disabled-btn">
            <span>{{ $t("platform.nodeMan.batchOperate") }}</span>
            <i
              class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
            ></i>
          </Button>
        </span>
        <!-- 复制IP：无权限时置灰 + hover 带锁 + 点击申请权限 -->
        <template v-if="hasOperateAuth">
          <copy-ip-dropdown
            :type="'agent'"
            :disabled="!hasSelection"
            :data="tableData"
            :list="[]"
            :is-cross-page-selection="isCrossPageSelection"
            :cross-page-query-params="crossPageQueryParams"
          ></copy-ip-dropdown>
        </template>
        <span
          v-else
          class="inline-flex items-center auth-lock-wrapper"
          @click="handleAuthClick"
          @mouseenter="authLockMouseEnter($event, false)"
          @mousemove="authLockMouseMove($event, false)"
          @mouseleave="authLockMouseLeave()"
        >
          <Button class="auth-disabled-btn">
            <span>{{ $t('components.copyIpDropdown.copy') }}</span>
            <i
              class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
            ></i>
          </Button>
        </span>
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
          class="w-[480px] z-99 bg-[#fff]"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('platform.nodeMan.agentSearchPlaceholder')"
          @update:model-value="handleSearchSelectChange"
          @paste.native="handleNativePaste"
        >
        </SearchSelect>
      </div>
    </section>
    <Loading
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
        <template #prepend>
          <div v-if="hasSelection" class="flex items-center justify-center h-[30px] bg-[#ebecf0] text-[12px]">
            <template v-if="isCrossPageSelection">
              <span>{{ $t('taskDetail.table.crossPageSelected') }} </span>
              <span class="font-bold mx-1"> {{ total - excludedIds.size }} </span>
              <span>{{ $t('taskDetail.table.items') }}</span>
              <Button text theme="primary" @click="handleClearSelection">
                {{ $t('taskDetail.table.cancelSelection') }}
              </Button>
            </template>
            <template v-else>
              <span>{{ $t('taskDetail.table.selected') }}</span>
              <span class="font-bold mx-1"> {{ selection.length }} </span>
              <span>{{ $t('taskDetail.table.items') }}</span>
              <Button
                v-if="total > pagination.limit"
                text theme="primary" @click="handleSelectAllCrossPage">
                {{ $t('taskDetail.table.selectAllPages', { x: total }) }}
              </Button>
              <Button v-else text theme="primary" @click="handleClearSelection">
                {{ $t('taskDetail.table.cancelSelection') }}
              </Button>
            </template>
          </div>
        </template>

        <TableColumn width="80" fixed="left">
          <template #header>
            <Button text class="flex items-center justify-start">
              <Checkbox
                :model-value="isCurrentPageAllChecked"
                :indeterminate="isIndeterminate"
                @change="handleHeaderClick" />
              <Dropdown trigger="click" placement="bottom-start">
                <i class="nodeman-icon nc-arrow-down ml-1 text-[18px]"></i>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem @click="handleSelectCurrentPage">
                      {{ $t('taskDetail.filter.currentPage') }}
                    </Dropdown.DropdownItem>
                    <Dropdown.DropdownItem @click="handleSelectAllCrossPage">
                      <Button text :disabled="total <= pagination.limit">
                        {{ $t('taskDetail.filter.crossSelected') }}
                      </Button>
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </Button>
          </template>
          <template #default="{ row }">
            <Checkbox class="mt-[6px]" :model-value="row.checked" @change="(val) => handleRowCheck(val, row)" />
          </template>
        </TableColumn>
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
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip_v6"
          :title="t('platform.nodeMan.inner_ipv6')"
          :min-width="150"
        ></TableColumn>
        <TableColumn
          field="bk_agent_id"
          :title="t('platform.nodeMan.agentId')"
          :min-width="320"
        ></TableColumn>
        <TableColumn
          field="bk_biz_id"
          :title="t('platform.nodeMan.bk_biz_id')"
          :filter="filterOptionSource.bk_biz_id"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ bizListMap.get(row.bk_biz_id) || row.bk_biz_id }}
          </template>
        </TableColumn>
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
            {{ networkUnitListMap.get(row.bk_networkunit_id) || row.bk_networkunit_name }}
          </template>
        </TableColumn>
        <TableColumn
          field="dept_name"
          :title="t('platform.nodeMan.dept_name')"
          :filter="filterOptionSource.dept_name"
          :min-width="120"
        ></TableColumn>
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
              <span>{{ statusMap.get(row.node_status) || row.node_status }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span>--</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          :title="t('platform.nodeMan.installAgentPage.pluginNum')"
          field="pluginNum"
          :min-width="122"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.pluginNum || 0 }}
              <i class="nodeman-icon nc-plug-in ml-[5px]"></i>
            </Button>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="t('platform.nodeMan.operate')"
          :min-width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <!-- 重装按钮：有权限正常点击，无权限灰色+锁hover+点击申请权限 -->
            <Button
              v-if="hasOperateAuth"
              theme="primary"
              text
              ext-cls="reinstall"
              @click="handleOperate('reinstall', [row])"
            >
              {{ $t("platform.nodeMan.agentStatus.reinstall") }}
            </Button>
            <span
              v-else
              class="inline-flex items-center auth-lock-wrapper"
              @click="handleAuthClick"
              @mouseenter="authLockMouseEnter($event, false)"
              @mousemove="authLockMouseMove($event, false)"
              @mouseleave="authLockMouseLeave()"
            >
              <Button theme="primary" text class="auth-disabled-text-btn">
                {{ $t("platform.nodeMan.agentStatus.reinstall") }}
              </Button>
            </span>

            <!-- 更多操作下拉：始终可点击打开，无权限时 Dropdown 内的 item 全部置灰+hover带锁+点击申请权限 -->
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
                  <!-- 有权限的操作项：正常显示和交互 -->
                  <Dropdown.DropdownItem
                    v-for="item in operate"
                    :key="item.id"
                    v-show="getOperateShow(row, item) && getItemHasAuth(item)"
                    :disabled="getRowOperateDisabled(row, item).disabled"
                    :class="{ 'operate-item-disabled': getRowOperateDisabled(row, item).disabled }"
                    v-bk-tooltips="{
                      content: getRowOperateDisabled(row, item).tooltip,
                      disabled: !getRowOperateDisabled(row, item).disabled,
                    }"
                    @click.stop="!getRowOperateDisabled(row, item).disabled && handleOperate(item.id, [row])"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                  <!-- 无权限的操作项：置灰+hover带锁+点击申请权限 -->
                  <Dropdown.DropdownItem
                    v-for="item in operate.filter(i => getOperateShow(row, i) && !getItemHasAuth(i))"
                    :key="'noauth-' + item.id"
                    class="auth-lock-dropdown-item"
                    @mouseenter="getItemAuthMouseEnter(item)($event)"
                    @mousemove="getItemAuthMouseMove(item)($event)"
                    @mouseleave="authLockMouseLeave()"
                    @click.stop="getItemAuthClick(item)()"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                </Dropdown.DropdownMenu>
              </template>
            </Dropdown>
          </template>
        </TableColumn>
      </Table>
    </Loading>

    <bk-footer></bk-footer>

    <upgrade-sideslider
      v-model:is-show="upgradePreviewData.isShow"
      :hosts="upgradePreviewData.hosts"
      release-type="agent"
    />

    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      @confirm="operateJob"
    ></operate-dialog>

    <!-- 侧边栏 -->
    <processSideslider
      v-model:is-show="isShowSideslider"
      type="node"
      node-type="agent"
      :node="currentAgent"
    ></processSideslider>
  </div>
</template>
<script setup lang="ts">
import { Button, Checkbox, Dropdown, InfoBox, Loading, SearchSelect } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, nextTick, onBeforeMount, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from '../plugin/process-sideslider.vue';
import UpgradeSideslider from './upgrade-sideslider.vue';

import type { TopoHostDistinctRespData } from '@/@types/topo';
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import { NodeAgentService } from '@/api/modules/node_agent';
import { ProcessAPIService } from '@/api/modules/process';
import { TopoService } from '@/api/modules/topo';
import { isNetworkUnitAssigned } from '@/common/const';
import useAuthLock from '@/composables/use-auth-lock';
import useTableSetting from '@/composables/use-table-setting';
import BkFooter from '@/pages/app/footer.vue';
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

// ===== agent_operate 权限控制（批量操作、复制IP、行内重装/更多操作）=====
const {
  hasAuth: hasOperateAuth,
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
  handleAuthClick,
} = useAuthLock('agent_operate', () => mainStore.selectedBusinessId);

// 注：「分配管控单元」不再做前端权限校验，由后端按需返回权限错误
const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;
const AGENT_ID_REG = /^0[12]/; // AgentID 以 01 或 02 开头
const AREA_IP_REG = /^(\d+):(.+)$/; // 管控区域ID:IP 格式

// ---------- 响应式数据 ----------
const tableData = ref<Host[]>([]);
const agentList = ref<Host[]>([]);
const loading = ref(false);
// 【优化点1】标记基础数据（区域、单元列表）是否加载完成
const isInitialDataLoaded = ref(false);

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });

// 搜索和筛选
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const filterOptionSource: Record<string, FilterOption> = reactive({
  bk_biz_id: { list: [], checked: [], filterScope: 'all' },
  bk_networkarea_id: { list: [], checked: [], filterScope: 'all' },
  bk_networkunit_id: { list: [], checked: [], filterScope: 'all' },
  dept_name: { list: [], checked: [], filterScope: 'all' },
  os_type: { list: [], checked: [], filterScope: 'all' },
  node_version: { list: [], checked: [], filterScope: 'all' },
  node_status: { list: [], checked: [], filterScope: 'all' },
});

// 其他UI相关响应式数据
const chooseVersionData = reactive({ title: '', isShow: false, data: [], batch: false });
const operateDialogIsShow = ref(false);
const operateDialogData = { type: '', title: '', subTitle: '' };
const dropdownShow = ref(false);
const operateData = ref<Host[]>([]);
const topo = ref([]);
const upgradePreviewData = reactive({ isShow: false, hosts: [] as Host[] });

const assignUnitDisabledState = computed(() => {
  const selected = selection.value;
  if (selected.length === 0) return { disabled: false, tooltip: '' };

  const hasAssigned = selected.some((h: any) => isNetworkUnitAssigned(h.bk_networkunit_id));
  if (hasAssigned) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.assignUnitDisabledAssigned'),
    };
  }

  const areaIds = new Set(selected.map((h: any) => h.bk_networkarea_id));
  if (areaIds.size > 1) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.assignUnitDisabledMultiArea'),
    };
  }

  return { disabled: false, tooltip: '' };
});

// 升级操作禁用状态：只检查 Agent 状态
const upgradeDisabledState = computed(() => {
  const selected = selection.value;
  if (selected.length === 0) return { disabled: false, tooltip: '' };

  const hasNonRunning = selected.some((h: any) => h.node_status !== 'running');
  if (hasNonRunning) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledNotRunning'),
    };
  }

  const hasUnassigned = selected.some((h: any) => !isNetworkUnitAssigned(h.bk_networkunit_id));
  if (hasUnassigned) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledUnassigned'),
    };
  }

  return { disabled: false, tooltip: '' };
});

// 重启操作禁用状态：检查 Agent 状态 + 管控单元
const restartDisabledState = computed(() => {
  const selected = selection.value;
  if (selected.length === 0) return { disabled: false, tooltip: '' };

  const hasNonRunning = selected.some((h: any) => h.node_status !== 'running');
  if (hasNonRunning) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledNotRunning'),
    };
  }

  const hasUnassigned = selected.some((h: any) => !isNetworkUnitAssigned(h.bk_networkunit_id));
  if (hasUnassigned) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledUnassigned'),
    };
  }

  return { disabled: false, tooltip: '' };
});

// 卸载操作禁用状态：检查管控单元
const uninstallDisabledState = computed(() => {
  const selected = selection.value;
  if (selected.length === 0) return { disabled: false, tooltip: '' };

  const hasUnassigned = selected.some((h: any) => !isNetworkUnitAssigned(h.bk_networkunit_id));
  if (hasUnassigned) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledUnassigned'),
    };
  }

  return { disabled: false, tooltip: '' };
});

// ---------- 常量定义 ----------
const topoBizFilterList = computed(() => mainStore.businessList);
const operate = computed(() => [
  { id: 'reinstall', name: t('platform.nodeMan.agentStatus.reinstall'), disabled: false, tooltip: '', show: true },
  { id: 'upgrade', name: t('platform.nodeMan.agentStatus.upgrade'), disabled: upgradeDisabledState.value.disabled, tooltip: upgradeDisabledState.value.tooltip, show: true },
  { id: 'restart', name: t('platform.nodeMan.agentStatus.restart'), disabled: restartDisabledState.value.disabled, tooltip: restartDisabledState.value.tooltip, show: true },
  { id: 'uninstall', name: t('platform.nodeMan.agentStatus.uninstall'), disabled: uninstallDisabledState.value.disabled, tooltip: uninstallDisabledState.value.tooltip, show: true },
  { id: 'assign_unit', name: t('platform.nodeMan.agentStatus.assignUnit'), disabled: assignUnitDisabledState.value.disabled, tooltip: assignUnitDisabledState.value.tooltip, show: true },
]);
const agentInstallType = [
  { id: 'setup', name: '普通远程安装' },
  { id: 'import', name: 'Excel 导入远程安装' },
  { id: 'manual', name: '手动安装' },
];

const statusMap = computed(() => new Map<string, string>([
  ['init', t('platform.nodeMan.agentStatus.init')],
  ['running', t('platform.nodeMan.agentStatus.running')],
  ['damaged', t('platform.nodeMan.agentStatus.damaged')],
  ['unknown', t('platform.nodeMan.agentStatus.unknown')],
]));

const fuzzyKeys = new Set(['bk_host_innerip', 'bk_host_innerip_v6', 'bk_host_name', 'dept_name']);

// ---------- 计算属性 ----------
const maxHeight = computed(() => mainStore.windowInnerHeight - 264 - (mainStore.noticeShow ? 40 : 0));
const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const total = computed(() => pagination.count);
// eslint-disable-next-line max-len
const bizListMap = computed(() => new Map<number, string>(mainStore.businessList.map((item: any) => [item.bk_biz_id, item.bk_biz_name])));
const networkAreaListMap = ref(new Map<number, string>([]));
const networkUnitListMap = ref(new Map<number, string>([]));
const hostDistinct = ref<TopoHostDistinctRespData | null>();

const searchSelectData = computed(() => [
  { id: 'ip', name: 'IP', multiple: true }, // 合并后的 IP 筛选
  { id: 'area_ip', name: t('platform.nodeMan.bk_cloud_name') + 'ID:IP', multiple: true }, // 管控区域ID:IP
  { id: 'bk_agent_id', name: 'Agent ID', multiple: true },
  {
    id: 'bk_biz_id',
    name: t('platform.nodeMan.bk_biz_id'),
    children: getUniqueChildrenFrom('bk_biz_id', bizListMap.value),
    multiple: true,
  },
  {
    id: 'bk_networkarea_id',
    name: t('platform.nodeMan.bk_cloud_name'),
    children: getUniqueChildrenFrom('bk_networkarea_id', networkAreaListMap.value),
    multiple: true,
  },
  {
    id: 'bk_networkunit_id',
    name: t('platform.nodeMan.bk_cloud_unit'),
    children: getUniqueChildrenFrom('bk_networkunit_id', networkUnitListMap.value),
    multiple: true,
  },
  {
    id: 'dept_name',
    name: t('platform.nodeMan.dept_name'),
    children: getUniqueChildrenFrom('dept_name'),
    multiple: true,
  },
  {
    id: 'os_type',
    name: t('platform.nodeMan.os_type'),
    children: getUniqueChildrenFrom('os_type'),
    multiple: true,
  },
  {
    id: 'node_version',
    name: t('platform.nodeMan.agent_version'),
    children: getUniqueChildrenFrom('node_version'),
    multiple: true,
  },
  {
    id: 'node_status',
    name: t('platform.nodeMan.status'),
    children: getUniqueChildrenFrom('node_status', statusMap.value),
    multiple: true,
  },
]);

// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_biz_id',
    'dept_name',
    'bk_agent_id',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'os_type',
    'node_version',
    'node_status',
    'pluginNum',
    'action',
  ],
  disabled: ['action'],
}, 'nodeMng-agent');

// --- 跨页全选核心状态 ---
const isCrossPageSelection = ref(false); // 是否开启跨页全选模式
const excludedIds = ref<Set<number>>(new Set()); // 全选模式下，用户手动“取消勾选”的 ID 集合
const crossPageQueryParams = computed(() => ({
  exact_include_conditions: getParams().exact_include_conditions,
  fuzzy_include_conditions: getParams().fuzzy_include_conditions,
  exact_exclude_conditions: {
    bk_host_id: [...excludedIds.value],
  },
}));

// 计算属性：是否有任何选中（用于禁用批量按钮）
const hasSelection = computed(() => tableData.value.some(item => item.checked) || isCrossPageSelection.value);

// 计算属性：主按钮文案（安装/重装动态切换）
const primaryButtonLabel = computed(() => (hasSelection.value ? t('platform.nodeMan.reinstallAgent') : t('platform.nodeMan.installAgent')));

// 主按钮点击事件：根据勾选状态分支调用安装或重装
const handlePrimaryButtonClick = () => {
  if (hasSelection.value) {
    handleOperate('reinstall', selection.value, true);
  } else {
    triggerHandler('setup');
  }
};

// 计算属性：当前页是否全选（用于表头 Checkbox 状态）
// eslint-disable-next-line max-len
const isCurrentPageAllChecked = computed(() => tableData.value.length > 0 && tableData.value.every(item => item.checked));
const isIndeterminate = computed(() => {
  const selectedCount = tableData.value.filter(item => item.checked).length;
  return selectedCount > 0 && selectedCount < tableData.value.length;
});

const setRowCheckedByHostId = (hostId: number, checked: boolean) => {
  const target = tableData.value.find(item => item.bk_host_id === hostId);
  if (target) {
    target.checked = checked;
  }
};

// 1. 处理单行勾选
const handleRowCheck = (checked: boolean, row: any) => {
  setRowCheckedByHostId(row.bk_host_id, checked);
  if (isCrossPageSelection.value) {
    if (!checked) {
      excludedIds.value.add(row.bk_host_id);
    } else {
      excludedIds.value.delete(row.bk_host_id);
    }
  }
};

// 2. 跨页全选
const handleSelectAllCrossPage = async () => {
  isCrossPageSelection.value = true;
  excludedIds.value.clear();

  // 更新当前页面的选中状态
  tableData.value.forEach(item => (item.checked = true));
};

// 3. 取消选择
const handleClearSelection = () => {
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  tableData.value.forEach(item => (item.checked = false));
};

// 4. 本页全选
const handleSelectCurrentPage = () => {
  isCrossPageSelection.value = false;
  tableData.value.forEach(item => (item.checked = true));
};

// 5. 表头 Checkbox 快速切换
const handleHeaderClick = () => {
  isCurrentPageAllChecked.value ? handleClearSelection() : handleSelectCurrentPage();
};

// 获取跨页全选的host_id数据
const crossPageSelectLoading = ref(false);
const crossPageHostIdData = ref<number[]>([]);
const getCorssPageHostIds = async () => {
  try {
    crossPageSelectLoading.value = true;
    const res = await TopoService.HostSelectHostID({
      exact_include_conditions: getParams().exact_include_conditions,
      fuzzy_include_conditions: getParams().fuzzy_include_conditions,
      exact_exclude_conditions: {
        bk_host_id: [...excludedIds.value],
      },
    });
    crossPageHostIdData.value = res.items;
  } catch (error) {
    console.error('获取跨页全选数据失败:', error);
  } finally {
    crossPageSelectLoading.value = false;
  }
};

// ---------- 辅助函数 ----------
function getUniqueChildrenFrom <K extends keyof TopoHostDistinctRespData>(
  prop: K,
  keyMap?: Map<number | string, string | number>,
) {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap?.get(value) || String(value),
    }));
}

const getParams = () => {
  const params = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: { bk_biz_id: mainStore.selectedBusinessId, node_role: ['agent', 'blank'] } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  
  searchSelectValue.value.forEach((item: any) => {
    // 处理合并后的 IP 字段
    if (item.id === 'ip') {
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      
      item.values?.forEach((value: any) => {
        const ip = value.id;
        if (IPV4_REG.test(ip)) {
          ipv4List.push(ip);
        } else if (IPV6_REG.test(ip)) {
          ipv6List.push(ip);
        }
      });
      
      if (ipv4List.length > 0) {
        params.fuzzy_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.fuzzy_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }
    
    // 处理管控区域ID:IP 组合字段
    if (item.id === 'area_ip') {
      const areaIds: number[] = [];
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      
      item.values?.forEach((value: any) => {
        const match = value.id.match(AREA_IP_REG);
        if (match) {
          const areaId = Number(match[1]);
          const ip = match[2];
          
          if (!areaIds.includes(areaId)) {
            areaIds.push(areaId);
          }
          
          if (IPV4_REG.test(ip)) {
            ipv4List.push(ip);
          } else if (IPV6_REG.test(ip)) {
            ipv6List.push(ip);
          }
        }
      });
      
      if (areaIds.length > 0) {
        params.exact_include_conditions.bk_networkarea_id = areaIds;
      }
      if (ipv4List.length > 0) {
        params.fuzzy_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.fuzzy_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }
    
    // 处理其他字段
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  
  return params;
};

// ---------- API 请求函数 ----------

/**
 * 获取管控区域列表
 */
const getNetworkAreaList = async (data: {bk_networkarea_id: number[]} | null) => {
  const res = await TopoService.NetworkAreaList({
    page: { limit: 0 },
    exact_include_conditions: { bk_networkarea_id: data?.bk_networkarea_id || [] },
  }).catch((err: any) => {
    console.error('获取管控区域列表失败:', err);
    return { total: 0, items: [] };
  });
  networkAreaListMap.value.set(-1, t('platform.nodeMan.agentStatus.unassigned'));
  res.items.forEach((item) => {
    networkAreaListMap.value.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
};

/**
 * 获取管控单元列表
 */
const getNetworkUnitList = async (data: {bk_networkunit_id: number[]} | null) => {
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: { bk_networkunit_id: data?.bk_networkunit_id || [] },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return { total: 0, items: [] };
  });
  networkUnitListMap.value.set(-1, t('platform.nodeMan.agentStatus.unassigned'));
  res.items.forEach((item) => {
    networkUnitListMap.value.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
};

/**
 * 获取主机筛选条件的唯一值
 */
const getHostDistinct = async () => {
  const params = {
    exact_include_conditions: {
      node_role: ['agent', 'blank'],
      bk_biz_id: mainStore.selectedBusinessId,
    },
  };
  const res = await TopoService.HostDistinct(params).catch((err: any) => {
    console.error('获取主机筛选条件唯一值失败:', err);
    return null;
  });
  await Promise.all([
    getNetworkAreaList(res),
    getNetworkUnitList(res),
  ]);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = res[key]
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text = value;
            if (key === 'bk_biz_id') text = bizListMap.value.get(Number(value)) || value;
            if (key === 'bk_networkarea_id') text = networkAreaListMap.value.get(Number(value)) || value;
            if (key === 'bk_networkunit_id') text = networkUnitListMap.value.get(Number(value)) || value;
            if (key === 'node_status') text = statusMap.value.get(value as string) || value;
            return { text, value };
          });
      }
    });
  }
};

// ---------- 侧边栏 ----------
const isShowSideslider = ref(false);
const currentAgent = ref();
// 打开侧边栏并加载进程列表
const openSidebar = async (agent) => {
  isShowSideslider.value = true;
  currentAgent.value = agent;
};


/**
 * 获取Agent列表
 */
const getAgentList = async () => {
  // Guard: skip if initial data not ready
  if (!isInitialDataLoaded.value) {
    return;
  }

  loading.value = true;
  try {
    const res = await TopoService.HostList(getParams()).catch((err: any) => {
      console.error('获取Agent列表失败:', err);
      return { total: 0, items: [] };
    });
    pagination.count = res.total;

    const pluginNumMap = await ProcessAPIService.GetProcessDistributionByHostID({
      exact_include_conditions: {
        bk_host_id: res.items.map((item: any) => item.bk_host_id),
      },
    }).catch((err: any) => {
      console.error('获取插件数量失败:', err);
      return {} as Record<number, number>;
    });
    tableData.value = res.items.map((item: any) => {
      let isChecked = false;
      if (isCrossPageSelection.value) {
        // 如果是跨页模式，只要不在排除名单里就是选中
        isChecked = !excludedIds.value.has(item.bk_host_id);
      }
      return {
        ...item.state,
        ...item.info,
        ...item,
        bk_host_innerip: item.info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6: item.info.bk_host_innerip_v6_list?.join(','),
        pluginNum: pluginNumMap[item.bk_host_id] || 0,
        checked: isChecked,
      };
    });
    agentList.value = tableData.value;
  } catch (err) {
    console.error('获取Agent列表失败:', err);
    tableData.value = [];
    pagination.count = 0;
  } finally {
    loading.value = false;
  }
};

// 【优化点2】使用debounce包装getAgentList，延迟300ms执行，防止重复请求
const debouncedGetAgentList = debounce(getAgentList, 300);

/**
 * 加载所有初始化数据（区域、单元、筛选条件）
 */
const loadInitialData = async () => {
  if (isInitialDataLoaded.value) {
    return;
  }

  try {
    // 【修复点】并行执行三个基础请求，但只执行一次
    await getHostDistinct();
    isInitialDataLoaded.value = true; // 标记基础数据已加载完成
  } catch (error) {
    console.error('加载初始化数据失败:', error);
    // 即使API失败，也标记为已加载完成，避免无限重试
    isInitialDataLoaded.value = true;
  }
  // 注：networkunit_use_for_agent 权限已合并到 PAGE_AUTHORIZED_ITEMS.agent，由 App.vue 统一下发
};

// ---------- 事件处理函数 ----------
/**
 * 解析多分隔符输入，支持空格、换行、分号、逗号
 */
const parseMultiDelimiterInput = (text: string): string[] => {
  // 使用正则匹配多种分隔符：空格、换行、分号、逗号、竖线、顿号
  return text
    .split(/[\s\n;,|、]+/)
    .map(item => item.trim())
    .filter(item => item.length > 0);
};

/**
 * 智能识别输入类型
 */
const detectInputType = (text: string): { type: 'ip' | 'area_ip' | 'agent_id' | null; value: string } => {
  // 1. 检测管控区域ID:IP 格式
  if (AREA_IP_REG.test(text)) {
    return { type: 'area_ip', value: text };
  }

  // 2. 检测 AgentID（01或02开头）
  if (AGENT_ID_REG.test(text)) {
    return { type: 'agent_id', value: text };
  }

  // 3. 检测 IPv4
  if (IPV4_REG.test(text)) {
    return { type: 'ip', value: text };
  }

  // 4. 检测 IPv6
  if (IPV6_REG.test(text)) {
    return { type: 'ip', value: text };
  }
  
  return { type: null, value: text };
};

/**
 * 拦截原生 paste 事件，将空格分隔符转换为组件能识别的逗号
 */
const handleNativePaste = (event: ClipboardEvent) => {
  const text = event.clipboardData?.getData('text');
  if (!text) return;

  // 如果包含空格但不包含组件默认分隔符，则替换空格为逗号
  if (text.includes(' ') && !/[|,、\r\n\n]/.test(text)) {
    event.preventDefault();
    const normalizedText = text.replace(/\s+/g, ',');

    // 手动触发粘贴
    const target = event.target as HTMLElement;
    if (target && target.isContentEditable) {
      document.execCommand('insertText', false, normalizedText);
    }
  }
};

/**
 * 处理粘贴/快速输入的逻辑
 */
const handleInputPaste = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  if (data.length === 0) return;

  const lastItem = data[data.length - 1];
  
  // 如果用户已经选择了类型（ip、area_ip、bk_agent_id），处理多分隔符输入
  if (['ip', 'area_ip', 'bk_agent_id'].includes(lastItem.id) && lastItem.values.length > 0) {
    // 遍历所有 values，解析每个可能包含多分隔符的值
    const allParsedItems: string[] = [];
    lastItem.values.forEach((value: any) => {
      const parsedItems = parseMultiDelimiterInput(value.id);
      allParsedItems.push(...parsedItems);
    });

    // 去重
    const uniqueItems = Array.from(new Set(allParsedItems));
    
    if (uniqueItems.length > 0) {
      // 替换为解析后的列表
      lastItem.values = uniqueItems.map(item => ({ id: item, name: item }));
      return;
    }
  }

  // 如果用户直接粘贴没选择类型，自动识别
  // SearchSelect 在多值粘贴后可能会生成多个“原始输入项”，这里从末尾聚合后统一识别。
  const searchFieldIds = new Set(searchSelectData.value.map(item => item.id));
  const tailRawInputIds: string[] = [];
  for (let i = data.length - 1; i >= 0; i--) {
    const current = data[i];
    if (searchFieldIds.has(current.id) || current.values?.length) break;
    tailRawInputIds.unshift(current.id);
  }

  const parsedItems = (tailRawInputIds.length > 0
    ? tailRawInputIds
    : [lastItem.id])
    .flatMap(text => parseMultiDelimiterInput(text));
  const uniqueParsedItems = Array.from(new Set(parsedItems));
  if (uniqueParsedItems.length === 0) return;
  
  // 智能识别第一个项的类型
  const firstDetection = detectInputType(uniqueParsedItems[0]);
  if (!firstDetection.type) return;

  // 验证所有项是否为同一类型
  const allSameType = uniqueParsedItems.every(item => {
    const detection = detectInputType(item);
    return detection.type === firstDetection.type;
  });

  if (!allSameType) {
    // 类型不一致，不自动识别
    return;
  }

  // 构造目标字段
  let targetId = '';
  let targetName = '';

  switch (firstDetection.type) {
    case 'ip':
      targetId = 'ip';
      targetName = 'IP';
      break;
    case 'area_ip':
      targetId = 'area_ip';
      targetName = t('platform.nodeMan.bk_cloud_name') + 'ID:IP';
      break;
    case 'agent_id':
      targetId = 'bk_agent_id';
      targetName = 'Agent ID';
      break;
  }

  if (targetId) {
    // 移除末尾原始输入项 + 已存在的同类型筛选，避免重复。
    searchSelectValue.value = searchSelectValue.value.filter((item: any) => {
      if (item.id === targetId) return false;
      return !tailRawInputIds.includes(item.id);
    });

    // 添加新的筛选条件
    searchSelectValue.value.push({
      id: targetId,
      name: targetName,
      values: uniqueParsedItems.map(item => ({ id: item, name: item })),
    });
  }
};
const handleSearchSelectChange = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id]) {
      filterOptionSource[item.id].checked = item.values.map((value: any) => value.id);
    }
  });
  // 【优化点2】使用防抖后的函数
  debouncedGetAgentList();
};

const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  if (index > -1) searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: field,
      values: checked.map((item: any) => {
        let name = item;
        if (field === 'bk_biz_id') name = bizListMap.value.get(Number(item)) || item;
        if (field === 'bk_networkarea_id') name = networkAreaListMap.value.get(Number(item)) || item;
        if (field === 'bk_networkunit_id') name = networkUnitListMap.value.get(Number(item)) || item;
        if (field === 'node_status') name = statusMap.value.get(item) || item;
        return { id: item, name };
      }),
    });
  }
};

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1; // 页码重置为1
  await getAgentList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getAgentList(); // 分页变化不防抖，立即执行
};

// const handleInstall = () => {
//   if (selection.value.length) {
//     dropdownShow.value = false;
//     triggerHandler('reinstall');
//   } else {
//     dropdownShow.value = !dropdownShow.value;
//   }
// };

const triggerHandler = (type: string, setupType = 'setup') => {
  switch (type) {
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

const getOperateShow = (row: Host, config: any) => {
  if (config.id === 'reinstall') {
    return false;
  }
  return config.show;
};

/** 获取操作项是否有权限：assign_unit 不做前端校验（直接放行），其余用 agent_operate */
const getItemHasAuth = (item: any): boolean => {
  if (item.id === 'assign_unit') return true;
  return hasOperateAuth.value;
};

/** 获取操作项对应的权限锁 mouseenter handler */
const getItemAuthMouseEnter = (_item: any) => (e: MouseEvent) => authLockMouseEnter(e, false);

/** 获取操作项对应的权限锁 mousemove handler */
const getItemAuthMouseMove = (_item: any) => (e: MouseEvent) => authLockMouseMove(e, false);

/** 获取操作项对应的权限锁 click handler */
const getItemAuthClick = (_item: any) => () => handleAuthClick();

const getRowOperateDisabled = (row: Host, config: any): { disabled: boolean; tooltip: string } => {
  if (config.id === 'assign_unit' && isNetworkUnitAssigned((row as any).bk_networkunit_id)) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.assignUnitDisabledRowAssigned'),
    };
  }
  // Check Agent status for upgrade, restart, and uninstall
  if ((config.id === 'upgrade' || config.id === 'restart' || config.id === 'uninstall') && (row as any).node_status !== 'running') {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledRowNotRunning'),
    };
  }
  // Check network unit assignment for upgrade, restart, and uninstall
  if ((config.id === 'upgrade' || config.id === 'restart' || config.id === 'uninstall') && !isNetworkUnitAssigned((row as any).bk_networkunit_id)) {
    return {
      disabled: true,
      tooltip: t('platform.nodeMan.agentStatus.operateDisabledRowUnassigned'),
    };
  }
  return { disabled: false, tooltip: '' };
};

const handleOperate = async (type: string, data: Host[], batch = false) => {
  // 如果是跨页全选模式，获取所有数据
  let operateData = data;
  if (isCrossPageSelection.value && type !== 'reinstall' && type !== 'assign_unit') {
    await getCorssPageHostIds();
    operateData = crossPageHostIdData.value.map((item: any) => ({ bk_host_id: item }));
    batch = true; // 强制设置为批量模式
  }

  switch (type) {
    case 'restart':
      handleOperatetHost(operateData, batch, 'restart');
      break;
    case 'uninstall':
      handleOperatetHost(operateData, batch, 'uninstall');
      break;
    case 'upgrade':
      handleOperatetHost(operateData, batch, 'upgrade');
      break;
    case 'assign_unit':
      nodeManageStore.updateAssignUnitParams({
        tableData: operateData.map((item: any) => ({ ...item })),
        isCrossPageSelection: isCrossPageSelection.value,
        queryParams: crossPageQueryParams.value,
      });
      router.push({ name: 'assignUnit' });
      return;
  }
  if (type !== 'reinstall') return;
  const params = {
    tableData: operateData.map((item: any) => ({ ...item })),
    type: 'reinstall',
    isCrossPageSelection: isCrossPageSelection.value,
    queryParams: crossPageQueryParams.value,
  };
  nodeManageStore.updateAgentEditRowData(params);
  router.push({ name: 'agentEdit' });
};
const handleSelectChange = ({ checked, row }: { checked: boolean; row: any }) => {
  setRowCheckedByHostId(row.bk_host_id, checked);
};

const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  tableData.value.forEach((item: any) => (item.checked = checked));
};

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
    result = await NodeAgentService.NodeAgentReconfig(params).catch(() => ({ workflow_id: '' }));
  } else {
    result = await NodeAgentService.NodeAgentRestart(params).catch(() => ({ workflow_id: '' }));
  }
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleUpgrade = async (osVersion: any[], upgradeData: {force: boolean, graceful_restart_timeout_sec: number}) => {
  loading.value = true;
  const params = {
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      target_version: osVersion[0].version,
      force: upgradeData.force,
      graceful_restart_timeout_sec: upgradeData.graceful_restart_timeout_sec,
    })),
  };
  const result = await NodeAgentService.NodeAgentUpgrade(params).catch(() => ({ workflow_id: '' }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleUninstall = async () => {
  loading.value = true;
  const result = await NodeAgentService.NodeAgentUninstall({
    host: operateData.value?.map((item: any) => ({ bk_host_id: item.bk_host_id })),
  }).catch(() => ({ workflow_id: '' }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleOperatetHost = async (data: Host[], batch: boolean, operateType: string) => {
  const titleObj = {
    firstIp: isCrossPageSelection.value ? selection.value[0].bk_host_innerip : data[0].bk_host_innerip,
    num: data.length,
  };
  let type = '';
  switch (operateType) {
    case 'restart': type = t('platform.nodeMan.agentStatus.restart'); break;
    case 'upgrade': type = t('platform.nodeMan.agentStatus.upgrade'); break;
    case 'uninstall': type = t('platform.nodeMan.agentStatus.uninstall'); break;
  }
  operateData.value = data;
  if (operateType === 'upgrade') {
    upgradePreviewData.hosts = data;
    upgradePreviewData.isShow = true;
  } else if (operateType === 'restart') {
    operateDialogIsShow.value = true;
    operateDialogData.type = operateType;
    operateDialogData.title = batch
      ? t('platform.nodeMan.agentStatus.confirmIsBatch', { type })
      : t('platform.nodeMan.agentStatus.confirmIsSingle', { type });
    operateDialogData.subTitle = batch
      ? t('platform.nodeMan.agentStatus.batchOperate', { type, firstIp: titleObj.firstIp, num: titleObj.num })
      : t('platform.nodeMan.agentStatus.singleOperate', { type, firstIp: titleObj.firstIp });
  } else if (operateType === 'uninstall') {
    InfoBox({
      title: batch
        ? t('platform.nodeMan.agentStatus.confirmIsBatch', { type })
        : t('platform.nodeMan.agentStatus.confirmIsSingle', { type }),
      subTitle: batch
        ? t('platform.nodeMan.agentStatus.batchOperate', { type, firstIp: titleObj.firstIp, num: titleObj.num })
        : t('platform.nodeMan.agentStatus.singleOperate', { type, firstIp: titleObj.firstIp }),
      onConfirm: () => {
        handleUninstall();
      },
    });
  }
};


// ---------- 监听与生命周期 ----------

watch(() => route.query, async (newQuery, oldQuery) => {
  if (JSON.stringify(newQuery) === JSON.stringify(oldQuery)) return;

  const { os_type, cpu_arch, node_version, bk_networkarea_id, bk_networkunit_id, bk_networkunit_name } = newQuery;

  if (os_type && cpu_arch && node_version) {
    searchSelectValue.value = [
      ...searchSelectValue.value.filter(item => !['os_type', 'cpu_arch', 'node_version'].includes(item.id)),
      { id: 'os_type', name: t('platform.nodeMan.os_type'), values: [{ id: os_type, name: os_type }] },
      { id: 'cpu_arch', name: t('platform.nodeMan.cpu_arch'), values: [{ id: cpu_arch, name: cpu_arch }] },
      { id: 'node_version', name: t('platform.nodeMan.agent_version'), values: [{ id: node_version, name: node_version }] },
    ];
  } else if (bk_networkarea_id !== undefined && bk_networkunit_id !== undefined) {
    const areaId = Number(bk_networkarea_id);
    const unitId = Number(bk_networkunit_id);
    
    // 等待基础数据加载完成
    const setSearchValue = () => {
      searchSelectValue.value = [
        ...searchSelectValue.value.filter(item => !['bk_networkarea_id', 'bk_networkunit_id'].includes(item.id)),
        {
          id: 'bk_networkarea_id',
          name: t('platform.nodeMan.bk_cloud_name'),
          values: [{ id: areaId, name: networkAreaListMap.value.get(areaId) || areaId }],
        },
        {
          id: 'bk_networkunit_id',
          name: t('platform.nodeMan.bk_cloud_unit'),
          values: [{ id: unitId, name: bk_networkunit_name || networkUnitListMap.value.get(unitId) }],
        },
      ];
    };
    
    // 如果数据已加载，直接设置；否则等待
    if (isInitialDataLoaded.value) {
      setSearchValue();
    } else {
      const unwatch = watch(isInitialDataLoaded, (loaded) => {
        if (loaded) {
          setSearchValue();
          unwatch();
        }
      });
    }
  } else if (searchSelectValue.value.length) {
    searchSelectValue.value = [];
  }
}, { immediate: true });

// 等业务初始化完成后再触发请求，避免 selectedBusinessId 从 [] 变为实际值时重复请求
const isInitialLoading = ref(false);
watch(
  [() => mainStore.isBusinessReady, () => mainStore.selectedBusinessId],
  async ([ready]) => {
    if (!ready) return;

    isInitialLoading.value = true;
    isInitialDataLoaded.value = false;
    tableData.value = [];
    pagination.count = 0;
    pagination.current = 1;

    await loadInitialData();
    if (isInitialDataLoaded.value) {
      // 等待 pending watchers flush（如 route.query 注册的 watch(isInitialDataLoaded) 设置 searchSelectValue）
      await nextTick();
      await getAgentList();
    }
    isInitialLoading.value = false;
  },
  { immediate: true },
);

watch(
  searchSelectValue,
  () => {
    if (isInitialDataLoaded.value && !isInitialLoading.value) {
      pagination.current = 1;
      debouncedGetAgentList();
    }
  },
  { deep: true },
);

onUnmounted(() => {
  debouncedGetAgentList.cancel();
});

</script>
<style lang="postcss" scoped>
.operate-item-disabled {
  color: #c4c6cc !important;
  cursor: not-allowed !important;
  pointer-events: auto !important;

  &:hover {
    color: #c4c6cc !important;
    background-color: transparent !important;
  }
}
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

/* 无权限按钮：模拟 disabled 视觉效果但保持鼠标事件可响应 */
.auth-lock-wrapper {
  cursor: pointer;

  .auth-disabled-btn {
    opacity: 0.5;
    cursor: pointer !important;
    pointer-events: auto !important;
  }

  .auth-disabled-text-btn {
    color: #c4c6cc !important;
    cursor: pointer !important;
    pointer-events: auto !important;

    &:hover {
      color: #c4c6cc !important;
      background: transparent !important;
    }
  }
}
</style>
