<template>
  <Dialog
    :is-show="isShow"
    :width="1048"
    title="按操作系统选定版本"
    @closed="isShow = false"
    @confirm="handleConfirm"
    @cancel="handleCancel">
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
            <i :class="[os.icon, 'mr-[6px]']"></i>
            <span :class="{'text-[#3A84FF]': os.selected }">{{ os.name }}</span>
          </div>
          <template v-if="os.version">
            <Tag v-if="os.selected" theme="info" type="filled">{{ os.version }}</Tag>
            <Tag v-else type="filled">{{ os.version }}</Tag>
          </template>
        </div>
      </div>

      <!-- Agent Version -->
      <div class="w-[227px] ml-[11px]">
        <Table
          :data="selectedOs?.versions"
          :empty-text="'暂无数据'"
          :sort-config="sortConfig"
        >
          <TableColumn width="34">
            <template #default="{ row }">
              <div class="flex items-center">
                <Radio :label="row.version" v-model="selectedRadio" @change="handleChange"/>
              </div>
            </template>
          </TableColumn>
          <TableColumn field="version" :title="t('Agent 版本')" sortable></TableColumn>
          <TableColumn fixed="right" width="50">
            <template #default="{ row }">
              <div class="flex items-center">
                <right-shape v-if="row.version === selectedVersion?.version"/>
              </div>
            </template>
          </TableColumn>
        </Table>
      </div>

      <!-- Version Details -->
      <div class="flex-1">
        <div class="text-[14px] bg-[#FAFBFD] border border-l-none border-[#DCDEE5] h-[40.69px] leading-[40.69px] pl-[24px]">
          {{ selectedVersion?.version }} 的详细信息
        </div>
        <p class="text-[12px] border border-t-none h-full">{{ selectedVersion?.description }}</p>
      </div>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
interface RowVO {
  name: string
  role: string
  num: number
  num1: string
  num2: string
}

import { RightShape } from 'bkui-vue/lib/icon';
import { Button, Dialog, Tag, Radio } from 'bkui-vue';
import { PackageService } from '@/api/modules/pkg';
import {ref,  watch } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { useI18n } from 'vue-i18n';
import type { VxeTablePropTypes } from 'vxe-table';

const props = defineProps({
  data: {
    type: Object,
    default: {
      os: '',
      version: ''
    }
  }
});
const isShow = defineModel('isShow', { type: Boolean, default: false });
const emit = defineEmits(['confirm', 'cancel']);
const { t } = useI18n();

const osVersions = ref<{name: string, version: string, selected: boolean, icon: string}>();

const selectedOs = ref();
const selectedVersion = ref<{version: string, description: string, disabled: boolean, label: Object}>();
const selectedRadio = ref('');

function selectOs(os) {
  if (props.data?.os) return;
  osVersions.value.forEach(o => o.selected = false);
  os.selected = true;
  selectedOs.value = os;
  selectedVersion.value = os.versions[0];
}

const handleChange = (val: string) => {
  selectedVersion.value = selectedOs.value?.versions.find(item => item.version === val);
}
function handleConfirm() {
  emit('confirm', selectedRadio.value);
  isShow.value = false;
}

function handleCancel() {
  isShow.value = false;
}
const capitalizeFirstLetter = (str: string) => {
  if (!str) return str; // 处理空字符串的情况
  return str.charAt(0).toUpperCase() + str.slice(1);
}
const sortConfig = ref<VxeTablePropTypes.SortConfig<RowVO>>({
  sortMethod ({ data, sortList }) {
    const sortItem = sortList[0]
    // 取出第一个排序的列
    const { field, order } = sortItem
    let list: RowVO[] = []
    if (order === 'desc') {
      if (field === 'version') {
        list = data.sort((a, b) => {
          return order === 'ascending'
            ? compareVersions(a.version,b.version)
            : compareVersions(b.version,a.version);
        });
      }
    }
    return list
  }
});
const compareVersions = (a: string, b: string) => {
  if (a === b) return 0;
  if (!a || !b) return a ? 1 : -1;
  // 解析版本号字符串，返回数字数组
  const parseVersion = (version: string) => version?.match(/\d+/g)?.map(Number);

  // 获取版本号的数字数组
  const versionA = parseVersion(a) || [];
  const versionB = parseVersion(b) || [];

  // 比较主版本号、次版本号、补丁号和附加编号
  for (let i = 0; i < versionA.length; i++) {
    const diff = versionA[i] - versionB[i];
    if (diff !== 0) return diff;
  }

  return 0; // 全部相同
}
const getVersions = async () => {
  const res = await PackageService.ListRelease({
    exact_include_conditions: {
      generation: [2],
      release_type: ['agent'],
    }
  }).catch(() => ({
    total: 0,
    items: []
  }));
  const osMap: any = {};
  res.items.forEach(item => {
    const key = `${capitalizeFirstLetter(item.os_type)}_${item.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        version: item.as_default ? item.version : '',
        selected: false,
        versions: [],
        icon: `nodeman-icon nc-${item.os_type}`
      };
    }

    osMap[key].versions.push({
      version: item.version,
      disabled: !item.enabled,
      lable: item.labels,
      packages: [item.file_name],
      description: item.change_log_zh
    });
  })
  osVersions.value = Object.values(osMap);
}
watch(() => isShow, async () => {
  if(isShow.value) {
    await getVersions();
    if (props.data?.os) {
      const os = capitalizeFirstLetter(props.data.os)
      selectedOs.value = osVersions.value.find(item => item.name === os);
    } else {
      selectedOs.value = osVersions.value[0];
    }
    if (props.data?.version) {
      selectedVersion.value = selectedOs.value?.versions.find(item => item.version === props.data?.version);
    } else {
      selectedVersion.value = selectedOs.value?.versions[0];
    }
    selectedOs.value.selected = true;
    selectedRadio.value = selectedVersion.value?.version || '';
  }
},{immediate: true, deep: true});
</script>
