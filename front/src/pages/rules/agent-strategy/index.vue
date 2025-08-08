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
          v-model="searchSelectValue"
          :uniqueSelect="true"
          :placeholder="'请输入 配置名称、范围、修改人 搜索'"
          @update:modelValue="handleSearchSelectChange"
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
          field="version"
          :title="'版本'"
          :min-width="120"
        ></TableColumn>
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
          show-overflow="tooltip"
        >
          <template #default="{ row }">
            <Button text theme="primary">{{ row.scopes?.length }}</Button>
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
          :min-width="120"
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
            <Tag v-else >未启用</Tag>
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
              <Button
                theme="primary"
                text
                :disabled="row.enabled"
                @click="handleDelete(row)"
                v-bk-tooltips="{
                  content: '启用中的配置不可删除',
                  disabled: !row.enabled
                }"
              >删除</Button>
            </div>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>
<script lang="ts" setup>
import { Button, Tab, SearchSelect, Loading, Tag } from 'bkui-vue';
import { Table, TableColumn } from "@blueking/table";
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useMainStore } from "@/stores/main";
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import useTableSetting from "@/composables/use-table-setting";
import { useRoute, useRouter } from "vue-router";
import { formatTimestamp } from '@/common/util';

const mainStore = useMainStore();
const route = useRoute();
const router = useRouter();
const nodeRole = computed(() => route.name === 'agentStrategy' ? 'agent' : 'proxy');
const maxHeight = computed(() => mainStore.windowInnerHeight - 255);
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableData = ref<ConfigPolicy[]>([]);
const loading = ref(false);
const active = ref('');
const panels = ref([
  {name: 'configStrategy', label: '配置策略'},
  {name: 'deploymentStrategy', label: '部署策略'}
]);
// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    "configpolicy_name",
    "version",
    "biz_name ",
    "remark",
    "scopes",
    "operator",
    "updated_time",
    "enabled",
    "action",
  ],
  disabled: ["action"],
});
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = ref([
  {
    id: "name",
    name: '配置名称',
    children: [],
  },
  {
    id: "range",
    name: '范围',
    children: [],
  },
  {
    id: "operate",
    name: '修改人',
    children: [],
  },
]);
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string,name: string}[]}[]) => {
  
}
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex(
    (item: any) => item.id === field
  );
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: field,
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'enabled':
            name = item ? "启用" : "禁用";
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
  mainStore.updateConfigEditData({...row})
  router.push({
    name: 'editConfig',
    params: {
      node_role: nodeRole.value
    }
  });
};
const handleCreate = () => {
  router.push({
    name: 'createConfig',
    params: {
      node_role: nodeRole.value
    }
  });
}
const handleEnabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyEnable({
    configpolicy_id: [row.configpolicy_id]
  });
  await getConfigPolicyList();
};
const handleDisabled = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDisable({
    configpolicy_id: [row.configpolicy_id]
  });
  await getConfigPolicyList();
};
const handleDelete = async (row: ConfigPolicy) => {
  await ConfigPolicyAPIService.ConfigPolicyDelete({
    configpolicy_id: [row.configpolicy_id]
  });
  await getConfigPolicyList();
};
const getParam = () => {
  const params: {
    page: Page,
    exact_include_conditions: {
      node_role: string[];
    },
    fuzzy_include_conditions?: {
      configpolicy_name?: string[]
    }
  } = {
    page: {
      offset: pagination.current - 1,
      limit: pagination.limit
    },
    exact_include_conditions: {
      node_role: [nodeRole.value]
    },
  }
  let configpolicy_name: string[] = []
  searchSelectValue.value.forEach((item: any) => {
    if(item.id === 'configpolicy_name') {
      configpolicy_name = item.values.map((val: any) => val.id)
    }
  });
  if(configpolicy_name.length && params.fuzzy_include_conditions) {
    params.fuzzy_include_conditions.configpolicy_name = configpolicy_name
  }
  return params;
}
const getConfigPolicyList = async () => {
  loading.value = true;
  const res = await ConfigPolicyAPIService.ConfigPolicyList(getParam()).catch(() => ({
    total: 0,
    items: []
  }));
  loading.value = false;
  tableData.value = res.items;
}
watch(() => route.name, async () => {
  await getConfigPolicyList();
})
onMounted(async () => {
  await getConfigPolicyList();
});
</script>