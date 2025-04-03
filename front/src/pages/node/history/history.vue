<template>
  <div class="p-[24px]">
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[12px]">
        <Checkbox v-model="hideAutoTask">隐藏自动部署任务</Checkbox>
        <DatePicker
          v-model="dateValue"
          :shortcut-selected-index="1"
          :shortcuts="shortcutsRange"
          format="yyyy-MM-dd HH:mm:ss"
          type="datetimerange"
          use-shortcut-text
          @change="dateChange" />
      </div>
      <div class="flex-1 ml-[8px]">
        <SearchSelect ref="searchSelect" :data="searchSelectData" v-model="searchSelectValue" :uniqueSelect="true"
          :placeholder="'搜索 任务ID、执行人、任务类型、操作类型、部署策略、执行状态 搜索'" @update:modelValue="handleSearchSelectChange">
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
      >
        <TableColumn field="bk_task_id" :title="t('任务ID')" width="150" fixed="left">
          <template #default="{ row }">
            <Button text theme="primary" @click="handleTaskDetail(row.bk_task_id)">{{ row.bk_task_id }}</Button>
          </template>
        </TableColumn>
        <TableColumn field="bk_operate_type" :title="t('操作类型')" :filter="operateFilterOption"></TableColumn>
        <TableColumn field="bk_task_type" :title="t('任务类型')" :filter="taskFilterOption"></TableColumn>
        <TableColumn field="bk_bussiness" :title="t('业务')" width="150"></TableColumn>
        <TableColumn field="bk_policy_name" :title="t('部署策略')"></TableColumn>
        <TableColumn field="created_by" :title="t('执行人')" :filter="createdFilterOption"></TableColumn>
        <TableColumn field="start_time" :title="t('执行时间')">
          <template #default={row}>
            <span>{{ timeFormatter(row.start_time) }}</span>
          </template>
        </TableColumn>
        <TableColumn field="cost_time" :title="t('总耗时')"></TableColumn>
        <TableColumn field="task_status" :title="t('执行状态')" width="150" :filter="statusFilterOption">
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <span :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"></span>
              <span>{{ row.node_status }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.node_status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn field="count" :title="t('总数/成功/失败/忽略')"></TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { useI18n } from 'vue-i18n';
import { toLower } from 'lodash';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Button, Dropdown, Cascader, SearchSelect, DatePicker, Checkbox } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { WorkflowService } from '@/api/modules/workflow';
import { TopoService } from '@/api/modules/topo';
import useTableSetting from '@/composables/use-table-setting';

import dayjs from 'dayjs';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const tableData = ref<Host[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
// 分页
const {
  pagination
} = usePage(tableData);
// 跨页全选
const loading = ref(false);

// 日期选择
const dateValue = reactive([new Date(), new Date()]);
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
const dateChange = (val: string[]) => {
  console.log(val);
}

// 隐藏自动部署任务
const hideAutoTask = ref(false);

// 搜索
const searchSelectValue = ref([]);
const handleSearchSelectChange = ({id, name, values}: {id: number, name: string, values: []}) => {}

const timeFormatter = (val: string, format = 'YYYY-MM-DD HH:mm:ss') => {
  return val ? dayjs(val).format(format) : '--';
}

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_task_id',
    'bk_bussiness',
    'bk_operate_type',
    'bk_task_type',
    'bk_policy_name',
    'created_by',
    'start_time',
    'cost_time',
    'task_status',
    'count'
  ],
  disabled: [],
});
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    id: value,
    name: value
  }))
}
const searchSelectData = computed(() => [
  {id: 'bk_task_id', name: '任务ID'},
  {id: 'bk_task_type', name: '任务类型', children: getUniqueChildren('bk_task_type')},
  {id: 'bk_operate_type', name: '操作类型', children: getUniqueChildren('bk_operate_type')},
  {id: 'created_by', name: '执行者', children: getUniqueChildren('created_by')},
  {id: 'task_status', name: '执行状态', children: getUniqueChildren('task_status')},
]);
const filterOptionConfig = (prop: string) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return {
    list: uniqueValues.map(value => ({
      text: value,
      value: value
    })),
    checked: [] as string[],
    filterScope: 'all',
  }
}
const taskFilterOption = computed(() => filterOptionConfig('bk_task_type'));
const operateFilterOption = computed(() => filterOptionConfig('bk_operate_type'));
const createdFilterOption = computed(() => filterOptionConfig('created_by'));
const statusFilterOption = computed(() => filterOptionConfig('task_status'));

// 跳转任务详情
const handleTaskDetail = (bk_task_id: number | string) => {
  router.push({
    name: 'taskDetail',
    params: {
      taskId: bk_task_id,
    }
  })
}

const getTaskList = async () => {
  loading.value = true;
  const res = await WorkflowService.WorkflowList({
    page: {
      limit: 0
    },
  }).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    }
  });
  tableData.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
    bk_task_id: item.bk_host_id
  }));
  loading.value = false;
}
onMounted(async () => {
  await getTaskList();
});
</script>
<style scoped lang="postcss">
:deep(.vxe-table--empty-content) {
  height: 200px;
  line-height: 200px;
}

.status-icon::before {
  content: '';
  display: inline-block;
  margin-right: 8px;
  width: 13px;
  height: 13px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
}
.nc-running {
  &::before {
    background: #3fc06d;
    border-color: #e5f6ea;
  }
}
.nc-terminated {
  &::before {
    border-color: #ffe6e6;
    background: #ea3636;
  }
}
.nc-unknown {
  &::before {
    border-color: #f0f1f5;
    background: #b2b5bd;
  }
}

</style>