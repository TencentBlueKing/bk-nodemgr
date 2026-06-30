<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="640"
    :title="$t('platform.nodeMan.agentStatus.agentUpgrade')"
    render-directive="if"
  >
    <template #default>
      <div class="py-[24px] px-[40px]">
        <Form :model="formData">
          <Form.FormItem
            property="bk_networkunit_id"
            :label="$t('platform.nodeMan.installAgentPage.cloud_unit')"
          >
            <Select
              class="w-[480px]"
              v-model="formData.bk_networkunit_id"
              clearable
              filterable
              :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.selectCloudUnit')"
              :loading="unitLoading"
            >
              <Select.Option
                v-for="option in networkUnitList"
                :key="option.bk_networkunit_id"
                :id="option.bk_networkunit_id"
                :name="option.bk_networkunit_name"
              >
                [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
              </Select.Option>
            </Select>
          </Form.FormItem>

          <Form.FormItem
            :label="$t('platform.nodeMan.installAgentPage.version')"
            required
          >
            <div class="w-[480px]">
              <Table
                :data="systemData"
                :border="true"
                width="480"
                :empty-text="$t('platform.nodeMan.installAgentPage.noAvailableVersion')"
              >
                <TableColumn
                  field="os"
                  :title="$t('platform.nodeMan.installAgentPage.osArch')"
                  width="160"
                >
                  <template #header>
                    <span class="text-[14px]">{{ $t('platform.nodeMan.installAgentPage.osArch') }}</span>
                  </template>
                  <template #default="{ row }">
                    {{ row.os?.replace('_', '/') }}
                  </template>
                </TableColumn>
                <TableColumn
                  field="version"
                  :title="$t('platform.nodeMan.installAgentPage.packageVersion')"
                  width="320"
                >
                  <template #header>
                    <span class="mr-[2px] text-[14px]">
                      {{ $t('platform.nodeMan.installAgentPage.packageVersion') }}
                    </span>
                    <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                    <Button text @click="handleBatchEditVersion">
                      <i class="nodeman-icon nc-bulk-edit cursor-pointer text-[14px]"></i>
                    </Button>
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

          <Form.FormItem :label="$t('platform.nodeMan.agentNodeStatus.forceUpgrade')">
            <Switcher v-model="formData.force" theme="primary" />
          </Form.FormItem>

          <Form.FormItem
            property="graceful_restart_timeout_sec"
            :label="$t('platform.nodeMan.agentStatus.gracefulRestartTimeout')"
          >
            <Input
              class="w-[200px]"
              type="number"
              :min="0"
              v-model="formData.graceful_restart_timeout_sec"
            >
              <template #suffix>
                <span class="px-[8px] text-[#979ba5]">s</span>
              </template>
            </Input>
          </Form.FormItem>
        </Form>
      </div>
    </template>
    <template #footer>
      <div class="flex gap-[8px] pl-[16px]">
        <Button
          theme="primary"
          :disabled="systemData.length === 0"
          v-bk-tooltips="{
            content: $t('platform.nodeMan.installAgentPage.noAvailableVersionTip'),
            disabled: systemData.length > 0
          }"
          @click="handleSubmit"
        >
          {{ $t('platform.nodeMan.preview.button.setup') }}
          <span
            class="mx-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
          >
            {{ hosts.length }}
          </span>
        </Button>
        <Button @click="isShow = false">{{ $t('action.cancel') }}</Button>
      </div>
    </template>
  </Sideslider>

  <choose-version-dialog
    v-model:is-show="isShowDialog"
    :data="dialogData"
    :batch="isBatch"
    :release-type="releaseType"
    @confirm="handleConfirmVersion"
  ></choose-version-dialog>
</template>
<script lang="ts" setup>
import { Button, Form, Input, Select, Sideslider, Switcher } from 'bkui-vue';
import { computed, ref, watch } from 'vue';

import { Table, TableColumn } from '@blueking/table';

import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { PACKAGE_GENERATION } from '@/common/const';
import Validate from '@/components/validate.vue';

const isShow = defineModel('isShow', { type: Boolean });

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

const emit = defineEmits<{
  (e: 'preview', formData: {
    networkUnitId?: number;
    targetVersions: Array<{ os_type: string; cpu_arch: string; version: string }>;
    force: boolean;
    gracefulRestartTimeoutSec: number;
  }): void;
}>();

const formData = ref({
  bk_networkunit_id: undefined as number | undefined,
  force: false,
  graceful_restart_timeout_sec: 0,
});

const networkUnitList = ref<NetworkUnitBrief[]>([]);
const unitLoading = ref(false);

const bk_networkarea_id = computed(() => {
  if (!props.hosts.length) return 0;
  return (props.hosts[0] as any).bk_networkarea_id
    ?? (props.hosts[0] as any).info?.bk_networkarea_id
    ?? 0;
});

const getNetworkUnitList = async () => {
  const areaId = bk_networkarea_id.value;
  if (!areaId) return;
  unitLoading.value = true;
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: {
      bk_networkarea_id: [areaId],
    },
  }).catch(() => ({ total: 0, items: [] }));
  unitLoading.value = false;
  networkUnitList.value = res.items;
};

const systemData = ref<Array<{ os: string; version: string }>>([]);

const getVersions = async () => {
  const listFn = props.releaseType === 'proxy'
    ? PackageService.ListReleaseProxyBrief
    : PackageService.ListReleaseAgentBrief;
  const res = await listFn({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      release_type: [props.releaseType],
    },
  }).catch(() => ({ total: 0, items: [] }));
  const osMap: Record<string, { name: string; enableVersions: any[] }> = {};
  res.items.forEach((item: any) => {
    const key = `${item.os_type}_${item.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = { name: key, enableVersions: [] };
    }
    if (item.enabled) {
      osMap[key].enableVersions.push({
        version: item.version,
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
      });
    }
  });
  systemData.value = Object.keys(osMap)
    .filter(key => osMap[key].enableVersions.length > 0)
    .map(key => ({ os: key, version: '' }));
};

const isShowDialog = ref(false);
const dialogData = ref([{ os: '', version: '' }]);
const isBatch = ref(false);
const inputRefs = ref<Map<string, any>>(new Map());

const setInputRef = (os: string, el: any) => {
  if (el) inputRefs.value.set(os, el);
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
  systemData.value.forEach((sys) => {
    const find = data.find((item) => sys.os === `${item.os_type}_${item.cpu_arch}`);
    if (find) sys.version = find.version;
  });
  isBatch.value = false;
};

const systemValidate = async () => {
  const refs = Array.from(inputRefs.value.values());
  const result = await Promise.all(refs.map(r => r.validate('blur')));
  return result.every(Boolean);
};

const handleSubmit = async () => {
  const versionValid = await systemValidate();
  if (!versionValid) return;

  const targetVersions = systemData.value
    .filter(item => !!item.version)
    .map((item) => {
      const [os_type, cpu_arch] = item.os.split('_');
      return { os_type, cpu_arch, version: item.version };
    });

  emit('preview', {
    networkUnitId: formData.value.bk_networkunit_id,
    targetVersions,
    force: formData.value.force,
    gracefulRestartTimeoutSec: formData.value.graceful_restart_timeout_sec,
  });
};

watch(
  () => isShow.value,
  async (val) => {
    if (val) {
      formData.value = { bk_networkunit_id: undefined, force: false, graceful_restart_timeout_sec: 0 };
      systemData.value = [];
      inputRefs.value.clear();
      await Promise.all([getNetworkUnitList(), getVersions()]);
    }
  },
);
</script>
