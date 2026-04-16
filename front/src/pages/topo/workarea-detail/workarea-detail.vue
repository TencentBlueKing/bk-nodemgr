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
import { getModuleAuthorizedItems } from '@/constants/auth';

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
    active.value = bk_networkunit_id;
    await loadUnitDetail();
  } catch (err) {
    console.error(err);
  } finally {
    contentLoading.value = false;
  }
};

/** 获取管控单元列表（仅 Brief，详情按需加载当前 active 单元） */
const fetchWorkUnits = async () => {
  const briefResult = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: {
      bk_networkarea_id: [workAreaId],
    },
  }).catch(() => ({ total: 0, items: [] }));
  const briefItems = (briefResult?.items || []) as NetworkUnitBrief[];

  workUnitList.value = briefItems.map(brief => ({
    ...brief,
    accesspoints: [],
    direct_endpoints: { cluster: [], file: [], data: [] },
    custom_deploy_config: {},
  } as NetworkUnitDetail));
};

/** 加载当前 active 单元的完整详情（含接入点） */
let loadingUnitId: number | null = null;
const loadUnitDetail = async () => {
  const unitId = active.value;
  console.log('[loadUnitDetail] called, active =', unitId, 'workUnitList length =', workUnitList.value.length);
  if (unitId == null) return;
  // 防重复：如果正在加载同一个 unitId，跳过
  if (loadingUnitId === unitId) {
    console.log('[loadUnitDetail] skip, already loading unitId =', unitId);
    return;
  }
  loadingUnitId = unitId;
  const detail = await TopoService.NetworkUnitGet({ bk_networkunit_id: unitId }).catch(() => null);
  console.log('[loadUnitDetail] NetworkUnitGet result, unitId =', unitId, 'detail =', !!detail, detail);
  if (detail) {
    // 兼容：get 返回 accesspoints 为空时，补充调用 accesspoint/list
    const hasAccessPoints = detail.accesspoints && detail.accesspoints.length > 0;
    if (!hasAccessPoints && detail.links) {
      // 从当前单元的 links 中提取所有关联的 accesspoint_id
      const linkedApIds = new Set<number>();
      for (const key of ['cluster', 'file', 'data'] as const) {
        const link = detail.links[key];
        if (link?.accesspoint_id != null) {
          linkedApIds.add(link.accesspoint_id);
        }
      }
      if (linkedApIds.size > 0) {
        const apResult = await TopoService.AccessPointList({
          page: { offset: 0, limit: 1 },
          only_count: false,
          exact_include_conditions: { bk_networkarea_id: [workAreaId], accesspoint_id: [...linkedApIds] },
        }).catch(() => null);
        if (apResult?.items) {
          detail.accesspoints = apResult.items as AccessPoint[];
        }
      } else {
        detail.accesspoints = [];
      }
      console.log('[loadUnitDetail] accesspoint/list 补充结果:', detail.accesspoints?.length || 0);
    }
    const idx = workUnitList.value.findIndex(u => u.bk_networkunit_id === unitId);
    if (idx !== -1) {
      workUnitList.value.splice(idx, 1, detail as NetworkUnitDetail);
    }
  }
  loadingUnitId = null;
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
    console.log('[initData] set active from sortedList:', sortedWorkUnitList.value[0]?.bk_networkunit_id);
  } else {
    active.value = Number(route.params.workUnit);
    console.log('[initData] set active from route:', Number(route.params.workUnit));
  }
  // 加载初始 active 单元的详情
  console.log('[initData] about to call loadUnitDetail, active =', active.value);
  await loadUnitDetail();
};

// 当排序列表变化时（如权限加载完成后重排），确保 active 仍指向有效的 panel
watch(sortedWorkUnitList, (list) => {
  console.log('[watch(sortedWorkUnitList)] triggered, list.length =', list.length, 'current active =', active.value);
  if (list.length > 0 && !list.some(item => item.bk_networkunit_id === active.value)) {
    console.log('[watch(sortedWorkUnitList)] setting active =', list[0].bk_networkunit_id);
    active.value = list[0].bk_networkunit_id;
  }
});

// 切换 tab 时加载当前单元详情
watch(active, () => {
  loadUnitDetail();
});

onMounted(async () => {
  // 确保 topoManager 模块权限数据已加载（页面刷新直接访问时可能未加载）
  const topoItems = getModuleAuthorizedItems('topoManager');
  await authStore.fetchAuthorized(topoItems, 'topoManager').catch(() => {});
  // 单独请求安装 Proxy 权限（networkunit_use_for_proxy），
  // 因后端 starts_with 兼容问题需独立调用，store 层已做降级处理
  authStore.fetchAuthorized([
    { action: 'networkunit_use_for_proxy', resource_type: 'networkunit' },
  ]);
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
