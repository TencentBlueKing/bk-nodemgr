<template>
  <div class="flex items-center place-content-between">
    <div :class="average ? 'flex-1' : ''"><slot name="left"></slot></div>
    <div :class="average ? 'flex-1' : ''"><slot name="right"></slot></div>
  </div>
</template>
<script lang="ts" setup>
interface IProps {
  average?: boolean
}

defineProps<IProps>();
</script>
