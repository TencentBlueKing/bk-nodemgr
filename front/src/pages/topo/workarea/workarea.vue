<template>
  <div class="p-[24px]">
    <FlexRow>
      <template #left>
        <div class="flex items-center">
          <!-- 新建 -->
          <Button theme="primary" class="mr-[8px]" @click="showUpsertWorkarea = true">
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
    <RegionTable ref="regionTableRef"></RegionTable>
    <UpsertWorkarea v-model:is-show="showUpsertWorkarea"></UpsertWorkarea>
  </div>
</template>

<script setup lang="ts">
import { Button, SearchSelect } from 'bkui-vue';
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { useDebounce } from '@vueuse/core';

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

const regionTableRef = ref();
const handleGetSelectData = async (type: string) => {
  let tableData = [];
  if (type === 'select') {
    // 获取勾选行
    tableData = regionTableRef.value.tableRef.getVxeTableInstance().getCheckboxRecords();
  } else {
    // 获取全部行
    tableData = await workareaStore.handleFetchAllWorkAreaList();
  }
  return tableData;
};

watch(debounceSearch, (newVal) => {
  console.log(newVal)
  for (const o of newVal) {
    const curValue = o.values[0].name;
    switch (o.id) {
      case 'workareaName':
        workareaStore.includeConditions.bk_networkarea_name = [curValue];
        break;
      case 'workareaId':
        workareaStore.includeConditions.bk_networkarea_name = [parseInt(curValue)];
        break;
      case 'vendor':
        workareaStore.includeConditions.bk_networkarea_name = [parseInt(curValue)];
        break;
    }
  }
});

</script>
