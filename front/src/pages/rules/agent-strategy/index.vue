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
      <div class="flex items-center gap-[10px]">
        <Button
          theme="primary"
          :class="{ 'unAuthorized': !hasManageAuth }"
          @click="hasManageAuth ? handleCreate() : handleAuthClick($event)"
          @mouseenter="handleMouseEnter($event, hasManageAuth)"
          @mousemove="handleMouseMove($event, hasManageAuth)"
          @mouseleave="handleMouseLeave()"
        >
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          {{ $t('agentStrategy.createConfig') }}
        </Button>
        <Button @click="handleSortPreview">
          {{ $t('agentStrategy.sortAndPreview') }}
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
          field="priority"
          :title="$t('agentStrategy.table.priority')"
          :min-width="80"
          fixed="left"
        >
          <template #default="{ row }">
            <Tag v-if="row.enabled" class="priority-tag">
              {{ enabledPriorityIndex(row) + 1 }}
            </Tag>
            <span v-else>-</span>
          </template>
        </TableColumn>
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
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleUpdate(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
              >{{ $t('agentStrategy.action.view') }}</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="!row.enabled"
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleEnabled(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
              >{{ $t('agentStrategy.action.enable') }}</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="row.enabled"
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleDisabled(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
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
                  :class="{ 'unAuthorized': !hasManageAuth }"
                  :disabled="row.enabled && hasManageAuth"
                  @click="!hasManageAuth ? handleAuthClick($event) : (row.isDeletePopShow = true)"
                  @mouseenter="handleMouseEnter($event, hasManageAuth)"
                  @mousemove="handleMouseMove($event, hasManageAuth)"
                  @mouseleave="handleMouseLeave()"
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
    :mode="sidesliderMode"
    :config-data="currentEditConfig"
    @save="handleSave"
    @request-edit="handleRequestEdit"
  />

  <!-- SortPreview 排序并预览组件 -->
  <SortPreview
    v-model:is-show="isShowSortPreview"
    :config-list="tableData"
    :configpolicy-type="configpolicyType"
    :biz-id="strategyBizId || 0"
    @save-sort="handleSaveSort"
    @view-policy="handleViewPolicy"
  />

</template>
<script lang="ts" setup>
import { Button, Loading, Message, PopConfirm, Popover, SearchSelect, Tab, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

// 导入create-config组件
import CreateConfig from './create-config.vue';
import SortPreview from './sort-preview.vue';

import type { ConfigPolicyExactConditions, ConfigPolicyFuzzyConditions } from '@/@types/configpolicy';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { TopoService } from '@/api/modules/topo';
import { formatTimestamp } from '@/common/util';
import useAuthLock from '@/composables/use-auth-lock';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const configpolicyType = computed(() => (route.name === 'agentStrategy' ? 'config_policy_agent' : 'config_policy_proxy'));

// Button-level permission: config_policy_manage for agent/proxy strategy
const manageAction = 'config_policy_manage';
const { hasAuth: hasManageAuth, handleMouseEnter, handleMouseMove, handleMouseLeave, handleAuthClick } = useAuthLock(
  manageAction,
  () => mainStore.strategyBizId,
);
const maxHeight = computed(() => mainStore.windowInnerHeight - 255 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<ConfigPolicy[]>([]);
const loading = ref(false);

// Priority: discrete index among enabled rows (0-based), theme by rank
const enabledRows = computed(() => tableData.value.filter(row => row.enabled));
const enabledPriorityIndex = (row: ConfigPolicy) => {
  const idx = enabledRows.value.indexOf(row);
  return idx >= 0 ? idx : -1;
};
const active = ref('');
const panels = computed(() => [
  { name: 'configStrategy', label: t('agentStrategy.configStrategy') },
  // { name: 'deploymentStrategy', label: t('agentStrategy.deploymentStrategy') },
]);

// Sideslider控制变量
const isShowSideslider = ref(false);
const sidesliderMode = ref<'create' | 'edit' | 'view'>('create');
const currentEditConfig = ref<ConfigPolicy | null>(null);// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'priority',
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
  sidesliderMode.value = 'view';
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

// 查看模式下点击编辑按钮，切换到编辑模式
const handleRequestEdit = () => {
  sidesliderMode.value = 'edit';
};

// ===== SortPreview =====
const isShowSortPreview = ref(false);
const handleSortPreview = () => {
  isShowSortPreview.value = true;
};

const handleSaveSort = async () => {
  await getConfigPolicyList();
};

// 从 sort-preview 点击策略名称跳转到查看页
const handleViewPolicy = async (configpolicyId: number) => {
  try {
    const res = await ConfigPolicyAPIService.ConfigPolicyGet({
      configpolicy_id: configpolicyId,
    });
    sidesliderMode.value = 'view';
    currentEditConfig.value = { ...res };
    mainStore.updateConfigEditData({ ...res });
    isShowSideslider.value = true;
  } catch {
    // error handled by fetch interceptor
  }
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
const networkUnitList = ref<NetworkUnitBrief[]>([]);
const getNetworkUnitList = async (data: {bk_networkunit_id: number}[]) => {
  const res = await TopoService.NetworkUnitListBrief({
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
      bk_biz_id: strategyBizId.value ? [strategyBizId.value] : [],
    } as ConfigPolicyExactConditions,
    fuzzy_include_conditions: {} as ConfigPolicyFuzzyConditions,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target: Record<string, any> = fuzzyKeys.has(item.id)
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
      .find(biz => item.bk_biz_id === biz.bk_biz_id)
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

<style lang="postcss">
.unAuthorized {
  color: #C4C6CC !important;
}
.priority-tag {
  min-width: 28px;
  text-align: center;
  background: #E1ECFF !important;
  color: #699DF4 !important;
  border-color: #E1ECFF !important;
}
</style>
