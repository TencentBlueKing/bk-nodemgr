<template>
  <Select
    class="w-[240px] area-selector-custom"
    v-model="internalValue"
    :prefix="noLimit ? t('platform.nodeMan.installAgentPage.cloud') : ''"
    :loading="localLoading"
    :clearable="false"
    :multiple="multiple"
    :filterable="true"
    collapse-tags
    :show-all="multiple"
    :all-option-id="multiple ? 'all' : undefined"
    :disabled="disabled"
    @change="handleSelectChange"
  >
    <Select.Option v-if="noLimit" id="-1" name="不限" value="-1"></Select.Option>
    <Select.Group :label="$t('topoManager.topo.select.default')">
      <Select.Option
        v-if="defaultArea"
        :key="defaultArea.bk_networkarea_id"
        :id="String(defaultArea.bk_networkarea_id)"
        :name="defaultArea.bk_networkarea_name"
        :value="String(defaultArea.bk_networkarea_id)"
      >
        <div class="flex items-center group/item">
          <Button text class="mr-2 w-[18px]">
            <i
              class="nodeman-icon nc-collect text-[#C4C6CC] text-[16px]"
            ></i>
          </Button>
          <span class="truncate">
            {{ `[${defaultArea.bk_networkarea_id}] ${defaultArea.bk_networkarea_name}` }}
          </span>
        </div>
      </Select.Option>
    </Select.Group>

    <Select.Group :label="$t('topoManager.topo.select.other')">
      <Select.Option
        v-for="item in sortedOtherList"
        :key="item.bk_networkarea_id"
        :id="String(item.bk_networkarea_id)"
        :name="item.bk_networkarea_name"
        :value="String(item.bk_networkarea_id)"
        v-bk-tooltips="{
          content: isAreaAuthorized(item.bk_networkarea_id)
            ? `[${item.bk_networkarea_id}] ${item.bk_networkarea_name}`
            : t('components.permission.noPermission'),
          disabled: isAreaAuthorized(item.bk_networkarea_id) && !textOverflowMap[item.bk_networkarea_id],
          boundary: 'parent',
          placement: 'right',
          offset: 10
        }"
      >
        <div
          class="w-full flex items-center area-select-option overflow-hidden"
          :class="{ 'unauthorized-area-row': !isAreaAuthorized(item.bk_networkarea_id) }"
          @click="handleOptionClick($event, item.bk_networkarea_id)"
          @mouseenter="handleOptionMouseEnter($event, item.bk_networkarea_id)"
          @mousemove="handleOptionMouseMove($event, item.bk_networkarea_id)"
          @mouseleave="handleOptionMouseLeave()"
        >
          <Button text class="mr-2 w-[18px] shrink-0" @click.stop="handleCollect(item.bk_networkarea_id)">
            <i
              v-if="isFavorited(item.bk_networkarea_id)"
              class="nodeman-icon nc-collect text-[#ffb848] text-[14px]"
            ></i>
            <i
              v-else
              class="nodeman-icon nc-not-favorited text-[#63656e] text-[14px] hidden"
            ></i>
          </Button>
          <div
            class="truncate"
            @mouseenter="handleTextMouseenter($event, item.bk_networkarea_id)"
          >
            {{ `[${item.bk_networkarea_id}] ${item.bk_networkarea_name}` }}
          </div>
        </div>
      </Select.Option>
    </Select.Group>
  </Select>
</template>

<script setup lang="ts">
import { Button, Select } from 'bkui-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import { useTopoStore } from '@/stores/topo';
import { useWorkareaStore } from '@/stores/workarea';

const props = withDefaults(defineProps<{
  multiple?: boolean;
  noLimit?: boolean;
  bk_networkarea_id?: number | string;
  disabled?: boolean;
}>(), {
  multiple: true,
  noLimit: false,
  bk_networkarea_id: undefined,
  disabled: false,
});

const emit = defineEmits<{
  (e: 'change', ids: any, rows: any[]): void
}>();

const { t } = useI18n();
const topoStore = useTopoStore();
const workareaStore = useWorkareaStore();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();

// ===== 权限控制（参考 biz-selector 模式）=====
const {
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
} = useAuthLock(
  'networkarea_view',
  () => undefined, // networkarea 不是 biz 级别，不需要 bizId
);

/** 判断某个管控区域是否有权限 */
function isAreaAuthorized(areaId: number): boolean {
  if (!authStore.authorizedLoaded) return false; // 未加载完成时默认无权限
  return authStore.hasAuthorizedResource('networkarea_view', areaId);
}

const handleOptionMouseEnter = (e: MouseEvent, areaId: number) => {
  authLockMouseEnter(e, isAreaAuthorized(areaId));
};
const handleOptionMouseMove = (e: MouseEvent, areaId: number) => {
  authLockMouseMove(e, isAreaAuthorized(areaId));
};
const handleOptionMouseLeave = () => {
  authLockMouseLeave();
};

/** 点击无权限区域 → 阻止选中 + 触发权限申请 */
const handleOptionClick = (e: MouseEvent, areaId: number) => {
  if (!isAreaAuthorized(areaId)) {
    e.stopPropagation();
    e.preventDefault();
    handleApplyPermission(areaId);
  }
};

const handleApplyPermission = async (areaId: number) => {
  await authStore.batchVerify([
    { id: 'networkarea_view', action: 'networkarea_view', resourceType: 'networkarea', routes: [] },
  ], undefined, areaId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
};

// ===== 原有逻辑 =====
const localLoading = ref(false);
// 内部统一用字符串管理 ID，解决数字 0 的显示 Bug
const internalValue = ref<any>(props.multiple ? [] : '');
const textOverflowMap = reactive<Record<number, boolean>>({});

onMounted(async () => {
  localLoading.value = true;
  try {
    // 并行：加载区域列表 + 权限数据
    await Promise.all([
      topoStore.handleFetchTopoWorkareaList(),
      authStore.fetchAuthorized(
        [{ action: 'networkarea_view', resource_type: 'networkarea' }],
        'areaSelector_networkarea_view',
      ),
    ]);
    workareaStore.syncFavoriteWorkareaList();

    // 优先使用 bk_networkarea_id prop 初始化（注意：0 = Default Area 也要识别）
    if (props.bk_networkarea_id !== undefined && props.bk_networkarea_id !== null) {
      internalValue.value = String(props.bk_networkarea_id);
    } else if (props.noLimit) {
      internalValue.value = '-1';
    }
  } finally {
    localLoading.value = false;
  }
});

// 外部 prop 变化时同步到 internalValue（支持 view 模式初始值回填）
watch(() => props.bk_networkarea_id, (val) => {
  if (val !== undefined && val !== null) {
    internalValue.value = String(val);
  }
});

// 计算属性和交互逻辑保持不变...
const defaultArea = computed(() => topoStore.allWorkareaList.find(i => i.bk_networkarea_id === 0));

const sortedOtherList = computed(() => {
  const others = topoStore.allWorkareaList.filter(i => i.bk_networkarea_id !== 0);
  return [...others].sort((a, b) => {
    // 优先级1: 有权限的排在最前面
    const aAuth = isAreaAuthorized(a.bk_networkarea_id);
    const bAuth = isAreaAuthorized(b.bk_networkarea_id);
    if (aAuth && !bAuth) return -1;
    if (!aAuth && bAuth) return 1;

    const aId = String(a.bk_networkarea_id);
    const bId = String(b.bk_networkarea_id);
    const aSel = props.multiple ? internalValue.value.includes(aId) : internalValue.value === aId;
    const bSel = props.multiple ? internalValue.value.includes(bId) : internalValue.value === bId;
    // 优先级2: 选中状态
    if (aSel !== bSel) return aSel ? -1 : 1;
    // 优先级3: 收藏状态
    if (isFavorited(a.bk_networkarea_id) !== isFavorited(b.bk_networkarea_id)) {
      return isFavorited(a.bk_networkarea_id) ? -1 : 1;
    }
    // 优先级4: 按 ID 从小到大
    return b.bk_networkarea_id - a.bk_networkarea_id;
  });
});

const isFavorited = (id: number) => workareaStore.favoriteWorkareaList.includes(id);

const handleCollect = (id: number) => {
  let collectList = JSON.parse(localStorage.getItem('collect_workarea') || '[]');
  collectList = collectList.includes(id) ? collectList.filter((i: number) => i !== id) : [...collectList, id];
  localStorage.setItem('collect_workarea', JSON.stringify(collectList));
  workareaStore.syncFavoriteWorkareaList();
};

const handleTextMouseenter = (e: MouseEvent, id: number) => {
  const el = e.target as HTMLElement;
  textOverflowMap[id] = el.scrollWidth > el.clientWidth;
};

// 核心：数据回传时还原类型
const handleSelectChange = (val: any) => {
  let selectedRows = [];
  let rawIds: any;

  if (props.multiple) {
    // 处理多选：'all' 保持字符串，数字 ID 转回 Number
    rawIds = val.map((v: string) => (v === 'all' ? 'all' : Number(v)));
    selectedRows = topoStore.allWorkareaList.filter(item => rawIds.includes('all') || rawIds.includes(item.bk_networkarea_id));
  } else {
    // 处理单选：转回 Number
    rawIds = val;
    const target = topoStore.allWorkareaList.find(item => String(item.bk_networkarea_id) === rawIds);
    selectedRows = target ? [target] : [];
  }

  emit('change', rawIds, selectedRows);
};
</script>

<style lang="postcss" scoped>
.area-select-option {
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
.unauthorized-area-row {
  color: #c4c6cc !important;
}
</style>
