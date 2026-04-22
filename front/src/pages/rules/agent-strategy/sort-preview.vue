<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    width="800"
    :before-close="handleBeforeClose"
  >
    <template #header>
      <div class="flex items-center">
        <span class="text-[16px] font-bold text-[#313238] mr-[24px]">{{ t('agentStrategy.sortAndPreview') }}</span>
        <!-- 步骤指示器 -->
        <div class="flex items-center">
          <div
            v-for="(step, index) in steps"
            :key="index"
            class="flex items-center"
          >
            <div
              :class="['flex items-center cursor-pointer', { 'opacity-50': index > currentStep }]"
              @click="handleStepClick(index)"
            >
              <div
                :class="['w-[24px] h-[24px] rounded-full flex items-center justify-center text-[12px] font-medium', getStepClass(index)]"
              >
                <span>{{ index + 1 }}</span>
              </div>
              <span
                :class="['ml-[8px] text-[14px]', index <= currentStep ? 'text-[#313238]' : 'text-[#979BA5]']"
              >{{ step }}</span>
            </div>
            <div
              v-if="index < steps.length - 1"
              class="w-[60px] h-[1px] bg-[#DCDEE5] mx-[12px]"
            ></div>
          </div>
        </div>
      </div>
    </template>

    <div class="p-[24px] h-full overflow-auto">
      <!-- Step 0: 策略排序 -->
      <div v-show="currentStep === 0">
        <Table
          ref="sortTableRef"
          :data="sortedConfigs"
          :max-height="maxTableHeight"
          :row-config="{ drag: true }"
          :row-drag-config="rowDragConfig"
          show-overflow-tooltip
          @row-dragend="handleDragEnd"
        >
          <TableColumn
            field="configpolicy_name"
            :title="t('agentStrategy.table.configName')"
            :min-width="200"
          >
            <template #default="{ row }">
              <div class="flex items-center" :class="{ 'opacity-50': !row.enabled }">
                <Tag class="mr-[8px]">{{ row.enabled ? enabledIndexOf(row) + 1 : '-' }}</Tag>
                <i :class="`nodeman-icon nc-${row.enabled ? 'success' : 'incomplete'} status-icon`"></i>
                <span>{{ row.configpolicy_name }}</span>
              </div>
            </template>
          </TableColumn>
          <TableColumn
            field="sort"
            drag-sort
            width="120"
            align="right"
          >
            <template #default="{ row }">
              <span v-if="!row.enabled" class="text-[12px] text-[#979BA5]">{{ t('agentStrategy.table.disabled') }}</span>
            </template>
          </TableColumn>
        </Table>
      </div>

      <!-- Step 1: 选择作用IP -->
      <div v-show="currentStep === 1">
        <div class="text-[14px] text-[#63656E] font-medium mb-[12px]">
          step 1. {{ t('agentStrategy.preview.selectIP') }}
        </div>
        <!-- 添加IP按钮 -->
        <div
          v-if="selectedHosts.length === 0"
          class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer mb-[10px]"
          @click="handleAddIP"
        >
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          <span class="text-[14px]">{{ $t('common.addIP') }}</span>
        </div>

        <!-- IP统计和操作 -->
        <div v-if="selectedHosts.length > 0" class="mb-[16px]">
          <div class="flex items-center mb-[10px]">
            <span class="text-[14px] text-[#63656E]">
              {{ t('common.totalAdded') }}
              <span class="text-[#3A84FF] font-medium mx-[2px]">{{ selectedHosts.length }}</span>
              {{ t('common.hostUnit') }}
            </span>
            <i class="nodeman-icon nc-icon-edit-2 text-[16px] text-[#979BA5] cursor-pointer ml-[10px]" @click="handleEditHosts"></i>
          </div>
        </div>

        <div class="text-[14px] text-[#63656E] font-medium mb-[12px] mt-[24px]">
          step 2. {{ t('agentStrategy.preview.previewMatchStrategy') }}
        </div>

        <!-- 预览匹配策略表格 -->
        <Loading :loading="previewLoading">
          <div v-if="previewData.length > 0">
            <Table
              ref="previewTableRef"
              :data="paginatedPreviewData"
              :max-height="300"
              :pagination="previewPagination"
              show-overflow-tooltip
              @page-value-change="handlePreviewPageChange"
              @page-limit-change="handlePreviewPageLimitChange"
            >
              <TableColumn
                field="bk_host_id"
                title="Host ID"
                :min-width="100"
              ></TableColumn>
              <TableColumn
                field="matched_policies"
                :title="t('agentStrategy.preview.matchedPolicy')"
                :min-width="200"
              >
                <template #default="{ row }">
                  <span v-if="row.matched_policies?.length">
                    {{ row.matched_policies.map((p: any) => p.configpolicy_name).join(', ') }}
                  </span>
                  <span v-else>-</span>
                </template>
              </TableColumn>
              <TableColumn
                field="action"
                :title="$t('table.action')"
                :min-width="100"
              >
                <template #default="{ row }">
                  <Button
                    text
                    theme="primary"
                    @click="handleViewConfigDetail(row)"
                  >{{ t('agentStrategy.preview.configDetail') }}</Button>
                </template>
              </TableColumn>
            </Table>
          </div>
        </Loading>
      </div>
    </div>

    <template #footer>
      <div class="flex items-center">
        <template v-if="currentStep === 0">
          <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleNextStep">
            {{ t('agentStrategy.preview.previewConfig') }}
          </Button>
          <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
        </template>
        <template v-else>
          <Button class="mr-[8px] w-[88px]" @click="handlePrevStep">
            {{ t('agentStrategy.preview.prevStep') }}
          </Button>
          <span
            v-bk-tooltips="{
              content: t('agentStrategy.preview.noSortData'),
              disabled: hasEnabledConfigs,
            }"
            class="mr-[8px] inline-block"
          >
            <Button
              theme="primary"
              class="w-[88px]"
              :loading="saveLoading"
              :disabled="!hasEnabledConfigs"
              @click="handleSaveSort"
            >
              {{ t('agentStrategy.preview.saveSort') }}
            </Button>
          </span>
          <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
        </template>
      </div>
    </template>

    <!-- 蓝鲸IP选择器 -->
    <IpSelector
      mode="dialog"
      :show-dialog="isShowIpSelector"
      :value="ipSelectorValue"
      @change="handleIpSelectorChange"
      @close-dialog="handleIpSelectorClose"
    />

    <!-- JSON配置详情弹窗 -->
    <Dialog
      :is-show="isShowConfigDetail"
      :title="t('agentStrategy.preview.jsonConfigDetail')"
      width="600"
      @closed="isShowConfigDetail = false"
      @confirm="isShowConfigDetail = false"
    >
      <div class="p-[16px]">
        <pre class="bg-[#F5F7FA] p-[16px] rounded-[4px] text-[13px] text-[#313238] overflow-auto max-h-[500px] whitespace-pre-wrap">{{ configDetailJson }}</pre>
      </div>
    </Dialog>
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Dialog, InfoBox, Loading, Message, Sideslider, Tag } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import type { ConfigPolicyPreviewRespPreviewItem } from '@/@types/configpolicy';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import IpSelector from '@/components/IpSelector';
import { fetchHostDetails, setPolicyType, setStrategyBizId } from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';

interface ISelectedHost {
  bk_host_id: number;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  bk_host_name: string;
  bk_networkarea_name: string;
  os_type: string;
  cpu_arch: string;
  bk_networkarea_id: number;
}

const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps<{
  configList: ConfigPolicy[];
  configpolicyType: string;
  bizId: number;
}>();

const emit = defineEmits(['save-sort']);
const { t } = useI18n();
const mainStore = useMainStore();

// 步骤
const steps = computed(() => [
  t('agentStrategy.preview.sortConfig'),
  t('agentStrategy.preview.previewConfig'),
]);
const currentStep = ref(0);
const maxTableHeight = 500;

// 排序配置
const sortTableRef = ref<any>(null);
const sortedConfigs = ref<ConfigPolicy[]>([]);
const saveLoading = ref(false);
// 是否存在启用的策略（保存排序按钮仅在有启用项时可用）
const hasEnabledConfigs = computed(() => sortedConfigs.value.some(item => item.enabled));

const rowDragConfig = {
  disabledMethod: ({ row }: { row: ConfigPolicy }) => !row.enabled,
  visibleMethod: ({ row }: { row: ConfigPolicy }) => row.enabled,
};

const enabledIndexOf = (row: ConfigPolicy) => {
  let idx = 0;
  for (const item of sortedConfigs.value) {
    if (!item.enabled) continue;
    if (item === row) return idx;
    idx++;
  }
  return -1;
};

const handleDragEnd = () => {
  // 从 table 内部获取拖拽后的最新排序数据，同步到 sortedConfigs
  const vxeInstance = sortTableRef.value?.getVxeTableInstance?.();
  if (vxeInstance) {
    sortedConfigs.value = vxeInstance.getTableData().fullData;
  }
  // 保证未启用的始终在最下方
  const enabled = sortedConfigs.value.filter(item => item.enabled);
  const disabled = sortedConfigs.value.filter(item => !item.enabled);
  sortedConfigs.value = [...enabled, ...disabled];
};

// IP选择器
const isShowIpSelector = ref(false);
const ipSelectorValue = ref({
  hostList: [],
  nodeList: [],
  dynamicGroupList: [],
  serviceTemplateList: [],
  setTemplateList: [],
});

const selectedHosts = ref<ISelectedHost[]>([]);

const handleAddIP = () => {
  setPolicyType(props.configpolicyType);
  setStrategyBizId(props.bizId);
  isShowIpSelector.value = true;
};

const handleEditHosts = () => {
  setPolicyType(props.configpolicyType);
  setStrategyBizId(props.bizId);
  isShowIpSelector.value = true;
};

const handleClearHosts = () => {
  InfoBox({
    title: t('common.confirmDelete'),
    subTitle: t('common.confirmDeleteAllHosts'),
    onConfirm: () => {
      ipSelectorValue.value = {
        hostList: [],
        nodeList: [],
        dynamicGroupList: [],
        serviceTemplateList: [],
        setTemplateList: [],
      };
      selectedHosts.value = [];
      previewData.value = [];
      previewPagination.count = 0;
      previewPagination.current = 1;
    },
  });
};

// IP选择器变化事件（dialog模式：选择即生效+即时预览）
const handleIpSelectorChange = async (value: any) => {
  // 用 hostId 列表查询完整主机信息，补回被库丢弃的字段
  const hostList = value.hostList || [];
  if (hostList.length > 0) {
    try {
      const detailRes = await fetchHostDetails({
        hostList: hostList.map((h: any) => ({ hostId: h.hostId, meta: h.meta })),
      });
      const detailMap = new Map<number, any>();
      (detailRes.data || []).forEach((h: any) => {
        detailMap.set(h.host_id, h);
      });
      hostList.forEach((host: any) => {
        const detail = detailMap.get(host.hostId);
        if (detail) {
          host.host_name = detail.host_name;
          host.os_type = detail.os_type;
          host.cpu_arch = detail.cpu_arch;
        }
      });
    } catch {
      // 补全失败不影响主流程
    }
  }

  ipSelectorValue.value = { ...value, hostList: [...hostList] };
  extractHosts();
  // 选择变化后即时触发预览
  if (selectedHosts.value.length > 0) {
    fetchPreviewMatchStrategy();
  }
};

// IP选择器关闭
const handleIpSelectorClose = () => {
  isShowIpSelector.value = false;
};

const extractHosts = () => {
  const hosts: ISelectedHost[] = [];
  if (ipSelectorValue.value.hostList && ipSelectorValue.value.hostList.length > 0) {
    ipSelectorValue.value.hostList.forEach((host: any) => {
      hosts.push({
        bk_host_id: host.hostId || host.host_id || host.bk_host_id,
        bk_host_innerip: host.ip || host.bk_host_innerip,
        bk_host_innerip_v6: host.ipv6 || host.bk_host_innerip_v6 || '',
        bk_host_name: host.hostName || host.host_name || host.bk_host_name || '',
        bk_networkarea_name: host.cloudArea?.name || host.cloud_area?.name || host.bk_cloud_name || '',
        os_type: host.osName || host.os_name || host.os_type || '',
        cpu_arch: host.cpuArch || host.cpu_arch || '',
        bk_networkarea_id: host.cloudArea?.id || host.cloud_area?.id || host.cloudId || host.cloud_id || host.bk_cloud_id || 0,
      });
    });
  }
  selectedHosts.value = hosts;
};

// 预览匹配策略
const previewData = ref<ConfigPolicyPreviewRespPreviewItem[]>([]);
const previewLoading = ref(false);
const previewPagination = reactive({
  current: 1,
  limit: 10,
  count: 0,
});

const paginatedPreviewData = computed(() => {
  const start = (previewPagination.current - 1) * previewPagination.limit;
  const end = start + previewPagination.limit;
  return previewData.value.slice(start, end);
});

const fetchPreviewMatchStrategy = async () => {
  previewLoading.value = true;
  try {
    const res = await ConfigPolicyAPIService.ConfigPolicyPreview({
      bk_biz_id: props.bizId,
      policy_type: props.configpolicyType,
      hosts: selectedHosts.value.map(host => ({
        bk_host_id: host.bk_host_id,
        bk_networkunit_id: 0,
        bk_networkarea_id: host.bk_networkarea_id || 0,
        os_type: host.os_type || '',
        cpu_arch: host.cpu_arch || '',
      })),
    });
    const items = [...(res.reliable_items || []), ...(res.unreliable_items || [])];
    previewData.value = items;
    previewPagination.count = items.length;
    previewPagination.current = 1;
  } catch {
    previewData.value = [];
    previewPagination.count = 0;
  } finally {
    previewLoading.value = false;
  }
};

const handlePreviewPageChange = (page: number) => {
  previewPagination.current = page;
};

const handlePreviewPageLimitChange = (limit: number) => {
  previewPagination.limit = limit;
  previewPagination.current = 1;
};

// 配置详情弹窗
const isShowConfigDetail = ref(false);
const configDetailJson = ref('');

const handleViewConfigDetail = (row: ConfigPolicyPreviewRespPreviewItem) => {
  const mergedConfig = {
    ...row.merged_configs_string,
    ...row.merged_configs_int,
    ...row.merged_configs_bool,
  };
  configDetailJson.value = JSON.stringify(mergedConfig, null, 2);
  isShowConfigDetail.value = true;
};

// 步骤导航
const handleStepClick = (index: number) => {
  if (index <= currentStep.value) {
    currentStep.value = index;
  }
};

const getStepClass = (index: number) => {
  if (index < currentStep.value) {
    return 'bg-[#E1ECFF] text-[#3A84FF]'; // completed
  }
  if (index === currentStep.value) {
    return 'bg-[#3A84FF] text-white'; // active
  }
  return 'bg-[#F0F1F5] text-[#979BA5]'; // pending
};

const handlePrevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value -= 1;
  }
};

const handleNextStep = () => {
  if (currentStep.value < steps.value.length - 1) {
    currentStep.value += 1;
  }
};

const handleSaveSort = async () => {
  saveLoading.value = true;
  try {
    await ConfigPolicyAPIService.ConfigPolicyPriorityReorder({
      bk_biz_id: props.bizId,
      configpolicy_type: props.configpolicyType,
      ordered_configpolicy_id: sortedConfigs.value.filter(item => item.enabled).map(item => item.configpolicy_id),
    });
    Message({ theme: 'success', message: t('agentStrategy.preview.saveSuccess') });
    emit('save-sort');
    isShow.value = false;
  } catch {
    // error handled by fetch interceptor
  } finally {
    saveLoading.value = false;
  }
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve) => {
  resolve(true);
  isShow.value = false;
});

// 初始化
watch(() => isShow.value, (val) => {
  if (val) {
    currentStep.value = 0;
    const enabled = props.configList.filter(item => item.enabled);
    const disabled = props.configList.filter(item => !item.enabled);
    sortedConfigs.value = [...enabled, ...disabled];
    selectedHosts.value = [];
    previewData.value = [];
    ipSelectorValue.value = {
      hostList: [],
      nodeList: [],
      dynamicGroupList: [],
      serviceTemplateList: [],
      setTemplateList: [],
    };
    previewPagination.count = 0;
    previewPagination.current = 1;
  }
});
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 8px;
  height: 8px;
  border: 1px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
}
.nc-success {
  &::before {
    border-color: #2DCC56;
    background: #cbf0da;
  }
}
.nc-incomplete {
  &::before {
    border-color: #90A4B2;
    background: #e2e7eb;
  }
}
</style>
