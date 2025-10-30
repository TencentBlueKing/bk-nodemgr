<template>
  <Tab
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
  </Tab>
  <div class="p-[24px] mt-[41px]">
    <div class="flex items-center justify-between">
      <div>
        <Button theme="primary" @click="handleCreate">
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          新建配置
        </Button>
      </div>
      <div class="">
        <SearchSelect
          class="w-[480px] z-99"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="'请输入 配置名称、修改人 搜索'"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </div>
    <Loading
      title="数据加载中"
      :loading="loading"
      class="mt-[16px] flex-1 overflow-auto"
    >
      <Table
        class="w-full"
        :max-height="maxHeight"
        :data="tableData"
        :empty-text="'暂无数据'"
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
          :title="'配置名称'"
          :min-width="150"
          fixed="left"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="configpolicy_id"
          :title="'配置ID'"
          :min-width="100"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="version"
          :title="'版本'"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ `V${row.version}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="biz_name"
          :title="'业务'"
          :min-width="120"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="remark"
          :title="'备注'"
          :min-width="180"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="scopes"
          :title="'范围'"
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
                  <TableColumn field="bk_networkarea_id" title="管控区域" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ networkAreaList?.find(item =>
                        item.bk_networkarea_id === scopesRow.bk_networkarea_id )?.bk_networkarea_name || '不限' }}
                    </template>
                  </TableColumn>
                  <TableColumn field="bk_networkunit_id" title="管控单元" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ networkUnitList?.find(item =>
                        item.bk_networkunit_id === scopesRow.bk_networkunit_id )?.bk_networkunit_name || '不限' }}
                    </template>
                  </TableColumn>
                  <TableColumn field="os_type" title="操作系统">
                    <template #default="{ row: scopesRow }">
                      {{ scopesRow.os_type || '不限' }}
                    </template>
                  </TableColumn>
                  <TableColumn field="cpu_arch" title="架构" :min-width="120">
                    <template #default="{ row: scopesRow }">
                      {{ scopesRow.cpu_arch || '不限' }}
                    </template>
                  </TableColumn>
                </Table>
              </template>
            </Popover>
          </template>
        </TableColumn>
        <TableColumn
          field="operator"
          :title="'修改人'"
          :min-width="120"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="updated_time"
          :title="'修改时间'"
          :min-width="150"
        >
          <template #default="{ row }">
            {{ formatTimestamp(row.updated_time) }}
          </template>
        </TableColumn>
        <TableColumn
          field="enabled"
          :title="'状态'"
          :min-width="100"
        >
          <template #default="{ row }">
            <Tag v-if="row.enabled" theme="success">启用</Tag>
            <Tag v-else>未启用</Tag>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="'操作'"
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
              >编辑</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="!row.enabled"
                @click="handleEnabled(row)"
              >启用</Button>
              <Button
                class="mr-[8px]"
                theme="primary"
                text
                v-if="row.enabled"
                @click="handleDisabled(row)"
              >停用</Button>
              <PopConfirm
                width="360"
                theme="light"
                trigger="click"
                placement="top-start"
                title="确认删除该配置策略？"
                confirm-text="删除"
                @confirm="handleDelete(row)"
              >
                <Button
                  theme="primary"
                  text
                  :disabled="row.enabled"
                  @click="row.isDeletePopShow = true"
                  v-bk-tooltips="{
                    content: '启用中的配置不可删除',
                    disabled: !row.enabled
                  }"
                >删除</Button>
                <template #content>
                  <div class="px-[4px] pb-[4px]">
                    <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">删除目标：{{row.configpolicy_name}}</div>
                    <div class="text-[12px] text-[#4D4F56] w-full">删除后不可恢复，请谨慎操作！</div>
                  </div>
                </template>
              </PopConfirm>
            </div>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>
<script lang="ts" setup>
import { Button, Loading, PopConfirm, Popover, SearchSelect, Tab, Tag } from 'bkui-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { ConfigPolicyExactConditions, ConfigPolicyFuzzyConditions } from '@/@types/configpolicy';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { TopoService } from '@/api/modules/topo';
import { formatTimestamp } from '@/common/util';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const mainStore = useMainStore();
const route = useRoute();
const router = useRouter();
const configpolicyType = computed(() => (route.name === 'agentStrategy' ? 'config_policy_agent' : 'config_policy_proxy'));
const maxHeight = computed(() => mainStore.windowInnerHeight - 255);
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<ConfigPolicy[]>([]);
const loading = ref(false);
const active = ref('');
const panels = ref([
  { name: 'configStrategy', label: '配置策略' },
  // { name: 'deploymentStrategy', label: '部署策略' },
]);
// 表格
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
const searchSelectData = ref([
  {
    id: 'configpolicy_name',
    name: '配置名称',
    children: [],
  },
  {
    id: 'operator',
    name: '修改人',
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
            name = item ? '启用' : '禁用';
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
  mainStore.updateConfigEditData({ ...res });
  router.push({
    name: 'editConfig',
    params: {
      configpolicy_type: configpolicyType.value,
    },
  });
};
const handleCreate = () => {
  router.push({
    name: 'createConfig',
    params: {
      configpolicy_type: configpolicyType.value,
    },
  });
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
const networkAreaList = ref<NetworkArea[]>([]);
// 管控区域
const getNetworkAreaList = async () => {
  const res = await TopoService.NetworkAreaList({}).catch(() => ({
    total: 0,
    items: [],
  }));
  networkAreaList.value = res.items;
};
// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({}).catch(() => ({
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
      biz_id: mainStore.selectedBusinessId,
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
  tableData.value = res.items.map((item: any) => ({
    ...item,
    biz_name: mainStore.businessList
      .filter(biz => item.biz_id.includes(biz.bk_biz_id))
      .map(item => item.bk_biz_name)
      .join(',') || '不限',
  }));
};
watch([
  () => route.name,
  searchSelectValue,
  () => mainStore.selectedBusinessId,
], async () => {
  await getConfigPolicyList();
}, { immediate: true, deep: true });
onMounted(async () => {
  await getNetworkAreaList();
  await getNetworkUnitList();
});
</script>
