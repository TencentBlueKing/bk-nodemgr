<template>
  <div class="p-[24px]">
    <div class="flex items-center justify-end">
      <SearchSelect
        class="w-[480px] z-99 bg-[#fff]"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="$t('deployPolicy.searchPlaceholder')"
        @update:model-value="handleSearchSelectChange"
      />
    </div>
    <Loading
      :title="$t('deployPolicy.loading')"
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
          field="deploy_policy_id"
          :title="$t('deployPolicy.table.deployPolicyId')"
          :min-width="120"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="dsu_id"
          :title="$t('deployPolicy.table.dsuId')"
          :min-width="100"
        ></TableColumn>
        <TableColumn
          field="name"
          :title="$t('deployPolicy.table.name')"
          :min-width="180"
          fixed="left"
          show-overflow="tooltip"
        >
          <template #default="{ row }">
            {{ row.meta?.name }}
          </template>
        </TableColumn>
        <TableColumn
          field="description"
          :title="$t('deployPolicy.table.description')"
          :min-width="220"
          show-overflow="tooltip"
        >
          <template #default="{ row }">
            {{ row.meta?.description }}
          </template>
        </TableColumn>
        <TableColumn
          field="operator"
          :title="$t('deployPolicy.table.operator')"
          :min-width="120"
          show-overflow="tooltip"
        >
          <template #default="{ row }">
            <UserNameDisplay :name="row.operator" />
          </template>
        </TableColumn>
        <TableColumn
          field="enabled"
          :title="$t('deployPolicy.table.status')"
          :min-width="100"
        >
          <template #default="{ row }">
            <Tag v-if="row.enabled" theme="success">{{ $t('deployPolicy.table.enabled') }}</Tag>
            <Tag v-else>{{ $t('deployPolicy.table.disabled') }}</Tag>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="$t('deployPolicy.table.action')"
          :min-width="140"
          fixed="right"
        >
          <template #default="{ row }">
            <Button theme="primary" text @click="handleViewHistory(row)">
              {{ $t('deployPolicy.action.viewHistory') }}
            </Button>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>

<script lang="ts" setup>
import { Button, Loading, SearchSelect, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { DeployPolicy } from '@/@types/deploy_policy';
import { DeployPolicyAPIService } from '@/api/modules/deploy_policy';
import UserNameDisplay from '@/components/user-name-display.vue';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();

const maxHeight = computed(() => mainStore.windowInnerHeight - 255 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<DeployPolicy[]>([]);
const loading = ref(false);

const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  { id: 'deploy_policy_name', name: t('deployPolicy.table.name'), children: [] },
  { id: 'operator', name: t('deployPolicy.table.operator'), children: [] },
]);

const handleSearchSelectChange = async () => {};

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  await getDeployPolicyList();
};
const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getDeployPolicyList();
};

// 模糊匹配字段
const fuzzyKeys = new Set(['deploy_policy_name', 'operator']);
const getParams = () => {
  const params: any = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: {},
    fuzzy_include_conditions: {},
  };
  searchSelectValue.value.forEach((item: any) => {
    const target: Record<string, any> = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};

const getDeployPolicyList = async () => {
  loading.value = true;
  const res = await DeployPolicyAPIService.ListDeployPolicy(getParams()).catch(() => ({ total: 0, items: [] }));
  loading.value = false;
  pagination.count = res.total;
  tableData.value = res.items || [];
};
const debounceDeployPolicyList = debounce(getDeployPolicyList, 300);

// 查看执行历史：跳转到插件 workflow 历史页，并按部署策略筛选
const handleViewHistory = (row: DeployPolicy) => {
  router.push({
    name: 'history',
    query: {
      active: 'plugin',
      f_deploy_policy_id: JSON.stringify([{ id: String(row.deploy_policy_id), name: String(row.deploy_policy_id) }]),
    },
  });
};

watch([searchSelectValue], async () => {
  await debounceDeployPolicyList();
}, { immediate: true, deep: true });
</script>
