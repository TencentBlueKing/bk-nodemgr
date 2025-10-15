<template>
  <div class="p-[24px]">
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[12px]">
        <DatePicker
          v-model="dateValue"
          :shortcut-selected-index="1"
          :shortcuts="shortcutsRange"
          format="yyyy-MM-dd HH:mm:ss"
          type="datetimerange"
          use-shortcut-text
          @change="pickSuccess" />
      </div>
      <div class="flex-1 ml-[8px]">
        <SearchSelect
          ref="searchSelect"
          :data="searchSelectData"
          v-model="searchSelectValue"
          :unique-select="true"
          :placeholder="t('请输入 包名称、类别、版本、操作、操作人 搜索')"
          @update:model-value="handleSearchSelectChange">
        </SearchSelect>
      </div>
    </section>
    <bk-loading title="数据加载中" :loading="loading">
      <Table
        :data="tableData"
        :empty-text="'暂无数据'"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        show-overflow-tooltip
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
      >
        <TableColumn field="workflow_id" :title="t('类别')" min-width="100" fixed="left"></TableColumn>
        <TableColumn field="type" :title="t('包名称')" min-width="150"></TableColumn>
        <TableColumn field="bk_biz_name" :title="t('版本')" min-width="150"></TableColumn>
        <TableColumn field="operator" :title="t('操作人')" min-width="150"></TableColumn>
        <TableColumn field="operate_time" :title="t('操作时间')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ timeFormatter(row.operate_time) }}</span>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import { Button, Cascader, Checkbox, DatePicker, Dropdown, InfoBox, SearchSelect } from 'bkui-vue';
import dayjs from 'dayjs';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { NodeWorkflowInfo } from '@/@types/node_workflow';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PackageService } from '@/api/modules/pkg';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const tableData = ref<NodeWorkflowInfo[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
// 分页
const {
  pagination,
} = usePage(tableData);
// 跨页全选
const loading = ref(false);

// 日期选择
const dateValue = ref([new Date().setTime(new Date().getTime() - 3600 * 1000 * 24 * 7), new Date()]);
const shortcutsRange = reactive([
  {
    text: '今天',
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: '近7天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: '近15天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: '近30天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  },
]);
const pickSuccess = async (val: string[]) => {
  await getTaskList();
};

const timeFormatter = (val: string, format = 'YYYY-MM-DD HH:mm:ss') => (val ? dayjs(val).format(format) : '--');

const formatTimeToMS = (duration: number) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = Math.floor((duration % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
};

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'workflow_id',
    'bk_biz_name',
    'type',
    'bk_policy_name',
    'operator',
    'operate_time',
    'cost_time',
    'status',
    'count',
  ],
  disabled: [],
});
const getUniqueChildren = (prop: string, map?: Record<string, any>) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    id: value,
    name: map && map[value as string] ? map[value as string].text : value,
  }));
};
const searchSelectData = computed(() => [
  // {id: 'workflow_id', name: t('platform.nodeMan.taskHistory.label.taskID')},
  // {id: 'type', name: 'platform.nodeMan.taskHistory.label.taskType', children: getUniqueChildren('type', typeMap)},
  // {id: 'bk_biz_id', name: 'platform.nodeMan.taskHistory.label.business', children: bussinessMap.value, multiple: true},
  // {id: 'operator', name: 'platform.nodeMan.taskHistory.label.operator', children: getUniqueChildren('operator')},
  // {id: 'status', name: 'platform.nodeMan.taskHistory.label.status', children: getUniqueChildren('status')},
]);

// 搜索
const searchSelectValue = ref<{id: string, name: string, values: any}[]>([]);
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string, name: string}[]}[]) => {

};

const getTimestampInSeconds = (originalDate: number | Date) => {
  const timestampInMilliseconds = new Date(originalDate).getTime();
  const timestampInSeconds = Math.floor(timestampInMilliseconds / 1000);
  return timestampInSeconds;
};
const getParams = () => {
  const params = {
    page: {
      limit: 0,
      offset: 0,
    },
    exact_include_conditions: {
      bk_biz_id: mainStore.selectedBusinessId,
    },
    fuzzy_include_conditions: {} as Record<string, string[]>,
    operate_time_range: {
      start_timestamp_sec: getTimestampInSeconds(dateValue.value[0]),
      end_timestamp_sec: getTimestampInSeconds(dateValue.value[1]),
    },
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getTaskList = async () => {
  loading.value = true;
  const res = await NodeWorkflowService.NodeWorkflowList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  tableData.value = res.items;
  loading.value = false;
};

watch(() => searchSelectValue, async () => {
  await getTaskList();
}, { deep: true });
onMounted(async () => {
  await getTaskList();
});
</script>
