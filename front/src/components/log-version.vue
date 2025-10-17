<template>
  <Dialog
    width="1105"
    v-model:value="isShow"
    :show-footer="false"
    ext-cls="log-version-dialog"
    @value-change="handleValueChange"
  >
    <div class="log-version" v-bkloading="{ isLoading: loading }">
      <div class="log-version-left">
        <ul class="left-list">
          <li
            v-for="(item, index) in logList"
            :class="['left-list-item', { 'item-active': index === active }]"
            :key="index"
            @click="handleItemClick(index)"
          >
            <span class="item-title">{{ item.title }}</span>
            <span class="item-date">{{ item.date }}</span>
            <span v-if="index === current" class="item-current">
              {{ $t("当前版本") }}
            </span>
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

interface ILog {
  title: string;
  date: string;
  detail: string;
}

// 使用defineModel定义双向绑定的isShow属性
const isShow = defineModel('isShow', { type: Boolean });

// 定义组件Emits（仅保留change事件）
const emit = defineEmits<{
  (e: 'change', value: boolean): void;
}>();

// 响应式变量
const loading = ref(false);
const current = ref(0);
const active = ref(0);
const logList = ref<ILog[]>([]);

// 计算属性 - 当前选中的日志
const currentLog = computed(() => logList.value[active.value] || { title: '', date: '', detail: '' });

// dialog显示状态变更
const handleValueChange = (value: boolean) => {
  emit('change', value);
  // 同步更新isShow的值
  isShow.value = value;
};

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
    const baseURL = window.PROJECT_CONFIG.SITE_URL;
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
  const data = await fetchData('version_log/version_logs_list/');
  if (!data) return [];

  return data.map((item: string[]) => ({
    title: item[0],
    date: item[1],
    detail: '',
  }));
};

// 获取版本日志详情
const getVersionLogsDetail = async () => {
  const data = await fetchData('version_log/version_log_detail/', {
    log_version: currentLog.value.title,
  });

  return data || '';
};
// 监听isShow变化
watch(isShow, async (v) => {
  if (v) {
    loading.value = true;
    logList.value = await getVersionLogsList();
    if (logList.value.length) {
      await handleItemClick(0);
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
.log-version {
  display: flex;
  margin: -33px -24px -26px;

  &-left {
    flex: 0 0 260px;
    background-color: #fafbfd;
    border-right: 1px solid #dcdee5;
    padding: 40px 0;
    display: flex;
    font-size: 12px;

    .left-list {
      border-top: 1px solid #dcdee5;
      border-bottom: 1px solid #dcdee5;
      height: 520px;
      overflow: auto;
      display: flex;
      flex-direction: column;
      width: 100%;

      &-item {
        flex: 0 0 54px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        padding-left: 30px;
        position: relative;
        border-bottom: 1px solid #dcdee5;

        &:hover {
          cursor: pointer;
          background-color: #fff;
        }

        .item-title {
          color: #313238;
          font-size: 16px;
        }

        .item-date {
          color: #979ba5;
        }

        .item-current {
          position: absolute;
          right: 20px;
          top: 8px;
          background-color: #699df4;
          border-radius: 2px;
          width: 58px;
          height: 20px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: #fff;
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

  &-right {
    flex: 1;
    padding: 25px 30px 50px 45px;

    .detail-container {
      max-height: 525px;
      overflow: auto;
    }
  }
}
</style>
