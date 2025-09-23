<template>
  <Loading :loading="contentLoading">
    <div class="p-[24px]">
      <WorkUnitInfo
        :work-unit-count="workUnitList.length"
      ></WorkUnitInfo>
      <!-- 管控单元面板 -->
      <Tab v-model:active="active" type="card-grid">
        <Tab.TabPanel
          v-for="item in workUnitList"
          :key="item.bk_networkunit_id"
          :label="item.bk_networkunit_name"
          :name="item.bk_networkunit_id"
          render-directive="if"
        >
          <template #label>
            <div class="w-full">
              <span>{{ item.bk_networkunit_name }}</span>
              <span v-if="item.is_direct" class="text-[10px] ml-[5px]">
                {{ $t('topoManager.workAreaDetail.tab.direct') }}
              </span>
            </div>
          </template>
          <div>
            <!-- 上下游接入点信息 -->
            <FlexRow class="mb-[30px] !items-start">
              <template #left>
                <AccessPoint
                  :upstream-data="curWorkUnit.links"
                  :downstream-data="curWorkUnit.accesspoints"
                  :is_direct="curWorkUnit.is_direct"
                  :direct-endpoints="curWorkUnit.direct_endpoints">
                </AccessPoint>
              </template>
              <template #right>
                <Button theme="primary" class="mr-[10px]" outline @click="handleEditWorkUnit">
                  <i class="nodeman-icon nc-icon-edit-2 mr-[4px]"></i>
                  <span>{{ $t('action.edit') }}</span>
                </Button>
                <Button outline class="w-[32px] h-[32px] hover:border-[#EA3636]" @click="isShowDelete = true">
                  <i class="nodeman-icon nc-delete-3 hover:text-[#EA3636]"></i>
                </Button>
              </template>
            </FlexRow>
            <template v-if="!curWorkUnit.is_direct">
              <Divider type="solid"></Divider>
              <proxy-info :active="active"></proxy-info>
            </template>
          </div>
        </Tab.TabPanel>
        <template #add>
          <div
            class="rounded-[50px] hover:bg-[#EAEBF0] cursor-pointer
            h-[32px] w-[32px] text-center leading-[32px] ml-[-10px]"
            @click="handleCreateWorkUnit">
            <i class="nodeman-icon nc-plus-line hover:text-[#979BA5] text-[#979BA5]"></i>
          </div>
        </template>
        <template v-if="workUnitList.length === 0">
          <Exception
            :description="$t('topoManager.workUnit.empty')"
            scene="part"
            type="empty"
          />
        </template>
      </Tab>
      <UpsertWorkUnit
        v-model:is-show="isShow"
        :work-unit-id="active"
        :is-create="isCreate"
        @save="handleWorkUnitSave"
      />
      <DeleteWorkUnit
        v-model:is-show="isShowDelete"
        :work-unit-name="curWorkUnit?.bk_networkunit_name"
        :work-unit-id="curWorkUnit?.bk_networkunit_id"
        @delete="handleAfterDelete"
      />
    </div>
  </Loading>
</template>

<script lang="ts" setup>
import { Button, Divider, Exception, Loading, SearchSelect, Tab } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import UpsertWorkUnit from '../upsert-workunit/upsert-work-unit.vue';

import AccessPoint from './components/access-point.vue';
import DeleteWorkUnit from './components/delete-work-unit.vue';
import proxyInfo from './components/proxy-info.vue';
import WorkUnitInfo from './components/work-unit-info.vue';

import { useRouteSubTitle } from '@/stores/route-sub-title';
import { useWorkareaStore } from '@/stores/workarea';

const {
  handleFetchAllWorkarea,
  handleFetchAllWorkUnit,
} = useWorkareaStore();
const workareaStore = useWorkareaStore();
const routeSubTitle = useRouteSubTitle();

const { t } = useI18n();

const route = useRoute();
const workAreaId = Number(route.params.workarea);

const isShow = ref(false);
const isCreate = ref(false);
const active  = ref();
const contentLoading = ref(false);

const handleCreateWorkUnit = () => {
  isShow.value = true;
  isCreate.value = true;
};

const handleEditWorkUnit = () => {
  isShow.value = true;
  isCreate.value = false;
};

const isShowDelete = ref(false);

const curWorkarea = ref<NetworkArea>();
const workUnitList = ref<NetworkUnit[]>([]);

// eslint-disable-next-line max-len
const curWorkUnit = computed(() => workUnitList.value.find(unit => unit.bk_networkunit_id === active.value) as NetworkUnit);

const handleAfterDelete = async () => {
  try {
    // 更新管控单元数据
    await handleFetchAllWorkUnit();
    // 更新tabList
    workUnitList.value = workareaStore.allWorkUnitList.get(workAreaId) || [];
    // 更新active
    active.value = workUnitList.value[0]?.bk_networkarea_id;
    contentLoading.value = true;
  } catch (err) {
    console.error(err);
  } finally {
    contentLoading.value = false;
  }
};

const handleWorkUnitSave = async () => {
  contentLoading.value = true;
  try {
    // 更新管控单元数据
    await handleFetchAllWorkUnit();
    workUnitList.value = workareaStore.allWorkUnitList.get(workAreaId) || [];
  } catch (err) {
    console.error(err);
  } finally {
    contentLoading.value = false;
  }
};

const fetchData = async () => {
  contentLoading.value = true;
  try {
    await Promise.all([
      handleFetchAllWorkarea(),
      handleFetchAllWorkUnit(),
    ]);
    curWorkarea.value = workareaStore.allWorkareaList.get(workAreaId);
    workUnitList.value = workareaStore.allWorkUnitList.get(workAreaId) || [];
  } catch (err) {
    console.error(err);
  } finally {
    contentLoading.value = false;
  }
};

const initData = async () => {
  await fetchData();
  // 初始化副标题
  const subTitle = `${curWorkarea.value?.bk_networkarea_name}-#${curWorkarea.value?.bk_networkarea_id}`;
  routeSubTitle.subTitle = subTitle;

  // 初始化 tab焦点
  if (!route.params?.workUnit) {
    active.value = workUnitList.value[0]?.bk_networkarea_id;
  } else {
    active.value = Number(route.params.workUnit);
  }
};

onMounted(() => {
  initData();
});

</script>

<style lang="less" scoped>
:deep(.bk-tab-header-operation > .bk-tab-header-item) {
  background-color: unset !important;
}
</style>
