<template>
  <div class="p-[24px]">
    <FlexRow>
      <template #left></template>
      <template #right>
        <div class="flex items-center">
          <DatePicker
            type="daterange"
            v-model="operateTime"
            :placeholder="$t('topoManager.record.searchPlaceholder.date')"
            class="mr-[8px]"
          >
          </DatePicker>
          <SearchSelect
            class="w-[480px]"
            unique-select
            v-model.trim="searchSelectValue"
            :data="searchSelectData">
          </SearchSelect>
        </div>
      </template>
    </FlexRow>
    <Loading :loading="loading">
      <Table
        class="mt-[16px] w-full"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @setting-change="handleSettingChange">
        <TableColumn
          :label="'配置名称'"
          field="configpolicy_name"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.bk_networkarea_name || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="'配置ID'"
          field="configpolicy_id">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operator')"
          field="operator"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.operator || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operateTime')"
          field="operate_time"
          show-overflow="tooltip"
          sortable>
          <template #default="{ row }">
            {{ filterTimeFormat(row.operate_time) || '--' }}
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>

<script lang="ts" setup>
import { DatePicker, Loading, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import type { TopoEventExactConditions, TopoEventFuzzyConditions } from '@/@types/topo';
import { filterTimeFormat, getTimeStamp } from '@/common/util';
import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';

const { t } = useI18n();
const searchSelectValue = ref<ISearchValue[]>([]);
const loading = ref(false);

const searchSelectData = ref<ISearchItem[]>([]);

const pagination = reactive({ count: 0, limit: 50, current: 1 });
const sortConfig = ref({ multiple: true });

const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'configpolicy_name',
    'configpolicy_id',
    'operator',
    'operate_time',
  ],
});

const list = ref<TopoEvent[]>([]);
const operateTime = ref([]);

const operateTimeRange = computed(() => {
  if (operateTime.value.length !== 0 && operateTime.value.every(item => item !== '')) {
    const operate_time_range: TimeRange = {
      start_timestamp_sec: getTimeStamp(operateTime.value[0]),
      end_timestamp_sec: getTimeStamp(operateTime.value[1]),
    };
    return operate_time_range;
  }
  return null as unknown as TimeRange;
});

const fetchRecordList = async () => {};

const initSearchList = async () => {
};

onMounted(() => {
  fetchRecordList();
  initSearchList();
});

</script>
