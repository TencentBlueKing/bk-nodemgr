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
        <TableColumn field="workflow_id" :title="t('任务ID')" width="300" fixed="left">
          <template #default="{ row }">
            <Button text theme="primary" @click="detailHandle(row, row.status)">{{ row.workflow_id }}</Button>
          </template>
        </TableColumn>
        <TableColumn field="type" :title="t('任务类型')" :filter="taskFilterOption" width="150"></TableColumn>
        <TableColumn field="bk_biz_name" :title="t('业务')" width="200"></TableColumn>
        <TableColumn field="operator" :title="t('执行人')" :filter="createdFilterOption"></TableColumn>
        <TableColumn field="operate_time" :title="t('执行时间')">
          <template #default={row}>
            <span>{{ timeFormatter(row.operate_time) }}</span>
          </template>
        </TableColumn>
        <TableColumn field="cost_time" :title="t('总耗时')">
          <template #default={row}>
            <span>{{ formatTimeToMS(row.cost_time) }}</span>
          </template>
        </TableColumn>
        <TableColumn field="status" :title="t('执行状态')" width="150" :filter="statusFilterOption">
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.status">
              <Spinner v-if="row.status === 'running'" class="mr-[8px]"/>
              <template v-else>
                <i :class="`nodeman-icon nc-${statusMap[row.status].icon} status-icon`"></i>
              </template>
              <span>{{ statusMap[row.status].text }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn field="count" :title="t('总数/成功/失败/忽略')">
          <template #default="{ row }">
            <template v-if="row.statistics">
              <span class="pr-[4px]">{{ row.statistics.totalCount || 0 }}</span>/
              <a class="text-[#2dcb56] pr-[4px]"
                @click.stop="detailHandle(row, 'success')">{{ row.statistics.successCount || 0 }}</a>/
              <a class="text-[#ea3636] pr-[4px]"
                @click.stop="detailHandle(row, 'failed')">{{ row.statistics.failedCount || 0 }}</a>/
              <a class="text-[#ff9c01] pr-[4px]"
                @click.stop="detailHandle(row, 'ignored')">{{ row.statistics.ignoredCount || 0 }}</a>
            </template>
            <span v-else>--</span>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Spinner } from 'bkui-vue/lib/icon';
import { Table, TableColumn } from '@blueking/table';
import { useI18n } from 'vue-i18n';
import { toLower } from 'lodash';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Button, Dropdown, Cascader, SearchSelect, DatePicker, Checkbox } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import useTableSetting from '@/composables/use-table-setting';
import dayjs from 'dayjs';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const tableData = ref<NodeWorkflowInfo[]>([]);
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
const statusMap = {
  running: {
    text: t('执行中')
  },
  failed: {
    text: t('失败'),
    icon: 'terminated'
  },
  success: {
    text: t('成功'),
    icon: 'running'
  },
  partial_failed: {
    text: t('部分失败'),
    icon: 'warning'
  },

}
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

const formatTimeToMS = (duration: number) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = Math.floor((duration % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

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
  {id: 'workflow_id', name: '任务ID'},
  {id: 'type', name: '任务类型', children: getUniqueChildren('type')},
  {id: 'operator', name: '执行者', children: getUniqueChildren('created_by')},
  {id: 'status', name: '执行状态', children: getUniqueChildren('task_status')},
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
const taskFilterOption = computed(() => filterOptionConfig('type'));
const createdFilterOption = computed(() => filterOptionConfig('created_by'));
const statusFilterOption = computed(() => filterOptionConfig('task_status'));

const getTaskList = async () => {
  loading.value = true;
  const res = await NodeWorkflowService.NodeWorkflowList({
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
  const statistics = await NodeWorkflowService.NodeWorkflowStatistics({
    workflow_id: res.items.map(item => item.workflow_id)
  }).catch((err) => {
    console.log(err);
    return {
      items: [],
    }
  });
  tableData.value = res.items.map(item => {
    const statisticsItem = statistics.items.find(statistic => statistic.workflow_id === item.workflow_id);
    return {
      statistics: statisticsItem,
      ...item,
      bk_biz_name: item.bk_biz_name.filter(item => item),
      cost_time: item.finish_time > 0 ? (item.finish_time - item.operate_time) : 0
    }
  });
  console.log(tableData.value)
  loading.value = false;
}

// 跳转详情
const detailHandle = (row: NodeWorkflowInfo, status: string) => {
  nodeManageStore.updateCurrentRowData(row);
  nodeManageStore.updateCurrentStatus(status);
  router.push({
    name: 'taskDetail',
    params: {
      taskId: row.workflow_id,
    },
  });
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
  width: 8px;
  height: 8px;
  border: 1px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
}
.nc-running {
  &::before {
    background: #CBF0DA;
    border-color: #2CAF5E;
  }
}
.nc-terminated {
  &::before {
    border-color: #EA3636;
    background: #FFDDDD;
  }
}
.nc-warning {
  &::before {
    border-color: #F59500;
    background: #FCE5C0;
  }
}
.nc-unknown {
  &::before {
    border-color: #b2b5bd;
    background: #f0f1f5;
  }
}

</style>