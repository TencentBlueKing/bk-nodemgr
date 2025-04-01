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
          <!-- 复制 -->
          <CopyIp :get-select-data="handleGetSelectData"></CopyIp>
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
      :list="tableData"
      :vendor-list="workareaStore.vendorList"
      @edit="handleEditWorkarea"
    />
    <UpsertWorkarea
      v-model:is-show="showUpsertWorkarea"
      :is-create="isCreate"
      @install-proxy="handleInstallProxy"
    />
    <InstallProxy v-bind:is-show="isInstallProxyShow" />
  </div>
</template>

<script setup lang="ts">
import { Button, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import InstallProxy from '../install-proxy/install-proxy.vue';

import RegionTable from './components/region-table.vue';
import UpsertWorkarea from './components/upsert-workarea.vue';
import { vendorMap } from './vendorMap';

import CopyIp from '@/components/copy-ip.vue';
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

// 复制
const regionTableRef = ref();
const handleGetSelectData = async (type: string) => {
  let tableData = [];
  if (type === 'select') {
    // 获取勾选行
    tableData = regionTableRef.value.tableRef.getVxeTableInstance().getCheckboxRecords();
  } else {
    // 获取全部行
    tableData = await workareaStore.handleGetAllWorkareaList();
  }
  return tableData;
};

// 表格数据
const tableData = ref<INetWorkArea[]>([]);

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
        id: String(id),
        name: t(String(vendorMap[item]?.label || '')),
      })),
    },
  ];
};
// 前端过滤数据
watch(searchKey, (newVal) => {
  let data: INetWorkArea[] = workareaStore.workareaList;
  for (const item of newVal) {
    switch (item.id) {
      case 'workareaName':
        data = data.filter(area => item.values?.find(o => o.id === area.bk_networkarea_name));
        break;
      case 'workareaId':
        data = data.filter(area => item.values?.find(o => o.id === String(area.bk_networkarea_id)));
        break;
      case 'vendor':
        data = data.filter(area => item.values?.find(o => o.id === area.cloud_vendor));
        break;
    }
  }
  tableData.value = data;
});

onMounted(async () => {
  await Promise.all([
    workareaStore.handleGetAllWorkareaList(),
    workareaStore.handleFetchVendorAndOs(),
  ]);
  tableData.value = workareaStore.workareaList;
  initSearchData();
});

</script>
