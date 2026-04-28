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
            mode="section"
            :panel-list="['manualInput']"
            :value="ipSelectorValue"
            :keep-host-field-output="true"
            @change="handleIpSelectorChange"
          />

        </div>
      </Form.FormItem>

      <!-- 高级选项展开按钮 -->
      <Form.FormItem>
        <Button
          text
          theme="primary"
          @click="showAdvanced = !showAdvanced"
        >
          <span class="mr-[8.5px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
          <down-shape :class="['text-[14px]', { 'transform rotate-180': showAdvanced }]"/>
        </Button>
      </Form.FormItem>

      <!-- 版本选择表格 -->
      <Form.FormItem v-if="showAdvanced" required>
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
                {{ row.os?.replace('_', '/') }}
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
                  :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.select')"
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
import { setBizId, setType } from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';

const mainStore = useMainStore();
const route = useRoute();

// 初始化 IP 选择器（onMounted 中设置，确保 store 已就绪）






const { t } = useI18n();

const props = defineProps<{
  initialPluginName?: string;
}>();

const formData = defineModel<{
  strategyName: string;
  pluginName: string;
  selectedHosts: any[];
  selectedVersion: string;
  paramConfig: Record<string, any>;
}>('formData', { required: true });

const formRef = ref();
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

onMounted(() => {
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
});


// 部署目标标签
const deployTargetLabel = computed(() => t('pluginOperation.form.deployTarget'));


// ---------- IP 选择器 ----------
const ipSelectorValue = computed(() => ({
  hostList: formData.value.selectedHosts.map((h: any) => ({
    host_id: h.host_id || h.bk_host_id || h.hostId || 0,
    ip: h.ip || h.bk_host_innerip || '',
    ipv6: h.ipv6 || h.bk_host_innerip_v6 || '',
    host_name: h.host_name || h.bk_host_name || '',
    os_name: h.os_name || h.os_type || '',
    cloud_area: h.cloud_area || h.cloudArea || { id: h.bk_cloud_id || 0, name: '' },
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
      os_type: h.osName || h.os_name || h.os_type || '',
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
const showAdvanced = ref(false);

// 系统/架构列表 — 由 plugin release 动态拼接
const systemData = ref<{ os: string; version: string }[]>([]);
// 系统/架构数据是否已加载过（防止重复调接口）
const systemLoaded = ref(false);

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

  items.forEach((item: any) => {
    if (item.os_type && item.cpu_arch) {
      const key = `${item.os_type}_${item.cpu_arch}`;
      osSet.add(key);
      // 记录默认版本
      if (item.as_default && !defaultVersionMap.has(key)) {
        defaultVersionMap.set(key, item.version);
      }
    }
  });

  systemData.value = Array.from(osSet).filter(Boolean).map(os => ({
    os,
    version: defaultVersionMap.get(os) || '',
  }));

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

// 点击单行版本输入框 → 打开选择弹窗（单选模式）
const handleChooseVersion = (row: { os: string; version: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
};

// 点击表头批量编辑图标 → 打开选择弹窗（批量模式）
const handleBatchEditVersion = () => {
  isShowDialog.value = true;
  dialogData.value = systemData.value;
  isBatch.value = true;
};

// 弹窗确认后，将选中版本回填到 systemData
const handleConfirmVersion = (data: any[]) => {
  systemData.value.forEach((sys) => {
    const find = data.find(item => sys.os === `${item.os_type}_${item.cup_arch}`);
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

defineExpose({ validate, systemData });
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
