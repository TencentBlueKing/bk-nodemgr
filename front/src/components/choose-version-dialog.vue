<template>
  <Dialog
    :is-show="isShow"
    :width="1048"
    :title="title"
    @closed="isShow = false"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="flex h-[489px]">
      <!-- OS List -->
      <div class="w-[275px] bg-[#F5F7FA]">
        <div
          v-for="os in osVersions"
          :key="os.name"
          @click="selectOs(os)"
          :class="{ 'bg-[#E1ECFF]': os.selected }"
          class="flex items-center justify-between w-full px-[12px] hover:bg-[#E1ECFF] h-[36px] cursor-pointer"
        >
          <div>
            <i :class="[os.icon, 'mr-[6px]', { 'text-[#3A84FF]': os.selected }]"></i>
            <span :class="{ 'text-[#3A84FF]': os.selected }">{{
              os.name.replace('_', '/')
            }}</span>
          </div>
          <template v-if="os.selectedVersion?.version">
            <Tag v-if="os.selected" theme="info" type="filled">{{
              os.selectedVersion.version
            }}</Tag>
            <Tag v-else type="filled">{{
              os.selectedVersion.version
            }}</Tag>
          </template>
        </div>
      </div>

      <!-- Agent Version -->
      <div class="w-[280px] ml-[11px]">
        <Table
          :data="selectedOs?.versions"
          :empty-text="'暂无数据'"
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
            :title="t('Agent 版本')"
            min-width="130"
            sortable
          ></TableColumn>
          <TableColumn
            field="tag"
            min-width="80"
          >
            <template #default="{ row }">
              <Tag v-if="row.as_default">默认版本</Tag>
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
          <span v-if="selectedVersion?.version">{{ selectedVersion?.version }} 的详细信息</span>
        </div>
        <p class="text-[12px] border border-t-none h-full p-[16px]">
          {{ selectedVersion?.description }}
        </p>
      </div>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
import { Button, Dialog, Radio, Tag } from 'bkui-vue';
import { RightShape } from 'bkui-vue/lib/icon';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import { PackageService } from '@/api/modules/pkg';
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
});
const emit = defineEmits(['confirm', 'cancel']);
const { t } = useI18n();

const mainStore = useMainStore();

const osVersions = ref<IOsversion[]>();
const title = computed(() => props.title || t('components.chooseVersion.title'));
const selectedOs = ref();
const selectedVersion = ref<any>();
const selectedRadio = computed(() => selectedVersion.value.version || '');

function selectOs(os: IOsversion) {
  if (!props.batch) return;
  osVersions.value?.forEach((o: IOsversion) => (o.selected = false));
  os.selected = true;
  selectedOs.value = os;
  selectedVersion.value = os.selectedVersion || os.versions[0];
}

const handleChange = (val: string) => {
  selectedVersion.value = selectedOs.value?.versions.find((item: any) => item.version === val);
  const findOs = osVersions.value?.find(item => item.name === selectedOs.value.name);
  findOs.selectedVersion = selectedVersion.value;
};
function handleConfirm() {
  emit('confirm', osVersions.value?.map(item => ({
    version: item.selectedVersion.version,
    os_type: item.selectedVersion.os_type,
    cup_arch: item.selectedVersion.cpu_arch,
  })));
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
    res = await PackageService.ListReleaseAgent({
      page: { limit: 500, offset: 0 },
      generation: 2,
      exact_include_conditions: {
        release_type: [props.releaseType],
        enabled: [true],
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));
  } else {
    res = await PackageService.ListReleaseProxy({
      page: { limit: 500, offset: 0 },
      generation: 2,
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
    .filter(item => filterOs?.some(os => `${item.release.os_type}_${item.release.cpu_arch}`.includes(os)))
    .forEach((item) => {
      const key = `${item.release.os_type}_${item.release.cpu_arch}`;
      const iconType = item.release.os_type === 'darwin' ? 'macos' : item.release.os_type;
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
        as_default: item.release.as_default,
        version: item.release.version,
        disabled: !item.release.enabled,
        lable: item.release.labels,
        packages: [item.release.file_name],
        os_type: item.release.os_type,
        cpu_arch: item.release.cpu_arch,
        description: mainStore.curLanguage === 'zh-CN' ? item.change_log_zh : item.change_log_en,
      };
      osMap[key].versions.push(versionObj);
      if (item.release.as_default) {
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
        selectedVersion.value = selectedOs.value?.versions.find(item => item.version === props.data?.[0]?.version);
      } else {
        selectedVersion.value = selectedOs.value?.versions.find(item => item.as_default) || selectedOs.value?.versions[0];
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
