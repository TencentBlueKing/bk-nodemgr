<template>
  <div
    class="w-full h-full min-w-[100px] min-h-[22px] cursor-pointer"
    v-if="!isEditing" @click="handleEdit">
    <AutoFitTags :tags="localTags ?? []" />
  </div>
  <div class="w-full edit-tag" v-else>
    <TagInput
      class="w-full max-w-[300px]"
      ref="tagInputRef"
      v-model="localTags"
      :list="tagList"
      allow-create
      has-delete-icon
      collapse-tags
      trigger="focus"
      @input="handleInputchange"
      @blur="handleBlur"
    ></TagInput>
    <div
      class="absolute cursor-pointer w-full mt-[4px] h-[40px] bg-[#fff] border border-[DCDEE5] rounded-[2px] z-999"
      v-if="popShow"
      @click="handleCreateTag"
    >
      <div
        class="w-full h-[32px] leading-[32px] mt-[4px] bg-[#F5F7FA] px-[14px] text-[#63656E] text-[12px]"
      >
        <span>{{ $t('action.create') }} “</span>
        <span class="text-[#3A84FF]">{{ createTag }}</span>
        <span>” {{ $t('action.tag') }}</span>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { TagInput } from 'bkui-vue';
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import AutoFitTags from '@/components/auto-fit-tags.vue';

import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';
import { usePackageStore } from '@/stores/package';

const props = defineProps({
  data: {
    type: Object,
    default: null,
  },
});
const emit = defineEmits(['blur']);
const packageStore = usePackageStore();
const tagList = computed(() => packageStore.tagList.map((tag: string) => ({
  id: tag,
  name: tag,
})));
const localTags = ref<string[]>([]);

const route = useRoute();
const currentType = computed(() => String(route.name).split('PackageMng')[0]);

const popShow = ref(false);
const createTag = ref('');
const isEditing = ref(false);
const tagInputRef = ref();
const handleEdit = () => {
  // 关闭其他弹框
  document.body.click();
  isEditing.value = true;

  nextTick(() => {
    tagInputRef.value?.focusInputTrigger();
  });
};

const handleInputchange = (value: string) => {
  const inputVal = value.trim();
  if (value.length > 32) {
    return;
  }
  const index = tagList.value.findIndex((item: any) => item.id === value);
  if (index === -1 && inputVal) {
    popShow.value = true;
    createTag.value = inputVal;
  } else {
    popShow.value = false;
    createTag.value = '';
  }
};
// 新建标签
const handleCreateTag = async () => {
  popShow.value = false;
  localTags.value.push(createTag.value);
};
const handleBlur = async () => {
  const serviceMethod = currentType.value === 'agent' ? PackageService.SetReleaseAgentLabelsMany : PackageService.SetReleaseProxyLabelsMany;

  // 标签输入框更新标签
  await serviceMethod({
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      platform: [props.data.platform],
      version: [props.data.version],
    },
    labels: [...localTags.value],
  });
  // 用于更新标签信息
  packageStore.getPackages();
  emit('blur');
  isEditing.value = false;
};
watch(() => props.data, () => {
  localTags.value = [...props.data.labels];
}, { immediate: true, deep: true });
</script>
