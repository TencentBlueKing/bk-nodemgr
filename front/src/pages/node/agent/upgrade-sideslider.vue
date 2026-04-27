<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="drawerTitle"
    width="1200"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <div class="py-[20px] px-[40px]">
      <Form :model="formData" class="mt-[24px]">
        <Form.FormItem
          :label="$t('platform.nodeMan.agentStatus.upgradeMode')"
          label-width="100"
          required
        >
          <div class="upgrade-mode-selector">
            <div
              v-for="item in upgradeModeList"
              :key="item.type"
              :class="['upgrade-mode-item', { active: upgradeMode === item.type }]"
              @click="upgradeMode = item.type"
            >
              <div class="prefix">
                <i :class="['text-[18px]', 'nodeman-icon', item.icon]"></i>
              </div>
              <Popover
                theme="light"
                trigger="hover"
                placement="top"
                :arrow="true"
                :max-width="280"
                :offset="8"
                :popover-delay="[0, 100]"
                :component-event-delay="0"
              >
                <div class="text-[12px] text-[#313238] leading-[19px] ml-[12px] content cursor-default">
                  {{ item.name }}
                </div>
                <template #content>
                  <div class="text-[12px] leading-[20px]">{{ item.desc }}</div>
                </template>
              </Popover>
              <div class="checked" v-show="upgradeMode === item.type">
                <i class="nodeman-icon nc-check-small"></i>
              </div>
            </div>
          </div>
        </Form.FormItem>

        <Form.FormItem
          v-if="upgradeMode === 'graceful'"
          :label="$t('platform.nodeMan.agentStatus.timeout')"
          label-width="100"
        >
          <Input
            class="w-[120px]"
            type="number"
            :min="1"
            v-model="formData.gracefulRestartTimeoutSec"
          >
            <template #suffix>
              <span class="px-[8px] text-[#979ba5]">s</span>
            </template>
          </Input>
        </Form.FormItem>

        <Form.FormItem
          :label="$t('platform.nodeMan.agentStatus.hostInfo')"
          property=""
          label-width="100"
          required
        >
          <install-table
            ref="installTableRef"
            v-model:data="hostTableData"
            :release-type="releaseType"
            :is-reinstall="true"
            :is-upgrade="true"
            :current-settings="hostTableSettings"
            :max-height="360"
          />
        </Form.FormItem>

        <Form.FormItem
          :label="versionLabel"
          label-width="100"
          required
        >
          <div class="w-[488px]">
            <Table
              :data="systemData"
              :border="true"
              :empty-text="emptyVersionText"
            >
              <TableColumn
                field="displayName"
                :title="$t('platform.nodeMan.installAgentPage.osArch')"
                width="200"
              >
                <template #default="{ row }">
                  {{ row.os.replace('_', '/') }}
                </template>
              </TableColumn>
              <TableColumn
                field="version"
                :title="$t('platform.nodeMan.installAgentPage.packageVersion')"
                width="288"
              >
                <template #header>
                  <div class="flex items-center">
                    <span class="mr-[2px] text-[14px]">
                      {{ $t('platform.nodeMan.installAgentPage.packageVersion') }}
                    </span>
                    <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                    <Button text @click="handleBatchEditVersion">
                      <i class="nodeman-icon nc-bulk-edit cursor-pointer text-[14px]"></i>
                    </Button>
                  </div>
                </template>
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="(el: any) => setInputRef(row.os, el)"
                  >
                    <Input
                      :model-value="row.version"
                      :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.select')"
                      @click="handleChooseVersion(row)"
                    />
                  </Validate>
                </template>
              </TableColumn>
            </Table>
          </div>
        </Form.FormItem>
      </Form>
    </div>

    <template #footer>
      <div class="flex mt-[32px] ml-[115px]">
        <Button
          theme="primary"
          class="mr-[8px]"
          :disabled="systemData.length === 0"
          v-bk-tooltips="{
            content: emptyVersionTip,
            disabled: systemData.length > 0,
          }"
          @click="handlePreview"
        >
          <span>{{ $t('platform.nodeMan.agentStatus.goUpgrade') }}</span>
          <span
            class="ml-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
          >
            {{ hosts.length }}
          </span>
        </Button>
        <Button @click="handleBeforeClose">
          {{ $t('action.cancel') }}
        </Button>
      </div>
    </template>

    <upgrade-preview-sideslider
      v-model:is-show="previewData.isShow"
      :hosts="previewData.hosts"
      :form-data="previewData.formData"
      :release-type="releaseType"
    />

    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :batch="isBatch"
      :release-type="releaseType"
      @confirm="handleConfirmVersion"
    />
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Popover, Sideslider } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import type { UpgradeFormData } from './upgrade-preview-sideslider.vue';
import UpgradePreviewSideslider from './upgrade-preview-sideslider.vue';

import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';
import ChooseVersionDialog from '@/components/choose-version-dialog.vue';
import InstallTable from '@/components/install-table.vue';
import Validate from '@/components/validate.vue';

const isShow = defineModel<boolean>('isShow', { default: false });

const props = defineProps({
  hosts: {
    type: Array as () => Host[],
    default: () => [],
  },
  releaseType: {
    type: String as () => 'agent' | 'proxy',
    default: 'agent',
  },
});

const { t } = useI18n();

const drawerTitle = computed(() => (
  props.releaseType === 'proxy'
    ? t('topoManager.workAreaDetail.dropdown.upgrade')
    : t('platform.nodeMan.agentStatus.agentUpgrade')
));

const versionLabel = computed(() => (
  props.releaseType === 'proxy'
    ? t('installProxy.proxyVersion')
    : t('platform.nodeMan.installAgentPage.version')
));

const emptyVersionText = computed(() => (
  props.releaseType === 'proxy'
    ? t('installProxy.noAvailableVersion')
    : t('platform.nodeMan.installAgentPage.noAvailableVersion')
));

const emptyVersionTip = computed(() => (
  props.releaseType === 'proxy'
    ? t('installProxy.noAvailableVersionTip')
    : t('platform.nodeMan.installAgentPage.noAvailableVersionTip')
));

const hostTableSettings = reactive({
  fields: [
    { field: 'bk_biz_id', title: t('installProxy.business') },
    { field: 'bk_networkarea_name', title: t('platform.nodeMan.bk_cloud_name') },
    { field: 'bk_networkunit_id', title: t('platform.nodeMan.bk_cloud_unit') },
    { field: 'bk_host_innerip', title: t('installProxy.innerIPv4') },
    { field: 'bk_host_innerip_v6', title: t('installProxy.innerIPv6') },
    { field: 'os_type', title: t('installProxy.os') },
    { field: 'cpu_arch', title: t('components.installTable.cpuArch') },
  ],
  checked: [
    'bk_biz_id',
    'bk_networkarea_name',
    'bk_networkunit_id',
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'cpu_arch',
  ],
  disabled: [
    'bk_biz_id',
    'bk_networkarea_name',
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'cpu_arch',
  ],
  size: 'medium' as const,
});

const hostTableData = ref<any[]>([]);
const installTableRef = ref<InstanceType<typeof InstallTable> | null>(null);

const formData = reactive<UpgradeFormData>({
  networkUnitId: undefined,
  targetVersions: [],
  force: false,
  gracefulRestartTimeoutSec: 120,
});

const upgradeMode = ref<'force' | 'graceful'>('graceful');
const upgradeModeList = computed(() => {
  const forceOption = {
    type: 'force' as const,
    name: t('platform.nodeMan.agentStatus.forceUpgrade'),
    icon: 'nc-deploy',
    desc: t('platform.nodeMan.agentStatus.forceUpgradeDesc'),
  };
  if (props.releaseType === 'proxy') {
    return [forceOption];
  }
  return [
    {
      type: 'graceful' as const,
      name: t('platform.nodeMan.agentStatus.gracefulUpgrade'),
      icon: 'nc-strategy',
      desc: t('platform.nodeMan.agentStatus.gracefulUpgradeDesc'),
    },
    forceOption,
  ];
});
const systemData = ref<Array<{ os: string; version: string }>>([]);
const isShowDialog = ref(false);
const isBatch = ref(false);
const dialogData = ref([{ os: '', version: '' }]);
const inputRefs = ref<Map<string, any>>(new Map());
const previewData = reactive({
  isShow: false,
  hosts: [] as Host[],
  formData: null as UpgradeFormData | null,
});

const setInputRef = (os: string, el: any) => {
  if (el) {
    inputRefs.value.set(os, el);
  }
};

const defaultVersionMap = ref<Map<string, string>>(new Map());

const fetchDefaultVersions = async () => {
  try {
    const service = props.releaseType === 'proxy'
      ? PackageService.ListReleaseProxyBrief
      : PackageService.ListReleaseAgentBrief;

    const res = await service({
      page: { limit: 500, offset: 0 },
      generation: PACKAGE_GENERATION,
      exact_include_conditions: {
        enabled: [true],
      },
    });

    const versionMap = new Map<string, string>();
    const firstVersionMap = new Map<string, string>();

    if (res?.items) {
      res.items.forEach((item: any) => {
        const key = `${item.os_type}_${item.cpu_arch}`;
        if (item.as_default) {
          versionMap.set(key, item.version);
        }
        if (!firstVersionMap.has(key)) {
          firstVersionMap.set(key, item.version);
        }
      });
    }

    firstVersionMap.forEach((version, key) => {
      if (!versionMap.has(key)) {
        versionMap.set(key, version);
      }
    });

    defaultVersionMap.value = versionMap;
  } catch {
    defaultVersionMap.value = new Map();
  }
};

const updateSystemData = async () => {
  const osArchSet = new Set<string>();
  hostTableData.value.forEach((row: any) => {
    if (row.os_type && row.cpu_arch) {
      osArchSet.add(`${row.os_type}_${row.cpu_arch}`);
    }
  });

  const currentOsArchs = new Set(systemData.value.map(item => item.os));
  const newOsArchs = [...osArchSet];

  if (newOsArchs.length !== currentOsArchs.size
    || !newOsArchs.every(os => currentOsArchs.has(os))) {
    const oldVersions = new Map(systemData.value.map(item => [item.os, item.version]));

    if (defaultVersionMap.value.size === 0) {
      await fetchDefaultVersions();
    }

    systemData.value = newOsArchs.map(os => ({
      os,
      version: oldVersions.get(os) || defaultVersionMap.value.get(os) || '',
    }));
  }
};

const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
  isBatch.value = false;
};

const handleBatchEditVersion = () => {
  isShowDialog.value = true;
  dialogData.value = systemData.value;
  isBatch.value = true;
};

const handleConfirmVersion = (data: any[]) => {
  systemData.value.forEach(sys => {
    const matched = data.find(item => sys.os === `${item.os_type}_${item.cpu_arch}`);
    if (matched) {
      sys.version = matched.version;
    }
  });
  isBatch.value = false;
};

const systemValidate = async () => {
  const refs = Array.from(inputRefs.value.values());
  const result = await Promise.all(refs.map(item => item.validate('blur')));
  return result.every(Boolean);
};

const handlePreview = async () => {
  const tableValid = await installTableRef.value?.tableValidate();
  if (!tableValid) {
    return;
  }

  const valid = await systemValidate();
  if (!valid) {
    return;
  }

  formData.targetVersions = systemData.value.map(item => {
    const [os_type, cpu_arch] = item.os.split('_');
    return { os_type, cpu_arch, version: item.version };
  });
  previewData.hosts = hostTableData.value.map((host: any) => ({ ...host }));
  previewData.formData = {
    networkUnitId: formData.networkUnitId,
    targetVersions: formData.targetVersions,
    force: upgradeMode.value === 'force',
    gracefulRestartTimeoutSec: upgradeMode.value === 'graceful'
      ? (Number(formData.gracefulRestartTimeoutSec) || 0)
      : 0,
  };
  previewData.isShow = true;
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      isShow.value = false;
      resolve(true);
    },
    onCancel: () => reject(),
  });
});

watch(
  () => isShow.value,
  async (value) => {
    if (!value) {
      return;
    }

    formData.networkUnitId = undefined;
    formData.targetVersions = [];
    formData.gracefulRestartTimeoutSec = 120;
    upgradeMode.value = props.releaseType === 'proxy' ? 'force' : 'graceful';
    previewData.isShow = false;
    previewData.hosts = [];
    previewData.formData = null;
    systemData.value = [];
    inputRefs.value.clear();

    hostTableData.value = props.hosts.map((host: any) => {
      const info = host.info ?? host;
      const state = host.state ?? {};
      return {
        bk_host_id: host.bk_host_id,
        bk_biz_id: info.bk_biz_id || '',
        bk_host_innerip: (info.bk_host_innerip_list ?? []).join(';') || info.bk_host_innerip || '',
        bk_host_innerip_v6: (info.bk_host_innerip_v6_list ?? []).join(';') || info.bk_host_innerip_v6 || '',
        bk_networkarea_id: info.bk_networkarea_id ?? '',
        bk_networkarea_name: info.bk_networkarea_name ?? '',
        bk_networkunit_id: info.bk_networkunit_id != null ? String(info.bk_networkunit_id) : '',
        bk_networkunit_name: info.bk_networkunit_name ?? '',
        os_type: info.os_type || '',
        cpu_arch: info.cpu_arch || '',
        node_version: state.node_version || info.node_version || '',
      };
    });

    await fetchDefaultVersions();
    updateSystemData();
  },
);

watch(
  () => hostTableData.value.map((row: any) => `${row.os_type}_${row.cpu_arch}`).join(','),
  () => {
    updateSystemData();
  },
);
</script>

<style lang="postcss" scoped>
.upgrade-mode-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 568px;

  .upgrade-mode-item {
    position: relative;
    width: 50%;
    height: 32px;
    background: #ffffff;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    display: flex;
    align-items: center;
    cursor: pointer;

    &:hover {
      border-color: #3a84ff;
    }

    &.active {
      border-color: #3a84ff;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);

      .prefix {
        background: #e1ecff;
        color: #3a84ff;
      }
    }

    .prefix {
      min-width: 40px;
      height: 100%;
      border-right: 1px solid #c4c6cc;
      background: #f5f7fa;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 21px;
      color: #979ba5;
    }

    .content {
      border-bottom: 1px dashed #c4c6cc;
    }

    .checked {
      position: absolute;
      top: 0;
      right: 0;
      width: 0;
      height: 0;
      border-style: solid;
      border-width: 0 32px 32px 0;
      border-color: transparent #3a84ff transparent transparent;

      .nc-check-small {
        position: absolute;
        font-size: 20px;
        color: #fff;
        top: 0;
        right: -32px;
      }
    }
  }
}
</style>
