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
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import { useWorkareaStore } from '@/stores/workarea';
import { TopoService } from '@/api/modules/topo';
import useAuthLock from '@/composables/use-auth-lock';
import { getModuleAuthorizedItems } from '@/constants/auth';

const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const routeSubTitle = useRouteSubTitle();
const workareaStore = useWorkareaStore();

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
  // 把本区域管控单元写入 store，供 access-point 组件反查单元名字
  workareaStore.allWorkUnitList.set(workAreaId, workUnitList.value);
};

/** 加载当前 active 单元的完整详情（含接入点） */
let loadingUnitId: number | null = null;
const loadUnitDetail = async () => {
  const unitId = active.value;
  if (unitId == null) return;
  // 防重复：如果正在加载同一个 unitId，跳过
  if (loadingUnitId === unitId) return;
  loadingUnitId = unitId;
  const detail = await TopoService.NetworkUnitGet({ bk_networkunit_id: unitId }).catch(() => null);
  if (detail) {
    // 当前单元自身的接入点（下游展示用）
    if (detail.accesspoints?.length) {
      workareaStore.allAccessPointList.set(unitId, detail.accesspoints as AccessPoint[]);
    }
    // 上游接入点：access-point 组件按 bk_networkunit_id 反查 allAccessPointList 拿名字和 endpoints
    // 从 links 提取所有 (bk_networkunit_id, accesspoint_id) 对，按 bk_networkunit_id 分组 batch 拉
    if (detail.links) {
      const unitToApIds = new Map<number, Set<number>>();
      // 建立 accesspoint_id -> bk_networkunit_id 映射，避免依赖后端 AccessPoint 响应中的 bk_networkunit_id
      const apIdToUnitId = new Map<number, number>();
      for (const key of ['cluster', 'file', 'data'] as const) {
        const link = detail.links[key];
        if (link?.bk_networkunit_id != null && link.accesspoint_id != null) {
          // 跳过已经在 store 里并且包含目标 accesspoint_id 的情况，避免重复请求
          const cached = workareaStore.allAccessPointList.get(link.bk_networkunit_id);
          if (cached?.some(ap => ap.accesspoint_id === link.accesspoint_id)) continue;
          if (!unitToApIds.has(link.bk_networkunit_id)) {
            unitToApIds.set(link.bk_networkunit_id, new Set());
          }
          unitToApIds.get(link.bk_networkunit_id)!.add(link.accesspoint_id);
          apIdToUnitId.set(link.accesspoint_id, link.bk_networkunit_id);
        }
      }
      if (unitToApIds.size > 0) {
        const allApIds = [...unitToApIds.values()].reduce<number[]>((acc, set) => acc.concat([...set]), []);
        // 按 accesspoint_id batch 查，不带 bk_networkarea_id 限制（支持跨区域 upstream）
        const apResult = await TopoService.AccessPointList({
          page: { offset: 0, limit: allApIds.length },
          only_count: false,
          exact_include_conditions: { accesspoint_id: allApIds },
        }).catch(() => null);
        if (apResult?.items?.length) {
          // 按 bk_networkunit_id 分组塞进 allAccessPointList（与 access-point.vue 反查维度一致）
          const grouped = new Map<number, AccessPoint[]>();
          for (const ap of apResult.items as AccessPoint[]) {
            const unitKey = apIdToUnitId.get(ap.accesspoint_id);
            if (unitKey == null) continue;
            if (!grouped.has(unitKey)) grouped.set(unitKey, []);
            grouped.get(unitKey)!.push(ap);
          }
          // 合并进 store（保留已有的，追加新的，去重）
          for (const [unitKey, aps] of grouped) {
            const existed = workareaStore.allAccessPointList.get(unitKey) || [];
            const merged = [...existed];
            for (const ap of aps) {
              if (!merged.some(e => e.accesspoint_id === ap.accesspoint_id)) merged.push(ap);
            }
            workareaStore.allAccessPointList.set(unitKey, merged);
          }
        }
      }
      // 补充加载上游区域和单元名称，供 access-point 组件反查完整路径
      const upstreamAreaIds = new Set<number>();
      const upstreamUnitIds = new Set<number>();
      for (const key of ['cluster', 'file', 'data'] as const) {
        const link = detail.links[key];
        if (link?.bk_networkarea_id != null && !workareaStore.allWorkareaList.has(link.bk_networkarea_id)) {
          upstreamAreaIds.add(link.bk_networkarea_id);
        }
        if (link?.bk_networkunit_id != null) {
          const cached = workareaStore.allWorkUnitList.get(link.bk_networkarea_id);
          if (!cached?.some(u => u.bk_networkunit_id === link.bk_networkunit_id)) {
            upstreamUnitIds.add(link.bk_networkunit_id);
          }
        }
      }
      if (upstreamAreaIds.size > 0) {
        const areaRes = await TopoService.NetworkAreaList({
          page: { offset: 0, limit: upstreamAreaIds.size },
          only_count: false,
          exact_include_conditions: { bk_networkarea_id: Array.from(upstreamAreaIds) },
        }).catch(() => ({ items: [] }));
        for (const area of areaRes.items || []) {
          workareaStore.allWorkareaList.set(area.bk_networkarea_id, area);
        }
      }
      if (upstreamUnitIds.size > 0) {
        const unitRes = await TopoService.NetworkUnitListBrief({
          page: { offset: 0, limit: upstreamUnitIds.size },
          only_count: false,
          exact_include_conditions: { bk_networkunit_id: Array.from(upstreamUnitIds) },
        }).catch(() => ({ items: [] }));
        const grouped = new Map<number, any[]>();
        for (const unit of unitRes.items || []) {
          if (!grouped.has(unit.bk_networkarea_id)) grouped.set(unit.bk_networkarea_id, []);
          grouped.get(unit.bk_networkarea_id)!.push(unit);
        }
        for (const [areaId, units] of grouped) {
          const existed = workareaStore.allWorkUnitList.get(areaId) || [];
          const merged = [...existed];
          for (const unit of units) {
            if (!merged.some(u => u.bk_networkunit_id === unit.bk_networkunit_id)) merged.push(unit);
          }
          workareaStore.allWorkUnitList.set(areaId, merged);
        }
      }
    }
    const idx = workUnitList.value.findIndex(u => u.bk_networkunit_id === unitId);
    if (idx !== -1) {
      workUnitList.value.splice(idx, 1, detail as NetworkUnitDetail);
    }
  }
  loadingUnitId = null;
};

/** 获取当前管控区域信息（只查当前区域，不拉全量）
 *  同时把当前区域写入 workareaStore.allWorkareaList，供 access-point 组件反查区域名字
 */
const fetchCurrentWorkarea = async () => {
  const result = await TopoService.NetworkAreaList({
    page: { offset: 0, limit: 1 },
    only_count: false,
    exact_include_conditions: {
      bk_networkarea_id: [workAreaId],
      cloud_vendor: [],
    },
    fuzzy_include_conditions: {
      bk_networkarea_name: [],
    },
  }).catch(() => ({ total: 0, items: [] }));
  curWorkarea.value = (result?.items || [])[0];
  if (curWorkarea.value) {
    workareaStore.allWorkareaList.set(workAreaId, curWorkarea.value);
  }
};

const fetchData = async () => {
  contentLoading.value = true;
  try {
    await Promise.all([
      fetchCurrentWorkarea(),
      fetchWorkUnits(),
    ]);
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
  // 加载初始 active 单元的详情
  await loadUnitDetail();
};

// 当排序列表变化时（如权限加载完成后重排），确保 active 仍指向有效的 panel
watch(sortedWorkUnitList, (list) => {
  if (list.length > 0 && !list.some(item => item.bk_networkunit_id === active.value)) {
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
