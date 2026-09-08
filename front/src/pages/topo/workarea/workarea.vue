<template>
  <div class="p-[24px]">
    <FlexRow>
      <template #left>
        <div class="flex items-center">
          <!-- 新建 -->
          <Button
            theme="primary"
            class="mr-[8px]"
            :class="{ 'unAuthorized': !hasCreateAuth }"
            @click="hasCreateAuth ? handleCreateWorkarea() : createAuthClick($event)"
            @mouseenter="createMouseEnter($event, hasCreateAuth)"
            @mousemove="createMouseMove($event, hasCreateAuth)"
            @mouseleave="createMouseLeave()"
          >
            <i class="nodeman-icon nc-plus-line mr-[4.5px]"></i>
            <span>{{ $t('action.create') }}</span>
          </Button>
          <Dropdown
            theme="light"
            trigger="click"
            :popover-options="{ clickContentAutoHide: true }"
          >
            <Button :disabled="selectedAreas.length === 0">
              <span>{{ $t('topoManager.workArea.batchCreate.batchOperate') }}</span>
              <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
            </Button>
            <template #content>
              <Dropdown.DropdownMenu>
                <Dropdown.DropdownItem
                  :class="{ 'batch-item-disabled': isBatchCreateDisabled }"
                  :aria-disabled="isBatchCreateDisabled"
                  v-bk-tooltips="{
                    content: batchCreateDisabledTip,
                    disabled: !isBatchCreateDisabled,
                  }"
                  @click="handleBatchCreateClick"
                >
                  {{ $t('topoManager.workArea.batchCreate.createDefaultUnit') }}
                </Dropdown.DropdownItem>
              </Dropdown.DropdownMenu>
            </template>
          </Dropdown>
        </div>
      </template>
      <template #right>
        <SearchSelect
          :max-height="240"
          class="w-[480px] bg-[#fff]"
          unique-select
          :placeholder="$t('topoManager.workArea.search.placeholder')"
          v-model.trim="searchKey"
          :data="searchSelectData">
        </SearchSelect>
      </template>
    </FlexRow>
    <RegionTable
      ref="regionTableRef"
      :list="displayTableData"
      :vendor-list="workareaStore.vendorList"
      :selected-ids="selectedAreaIds"
      @edit="handleEditWorkarea"
      @filter="handleFilter"
      @selection-change="handleSelectionChange"
      @clear-selection="clearSelection"
    />
    <UpsertWorkarea
      v-model:is-show="showUpsertWorkarea"
      :is-create="isCreate"
      :cur-workarea-data="curWorkareaData"
      @install-proxy="handleInstallProxy"
      @update="handleUpdate"
    />
    <BatchCreateDefaultWorkUnit
      v-model:is-show="showBatchCreateWorkUnit"
      :selected-areas="selectedAreas"
      @submitted="handleBatchCreateSubmitted"
    />
    <InstallProxy v-bind:is-show="isInstallProxyShow" />
  </div>
</template>

<script setup lang="ts">
import { Button, Dropdown, Message, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import InstallProxy from '../install-proxy/install-proxy.vue';

import BatchCreateDefaultWorkUnit from './components/batch-create-default-work-unit.vue';
import RegionTable from './components/region-table.vue';
import UpsertWorkarea from './components/upsert-workarea.vue';
import { vendorMap } from './vendorMap';

import type {
  TopoNetworkUnitCreateDefaultMultiResp,
} from '@/@types/topo';
import useAuthLock from '@/composables/use-auth-lock';
import { getModuleAuthorizedItems } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';
import type { INetWorkArea } from '@/stores/workarea';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const authStore = useAuthStore();
const workareaStore = useWorkareaStore();

const selectedAreaMap = ref(new Map<number, INetWorkArea>());
const selectedAreaIds = computed(() => Array.from(selectedAreaMap.value.keys()));
const selectedAreas = computed(() => Array.from(selectedAreaMap.value.values()));
const showBatchCreateWorkUnit = ref(false);

// networkunit_count 为异步统计值，需从 workareaList 实时取最新值判断
const selectedUnitCounts = computed(() => {
  const countById = new Map(workareaStore.workareaList.map(area => [area.bk_networkarea_id, area.networkunit_count]));
  return selectedAreaIds.value.map(id => countById.get(id));
});
const hasUnknownSelected = computed(() => selectedUnitCounts.value.some(count => count === undefined));
const hasNonEmptySelected = computed(() => selectedUnitCounts.value.some(count => (count ?? 0) > 0));
const isBatchCreateDisabled = computed(() => hasUnknownSelected.value || hasNonEmptySelected.value);
const batchCreateDisabledTip = computed(() => (
  hasUnknownSelected.value
    ? t('topoManager.workArea.batchCreate.statisticsUnavailable')
    : t('topoManager.workArea.batchCreate.nonEmptySelection')
));

const handleBatchCreateClick = (event: MouseEvent) => {
  if (isBatchCreateDisabled.value) {
    event.preventDefault();
    event.stopPropagation();
    return;
  }

  showBatchCreateWorkUnit.value = true;
};

const handleSelectionChange = ({ row, checked }: { row: INetWorkArea; checked: boolean }) => {
  if (checked) {
    if (!selectedAreaMap.value.has(row.bk_networkarea_id) && selectedAreaIds.value.length >= 100) {
      Message({
        theme: 'warning',
        message: t('topoManager.workArea.batchCreate.limitReached'),
      });
      return;
    }
    selectedAreaMap.value.set(row.bk_networkarea_id, row);
    return;
  }

  selectedAreaMap.value.delete(row.bk_networkarea_id);
};

const clearSelection = () => {
  selectedAreaMap.value.clear();
};

const handleBatchCreateSubmitted = async (result: TopoNetworkUnitCreateDefaultMultiResp['data']) => {
  const successfulAreaIds = (result.items || [])
    .filter(item => item.success)
    .map(item => item.bk_networkarea_id);
  successfulAreaIds.forEach(id => selectedAreaMap.value.delete(id));
  await getTableData();
};

// networkarea_create permission
const { hasAuth: hasCreateAuth, handleMouseEnter: createMouseEnter, handleMouseMove: createMouseMove, handleMouseLeave: createMouseLeave, handleAuthClick: createAuthClick } = useAuthLock('networkarea_create', () => undefined, { resourceType: 'networkarea' });

// 新增/修改 workarea dialog
const showUpsertWorkarea = ref(false);
const isCreate = ref(true);
const curWorkareaData = ref();
const handleCreateWorkarea = () => {
  isCreate.value = true;
  showUpsertWorkarea.value = true;
};
const handleEditWorkarea = (workareaData: NetworkArea) => {
  isCreate.value = false;
  curWorkareaData.value = workareaData;
  showUpsertWorkarea.value = true;
};

// 安装 proxy dialog
const isInstallProxyShow = ref(false);
const handleInstallProxy = () => {
  isInstallProxyShow.value = true;
};
// 编辑更新
const handleUpdate = async () => {
  await getTableData();
};

// 表格数据
const tableData = ref<INetWorkArea[]>([]);

// 管控单元数量筛选（前端统计值，无法后端查询，前端过滤）
const unitCountFilterChecked = ref<string[]>([]);
const displayTableData = computed(() => {
  if (unitCountFilterChecked.value.length === 0) return tableData.value;
  return tableData.value.filter((row) => {
    if (!Number.isFinite(row.networkunit_count)) return false;
    const isEmpty = row.networkunit_count === 0;
    return unitCountFilterChecked.value.some(v => (v === '0' ? isEmpty : !isEmpty));
  });
});

// 下拉搜索框value、list
const searchKey = ref<ISearchValue[]>([]);
const searchSelectData = ref<ISearchItem[]>([]);
// 初始化searchSelect checkbox options
const initSearchData = () => {
  searchSelectData.value = [
    {
      name: t('topoManager.workArea.search.workareaName'),
      id: 'workareaName',
      multiple: true,
      children: workareaStore.workareaList.map(item => ({
        id: item.bk_networkarea_name,
        name: item.bk_networkarea_name,
      })),
    },
    {
      name: t('topoManager.workArea.search.workareaId'),
      id: 'workareaId',
      multiple: true,
      children: workareaStore.workareaList.map(item => ({
        id: String(item.bk_networkarea_id),
        name: String(item.bk_networkarea_id),
      })),
    },
    {
      name: t('topoManager.workArea.search.vendor'),
      id: 'vendor',
      multiple: true,
      children: workareaStore.vendorList.map(item => ({
        id: String(item),
        name: t(String(vendorMap[item]?.label || '')),
      })),
    },
  ];
};

// 表头过滤
const handleFilter = async ({ checked, field }: { checked: string[]; field: string }) => {
  // 管控单元数量为前端统计值，筛选前先补齐统计数据
  if (field === 'networkunit_count') {
    workareaStore.frontPageConf.current = 1;
    if (checked.length > 0) {
      await workareaStore.handleFetchAllWorkareaStatistics();
    }
    unitCountFilterChecked.value = checked;
    return;
  }

  const index = searchKey.value.findIndex((item: any) => item.id === field);
  if (index > -1) searchKey.value.splice(index, 1);
  if (checked.length) {
    searchKey.value.push({
      id: field === 'cloud_vendor' ? 'vendor' : field,
      name: field === 'cloud_vendor' ? t('topoManager.workArea.search.vendor') : field,
      values: checked.map((item: any) => {
        let name = item;
        let value_id = item;
        if (field === 'cloud_vendor') {
          value_id = workareaStore.vendorList[item];
          const label = vendorMap[workareaStore.vendorList[item]]?.label;
          name = label ? t(String(label)) : item;
        }
        return { id: value_id, name };
      }),
    });
  }
};

// 后端端过滤数据
watch(searchKey, async (newVal) => {
  workareaStore.includeConditions.bk_networkarea_name = [];
  workareaStore.includeConditions.bk_networkarea_id = [];
  workareaStore.includeConditions.cloud_vendor = [];
  newVal.forEach((item: any) => {
    const keyMap: Record<string, keyof typeof workareaStore.includeConditions> = {
      workareaName: 'bk_networkarea_name',
      workareaId: 'bk_networkarea_id',
      vendor: 'cloud_vendor',
    };
    const field = keyMap[String(item.id)];
    if (!field) return;
    const values = item.values.map((value: any) => value.id);
    if (field === 'bk_networkarea_id') {
      workareaStore.includeConditions[field] = values.map(Number);
      return;
    }
    workareaStore.includeConditions[field] = values;
  });
  await workareaStore.handleFetchWorkareaList();
  tableData.value = workareaStore.workareaList;
}, { deep: true });
const getTableData = async () => {
  await Promise.all([
    workareaStore.handleGetAllWorkareaList(),
    workareaStore.handleFetchVendorAndOs(),
  ]);
  tableData.value = workareaStore.workareaList;
  initSearchData();
};
// 对于NetworkAreaStatistics异步请求变化的数据监听
watch(() => workareaStore.workareaList, () =>  {
  tableData.value = workareaStore.workareaList;
}, { deep: true });
onMounted(async () => {
  // 确保 topoManager 模块权限数据已加载（statistics 接口依赖 authorized 判断有权限的区域）
  const topoItems = getModuleAuthorizedItems('topoManager');
  await authStore.fetchAuthorized(topoItems, 'topoManager').catch(() => {});
  await getTableData();
});

</script>

<style lang="postcss">
.bk-dropdown-popover .bk-dropdown-item.batch-item-disabled,
.bk-dropdown-popover .bk-dropdown-item.batch-item-disabled:hover {
  color: #c4c6cc;
  cursor: not-allowed;
  background-color: #fff;
}
</style>
