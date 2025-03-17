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
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { useDebounce } from '@vueuse/core';

import InstallProxy from '../install-proxy/install-proxy.vue';

import RegionTable from './components/region-table.vue';
import UpsertWorkarea from './components/upsert-workarea.vue';

import CopyIp from '@/components/copy-ip.vue';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const workareaStore = useWorkareaStore();

const showUpsertWorkarea = ref(false);
// 搜索
const searchKey = ref([]);
const debounceSearch = useDebounce(searchKey, 300);
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

const searchSelectData = ref([
  {
    name: t('topoManager.workArea.search.workareaName'),
    id: 'workareaName',
  },
  {
    name: t('topoManager.workArea.search.workareaId'),
    id: 'workareaId',
  },
  {
    name: t('topoManager.workArea.search.vendor'),
    id: 'vendor',
    multiple: true,
  },
]);
const isInstallProxyShow = ref(false);
const handleInstallProxy = () => {
  isInstallProxyShow.value = true;
};

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

watch(debounceSearch, (newVal) => {
  // todo
});

</script>
