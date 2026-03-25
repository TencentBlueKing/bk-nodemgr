<template>
  <!-- <Tab
    v-model:active="active"
    type="unborder-card"
    :label-height="41"
    class="text-[14px] bg-[#fff] h-[41px] absolute z-10 w-full"
  >
    <Tab.TabPanel
      v-for="item in panels"
      :key="item.name"
      :label="item.label"
      :name="item.name"
    >
    </Tab.TabPanel>
  </Tab> -->
  <div class="p-[24px]">
    <div class="flex items-center justify-between">
      <div>
        <Button theme="primary" @click="handleCreate">
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          {{ $t('agentStrategy.createConfig') }}
        </Button>
      </div>
      <div class="">
        <SearchSelect
          class="w-[480px] z-99 bg-[#fff]"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('agentStrategy.searchPlaceholder')"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
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
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn
          field="configpolicy_name"
          :title="$t('agentStrategy.table.configName')"
          :min-width="150"
          fixed="left"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="configpolicy_id"
          :title="$t('agentStrategy.table.configId')"
          :min-width="100"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="version"
          :title="$t('agentStrategy.table.version')"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ `V${row.version}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="biz_name"
          :title="$t('agentStrategy.table.business')"
          :min-width="120"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="remark"
          :title="$t('agentStrategy.table.remark')"
          :min-width="180"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="scopes"
          :title="$t('agentStrategy.table.scope')"
          :min-width="120"
        >
          <template #default="{ row }">
            <Popover
              theme="light"
              trigger="hover"
            >
              <Button text theme="primary">{{ row.scopes?.length }}</Button>
              <template #content>
                <Table :data="row.scopes" :min-width="600">
                  <TableColumn field="bk_networkarea_id" :title="$t('taskDetail.table.workarea')" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      <!-- eslint-disable-next-line max-len -->
                      {{ networkAreaList?.find(item => item.bk_networkarea_id === scopesRow.bk_networkarea_id )?.bk_networkarea_name || $t('agentStrategy.table.unlimited') }}
                    </template>
                  </TableColumn>
                  <TableColumn field="bk_networkunit_id" :title="$t('taskDetail.table.workUnit')" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      <!-- eslint-disable-next-line max-len -->
                      {{ networkUnitList?.find(item => item.bk_networkunit_id === scopesRow.bk_networkunit_id )?.bk_networkunit_name || $t('agentStrategy.table.unlimited') }}
                    </template>
                  </TableColumn>
                  <TableColumn field="os_type" :title="$t('agentStrategy.table.os')">
                    <template #default="{ row: scopesRow }">
                      {{ scopesRow.os_type || $t('agentStrategy.table.unlimited') }}
                    </template>
                  </TableColumn>
                  <TableColumn field="cpu_arch" :title="$t('agentStrategy.table.arch')" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ scopesRow.cpu_arch || $t('agentStrategy.table.unlimited') }}
                    </template>
                  </TableColumn>
                </Table>
              </template>
            </Popover>
          </template>
        </TableColumn>
        <TableColumn
          field="operator"
          :title="$t('agentStrategy.table.operator')"
          :min-width="120"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="updated_time"
          :title="$t('agentStrategy.table.updateTime')"
          :min-width="150"
        >
          <template #default="{ row }">
            {{ formatTimestamp(row.updated_time) }}
          </template>
        </TableColumn>
        <TableColumn
          field="enabled"
          :title="$t('agentStrategy.table.status')"
          :min-width="100"
        >
          <template #default="{ row }">
            <Tag v-if="row.enabled" theme="success">{{ $t('agentStrategy.table.enabled') }}</Tag>
            <Tag v-else>{{ $t('agentStrategy.table.disabled') }}</Tag>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="$t('agentStrategy.table.action')"
          :min-width="120"
          fixed="right"
        >
          <template #default="{ row }">
            <div class="flex items-center">
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                @click="handleUpdate(row)"
              >{{ $t('agentStrategy.action.edit') }}</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="!row.enabled"
                @click="handleEnabled(row)"
              >{{ $t('agentStrategy.action.enable') }}</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="row.enabled"
                @click="handleDisabled(row)"
              >{{ $t('agentStrategy.action.disable') }}</Button>
              <PopConfirm
                width="360"
                theme="light"
                trigger="click"
                placement="top-start"
                :title="$t('agentStrategy.action.confirmDelete')"
                :confirm-text="$t('agentStrategy.action.delete')"
                @confirm="handleDelete(row)"
              >
                <Button
                  theme="primary"
                  text
                  :disabled="row.enabled"
                  @click="row.isDeletePopShow = true"
                  v-bk-tooltips="{
                    content: $t('agentStrategy.action.deleteDisabledTip'),
                    disabled: !row.enabled
                  }"
                >{{ $t('agentStrategy.action.delete') }}</Button>
                <template #content>
                  <div class="px-[4px] pb-[4px]">
                    <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">{{ $t('agentStrategy.action.deleteTarget', { name: row.configpolicy_name }) }}</div>
                    <div class="text-[12px] text-[#4D4F56] w-full">{{ $t('agentStrategy.action.deleteTip') }}</div>
                  </div>
                </template>
              </PopConfirm>
            </div>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>

  <!-- CreateConfig组件 -->
  <CreateConfig
    v-model:is-show="isShowSideslider"
    :configpolicy-type="configpolicyType"
    :is-edit="sidesliderMode === 'edit'"
    :config-data="currentEditConfig"
    @save="handleSave"
  />
</template>
<script lang="ts" setup>
import { Button, Loading, PopConfirm, Popover, SearchSelect, Sideslider, Tab, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

// 导入create-config组件
import CreateConfig from './create-config.vue';

import type { ConfigPolicyExactConditions, ConfigPolicyFuzzyConditions } from '@/@types/configpolicy';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { TopoService } from '@/api/modules/topo';
import { formatTimestamp } from '@/common/util';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const configpolicyType = computed(() => (route.name === 'agentStrategy' ? 'config_policy_agent' : 'config_policy_proxy'));
const maxHeight = computed(() => mainStore.windowInnerHeight - 255 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<ConfigPolicy[]>([]);
const loading = ref(false);
const active = ref('');
const panels = computed(() => [
  { name: 'configStrategy', label: t('agentStrategy.configStrategy') },
  // { name: 'deploymentStrategy', label: t('agentStrategy.deploymentStrategy') },
]);

// Sideslider控制变量
const isShowSideslider = ref(false);
const sidesliderMode = ref<'create' | 'edit'>('create');
const currentEditConfig = ref<ConfigPolicy | null>(null);// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'configpolicy_name',
    'version',
    'biz_name ',
    'remark',
    'scopes',
    'operator',
    'updated_time',
    'enabled',
    'action',
  ],
  disabled: ['action'],
}, `rulesMng-${configpolicyType.value}`);
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  {
    id: 'configpolicy_name',
    name: t('agentStrategy.table.configName'),
    children: [],
  },
  {
    id: 'operator',
    name: t('agentStrategy.table.operator'),
    children: [],
  },
]);
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string, name: string}[]}[]) => {

};
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: field,
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'enabled':
            name = item ? t('agentStrategy.filter.enabled') : t('agentStrategy.filter.disabled');
            break;
          default:
            name = item;
            break;
        }
        return {
          id: item,
          name,
        };
      }),
    });
  }
};
// 分页操作
const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  await getConfigPolicyList();
};
const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getConfigPolicyList();
};
const handleUpdate = async (row: ConfigPolicy) => {
  const res = await ConfigPolicyAPIService.ConfigPolicyGet({
    configpolicy_id: row.configpolicy_id,
  });
  sidesliderMode.value = 'edit';
  currentEditConfig.value = { ...res };
  mainStore.updateConfigEditData({ ...res });
  isShowSideslider.value = true;
};
const handleCreate = () => {
  // 打开侧边栏创建配置
  sidesliderMode.value = 'create';
  currentEditConfig.value = null;
  mainStore.updateConfigEditData(null);
  isShowSideslider.value = true;
};
const handleEnabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyEnable({
    configpolicy_id: [row.configpolicy_id],
  });
  await getConfigPolicyList();
};
const handleDisabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDisable({
    configpolicy_id: [row.configpolicy_id],
  });
  await getConfigPolicyList();
};
const handleDelete = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDelete({
    configpolicy_id: [row.configpolicy_id],
  });
  await getConfigPolicyList();
};

// Sideslider关闭后的回调
const handleSave = async () => {
  await getConfigPolicyList();
};
const networkAreaList = ref<NetworkArea[]>([]);
// 管控区域
const getNetworkAreaList = async (data: {bk_networkarea_id: number}[]) => {
  const res = await TopoService.NetworkAreaList({
    page: {
      limit: 0,
    },
    exact_include_conditions: {
      bk_networkarea_id: data.map((item: any) => item.bk_networkarea_id),
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  networkAreaList.value = res.items;
};
// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async (data: {bk_networkunit_id: number}[]) => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkunit_id: data.map((item: any) => item.bk_networkunit_id),
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  networkUnitList.value = res.items;
};
const fuzzyKeys = new Set([
  'configpolicy_name',
  'operator',
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      configpolicy_type: [configpolicyType.value],
      biz_id: strategyBizId.value ? [strategyBizId.value] : [],
    } as ConfigPolicyExactConditions,
    fuzzy_include_conditions: {} as ConfigPolicyFuzzyConditions,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getConfigPolicyList = async () => {
  loading.value = true;
  const res = await ConfigPolicyAPIService.ConfigPolicyList(getParams()).catch(() => ({
    total: 0,
    items: [],
  }));
  loading.value = false;
  pagination.count = res.total;
  const allScopes = res.items.flatMap(item => item.scopes || []);
  if (allScopes.length > 0) {
    await Promise.all([getNetworkAreaList(allScopes), getNetworkUnitList(allScopes)]);
  }
  tableData.value = res.items.map((item: any) => ({
    ...item,
    biz_name: mainStore.businessList
      .find(biz => item.biz_id === biz.bk_biz_id)
      ?.bk_biz_name || t('agentStrategy.table.unlimited'),
  }));
};
const debounceConfigPolicyList = debounce(getConfigPolicyList, 300);

// 从 store 获取策略专用的单选业务 ID
const strategyBizId = computed(() => mainStore.strategyBizId);

watch([
  () => route.name,
  searchSelectValue,
  strategyBizId,
], async () => {
  if (strategyBizId.value) {
    await debounceConfigPolicyList();
  }
}, { immediate: true, deep: true });
</script>
