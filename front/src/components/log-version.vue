<template>
  <Dialog
    width="1105"
    :is-show="isShow"
    ext-cls="log-version-dialog"
    @closed="isShow = false">
    <div class="log-version" v-bkloading="{ isLoading: loading }">
      <div class="log-version-left">
        <ul class="left-list">
          <li
            v-for="(item, index) in logList"
            :class="['left-list-item', { 'item-active': index === active }]"
            :key="index"
            @click="handleItemClick(index)"
          >
            <div class="item-head">
              <span class="item-title">{{ item.title }}</span>
              <span v-if="index === current" class="item-current">
                {{ $t("components.logVersion.current") }}
              </span>
            </div>
            <span class="item-date">{{ item.date }}</span>
          </li>
        </ul>
      </div>
      <div class="log-version-right">
        <!-- eslint-disable-next-line vue/no-v-html -->
        <div class="detail-container" v-html="currentLog.detail"></div>
      </div>
    </div>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { Dialog } from 'bkui-vue';
import { marked } from 'marked';
import xss from 'xss';

import { mapVersionLogs, resolveCurrentIndex } from './log-version-helper';

interface ILog {
  title: string;
  date: string;
  detail: string;
  isCurrent: boolean;
}

// 使用defineModel定义双向绑定的isShow属性
const isShow = defineModel('isShow', { type: Boolean });

// 响应式变量
const loading = ref(false);
const current = ref(0);
const active = ref(0);
const logList = ref<ILog[]>([]);

// 计算属性 - 当前选中的日志
const currentLog = computed(() => logList.value[active.value] || { title: '', date: '', detail: '' });

// 点击左侧日志项
const handleItemClick = async (v = 0) => {
  active.value = v;
  if (!currentLog.value.detail) {
    loading.value = true;
    try {
      const detail = await getVersionLogsDetail();
      logList.value[v].detail = detail;
    } finally {
      loading.value = false;
    }
  }
};

// 工具函数：处理fetch请求
const fetchData = async (url: string, params?: Record<string, string>) => {
  try {
    const fullUrl = new URL(url, location.origin);

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        fullUrl.searchParams.append(key, value);
      });
    }

    const response = await fetch(fullUrl.toString(), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Fetch error:', error);
    return null;
  }
};

// 获取版本日志列表
const getVersionLogsList = async () => {
  const data = await fetchData('/api/v3/version_log/version_logs_list/');
  return mapVersionLogs(data || {});
};

// 获取版本日志详情
const getVersionLogsDetail = async () => {
  const data = await fetchData('/api/v3/version_log/changelog/' + currentLog.value.title);
  const rawMarkdown = data?.data?.content || '';
  
  // Markdown → HTML + XSS filtering
  return xss(marked(rawMarkdown, { async: false }), { stripIgnoreTagBody: true });
};

watch(isShow, async (v) => {
  if (v) {
    loading.value = true;
    logList.value = await getVersionLogsList();
    current.value = resolveCurrentIndex(logList.value);
    active.value = current.value;
    if (logList.value.length) {
      await handleItemClick(active.value);
    }
    loading.value = false;
  }
});

// 组件销毁前处理
onBeforeUnmount(() => {
  isShow.value = false;
});
</script>

<style lang="postcss" scoped>
:deep(.log-version-dialog) {
  .bk-modal-content {
    border-radius: 8px;
    box-shadow: 0 10px 28px rgba(25, 25, 41, 0.12);
  }

  .bk-modal-footer {
    display: none;
  }

  .bk-dialog-header {
    padding: 20px 24px 14px;
  }

  .bk-dialog-content {
    padding: 0;
  }
}

.log-version {
  display: flex;
  margin: -14px -24px -24px;
  min-height: 560px;

  .log-version-left {
    flex: 0 0 260px;
    background-color: #fafbfd;
    border-right: 1px solid #dcdee5;
    padding: 18px 0;
    display: flex;
    font-size: 12px;

    .left-list {
      border-top: 1px solid #eaebf0;
      border-bottom: 1px solid #eaebf0;
      height: 524px;
      overflow: auto;
      display: flex;
      flex-direction: column;
      width: 100%;

      .left-list-item {
        flex: 0 0 62px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 4px;
        padding: 8px 20px 8px 24px;
        position: relative;
        border-bottom: 1px solid #eaebf0;
        transition: background-color .2s ease;

        &:hover {
          cursor: pointer;
          background-color: #f0f5ff;
        }

        .item-head {
          display: flex;
          align-items: center;
          justify-content: space-between;
          gap: 8px;
        }

        .item-title {
          color: #313238;
          font-size: 15px;
          line-height: 22px;
          font-weight: 500;
        }

        .item-date {
          color: #979ba5;
          line-height: 18px;
        }

        .item-current {
          background-color: #e1ecff;
          border-radius: 10px;
          padding: 0 8px;
          height: 20px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: #3a84ff;
          white-space: nowrap;
          font-size: 12px;
        }

        &.item-active {
          background-color: #fff;

          &::before {
            content: " ";
            position: absolute;
            top: 0px;
            bottom: 0px;
            left: 0;
            width: 6px;
            background-color: #3a84ff;
          }
        }
      }
    }
  }

  .log-version-right {
    flex: 1;
    padding: 24px 32px 30px;

    .detail-container {
      max-height: 524px;
      overflow: auto;
      color: #4d4f56;
      font-size: 14px;
      line-height: 1.75;

      :deep(h1),
      :deep(h2),
      :deep(h3) {
        color: #313238;
        line-height: 1.4;
        margin: 20px 0 10px;
      }

      :deep(h1) {
        font-size: 22px;
      }

      :deep(h2) {
        font-size: 18px;
      }

      :deep(h3) {
        font-size: 16px;
      }

      :deep(p),
      :deep(ul),
      :deep(ol),
      :deep(blockquote) {
        margin: 0 0 12px;
      }

      :deep(ul),
      :deep(ol) {
        padding-left: 20px;
      }

      :deep(li) {
        margin: 4px 0;
      }

      :deep(code) {
        font-size: 13px;
        background: #f5f7fa;
        border-radius: 4px;
        padding: 1px 5px;
      }

      :deep(pre) {
        background: #f5f7fa;
        border: 1px solid #eaebf0;
        border-radius: 6px;
        padding: 12px;
        overflow: auto;
      }

      :deep(pre code) {
        background: transparent;
        padding: 0;
      }
    }
  }
}
</style>
