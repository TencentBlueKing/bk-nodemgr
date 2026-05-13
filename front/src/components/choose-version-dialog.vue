<template>
  <Dialog
    :is-show="isShow"
    :width="1048"
    :title="title"
    @closed="isShow = false"
  >
    <div class="flex h-[489px]">
      <!-- OS List -->
      <div class="w-[275px] bg-[#F5F7FA] overflow-y-auto shrink-0">
        <div
          v-for="os in osVersions"
          :key="os.name"
          @click="selectOs(os)"
          :class="{ 'bg-[#E1ECFF]': os.selected }"
          class="flex items-center justify-between w-full px-[12px] hover:bg-[#E1ECFF] h-[36px] cursor-pointer"
        >
          <div class="flex items-center shrink-0 min-w-0">
            <i :class="[os.icon, 'mr-[6px]', { 'text-[#3A84FF]': os.selected }]"></i>
            <span :class="{ 'text-[#3A84FF]': os.selected }">{{
              os.name.replace('_', '/')
            }}</span>
          </div>
          <template v-if="os.selectedVersion?.version">
            <Tag
              v-bk-tooltips="{ content: os.selectedVersion.version, disabled: os.selectedVersion.version.length <= 12 }"
              :theme="os.selected ? 'info' : undefined"
              :type="'filled'"
              class="version-tag"
            >
              {{ os.selectedVersion.version }}
            </Tag>
          </template>
        </div>
      </div>

      <!-- Agent Version -->
      <div class="ml-[11px]">
        <Table
          :data="selectedOs?.versions"
          :empty-text="$t('table.empty')"
          :sort-config="sortConfig"
        >
          <TableColumn fixed="left" width="34">
            <template #default="{ row }">
              <div class="flex items-center">
                <Radio
                  :label="row.version"
                  :model-value="selectedRadio"
                  @change="handleChange"
                />
              </div>
            </template>
          </TableColumn>
          <TableColumn
            field="version"
            fixed="left"
            :min-width="hasDefaultVersion ? 138 : 230"
            sortable
          >
            <template #header>
              <span class="text-[14px]">{{ versionColumnTitle }}</span>
            </template>
            <template #default="{ row }">
              <span v-bk-tooltips="{ content: row.version, disabled: !row.version || row.version.length <= 15 }">
                {{ row.version }}
              </span>
            </template>
          </TableColumn>
          <TableColumn
            v-if="hasDefaultVersion"
            field="tag"
            min-width="100"
          >
            <template #default="{ row }">
              <Tag v-if="row.as_default">{{ $t('components.chooseVersion.DefaultVersion') }}</Tag>
            </template>
          </TableColumn>
          <TableColumn fixed="right" min-width="34">
            <template #default="{ row }">
              <div class="flex items-center">
                <right-shape v-if="row.version === selectedVersion?.version" />
              </div>
            </template>
          </TableColumn>
        </Table>
      </div>

      <!-- Version Details -->
      <div class="flex-1">
        <div
          class="text-[14px] bg-[#FAFBFD] border border-l-none border-[#DCDEE5] h-[40.69px] leading-[40.69px] pl-[24px]"
        >
          <span v-if="selectedVersion?.version">
            {{ $t('components.chooseVersion.versionInfoTitle', { version: selectedVersion?.version }) }}
          </span>
        </div>
        <p class="text-[12px] border border-t-none h-full p-[16px]">
          {{ selectedVersion?.description || '--' }}
        </p>
      </div>
    </div>
    <template #footer>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-[8px]" v-if="type === 'upgrade'">
          <Radio.Group v-model="force">
            <Radio.Button :label="true" :key="true">
              {{ $t('components.chooseVersion.forceUpgrade') }}
            </Radio.Button>
            <Radio.Button :label="false" :key="false">
              {{ $t('components.chooseVersion.gracefulUpgrade') }}
            </Radio.Button>
          </Radio.Group>
          <div class="flex items-center gap-[3px] ml-[20px]" v-show="!force">
            <span>{{ $t('components.operateDialog.gracefulTime') }}</span>
            <Input type="number" v-model="graceful_restart_timeout_sec" class="w-[80px] mx-[3px]"></Input>
            <span>{{ $t('components.operateDialog.seconds') }}</span>
          </div>
        </div>
        <div class="ml-auto">
          <Button class="mr-[8px]" theme="primary" @click="handleConfirm">{{ $t('action.confirm') }}</Button>
          <Button @click="handleCancel">{{ $t('action.cancel') }}</Button>
        </div>
      </div>
    </template>
  </Dialog>
</template>
<script lang="ts" setup>
import { Button, Dialog, Input, Radio, Tag } from 'bkui-vue';
import { RightShape } from 'bkui-vue/lib/icon';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';
import { compareVersions } from '@/common/util';
import { useMainStore } from '@/stores/main';

interface RowVO {
  name: string;
  role: string;
  num: number;
}

interface IVersion {
  version: string,
  os_type: string,
  cpu_arch: string,
}

interface IOsversion {
  name: string;
  version: string,
  versions: IVersion[];
  selected: boolean;
  selectedVersion: IVersion;
  icon: string;
}

const isShow = defineModel('isShow', { type: Boolean, default: false });
const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  data: {
    type: Array,
    default: [{
      os: '',
      version: '',
    }],
  },
  batch: {
    type: Boolean,
    default: false,
  },
  releaseType: {
    type: String,
    default: 'agent',
  },
  isCrossPageSelection: {
    type: Boolean,
    default: false,
  },
  type: {
    type: String,
    default: '',
  },
  pluginName: {
    type: String,
    default: '',
  },
});
const emit = defineEmits(['confirm', 'cancel']);
const { t } = useI18n();

const mainStore = useMainStore();

const osVersions = ref<IOsversion[]>();
const title = computed(() => props.title || t('components.chooseVersion.title'));
const selectedOs = ref();
const selectedVersion = ref<any>();
const selectedRadio = computed(() => selectedVersion.value.version || '');
const force = ref(false); // 是否强制升级
const graceful_restart_timeout_sec = ref(120);

// 当前 OS 下是否有默认版本标签，没有则隐藏 tag 列
const hasDefaultVersion = computed(() => selectedOs.value?.versions?.some((v: any) => v.as_default) ?? false);

// 根据 releaseType 动态显示版本列标题
const versionColumnTitle = computed(() => {
  const typeMap: Record<string, string> = {
    agent: t('components.chooseVersion.AgentVersion'),
    plugin: t('components.chooseVersion.PluginVersion'),
    proxy: t('components.chooseVersion.ProxyVersion'),
  };
  return typeMap[props.releaseType] || t('components.chooseVersion.AgentVersion');
});

function selectOs(os: IOsversion) {
  if (!props.batch) return;
  osVersions.value?.forEach((o: IOsversion) => { if (o) o.selected = false; });
  if (os) os.selected = true;
  selectedOs.value = os;
  selectedVersion.value = os?.selectedVersion || os?.versions?.[0];
}

const handleChange = (val: string) => {
  selectedVersion.value = selectedOs.value?.versions?.find((item: any) => item.version === val);
  const findOs = osVersions.value?.find(item => item.name === selectedOs.value?.name);
  if (findOs) findOs.selectedVersion = selectedVersion.value;
};
function handleConfirm() {
  emit('confirm', osVersions.value?.filter(item => item?.selectedVersion).map(item => ({
    version: item.selectedVersion.version,
    os_type: item.selectedVersion.os_type,
    cpu_arch: item.selectedVersion.cpu_arch,
  })), {
    force: force.value,
    graceful_restart_timeout_sec: graceful_restart_timeout_sec.value,
  });
  isShow.value = false;
}

function handleCancel() {
  isShow.value = false;
}

const sortConfig = ref<VxeTablePropTypes.SortConfig<RowVO>>({
  sortMethod({ data, sortList }) {
    const sortItem = sortList[0];
    // 取出第一个排序的列
    const { field, order } = sortItem;
    let list: [] = data;
    if (field === 'version') {
      list = data.sort((a, b) => (order === 'desc'
        ? compareVersions(b.version, a.version)
        : compareVersions(a.version, b.version)));
    }
    return list;
  },
});

const getVersions = async () => {
  let res;
  if (props.releaseType === 'agent') {
    res = await PackageService.ListReleaseAgentBrief({
      page: { limit: 500, offset: 0 },
      generation: PACKAGE_GENERATION,
      exact_include_conditions: {
        release_type: [props.releaseType],
        enabled: [true],
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));
  } else if (props.releaseType === 'plugin') {
    res = await PackageService.ListReleasePluginBrief({
      page: { limit: 500, offset: 0 },
      generation: PACKAGE_GENERATION,
      exact_include_conditions: {
        enabled: [true],
        ...(props.pluginName ? { name: [props.pluginName] } : {}),
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));
  } else {
    res = await PackageService.ListReleaseProxyBrief({
      page: { limit: 500, offset: 0 },
      generation: PACKAGE_GENERATION,
      exact_include_conditions: {
        release_type: [props.releaseType],
        enabled: [true],
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));
  }
  const osMap: any = {};
  const filterOs = props.isCrossPageSelection ? ['_'] : props.data.map((el: any) => (el.os ? el.os : `${el.os_type}_${el.cpu_arch}`));
  res.items
    .filter(item => filterOs?.some(os => `${item.os_type}_${item.cpu_arch}`.includes(os)))
    .forEach((item) => {
      const key = `${item.os_type}_${item.cpu_arch}`;
      const iconType = item.os_type === 'darwin' ? 'macos' : item.os_type;
      if (!osMap[key]) {
        osMap[key] = {
          name: key,
          version: '',
          selected: false,
          selectedVersion: '',
          versions: [],
          icon: `nodeman-icon nc-${iconType}`,
        };
      }
      const versionObj = {
        as_default: item.as_default,
        version: item.version,
        disabled: !item.enabled,
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
        description: mainStore.curLanguage === 'zh-CN' ? item.change_log_zh : item.change_log_en,
      };
      osMap[key].versions.push(versionObj);
      if (item.as_default) {
        osMap[key].selectedVersion = versionObj;
      }
    });
  osVersions.value = Object.values(osMap);
};
watch(
  () => isShow,
  async () => {
    if (isShow.value) {
      await getVersions();
      if (props.data?.[0]?.os && !props.batch) {
        const os = props.data?.[0].os || `${props.data?.[0].os_type}_${props.data?.[0].cpu_arch}`;
        selectedOs.value = osVersions.value?.find(item => item.name === os);
      } else {
        selectedOs.value = osVersions.value?.[0];
      }
      if (props.data?.[0]?.version && !props.batch) {
        selectedVersion.value = selectedOs.value?.versions?.find(item => item.version === props.data?.[0]?.version);
      } else {
        selectedVersion.value = selectedOs.value?.versions?.find(item => item.as_default) || selectedOs.value?.versions?.[0];
      }
      if (selectedOs.value) {
        selectedOs.value.selected = true;
        selectedOs.value.selectedVersion = selectedVersion.value;
      }
    }
  },
  { immediate: true, deep: true },
);
</script>

<style lang="postcss" scoped>
.version-tag {
  max-width: 120px;
  :deep(.bk-tag-text) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
