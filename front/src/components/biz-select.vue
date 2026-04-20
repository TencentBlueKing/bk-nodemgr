<template>
  <Select
    :class="selectClass"
    v-model="innerValue"
    auto-focus
    filterable
    :filter-option="filterOption"
    :show-selected-icon="false"
    :multiple="mode === 'multiple'"
    :clearable="mode === 'multiple'"
    :placeholder="placeholder"
    :popover-options="popoverOptions"
    @toggle="handleToggle"
    @change="mode === 'multiple' && emit('change', innerValue)"
  >
    <Select.Option
      v-for="item in sortedBusinessList"
      :key="item.bk_biz_id"
      :name="item.bk_biz_name"
      :id="item.bk_biz_id"
      v-bk-tooltips="{
        content: isBizAuthorized(item.bk_biz_id)
          ? `[${item.bk_biz_id}] ${item.bk_biz_name}`
          : noPermissionText,
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
        <div
          class="truncate"
          @mouseenter="handleTextMouseenter($event, item.bk_biz_id)">
          [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
        </div>
      </div>
    </Select.Option>
  </Select>
</template>

<script setup lang="ts">
import { Button, Select } from 'bkui-vue';
import { computed, onMounted, reactive, ref } from 'vue';
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

const noPermissionText = t('components.permission.noPermission');

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
const BIZ_ACCESS_ACTION = 'biz_access';
const isBizAuthorized = (bizId: number): boolean => {
  if (!authStore.authorizedLoaded) return true; // 未加载完成时默认有权限，避免闪烁
  return authStore.hasAuthorizedBiz(BIZ_ACCESS_ACTION, bizId);
};

// ===== 排序逻辑 =====
const sortedBusinessList = computed(() => {
  const list = [...businessList.value];
  return list.sort((a, b) => {
    // 有权限的排前面（biz_access 已加载时才生效）
    const aAuth = isBizAuthorized(a.bk_biz_id);
    const bAuth = isBizAuthorized(b.bk_biz_id);
    if (aAuth && !bAuth) return -1;
    if (!aAuth && bAuth) return 1;

    const aIsCollected = collectList.value.includes(a.bk_biz_id);
    const bIsCollected = collectList.value.includes(b.bk_biz_id);
    const aIsSelected = selectedBizIds.value.includes(a.bk_biz_id);
    const bIsSelected = selectedBizIds.value.includes(b.bk_biz_id);

    if (aIsSelected && !bIsSelected) return -1;
    if (!aIsSelected && bIsSelected) return 1;
    if (aIsCollected && !bIsCollected) return -1;
    if (!aIsCollected && bIsCollected) return 1;
    return a.bk_biz_id - b.bk_biz_id;
  });
});

// ===== 搜索过滤 =====
const filterOption = (input: any, options: { id: number; name: string }) => {
  const inputStr = String(input).trim();
  if (!inputStr) return false;
  const keywords = inputStr.split(/[\s,;]+/).filter(keyword => keyword.trim());
  if (keywords.length === 0) return false;
  const nameMatch = keywords.some((keyword) => {
    const safeKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    return options.name?.match(new RegExp(safeKeyword, 'i'));
  });
  const idMatch = keywords.some(keyword => String(keyword).trim() === String(options.id));
  return nameMatch || idMatch;
};

// ===== 下拉展开时排序 =====
const handleToggle = () => {
  sortedBusinessList.value;
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

// ===== 文字溢出检测 =====
const textOverflowMap = reactive<Record<number, boolean>>({});
const handleTextMouseenter = (e: MouseEvent, id: number) => {
  const el = e.target as HTMLElement;
  textOverflowMap[id] = el.scrollWidth > el.clientWidth;
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
</style>
