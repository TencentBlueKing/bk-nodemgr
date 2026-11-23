<template>
  <ReleaseNote
    v-model:show="isShow"
    :list="logList"
    :active="currentVersion"
    :current="currentVersionTag"
    :loading="detailLoading"
    :detail="currentDetail"
    :min-left-width="240"
  >
    <template #list>
      <div class="version-log-modes">
        <Button
          :theme="mode === 'overview' ? 'primary' : 'default'"
          :aria-pressed="mode === 'overview'"
          size="small"
          @click="handleModeChange('overview')"
        >
          {{ t('components.versionLog.overview') }}
        </Button>
        <Button
          :theme="mode === 'changelog' ? 'primary' : 'default'"
          :aria-pressed="mode === 'changelog'"
          size="small"
          @click="handleModeChange('changelog')"
        >
          {{ t('components.versionLog.changelog') }}
        </Button>
      </div>
      <button
        v-for="log in logList"
        :key="log.title"
        type="button"
        class="bk-release-note-version-item version-log-item"
        :class="{ 'is-active': currentVersion === log.title }"
        :aria-pressed="currentVersion === log.title"
        @click="handleSelected(log.title)"
      >
        <span v-if="log.isCurrent" class="bk-release-note-crt-tag">
          {{ t('components.versionLog.current') }}
        </span>
        <span class="bk-release-note-version-item-title">{{ log.title }}</span>
        <span v-if="log.date" class="bk-release-note-version-item-subtitle">{{ log.date }}</span>
      </button>
      <div v-if="!detailLoading && !logList.length && status !== 'list-error'" class="version-log-list-empty">
        {{ t('table.empty') }}
      </div>
    </template>
    <template v-if="status" #detail>
      <div class="version-log-state" role="status">
        <div v-if="currentVersion" class="version-log-state-version">{{ currentVersion }}</div>
        <div>{{ stateText }}</div>
        <Button v-if="status !== 'empty'" text theme="primary" @click="handleRetry">
          {{ t('components.versionLog.retry') }}
        </Button>
      </div>
    </template>
  </ReleaseNote>
</template>

<script setup lang="ts">
import { Button } from 'bkui-vue';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import ReleaseNote from '@blueking/release-note';

import type { ILog, IVersionLogsListResponse } from './log-version-helper';
import { mapVersionLogs, resolveCurrentIndex } from './log-version-helper';

import '@blueking/release-note/vue3/vue3.css';

type LogMode = 'overview' | 'changelog';
type LogStatus = '' | 'empty' | 'list-error' | 'detail-error';

const isShow = defineModel('isShow', { type: Boolean });
const { t } = useI18n();

const mode = ref<LogMode>('overview');
const logList = ref<ILog[]>([]);
const currentVersion = ref('');
const currentVersionTag = computed(() => logList.value.find(log => log.isCurrent)?.title || '');
const currentDetail = ref('');
const detailLoading = ref(false);
const status = ref<LogStatus>('');
let requestId = 0;

const stateText = computed(() => {
  if (status.value === 'list-error') return t('components.versionLog.listFailed');
  if (status.value === 'detail-error') return t('components.versionLog.detailFailed');
  return t(mode.value === 'overview' ? 'components.versionLog.emptyOverview' : 'components.versionLog.emptyChangelog');
});

const fetchData = async <T,>(url: string, allowNotFound = false): Promise<T | null> => {
  const fullUrl = new URL(url, location.origin);
  const response = await fetch(fullUrl.toString(), {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  });
  const body = await response.json();
  // RecordNotFound is the API's explicit missing-document response, not a network failure.
  if (allowNotFound && body.code === 3800003) return null;
  if (!response.ok || body.code !== 0) throw new Error(`Version log request failed: ${response.status}, ${body.code}`);
  return body;
};

const loadDetail = async (version: string) => {
  requestId += 1;
  const id = requestId;
  currentVersion.value = version;
  currentDetail.value = '';
  status.value = '';
  const log = logList.value.find(item => item.title === version);
  if (!log) {
    status.value = 'empty';
    detailLoading.value = false;
    return;
  }
  if (log.detail) {
    currentDetail.value = log.detail;
    if (!log.detail.trim()) status.value = 'empty';
    detailLoading.value = false;
    return;
  }
  detailLoading.value = true;
  const endpoint = mode.value === 'overview' ? 'overview' : 'changelog';
  try {
    const data = await fetchData<{ data?: { content?: string } }>(
      `/api/v3/version_log/${endpoint}/${encodeURIComponent(version)}`,
      true,
    );
    if (id !== requestId) return;
    const rawMarkdown = data?.data?.content || '';
    log.detail = rawMarkdown;
    currentDetail.value = rawMarkdown;
    if (!rawMarkdown.trim()) status.value = 'empty';
  } catch (error) {
    if (id !== requestId) return;
    console.error('Version log detail error:', error);
    status.value = 'detail-error';
  } finally {
    if (id === requestId) detailLoading.value = false;
  }
};

const getVersionLogsList = async () => {
  requestId += 1;
  const id = requestId;
  logList.value = [];
  currentVersion.value = '';
  currentDetail.value = '';
  status.value = '';
  detailLoading.value = true;
  const endpoint = mode.value === 'overview' ? 'overview_list' : 'version_logs_list';
  try {
    const data = await fetchData<IVersionLogsListResponse>(`/api/v3/version_log/${endpoint}/`);
    if (id !== requestId) return;
    logList.value = mapVersionLogs(data || {});
    const version = mode.value === 'overview'
      ? data?.data?.current_version || ''
      : logList.value[resolveCurrentIndex(logList.value)]?.title || '';
    await loadDetail(version);
  } catch (error) {
    if (id !== requestId) return;
    console.error('Version log list error:', error);
    status.value = 'list-error';
  } finally {
    if (id === requestId) detailLoading.value = false;
  }
};

const handleSelected = (version: string) => {
  loadDetail(version);
};

const handleModeChange = (value: LogMode) => {
  if (mode.value === value) return;
  mode.value = value;
  getVersionLogsList();
};

const handleRetry = () => {
  if (status.value === 'list-error') {
    getVersionLogsList();
  } else {
    loadDetail(currentVersion.value);
  }
};

watch(isShow, (value) => {
  requestId += 1;
  if (value) {
    mode.value = 'overview';
    getVersionLogsList();
  }
}, { flush: 'sync', immediate: true });

onBeforeUnmount(() => {
  requestId += 1;
  isShow.value = false;
});
</script>

<style>
.version-log-modes {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px;
  background: #fff;
}
.version-log-item {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  cursor: pointer;
}
.version-log-item > span {
  display: block;
}
.version-log-list-empty,
.version-log-state {
  padding: 32px 16px;
  color: #979ba5;
  text-align: center;
}
.version-log-state-version {
  margin-bottom: 12px;
  color: #63656e;
}
.version-log-state .bk-button {
  margin-top: 12px;
}
.bk-release-note .bk-release-note-detail-content > :first-child {
  margin-top: 0 !important;
}
.bk-release-note .bk-release-note-detail-content a {
  color: #3a84ff;
  text-decoration: none;
}
.bk-release-note .bk-release-note-detail-content a:hover {
  text-decoration: underline;
}
</style>
