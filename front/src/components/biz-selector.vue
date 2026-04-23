<template>
  <div class="nm-menu-biz mb-[10px]" v-if="isVisible">
    <div
      v-show="!expanded"
      class="w-[30px] h-[30px] text-[12px] bg-[#F0F1F5] m-auto cursor-pointer flex items-center justify-center"
    >
      {{ shrinkText }}
    </div>
    <!-- 策略管理：单选业务 -->
    <Select
      v-if="isSingle"
      v-show="expanded"
      class="mx-[12px]"
      v-model="singleBusiness"
      :filter-option="filterOption"
      :show-selected-icon="false"
      :clearable="false"
      filterable
      :placeholder="hasNoAuthorizedBiz() ? undefined : t('platform.nodeMan.allBusiness')"
      :popover-options="{ boundary: 'document.body', width: '235px' }"
      @toggle="handleToggle"
    >
      <Select.Option
        v-for="item in businessList"
        :key="item.bk_biz_id"
        :name="item.bk_biz_name"
        :id="item.bk_biz_id"
        v-bk-tooltips="{
          content: isBizAuthorized(item.bk_biz_id)
            ? `[${item.bk_biz_id}] ${item.bk_biz_name}`
            : t('components.permission.noPermission'),
          disabled: isBizAuthorized(item.bk_biz_id) && !textOverflowMap[item.bk_biz_id],
          boundary: 'parent',
          placement: 'right',
          offset: 10
        }"
      >
        <div
          class="w-full flex items-center biz-select-option overflow-hidden"
          :class="{ 'unauthorized-biz-row': !isBizAuthorized(item.bk_biz_id) }"
          @click="handleOptionClick($event, item.bk_biz_id)"
          @mouseenter="handleOptionMouseEnter($event, item.bk_biz_id)"
          @mousemove="handleOptionMouseMove($event, item.bk_biz_id)"
          @mouseleave="handleOptionMouseLeave()"
        >
          <Button
            class="mr-[8px] w-[18px] shrink-0"
            text
            @click.native.stop="handleCollect(item.bk_biz_id)">
            <i
              class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
              v-if="collectList.includes(item.bk_biz_id)">
            </i>
            <i
              class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
              v-else>
            </i>
          </Button>
          <div
            class="truncate"
            @mouseenter="handleTextMouseenter($event, item.bk_biz_id)">
            [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
          </div>
        </div>
      </Select.Option>
    </Select>
    <!-- 其他模块：多选业务 -->
    <Select
      v-else
      v-show="expanded"
      class="mx-[12px]"
      v-model="multiBusiness"
      :filter-option="filterOption"
      :show-selected-icon="false"
      multiple
      filterable
      :placeholder="hasNoAuthorizedBiz() ? undefined : t('platform.nodeMan.allBusiness')"
      :popover-options="{ boundary: 'document.body', width: '235px' }"
      @change="handleMultiChange"
      @toggle="handleToggle"
    >
      <Select.Option
        v-for="item in businessList"
        :key="item.bk_biz_id"
        :name="item.bk_biz_name"
        :id="item.bk_biz_id"
        v-bk-tooltips="{
          content: isBizAuthorized(item.bk_biz_id)
            ? `[${item.bk_biz_id}] ${item.bk_biz_name}`
            : t('components.permission.noPermission'),
          disabled: isBizAuthorized(item.bk_biz_id) && !textOverflowMap[item.bk_biz_id],
          boundary: 'parent',
          placement: 'right',
          offset: 10
        }"
      >
        <div
          class="w-full flex items-center biz-select-option overflow-hidden"
          :class="{ 'unauthorized-biz-row': !isBizAuthorized(item.bk_biz_id) }"
          @click="handleOptionClick($event, item.bk_biz_id)"
          @mouseenter="handleOptionMouseEnter($event, item.bk_biz_id)"
          @mousemove="handleOptionMouseMove($event, item.bk_biz_id)"
          @mouseleave="handleOptionMouseLeave()"
        >
          <Button
            class="mr-[8px] w-[18px] shrink-0"
            text
            @click.native.stop="handleCollect(item.bk_biz_id)">
            <i
              class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
              v-if="collectList.includes(item.bk_biz_id)">
            </i>
            <i
              class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
              v-else>
            </i>
          </Button>
          <div
            class="truncate"
            @mouseenter="handleTextMouseenter($event, item.bk_biz_id)">
            [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
          </div>
        </div>
      </Select.Option>
    </Select>
    <!-- 锁图标由 use-auth-lock hook 通过 DOM 管理，无需 Teleport -->
  </div>
</template>

<script setup lang="ts">
import { Button, Select } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';
import { usePermissionStore } from '@/stores/permission';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const authStore = useAuthStore();

defineProps<{
  /** 左侧菜单是否展开 */
  expanded: boolean;
}>();

// ===== 可见性与模式 =====
const NO_BIZ_SELECTOR_MAIN_MENUS = ['topoManager', 'pkgManager'];

// 业务访问权限 action
const BIZ_ACCESS_ACTION = 'biz_access';

const isVisible = computed(() => {
  if (route.path.includes('topo-manager') || route.path.includes('pkg-manager')) return false;
  // 403/404 页面：通过 mainMenu 判断原模块是否不需要业务选择器
  if (route.name === '403' || route.name === '404') {
    const mainMenu = route.query.mainMenu as string;
    if (mainMenu && NO_BIZ_SELECTOR_MAIN_MENUS.includes(mainMenu)) return false;
  }
  return true;
});
const isSingle = computed(() => route.path.includes('rule-manager'));

// ===== 菜单路由 → authorized action 映射 =====
// 该映射仅用于「点击无权限业务申请权限」时，定位需要申请哪个菜单 view action。
// 业务可访问范围判断统一使用 biz_access（与菜单 view 权限解耦，由后端约定）。
const MENU_ROUTE_ACTION_MAP: Record<string, string> = {
  // nodeManager
  agent: 'agent_view',
  proxy: 'proxy_view',
  agentSetup: 'agent_operate',
  agentEdit: 'agent_operate',
  assignUnit: 'agent_operate',
  plugin: 'plugin_view',
  history: 'agent_history_view',
  taskDetail: 'agent_history_view',
  log: 'agent_history_view',
  // ruleManager
  agentStrategy: 'config_policy_view',
  proxyStrategy: 'config_policy_view',
  pluginStrategy: 'config_policy_view',
  strategyTaskHistory: 'config_policy_history_view',
};

// 顶层导航模块 → 默认 view action（用于 403 页面时通过 mainMenu 反推）
const MAIN_MENU_DEFAULT_ACTION: Record<string, string> = {
  nodeManager: 'agent_view',
  ruleManager: 'config_policy_view',
};

// 缓存上次有效的 action，用于 403 等无法匹配路由名的场景
let cachedAction: string | undefined;

// 根据当前路由名获取对应的菜单 view action（仅用于申请权限）
function getActionForRoute(): string | undefined {
  const name = route.name;
  if (typeof name === 'string' && MENU_ROUTE_ACTION_MAP[name]) {
    cachedAction = MENU_ROUTE_ACTION_MAP[name];
    return cachedAction;
  }
  // 403/404 页面：通过 query.mainMenu 反推
  if (name === '403' || name === '404') {
    const mainMenu = route.query.mainMenu as string;
    if (mainMenu && MAIN_MENU_DEFAULT_ACTION[mainMenu]) {
      return MAIN_MENU_DEFAULT_ACTION[mainMenu];
    }
  }
  // 兜底：使用缓存
  return cachedAction;
}

// 判断某个业务是否有访问权限（业务选择器统一检查 biz_access 权限）
function isBizAuthorized(bizId: number): boolean {
  if (!authStore.authorizedLoaded) return true; // 未加载完成时默认有权限，避免闪烁
  return authStore.hasAuthorizedBiz(BIZ_ACCESS_ACTION, bizId);
}

const filteredBusinessList = computed(() => mainStore.businessList);

// ===== 业务列表（本地排序副本，不修改 store 原数组）=====
const businessList = ref<Business[]>([]);

// ===== 收藏（同步初始化，避免硬刷新时收藏图标闪烁丢失）=====
const collectList = ref<number[]>([]);
(() => {
  try {
    const raw = localStorage.getItem('collect');
    if (raw) collectList.value = JSON.parse(raw);
  } catch { /* ignore */ }
})();

// ===== 多选业务 =====
const multiBusiness = ref<number[]>([]);

// ===== 策略单选业务 =====
const singleBusiness = ref<number | ''>('');

const hasExplicitlyClearedMultiBusiness = () => {
  const bizIdsJson = localStorage.getItem('bk_biz_id');
  if (!bizIdsJson) return false;

  try {
    const bizIds = JSON.parse(bizIdsJson);
    return Array.isArray(bizIds) && bizIds.length === 0;
  } catch {
    return false;
  }
};

const hasPersistedMultiBusiness = () => {
  const bizIdsJson = localStorage.getItem('bk_biz_id');
  if (!bizIdsJson) return false;

  try {
    const bizIds = JSON.parse(bizIdsJson);
    return Array.isArray(bizIds);
  } catch {
    return false;
  }
};

// ===== 节点管理模块：权限加载后，多选模式无选中业务时默认选中第一个有权限的业务 =====
watch(() => authStore.authorizedLoaded, (loaded) => {
  if (
    !loaded
    || isSingle.value
    || multiBusiness.value.length > 0
    || filteredBusinessList.value.length === 0
    || hasPersistedMultiBusiness()
    || hasExplicitlyClearedMultiBusiness()
  ) return;
  // 排序后取第一个有权限的，都没权限则不选
  sortBusinessList();
  const defaultBizId = getDefaultBizId();
  if (defaultBizId !== undefined) {
    multiBusiness.value = [defaultBizId];
    mainStore.updateCurBusiness([defaultBizId]);
    localStorage.setItem('bk_biz_id', JSON.stringify([defaultBizId]));
  }
}, { immediate: true });

const syncMultiBusiness = (bizIds: number[] = []) => {
  mainStore.updateCurBusiness(bizIds);
  localStorage.setItem('bk_biz_id', JSON.stringify(bizIds));
};

// ===== 多选业务持久化 & 同步到 store =====
watch(multiBusiness, (val) => {
  if (isSingle.value) return;
  syncMultiBusiness(val);
}, { deep: true });

// ===== 策略单选业务持久化 & 同步到 store =====
watch(singleBusiness, (val) => {
  if (val !== '' && val !== undefined) {
    localStorage.setItem('strategy_biz_id', String(val));
    mainStore.updateStrategyBizId(val);
  }
});

// ===== hover 无权限业务时跟随鼠标的锁图标（使用 use-auth-lock hook）=====
const {
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
} = useAuthLock(
  BIZ_ACCESS_ACTION, // 业务选择器统一使用 biz_access 权限
  () => mainStore.selectedBusinessId[0],
);

// 业务选择器中每个 option 的 action 由路由决定，需根据 isBizAuthorized 判断
const handleOptionMouseEnter = (e: MouseEvent, bizId: number) => {
  authLockMouseEnter(e, isBizAuthorized(bizId));
};
const handleOptionMouseMove = (e: MouseEvent, bizId: number) => {
  authLockMouseMove(e, isBizAuthorized(bizId));
};
const handleOptionMouseLeave = () => {
  authLockMouseLeave();
};

// ===== 文字溢出检测 =====
const textOverflowMap = reactive<Record<number, boolean>>({});
const handleTextMouseenter = (e: MouseEvent, id: number) => {
  const el = e.target as HTMLElement;
  textOverflowMap[id] = el.scrollWidth > el.clientWidth;
};

// ===== 收起状态显示文案 =====
const shrinkText = computed(() => {
  // 都没有 biz_access 权限时不显示文字
  if (authStore.authorizedLoaded && hasNoAuthorizedBiz()) return '';

  if (isSingle.value) {
    if (!singleBusiness.value) return t('platform.nodeMan.all');
    const biz = mainStore.businessList.find(b => b.bk_biz_id === singleBusiness.value);
    return biz?.bk_biz_name?.[0] || '';
  }
  if (!mainStore.selectedBusinessName.length) {
    return t('platform.nodeMan.all');
  }
  const len = mainStore.selectedBusinessName.length;
  return len > 1 ? len : mainStore.selectedBusinessName[0]?.[0];
});

/** 是否没有任何业务有 biz_access 权限（true=都没权限，应隐藏文字） */
function hasNoAuthorizedBiz(): boolean {
  if (!authStore.authorizedLoaded) return false; // 未加载完默认有权限
  const ids = authStore.getAuthorizedBizIds(BIZ_ACCESS_ACTION);
  // null 表示全有权限
  if (ids === null) return false;
  // 有权限列表为空，或与业务列表无交集 → 都没权限
  return !mainStore.businessList.some(b => ids.includes(String(b.bk_biz_id)));
}

// ===== 排序逻辑（拷贝后排序，不影响 store 原数组）=====
const sortBusinessList = () => {
  businessList.value = [...filteredBusinessList.value].sort((a: Business, b: Business) => {
    // 优先级1: 有权限的排在前面
    const aAuthorized = isBizAuthorized(a.bk_biz_id);
    const bAuthorized = isBizAuthorized(b.bk_biz_id);
    if (aAuthorized && !bAuthorized) return -1;
    if (!aAuthorized && bAuthorized) return 1;

    const aIsCollected = collectList.value.includes(a.bk_biz_id);
    const bIsCollected = collectList.value.includes(b.bk_biz_id);

    // 策略单选模式按 singleBusiness 判断选中，多选模式按 multiBusiness 判断选中
    const aIsSelected = isSingle.value
      ? a.bk_biz_id === singleBusiness.value
      : multiBusiness.value.includes(a.bk_biz_id);
    const bIsSelected = isSingle.value
      ? b.bk_biz_id === singleBusiness.value
      : multiBusiness.value.includes(b.bk_biz_id);

    // 优先级2: 选中状态（选中的排在前面）
    if (aIsSelected && !bIsSelected) return -1;
    if (!aIsSelected && bIsSelected) return 1;
    // 优先级3: 收藏状态（收藏在前）
    if (aIsCollected && !bIsCollected) return -1;
    if (!aIsCollected && bIsCollected) return 1;
    // 优先级4: 按ID从小到大
    return a.bk_biz_id - b.bk_biz_id;
  });
};

// ===== 下拉展开时排序 =====
const handleToggle = () => {
  sortBusinessList();
};

// ===== 收藏操作 =====
const handleCollect = (val: number) => {
  if (collectList.value.includes(val)) {
    collectList.value = collectList.value.filter(item => item !== val);
  } else {
    collectList.value.push(val);
  }
  localStorage.setItem('collect', JSON.stringify(collectList.value));
};

// ===== 点击无权限业务，申请当前菜单页的查看权限 + 业务访问权限 =====
const permissionStore = usePermissionStore();

// 统一处理 option 点击：无权限时阻止选中并触发申请，有权限时不干预（让 Select 正常选中）
const handleOptionClick = (e: MouseEvent, bizId: number) => {
  if (!isBizAuthorized(bizId)) {
    e.stopPropagation(); // 阻止冒泡到 Select.Option，防止被选中
    e.preventDefault();
    handleApplyPermission(bizId);
  }
  // 有权限时不管，让事件正常冒泡给 Select.Option 处理选中
};

const handleApplyPermission = async (bizId: number) => {
  const action = getActionForRoute();
  if (!action) return;

  const authItems = [
    { id: action, action, resourceType: 'biz', routes: [] },
    { id: BIZ_ACCESS_ACTION, action: BIZ_ACCESS_ACTION, resourceType: 'biz', routes: [] },
  ];
  await authStore.batchVerify(authItems, bizId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
};

// ===== 多选业务变更 =====
const handleMultiChange = (val: number[]) => {
  syncMultiBusiness(val);
};

// ===== 自定义搜索方法 =====
const filterOption = (input: any, options: { id: number; name: string }) => {
  const inputStr = String(input).trim();
  if (!inputStr) return false;

  const keywords = inputStr.split(/[\s,;]+/).filter(keyword => keyword.trim());
  if (keywords.length === 0) return false;

  const nameMatch = keywords.some((keyword) => {
    const safeKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const nameRegex = new RegExp(safeKeyword, 'i');
    return options.name?.match(nameRegex);
  });

  const idMatch = keywords.some((keyword) => {
    return String(keyword).trim() === String(options.id);
  });

  return nameMatch || idMatch;
};

const getDefaultBizId = () => {
  // 排序后第一个就是最有权限优先级的，无权限说明都没权限
  if (businessList.value.length === 0) return undefined;
  const first = businessList.value[0];
  return isBizAuthorized(first.bk_biz_id) ? first.bk_biz_id : undefined;
};

const getPersistedStrategyBizId = () => {
  const savedBizId = localStorage.getItem('strategy_biz_id');
  if (savedBizId) {
    const parsedId = Number(savedBizId);
    // 缓存值仍需有权限才用
    if (isBizAuthorized(parsedId)) {
      const exists = filteredBusinessList.value.some(b => b.bk_biz_id === parsedId);
      if (exists) return parsedId;
    }
  }

  return getDefaultBizId();
};

// ===== 路由切换到策略页面时，自动初始化单选业务 =====
watch(isSingle, (val) => {
  if (val && (singleBusiness.value === '' || singleBusiness.value === undefined) && filteredBusinessList.value.length > 0) {
    // 权限未加载完时不初始化，等加载完后 watcher 会再触发
    if (!authStore.authorizedLoaded) return;
    sortBusinessList();
    const persistedBizId = getPersistedStrategyBizId();
    if (persistedBizId !== undefined) {
      singleBusiness.value = persistedBizId;
    }
  }

  // 从策略（单选）切回其他模块（多选）时，从 localStorage.bk_biz_id 恢复多选选中状态
  if (!val) {
    const bizIdsJson = localStorage.getItem('bk_biz_id');
    if (bizIdsJson) {
      try {
        const bizIds = JSON.parse(bizIdsJson);
        if (Array.isArray(bizIds)) {
          multiBusiness.value = bizIds;
          mainStore.updateCurBusiness(bizIds);
          sortBusinessList();
        }
      } catch { /* ignore */ }
    }
  }
});

// 兜底：当 businessList 异步加载完成且 isSingle 已为 true 时，确保单选业务被初始化
// （覆盖直接刷新策略页面、或 businessList 在 isSingle 之后才加载到的场景）
watch(
  [isSingle, () => filteredBusinessList.value.length],
  ([single, len]) => {
    if (single && len > 0 && (singleBusiness.value === '' || singleBusiness.value === undefined)) {
      // 权限未加载完时不初始化
      if (!authStore.authorizedLoaded) return;
      sortBusinessList();
      const persistedBizId = getPersistedStrategyBizId();
      if (persistedBizId !== undefined) {
        singleBusiness.value = persistedBizId;
      }
    }
  },
  { immediate: true },
);

// ===== 初始化 =====
const init = () => {
  // 1. 恢复收藏列表
  const collectsJson = localStorage.getItem('collect');
  if (collectsJson) {
    collectList.value = JSON.parse(collectsJson);
  }

  // 2. 恢复多选业务
  const bizIdsJson = localStorage.getItem('bk_biz_id');
  if (bizIdsJson) {
    const bizIds = JSON.parse(bizIdsJson);
    multiBusiness.value = bizIds;
    mainStore.updateCurBusiness(bizIds);
  }

  // 3. 排序
  if (filteredBusinessList.value.length > 0) {
    sortBusinessList();
  }

  // 4. 策略模块：恢复/默认单选业务（排序后取第一个，没有则取id最小的）
  if (isSingle.value && filteredBusinessList.value.length > 0 && authStore.authorizedLoaded) {
    const persistedBizId = getPersistedStrategyBizId();
    if (persistedBizId !== undefined) {
      singleBusiness.value = persistedBizId;
    }
  }
};

defineExpose({ init });
</script>

<style lang="postcss" scoped>
.nm-menu-biz {
  :deep(.bk-select-trigger) {
    .bk-input {
      border: none;
    }
    .bk-input--text {
      background: #f0f1f5 !important;
    }
  }
}
.biz-select-option {
  min-height: 32px;
  padding: 0 12px;
  box-sizing: border-box;
  &:hover {
    .nc-not-favorited {
      display: inline;
    }
  }
}
.nm-menu-biz {
  :deep(.bk-option) {
    padding: 0 !important;
  }
  :deep(.bk-option-content) {
    width: 100%;
    padding: 0 !important;
  }
}
</style>

<style lang="postcss">
/* 非 scoped：置灰样式需穿透 Select Option popover */
.unauthorized-biz-row {
  color: #c4c6cc !important;
}
</style>
