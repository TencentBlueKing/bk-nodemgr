<template>
  <div class="p-[24px] h-full flex flex-col">
    <Alert theme="info" class="mb-[16px]" :title="$t('globalSettings.storedValueNotice')" />
    <div class="flex items-center justify-between gap-[16px] mb-[16px]">
      <Button :loading="loading" @click="getSettings">{{ $t('globalSettings.refresh') }}</Button>
      <SearchSelect
        v-model="searchValues"
        class="w-[640px] bg-white"
        :data="searchOptions"
        :unique-select="true"
        :placeholder="$t('globalSettings.searchPlaceholder')"
      />
    </div>
    <Alert v-if="failed" theme="danger" class="mb-[16px]" :title="$t('globalSettings.loadFailed')" />
    <Loading :loading="loading" :title="$t('table.loading')" class="flex-1 overflow-auto">
      <Table
        :data="settings"
        :pagination="pagination"
        :max-height="maxHeight"
        :empty-text="failed ? $t('globalSettings.loadFailed') : $t('table.empty')"
        @page-limit-change="handleLimitChange"
        @page-value-change="handlePageChange"
      >
        <TableColumn field="setting_name" :title="$t('globalSettings.name')" min-width="320" show-overflow-tooltip />
        <TableColumn field="value" :title="$t('globalSettings.value')" min-width="360">
          <template #default="{ row }">
            <span class="block truncate">{{ row.value === '' ? $t('globalSettings.emptyString') : row.value }}</span>
          </template>
        </TableColumn>
        <TableColumn :title="$t('table.action')" width="160" fixed="right">
          <template #default="{ row }">
            <Button theme="primary" text @click="showValue(row)">{{ $t('globalSettings.viewValue') }}</Button>
          </template>
        </TableColumn>
      </Table>
    </Loading>
    <Sideslider
      v-model:is-show="valueVisible"
      :title="$t('globalSettings.viewValue')"
      :width="960"
      render-directive="if"
    >
      <div v-if="selectedSetting" class="p-[24px]">
        <div class="flex items-start justify-between gap-[16px] mb-[16px]">
          <h3 class="m-0 text-[14px] break-all">{{ selectedSetting.setting_name }}</h3>
          <Button class="shrink-0" @click="copyValue">{{ $t('globalSettings.copyValue') }}</Button>
        </div>
        <p class="text-[12px] text-[#979ba5]">{{ $t('globalSettings.rawValueNotice') }}</p>
        <pre class="raw-value">{{ selectedSetting.value === ''
          ? $t('globalSettings.emptyString') : selectedSetting.value }}</pre>
      </div>
    </Sideslider>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading, Message, SearchSelect, Sideslider } from 'bkui-vue';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';
import { useClipboard } from '@vueuse/core';

import type { GlobalSetting } from '@/@types/globalsettings';
import { GlobalSettingsService } from '@/api/modules/globalsettings';
import { useMainStore } from '@/stores/main';

interface SearchValue {
  id: 'setting_name';
  name: string;
  values: { id: string; name: string }[];
}

const { t } = useI18n();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 280 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const settings = ref<GlobalSetting[]>([]);
const loading = ref(false);
const failed = ref(false);
const searchValues = ref<SearchValue[]>([]);
const searchOptions = computed(() => [{ id: 'setting_name', name: t('globalSettings.name'), multiple: true }]);
const valueVisible = ref(false);
const selectedSetting = ref<GlobalSetting>();
const { copy } = useClipboard({ legacy: true });
let requestId = 0;

const getSettings = async () => {
  requestId += 1;
  const currentRequest = requestId;
  loading.value = true;
  failed.value = false;
  try {
    const names = searchValues.value.flatMap(item => item.values?.map(value => value.id) || []);
    const res = await GlobalSettingsService.ListGlobalSettings({
      page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
      exact_include_conditions: { setting_name: names },
    });
    if (currentRequest !== requestId) return;
    settings.value = res.items || [];
    pagination.count = res.total;
  } catch {
    if (currentRequest !== requestId) return;
    settings.value = [];
    pagination.count = 0;
    failed.value = true;
  } finally {
    if (currentRequest === requestId) loading.value = false;
  }
};
const handleLimitChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
  getSettings();
};
const handlePageChange = (page: number) => {
  pagination.current = page;
  getSettings();
};
const showValue = (setting: GlobalSetting) => {
  selectedSetting.value = setting;
  valueVisible.value = true;
};
const copyValue = async () => {
  if (!selectedSetting.value) return;
  try {
    await copy(selectedSetting.value.value);
    Message({ theme: 'success', message: t('globalSettings.copySuccess') });
  } catch {
    Message({ theme: 'error', message: t('globalSettings.copyFailed') });
  }
};
watch(searchValues, () => {
  pagination.current = 1;
  getSettings();
}, { deep: true, immediate: true });
onBeforeUnmount(() => {
  requestId += 1;
});
</script>

<style scoped>
.raw-value {
  padding: 16px;
  margin: 0;
  min-height: 240px;
  color: #313238;
  background: #f5f7fa;
  font-size: 12px;
  line-height: 20px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
