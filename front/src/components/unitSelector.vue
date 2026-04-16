<template>
  <Select
    :class="selectClass"
    v-model="internalValue"
    :loading="localLoading"
    :clearable="false"
    :filterable="true"
    :disabled="disabled"
    auto-focus
    @change="handleSelectChange"
  >
    <Select.Option v-if="noLimit" :label="t('agentStrategy.form.unlimited')" value="-1"></Select.Option>
    <Select.Option
      v-for="option in filteredList"
      :key="option.bk_networkunit_id"
      :id="String(option.bk_networkunit_id)"
      :name="option.bk_networkunit_name"
      :disabled="disableDirect && option.is_direct"
      :class="{ 'unauthorized-unit-row': !isUnitAuthorized(option.bk_networkunit_id) }"
      v-bk-tooltips="{
        content: directTip,
        disabled: !(disableDirect && option.is_direct),
        boundary: 'parent',
        placement: 'left'
      }"
    >
      <div
        class="w-full flex items-center unit-select-option"
        @click="handleOptionClick($event, option.bk_networkunit_id)"
        @mouseenter="handleOptionMouseEnter($event, option.bk_networkunit_id)"
        @mousemove="handleOptionMouseMove($event, option.bk_networkunit_id)"
        @mouseleave="handleOptionMouseLeave()"
      >
        <span class="truncate">[{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}</span>
      </div>
    </Select.Option>
  </Select>
</template>

<script setup lang="ts">
import { Select } from 'bkui-vue';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import useAuthLock from '@/composables/use-auth-lock';
import { TopoService } from '@/api/modules/topo';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const props = withDefaults(defineProps<{
  /** 绑定值（bk_networkunit_id） */
  modelValue?: number | string;
  /** 按管控区域 ID 过滤 */
  filterByAreaId?: number | string;
  /** 是否显示"不限"选项 */
  noLimit?: boolean;
  /** 是否禁用 is_direct 的选项（Proxy 安装场景） */
  disableDirect?: boolean;
  /** is_direct 选项禁用时的提示文案 */
  directTip?: string;
  /** 整体禁用 */
  disabled?: boolean;
  /** Select 自定义 class */
  selectClass?: string;
  /** 使用 Brief 接口（轻量，不含 direct_endpoints） */
  useBrief?: boolean;
  /** 按区域 ID 数组过滤（install-proxy 场景，多个区域） */
  filterByAreaIds?: number[];
  /** 外部传入选项列表（传入时组件不自加载数据） */
  options?: any[];
}>(), {
  modelValue: undefined,
  filterByAreaId: undefined,
  noLimit: false,
  disableDirect: false,
  directTip: '',
  disabled: false,
  selectClass: '',
  useBrief: true,
  filterByAreaIds: undefined,
  options: undefined,
});

const emit = defineEmits<{
  (e: 'update:modelValue', id: number | string): void;
  (e: 'change', id: number | string, row: any): void;
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();

// ===== 权限控制 =====
const {
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
} = useAuthLock('networkunit_view', () => undefined, { resourceType: 'networkunit' });

function isUnitAuthorized(unitId: number | string): boolean {
  if (!authStore.authorizedLoaded) return true;
  return authStore.hasAuthorizedResource('networkunit_view', unitId);
}

const handleOptionMouseEnter = (e: MouseEvent, unitId: number | string) => {
  authLockMouseEnter(e, isUnitAuthorized(unitId));
};
const handleOptionMouseMove = (e: MouseEvent, unitId: number | string) => {
  authLockMouseMove(e, isUnitAuthorized(unitId));
};
const handleOptionMouseLeave = () => {
  authLockMouseLeave();
};

const handleOptionClick = (e: MouseEvent, unitId: number | string) => {
  if (!isUnitAuthorized(unitId)) {
    e.stopPropagation();
    e.preventDefault();
    handleApplyPermission(unitId);
  }
};

const handleApplyPermission = async (unitId: number | string) => {
  await authStore.batchVerify([
    { id: 'networkunit_view', action: 'networkunit_view', resourceType: 'networkunit', routes: [] },
  ], undefined, unitId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
};

// ===== 数据加载 =====
const localLoading = ref(false);
const unitList = ref<any[]>([]);
const internalValue = ref<string>(props.modelValue !== undefined ? String(props.modelValue) : '');

/** 是否使用外部传入的 options */
const useExternalOptions = computed(() => Array.isArray(props.options));

const filteredList = computed(() => {
  let list = useExternalOptions.value ? props.options! : unitList.value;
  if (props.filterByAreaIds && props.filterByAreaIds.length > 0) {
    list = list.filter((item: any) => props.filterByAreaIds!.includes(item.bk_networkarea_id));
  } else if (props.filterByAreaId !== undefined && props.filterByAreaId !== '') {
    list = list.filter((item: any) => item.bk_networkarea_id === Number(props.filterByAreaId) || item.bk_networkunit_id === -1);
  }
  // 排序：有权限的在前
  return [...list].sort((a, b) => {
    const aAuth = isUnitAuthorized(a.bk_networkunit_id);
    const bAuth = isUnitAuthorized(b.bk_networkunit_id);
    if (aAuth && !bAuth) return -1;
    if (!aAuth && bAuth) return 1;
    return a.bk_networkunit_id - b.bk_networkunit_id;
  });
});

const fetchUnitList = async () => {
  if (useExternalOptions.value) return; // 外部传入时不自加载
  localLoading.value = true;
  try {
    const res = props.useBrief
      ? await TopoService.NetworkUnitListBrief({})
      : await TopoService.NetworkUnitList({});
    unitList.value = res.items || [];
  } catch {
    unitList.value = [];
  } finally {
    localLoading.value = false;
  }
};

onMounted(async () => {
  const tasks: Promise<any>[] = [fetchUnitList()];
  // 外部传入 options 时仍需加载权限信息用于置灰/锁图标展示
  tasks.push(authStore.fetchAuthorized(
    [{ action: 'networkunit_view', resource_type: 'networkunit' }],
    'unitSelector_networkunit_view',
  ));
  await Promise.all(tasks);
});

watch(() => props.modelValue, (val) => {
  if (val !== undefined) {
    internalValue.value = String(val);
  }
});

const handleSelectChange = (val: any) => {
  const numVal = val === '-1' ? '-1' : Number(val);
  emit('update:modelValue', numVal);
  const sourceList = useExternalOptions.value ? props.options! : unitList.value;
  const target = sourceList.find((item: any) => String(item.bk_networkunit_id) === val);
  emit('change', numVal, target || null);
};
</script>

<style lang="postcss">
/* 非 scoped：置灰样式需穿透 Select Option popover */
.unauthorized-unit-row {
  color: #c4c6cc !important;
  cursor: pointer !important;
  pointer-events: auto !important;

  &:hover {
    color: #c4c6cc !important;
    background-color: #f5f7fa !important;
  }
}
</style>
