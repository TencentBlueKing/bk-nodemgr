<template>
  <div class="step-deploy-target bg-[#fff] p-[24px] rounded-[2px]">
    <Form ref="formRef" :model="formData" :rules="rules" form-type="vertical">
      <!-- 插件名称 -->
      <Form.FormItem :label="$t('platform.nodeMan.pluginOperation.form.pluginName')" property="pluginName" required>
        <Select
          v-model="formData.pluginName"
          class="w-[568px]"
          filterable
          :disabled="!!initialPluginName"
          :placeholder="$t('platform.nodeMan.pluginOperation.form.pluginNamePlaceholder')"
          :loading="pluginListLoading"
        >
          <Select.Option
            v-for="item in pluginOptions"
            :key="item.name"
            :id="item.name"
            :name="item.name"
          />
        </Select>
      </Form.FormItem>

      <!-- 部署目标 -->
      <Form.FormItem :label="deployTargetLabel" property="selectedHosts" required>
        <!-- 手动输入提示：需在业务范围内 -->
        <div class="text-[12px] text-[#979BA5] mb-[8px]">
          {{ $t('pluginOperation.form.manualInputBizTip') }}
        </div>
        <div class="deploy-target-wrapper">
          <!-- IP 选择器组件 — section 模式 -->
          <IpSelector
            :key="ipSelectorKey"
            mode="section"
            :panel-list="['manualInput']"
            :value="ipSelectorValue"
            :keep-host-field-output="true"
            @change="handleIpSelectorChange"
          />

        </div>
      </Form.FormItem>

      <!-- 高级选项展开按钮（重启/停止操作不需要版本选择，隐藏高级选项） -->
      <Form.FormItem v-if="!isSimpleOperation">
        <Button
          text
          theme="primary"
          @click="showAdvanced = !showAdvanced"
        >
          <span class="mr-[8.5px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
          <down-shape :class="['text-[14px]', { 'transform rotate-180': showAdvanced }]"/>
        </Button>
      </Form.FormItem>

      <!-- 版本选择表格（重启/停止操作不需要） -->
      <Form.FormItem v-if="showAdvanced && !isSimpleOperation" required>
        <template #label>
          <span>{{ $t('pluginOperation.steps.packageVersion') }}</span>
        </template>
        <div class="w-[568px]">
          <Table
            :data="systemData"
            :border="true"
            width="568"
            :empty-text="$t('platform.nodeMan.installAgentPage.noAvailableVersion')"
          >
            <TableColumn
              field="os"
              :title="$t('platform.nodeMan.installAgentPage.osArch')"
              width="200"
            >
              <template #header>
                <span class="text-[14px]">{{ $t('platform.nodeMan.installAgentPage.osArch') }}</span>
              </template>
              <template #default="{ row }">
                <span :class="{ 'text-[#C4C6CC]': !isRowSelectable(row) }">{{ row.os?.replace('_', '/') }}</span>
              </template>
            </TableColumn>
            <TableColumn field="version" :title="$t('platform.nodeMan.installAgentPage.packageVersion')" width="368">
              <template #header>
                <span class="mr-[2px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.packageVersion') }}</span>
                <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                <Button text @click="handleBatchEditVersion">
                  <i class="nodeman-icon nc-bulk-edit cursor-pointer text-[14px]"></i>
                </Button>
              </template>
              <template #default="{ row }">
                <Input
                  :model-value="row.version"
                  :placeholder="isRowSelectable(row) ? $t('platform.nodeMan.installAgentPage.placeholder.select') : '--'"
                  :disabled="!isRowSelectable(row)"
                  @click="handleChooseVersion(row)"
                />
              </template>
            </TableColumn>
          </Table>
        </div>
      </Form.FormItem>
    </Form>

    <!-- 版本选择弹窗 -->
    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :batch="isBatch"
      :release-type="'plugin'"
      :plugin-name="formData.pluginName"
      @confirm="handleConfirmVersion"
    />
  </div>
</template>

<script lang="ts" setup>
import { Button, Form, Input, Select } from 'bkui-vue';


import { DownShape } from 'bkui-vue/lib/icon';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { PackageService } from '@/api/modules/pkg';
import { PluginAPIService } from '@/api/modules/plugin';
import { PACKAGE_GENERATION } from '@/common/const';
import ChooseVersionDialog from '@/components/choose-version-dialog.vue';
import IpSelector from '@/components/IpSelector';
import { fetchHostDetails } from '@/services/ip-selector';
import { setBizId, setType } from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';

const mainStore = useMainStore();
const route = useRoute();

// 初始化 IP 选择器（onMounted 中设置，确保 store 已就绪）






const { t } = useI18n();

const props = defineProps<{
  initialPluginName?: string;
  operationType?: string;
}>();

const formData = defineModel<{
  strategyName: string;
  pluginName: string;
  selectedHosts: any[];
  selectedVersion: string;
  paramConfig: Record<string, any>;
}>('formData', { required: true });

// 重启/停止操作不需要版本选择和参数配置
const isSimpleOperation = computed(() => ['restart', 'stop'].includes(props.operationType || ''));

const formRef = ref();
// IP 选择器 key：回填完成后递增，强制组件重建（解决 Vue3 包裹 Vue2 组件 props 不响应的问题）
const ipSelectorKey = ref(0);
const rules = {
  pluginName: [
    { required: true, message: t('platform.nodeMan.pluginOperation.form.pluginNameRequired'), trigger: 'change' },
  ],
  selectedHosts: [
    {
      required: true,
      message: t('pluginOperation.form.deployTargetRequired'),
      trigger: 'change',
      validator: (val: any[]) => Array.isArray(val) && val.length > 0,
    },
  ],
};

// ---------- 插件名称下拉 ----------
const pluginOptions = ref<{ name: string }[]>([]);
const pluginListLoading = ref(false);
const initialPluginName = computed(() => props.initialPluginName || '');

const loadPluginList = async () => {
  pluginListLoading.value = true;
  const res = await PluginAPIService.ListPlugins({
    page: { limit: 500, offset: 0 },
    exact_include_conditions: {
      visible_biz_ids: mainStore.selectedBusinessId,
    } as any,
    fuzzy_include_conditions: {} as any,
  }).catch(() => ({ total: 0, items: [] }));
  pluginOptions.value = (res.items || []).map((p: any) => ({ name: p.name }));
  pluginListLoading.value = false;
};

onMounted(async () => {
  // 初始化 IP 选择器：始终设置业务 ID（限制拓扑树只显示当前业务）
  setBizId(mainStore.selectedBusinessId);
  // 根据路由 type 参数设置 node_role 过滤（刷新后从 URL 恢复）
  // 有 type：按类型过滤（agent/proxy）；无 type：由 resolveNodeRoleFilter 自适应权限过滤
  const queryType = route.query.type as string;
  if (queryType) {
    setType(queryType);
  }

  loadPluginList();
  loadDefaultVersions();

  // 回填预填主机数据（重装/升级从 process/list 获取的主机ID列表）
  if (prefillHosts.value.length > 0) {
    // prefillHosts 现在是 host_id[] 数字数组
    const hostIds = prefillHosts.value.filter((id: any) => typeof id === 'number');
    if (hostIds.length > 0) {
      // 通过 fetchHostDetails 获取完整主机信息（IP、主机名、管控区域等）
      try {
        const hostDetailRes = await fetchHostDetails({
          hostList: hostIds.map((id: number) => ({ hostId: id })),
        });
        const hostDetailList = hostDetailRes?.data || [];
        if (hostDetailList.length > 0) {
          // 版本映射：从 prefillVersions 中取进程当前版本
          const versionMap = prefillVersions.value;
          formData.value.selectedHosts = hostDetailList.map((h: any) => ({
            bk_host_id: h.host_id || h.bk_host_id,
            host_id: h.host_id || h.bk_host_id,
            os_type: h.os_type || '',
            cpu_arch: h.cpu_arch || '',
            ip: h.ip || '',
            ipv6: h.ipv6 || '',
            host_name: h.host_name || '',
            bk_host_innerip: h.ip || '',
            bk_host_innerip_v6: h.ipv6 || '',
            bk_host_name: h.host_name || '',
            os_name: h.os_name || h.os_type || '',
            cloud_area: h.cloud_area || { id: 0, name: '' },
            alive: h.alive ?? 1,
            meta: h.meta || { scope_type: '', scope_id: '', bk_biz_id: '' },
            // 附加进程当前版本，供版本回填使用
            _processVersion: versionMap[`${h.os_type || ''}_${h.cpu_arch || ''}`] || '',
          }));
        }
      } catch {
        // 获取主机详情失败，用最小数据回填（仅 hostId，IP 选择器可后续自动补全）
        formData.value.selectedHosts = hostIds.map((id: number) => ({
          bk_host_id: id,
          host_id: id,
          os_type: '',
          cpu_arch: '',
          ip: '',
          ipv6: '',
          host_name: '',
          bk_host_innerip: '',
          bk_host_innerip_v6: '',
          bk_host_name: '',
          os_name: '',
          cloud_area: { id: 0, name: '' },
          alive: 1,
          meta: { scope_type: '', scope_id: '', bk_biz_id: '' },
        }));
      }
      // 强制 IP 选择器重建，使回填数据生效（Vue3 包裹 Vue2 组件 props 不响应式更新）
      ipSelectorKey.value++;
    }
  }
});


// 部署目标标签
const deployTargetLabel = computed(() => t('pluginOperation.form.deployTarget'));


// ---------- IP 选择器 ----------
// nameStyle: 'camelCase' 配置要求字段名使用 camelCase
const ipSelectorValue = computed(() => ({
  hostList: formData.value.selectedHosts.map((h: any) => ({
    hostId: h.host_id || h.bk_host_id || h.hostId || 0,
    ip: h.ip || h.bk_host_innerip || '',
    ipv6: h.ipv6 || h.bk_host_innerip_v6 || '',
    hostName: h.host_name || h.bk_host_name || '',
    osName: h.os_name || h.os_type || '',
    cloudArea: h.cloud_area || h.cloudArea || { id: h.bk_cloud_id || 0, name: '' },
    alive: h.alive ?? 1,
    meta: h.meta || { scope_type: '', scope_id: '', bk_biz_id: '' },
  })),
  nodeList: [],
  dynamicGroupList: [],
  serviceTemplateList: [],
  setTemplateList: [],
  serviceInstanceList: [],
}));

const handleIpSelectorChange = (value: any) => {
  const hosts = value?.hostList || [];
  formData.value.selectedHosts = hosts.map((h: any) => {
    const hostId = h.hostId || h.host_id || h.bk_host_id || 0;
    return {
      bk_host_id: hostId,
      host_id: hostId,
      bk_host_innerip: h.ip || '',
      bk_host_innerip_v6: h.ipv6 || '',
      bk_host_name: h.hostName || h.host_name || '',
      os_type: h.osType || h.os_type || '',
      cpu_arch: h.cpuArch || h.cpu_arch || '',
      ip: h.ip || '',
      ipv6: h.ipv6 || '',
      host_name: h.hostName || h.host_name || '',
      os_name: h.osName || h.os_name || '',
      cloud_area: h.cloudArea || h.cloud_area,
      alive: h.alive,
      meta: h.meta,
    };
  });
};




// ---------- 高级选项 & 版本选择 ----------
// 升级操作默认展开高级选项
const showAdvanced = ref(props.operationType === 'upgrade');

// 从路由 query 中解析预填数据（重装/升级传入的 host_id 列表）
const prefillHosts = computed(() => {
  try {
    const raw = route.query.prefillHosts as string;
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
});
const prefillVersions = computed(() => {
  try {
    const raw = route.query.prefillVersions as string;
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
});

// 系统/架构列表 — 由 plugin release 动态拼接
const systemData = ref<{ os: string; version: string }[]>([]);
// 系统/架构数据是否已加载过（防止重复调接口）
const systemLoaded = ref(false);

/** 根据已选主机计算出的平台集合（如 Set(['linux_x86_64', 'linux_aarch64'])） */
const selectedPlatforms = computed(() => {
  const set = new Set<string>();
  (formData.value.selectedHosts || []).forEach((h: any) => {
    const osType = h.os_type || '';
    const cpuArch = h.cpu_arch || '';
    if (osType && cpuArch) {
      set.add(`${osType}_${cpuArch}`);
    }
  });
  return set;
});

/** 判断某行是否可选（已选主机中存在该平台） */
const isRowSelectable = (row: { os: string }) => {
  // 没有选择主机时，所有行都可选
  if (selectedPlatforms.value.size === 0) return true;
  return selectedPlatforms.value.has(row.os);
};

/** 过滤出可选择的系统/架构（供 expose 和 param-config 使用） */
const selectableSystemData = computed(() => systemData.value.filter(item => isRowSelectable(item)));

const loadSystemArch = async () => {
  // 已加载过则直接返回，避免重复请求
  if (systemLoaded.value) return;
  systemLoaded.value = true;
  const pluginName = formData.value.pluginName;
  const res = await PackageService.ListReleasePluginBrief({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      enabled: [true],
      ...(pluginName ? { name: [pluginName] } : {}),
    },
  }).catch(() => ({ total: 0, items: [] }));

  const items = res.items || [];
  const osSet = new Set<string>();
  const defaultVersionMap = new Map<string, string>();
  // 所有可用版本集合（os_cpuArch → Set<version>），用于判断进程版本是否存在
  const availableVersionsMap = new Map<string, Set<string>>();

  items.forEach((item: any) => {
    if (item.os_type && item.cpu_arch) {
      const key = `${item.os_type}_${item.cpu_arch}`;
      osSet.add(key);
      // 记录默认版本
      if (item.as_default && !defaultVersionMap.has(key)) {
        defaultVersionMap.set(key, item.version);
      }
      // 收集所有可用版本
      if (!availableVersionsMap.has(key)) {
        availableVersionsMap.set(key, new Set());
      }
      availableVersionsMap.get(key)!.add(item.version);
    }
  });

  systemData.value = Array.from(osSet).filter(Boolean).map(os => ({
    os,
    version: defaultVersionMap.get(os) || '',
  }));

  // 用 prefillVersions（进程当前版本）回填：若进程版本在可用版本列表中则使用，否则保持默认版本
  const prefill = prefillVersions.value;
  if (Object.keys(prefill).length > 0) {
    systemData.value.forEach((sys) => {
      const processVersion = prefill[sys.os];
      if (processVersion) {
        const availableVersions = availableVersionsMap.get(sys.os);
        // 进程版本在可用列表中 → 使用进程版本；否则保持默认版本
        if (availableVersions && availableVersions.has(processVersion)) {
          sys.version = processVersion;
        }
      }
    });
  }

  // 自动设置第一个有默认版本的作为 selectedVersion
  const firstDefault = systemData.value.find(s => s.version);
  if (firstDefault && !formData.value.selectedVersion) {
    formData.value.selectedVersion = firstDefault.version;
  }
};

// 组件挂载时预加载版本数据（不需要展开高级选项）
const loadDefaultVersions = () => {
  loadSystemArch();
};

watch(
  () => showAdvanced.value,
  (visible) => {
    if (visible) {
      loadSystemArch();
    }
  },
);
watch(() => formData.value.pluginName, () => {
  // 切换插件时重置加载标志并清空旧数据，重新拉取对应插件的系统/架构列表
  systemLoaded.value = false;
  systemData.value = [];
  formData.value.selectedVersion = '';
  loadSystemArch();
});
const isShowDialog = ref(false);
const dialogData = ref<{ os: string; version: string }[]>([{ os: '', version: '' }]);
const isBatch = ref(false);

// 点击单行版本输入框 → 打开选择弹窗（单选模式，不可选行不响应）
const handleChooseVersion = (row: { os: string; version: string }) => {
  if (!isRowSelectable(row)) return;
  isShowDialog.value = true;
  dialogData.value = [row];
};

// 点击表头批量编辑图标 → 打开选择弹窗（批量模式，只传可选择的行）
const handleBatchEditVersion = () => {
  isShowDialog.value = true;
  dialogData.value = selectableSystemData.value;
  isBatch.value = true;
};

// 弹窗确认后，将选中版本回填到 systemData
const handleConfirmVersion = (data: any[]) => {
  systemData.value.forEach((sys) => {
    const find = data.find(item => sys.os === `${item.os_type}_${item.cpu_arch}`);
    if (find) {
      sys.version = find.version;
    }
  });
  isBatch.value = false;

  // 将选中的版本信息同步到 formData（取第一个有值的）
  const firstSelected = systemData.value.find(s => s.version);
  if (firstSelected) {
    formData.value.selectedVersion = firstSelected.version;
  }
};

// ---------- 校验 ----------
const validate = async () => {
  try {
    const result = await formRef.value?.validate().catch(() => false);
    return !!result;
  } catch {
    return false;
  }
};

defineExpose({ validate, systemData: selectableSystemData });
</script>

<style lang="postcss" scoped>
.step-deploy-target {
  /* 填满父容器高度，不额外撑开 */
  display: flex;
  flex-direction: column;
}

.deploy-target-wrapper {
  width: 100%;
  height: 632px;
}


:deep(.ip-selector-agent-status) {
  display: inline-flex;
  align-items: center;
}

:deep(.ip-selector-agent-status .ip-selector-icon) {
  flex-shrink: 0;
}


</style>
