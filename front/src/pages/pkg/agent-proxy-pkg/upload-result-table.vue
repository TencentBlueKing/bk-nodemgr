<template>
  <div>
    <div class="text-[12px] text-[#4D4F56] mb-[8px]">结果预览</div>
    <bk-loading title="数据解析中" :loading="loading">
      <Table
        :data="tableData"
      >
        <TableColumn
          field="name"
          :title="currentType === 'cert' ? '文件名' : '包名'"
          v-if="currentType !== 'bintool'"
          min-width="350"
        ></TableColumn>
        <TableColumn
          field="os_type"
          :title="'操作系统架构'"
          v-if="currentType !== 'cert'"
          min-width="150"
        >
          <template #default="{ row }">
            {{ `${row.os_type}_${row.cpu_arch}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="release_type"
          :title="'包类型'"
          v-if="!['cert', 'bintool'].includes(currentType)"
          min-width="70"
        ></TableColumn>
      </Table>
    </bk-loading>
    <template v-if="data?.change_log_zh || data?.change_log_en || data?.description">
      <div class="text-[12px] text-[#4D4F56] mt-[24px] mb-[8px] flex items-center gap-[16px]">
        <span>描述</span>
        <Radio.Group v-model="changLog" v-if="data?.change_log_zh || data?.change_log_en">
          <Radio.Button label="ZH"></Radio.Button>
          <Radio.Button label="EN"></Radio.Button>
        </Radio.Group>
      </div>
      <div
        class="w-full bg-[#FAFBFD] min-h-[60px] border
        border-[#DCDEE5] text-[#4D4F56] text-[12px] px-[10px] py-[6px] formatted-text">
        <span v-if="data?.change_log_zh || data?.change_log_en">
          {{ changLog === 'ZH' ? data.change_log_zh : data.change_log_en }}
        </span>
        <div v-else-if="currentType === 'plugin'" class="flex flex-col gap-[10px] flex-wrap">
          <span>描述信息：{{ data.description }}</span>
          <span>配置文件：{{ data.config_file }}</span>
          <span>配置格式：{{ data.config_format }}</span>
          <span>运行节点类型：{{ data.launch_node }}</span>
        </div>
      </div>
    </template>
  </div>
</template>
<script lang="ts" setup>
import { Button, PopConfirm, Radio, Select } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { usePackageStore } from '@/stores/package';

const props = defineProps({
  data: {
    type: Object,
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});
const route = useRoute();
const packageStore = usePackageStore();
const tagList = computed(() => packageStore.tagList.map((tag: string) => ({
  value: tag,
  label: tag,
})));
const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
const changLog = ref('ZH');
const selectTag = ref<string[]>([]);
const tableData = ref();
const batchUpdateTag = async () => {
  tableData.value = tableData.value.map((item: any) => ({
    ...item,
    labels: [...selectTag.value],
  }));
};
watch(() => props.data, () => {
  if (['agent', 'proxy', 'plugin'].includes(currentType.value)) {
    tableData.value = props.data?.platforms?.map((item: Platform) => {
      const fileName = props.data.name
        .replace('all', `${item.os_type}_${item.cpu_arch}`)
        .replace('_origin', '')
        .replace(/\.tgz.*$/, '.tgz');

      return {
        ...item,
        name: fileName,
        release_type: currentType.value,
        labels: [],
        version: props.data.version,
      };
    }) ?? []; // 添加默认空数组
  } else if (currentType.value === 'cert') {
    tableData.value = props.data?.cert_files?.map((item: any) => ({
      ...item,
      name: item,
      release_type: currentType.value,
      labels: [],
      version: props.data.version,
    }));
  } else if (currentType.value === 'bintool') {
    tableData.value = (props.data?.agent_platforms || props.data?.proxy_platforms)?.map((item: Platform) => ({
      ...item,
      name: props.data.name.replace(
        'all',
        `${item.os_type}_${item.cpu_arch}`,
      ).replace(/\.tgz.*$/, '.tgz'),
      release_type: currentType.value,
      labels: [],
      version: props.data.version,
    }));
  }
}, { immediate: true, deep: true });
</script>
<style scoped>
.formatted-text {
  white-space: pre-wrap;
}
</style>
