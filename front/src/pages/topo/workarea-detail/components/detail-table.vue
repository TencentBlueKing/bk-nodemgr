<template>
  <div>
    <bk-loading
      title="数据加载中"
      :loading="loading"
      class="w-full overflow-auto mt-[16px]"
    >
      <Table
        class="w-full filterTable"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        empty-cell-text="--"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
        @setting-change="handleSettingChange"
        @column-filter="handleColumnFilter">
        <TableColumn type="checkbox" :width="60" :resizable="false" fixed="left" />
        <TableColumn
          field="bk_host_id"
          title="Host ID"
          fixed="left"
          :min-width="100"
        ></TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv4')"
          field="bk_host_innerip"
          show-overflow="tooltip"
          fixed="left"
          :min-width="150">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv6')"
          field="bk_host_innerip_v6"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          label="出口IP"
          field="export_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          label="服务IP"
          field="advertise_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          label="登录IP"
          field="login_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          label="所属业务"
          field="bk_biz_id"
          show-overflow="tooltip"
          :min-width="100">
          <template #default="{ row }">
            {{ businessList.find(item => item.bk_biz_id === row.bk_biz_id)?.bk_biz_name }}
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
          :min-width="150">
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
          label="proxy服务"
          field="proxy_tags"
          show-overflow="tooltip"
          :min-width="300">
          <template #default="{ row }">
            <div class="flex items-center gap-[4px]">
              <div v-for="tag in row.proxy_tags" :key="tag">
                <Tag>{{ proxyTagMap[tag] }}</Tag>
              </div>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          title="插件数"
          field="pluginNum"
          :min-width="122"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.pluginNum || 0 }}
              <i class="nodeman-icon nc-plug-in ml-[5px]"></i>
            </Button>
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
              <Button theme="primary" text class="mr-[12px]" @click="handleEdit(row)">
                {{ $t('topoManager.workAreaDetail.table.edit') }}
              </Button>
              <MoreAction
                :ipv4="row.bk_host_innerip"
                :data="[row]"
                @reinstall="handleReinstall(row)">
                <i class="nodeman-icon nc-more cursor"></i>
              </MoreAction>
            </div>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
    <edit-proxy-sideslider v-model:is-show="sidesliderData.isShow" :data="sidesliderData.data" @update="handleUpdate">
    </edit-proxy-sideslider>
    <ReinstallProxy v-model:is-show="isShowInstallProxy" :data="reinstallData" :bk_networkunit_id="bkNetworkunitId" />
    <!-- 侧边栏 -->
    <processSideslider
      v-model:is-show="isShowSideslider"
      type="node"
      :node="currentProxy"
    ></processSideslider>
  </div>
</template>

<script lang="ts" setup>
import { Button, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from '../../../node/plugin/process-sideslider.vue';
import ReinstallProxy from '../../install-proxy/reinstall-proxy.vue';

import EditProxySideslider from './edit-proxy-sideslider.vue';
import MoreAction from './more-action.vue';

import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import { ProcessAPIService } from '@/api/modules/process';
import { TopoService } from '@/api/modules/topo';
import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const props = defineProps({
  searchSelectValue: {
    type: Array,
    default: [],
  },
  bkNetworkunitId: {
    type: Number,
    default: 0,
  },
});
const emit = defineEmits(['selectChange', 'getData']);
const route = useRoute();
const workAreaId = Number(route.params.workarea);
const list = ref<Host[]>([]);
const pagination = reactive({ count: 0, limit: 20, current: 1 });
const sortConfig = ref({ multiple: true });
const mainStore = useMainStore();
const businessList = computed(() => mainStore.businessList);
const sidesliderData = reactive<{
  isShow: boolean,
  data: Host | null
}>({
  isShow: false,
  data: null,
});
const proxyTagMap = {
  dedicated_installer: '安装跳板',
  cluster_tunnel: 'Agent控制',
  file_tunnel: '文件传输',
  data_tunnel: '数据上报',
};
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'export_ip',
    'advertise_ip',
    'login_ip',
    'bk_biz_id',
    'bk_agent_id',
    'node_version',
    'node_status',
    'proxy_tags',
    'pluginNum',
    'action',
  ],
  disabled: ['action'],
}, `topoMng-workarea-detail-${String(route.name)}`);
// 表格勾选
const selection = computed(() => list.value.filter((item: any) => item.checked));
const handleSelectChange = ({
  checked,
  row,
}: {
  checked: boolean;
  row: any;
}) => {
  row.checked = checked;
  emit('selectChange', selection.value);
};

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  list.value.forEach((item: any) => (item.checked = checked));
  emit('selectChange', selection.value);
};
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
  list: [],
  checked: [],
});

const filterOptionConfig = (prop: string, valMap?: Record<string, any>) => {
  const uniqueValues = Array.from(new Set(list.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    text:
      valMap && valMap[value as string] ? valMap[value as string].text : value,
    value,
  }));
};

// ---------- 侧边栏 ----------
const isShowSideslider = ref(false);
const currentProxy = ref();
// 打开侧边栏并加载进程列表
const openSidebar = async (proxy) => {
  isShowSideslider.value = true;
  currentProxy.value = proxy;
};

const handleEdit = (row: Host) => {
  sidesliderData.isShow = true;
  sidesliderData.data = row;
};
const fuzzyKeys = new Set([
  'bk_host_innerip',
  'bk_host_innerip_v6',
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      node_role: ['proxy'],
    } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  if (route.name === 'proxy') {
    params.exact_include_conditions.bk_biz_id = mainStore.selectedBusinessId;
  }
  if (props.bkNetworkunitId) {
    params.exact_include_conditions.bk_networkunit_id = [props.bkNetworkunitId];
  }
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getProxyList = async () => {
  loading.value = true;
  const res = await TopoService.HostList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;

  const pluginNumMap = await ProcessAPIService.GetProcessDistributionByHostID({
    exact_include_conditions: {
      bk_host_id: res.items.map((item: any) => item.bk_host_id),
    },
  }).catch((err: any) => {
    console.error('获取插件数量失败:', err);
    return {} as Record<number, number>;
  });

  list.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
    bk_host_innerip: item.info.bk_host_innerip_list.join(','),
    bk_host_innerip_v6: item.info.bk_host_innerip_v6_list.join(','),
    export_ip: item.info.bk_host_outerip_list.join(','),
    advertise_ip: item.info.bk_host_outerip_v6_list.join(','),
    pluginNum: pluginNumMap[item.bk_host_id] || 0,
  }));
  emit('getData', list.value);
  loading.value = false;
};
const handleUpdate = async () => {
  await getProxyList();
};
const isShowInstallProxy = ref(false);
const reinstallData = ref<Host[]>([]);
const handleReinstall = (row: Host) => {
  isShowInstallProxy.value = true;
  reinstallData.value = [row];
};
watch(
  () => list,
  () => {
    proxyVersionFilter.list = filterOptionConfig('node_version');
    proxyStatusFilter.list = filterOptionConfig('node_status');
  },
  { deep: true, immediate: true },
);
watch(route, async () => {
  if (route.query.os_type) {
    searchSelectValue.value.push(...[
      {
        id: 'os_type',
        name: '操作系统',
        values: [{
          id: route.query.os_type,
          name: route.query.os_type,
        }],
      },
      {
        id: 'cpu_arch',
        name: '架构',
        values: [{
          id: route.query.cpu_arch,
          name: route.query.cpu_arch,
        }],
      },
      {
        id: 'node_version',
        name: 'Proxy 版本',
        values: [{
          id: route.query.node_version,
          name: route.query.node_version,
        }],
      },
    ]);
    await getProxyList();
  }
}, { immediate: true, deep: true });
const debounceGetProxyList = debounce(() => {
  getProxyList();
}, 300);
watch(
  [
    () => mainStore.selectedBusinessId,
    () => searchSelectValue.value,
  ],
  async () => {
    await debounceGetProxyList();
  },
  { immediate: true, deep: true },
);
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 14px;
  height: 14px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
  vertical-align: middle;
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
