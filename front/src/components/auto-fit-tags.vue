<template>
  <div ref="containerRef" class="auto-fit-tags relative w-full overflow-hidden">
    <!-- Hidden measurement layer: renders all tags invisibly to measure their widths -->
    <div
      ref="measureRef"
      class="flex items-center gap-[4px] absolute left-0 top-0 w-full"
      style="visibility: hidden; pointer-events: none; height: 0; overflow: hidden;"
    >
      <span
        v-for="(tag, i) in tags"
        :key="'measure-' + i"
        ref="tagMeasureRefs"
        class="inline-flex shrink-0"
      >
        <Tag>{{ getLabel(tag) }}</Tag>
      </span>
      <span ref="overflowMeasureRef" class="inline-flex shrink-0">
        <Tag>+99</Tag>
      </span>
    </div>
    <!-- Visible display layer -->
    <div class="flex items-center gap-[4px]" :style="{ visibility: measured ? 'visible' : 'hidden' }">
      <Tag v-for="(tag, i) in visibleTags" :key="i">{{ getLabel(tag) }}</Tag>
      <Tag
        v-if="hiddenCount > 0"
        v-bk-tooltips="{ content: tooltipContent, maxWidth: 400 }"
      >+{{ hiddenCount }}</Tag>
    </div>
    <!-- Placeholder to prevent layout shift before measurement -->
    <div v-if="!measured && tags.length > 0" class="h-[22px]"></div>
  </div>
</template>

<script lang="ts" setup>
import { Tag } from 'bkui-vue';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

const props = defineProps({
  /** Array of tag values to display */
  tags: {
    type: Array as () => (string | number)[],
    default: () => [],
  },
  /** Optional label mapping: { tagValue: displayLabel } */
  labelMap: {
    type: Object as () => Record<string, string> | null,
    default: null,
  },
});

const containerRef = ref<HTMLElement>();
const tagMeasureRefs = ref<HTMLElement[]>([]);
const overflowMeasureRef = ref<HTMLElement>();
const visibleCount = ref(0);
const measured = ref(false);

const GAP = 4; // px, matches gap-[4px]

const getLabel = (tag: string | number): string => {
  if (props.labelMap) {
    return props.labelMap[String(tag)] ?? String(tag);
  }
  return String(tag);
};

const visibleTags = computed(() => props.tags.slice(0, visibleCount.value));
const hiddenCount = computed(() => Math.max(0, props.tags.length - visibleCount.value));
const tooltipContent = computed(() =>
  props.tags
    .slice(visibleCount.value)
    .map(t => getLabel(t))
    .join(', '),
);

const calculateVisible = () => {
  const container = containerRef.value;
  if (!container || props.tags.length === 0) {
    visibleCount.value = props.tags.length;
    measured.value = true;
    return;
  }

  const containerWidth = container.clientWidth;
  if (containerWidth <= 0) {
    visibleCount.value = props.tags.length;
    measured.value = true;
    return;
  }

  const measureEls = tagMeasureRefs.value;
  const overflowEl = overflowMeasureRef.value;
  const overflowWidth = overflowEl ? overflowEl.offsetWidth : 36;

  if (!measureEls || measureEls.length === 0) {
    visibleCount.value = props.tags.length;
    measured.value = true;
    return;
  }

  // If all tags fit, show all
  let totalWidth = 0;
  for (let i = 0; i < measureEls.length; i++) {
    totalWidth += measureEls[i].offsetWidth + (i > 0 ? GAP : 0);
  }
  if (totalWidth <= containerWidth) {
    visibleCount.value = props.tags.length;
    measured.value = true;
    return;
  }

  // Otherwise, find how many fit with the overflow tag
  let usedWidth = 0;
  let count = 0;
  for (let i = 0; i < measureEls.length; i++) {
    const tagWidth = measureEls[i].offsetWidth;
    const widthWithThis = usedWidth + tagWidth + (count > 0 ? GAP : 0);
    // Reserve space for overflow tag
    const reserveForOverflow = overflowWidth + GAP;

    if (widthWithThis + reserveForOverflow <= containerWidth) {
      usedWidth = widthWithThis;
      count++;
    } else {
      break;
    }
  }

  visibleCount.value = Math.max(1, count); // Show at least 1 tag
  measured.value = true;
};

let observer: ResizeObserver | null = null;

const setupObserver = () => {
  if (observer) observer.disconnect();
  if (!containerRef.value) return;

  observer = new ResizeObserver(() => {
    measured.value = false;
    nextTick(calculateVisible);
  });
  observer.observe(containerRef.value);
};

onMounted(() => {
  nextTick(() => {
    calculateVisible();
    setupObserver();
  });
});

watch(
  () => [...props.tags],
  () => {
    measured.value = false;
    nextTick(calculateVisible);
  },
);

onBeforeUnmount(() => {
  observer?.disconnect();
  observer = null;
});
</script>
