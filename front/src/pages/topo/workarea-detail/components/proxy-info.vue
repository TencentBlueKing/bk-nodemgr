<template>
  <div :class="{ 'mx-[24px]': isProxyStatus }">
    <FlexRow class="mt-[24px]">
      <template #left>
        <div class="flex items-center">
          <!-- 新建 -->
          <Button theme="primary" class="mr-[8px]" @click="handleInstallProxy">
            <span>{{
              $t("topoManager.workAreaDetail.button.installProxy")
            }}</span>
          </Button>
          <MoreAction
            :data="selectTableData"
            placement="bottom-start"
            :batch="true"
            @reinstall="handleReinstall(selectTableData)"
          >
            <Button :disabled="!selectTableData.length" class="mr-[8px]">
              <span>{{ $t("topoManager.workAreaDetail.button.batch") }}</span>
              <i
                class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
              ></i>
            </Button>
          </MoreAction>
          <!-- 复制 -->
          <copy-ip-dropdown
            :disabled="!selectTableData.length"
            :data="tableData"
            :list="list"
          ></copy-ip-dropdown>
        </div>
      </template>
      <template #right>
        <SearchSelect
          class="w-[480px]"
          :placeholder="$t('topoManager.workAreaDetail.searchSelect.placeholder')"
          :unique-select="true"
          v-model.trim="searchKey"
          :data="searchSelectData"
        >
        </SearchSelect>
      </template>
    </FlexRow>
    <!-- table -->
    <DetailTable
      :search-select-value="searchKey"
      :bk-networkunit-id="active"
      :is-batch-reinstall="batchReinstall"
      @select-change="handleSelectChange"
      @get-data="handleGetData">
    </DetailTable>
  </div>
  <InstallProxy v-model:is-show="isShowInstallProxy" :bk_networkunit_id="active" />
  <ReinstallProxy v-model:is-show="isShowReinstallProxy" :data="reinstallData" />
</template>
<script setup lang="ts">
import { Button, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import DetailTable from './detail-table.vue';
import MoreAction from './more-action.vue';

import CopyIp from '@/components/copy-ip.vue';
import InstallProxy from '@/pages/topo/install-proxy/install-proxy.vue';
import ReinstallProxy from '@/pages/topo/install-proxy/reinstall-proxy.vue';

const props = defineProps({
  active: {
    type: Number,
    default: null,
  },
});
const { t } = useI18n();
const route = useRoute();
const isProxyStatus = computed(() => route.name === 'proxy');
// 搜索
const searchKey = ref<ISearchValue[]>([]);
const searchSelectData = ref<ISearchItem[]>([
  {
    name: t('topoManager.workAreaDetail.table.ipv4'),
    id: 'bk_host_innerip',
  },
  {
    name: t('topoManager.workAreaDetail.table.ipv6'),
    id: 'bk_host_innerip_v6',
  },
  {
    name: 'AgentID',
    id: 'bk_agent_id',
  },
]);
// 复制
const subList = [
  {
    id: 'ipv4',
    name: 'IPv4',
  },
  {
    id: 'ipv6',
    name: 'IPv6',
  },
  {
    id: 'workarea+ipv4',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv4`,
  },
  {
    id: 'workarea+ipv6',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv6`,
  },
];
const list = [
  {
    id: 'select',
    name: t('topoManager.workArea.copy.select'),
    children: subList,
  },
  {
    id: 'all',
    name: t('topoManager.workArea.copy.selectAll'),
    children: subList,
  },
];
const batchReinstall = ref(false);
const isShowReinstallProxy = ref(false);
const reinstallData = ref<Host[]>([]);
const handleReinstall = (data: Host[]) => {
  isShowReinstallProxy.value = true;
  reinstallData.value = data;
};
const isShowInstallProxy = ref(false);
const handleInstallProxy = () => {
  isShowInstallProxy.value = true;
};
const selectTableData = ref<Host[]>([]);
const tableData = ref<Host[]>([]);
const handleSelectChange = (tableList: Host[]) => {
  selectTableData.value = tableList.filter((item: any) => item.checked);
  tableData.value = tableList;
};
const handleGetData = (tableList: Host[]) => {
  selectTableData.value = [];
  tableData.value = tableList;
};
</script>
