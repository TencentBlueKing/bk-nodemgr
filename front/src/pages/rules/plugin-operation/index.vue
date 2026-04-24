<template>
  <div class="p-[24px]">
    <div class="flex items-center justify-between">
      <div>
        <Button theme="primary" @click="handleCreate">
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          {{ $t('pluginOperation.createStrategy') }}
        </Button>
      </div>
      <div>
        <SearchSelect
          class="w-[480px] z-99 bg-[#fff]"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('common.searchPlaceholder')"
        />
      </div>
    </div>
    <Loading
      :title="$t('agentStrategy.loading')"
      :loading="loading"
      class="mt-[16px] flex-1 overflow-auto"
    >
      <Table
        class="w-full"
        :max-height="maxHeight"
        :data="tableData"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        show-overflow-tooltip
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn
          field="strategy_name"
          :title="$t('pluginOperation.table.strategyName')"
          :min-width="150"
          fixed="left"
          show-overflow="tooltip"
        />
        <TableColumn
          field="plugin_name"
          :title="$t('pluginOperation.table.pluginName')"
          :min-width="120"
        />
        <TableColumn
          field="operation_type"
          :title="$t('pluginOperation.table.operationType')"
          :min-width="100"
        />
        <TableColumn
          field="target_version"
          :title="$t('pluginOperation.table.targetVersion')"
          :min-width="120"
        />
        <TableColumn
          field="creator"
          :title="$t('pluginOperation.table.creator')"
          :min-width="120"
        />
        <TableColumn
          field="create_time"
          :title="$t('pluginOperation.table.createTime')"
          :min-width="180"
        />
        <TableColumn
          field="status"
          :title="$t('pluginOperation.table.status')"
          :min-width="100"
        >
          <template #default="{ row }">
            <Tag v-if="row.enabled" theme="success">{{ $t('pluginOperation.table.enabled') }}</Tag>
            <Tag v-else>{{ $t('pluginOperation.table.disabled') }}</Tag>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="$t('pluginOperation.table.action')"
          :min-width="120"
          fixed="right"
        >
          <template #default="{ row }">
            <div class="flex items-center">
              <Button class="mr-[8px]" theme="primary" text @click="handleEdit(row)">
                {{ $t('pluginOperation.table.edit') }}
              </Button>
              <Button theme="primary" text @click="handleDelete(row)">
                {{ $t('pluginOperation.table.delete') }}
              </Button>
            </div>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>

<script lang="ts" setup>
import { Button, Loading, SearchSelect, Tag } from 'bkui-vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 255 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<any[]>([]);
const loading = ref(false);
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  {
    id: 'strategy_name',
    name: t('pluginOperation.table.strategyName'),
    children: [],
  },
  {
    id: 'plugin_name',
    name: t('pluginOperation.table.pluginName'),
    children: [],
  },
]);

const pageLimitChange = (limit: number) => {
  pagination.limit = limit;
};
const pageValueChange = (current: number) => {
  pagination.current = current;
};

const handleCreate = () => {
  router.push({ name: 'createPluginOperation' });
};

const handleEdit = (row: any) => {
  console.log('edit', row);
};

const handleDelete = (row: any) => {
  console.log('delete', row);
};

onMounted(() => {
  // 模拟数据，后续对接接口
  tableData.value = [];
  pagination.count = 0;
});
</script>
