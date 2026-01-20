<template>
  <Select
    class="w-[240px] area-selector-custom"
    v-model="internalValue"
    :prefix="noLimit ? '管控区域' : ''"
    :loading="localLoading"
    :clearable="false"
    :multiple="multiple"
    :filterable="true"
    collapse-tags
    :show-all="multiple"
    :all-option-id="multiple ? 'all' : undefined"
    @change="handleSelectChange"
  >
    <Select.Option v-if="noLimit" label="不限" value="-1"></Select.Option>
    <Select.Group :label="$t('topoManager.topo.select.default')">
      <Select.Option
        v-if="defaultArea"
        :key="defaultArea.bk_networkarea_id"
        :id="String(defaultArea.bk_networkarea_id)"
        :name="defaultArea.bk_networkarea_name"
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
      >
        <div class="flex items-center group/item overflow-hidden">
          <Button text class="mr-2 w-[18px] shrink-0" @click.stop="handleCollect(item.bk_networkarea_id)">
            <i
              v-if="isFavorited(item.bk_networkarea_id)"
              class="nodeman-icon nc-collect text-[#ffb848] text-[14px]"
            ></i>
            <i
              v-else
              class="nodeman-icon nc-not-favorited text-[#C4C6CC] text-[14px] opacity-0 group-hover/item:opacity-100"
            ></i>
          </Button>
          <div
            class="flex-1 truncate"
            @mouseenter="handleTextMouseenter($event, item.bk_networkarea_id)"
            v-bk-tooltips="{
              content: `[${item.bk_networkarea_id}] ${item.bk_networkarea_name}`,
              disabled: !textOverflowMap[item.bk_networkarea_id],
              boundary: 'body'
            }"
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
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useTopoStore } from '@/stores/topo';
import { useWorkareaStore } from '@/stores/workarea';

const props = withDefaults(defineProps<{
  multiple?: boolean;
  noLimit?: boolean;
  bk_networkarea_id?: number;
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

const localLoading = ref(false);
// 内部统一用字符串管理 ID，解决数字 0 的显示 Bug
const internalValue = ref<any>(props.multiple ? [] : '');
const textOverflowMap = reactive<Record<number, boolean>>({});

onMounted(async () => {
  localLoading.value = true;
  try {
    await topoStore.handleFetchTopoWorkareaList();
    workareaStore.syncFavoriteWorkareaList();

    if (props.noLimit) {
      internalValue.value = '-1';
    }
    if (props.bk_networkarea_id) {
      internalValue.value = String(props.bk_networkarea_id);
    }
    // 初始值赋值逻辑
    // const favoriteIds = workareaStore.favoriteWorkareaList;
    // if (props.multiple) {
    //   // 多选：优先收藏，无收藏默认 ['all']
    //   internalValue.value = favoriteIds.length > 0
    //     ? favoriteIds.map(id => String(id))
    //     : ['all'];
    // } else {
    //   // 单选：优先收藏第一个，无收藏默认 ''
    //   internalValue.value = favoriteIds.length > 0
    //     ? String(favoriteIds[0])
    //     : '';
    // }

    // handleSelectChange(internalValue.value);
  } finally {
    localLoading.value = false;
  }
});

// 计算属性和交互逻辑保持不变...
const defaultArea = computed(() => topoStore.allWorkareaList.find(i => i.bk_networkarea_id === 0));

const sortedOtherList = computed(() => {
  const others = topoStore.allWorkareaList.filter(i => i.bk_networkarea_id !== 0);
  return [...others].sort((a, b) => {
    const aId = String(a.bk_networkarea_id);
    const bId = String(b.bk_networkarea_id);
    const aSel = props.multiple ? internalValue.value.includes(aId) : internalValue.value === aId;
    const bSel = props.multiple ? internalValue.value.includes(bId) : internalValue.value === bId;
    if (aSel !== bSel) return aSel ? -1 : 1;
    if (isFavorited(a.bk_networkarea_id) !== isFavorited(b.bk_networkarea_id)) {
      return isFavorited(a.bk_networkarea_id) ? -1 : 1;
    }
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
