<template>
  <span
    v-if="!hasPermission"
    v-bk-tooltips="{
      content: tooltipContent,
      placement: 'top',
    }"
    class="auth-button-wrapper auth-button-disabled"
  >
    <Button v-bind="$attrs" :disabled="true">
      <slot />
    </Button>
    <i class="auth-lock-icon" />
  </span>
  <Button v-else v-bind="$attrs">
    <slot />
  </Button>
</template>

<script setup lang="ts">
import { Button } from 'bkui-vue';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  hasPermission: boolean;
  actionName?: string;
}>();

const { t } = useI18n();

const tooltipContent = computed(() => {
  if (props.actionName) {
    return t('components.permission.missingAction', { action: props.actionName });
  }
  return t('components.permission.noPermission');
});
</script>

<style lang="postcss" scoped>
.auth-button-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
  cursor: not-allowed;
}

.auth-button-disabled :deep(.bk-button) {
  pointer-events: none;
}

.auth-lock-icon {
  position: absolute;
  top: 50%;
  right: 4px;
  width: 14px;
  height: 14px;
  transform: translateY(-50%);
  mask-image: url('/images/lock.svg');
  mask-size: contain;
  mask-repeat: no-repeat;
  mask-position: center;
  -webkit-mask-image: url('/images/lock.svg');
  -webkit-mask-size: contain;
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-position: center;
  background-color: #979ba5;
}
</style>
