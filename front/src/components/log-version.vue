<template>
  <ReleaseNote
    v-model:show="isShow"
    :list="logList"
    :current="currentVersion"
    :loading="detailLoading"
    :detail="currentDetail"
    :min-left-width="240"
    @selected="handleSelected"
  />
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import ReleaseNote from '@blueking/release-note';
import '@blueking/release-note/vue3/vue3.css';

import { mapVersionLogs, resolveCurrentIndex } from './log-version-helper';

interface ILog {
  title: string;
  date: string;
  detail: string;
  isCurrent: boolean;
}

const isShow = defineModel('isShow', { type: Boolean });

const logList = ref<ILog[]>([]);
const currentVersion = ref('');
const currentDetail = ref('');
const detailLoading = ref(false);

const fetchData = async (url: string) => {
  try {
    const fullUrl = new URL(url, location.origin);
    const response = await fetch(fullUrl.toString(), {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
    });
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    return await response.json();
  } catch (error) {
    console.error('Fetch error:', error);
    return null;
  }
};

const getVersionLogsList = async () => {
  const data = await fetchData('/api/v3/version_log/version_logs_list/');
  return mapVersionLogs(data || {});
};

const loadDetail = async (version: string) => {
  const log = logList.value.find(item => item.title === version);
  if (log?.detail) {
    currentDetail.value = log.detail;
    return;
  }
  detailLoading.value = true;
  try {
    const data = await fetchData(`/api/v3/version_log/changelog/${version}`);
    const rawMarkdown = data?.data?.content || '';
    if (log) log.detail = rawMarkdown;
    currentDetail.value = rawMarkdown;
  } finally {
    detailLoading.value = false;
  }
};

const handleSelected = (version: Record<string, string>) => {
  loadDetail(version.title);
};

watch(isShow, async (v) => {
  if (v) {
    detailLoading.value = true;
    logList.value = await getVersionLogsList();
    const currentIndex = resolveCurrentIndex(logList.value);
    currentVersion.value = logList.value[currentIndex]?.title || '';
    if (currentVersion.value) {
      await loadDetail(currentVersion.value);
    }
    detailLoading.value = false;
  }
});

onBeforeUnmount(() => {
  isShow.value = false;
});
</script>

<style>
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
