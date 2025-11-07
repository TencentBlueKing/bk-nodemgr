<template>
  <div class="text-[12px]">
    <div class="mb-[14px] flex items-start">
      <div class="text-[#4D4F56] w-[72px] mr-[3px]">{{ $t('topoManager.workUnit.accessPoints.upstream') }} :</div>
      <template v-if="is_direct">
        <Popover
          theme="light"
          trigger="hover"
          placement="right"
        >
          <div>
            <Tag theme="info">直连</Tag>
          </div>
          <template #content>
            <Table
              :data="upstreamTableData"
              :min-width="600"
              :maxHeight="800"
              auto-resize
              :show-overflow="false"
              :row-config="{ isHover: true, height: 'auto' }">
              <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                <template #default>
                  <Tag theme="info">直连</Tag>
                </template>
              </TableColumn>
              <TableColumn field="cluster" title="cluster" :min-width="200">
                <template #default>
                  <span>{{ directEndpoints.cluster.join('\n') }}</span>
                </template>
              </TableColumn>
              <TableColumn field="file" title="file" :min-width="200">
                <template #default>
                  <span>{{ directEndpoints.cluster.join('\n') }}</span>
                </template>
              </TableColumn>
              <TableColumn field="data" title="data" :min-width="200">
                <template #default>
                  <span>{{ directEndpoints.cluster.join('\n') }}</span>
                </template>
              </TableColumn>
            </Table>
          </template>
        </Popover>
      </template>
      <div v-else class="flex cursor-pointer">
        <div>
          <Popover
            theme="light"
            trigger="hover"
            placement="right"
          >
            <div class="flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">Cluster :</div>
              <div class="flex items-center">
                <span class="text-[#3a84ff]">{{ clusterData.workarea_name }}</span>
                <span v-show="clusterData.workunit_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ clusterData.workunit_name }}</span>
                <span v-show="clusterData.accesspoint_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ clusterData.accesspoint_name }}</span>
              </div>
            </div>
            <template #content>
              <Table
                :data="clusterData.accesspoint_name ? upstreamTableData : []"
                :min-width="600"
                :maxHeight="800"
                auto-resize
                :show-overflow="false"
                :row-config="{ isHover: true, height: 'auto' }">
                <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                  <template #default>
                    {{ clusterData.accesspoint_name }}
                  </template>
                </TableColumn>
                <TableColumn field="cluster" title="cluster" :min-width="200">
                  <template #default>
                    <span style="white-space: pre-line;">{{ clusterData.cluster?.join('\n') }}</span>
                  </template>
                </TableColumn>
              </Table>
            </template>
          </Popover>
          <Popover
            theme="light"
            trigger="hover"
            placement="right"
          >
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">File :</div>
              <div class="flex items-center">
                <span class="text-[#3a84ff]">{{ fileData.workarea_name }}</span>
                <span v-show="fileData.workunit_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ fileData.workunit_name }}</span>
                <span v-show="fileData.accesspoint_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ fileData.accesspoint_name }}</span>
              </div>
            </div>
            <template #content>
              <Table
                :data="fileData.accesspoint_name ? upstreamTableData : []"
                :min-width="600"
                :maxHeight="800"
                auto-resize
                :show-overflow="false"
                :row-config="{ isHover: true, height: 'auto' }">
                <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                  <template #default>
                    {{ fileData.accesspoint_name }}
                  </template>
                </TableColumn>
                <TableColumn field="file" title="file" :min-width="200">
                  <template #default>
                    <span>{{ fileData.file?.join('\n') }}</span>
                  </template>
                </TableColumn>
              </Table>
            </template>
          </Popover>
          <Popover
            theme="light"
            trigger="hover"
            placement="right"
          >
            <div class="mt-[12px] flex items-center h-[20px]">
              <div class="w-[45px] text-right mr-[8px]">Data :</div>
              <div class="flex items-center">
                <span class="text-[#3a84ff]">{{ dataData.workarea_name }}</span>
                <span v-show="dataData.workunit_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ dataData.workunit_name }}</span>
                <span v-show="dataData.accesspoint_name" class="mx-[3px]">/</span>
                <span class="text-[#3a84ff]">{{ dataData.accesspoint_name }}</span>
              </div>
            </div>
            <template #content>
              <Table
                :data="dataData.accesspoint_name ? upstreamTableData : []"
                :min-width="600"
                :maxHeight="800"
                auto-resize
                :show-overflow="false"
                :row-config="{ isHover: true, height: 'auto' }">
                <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                  <template #default>
                    {{ dataData.accesspoint_name }}
                  </template>
                </TableColumn>
                <TableColumn field="data" title="data" :min-width="200">
                  <template #default>
                    <span>{{ dataData.data?.join('\n') }}</span>
                  </template>
                </TableColumn>
              </Table>
            </template>
          </Popover>
        </div>
      </div>
    </div>
    <div class="flex items-center">
      <div class="text-[#4D4F56] w-[72px] mr-[3px] shrink-0">
        {{ $t('topoManager.workUnit.accessPoints.downstream') }} :
      </div>
      <span v-if="isDownStreamDataExist" class="responsive-container cursor-pointer text-[#3a84ff]">
        <Popover
          theme="light"
          trigger="hover"
          placement="right"
        >
          <div>
            {{ downstreamData?.map(item => item.accesspoint_name).join(', ') }}
          </div>
          <template #content>
            <Table
              :data="downstreamData"
              :min-width="600"
              :maxHeight="800"
              auto-resize
              :show-overflow="false"
              :row-config="{ isHover: true, height: 'auto' }"
            >
              <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                <template #default="{ row }">
                  {{ row.accesspoint_name }}
                </template>
              </TableColumn>
              <TableColumn field="cluster" title="cluster" :min-width="200">
                <template #default="{ row }">
                  <span style="white-space: pre-line;">{{ row.endpoints.cluster.join('\n') }}</span>
                </template>
              </TableColumn>
              <TableColumn field="file" title="file" :min-width="200">
                <template #default="{ row }">
                  {{ row.endpoints.file.join('\n') }}
                </template>
              </TableColumn>
              <TableColumn field="data" title="data" :min-width="200">
                <template #default="{ row }">
                  {{ row.endpoints.data.join('\n') }}
                </template>
              </TableColumn>
            </Table>
          </template>
        </Popover>
      </span>
      <span v-else>--</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Button, Tag, Popover } from 'bkui-vue';
import { isEqual } from 'lodash';
import type { PropType } from 'vue';
import { computed, ref } from 'vue';

import { Table, TableColumn } from '@blueking/table';

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
  is_direct: {
    type: Boolean,
    default: false,
  },
  directEndpoints: {
    type: Object as PropType<Endpoints>,
    default: null,
  },
});
const workareaStore = useWorkareaStore();
const isDownStreamDataExist = computed(() => props.downstreamData.length !== 0);

// 上游接入点的cluster/file/data是否一致
// 如一致简化展示
// 如不一致单独展开上游接入点,不再展示简化信息(不影响isExpand)
const isShowUpstreamDefaultInfo = computed(() => {
  const { cluster, file, data } = props.upstreamData;
  const equal = isEqual(cluster, file) && isEqual(file, data);
  return equal;
});

const clusterData = computed(() => {
  const endpoints = workareaStore.allAccessPointList.get(props.upstreamData.cluster?.bk_networkunit_id)
    ?.find(item => item.accesspoint_id === props.upstreamData.cluster?.accesspoint_id)
    ?.endpoints;
  const { cluster = [''], file = [''], data = [''] } = endpoints as Endpoints;
  return {
    workarea_name: getWorkareaName(props.upstreamData.cluster?.bk_networkarea_id),
    workunit_name: getWorkUnitName(
      props.upstreamData.cluster?.bk_networkarea_id,
      props.upstreamData.cluster?.bk_networkunit_id,
    ),
    accesspoint_name: getAccessPointName(
      props.upstreamData.cluster?.bk_networkunit_id,
      props.upstreamData.cluster?.accesspoint_id,
    ),
    cluster,
    file,
    data,
  };
});

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
    file: workareaStore.allAccessPointList.get(props.upstreamData.file?.bk_networkunit_id)
      ?.find(item => item.accesspoint_id === props.upstreamData.file?.accesspoint_id)
      ?.endpoints.file,
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
    data: workareaStore.allAccessPointList.get(props.upstreamData.data?.bk_networkunit_id)
      ?.find(item => item.accesspoint_id === props.upstreamData.data?.accesspoint_id)
      ?.endpoints.data,
  };
});

const upstreamTableData = ref([
  {
    cluster: '',
    file: '',
    data: '',
  },
]);
// 根据缓存和workareaId获取workareaName
const getWorkareaName = (workareaId: number): string => workareaStore.allWorkareaList.get(workareaId)?.bk_networkarea_name || '';
// 根据缓存和workareaId, workUnitId获取workUnitName
const getWorkUnitName = (workareaId: number, workUnitId: number): string => workareaStore.allWorkUnitList
  .get(workareaId)?.find(unit => unit.bk_networkunit_id === workUnitId)?.bk_networkunit_name || '';
// 根据缓存和workUnitId, accessPointId获取accessPointName
const getAccessPointName = (workUnitId: number, accessPointId: number): string => workareaStore.allAccessPointList
  .get(workUnitId)?.find(point => point.accesspoint_id === accessPointId)?.accesspoint_name || '';


</script>
<style scoped lang="postcss">
.responsive-container {
  display: flex;
  align-items: baseline;
  overflow-x: auto;
  max-width: min(700px, 70vw);
  width: 100%;
}

@media (min-width: 2001px) {
  .responsive-container {
    max-width: 70vw;
  }
}

@media (min-width: 1801px) and (max-width: 2000px) {
  .responsive-container {
    max-width: 67vw;
  }
}

@media (min-width: 1601px) and (max-width: 1800px) {
  .responsive-container {
    max-width: 65vw;
  }
}

@media (min-width: 1401px) and (max-width: 1600px) {
  .responsive-container {
    max-width: 63vw;
  }
}
@media (min-width: 1201px) and (max-width: 1400px) {
  .responsive-container {
    max-width: 61vw;
  }
}

@media (max-width: 1200px) {
  .responsive-container {
    max-width: 900px;
  }
}
</style>
