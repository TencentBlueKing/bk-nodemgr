<template>
  <Loading :loading="contentLoading">
    <div class="p-[24px]">
      <WorkUnitInfo
        :work-unit-count="workUnitList.length"
      ></WorkUnitInfo>
      <!-- 管控单元面板 -->
      <Tab :active="active" type="card-grid" @change="handleTabChange">
        <Tab.TabPanel
          v-for="item in sortedWorkUnitList"
          :key="item.bk_networkunit_id"
          :label="item.bk_networkunit_name"
          :name="item.bk_networkunit_id"
          render-directive="if"
        >
          <template #label>
            <div
              class="w-full"
              :class="{ 'text-[#C4C6CC]': !isUnitAuthorized(item.bk_networkunit_id) }"
              @mouseenter="viewMouseEnter($event, isUnitAuthorized(item.bk_networkunit_id))"
              @mousemove="viewMouseMove($event, isUnitAuthorized(item.bk_networkunit_id))"
              @mouseleave="viewMouseLeave()"
            >
              <span>{{ item.bk_networkunit_name }}</span>
              <span v-if="item.is_direct" class="text-[10px] ml-[5px]">
                {{ $t('topoManager.workAreaDetail.tab.direct') }}
              </span>
            </div>
          </template>
          <!-- 有权限 tab 内容：正常展示 -->
          <div v-if="isUnitAuthorized(item.bk_networkunit_id)">
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
                <Button
                  theme="primary"
                  class="mr-[10px]"
                  outline
                  :class="{ 'unAuthorized': !hasUnitEditAuth }"
                  @click="hasUnitEditAuth ? handleEditWorkUnit() : editAuthClick($event)"
                  @mouseenter="editMouseEnter($event, hasUnitEditAuth)"
                  @mousemove="editMouseMove($event, hasUnitEditAuth)"
                  @mouseleave="editMouseLeave()"
                >
                  <i class="nodeman-icon nc-icon-edit-2 mr-[4px]"></i>
                  <span>{{ $t('action.edit') }}</span>
                </Button>
                <Button
                  outline
                  class="w-[32px] h-[32px]"
                  :class="{ 'unAuthorized': !hasUnitDeleteAuth }"
                  @click="hasUnitDeleteAuth ? (isShowDelete = true) : deleteAuthClick($event)"
                  @mouseenter="deleteMouseEnter($event, hasUnitDeleteAuth)"
                  @mousemove="deleteMouseMove($event, hasUnitDeleteAuth)"
                  @mouseleave="deleteMouseLeave()"
                >
                  <i class="nodeman-icon nc-delete-3" :class="{ 'hover:text-[#EA3636]': hasUnitDeleteAuth }"></i>
                </Button>
              </template>
            </FlexRow>
            <!-- proxy 表格信息-->
            <template v-if="!curWorkUnit.is_direct">
              <Divider type="solid"></Divider>
              <proxy-info :active="active"></proxy-info>
            </template>
          </div>
          <!-- 当前 active tab 无权限：显示 403 空状态（仅当所有单元都无权限时才会命中此分支） -->
          <div v-else>
            <Exception
              class="exception-wrap-item"
              scene="part"
              type="403"
            >
              <Button theme="primary" @click="handleUnitViewAuthClick($event, item.bk_networkunit_id)">
                {{ $t('components.permission.apply') }}
              </Button>
            </Exception>
          </div>
        </Tab.TabPanel>
        <template #add>
          <div
            class="rounded-[50px] hover:bg-[#EAEBF0] cursor-pointer
            h-[32px] w-[32px] text-center leading-[32px] ml-[-10px]"
            @click="hasUnitCreateAuth ? handleCreateWorkUnit() : createAuthClick($event)"
            @mouseenter="createMouseEnter($event, hasUnitCreateAuth)"
            @mousemove="createMouseMove($event, hasUnitCreateAuth)"
            @mouseleave="createMouseLeave()">
            <i class="nodeman-icon nc-plus-line hover:text-[#979BA5] text-[#979BA5]"></i>
          </div>
        </template>
        <!-- 没有任何管控单元时的空状态 -->
        <template v-if="workUnitList.length === 0">
          <Exception
            class="exception-wrap-item"
            :title="$t('topoManager.workUnit.empty')"
            :description="$t('topoManager.workUnit.emptyTips')"
            scene="part"
            type="empty"
          >
            <Button
              theme="primary"
              :class="{ 'unAuthorized': !hasUnitCreateAuth }"
              @click="hasUnitCreateAuth ? handleCreateWorkUnit() : createAuthClick($event)"
              @mouseenter="createMouseEnter($event, hasUnitCreateAuth)"
              @mousemove="createMouseMove($event, hasUnitCreateAuth)"
              @mouseleave="createMouseLeave()"
            >
              {{ $t('topoManager.workUnit.title.create') }}
            </Button>
          </Exception>
        </template>
      </Tab>
      <UpsertWorkUnit
        v-model:is-show="isShow"
        :work-unit-id="active ?? -1"
        :is-create="isCreate"
        @save="handleWorkUnitSave"
      />
      <DeleteWorkUnit
        v-model:is-show="isShowDelete"
        :work-unit-name="curWorkUnit?.bk_networkunit_name"
        :work-unit-id="curWorkUnit?.bk_networkunit_id"
        :is-direct="curWorkUnit?.is_direct"
        @delete="handleAfterDelete"
      />
    </div>
  </Loading>
</template>

<script lang="ts" setup>
import { Button, Divider, Exception, Loading, Tab } from 'bkui-vue';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import UpsertWorkUnit from '../upsert-workunit/upsert-work-unit.vue';

import AccessPoint from './components/access-point.vue';
import DeleteWorkUnit from './components/delete-work-unit.vue';
import proxyInfo from './components/proxy-info.vue';
import WorkUnitInfo from './components/work-unit-info.vue';

import { useRouteSubTitle } from '@/stores/route-sub-title';
import { useWorkareaStore } from '@/stores/workarea';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import { TopoService } from '@/api/modules/topo';
import useAuthLock from '@/composables/use-auth-lock';

const {
  handleFetchAllWorkarea,
} = useWorkareaStore();
const workareaStore = useWorkareaStore();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const routeSubTitle = useRouteSubTitle();

const { t } = useI18n();

const route = useRoute();
const workAreaId = Number(route.params.workarea);

// networkunit permissions (hover lock + click apply)
const { hasAuth: hasUnitCreateAuth, handleMouseEnter: createMouseEnter, handleMouseMove: createMouseMove, handleMouseLeave: createMouseLeave, handleAuthClick: createAuthClick } = useAuthLock(
  'networkunit_create', () => undefined, { resourceType: 'networkunit' },
);
const { hasAuth: hasUnitEditAuth, handleMouseEnter: editMouseEnter, handleMouseMove: editMouseMove, handleMouseLeave: editMouseLeave, handleAuthClick: editAuthClick } = useAuthLock(
  'networkunit_edit', () => active.value, { resourceType: 'networkunit' },
);
const { hasAuth: hasUnitDeleteAuth, handleMouseEnter: deleteMouseEnter, handleMouseMove: deleteMouseMove, handleMouseLeave: deleteMouseLeave, handleAuthClick: deleteAuthClick } = useAuthLock(
  'networkunit_delete', () => active.value, { resourceType: 'networkunit' },
);
// networkunit_view hover lock (for tab label)
const { handleMouseEnter: viewMouseEnter, handleMouseMove: viewMouseMove, handleMouseLeave: viewMouseLeave } = useAuthLock(
  'networkunit_view', () => undefined, { resourceType: 'networkunit' },
);

/** 判断某个单元是否有 view 权限 */
function isUnitAuthorized(unitId: number): boolean {
  if (!authStore.authorizedLoaded) return false; // 默认无权限
  return authStore.hasAuthorizedResource('networkunit_view', unitId);
}

/** 无权限 tab 点击 → 触发 networkunit_view 权限申请 */
async function handleUnitViewAuthClick(e: MouseEvent, unitId: number) {
  e.stopPropagation();
  await authStore.batchVerify([
    { id: 'networkunit_view', action: 'networkunit_view', resourceType: 'networkunit', routes: [] },
  ], undefined, unitId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
}

const isShow = ref(false);
const isCreate = ref(false);
const active  = ref();
const contentLoading = ref(false);

/** Tab 切换拦截：无权限 tab 不切换，直接触发权限申请 */
function handleTabChange(name: number) {
  if (!isUnitAuthorized(name)) {
    handleUnitViewAuthClick(new MouseEvent('click'), name);
    return;
  }
  active.value = name;
}

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
const workUnitList = ref<NetworkUnitDetail[]>([]);

/** 排序后的管控单元列表：有权限的排前面（id 升序），无权限的排后面（id 升序） */
const sortedWorkUnitList = computed(() => {
  return [...workUnitList.value].sort((a, b) => {
    const aAuth = isUnitAuthorized(a.bk_networkunit_id);
    const bAuth = isUnitAuthorized(b.bk_networkunit_id);
    if (aAuth !== bAuth) return aAuth ? -1 : 1;
    return a.bk_networkunit_id - b.bk_networkunit_id;
  });
});

// eslint-disable-next-line max-len
const curWorkUnit = computed(() => workUnitList.value.find(unit => unit.bk_networkunit_id === active.value) as NetworkUnitDetail);

const handleAfterDelete = async () => {
  try {
    await fetchWorkUnits();
    active.value = sortedWorkUnitList.value[0]?.bk_networkunit_id;
  } catch (err) {
    console.error(err);
  }
};

const handleWorkUnitSave = async (bk_networkunit_id: number) => {
  contentLoading.value = true;
  try {
    await fetchWorkUnits();
    setTimeout(() => {
      active.value = bk_networkunit_id;
    }, 0);
  } catch (err) {
    console.error(err);
  } finally {
    contentLoading.value = false;
  }
};

/** 获取管控单元数据：Brief（当前区域全部）+ Get（有权限的逐个获取详情） */
const fetchWorkUnits = async () => {
  // 1. Brief 接口获取当前区域所有单元基本信息（用于 tab 列表展示）
  const briefResult = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: {
      bk_networkarea_id: [workAreaId],
    },
  }).catch(() => ({ total: 0, items: [] }));
  const briefItems = (briefResult?.items || []) as NetworkUnitBrief[];

  // 2. 筛选有 networkunit_view 权限的单元 ID
  const authorizedUnitIds = briefItems
    .map(item => item.bk_networkunit_id)
    .filter(id => isUnitAuthorized(id));

  // 3. 对有权限的单元并行调用 NetworkUnitGet 获取完整数据（含接入点详情）
  const detailMap = new Map<number, NetworkUnitDetail>();
  if (authorizedUnitIds.length > 0) {
    const detailResults = await Promise.all(
      authorizedUnitIds.map(id =>
        TopoService.NetworkUnitGet({ bk_networkunit_id: id }).catch(() => null),
      ),
    );
    for (const detail of detailResults) {
      if (detail) {
        detailMap.set((detail as NetworkUnitDetail).bk_networkunit_id, detail as NetworkUnitDetail);
      }
    }
  }

  // 4. 合并：有权限用完整数据，无权限用 Brief 填充
  workUnitList.value = briefItems.map((brief) => {
    const detail = detailMap.get(brief.bk_networkunit_id);
    if (detail) return detail;
    return {
      ...brief,
      accesspoints: [],
      direct_endpoints: { cluster: [], file: [], data: [] },
      custom_deploy_config: {},
    } as NetworkUnitDetail;
  });
};

const fetchData = async () => {
  contentLoading.value = true;
  try {
    await Promise.all([
      handleFetchAllWorkarea(),
      fetchWorkUnits(),
    ]);
    curWorkarea.value = workareaStore.allWorkareaList.get(workAreaId);
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

  // 初始化 tab焦点：使用排序后的列表（有权限优先）
  if (!route.params?.workUnit) {
    active.value = sortedWorkUnitList.value[0]?.bk_networkunit_id;
  } else {
    active.value = Number(route.params.workUnit);
  }
};

// 当排序列表变化时（如权限加载完成后重排），确保 active 仍指向有效的 panel
watch(sortedWorkUnitList, (list) => {
  if (list.length > 0 && !list.some(item => item.bk_networkunit_id === active.value)) {
    active.value = list[0].bk_networkunit_id;
  }
});

onMounted(async () => {
  await initData();
});

</script>

<style lang="less" scoped>
:deep(.bk-tab-header-operation > .bk-tab-header-item) {
  background-color: unset !important;
}
</style>
<style lang="postcss">
.unAuthorized {
  color: #C4C6CC !important;
}
.exception-wrap-item {
  height: 435px;
  .bk-exception-img {
    width: 440px;
    height: 200px;
    img {
      width: 440px;
      height: 200px;
    }
  }
  .bk-exception-title {
    font-size: 24px;
    color: #313238;
    letter-spacing: 0;
    line-height: 32px;
  }
  .bk-exception-description {
    margin-top: 16px;
    font-size: 14px;
  }
  .bk-exception-footer {
    margin-top: 24px;
  }
}
</style>
