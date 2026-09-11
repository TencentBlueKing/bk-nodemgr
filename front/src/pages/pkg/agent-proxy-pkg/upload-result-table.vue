<template>
  <div>
    <div class="text-[12px] text-[#4D4F56] mb-[8px]">{{ t('pkgUpload.resultPreview') }}</div>
    <bk-loading :title="t('pkgUpload.parsing')" :loading="loading">
      <Table
        v-if="currentType !== 'plugin_bintool'"
        :data="tableData"
      >
        <TableColumn
          field="name"
          :title="currentType === 'cert' ? t('pkgUpload.fileName') : t('pkgUpload.packageNameLabel')"
          v-if="!['bintool', 'plugin_bintool'].includes(currentType)"
          min-width="350"
        ></TableColumn>
        <TableColumn
          field="os_type"
          :title="t('pkgUpload.osArch')"
          v-if="currentType !== 'cert'"
          min-width="150"
        >
          <template #default="{ row }">
            {{ `${row.os_type}_${row.cpu_arch}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="release_type"
          :title="t('pkgUpload.packageType')"
          v-if="!['cert', 'bintool', 'plugin_bintool'].includes(currentType)"
          min-width="70"
        ></TableColumn>
      </Table>
      <template v-else>
        <div v-for="item in plugin_bintool_list" :key="item.name" class="mb-[10px]">
          <div>{{ item.name }}</div>
          <Table
            :data="item.platforms"
          >
            <TableColumn
              field="name"
              :title="t('pkgUpload.packageNameLabel')"
              v-if="!['bintool', 'plugin_bintool'].includes(currentType)"
              min-width="350"
            ></TableColumn>
            <TableColumn
              field="os_type"
              :title="t('pkgUpload.osArch')"
              min-width="150"
            >
              <template #default="{ row }">
                {{ `${row.os_type}_${row.cpu_arch}` }}
              </template>
            </TableColumn>
            <TableColumn
              field="release_type"
              :title="t('pkgUpload.packageType')"
              v-if="!['cert', 'bintool', 'plugin_bintool'].includes(currentType)"
              min-width="70"
            ></TableColumn>
          </Table>
        </div>
      </template>
    </bk-loading>
    <template
      v-if="data?.change_log_zh || data?.change_log_en || data?.description || data?.description_en
        || data?.scenario || data?.scenario_en"
    >
      <div class="text-[12px] text-[#4D4F56] mt-[24px] mb-[8px] flex items-center gap-[16px]">
        <span>{{ t('pkgUpload.description') }}</span>
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
          <span>{{ t('pkgUpload.descInfo', {
            desc: locale.startsWith('zh') ? data.description : data.description_en,
          }) }}</span>
          <span>
            {{ t('pkgUpload.scenario', {
              scenario: locale.startsWith('zh') ? data.scenario : data.scenario_en,
            }) }}
          </span>
          <span v-if="!isV3Plugin">{{ t('pkgUpload.configFile', { file: data.config_file }) }}</span>
          <span v-if="!isV3Plugin">{{ t('pkgUpload.configFormat', { format: data.config_format }) }}</span>
          <span>{{ t('pkgUpload.launchNode', { node: data.launch_node }) }}</span>
        </div>
      </div>
    </template>
  </div>
</template>
<script lang="ts" setup>
import { Button, PopConfirm, Radio, Select } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { usePackageStore } from '@/stores/package';

const { t, locale } = useI18n();
const props = defineProps({
  data: {
    type: Object,
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  // 插件包类型：v3 标准插件不展示「配置文件/配置格式」（v3 响应里就没有这俩字段）
  pluginUploadType: {
    type: String,
    default: '',
  },
});
const isV3Plugin = computed(() => props.pluginUploadType === 'v3/plugin');
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
const tableData = ref<any[]>([]);
const plugin_bintool_list = ref<any[]>([]);
watch(() => props.data, () => {
  if (['agent', 'proxy', 'plugin'].includes(currentType.value)) {
    tableData.value = props.data?.platforms?.map((item: Platform) => {
      const fileName = props.data.name
        .replace('all', `${item.os_type}_${item.cpu_arch}`)
        .replace('origin_server', 'proxy')
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
  } else if (currentType.value === 'plugin_bintool') {
    const v2 = props.data?.v2?.platforms.map((item: Platform) => ({
      ...item,
    })) || [];
    const v3 = props.data?.v3?.platforms.map((item: Platform) => ({
      ...item,
    })) || [];
    plugin_bintool_list.value = [
      { name: 'v2', platforms: v2 },
      { name: 'v3', platforms: v3 },
    ];
  }
}, { immediate: true, deep: true });
</script>
<style scoped>
.formatted-text {
  white-space: pre-wrap;
}
</style>
