<template>
  <Dialog
    :is-show="isShow"
    :width="1048"
    title="按操作系统选定版本"
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
            <i :class="[os.icon, 'mr-[6px]']"></i>
            <span :class="{ 'text-[#3A84FF]': os.selected }">{{
              os.name
            }}</span>
          </div>
          <template v-if="os.version">
            <Tag v-if="os.selected" theme="info" type="filled">{{
              os.version
            }}</Tag>
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
                <Radio
                  :label="row.version"
                  v-model="selectedRadio"
                  @change="handleChange"
                />
              </div>
            </template>
          </TableColumn>
          <TableColumn
            field="version"
            :title="t('Agent 版本')"
            sortable
          ></TableColumn>
          <TableColumn fixed="right" width="50">
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
          {{ selectedVersion?.version }} 的详细信息
        </div>
        <p class="text-[12px] border border-t-none h-full p-[16px]">
          {{ selectedVersion?.description }}
        </p>
      </div>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
interface RowVO {
  name: string;
  role: string;
  num: number;
}

import { RightShape } from "bkui-vue/lib/icon";
import { Button, Dialog, Tag, Radio } from "bkui-vue";
import { PackageService } from "@/api/modules/pkg";
import { ref, watch } from "vue";
import { Table, TableColumn } from "@blueking/table";
import { useI18n } from "vue-i18n";
import type { VxeTablePropTypes } from "vxe-table";
import { compareVersions, capitalizeFirstLetter } from "@/common/util";

const props = defineProps({
  data: {
    type: Object,
    default: {
      os: "",
      version: "",
    },
  },
  releaseType: {
    type: String,
    default: "agent",
  },
});
const isShow = defineModel("isShow", { type: Boolean, default: false });
const emit = defineEmits(["confirm", "cancel"]);
const { t } = useI18n();

const osVersions = ref<{
  name: string;
  version: string;
  selected: boolean;
  icon: string;
}>();

const selectedOs = ref();
const selectedVersion = ref<{
  version: string;
  description: string;
  disabled: boolean;
  label: Object;
}>();
const selectedRadio = ref("");

function selectOs(os) {
  if (props.data?.os) return;
  osVersions.value.forEach((o) => (o.selected = false));
  os.selected = true;
  selectedOs.value = os;
  selectedVersion.value = os.versions[0];
}

const handleChange = (val: string) => {
  selectedVersion.value = selectedOs.value?.versions.find(
    (item: any) => item.version === val
  );
};
function handleConfirm() {
  emit("confirm", selectedRadio.value);
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
    if (field === "version") {
      list = data.sort((a, b) => {
        return order === "desc"
          ? compareVersions(b.version, a.version)
          : compareVersions(a.version, b.version);
      });
    }
    return list;
  },
});

const getVersions = async () => {
  const res = await PackageService.ListRelease({
    generation: [2],
    release_type: props.releaseType,
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const osMap: any = {};
  res.items.forEach((item) => {
    const key = `${capitalizeFirstLetter(item.os_type)}_${item.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        version: item.as_default ? item.version : "",
        selected: false,
        versions: props.releaseType === 'proxy' ? [{
          version: '默认',
          disabled: false,
          lable: [],
          packages: [],
          description: '',
        }] : [],
        icon: `nodeman-icon nc-${item.os_type}`,
      };
    }

    osMap[key].versions.push({
      version: item.version,
      disabled: !item.enabled,
      lable: item.labels,
      packages: [item.file_name],
      description: item.change_log_zh,
    });
  });
  osVersions.value = Object.values(osMap);
};
watch(
  () => isShow,
  async () => {
    if (isShow.value) {
      await getVersions();
      if (props.data?.os) {
        const os = capitalizeFirstLetter(props.data.os);
        selectedOs.value = osVersions.value.find((item) => item.name === os);
      } else {
        selectedOs.value = osVersions.value[0];
      }
      if (props.data?.version) {
        selectedVersion.value = selectedOs.value?.versions.find(
          (item) => item.version === props.data?.version
        );
      } else {
        selectedVersion.value = selectedOs.value?.versions[0];
      }
      selectedOs.value && (selectedOs.value.selected = true);
      selectedRadio.value = selectedVersion.value?.version || "";
    }
  },
  { immediate: true, deep: true }
);
</script>
