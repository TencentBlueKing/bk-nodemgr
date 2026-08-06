<template>
  <Select
    ref="selectRef"
    :class="selectClass"
    v-model="innerValue"
    auto-focus
    filterable
    :list="businessOptions"
    id-key="id"
    display-key="name"
    enable-virtual-render
    :scroll-height="360"
    :min-height="360"
    :popover-min-width="280"
    :filter-option="filterOption"
    show-selected-icon
    selected-style="check"
    :multiple="mode === 'multiple'"
    :clearable="mode === 'multiple'"
    :placeholder="placeholder"
    :popover-options="resolvedPopoverOptions"
    @toggle="handleToggle"
    @change="mode === 'multiple' && emit('change', innerValue)"
    @search-change="handleSearchChange"
  >
    <template #optionRender="{ item }">
      <div
        class="w-full flex items-center biz-select-option overflow-hidden"
        :class="{ 'unauthorized-biz-row': !item.authorized }"
        :title="item.title"
        @click="handleOptionClick($event, item.id)"
        @mouseenter="handleOptionMouseEnter($event, item.id)"
        @mousemove="handleOptionMouseMove($event, item.id)"
        @mouseleave="handleOptionMouseLeave()"
      >
        <Button
          class="mr-[8px] w-[18px] shrink-0"
          text
          @click.stop="handleCollect(item.id)"
        >
          <i
            class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
            v-if="item.collected">
          </i>
          <i
            class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
            v-else>
          </i>
        </Button>
        <div class="truncate">[{{ item.id }}] {{ item.name }}</div>
      </div>
    </template>
  </Select>
</template>

<script setup lang="ts">
import { Button, Select } from 'bkui-vue';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';
import { usePermissionStore } from '@/stores/permission';

const props = withDefaults(defineProps<{
  /** 选中值，v-model 绑定 */
  modelValue: number | string | number[] | string[];
  /** 选择模式：单选 / 多选 */
  mode?: 'single' | 'multiple';
  /** IAM action，仅用于点击无权限业务时申请权限（业务可访问范围统一由 biz_access 判断） */
  action?: string;
  /** 额外需要申请的 action（默认包含 biz_access） */
  extraAuthActions?: string[];
  /** Select 自定义 class */
  selectClass?: string;
  /** 占位文本 */
  placeholder?: string;
  /** popover 配置 */
  popoverOptions?: Record<string, any>;
}>(), {
  mode: 'single',
  action: '',
  extraAuthActions: () => ['biz_access'],
  selectClass: '',
  placeholder: '',
  popoverOptions: () => ({}),
});

const emit = defineEmits<{
  'update:modelValue': [value: number | string | number[] | string[]];
  'change': [value: number | string | number[] | string[]];
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const mainStore = useMainStore();
const permissionStore = usePermissionStore();
const selectRef = ref<{
  virtualRenderRef?: {
    scrollTo: (x: number, y: number) => void;
  };
} | null>(null);

const isPopoverOpen = ref(false);
const searchKeyword = ref('');
const displayedBusinessList = ref<Business[]>([]);

const resolvedPopoverOptions = computed(() => ({
  width: 280,
  maxWidth: 280,
  maxHeight: 400,
  extCls: 'nm-biz-select-popover',
  ...props.popoverOptions,
}));

// ===== v-model =====
const innerValue = computed({
  get: () => props.modelValue,
  set: (val) => {
    emit('update:modelValue', val);
    emit('change', val);
  },
});

// ===== 业务列表 =====
const businessList = computed(() => mainStore.businessList);

// ===== 收藏 =====
const collectList = ref<number[]>([]);

// ===== 选中业务 ID 集合（用于排序） =====
const selectedBizIds = computed(() => {
  if (Array.isArray(innerValue.value)) {
    return (innerValue.value as Array<string | number>).map(id => Number(id)).filter(id => !Number.isNaN(id));
  }
  if (innerValue.value === '' || innerValue.value === undefined || innerValue.value === null) return [];
  const v = Number(innerValue.value);
  return Number.isNaN(v) ? [] : [v];
});

// ===== 权限判断 =====
// 业务可访问范围统一通过 biz_access 判断（与菜单页 view 权限解耦）
// 若传入了 action prop，则叠加校验该 action 的 biz 权限
const BIZ_ACCESS_ACTION = 'biz_access';
const isBizAuthorized = (bizId: number): boolean => {
  if (!authStore.authorizedLoaded) return true; // 未加载完成时默认有权限，避免闪烁
  const hasBizAccess = authStore.hasAuthorizedBiz(BIZ_ACCESS_ACTION, bizId);
  if (!hasBizAccess) return false;
  if (props.action?.length) {
    return authStore.hasAuthorizedBiz(props.action, bizId);
  }
  return true;
};

// ===== 排序逻辑 =====
const sortBusinessList = () => {
  const list = [...businessList.value];
  const bizAccessIds = authStore.getAuthorizedBizIds(BIZ_ACCESS_ACTION);
  const bizAccessSet = bizAccessIds === null ? null : new Set(bizAccessIds);
  const actionIds = props.action
    ? authStore.getAuthorizedBizIds(props.action)
    : null;
  const actionSet = actionIds === null ? null : new Set(actionIds || []);
  const collectedSet = new Set(collectList.value);
  const selectedSet = new Set(selectedBizIds.value);
  list.sort((a, b) => {
    // 有权限的排前面（biz_access 已加载时才生效）
    const aAuth = isAuthorizedWithSets(a.bk_biz_id, bizAccessSet, actionSet);
    const bAuth = isAuthorizedWithSets(b.bk_biz_id, bizAccessSet, actionSet);
    if (aAuth && !bAuth) return -1;
    if (!aAuth && bAuth) return 1;

    const aIsCollected = collectedSet.has(a.bk_biz_id);
    const bIsCollected = collectedSet.has(b.bk_biz_id);
    const aIsSelected = selectedSet.has(a.bk_biz_id);
    const bIsSelected = selectedSet.has(b.bk_biz_id);

    if (aIsSelected && !bIsSelected) return -1;
    if (!aIsSelected && bIsSelected) return 1;
    if (aIsCollected && !bIsCollected) return -1;
    if (!aIsCollected && bIsCollected) return 1;
    return a.bk_biz_id - b.bk_biz_id;
  });
  displayedBusinessList.value = list;
};

const businessOptions = computed(() => {
  const collectedSet = new Set(collectList.value);
  return displayedBusinessList.value.map((item) => {
    const authorized = isBizAuthorized(item.bk_biz_id);
    return {
      id: item.bk_biz_id,
      name: item.bk_biz_name,
      title: authorized
        ? `[${item.bk_biz_id}] ${item.bk_biz_name}`
        : t('components.permission.noPermission'),
      authorized,
      collected: collectedSet.has(item.bk_biz_id),
    };
  });
});

watch(businessList, sortBusinessList, { immediate: true });

function isAuthorizedWithSets(
  bizId: number,
  bizAccessSet: Set<string> | null,
  actionSet: Set<string> | null,
): boolean {
  if (!authStore.authorizedLoaded) return true;
  if (bizAccessSet !== null && !bizAccessSet.has(String(bizId))) return false;
  return actionSet === null || actionSet.has(String(bizId));
}

// ===== 搜索过滤 =====
const filterOption = (input: any, options: { id: number; name: string }) => {
  const inputStr = String(input).trim();
  if (!inputStr) return true;
  const keywords = inputStr.split(/[\s,;]+/).filter(keyword => keyword.trim());
  if (keywords.length === 0) return false;
  const nameMatch = keywords.some((keyword) => {
    const safeKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    return options.name?.match(new RegExp(safeKeyword, 'i'));
  });
  const idMatch = keywords.some(keyword => String(keyword).trim() === String(options.id));
  return nameMatch || idMatch;
};

const handleSearchChange = (val: string) => {
  searchKeyword.value = val;
};

const handleEnterSelectFilteredBusiness = (event: KeyboardEvent) => {
  if (props.mode !== 'multiple' || !isPopoverOpen.value || event.key !== 'Enter' || event.isComposing) return;

  const keyword = searchKeyword.value.trim();
  if (!keyword) return;

  event.preventDefault();
  event.stopPropagation();

  const matchedBizIds = businessOptions.value
    .filter(item => item.authorized)
    .filter(item => filterOption(keyword, { id: item.id, name: item.name }))
    .map(item => item.id);
  if (matchedBizIds.length === 0) return;

  const selectedIds = Array.isArray(innerValue.value) ? innerValue.value : [];
  const nextIds = Array.from(new Set([
    ...selectedIds.map(item => Number(item)).filter(item => Number.isFinite(item)),
    ...matchedBizIds,
  ]));

  innerValue.value = nextIds;
};

// ===== 下拉展开状态 =====
const resetVirtualListScroll = () => {
  selectRef.value?.virtualRenderRef?.scrollTo(0, 0);

  const virtualList = document.querySelector(
    '.nm-biz-select-popover .bk-virtual-render',
  ) as HTMLElement | null;
  virtualList?.scrollTo({ left: 0, top: 0 });
};

const handleToggle = (isOpen: boolean) => {
  isPopoverOpen.value = isOpen;
  if (isOpen) {
    sortBusinessList();
    window.setTimeout(resetVirtualListScroll, 20);
  }
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

// ===== hover 无权限锁图标 =====
const {
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
} = useAuthLock(
  BIZ_ACCESS_ACTION,
  () => {
    if (Array.isArray(innerValue.value)) return innerValue.value as number[];
    return innerValue.value ? [Number(innerValue.value)] : [];
  },
);

const handleOptionMouseEnter = (e: MouseEvent, bizId: number) => {
  authLockMouseEnter(e, isBizAuthorized(bizId));
};
const handleOptionMouseMove = (e: MouseEvent, bizId: number) => {
  authLockMouseMove(e, isBizAuthorized(bizId));
};
const handleOptionMouseLeave = () => {
  authLockMouseLeave();
};

// ===== 点击无权限 option → 阻止选中 + 申请权限 =====
const handleOptionClick = async (e: MouseEvent, bizId: number) => {
  if (isBizAuthorized(bizId)) return;
  e.stopPropagation();
  e.preventDefault();
  // 申请权限时：若外部指定了菜单 view action 则同时申请，再叠加 extraAuthActions（默认含 biz_access）
  const actions = [
    ...(props.action ? [props.action] : []),
    ...props.extraAuthActions,
  ];
  if (!actions.length) return;
  const authItems = actions.map(a => ({ id: a, action: a, resourceType: 'biz', routes: [] }));
  await authStore.batchVerify(authItems, bizId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
};

onMounted(() => {
  document.addEventListener('keydown', handleEnterSelectFilteredBusiness, true);
});

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEnterSelectFilteredBusiness, true);
});

// ===== 初始化收藏 =====
onMounted(() => {
  const collectsJson = localStorage.getItem('collect');
  if (collectsJson) {
    try {
      const collects = JSON.parse(collectsJson) as unknown;
      if (Array.isArray(collects)) {
        collectList.value = collects.map(item => Number(item)).filter(item => Number.isFinite(item));
      }
    } catch { /* ignore */ }
  }
});
</script>

<style lang="postcss" scoped>
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
</style>

<style lang="postcss">
/* 非 scoped：置灰样式需穿透 Select Option popover */
.unauthorized-biz-row {
  color: #c4c6cc !important;
}

.nm-biz-select-popover {
  width: 280px !important;
  min-width: 280px !important;
  max-width: 280px !important;
}

.nm-biz-select-popover .bk-select-content-wrapper,
.nm-biz-select-popover .bk-select-content,
.nm-biz-select-popover .bk-select-dropdown {
  width: 100% !important;
}

.nm-biz-select-popover .bk-select-dropdown {
  min-height: 360px !important;
  max-height: 360px !important;
}
</style>
