<template>
  <page-header :title="'route.proxyStatus'" :back="Object.keys(route.query).length > 0" v-if="isProxyStatus">
    <!-- Proxy 状态统计：标题后、subtitle 前 -->
    <template #after-title>
      <bk-loading :loading="proxyStatusLoading" class="flex items-center">
        <div class="w-[1px] h-[14px] bg-[#DCDEE5] mx-[16px]"></div>
        <span
          class="inline-flex items-center cursor-pointer mr-[20px]"
          v-bk-tooltips="{ content: `${$t('platform.nodeMan.proxyStatus.onlineProxyCount')}: ${proxyStatus.online}` }">
          <span class="w-[6px] h-[6px] rounded-full bg-[#3FC06D] mr-[4px]" />
          <span class="text-[12px] mr-[2px]">{{ $t('platform.nodeMan.proxyStatus.online') }}</span>
          <span class="text-[#3FC06D] text-[12px] font-medium">{{ proxyStatus.online }}</span>
        </span>
        <span
          class="inline-flex items-center cursor-pointer mr-[20px]"
          v-bk-tooltips="{ content: `${$t('platform.nodeMan.proxyStatus.abnormalProxyCount')}: ${proxyStatus.abnormal}` }">
          <span class="w-[6px] h-[6px] rounded-full bg-[#EA3636] mr-[4px]" />
          <span class="text-[12px] mr-[2px]">{{ $t('platform.nodeMan.proxyStatus.abnormal') }}</span>
          <span class="text-[#EA3636] text-[12px] font-medium">{{ proxyStatus.abnormal }}</span>
        </span>
        <span
          class="inline-flex items-center cursor-pointer"
          v-bk-tooltips="{ content: `${$t('platform.nodeMan.proxyStatus.notInstalledProxyCount')}: ${proxyStatus.notInstalled}` }">
          <span class="w-[6px] h-[6px] rounded-full bg-[#979BA5] mr-[4px]" />
          <span class="text-[12px] mr-[2px]">{{ $t('platform.nodeMan.proxyStatus.notInstalled') }}</span>
          <span class="text-[#979BA5] text-[12px] font-medium">{{ proxyStatus.notInstalled }}</span>
        </span>
      </bk-loading>
    </template>
  </page-header>
  <div :class="{ 'mx-[24px]': isProxyStatus }">
    <FlexRow class="mt-[24px]">
      <template #left>
        <div class="flex items-center">
          <!-- 安装/重装 Proxy：需要 networkunit_use_for_proxy 权限 -->
          <template v-if="hasInstallProxyAuth">
            <Button
              theme="primary"
              :class="['mr-[8px]', 'w-[130px]', { 'btn-reinstall': hasSelection }]"
              @click="handlePrimaryButtonClick">
              <span>{{ primaryButtonLabel }}</span>
            </Button>
          </template>
          <span
            v-else
            class="inline-flex items-center auth-lock-wrapper mr-[8px]"
            @click="handleInstallAuthClick"
            @mouseenter="installMouseEnter($event, false)"
            @mousemove="installMouseMove($event, false)"
            @mouseleave="installMouseLeave()"
          >
            <Button theme="primary" class="auth-disabled-btn w-[130px]">
              <span>{{ primaryButtonLabel }}</span>
            </Button>
          </span>
          <!-- 批量操作：无权限时置灰 + hover 带锁 + 点击申请权限 -->
          <template v-if="hasProxyOperateAuth">
            <MoreAction
              :data="selectTableData"
              placement="bottom-start"
              :batch="true"
              :has-auth="hasProxyOperateAuth"
              :unit-proxy-count="unitProxyCount"
              :unit-agent-count="unitAgentCount"
              @reinstall="handleReinstall"
              @assign-unit="handleAssignUnit"
              @ops-setting="handleOpsSetting"
              :is-cross-page-selection="isCrossPageSelection"
              :cross-page-query-params="crossPageQueryParams"
            >
              <Button :disabled="!hasSelection" class="mr-[8px]">
                <span>{{ $t("topoManager.workAreaDetail.button.batch") }}</span>
                <i
                  class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
                ></i>
              </Button>
            </MoreAction>
          </template>
          <span
            v-else
            class="inline-flex items-center auth-lock-wrapper mr-[8px]"
            @click="handleAuthClick"
            @mouseenter="authLockMouseEnter($event, false)"
            @mousemove="authLockMouseMove($event, false)"
            @mouseleave="authLockMouseLeave()"
          >
            <Button class="auth-disabled-btn">
              <span>{{ $t("topoManager.workAreaDetail.button.batch") }}</span>
              <i
                class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
              ></i>
            </Button>
          </span>
          <!-- 复制IP：无权限时置灰 + hover 带锁 + 点击申请权限 -->
          <template v-if="hasProxyOperateAuth">
            <copy-ip-dropdown
              :disabled="!hasSelection"
              :data="tableData"
              :list="list"
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
      </template>
      <template #right>
        <SearchSelect
          :max-height="240"
          class="w-[480px] bg-[#fff]"
          :placeholder="$t('platform.nodeMan.proxySearchPlaceholder')"
          :unique-select="true"
          v-model.trim="searchKey"
          :data="searchSelectData"
          @update:model-value="handleSearchSelectChange"
          @paste="handleNativePaste"
        >
        </SearchSelect>
      </template>
    </FlexRow>
    <!-- table -->
    <DetailTable
      ref="detailTableRef"
      v-model:search-select-value="searchKey"
      :bk-networkunit-id="active"
      :is-batch-reinstall="batchReinstall"
      :has-proxy-operate-auth="hasProxyOperateAuth"
      :unit-proxy-count="unitProxyCount"
      :unit-agent-count="unitAgentCount"
      @select-change="handleSelectChange"
      @get-data="handleGetData"
      @update-cross-page="handleUpdateCrossPage"
      @excluded-ids-change="handleExcludedIdsChange"
      @update-search-select-data="handleUpdateSearchSelectData"
      @assign-unit="handleAssignUnit"
      @ops-setting="handleOpsSetting"
      @auth-click="handleAuthClick">
    </DetailTable>
  </div>
  <InstallProxy
    v-model:is-show="isShowInstallProxy"
    :bk_networkunit_id="active"
  />
  <ReinstallProxy
    v-model:is-show="isShowReinstallProxy"
    :data="reinstallData"
    :is-cross-page-selection="isCrossPageSelection"
    :params="crossPageQueryParams"
  />
  <AssignUnit
    v-model:is-show="isShowAssignUnit"
    :data="assignUnitData"
    :is-cross-page-selection="isCrossPageSelection"
    :params="crossPageQueryParams"
  />
</template>
<script setup lang="ts">
import { Button, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { debounce } from 'lodash';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import DetailTable from './detail-table.vue';
import MoreAction from './more-action.vue';

import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import useAuthLock from '@/composables/use-auth-lock';
import InstallProxy from '@/pages/topo/install-proxy/install-proxy.vue';
import ReinstallProxy from '@/pages/topo/install-proxy/reinstall-proxy.vue';
import AssignUnit from '@/pages/topo/install-proxy/assign-unit.vue';
import { TopoService } from '@/api/modules/topo';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';
import { useAuthStore } from '@/stores/auth';

const props = defineProps({
  active: {
    type: Number,
    default: null,
  },
});
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const authStore = useAuthStore();
const nodeManageStore = useNodeManageStore();

const isProxyStatus = computed(() => route.name === 'proxy');

// 管控单元下 proxy 和 agent 总数（用于限制最后一个 proxy 的卸载）
const unitProxyCount = ref(0);
const unitAgentCount = ref(0);
const fetchUnitStats = async () => {
  if (!props.active) return;
  const res = await TopoService.TopoGraphNodeGetReq({
    bk_networkunit_id: [props.active],
  }).catch(() => ({ graph_node_info: [] as any[] }));
  const info = (res.graph_node_info || []).find(
    (item: any) => item.bk_networkunit_id === props.active,
  );
  if (info) {
    unitProxyCount.value = info.total_proxy ?? 0;
    unitAgentCount.value = info.total_agent ?? 0;
  }
};

// 切换到 proxy tab 且有单元 ID 时，主动加载 proxy_view / proxy_operate 权限
// 仅「管控区域详情」场景（非节点管理 proxy 列表页）才额外加载 networkunit_use_for_proxy
watch(() => props.active, (id) => {
  if (!id) return;
  const items = [
    { action: 'proxy_view', resource_type: 'networkunit' },
    { action: 'proxy_operate', resource_type: 'networkunit' },
  ];
  if (!isProxyStatus.value) {
    items.push({ action: 'networkunit_use_for_proxy', resource_type: 'networkunit' });
  }
  authStore.fetchAuthorized(items);
  fetchUnitStats();
}, { immediate: true });
// ===== proxy_operate 权限控制（批量操作、复制IP） =====
const {
  hasAuth: hasProxyOperateAuth,
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
  handleAuthClick,
} = useAuthLock('proxy_operate', () => mainStore.selectedBusinessId);
const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;
const AGENT_ID_REG = /^0[12]/; // AgentID 以 01 或 02 开头
const AREA_IP_REG = /^(\d+):(.+)$/; // 管控区域ID:IP 格式

// ===== networkunit_use_for_proxy 权限控制（仅「管控区域详情」场景的安装/重装 Proxy 按钮） =====
// 节点管理（route.name === 'proxy'）的安装 Proxy 按钮不做权限判断，视为始终有权限
const {
  hasAuth: hasInstallProxyAuthRaw,
  handleMouseEnter: installMouseEnter,
  handleMouseMove: installMouseMove,
  handleMouseLeave: installMouseLeave,
  handleAuthClick: handleInstallAuthClick,
} = useAuthLock('networkunit_use_for_proxy', () => props.active, { resourceType: 'networkunit' });
const hasInstallProxyAuth = computed(() => isProxyStatus.value || hasInstallProxyAuthRaw.value);
// 搜索
const searchKey = ref<ISearchValue[]>([]);
const searchSelectData = ref<ISearchItem[]>([
  { id: 'ip', name: 'IP', multiple: true }, // 合并后的 IP 筛选
  { id: 'area_ip', name: `${t('topoManager.workAreaDetail.table.networkArea')}ID:IP`, multiple: true }, // 管控区域ID:IP
  {
    name: t('platform.nodeMan.bk_cloud_name'),
    id: 'bk_networkarea_id',
  },
  {
    name: 'AgentID',
    id: 'bk_agent_id',
  },
  {
    name: 'Host ID',
    id: 'bk_host_id',
  },
  {
    name: t('installProxy.proxyVersion'),
    id: 'node_version',
  },
  {
    name: t('topoManager.workAreaDetail.table.proxyStatus'),
    id: 'node_status',
  },
  {
    id: 'dept_name',
    name: t('platform.nodeMan.dept_name'),
  },
]);
// 复制
const subList = [
  {
    id: 'ipv4',
    name: 'IPv4',
  },
  {
    id: 'ipv6',
    name: 'IPv6',
  },
  {
    id: 'workarea+ipv4',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv4`,
  },
  {
    id: 'workarea+ipv6',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv6`,
  },
];
const list = [
  {
    id: 'select',
    name: t('topoManager.workArea.copy.select'),
    children: subList,
  },
  {
    id: 'all',
    name: t('topoManager.workArea.copy.selectAll'),
    children: subList,
  },
];
const batchReinstall = ref(false);
const isShowReinstallProxy = ref(false);
const reinstallData = ref<Host[]>([]);
const selectTableData = ref<Host[]>([]);

// 获取跨页全选的host_id数据
const excludedIds = ref<number[]>([]);

const getParams = () => {
  const fuzzyKeys = new Set<string>();
  const params = {
    exact_include_conditions: {
      node_role: ['proxy'],
    } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  if (route.name === 'proxy') {
    params.exact_include_conditions.bk_biz_id = mainStore.selectedBusinessId;
  }
  searchKey.value.forEach((item: any) => {
    // IP 字段：自动识别 IPv4/IPv6 并分类
    if (item.id === 'ip' && item.values?.length) {
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        if (IPV4_REG.test(value.id)) {
          ipv4List.push(value.id);
        } else if (IPV6_REG.test(value.id)) {
          ipv6List.push(value.id);
        }
      });
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

    // 管控区域ID:IP：自动拆分为 bk_networkarea_id + IP 列表
    if (item.id === 'area_ip' && item.values?.length) {
      const areaIds = new Set<number>();
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        const match = AREA_IP_REG.exec(value.id);
        if (match) {
          const areaId = Number(match[1]);
          const ip = match[2];
          areaIds.add(areaId);
          if (IPV4_REG.test(ip)) {
            ipv4List.push(ip);
          } else if (IPV6_REG.test(ip)) {
            ipv6List.push(ip);
          }
        }
      });
      if (areaIds.size > 0) {
        params.exact_include_conditions.bk_networkarea_id = Array.from(areaIds);
      }
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

    if (item.id === 'bk_host_id') {
      const hostIds: number[] = (item.values ?? [])
        .map((value: any) => String(value.id).trim())
        .filter((value: string) => /^\d+$/.test(value))
        .map((value: string) => Number(value))
        .filter((value: number) => Number.isSafeInteger(value) && value > 0);

      const uniqueHostIds: number[] = Array.from(new Set<number>(hostIds));
      if (uniqueHostIds.length > 0) {
        params.exact_include_conditions.bk_host_id = uniqueHostIds;
      }
      return;
    }

    // 管控区域ID：字符串转数字（从管控区域列表跳转或搜索框筛选）
    if (item.id === 'bk_networkarea_id' && item.values?.length) {
      const areaIds = item.values
        .map((value: any) => Number(value.id))
        .filter((value: number) => Number.isSafeInteger(value));
      if (areaIds.length > 0) {
        params.exact_include_conditions.bk_networkarea_id = Array.from(new Set(areaIds));
      }
      return;
    }

    // 其他字段：保持原有 fuzzy/exact 分类逻辑
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const crossPageQueryParams = computed(() => ({
  exact_include_conditions: getParams().exact_include_conditions,
  fuzzy_include_conditions: getParams().fuzzy_include_conditions,
  exact_exclude_conditions: {
    bk_host_id: [...excludedIds.value],
  },
}));

const handleReinstall = () => {
  isShowReinstallProxy.value = true;
  reinstallData.value = selectTableData.value;
};
const isShowInstallProxy = ref(false);
const handleInstallProxy = () => {
  isShowInstallProxy.value = true;
};
const isShowAssignUnit = ref(false);
const assignUnitData = ref<Host[]>([]);
const handleAssignUnit = (row?: Host) => {
  assignUnitData.value = row ? [row] : selectTableData.value;
  isShowAssignUnit.value = true;
};

const handleOpsSetting = (row?: Host) => {
  const data = row ? [row] : selectTableData.value;
  nodeManageStore.updateOpsFieldsParams({
    tableData: data.map((item: any) => ({ ...item })),
    isCrossPageSelection: !row && isCrossPageSelection.value,
    nodeType: 'proxy',
    queryParams: crossPageQueryParams.value,
  });
  router.push({ name: 'opsSetting' });
};

// 计算属性：是否有勾选主机
const hasSelection = computed(() => {
  const hasCheckedRow = tableData.value.some(item => item.checked);
  return selectTableData.value.length > 0 || hasCheckedRow || isCrossPageSelection.value;
});

// 计算属性：主按钮文案（安装/重装动态切换）
const primaryButtonLabel = computed(() => (hasSelection.value ? t('topoManager.workAreaDetail.button.reinstallProxy') : t('topoManager.workAreaDetail.button.installProxy')));

// 主按钮点击事件：根据勾选状态分支调用安装或重装
const handlePrimaryButtonClick = () => {
  if (hasSelection.value) {
    handleReinstall();
  } else {
    handleInstallProxy();
  }
};
const tableData = ref<Host[]>([]);
const handleSelectChange = (tableList: Host[]) => {
  // detail-table 已经回传选中列表，这里直接接管即可，避免状态二次过滤导致丢失
  selectTableData.value = tableList;
};
const handleGetData = (tableList: Host[]) => {
  selectTableData.value = [];
  tableData.value = tableList;
};

const isCrossPageSelection = ref(false);
const handleUpdateCrossPage = (crossPage: boolean) => {
  isCrossPageSelection.value = crossPage;
};
// 更新detail-table.vue中的excludedIds
const handleExcludedIdsChange = (ids: number[]) => {
  excludedIds.value = ids;
};

const handleUpdateSearchSelectData = (data: any) => {
  searchSelectData.value = data;
};

/**
 * 解析多分隔符输入，支持空格、换行、分号、逗号
 */
const parseMultiDelimiterInput = (text: string): string[] => text
  // 使用正则匹配多种分隔符：空格、换行、分号、逗号、竖线、顿号
  .split(/[\s\n;,|、]+/)
  .map(item => item.trim())
  .filter(item => item.length > 0);

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
  const allSameType = uniqueParsedItems.every(item => detectInputType(item).type === firstDetection.type);

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
      targetName = `${t('topoManager.workAreaDetail.table.networkArea')}ID:IP`;
      break;
    case 'agent_id':
      targetId = 'bk_agent_id';
      targetName = 'Agent ID';
      break;
  }

  if (targetId) {
    // 移除末尾原始输入项 + 已存在的同类型筛选，避免重复。
    searchKey.value = searchKey.value.filter((item: any) => {
      if (item.id === targetId) return false;
      return !tailRawInputIds.includes(item.id);
    });

    // 添加新的筛选条件
    searchKey.value.push({
      id: targetId,
      name: targetName,
      values: uniqueParsedItems.map(item => ({ id: item, name: item })),
    });
  }
};

const handleSearchSelectChange = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
};

// DetailTable ref for refreshing list
const detailTableRef = ref<InstanceType<typeof DetailTable>>();

// ---------- Proxy 状态统计 ----------
const proxyStatusLoading = ref(true);
const proxyStatus = ref({ online: 0, abnormal: 0, notInstalled: 0 });
const fetchProxyStatusCount = async () => {
  const bizIds = mainStore.selectedBusinessId || [];
  const listParams = getParams();

  const { node_status: userNodeStatus, bk_biz_id: _, node_role: __, ...otherConditions } = listParams.exact_include_conditions;

  proxyStatusLoading.value = true;
  try {
    if (userNodeStatus?.length) {
      // 用户有 node_status 筛选，只统计已安装 proxy（node_role=proxy），不查 blank
      const hasRunning = userNodeStatus.includes('running');
      const abnormalStatuses = userNodeStatus.filter((s: string) => s !== 'running');

      const [onlineRes, abnormalRes] = await Promise.all([
        hasRunning
          ? TopoService.HostList({
            page: { offset: 0, limit: 0 },
            only_count: true,
            exact_include_conditions: { ...otherConditions, bk_biz_id: bizIds, node_role: ['proxy'], node_status: ['running'] },
            fuzzy_include_conditions: { ...listParams.fuzzy_include_conditions },
          }).catch(() => ({ total: 0 }))
          : { total: 0 },
        abnormalStatuses.length > 0
          ? TopoService.HostList({
            page: { offset: 0, limit: 0 },
            only_count: true,
            exact_include_conditions: { ...otherConditions, bk_biz_id: bizIds, node_role: ['proxy'], node_status: abnormalStatuses },
            fuzzy_include_conditions: { ...listParams.fuzzy_include_conditions },
          }).catch(() => ({ total: 0 }))
          : { total: 0 },
      ]);

      proxyStatus.value = {
        online: (onlineRes as any).total ?? 0,
        abnormal: (abnormalRes as any).total ?? 0,
        notInstalled: 0,
      };
    } else {
      // 无状态筛选：只统计已安装 proxy，abnormal = total - online（含初始化等所有非 running）
      const [onlineRes, totalRes] = await Promise.all([
        TopoService.HostList({
          page: { offset: 0, limit: 0 },
          only_count: true,
          exact_include_conditions: { ...otherConditions, bk_biz_id: bizIds, node_role: ['proxy'], node_status: ['running'] },
          fuzzy_include_conditions: { ...listParams.fuzzy_include_conditions },
        }).catch(() => ({ total: 0 })),
        TopoService.HostList({
          page: { offset: 0, limit: 0 },
          only_count: true,
          exact_include_conditions: { ...otherConditions, bk_biz_id: bizIds, node_role: ['proxy'] },
          fuzzy_include_conditions: { ...listParams.fuzzy_include_conditions },
        }).catch(() => ({ total: 0 })),
      ]);
      const online = onlineRes.total ?? 0;
      const totalProxy = totalRes.total ?? 0;
      proxyStatus.value = {
        online,
        abnormal: Math.max(totalProxy - online, 0),
        notInstalled: 0,
      };
    }
  } catch {
    proxyStatus.value = { online: 0, abnormal: 0, notInstalled: 0 };
  } finally {
    proxyStatusLoading.value = false;
  }
};
const debouncedFetchProxyStatusCount = debounce(fetchProxyStatusCount, 300);
watch(() => [mainStore.selectedBusinessId, isProxyStatus.value, searchKey.value], () => { if (isProxyStatus.value) debouncedFetchProxyStatusCount(); }, { immediate: true, deep: true });

// Listen for assign unit success event to refresh list
const handleAssignUnitSuccess = () => {
  detailTableRef.value?.handleUpdate?.();
  if (isProxyStatus.value) debouncedFetchProxyStatusCount();
};

onMounted(() => {
  // 从管控区域列表跳转时，设置搜索框显示值（筛选参数由 getParams 从 route.query 直接读取）
  if (route.query.bk_networkarea_id) {
    const areaId = Number(route.query.bk_networkarea_id);
    searchKey.value = [{
      id: 'bk_networkarea_id',
      name: t('platform.nodeMan.bk_cloud_name'),
      values: [{ id: String(areaId), name: (route.query.bk_networkarea_name as string) || String(areaId) }],
    }];
  }
  window.addEventListener('proxy-assign-unit-success', handleAssignUnitSuccess);
});

onUnmounted(() => {
  window.removeEventListener('proxy-assign-unit-success', handleAssignUnitSuccess);
});
</script>

<style lang="postcss" scoped>
/* 无权限按钮：模拟 disabled 视觉效果但保持鼠标事件可响应 */
.auth-lock-wrapper {
  cursor: pointer;

  .auth-disabled-btn {
    opacity: 0.5;
    cursor: pointer !important;
    pointer-events: auto !important;
  }
}
</style>
