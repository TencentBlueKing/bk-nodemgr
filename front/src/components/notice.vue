<template>
  <div class="bk-notice-component-alert" v-if="alertNoticeList.length && !isClosed">
    <img class="notice-icon" src="../../public/images/alert.svg" width="14" height="14" />
    <bk-popover
      :disabled="disableTooltips"
      :width="500"
      :allow-html="true"
      :popover-delay="[400, 100]"
      theme="light"
      placement="bottom"
    >
      <template #content>
        <p
          class="bk-notice-popover-con"
          v-html="filterContent(renderAlertContent)"
        ></p>
      </template>
      <div class="bk-notice-content-container">
        <p
          class="bk-notice-content"
          :style="{ 'margin-left': alertLeft + 'px' }"
          v-html="filterContent(renderAlertContent)"
          @mouseover="alertPause = true"
          @mouseout="alertPause = false"
        ></p>
      </div>
    </bk-popover>
    <div v-if="alertNoticeList.length > 1" class="alert-pagination">
      <span
        class="operate-icon"
        :class="{ disabled: alertIndex <= 0 }"
        @click="changeAlertIndex('prev')"
      >
        &lt
      </span>
      <span class="alert-num">
        {{ alertIndex + 1 }}/{{ alertNoticeList.length }}
      </span>
      <span
        class="operate-icon"
        :class="{ disabled: alertIndex >= alertNoticeList.length - 1 }"
        @click="changeAlertIndex('next')"
      >
        >
      </span>
    </div>
    <i
      v-if="showCloseIcon"
      class="nodeman-icon nc-delete cursor-pointer text-[22px] text-[#979aa8] noticeCloseIcon"
      @click="handleClose">
    </i>
  </div>
  <bk-dialog
    ext-cls="bk-notice-component-dialog"
    header-align="center"
    width="480"
    show-mask
    :z-index="5000"
    :is-show="showDialog"
    :title="currentDialog.title"
    :close-icon="false"
    :esc-close="false"
    :quick-close="false"
  >
    <div
      class="notice-content"
      v-html="filterContent(currentDialog.content)"
    ></div>
    <template #footer>
      <div class="dialog-footer notice-buttons">
        <Button v-if="dialogIndex > 0" @click="changeDialogIndex('prev')">{{
          getLangText("上一条")
        }}</Button>
        <Button
          v-if="dialogIndex < showDialogNoticeList.length - 1"
          theme="primary"
          @click="changeDialogIndex('next')"
        >{{ getLangText("下一条") }}
        </Button>
        <bk-button
          class="ml-[8px]"
          v-if="dialogIndex >= showDialogNoticeList.length - 1"
          theme="primary"
          @click="confirmDialog"
        >{{ getLangText("确定") }}
        </bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<script setup>
import {
  Button,
  Dialog,
  Popover,
} from 'bkui-vue';
import DOMPurify from 'dompurify';
import {
  computed,
  nextTick,
  onBeforeMount,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from 'vue';

import { parseCookies } from '@/common/util';


defineOptions({
  name: 'NoticeComponent',
});

const props = defineProps({
  apiUrl: {
    type: String,
    default: '',
  },
  loopTime: {
    type: Number,
    default: 60,
  },
  alertSpeed: {
    type: Number,
    default: 30,
  },
  showCloseIcon: {
    type: Boolean,
    default: true,
  },
});

const emit = defineEmits(['show-alert-change']);
const EN_MAP = {
  上一条: 'Previous',
  下一条: 'Next',
  确定: 'Confirm',
};

// 兼容各浏览器的hidden属性
let hidden;
let visibilityChange;
if (typeof document.hidden !== 'undefined') {
  hidden = 'hidden';
  visibilityChange = 'visibilitychange';
} else if (typeof document.msHidden !== 'undefined') {
  hidden = 'msHidden';
  visibilityChange = 'msvisibilitychange';
} else if (typeof document.webkitHidden !== 'undefined') {
  hidden = 'webkitHidden';
  visibilityChange = 'webkitvisibilitychange';
}

const renderList = ref([]);

const alertIndex = ref(0);
const dialogIndex = ref(0);

const alertLeft = ref(0);
const alertPause = ref(false);
const disableTooltips = ref(true);
// 这两个是跑马灯相关的变量
let interval = null;
let timer = null;

// api轮询Interval
let apiInterval = null;
let apiErrTimes = 0;

const showDialog = ref(false);
const showDialogNoticeList = ref([]);

onBeforeMount(async () => {
  if (props.apiUrl) {
    fetchApiData();
    apiInterval = setInterval(fetchApiData, 1000 * props.loopTime);
  }
});

onMounted(() => {
  document.addEventListener(visibilityChange, handleVisibilityChange);
});

onBeforeUnmount(() => {
  clearAllTimer();
  apiInterval && clearInterval(apiInterval);
  document.removeEventListener(visibilityChange, handleVisibilityChange);
});

const alertNoticeList = computed(() => renderList.value.filter(item => item.announce_type === 'announce') || []);

const dialogNoticeList = computed(() => renderList.value.filter(item => item.announce_type === 'event') || []);

const currentAlert = computed(() => alertNoticeList.value[alertIndex.value] || {});

const renderAlertContent = computed(() => currentAlert.value.content
  ?.replaceAll('<p style="line-height: 1.5;"><br></p>', '')
  ?.replaceAll('<br>', '')
  ?.replaceAll('<p', '<span')
  ?.replaceAll('</p>', '</span>'));

const currentDialog = computed(() => showDialogNoticeList.value[dialogIndex.value] || {});

const isClosed = ref(false);
const handleClose = () => {
  isClosed.value = true;
  emit('show-alert-change', false);
};

watch(
  renderList,
  (val, oldVal) => {
    initData();
    // 如果先前只有一条不滚动的公告， 现在要开启轮询
    if (disableTooltips.value && val.length > 1 && oldVal.length === 1) {
      timer && clearTimeout(timer);
      timer = setTimeout(changeAlertIndex, 6000);
    }
    emit('show-alert-change', alertNoticeList.value.length > 0 && !isClosed.value);
  },
  {
    deep: true,
  },
);

// 当前的跑马灯公告改变
watch(
  () => currentAlert.value,
  (val, oldVal) => {
    if (val.id !== oldVal.id) {
      resetAlert();
    }
  },
);

const handleVisibilityChange = function () {
  if (document[hidden]) {
    apiInterval && clearInterval(apiInterval);
    console.log('页面不可见');
  } else {
    console.log('页面可见');
    if (props.apiUrl) {
      fetchApiData();
      apiInterval = setInterval(fetchApiData, 1000 * props.loopTime);
      apiErrTimes = 0;
    }
  }
};

// 处理滚动的跑马灯
const handleAlertMarquee = function () {
  alertLeft.value = 0;
  // 获取容器的宽度跟实际内容的宽度
  const container = document.querySelector('.bk-notice-content-container');
  const content = document.querySelector('.bk-notice-content');

  const scrollLeft = () => {
    interval = setInterval(() => {
      if (alertPause.value !== true) {
        alertLeft.value -= 1;
      }
      if (alertLeft.value < container.offsetWidth - content.offsetWidth) {
        clearAllTimer();
        if (alertNoticeList.value.length > 1) {
          timer = setTimeout(changeAlertIndex, 3000);
        } else {
          timer = setTimeout(handleAlertMarquee, 3000);
        }
      }
    }, props.alertSpeed);
  };
  // 如果内容宽度超出容器宽度， 开始滚动跑马灯
  if (content.offsetWidth > container.offsetWidth) {
    disableTooltips.value = false;
    timer && clearTimeout(timer);
    timer = setTimeout(scrollLeft, 3000);
  } else {
    disableTooltips.value = true;
    if (alertNoticeList.value.length > 1) {
      timer && clearTimeout(timer);
      timer = setTimeout(changeAlertIndex, 6000);
    }
  }
};

const resetAlert = function () {
  clearAllTimer();
  alertLeft.value = 0;
  if (currentAlert.value?.id) {
    nextTick(handleAlertMarquee);
  }
};

const clearAllTimer = function () {
  interval && clearInterval(interval);
  timer && clearTimeout(timer);
};

const fetchApiData = async function () {
  const lang = parseCookies().blueking_language || 'zh-cn';
  const res = await fetch(props.apiUrl, {
    mode: 'cors',
    credentials: 'include',
  });
  if (res.status === 200) {
    const data = await res.json();
    renderList.value = data?.data.map(item => {
      const findItem = item.content_list.find(content => content.language === lang);
      return {
        ...item,
        content: findItem ? findItem.content : '',
      };
    }) || [];
  } else {
    console.error(`获取消息通知公告失败：${res.statusText}(${res.status})`);
    // eslint-disable-next-line no-plusplus
    apiErrTimes++;
    if (apiErrTimes >= 5) {
      apiInterval && clearInterval(apiInterval);
    }
  }
};

const initData = function () {
  const cookieDialogIds = (localStorage.getItem('bk-dialog-notice-ids') || '').split(',') || [];
  // eslint-disable-next-line max-len
  showDialogNoticeList.value = dialogNoticeList.value.filter(item => cookieDialogIds.indexOf(item.id?.toString()) === -1);
  showDialog.value = showDialogNoticeList.value.length > 0;
  if (
    alertIndex.value > 0
    && alertIndex.value >= alertNoticeList.value.length
  ) {
    alertIndex.value = 0;
    resetAlert();
  }
};

// type 类型有 prev、next、loop
const changeAlertIndex = function (type = 'loop') {
  if (type === 'prev') {
    if (alertIndex.value > 0) {
      // eslint-disable-next-line no-plusplus
      alertIndex.value--;
    }
  } else if (type === 'next') {
    if (alertIndex.value < alertNoticeList.value.length - 1) {
      // eslint-disable-next-line no-plusplus
      alertIndex.value++;
    }
  } else {
    // 循环
    if (alertIndex.value < alertNoticeList.value.length - 1) {
      // eslint-disable-next-line no-plusplus
      alertIndex.value++;
    } else {
      alertIndex.value = 0;
    }
  }
};

const changeDialogIndex = function (type = 'next') {
  if (type === 'prev') {
    if (dialogIndex.value > 0) {
      // eslint-disable-next-line no-plusplus
      dialogIndex.value--;
    }
  } else {
    if (dialogIndex.value < showDialogNoticeList.value.length - 1) {
      // eslint-disable-next-line no-plusplus
      dialogIndex.value++;
    }
  }
};

const confirmDialog = function () {
  const ids = dialogNoticeList.value.map(item => item.id?.toString());
  let existIds = localStorage.getItem('bk-dialog-notice-ids')?.split(',') || [];
  existIds = existIds.concat(ids);
  const newIds = Array.from(new Set(existIds));
  const newStr = newIds.join(',');
  localStorage.setItem('bk-dialog-notice-ids', newStr);
  showDialog.value = false;
  setTimeout(() => {
    dialogIndex.value = 0;
  }, 500);
};

const getLangText = function (val) {
  const lang = parseCookies().blueking_language;
  if (lang === 'en') {
    return EN_MAP[val] || val;
  }
  return val;
};

const filterContent = function (con) {
  // 兼容打开企业微信链接
  const allowUriConfig = /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|wxwork):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i;
  return DOMPurify.sanitize(con, {
    ALLOWED_URI_REGEXP: allowUriConfig,
  });
};
</script>

<style lang="scss">
.noticeCloseIcon {
  font-weight: 600 !important;
}
.bk-notice-popover-con {
  max-height: calc(100vh - 120px);
  word-break: break-all;
  overflow-y: auto;
  * {
    display: inline-block;
    margin-block-start: 0;
    margin-block-end: 0;
    margin-inline-start: 0;
    margin-inline-end: 0;
  }
  span {
    display: inline;
  }
}
.bk-notice-component-alert {
  padding: 0 40px;
  height: 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #faf0d2;
  border-radius: 2px;
  font-size: 12px;
  color: #63656e;

  .notice-icon {
    color: #e6aa00;
    margin-right: 8px;
  }
  .bk-notice-content-container {
    height: 40px;
    /* line-height: 40px; */
    display: flex;
    align-items: center;
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    * {
      margin-block-start: 0;
      margin-block-end: 0;
      margin-inline-start: 0;
      margin-inline-end: 0;
    }
    .bk-notice-content {
      display: inline-block;
      transition: all 0.2s linear;
      * {
        display: inline-block;
      }
    }
  }
  .operate-icon {
    cursor: pointer;
  }
  .alert-pagination {
    width: 100px;
    text-align: right;
    margin-left: 8px;
    span {
      padding: 4px;
      font-size: 12px;
    }
    .alert-num {
      cursor: default;
      color: #979ba5;
    }
    .operate-icon {
      cursor: pointer;
      font-weight: 800;
      color: #3a84ff;
      &.disabled {
        cursor: not-allowed;
        color: #a3c5fd;
      }
    }
  }
}
.bk-notice-component-dialog.bk-modal-wrapper.bk-dialog-wrapper {
  .bk-modal-header {
    display: none;
  }
  .bk-modal-close {
    display: none;
  }
  .bk-modal-body {
    padding-bottom: 80px;
    .bk-modal-content {
      min-height: 220px;
      max-height: 420px;
      padding: 32px 24px 0;
      overflow-x: hidden;
      .notice-content {
        word-wrap: break-word;
      }
      p {
        margin: 0;
      }
    }
  }
  .bk-modal-footer {
    background-color: #fff;
    border-top: none;
    padding: 0;
    height: 32px;
    line-height: 32px;
    margin: 24px;
    .notice-buttons {
      .bk-button {
        width: 88px;
        margin-left: 8px;
      }
    }
  }
}
</style>
