<template>
  <div class="p-[24px]">
    <FlexRow>
      <template #left>
        <div class="flex items-center">
          <!-- 新建 -->
          <Button theme="primary" class="mr-[8px]" @click="handleCreateWorkarea">
            <i class="nodeman-icon nc-plus-line mr-[4.5px]"></i>
            <span>{{ $t('action.create') }}</span>
          </Button>
        </div>
      </template>
      <template #right>
        <SearchSelect
          class="w-[480px]"
          unique-select
          :placeholder="$t('topoManager.workArea.search.placeholder')"
          v-model.trim="searchKey"
          :data="searchSelectData">
        </SearchSelect>
      </template>
    </FlexRow>
    <RegionTable
      ref="regionTableRef"
      :list="sortTableData"
      :vendor-list="workareaStore.vendorList"
      @edit="handleEditWorkarea"
    />
    <UpsertWorkarea
      v-model:is-show="showUpsertWorkarea"
      :is-create="isCreate"
      :cur-workarea-data="curWorkareaData"
      @install-proxy="handleInstallProxy"
      @update="handleUpdate"
    />
    <InstallProxy v-bind:is-show="isInstallProxyShow" />
  </div>
</template>

<script setup lang="ts">
import { Button, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import InstallProxy from '../install-proxy/install-proxy.vue';

import RegionTable from './components/region-table.vue';
import UpsertWorkarea from './components/upsert-workarea.vue';
import { vendorMap } from './vendorMap';

import type {
  TopoNetworkAreaListReq,
  TopoNetworkAreaListReqExactConditions,
  TopoNetworkAreaListReqFuzzyConditions,
} from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';
import type { INetWorkArea } from '@/stores/workarea';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const workareaStore = useWorkareaStore();

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
const sortTableData = computed(() => tableData.value.sort((a: INetWorkArea, b: INetWorkArea) => {
  if (a.bk_networkarea_id === 0) {
    return -1;
  } if (b.bk_networkarea_id === 0) {
    return 1;
  }
  return b.bk_networkarea_id - a.bk_networkarea_id;
}));

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
      children: workareaStore.vendorList.map((item, id) => ({
        id: String(item),
        name: t(String(vendorMap[item]?.label || '')),
      })),
    },
  ];
};
// 后端端过滤数据
watch(searchKey, async (newVal) => {
  workareaStore.includeConditions.bk_networkarea_name = [];
  workareaStore.includeConditions.bk_networkarea_id = [];
  workareaStore.includeConditions.cloud_vendor = [];
  newVal.forEach((item: any) => {
    const keyMap = {
      workareaName: 'bk_networkarea_name',
      workareaId: 'bk_networkarea_id',
      vendor: 'cloud_vendor',
    };
    if (keyMap[item.id]) {
      workareaStore.includeConditions[keyMap[item.id]] = item.values.map((value: any) => (item.id === 'workareaId' ? Number(value.id) : value.id));
    }
  });
  await workareaStore.handleFetchWorkareaList();
  tableData.value = workareaStore.workareaList;
});
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
  await getTableData();
});

</script>
