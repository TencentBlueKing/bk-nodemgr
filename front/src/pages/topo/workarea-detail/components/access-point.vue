<template>
  <div class="text-[12px]">
    <!-- 收起状态 -->
    <div v-show="!isExpand">
      <!-- 上游接入点简化展示 -->
      <div class="mb-[14px] flex items-start" v-if="isShowUpstreamDefaultInfo">
        <div class="text-[#4D4F56] w-[72px] mr-[3px]">{{ $t('topoManager.workUnit.accessPoints.upstream') }} :</div>
        <div class="flex">
          <span>{{ clusterData.workarea_name }}</span>
          <span
            v-show="clusterData.workunit_name"
            class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] mt-[-3px] text-[24px]">
          </span>
          <span>{{ clusterData.workunit_name }}</span>
          <span
            v-show="clusterData.accesspoint_name"
            class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] mt-[-3px] text-[24px]">
          </span>
          <span>{{ clusterData.accesspoint_name }}</span>
        </div>
      </div>
      <!-- 上游接入点详细展示 -->
      <div class="mb-[14px] flex" v-else>
        <div class="text-[#4D4F56] w-[72px] mr-[3px]">
          {{ $t('topoManager.workUnit.accessPoints.upstream') }} :
        </div>
        <div>
          <div class="min-w-[295px] h-[112px] bg-[#F5F7FA] p-[14px] text-[#4D4F56]">
            <div class="flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">cluster :</div>
              <div class="flex items-center">
                <span>{{ clusterData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ clusterData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ clusterData.accesspoint_name }}</span>
              </div>
            </div>
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">file :</div>
              <div class="flex items-center">
                <span>{{ fileData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ fileData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ fileData.accesspoint_name }}</span>
              </div>
            </div>
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">data :</div>
              <div class="flex items-center">
                <span>{{ dataData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ dataData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ dataData.accesspoint_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- 下游接入点简化展示 -->
      <div class="flex items-start">
        <div class="text-[#4D4F56] w-[72px] mr-[3px]">
          {{ $t('topoManager.workUnit.accessPoints.downstream') }} :
        </div>
        <span v-if="isDownStreamDataExist">
          {{ downstreamData?.map(item => item.accesspoint_name).join(' , ') }}
        </span>
        <span v-else>--</span>
        <Button
          text
          theme="primary"
          class="ml-[15px]"
          @click="toggleExpand">
          <span>
            {{ $t('topoManager.workUnit.toggle.expand') }}
          </span>
          <i class="nodeman-icon nc-angle-double-down text-[24px]"></i>
        </Button>
      </div>
    </div>
    <!-- 展开状态 -->
    <div v-show="isExpand">
      <!-- 上游接入点详细展示 -->
      <div class="mb-[14px] flex">
        <div class="text-[#4D4F56] w-[72px] mr-[3px]">
          {{ $t('topoManager.workUnit.accessPoints.upstream') }} :
        </div>
        <div>
          <div class="min-w-[295px] h-[112px] bg-[#F5F7FA] p-[14px] text-[#4D4F56]">
            <div class="flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">cluster :</div>
              <div class="flex items-center">
                <span>{{ clusterData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ clusterData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ clusterData.accesspoint_name }}</span>
              </div>
            </div>
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">file :</div>
              <div class="flex items-center">
                <span>{{ fileData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ fileData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ fileData.accesspoint_name }}</span>
              </div>
            </div>
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">data :</div>
              <div class="flex items-center">
                <span>{{ dataData.workarea_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ dataData.workunit_name }}</span>
                <i class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
                <span>{{ dataData.accesspoint_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- 下游接入点详细展示 -->
      <div class="flex">
        <div class="text-[#4D4F56] w-[72px] mr-[3px]">
          {{ $t('topoManager.workUnit.accessPoints.downstream') }} :
        </div>
        <div v-if="isDownStreamDataExist">
          <div v-for="(item, index) in downstreamData" class="mb-[8px]" :key="index">
            <span>{{ item.accesspoint_name }}</span>
            <div class="min-w-[295px] h-[112px] bg-[#F5F7FA] p-[14px] text-[#4D4F56] mt-[6px]">
              <div class="flex items-center h-[20px]">
                <div class="w-[45px] text-right mr-[8px]">cluster :</div>
                <span>{{ item.endpoints.cluster.join(' ;') }}</span>
              </div>
              <div class="flex items-center h-[20px] mt-[12px]">
                <div class="w-[45px] text-right mr-[8px]">file :</div>
                <span>{{ item.endpoints.file.join(' ;') }}</span>
              </div>
              <div class="flex items-center h-[20px] mt-[12px]">
                <div class="w-[45px] text-right mr-[8px]">data :</div>
                <span>{{ item.endpoints.data.join(' ;') }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else>--</div>
        <div class="flex items-end">
          <Button
            text
            theme="primary"
            class="ml-[15px]"
            @click="toggleExpand">
            <span>
              {{ $t('topoManager.workUnit.toggle.close') }}
            </span>
            <i class="nodeman-icon nc-double-up text-[24px]"></i>
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Button } from 'bkui-vue';
import { isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, ref } from 'vue';

import { useWorkareaStore } from '@/stores/workarea';

const props = defineProps({
  upstreamData: {
    type: Object as PropType<Links>,
    required: true,
  },
  downstreamData: {
    type: Array as PropType<Array<AccessPoint>>,
    default: [],
  },
});
const workareaStore = useWorkareaStore();
const isDownStreamDataExist = computed(() => props.downstreamData.length !== 0);
const isExpand = ref(false);
const toggleExpand = () => {
  isExpand.value = !isExpand.value;
};

// 上游接入点的cluster/file/data是否一致
// 如一致简化展示
// 如不一致单独展开上游接入点,不再展示简化信息(不影响isExpand)
const isShowUpstreamDefaultInfo = computed(() => {
  const { cluster, file, data } = props.upstreamData;
  const equal = isEqual(cluster, file) && isEqual(file, data);
  return equal;
});

const clusterData = computed(() => ({
  workarea_name: getWorkareaName(props.upstreamData.cluster?.bk_networkarea_id),
  workunit_name: getWorkUnitName(
    props.upstreamData.cluster?.bk_networkarea_id,
    props.upstreamData.cluster?.bk_networkunit_id,
  ),
  accesspoint_name: getAccessPointName(
    props.upstreamData.cluster?.bk_networkunit_id,
    props.upstreamData.cluster?.accesspoint_id,
  ),
}));

const fileData = computed(() => {
  if (isShowUpstreamDefaultInfo.value) return clusterData.value;
  return {
    workarea_name: getWorkareaName(props.upstreamData.file.bk_networkarea_id),
    workunit_name: getWorkUnitName(
      props.upstreamData.file.bk_networkarea_id,
      props.upstreamData.file.bk_networkunit_id,
    ),
    accesspoint_name: getAccessPointName(
      props.upstreamData.file.bk_networkunit_id,
      props.upstreamData.file.accesspoint_id,
    ),
  };
});

const dataData = computed(() => {
  if (isShowUpstreamDefaultInfo.value) return clusterData.value;
  return {
    workarea_name: getWorkareaName(props.upstreamData.data.bk_networkarea_id),
    workunit_name: getWorkUnitName(
      props.upstreamData.data.bk_networkarea_id,
      props.upstreamData.data.bk_networkunit_id,
    ),
    accesspoint_name: getAccessPointName(
      props.upstreamData.data.bk_networkunit_id,
      props.upstreamData.data.accesspoint_id,
    ),
  };
});

// 根据缓存和workareaId获取workareaName
const getWorkareaName = (workareaId: number): string => workareaStore.allWorkareaList.get(workareaId)?.bk_networkarea_name || '';
// 根据缓存和workareaId, workUnitId获取workUnitName
const getWorkUnitName = (workareaId: number, workUnitId: number): string => workareaStore.allWorkUnitList
  .get(workareaId)?.find(unit => unit.bk_networkunit_id === workUnitId)?.bk_networkunit_name || '';
// 根据缓存和workUnitId, accessPointId获取accessPointName
const getAccessPointName = (workUnitId: number, accessPointId: number): string => workareaStore.allAccessPointList
  .get(workUnitId)?.find(point => point.accesspoint_id === accessPointId)?.accesspoint_name || '';


</script>
