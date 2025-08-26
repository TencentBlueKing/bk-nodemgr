<template>
  <div>
    <bk-loading
      title="数据加载中"
      :loading="loading"
      class="w-full overflow-auto"
    >
      <Table
        class="mt-[16px] w-full"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        empty-cell-text="--"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @setting-change="handleSettingChange"
        @column-filter="handleColumnFilter">
        <TableColumn type="checkbox" :width="60" :resizable="false" />
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv4')"
          field="bk_host_innerip"
          show-overflow="tooltip"
          :min-width="150">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv6')"
          field="bk_host_innerip_v6"
          show-overflow="tooltip"
          :min-width="150">
          <template #default="{ row }">
            <span class="!text-[12px]">
              {{ row.bk_host_innerip_v6 ? `#${row.bk_host_innerip_v6}` : '--' }}
            </span>
          </template>
        </TableColumn>
        <TableColumn
          label="Agent ID"
          field="bk_agent_id"
          show-overflow="tooltip"
          :min-width="300">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.proxyVersion')"
          field="node_version"
          show-overflow="tooltip"
          :filter="proxyVersionFilter"
          :min-width="120">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.proxyStatus')"
          field="node_status"
          show-overflow="tooltip"
          :filter="proxyStatusFilter"
          :min-width="130">
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <i
                :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"
              ></i>
              <div>{{ row.node_status }}</div>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.node_status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.action')"
          field="action"
          fixed="right"
          show-overflow="tooltip"
          :min-width="140">
          <template #default="{ row }">
            <div class="flex">
              <Button theme="primary" text class="mr-[12px]">
                {{ $t('topoManager.workAreaDetail.table.Reassembly') }}
              </Button>
              <MoreAction :ipv4="row.bk_host_innerip" :row="row"></MoreAction>
            </div>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>

<script lang="ts" setup>
import { Button } from 'bkui-vue';
import { reactive, ref, onMounted, computed, watch } from 'vue';

import { Table, TableColumn } from '@blueking/table';

import MoreAction from './more-action.vue';

import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from "@/@types/topo.d";
import { TopoService } from "@/api/modules/topo";
import { useRoute } from 'vue-router';

const props = defineProps({
  searchSelectValue: {
    type: Array,
    default: []
  },
  bkNetworkunitId: {
    type: Number,
    default: 0
  }
});
const route = useRoute();
const workAreaId = Number(route.params.workarea);
const list = ref<Host[]>([]);
const pagination = reactive({ count: 0, limit: 20, current: 1 });
const sortConfig = ref({ multiple: true });

const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_agent_id',
    'node_version',
    'node_status',
    'action',
  ],
  disabled: ['action'],
});

const tableRef = ref();
const loading = ref(false);
// 待优化 各影响table最大高度的元素的高度
const tableOffset = 445;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleColumnFilter = () => {

};
const searchSelectValue = computed(() => props.searchSelectValue);
const proxyVersionFilter = reactive({
  list: [],
  checked: [],
});
const proxyStatusFilter = reactive({
  list: [
    {
      text: 'success',
      value: 1,
    },
    {
      text: 'failed',
      value: 0,
    },
  ],
  checked: [],
});
const fuzzyKeys = new Set([
  "bk_host_innerip",
  "bk_host_innerip_v6",
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1)*pagination.limit,
    },
    exact_include_conditions: {} as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  params.exact_include_conditions.node_role = ["proxy"];
  params.exact_include_conditions.bk_networkunit_id = [props.bkNetworkunitId];
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values.map((value: any) => value.id);
  });
  return params;
};
const getAgentList = async () => {
  loading.value = true;
  const res = await TopoService.HostList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;
  list.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
  }));
  loading.value = false;
};
watch(searchSelectValue,async () => {
  await getAgentList();
},{immediate: true, deep: true});
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 13px;
  height: 13px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
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