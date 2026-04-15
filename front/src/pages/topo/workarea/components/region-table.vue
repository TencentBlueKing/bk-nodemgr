<template>
  <Loading :loading="workareaStore.loading">
    <Table
      class="mt-[16px] w-full"
      ref="tableRef"
      :data="tableData"
      :empty-text="$t('table.empty')"
      :pagination="pagination"
      :sort-config="sortConfig"
      :show-settings="isShowSetting"
      :settings="settings"
      :max-height="maxHeight"
      @page-limit-change="pageLimitChange"
      @page-value-change="pageValueChange"
      @setting-change="handleSettingChange"
      @column-filter="handleColumnFilter">
      <TableColumn
        :label="$t('topoManager.workArea.table.workareaName')"
        field="bk_networkarea_name"
        show-overflow="tooltip"
        fixed="left"
        :min-width="280">
        <template #default="{ row }">
          <Button
            v-if="row.bk_networkarea_id === 0"
            text
            class="mr-[12px]">
            <i class="nodeman-icon nc-collect text-[#C4C6CC] text-[18px]">
            </i>
          </Button>
          <Button
            v-else
            text
            class="mr-[12px]"
            @click.stop="handleCollect(row.bk_networkarea_id)">
            <i
              class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
              v-if="collectList.includes(row.bk_networkarea_id)">
            </i>
            <i
              class="nodeman-icon nc-not-favorited text-[#C4C6CC] text-[18px]"
              v-else>
            </i>
          </Button>
          <span
            v-if="!isAreaAuthorized(row.bk_networkarea_id)"
            class="text-[#C4C6CC] cursor-pointer"
            @click="handleAreaAuthClick($event, row.bk_networkarea_id, 'networkarea_view')"
            @mouseenter="viewMouseEnter($event, false)"
            @mousemove="viewMouseMove($event, false)"
            @mouseleave="viewMouseLeave()"
          >{{ row.bk_networkarea_name }}</span>
          <Button v-else theme="primary" text @click="handleToWorkareaDetail(row.bk_networkarea_id)">
            {{ row.bk_networkarea_name }}
          </Button>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.workareaId')"
        field="bk_networkarea_id"
        show-overflow="tooltip"
        :min-width="240">
        <template #default="{ row }">
          <span>
            {{ row.bk_networkarea_id || row.bk_networkarea_id === 0 ? `#${row.bk_networkarea_id}` : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.vendor')"
        field="cloud_vendor"
        show-overflow="tooltip"
        :filter="filterOption"
        :min-width="240">
        <template #default="{ row }">
          <span>
            {{ row.cloud_vendor && vendorMap[row.cloud_vendor] ?
              $t(
                vendorMap[row.cloud_vendor]?.label
              ) : '--'
            }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.workUnitsCount')"
        field="networkunit_count"
        show-overflow="tooltip"
        :min-width="240">
        <template #default="{ row }">
          <span>
            {{ !isNaN(row.networkunit_count) ? row.networkunit_count : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.proxyCount')"
        field="proxy_count"
        show-overflow="tooltip"
        :min-width="180">
        <template #default="{ row }">
          <span>
            {{ !isNaN(row.proxy_count) ? row.proxy_count : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.agentCount')"
        field="agent_count"
        show-overflow="tooltip"
        :min-width="180">
        <template #default="{ row }">
          <span>
            {{ !isNaN(row.agent_count) ? row.agent_count : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('table.action')"
        field="action"
        show-overflow="tooltip"
        fixed="right"
        :min-width="140">
        <template #default="{ row }">
          <div class="flex">
            <Button
              theme="primary"
              text
              class="mr-[12px]"
              :class="{ 'unAuthorized': !isRowEditAuth(row.bk_networkarea_id) }"
              @click="isRowEditAuth(row.bk_networkarea_id) ? handleEditWorkarea(row) : handleAreaAuthClick($event, row.bk_networkarea_id, 'networkarea_edit')"
              @mouseenter="editMouseEnter($event, isRowEditAuth(row.bk_networkarea_id))"
              @mousemove="editMouseMove($event, isRowEditAuth(row.bk_networkarea_id))"
              @mouseleave="editMouseLeave()"
            >
              {{ $t('action.edit') }}
            </Button>
            <Button
              theme="primary"
              text
              :class="{ 'unAuthorized': !isRowDeleteAuth(row.bk_networkarea_id) }"
              @click="isRowDeleteAuth(row.bk_networkarea_id) ? handleDeleteWorkarea(row.bk_networkarea_id) : handleAreaAuthClick($event, row.bk_networkarea_id, 'networkarea_delete')"
              @mouseenter="deleteMouseEnter($event, isRowDeleteAuth(row.bk_networkarea_id))"
              @mousemove="deleteMouseMove($event, isRowDeleteAuth(row.bk_networkarea_id))"
              @mouseleave="deleteMouseLeave()"
            >
              {{ $t('action.delete') }}
            </Button>
          </div>
        </template>
      </TableColumn>
    </Table>
  </Loading>
</template>

<script lang="ts" setup>
import { Button, InfoBox, Loading } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { vendorMap } from '../vendorMap';

import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import type { INetWorkArea } from '@/stores/workarea';
import { useWorkareaStore } from '@/stores/workarea';

interface IProps {
  list: INetWorkArea[],
  vendorList: string[],
}
const props = defineProps<IProps>();

const emit = defineEmits(['edit', 'filter']);

const tableData = ref(props.list);
const { t } = useI18n();
const router = useRouter();
const workareaStore = useWorkareaStore();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const sortConfig = ref({ multiple: true });

// networkarea_view hover lock for name column
const { handleMouseEnter: viewMouseEnter, handleMouseMove: viewMouseMove, handleMouseLeave: viewMouseLeave } = useAuthLock(
  'networkarea_view', () => undefined, { resourceType: 'networkarea' },
);

// networkarea_edit / networkarea_delete permissions (hover lock)
const { handleMouseEnter: editMouseEnter, handleMouseMove: editMouseMove, handleMouseLeave: editMouseLeave } = useAuthLock(
  'networkarea_edit', () => undefined, { resourceType: 'networkarea' },
);
const { handleMouseEnter: deleteMouseEnter, handleMouseMove: deleteMouseMove, handleMouseLeave: deleteMouseLeave } = useAuthLock(
  'networkarea_delete', () => undefined, { resourceType: 'networkarea' },
);

/** 判断某个区域是否有 view 权限 */
function isAreaAuthorized(areaId: number): boolean {
  if (!authStore.authorizedLoaded) return true;
  return authStore.hasAuthorizedResource('networkarea_view', areaId);
}

/** 按行判断编辑权限（需要 view + edit） */
function isRowEditAuth(areaId: number): boolean {
  return isAreaAuthorized(areaId) && authStore.hasAuthorizedResource('networkarea_edit', areaId);
}

/** 按行判断删除权限（需要 view + delete） */
function isRowDeleteAuth(areaId: number): boolean {
  return isAreaAuthorized(areaId) && authStore.hasAuthorizedResource('networkarea_delete', areaId);
}

/** 无权限时点击触发 verify 申请 */
async function handleAreaAuthClick(e: MouseEvent, areaId: number, action: string) {
  e.stopPropagation();
  await authStore.batchVerify([
    { id: action, action, resourceType: 'networkarea', routes: [] },
  ], undefined, areaId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
}

// 分页 - 使用store中的前端分页配置
const pagination = computed(() => workareaStore.frontPagination);

const pageLimitChange = async (limit: number) => {
  workareaStore.frontPageConf.limit = limit;
  workareaStore.frontPageConf.current = 1; // 页码重置为1
  await workareaStore.handleFetchCurrentPageStatistics();
};

const pageValueChange = async (current: number) => {
  workareaStore.frontPageConf.current = current;
  await workareaStore.handleFetchCurrentPageStatistics();
};

// table setting逻辑
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_networkarea_name',
    'bk_networkarea_id',
    'cloud_vendor',
    'networkunit_count',
    'proxy_count',
    'agent_count',
    'action',
  ],
  disabled: ['action'],
}, 'topoMng-workarea');

// table filter逻辑
const filterOption = reactive<{
  list: {
    value: any
    text: string
  }[],
  checked: string[]
}>({
  list: [],
  checked: [],
});
const handleColumnFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  emit('filter', { checked, field });
};

// 收藏管控区域
const collectList = ref<number[]>(JSON.parse(localStorage.getItem('collect_workarea') || '[]'));
const handleCollect = (val: number) => {
  if (collectList.value.includes(val)) {
    collectList.value = collectList.value.filter(item => item !== val);
  } else {
    collectList.value.push(val);
  }
  tableData.value.sort((a: any, b: any) => {
    // bk_networkarea_id为0的始终排在最前面
    if (a.bk_networkarea_id === 0) return -1;
    if (b.bk_networkarea_id === 0) return 1;
    const aIsFavorite = collectList.value.includes(a.bk_networkarea_id);
    const bIsFavorite = collectList.value.includes(b.bk_networkarea_id);
    if (aIsFavorite && bIsFavorite) return b.bk_networkarea_id - a.bk_networkarea_id;
    if (aIsFavorite && !bIsFavorite) return -1;
    if (!aIsFavorite && bIsFavorite) return 1;
    return b.bk_networkarea_id - a.bk_networkarea_id;
  });
  localStorage.setItem('collect_workarea', JSON.stringify(collectList.value));
  // 同步更新store中的收藏状态
  workareaStore.syncFavoriteWorkareaList();
};

// 改变includeConditions 重新请求table data
// const handleFilter = (currentChecked: string[]) => {
//   if (currentChecked.length === 0) {
//     tableData.value = props.list;
//     return;
//   }
//   tableData.value = tableData.value.filter(item => currentChecked.includes(parseInt(item.cloud_vendor)));
// };

// table height逻辑
// 52(顶部导航)+52(二级导航)+24*2(content-padding)+32(flex-row)+16(table-margin-top)
const tableOffset = 200;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleToWorkareaDetail = (bk_networkarea_id: number) => {
  router.push({
    name: 'workareaDetail',
    params: {
      workarea: bk_networkarea_id,
    },
  });
};

const handleEditWorkarea = (workareaData: NetworkArea) => {
  emit('edit', workareaData);
};

// todo
// 可能需要补充交互(message/重置筛选/重置pagination)
const handleDeleteWorkarea = (bk_networkarea_id: number) => {
  const workareaData = props.list.find(item => item.bk_networkarea_id === bk_networkarea_id);
  InfoBox({
    title: t('topoManager.workArea.delete.title', { x: workareaData.bk_networkarea_name || '' }),
    cancelText: t('action.cancel'),
    onConfirm() {
      workareaStore.handleDeleteWorkarea(bk_networkarea_id);
      workareaStore.handleFetchWorkareaList();
    },
  });
};

watch(() => props.vendorList, () => {
  filterOption.list = props.vendorList.map((item, index) => ({
    value: index,
    text: t(vendorMap[item]?.label || ''),
  }));
});

watch(() => props.list, () => {
  tableData.value = props.list;
}, { immediate: true });

const tableRef = ref();
defineExpose({
  tableRef,
});

</script>

<style lang="postcss">
.unAuthorized {
  color: #C4C6CC !important;
}
</style>
