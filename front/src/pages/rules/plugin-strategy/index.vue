<template>
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
          {{ $t('agentStrategy.createPluginConfig') }}
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
          field="plugin_name"
          :title="$t('agentStrategy.table.pluginName')"
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
            <Popover theme="light" trigger="hover">
              <Button text theme="primary">{{ row.scopes?.length }}</Button>
              <template #content>
                <Table :data="row.scopes" :min-width="600">
                  <TableColumn field="bk_networkarea_id" :title="$t('taskDetail.table.workarea')" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ networkAreaList?.find(item => item.bk_networkarea_id === scopesRow.bk_networkarea_id)?.bk_networkarea_name || $t('agentStrategy.table.unlimited') }}
                    </template>
                  </TableColumn>
                  <TableColumn field="bk_networkunit_id" :title="$t('taskDetail.table.workUnit')" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ networkUnitList?.find(item => item.bk_networkunit_id === scopesRow.bk_networkunit_id)?.bk_networkunit_name || $t('agentStrategy.table.unlimited') }}
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
        >
          <template #default="{ row }">
            <UserNameDisplay :name="row.operator" />
          </template>
        </TableColumn>
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
                class="mr-[8px]" theme="primary" text
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleUpdate(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
              >{{ $t('agentStrategy.action.view') }}</Button>
              <Button
                class="mr-[8px]" theme="primary" text
                v-if="!row.enabled"
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleEnabled(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
              >{{ $t('agentStrategy.action.enable') }}</Button>
              <Button
                class="mr-[8px]" theme="primary" text
                v-if="row.enabled"
                :class="{ 'unAuthorized': !hasManageAuth }"
                @click="hasManageAuth ? handleDisabled(row) : handleAuthClick($event)"
                @mouseenter="handleMouseEnter($event, hasManageAuth)"
                @mousemove="handleMouseMove($event, hasManageAuth)"
                @mouseleave="handleMouseLeave()"
              >{{ $t('agentStrategy.action.disable') }}</Button>
              <PopConfirm
                width="360" theme="light" trigger="click" placement="top-start"
                :title="$t('agentStrategy.action.confirmDelete')"
                :confirm-text="$t('agentStrategy.action.delete')"
                @confirm="handleDelete(row)"
              >
                <Button
                  theme="primary" text
                  :class="{ 'unAuthorized': !hasManageAuth }"
                  :disabled="row.enabled && hasManageAuth"
                  @click="!hasManageAuth ? handleAuthClick($event) : (row.isDeletePopShow = true)"
                  @mouseenter="handleMouseEnter($event, hasManageAuth)"
                  @mousemove="handleMouseMove($event, hasManageAuth)"
                  @mouseleave="handleMouseLeave()"
                  v-bk-tooltips="{ content: $t('agentStrategy.action.deleteDisabledTip'), disabled: !row.enabled }"
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

  <PluginCreateConfig
    v-model:is-show="isShowSideslider"
    :mode="sidesliderMode"
    :config-data="currentEditConfig"
    @save="handleSave"
    @request-edit="handleRequestEdit"
  />
</template>

<script lang="ts" setup>
import { Button, Loading, Message, PopConfirm, Popover, SearchSelect, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import PluginCreateConfig from './create-config.vue';

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

const manageAction = 'config_policy_manage';
const { hasAuth: hasManageAuth, handleMouseEnter, handleMouseMove, handleMouseLeave, handleAuthClick } = useAuthLock(
  manageAction,
  () => mainStore.strategyBizId,
);

const maxHeight = computed(() => mainStore.windowInnerHeight - 255 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<ConfigPolicy[]>([]);
const loading = ref(false);

const enabledRows = computed(() => tableData.value.filter(row => row.enabled));
const enabledPriorityIndex = (row: ConfigPolicy) => {
  const idx = enabledRows.value.indexOf(row);
  return idx >= 0 ? idx : -1;
};

const isShowSideslider = ref(false);
const sidesliderMode = ref<'create' | 'edit' | 'view'>('create');
const currentEditConfig = ref<ConfigPolicy | null>(null);
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: ['priority', 'configpolicy_name', 'version', 'biz_name', 'plugin_name', 'remark', 'scopes', 'operator', 'updated_time', 'enabled', 'action'],
  disabled: ['action'],
}, 'rulesMng-config_policy_plugin');

const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  { id: 'configpolicy_name', name: t('agentStrategy.table.configName'), children: [] },
  { id: 'plugin_name', name: t('agentStrategy.table.pluginName'), children: [] },
  { id: 'operator', name: t('agentStrategy.table.operator'), children: [] },
]);

const handleSearchSelectChange = async () => {};
const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field, name: field,
      values: checked.map((item: any) => ({
        id: item,
        name: field === 'enabled' ? (item ? t('agentStrategy.filter.enabled') : t('agentStrategy.filter.disabled')) : item,
      })),
    });
  }
};

const pageLimitChange = async (limit: number) => { pagination.limit = limit; await getConfigPolicyList(); };
const pageValueChange = async (current: number) => { pagination.current = current; await getConfigPolicyList(); };

const handleUpdate = async (row: ConfigPolicy) => {
  const res = await ConfigPolicyAPIService.ConfigPolicyGet({ configpolicy_id: row.configpolicy_id });
  sidesliderMode.value = 'view';
  currentEditConfig.value = { ...res };
  mainStore.updateConfigEditData({ ...res });
  isShowSideslider.value = true;
};

const handleCreate = () => {
  sidesliderMode.value = 'create';
  currentEditConfig.value = null;
  mainStore.updateConfigEditData(null);
  isShowSideslider.value = true;
};

const handleRequestEdit = () => { sidesliderMode.value = 'edit'; };

const handleEnabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyEnable({ configpolicy_id: [row.configpolicy_id] });
  await getConfigPolicyList();
};
const handleDisabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDisable({ configpolicy_id: [row.configpolicy_id] });
  await getConfigPolicyList();
};
const handleDelete = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDelete({ configpolicy_id: [row.configpolicy_id] });
  await getConfigPolicyList();
};

const handleSave = async () => { await getConfigPolicyList(); };

const networkAreaList = ref<NetworkArea[]>([]);
const getNetworkAreaList = async (data: { bk_networkarea_id: number }[]) => {
  const res = await TopoService.NetworkAreaList({
    page: { limit: 0 },
    exact_include_conditions: { bk_networkarea_id: data.map((item: any) => item.bk_networkarea_id) },
  }).catch(() => ({ total: 0, items: [] }));
  networkAreaList.value = res.items;
};

const networkUnitList = ref<NetworkUnitBrief[]>([]);
const getNetworkUnitList = async (data: { bk_networkunit_id: number }[]) => {
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: { bk_networkunit_id: data.map((item: any) => item.bk_networkunit_id) },
  }).catch(() => ({ total: 0, items: [] }));
  networkUnitList.value = res.items;
};

const fuzzyKeys = new Set(['configpolicy_name', 'plugin_name', 'operator']);
const getParams = () => {
  const params = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: {
      configpolicy_type: ['config_policy_plugin'],
      bk_biz_id: strategyBizId.value ? [strategyBizId.value] : [],
    } as ConfigPolicyExactConditions,
    fuzzy_include_conditions: {} as ConfigPolicyFuzzyConditions,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target: Record<string, any> = fuzzyKeys.has(item.id) ? params.fuzzy_include_conditions : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};

const getConfigPolicyList = async () => {
  loading.value = true;
  const res = await ConfigPolicyAPIService.ConfigPolicyList(getParams()).catch(() => ({ total: 0, items: [] }));
  loading.value = false;
  pagination.count = res.total;
  const allScopes = res.items.flatMap(item => item.scopes || []);
  if (allScopes.length > 0) {
    await Promise.all([getNetworkAreaList(allScopes), getNetworkUnitList(allScopes)]);
  }
  tableData.value = res.items.map((item: any) => ({
    ...item,
    biz_name: mainStore.businessList.find(biz => item.bk_biz_id === biz.bk_biz_id)?.bk_biz_name || t('agentStrategy.table.unlimited'),
  }));
};
const debounceConfigPolicyList = debounce(getConfigPolicyList, 300);

const strategyBizId = computed(() => mainStore.strategyBizId);

watch([searchSelectValue, strategyBizId], async () => {
  if (strategyBizId.value) await debounceConfigPolicyList();
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
