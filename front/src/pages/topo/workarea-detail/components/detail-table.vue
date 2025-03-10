<template>
  <div>
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
      @setting-change="handleSettingChange"
      @column-filter="handleColumnFilter">
      <TableColumn type="checkbox" :width="60" :resizable="false" />
      <TableColumn
        :label="$t('topoManager.workAreaDetail.table.ipv4')"
        field="ipv4"
        show-overflow="tooltip"
        :min-width="260">
        <template #default="{ row }">
          <Button theme="primary" class="!text-[12px]" text>{{ row.ipv4 }}</Button>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workAreaDetail.table.ipv6')"
        field="ipv6"
        show-overflow="tooltip"
        :min-width="260">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.ipv6 ? `#${row.ipv6}` : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        label="Agent ID"
        field="agentId"
        show-overflow="tooltip"
        :min-width="260">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.agentId || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workAreaDetail.table.proxyVersion')"
        field="proxyVersion"
        show-overflow="tooltip"
        :filter="proxyVersionFilter"
        :min-width="240">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.proxyVersion || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workAreaDetail.table.proxyStatus')"
        field="proxyStatus"
        show-overflow="tooltip"
        :filter="proxyStatusFilter"
        :min-width="130">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.proxyStatus || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workAreaDetail.table.action')"
        field="action"
        show-overflow="tooltip"
        :min-width="140">
        <template #default="{ row }">
          <div class="flex">
            <Button theme="primary" text class="mr-[12px]">
              {{ $t('topoManager.workAreaDetail.table.Reassembly') }}
            </Button>
            <Button text>
              <i class="nodeman-icon nc-more"></i>
            </Button>
          </div>
        </template>
      </TableColumn>
    </Table>
  </div>
</template>

<script lang="ts" setup>
import { Button } from 'bkui-vue';
import { reactive, ref } from 'vue';

import { Table, TableColumn } from '@blueking/table';

import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';

const list = ref([
  {
    ipv4: '3.45.345.23',
    ipv6: '3.45.345.23',
    agentId: '0201312d2f525400ee6b5b17367675080054',
    proxyVersion: '1.2.3',
    proxyStatus: 'success',
  },
  {
    ipv4: '11.345.117.43',
    ipv6: '11.345.117.43',
    agentId: '0201312d2f525400ee6b5b17367675080054',
    proxyVersion: '1.2.3',
    proxyStatus: 'success',
  },
  {
    ipv4: '54.34.64.854',
    ipv6: '54.34.64.854',
    agentId: '0201312d2f525400ee6b5b17367675080054',
    proxyVersion: '1.2.3',
    proxyStatus: 'success',
  },
  {
    ipv4: '1.1.1.1',
    ipv6: '1.1.1.1',
    agentId: '0201312d2f525400ee6b5b17367675080054',
    proxyVersion: '1.2.3',
    proxyStatus: 'success',
  },
  {
    ipv4: '4.34.34.32',
    ipv6: '4.34.34.32',
    agentId: '0201312d2f525400ee6b5b17367675080054',
    proxyVersion: '1.2.3',
    proxyStatus: 'success',
  },
]);
const pagination = reactive({ count: 0, limit: 20, current: 1 });
const sortConfig = ref({ multiple: true });

const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'ipv4',
    'ipv6',
    'agentId',
    'proxyVersion',
    'proxyStatus',
    'action',
  ],
  disabled: ['action'],
});

const tableRef = ref();

// 待优化 各影响table最大高度的元素的高度
const tableOffset = 445;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleColumnFilter = () => {

};

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
</script>
