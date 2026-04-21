<template>
  <!-- 基本信息 -->
  <div class="h-[44px] flex mb-[24px]">
    <!-- 管控单元数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.workUnitCount') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ stats.networkunit_count ?? '--' }}
      </div>
    </div>
    <!-- Proxy 数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.proxyQuantity') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ stats.proxy_count ?? 0 }}
      </div>
    </div>
    <!-- Agent 数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.agentQuantity') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ stats.agent_count ?? 0 }}
      </div>
    </div>
    <!-- 更新人 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updatePerson') }}
      </div>
      <div
        class="text-[12px] mt-[4px]"
        :class="{ 'auth-lock-text': !hasHistoryViewAuth }"
        @mouseenter="historyMouseEnter($event, !!hasHistoryViewAuth)"
        @mousemove="historyMouseMove($event, !!hasHistoryViewAuth)"
        @mouseleave="historyMouseLeave()"
        @click="!hasHistoryViewAuth && historyAuthClick($event)"
      >
        {{ eventInfo.operator || '--' }}<span v-if="!hasHistoryViewAuth" class="auth-asterisk">*</span>
      </div>
    </div>
    <!-- 更新时间 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updateTime') }}
      </div>
      <div
        class="text-[12px] mt-[4px]"
        :class="{ 'auth-lock-text': !hasHistoryViewAuth }"
        @mouseenter="historyMouseEnter($event, !!hasHistoryViewAuth)"
        @mousemove="historyMouseMove($event, !!hasHistoryViewAuth)"
        @mouseleave="historyMouseLeave()"
        @click="!hasHistoryViewAuth && historyAuthClick($event)"
      >
        {{ eventInfo.operate_time > 0 ? formatTimestamp(eventInfo.operate_time) : '--' }}<span v-if="!hasHistoryViewAuth" class="auth-asterisk">*</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, watch } from 'vue';
import { useRoute } from 'vue-router';

import { TopoService } from '@/api/modules/topo';
import { useAuthStore } from '@/stores/auth';
import useAuthLock from '@/composables/use-auth-lock';
import { formatTimestamp } from '@/common/util';

defineProps({
  workUnitCount: {
    type: Number,
    default: 0,
  },
});

const route = useRoute();
const authStore = useAuthStore();

const areaId = Number(route.params.workarea);

// networkarea_history_view 权限控制（更新人/更新时间的星号+锁 hover）
const {
  hasAuth: hasHistoryViewAuth,
  handleMouseEnter: historyMouseEnter,
  handleMouseMove: historyMouseMove,
  handleMouseLeave: historyMouseLeave,
  handleAuthClick: historyAuthClick,
} = useAuthLock(
  'networkarea_history_view',
  () => areaId,
  { resourceType: 'networkarea' },
);

/** 统计数据：从 /topo/networkarea/statistics 获取 */
const stats = reactive<{
  networkunit_count: number | null;
  proxy_count: number | null;
  agent_count: number | null;
}>({
  networkunit_count: null,
  proxy_count: null,
  agent_count: null,
});

/** 操作记录信息：从 /api/v3/topo/event/list 获取最新一条 */
const eventInfo = reactive<{
  operator: string;
  operate_time: number;
}>({
  operator: '',
  operate_time: 0,
});

/** 调 statistics 接口获取三个计数 */
const fetchStatistics = async () => {
  const result = await TopoService.NetworkAreaStatistics({
    bk_networkarea_id: [areaId],
  }).catch(() => ({ items: [] }));
  const item = (result?.items || []).find(i => i.bk_networkarea_id === areaId);
  if (item) {
    stats.networkunit_count = item.networkunit_count ?? null;
    stats.proxy_count = item.proxy_count ?? null;
    stats.agent_count = item.agent_count ?? null;
  }
};

/** 调 event/list 接口获取最新操作人/时间（只取第一条） */
const fetchLatestEvent = async () => {
  const result = await TopoService.EventList({
    page: { offset: 0, limit: 1 },
    only_count: false,
    exact_include_conditions: {
      bk_networkarea_id: [areaId],
    },
    fuzzy_include_conditions: {},
    operate_time_range: null,
  }).catch(() => ({ total: 0, items: [] }));
  const latest = (result.items || [])[0];
  if (latest) {
    eventInfo.operator = latest.operator || '';
    eventInfo.operate_time = latest.operate_time || 0;
  }
};

onMounted(async () => {
  // statistics 不依赖 history_view 权限，直接请求
  await fetchStatistics();
});

// 等 authorized 加载完再决定是否请求 EventList
watch(
  () => authStore.authorizedLoaded,
  async (loaded) => {
    if (loaded && hasHistoryViewAuth.value) {
      await fetchLatestEvent();
    }
  },
  { immediate: true },
);
</script>

<style lang="postcss" scoped>
.auth-lock-text {
  color: #c4c6cc;
  cursor: pointer;
}

.auth-asterisk {
  color: #c4c6cc;
  margin-left: 2px;
}
</style>
