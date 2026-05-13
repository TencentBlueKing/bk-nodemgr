<template>
  <div
    class="validate-cell"
    :class="{ 'validate-cell--error': shouldShowError }"
    @focusin="handleFocus"
    @focusout="handleBlur"
  >
    <!-- 插槽：放置 Input/Select -->
    <slot></slot>

    <!--
      错误提示图标
      使用 v-show 而非 v-if，避免 DOM 增减导致 VxeTable 重新计算列宽
    -->
    <div
      v-show="shouldShowError"
      class="error-indicator"
      ref="triggerRef"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
    >
      <i class="nodeman-icon nc-remind-fill text-[16px]"></i>
    </div>

    <!-- 全局传送的气泡 (解决 z-index 遮挡问题) -->
    <Teleport to="body">
      <div
        v-if="visible && shouldShowError"
        class="validate-cell-global-tooltip"
        :style="tooltipStyle"
      >
        {{ error }}
        <div class="tooltip-arrow"></div>
      </div>
    </Teleport>
  </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';

const props = defineProps({
  error: {
    type: String,
    default: '',
  },
});

const triggerRef = ref<HTMLElement | null>(null);
const visible = ref(false);
const isFocused = ref(false); // 新增：是否聚焦状态

// 计算属性：是否显示错误
// 只有在【有错误信息】且【当前未聚焦】时才显示红色样式
const shouldShowError = computed(() => !!props.error && !isFocused.value);

const tooltipStyle = reactive({
  top: '0px',
  left: '0px',
});

// 聚焦隐藏错误
const handleFocus = () => {
  isFocused.value = true;
  visible.value = false; // 聚焦时顺便把气泡关掉
};

// 失焦恢复检测
const handleBlur = () => {
  isFocused.value = false;
};

const handleMouseEnter = () => {
  if (!triggerRef.value) return;

  const rect = triggerRef.value.getBoundingClientRect();

  // --- 位置计算 (改为上方) ---
  // top: 使用图标的 top 作为基准
  // 配合 CSS 的 translateY(-100%)，气泡会显示在图标上方
  tooltipStyle.top = `${rect.top - 8}px`; // 减 8px 留点空隙

  // left: 对齐处理
  tooltipStyle.left = `${rect.left - 20}px`;

  visible.value = true;
};

const handleMouseLeave = () => {
  visible.value = false;
};
</script>

<style lang="postcss" scoped>
.validate-cell {
  position: relative;
  width: 100%;
}
.validate-cell :deep(.bk-select),
.validate-cell :deep(.bk-input),
.validate-cell :deep(.bk-upload) {
  width: 100%;
}
.validate-cell--error {
  &::v-deep(.bk-input--text),
  &::v-deep(.bk-textarea),
  &::v-deep(.bk-select-tag),
  &::v-deep(.bk-input) {
    border-color: #ea3636 !important;
    color: #ea3636 !important;
    background-color: #fff0f0 !important;
    .bk-input--suffix-icon { background-color: #fff0f0 !important; }
    /* 聚焦时底部装饰线也置为红色 */
    &::after {
      background-color: #ea3636 !important;
    }
  }
}
.error-indicator {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 10;
  display: flex;
  align-items: center;
  cursor: pointer;
  line-height: 1;
  color: #ea3636;
}
</style>

<style lang="postcss">
.validate-cell-global-tooltip {
  position: fixed;
  background: #333;
  color: #fff;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  pointer-events: none;
  z-index: 99999;
  transform: translateY(-100%); /* 向上生长 */
}

.validate-cell-global-tooltip .tooltip-arrow {
  position: absolute;
  top: 100%;
  left: 23px; /* 固定在左侧，不再随文字长度变化 */
  border-left: 5px solid transparent;
  border-right: 5px solid transparent;
  border-top: 5px solid #333;
}
</style>
