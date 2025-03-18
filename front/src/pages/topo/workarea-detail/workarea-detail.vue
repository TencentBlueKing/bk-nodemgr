<template>
  <div class="p-[24px]">
    <WorkUnitInfo></WorkUnitInfo>
    <!-- 管控单元面板 -->
    <Tab v-model:active="active" type="card-grid">
      <Tab.TabPanel
        v-for="item in mockUnit"
        :key="item.name"
        :label="item.label"
        :name="item.name"
      >
        <div>
          <!-- 上下游接入点信息 -->
          <FlexRow class="mb-[30px] !items-start">
            <template #left>
              <AccessPoint
                :upstream-data="['默认区域', 'default', '内网接入点']"
                :downstream-data="[
                  {
                    name: '内网接入点',
                    Cluster: '10.0.0.1',
                    File: '10.0.0.2',
                    Data: '10.0.0.3',
                  },
                  {
                    name: '外网接入点',
                    Cluster: '10.0.0.1',
                    File: '10.0.0.2',
                    Data: '10.0.0.3',
                  },
                ]">
              </AccessPoint>
            </template>
            <template #right>
              <Button theme="primary" outline>
                <i class="nodeman-icon nc-icon-edit-2 mr-[4px]"></i>
                <span>{{ $t('action.edit') }}</span>
              </Button>
            </template>
          </FlexRow>
          <Divider type="solid"></Divider>
          <FlexRow class="mt-[24px]">
            <template #left>
              <div class="flex items-center">
                <!-- 新建 -->
                <Button theme="primary" class="mr-[8px]">
                  <span>{{ $t('topoManager.workAreaDetail.button.installProxy') }}</span>
                </Button>
                <Button disabled class="mr-[8px]">
                  <span>{{ $t('topoManager.workAreaDetail.button.batch') }}</span>
                  <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
                </Button>
                <Button>
                  <span>{{ $t('action.copy') }}</span>
                  <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
                </Button>
                <!-- 复制 -->
                <!-- <CopyIp @change="handleCopyChange"></CopyIp> -->
              </div>
            </template>
            <template #right>
              <SearchSelect
                class="w-[480px]"
                :placeholder="$t('topoManager.workAreaDetail.searchSelect.placeholder')"
                v-model.trim="searchKey"
                :data="searchSelectData">
              </SearchSelect>
            </template>
          </FlexRow>
          <!-- table -->
          <DetailTable></DetailTable>
        </div>
      </Tab.TabPanel>
      <template #add>
        <div
          class="rounded-[50px] hover:bg-[#EAEBF0] cursor-pointer
          h-[32px] w-[32px] text-center leading-[32px] ml-[-10px]">
          <i class="nodeman-icon nc-plus-line hover:text-[#979BA5] text-[#979BA5]"></i>
        </div>
      </template>
    </Tab>
  </div>
</template>

<script lang="ts" setup>
import { Button, Divider, SearchSelect, Tab } from 'bkui-vue';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { useDebounce } from '@vueuse/core';

import AccessPoint from './components/access-point.vue';
import DetailTable from './components/detail-table.vue';
import WorkUnitInfo from './components/work-unit-info.vue';

import { TopoService } from '@/api/modules/topo';

const { t } = useI18n();
const active = ref('OSS');
const mockUnit = ref([
  {
    label: 'default',
    name: 'default',
  },
  {
    label: 'DevCloud',
    name: 'DevCloud',
  },
  {
    label: 'OSS',
    name: 'OSS',
  },
]);

const handleAddWorkunit = () => {

};

// 搜索
const searchKey = ref([]);
const debounceSearch = useDebounce(searchKey, 300);

const searchSelectData = ref([
  {
    name: t('topoManager.workAreaDetail.table.ipv4'),
    id: 1,
  },
  {
    name: t('topoManager.workAreaDetail.table.ipv6'),
    id: 2,
  },
  {
    name: 'AgentID',
    id: 3,
  },
]);


const route = useRoute();
const workAreaId = Number(route.params.workarea);

const workUnitList = ref([]);
const fetchWorkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: [workAreaId],
    },
  });
  console.log(res);
  workUnitList.value = res.items;
};

const fetchWorkUnit = async () => {
  const res = await TopoService.NetworkUnitGet({
    bk_networkunit_id: 0,
  });
};

onMounted(() => {
  fetchWorkUnitList();
});

</script>

<style lang="less" scoped>
:deep(.bk-tab-header-operation > .bk-tab-header-item) {
  background-color: unset !important;
}
</style>
