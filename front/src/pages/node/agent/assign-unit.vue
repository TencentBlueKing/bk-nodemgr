<template>
  <div class="setup pt-[24px] pb-[48px]">
    <div class="m-[24px]">
      <Form ref="formRef" :model="formData">
        <Form.FormItem
          :label="$t('platform.nodeMan.agentStatus.assignUnitHostInfo')"
          required
        >
          <Loading :loading="loading">
            <assign-unit-table
              ref="assignUnitTableRef"
              v-model:data="formData.info"
              :max-height="640"
              :disable-direct="false"
            ></assign-unit-table>
          </Loading>
        </Form.FormItem>
      </Form>
    </div>
    <div
      :class="[
        'h-[48px] w-full flex items-center pl-[174px]',
        { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom },
      ]"
      ref="footerRef"
    >
      <template v-if="hasAssignUnitAuth">
        <Button
          class="min-w-[120px] mr-[8px]"
          theme="primary"
          :loading="submitting"
          @click="handleConfirm"
        >
          <span>{{ $t('platform.nodeMan.agentStatus.assignUnitConfirmBtn', { count: formData.info.length }) }}</span>
        </Button>
      </template>
      <span
        v-else
        class="inline-flex items-center auth-lock-wrapper mr-[8px]"
        @click="handleAuthClick"
        @mouseenter="authLockMouseEnter($event, false)"
        @mousemove="authLockMouseMove($event, false)"
        @mouseleave="authLockMouseLeave()"
      >
        <Button theme="primary" class="auth-disabled-btn min-w-[120px]">
          <span>{{ $t('platform.nodeMan.agentStatus.assignUnitConfirmBtn', { count: formData.info.length }) }}</span>
        </Button>
      </span>
      <Button class="w-[88px]" @click="handleCancel">{{ $t("action.cancel") }}</Button>
    </div>

    <operate-dialog
      v-model:is-show="restartDialogShow"
      :title="restartDialogTitle"
      :type="'restart'"
      :sub-title="restartDialogSubTitle"
      @confirm="handleRestartConfirm"
    ></operate-dialog>
  </div>
</template>

<script lang="ts" setup>
import { Button, Form, InfoBox, Loading, Message } from 'bkui-vue';
import { debounce } from 'lodash';
import { h, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import type { NodeAgentAssignUnitRespData } from '@/@types/node_agent';
import { NodeAgentService } from '@/api/modules/node_agent';
import { TopoService } from '@/api/modules/topo';
import { UNASSIGNED_NETWORK_UNIT } from '@/common/const';
import { resolveLoginMode } from '@/common/util';
import AssignUnitTable from '@/components/assign-unit-table.vue';
import OperateDialog from '@/components/operate-dialog.vue';
import useAuthLock from '@/composables/use-auth-lock';
import { useNodeManageStore } from '@/stores/node-manage';

const { t } = useI18n();
const router = useRouter();
const nodeManageStore = useNodeManageStore();

// ===== networkunit_use_for_agent 权限控制（分配管控单元确认按钮） =====
const {
  hasAuth: hasAssignUnitAuth,
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
  handleAuthClick,
} = useAuthLock('networkunit_use_for_agent', () => undefined, { resourceType: 'networkunit' });

const formData = reactive({
  info: [] as any[],
});

const normalizeNetworkUnitId = (id: unknown) => {
  if (id === '' || id === null || id === undefined) return '';
  return Number(id) === UNASSIGNED_NETWORK_UNIT ? '' : String(id);
};

const loading = ref(false);
const submitting = ref(false);
const isAtBottom = ref(false);
const footerRef = ref<Element | null>(null);
const formRef = ref(null);
const assignUnitTableRef = ref<InstanceType<typeof AssignUnitTable> | null>(null);

const restartDialogShow = ref(false);
const restartDialogTitle = ref('');
const restartDialogSubTitle = ref('');
const assignedHostIds = ref<number[]>([]);

const handleCancel = () => {
  router.push({ name: 'agent' });
};

const handleConfirm = async () => {
  const valid = await assignUnitTableRef.value?.tableValidate();
  if (!valid) return;

  const grouped = new Map<number, number[]>();
  for (const row of formData.info) {
    const unitId = Number(row.bk_networkunit_id);
    if (unitId < 0 || isNaN(unitId) || row.bk_networkunit_id === '') continue;
    if (!grouped.has(unitId)) grouped.set(unitId, []);
    grouped.get(unitId)!.push(row.bk_host_id);
  }

  if (grouped.size === 0) {
    Message({ theme: 'warning', message: t('platform.nodeMan.agentStatus.assignUnitNoUnit') });
    return;
  }

  submitting.value = true;
  let totalSuccess = 0;
  let totalFailed = 0;
  const allFailedReasons: string[] = [];
  const successHostIds: number[] = [];

  try {
    for (const [unitId, hostIds] of grouped) {
      const result: NodeAgentAssignUnitRespData = await NodeAgentService.NodeAgentAssignUnit({
        bk_host_id: hostIds,
        bk_networkunit_id: unitId,
      });
      totalSuccess += result.success_count;
      totalFailed += result.failed_count;
      if (result.failed_reasons?.length) {
        allFailedReasons.push(...result.failed_reasons);
      }
      if (result.success_count > 0) {
        successHostIds.push(...hostIds);
      }
    }

    if (totalSuccess > 0) {
      assignedHostIds.value = successHostIds;

      InfoBox({
        title: t('platform.nodeMan.agentStatus.assignUnitSuccessTitle'),
        content: () => h('div', { style: 'font-size: 14px; line-height: 1.8;' }, [
          t('platform.nodeMan.agentStatus.assignUnitSuccessDescPrefix'),
          h('span', { style: 'color: #3fc06d; font-weight: bold;' }, totalSuccess),
          t('platform.nodeMan.agentStatus.assignUnitSuccessDescMid'),
          h('span', { style: 'color: #ea3636; font-weight: bold;' }, totalFailed),
          t('platform.nodeMan.agentStatus.assignUnitSuccessDescSuffix'),
        ]),
        confirmText: t('platform.nodeMan.agentStatus.restartNow'),
        cancelText: t('platform.nodeMan.agentStatus.handleLater'),
        onConfirm: () => {
          const firstIp = formData.info[0]?.bk_host_innerip || '';
          restartDialogTitle.value = t('platform.nodeMan.agentStatus.confirmIsBatch', {
            type: t('platform.nodeMan.agentStatus.restart'),
          });
          restartDialogSubTitle.value = t('platform.nodeMan.agentStatus.batchOperate', {
            type: t('platform.nodeMan.agentStatus.restart'),
            firstIp,
            num: assignedHostIds.value.length,
          });
          restartDialogShow.value = true;
        },
        onCancel: () => {
          router.push({ name: 'agent' });
        },
      });
    } else {
      Message({
        theme: 'error',
        message: allFailedReasons.join('; ') || t('platform.nodeMan.agentStatus.assignUnitAllFailed'),
      });
    }
  } catch (err: any) {
    Message({ theme: 'error', message: err.message || String(err) });
  } finally {
    submitting.value = false;
  }
};

const restartInProgress = ref(false);

const handleRestartConfirm = async (extraData: any = {}) => {
  restartInProgress.value = true;
  loading.value = true;
  const params = {
    host: assignedHostIds.value.map((id: number) => ({
      bk_host_id: id,
      force: extraData.isForce,
      graceful_restart_timeout_sec: extraData.time,
    })),
  };
  let result;
  if (extraData.isReconfig) {
    result = await NodeAgentService.NodeAgentReconfig(params).catch(() => ({ workflow_id: '' }));
  } else {
    result = await NodeAgentService.NodeAgentRestart(params).catch(() => ({ workflow_id: '' }));
  }
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: { active: 'node' },
    });
  } else {
    router.push({ name: 'agent' });
  }
};

watch(restartDialogShow, (newVal, oldVal) => {
  if (oldVal && !newVal && !restartInProgress.value) {
    router.push({ name: 'agent' });
  }
});

const checkIfAtBottom = () => {
  if (footerRef.value) {
    const { bottom } = footerRef.value.getBoundingClientRect();
    isAtBottom.value = bottom >= window.innerHeight;
  }
};
const debouncedCheck = debounce(checkIfAtBottom, 100);

onMounted(async () => {
  if (footerRef.value) {
    window.addEventListener('resize', debouncedCheck);
    checkIfAtBottom();
  }

  if (nodeManageStore.assignUnitParams.isCrossPageSelection) {
    const allHosts: any[] = [];
    const pageSize = 1000;
    let offset = 0;
    let hasMore = true;
    loading.value = true;

    while (hasMore) {
      const hostListData = await TopoService.HostList({
        page: { offset, limit: pageSize },
        only_count: false,
        exact_include_conditions: nodeManageStore.assignUnitParams.queryParams.exact_include_conditions || {},
        exact_exclude_conditions: nodeManageStore.assignUnitParams.queryParams.exact_exclude_conditions || {},
        fuzzy_include_conditions: nodeManageStore.assignUnitParams.queryParams.fuzzy_include_conditions || {},
      }).catch(() => ({ total: 0, items: [] }));

      if (hostListData.items && hostListData.items.length > 0) {
        allHosts.push(...hostListData.items);
        offset += pageSize;
        if (hostListData.items.length < pageSize) hasMore = false;
      } else {
        hasMore = false;
      }
    }

    formData.info = allHosts.map((host: any) => ({
      ...host.state,
      ...host.info,
      ...host,
      bk_networkunit_id: normalizeNetworkUnitId(host.info.bk_networkunit_id),
      bk_host_innerip: host.info.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6: host.info.bk_host_innerip_v6_list?.join(','),
      login_mode: resolveLoginMode(host.info?.login_mode),
    }));
    loading.value = false;
  } else {
    formData.info = nodeManageStore.assignUnitParams.tableData.map(({ info, state, ...rest }: any) => ({
      ...rest,
      bk_networkunit_id: normalizeNetworkUnitId(info?.bk_networkunit_id ?? rest.bk_networkunit_id),
      bk_host_innerip: info?.bk_host_innerip_list?.join(',') ?? rest.bk_host_innerip,
      bk_host_innerip_v6: info?.bk_host_innerip_v6_list?.join(',') ?? rest.bk_host_innerip_v6,
      bk_networkarea_id: info?.bk_networkarea_id ?? rest.bk_networkarea_id,
      bk_networkarea_name: info?.bk_networkarea_name ?? rest.bk_networkarea_name,
      bk_biz_id: info?.bk_biz_id ?? rest.bk_biz_id,
      login_mode: resolveLoginMode(rest.login_mode || info?.login_mode),
    }));
  }
});

onUnmounted(() => {
  if (footerRef.value) {
    window.removeEventListener('resize', debouncedCheck);
  }
});
</script>


