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
      :placeholder="t('platform.nodeMan.allBusiness')"
      :popover-options="{ boundary: 'document.body', width: '235px' }"
      @toggle="handleToggle"
    >
      <Select.Option
        v-for="item in businessList"
        :key="item.bk_biz_id"
        :name="item.bk_biz_name"
        :id="item.bk_biz_id"
        v-bk-tooltips="{
          content: `[${item.bk_biz_id}] ${item.bk_biz_name}`,
          disabled: !textOverflowMap[item.bk_biz_id],
          boundary: 'parent',
          placement: 'right',
          offset: 10
        }"
      >
        <div class="w-full flex items-center biz-select-option overflow-hidden">
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
      :placeholder="t('platform.nodeMan.allBusiness')"
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
          content: `[${item.bk_biz_id}] ${item.bk_biz_name}`,
          disabled: !textOverflowMap[item.bk_biz_id],
          boundary: 'parent',
          placement: 'right',
          offset: 10
        }"
      >
        <div class="w-full flex items-center biz-select-option overflow-hidden">
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
  </div>
</template>

<script setup lang="ts">
import { Button, Select } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();

defineProps<{
  /** 左侧菜单是否展开 */
  expanded: boolean;
}>();

// ===== 可见性与模式 =====
const isVisible = computed(() => !route.path.includes('topo-manager') && !route.path.includes('pkg-manager'));
const isSingle = computed(() => route.path.includes('rule-manager'));

// ===== 业务列表（本地排序副本，不修改 store 原数组）=====
const businessList = ref<Business[]>([]);

// ===== 收藏 =====
const collectList = ref<number[]>([]);

// ===== 多选业务 =====
const multiBusiness = ref<number[]>([]);

// ===== 策略单选业务 =====
const singleBusiness = ref<number | ''>('');

// ===== 策略单选业务持久化 & 同步到 store =====
watch(singleBusiness, (val) => {
  if (val !== '' && val !== undefined) {
    localStorage.setItem('strategy_biz_id', String(val));
    mainStore.updateStrategyBizId(val);
  }
});

// ===== 文字溢出检测 =====
const textOverflowMap = reactive<Record<number, boolean>>({});
const handleTextMouseenter = (e: MouseEvent, id: number) => {
  const el = e.target as HTMLElement;
  textOverflowMap[id] = el.scrollWidth > el.clientWidth;
};

// ===== 收起状态显示文案 =====
const shrinkText = computed(() => {
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

// ===== 排序逻辑（拷贝后排序，不影响 store 原数组）=====
const sortBusinessList = () => {
  businessList.value = [...mainStore.businessList].sort((a: Business, b: Business) => {
    const aIsCollected = collectList.value.includes(a.bk_biz_id);
    const bIsCollected = collectList.value.includes(b.bk_biz_id);

    // 策略单选模式按 singleBusiness 判断选中，多选模式按 multiBusiness 判断选中
    const aIsSelected = isSingle.value
      ? a.bk_biz_id === singleBusiness.value
      : multiBusiness.value.includes(a.bk_biz_id);
    const bIsSelected = isSingle.value
      ? b.bk_biz_id === singleBusiness.value
      : multiBusiness.value.includes(b.bk_biz_id);

    // 优先级1: 选中状态（选中的排在前面）
    if (aIsSelected && !bIsSelected) return -1;
    if (!aIsSelected && bIsSelected) return 1;
    // 优先级2: 收藏状态（收藏在前）
    if (aIsCollected && !bIsCollected) return -1;
    if (!aIsCollected && bIsCollected) return 1;
    // 优先级3: 按ID从小到大
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

// ===== 多选业务变更 =====
const handleMultiChange = (val: number[]) => {
  mainStore.updateCurBusiness(val);
  localStorage.setItem('bk_biz_id', JSON.stringify(val));
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

// ===== 路由切换到策略页面时，自动初始化单选业务 =====
watch(isSingle, (val) => {
  if (val && (singleBusiness.value === '' || singleBusiness.value === undefined) && mainStore.businessList.length > 0) {
    sortBusinessList();
    const savedBizId = localStorage.getItem('strategy_biz_id');
    const defaultBiz = businessList.value[0]?.bk_biz_id
      ?? [...mainStore.businessList].sort((a, b) => a.bk_biz_id - b.bk_biz_id)[0].bk_biz_id;
    if (savedBizId) {
      const parsedId = Number(savedBizId);
      const exists = mainStore.businessList.some(b => b.bk_biz_id === parsedId);
      singleBusiness.value = exists ? parsedId : defaultBiz;
    } else {
      singleBusiness.value = defaultBiz;
    }
  }
});

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
  if (mainStore.businessList.length > 0) {
    sortBusinessList();
  }

  // 4. 策略模块：恢复/默认单选业务（排序后取第一个，没有则取id最小的）
  if (isSingle.value && mainStore.businessList.length > 0) {
    const savedBizId = localStorage.getItem('strategy_biz_id');
    // 排序后的第一个作为默认值，兜底取id最小的
    const defaultBiz = businessList.value[0]?.bk_biz_id
      ?? [...mainStore.businessList].sort((a, b) => a.bk_biz_id - b.bk_biz_id)[0].bk_biz_id;
    if (savedBizId) {
      const parsedId = Number(savedBizId);
      const exists = mainStore.businessList.some(b => b.bk_biz_id === parsedId);
      singleBusiness.value = exists ? parsedId : defaultBiz;
    } else {
      singleBusiness.value = defaultBiz;
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
  &:hover {
    .nc-not-favorited {
      display: inline;
    }
  }
}
</style>
