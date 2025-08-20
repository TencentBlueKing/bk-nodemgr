<template>
  <div class="w-full h-full flex  items-center gap-[8px]" v-if="!isEditing" @click="handleEdit">
    <Tag
      v-for="tag in data.labels?.slice(0, 2)"
      :key="tag"
      >{{ tag }}
    </Tag
    >
    <Tag
      v-if="data.labels?.length > 2"
      v-bk-tooltips="data.labels?.join(', ')"
      >+{{ data.labels?.length - 2 }}
    </Tag>
  </div>
  <div class="w-full edit-tag" v-else>
    <TagInput
      class="w-full max-w-[300px]"
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
        <span>新建 “</span>
        <span class="text-[#3A84FF]">{{ createTag }}</span>
        <span>” 标签</span>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed, ref, onMounted } from "vue";
import { Tag, TagInput } from "bkui-vue";
import { PackageService } from "@/api/modules/pkg";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();
const props = defineProps({
  data: {
    type: Object,
    default: null,
  },
});
const emit = defineEmits(['blur']);
const tagList = computed(() =>
  packageStore.tagList.map((tag: string) => ({
    id: tag,
    name: tag,
  }))
);
const localTags = ref<string[]>([...props.data.labels])
const popShow = ref(false);
const createTag = ref("");
const isEditing = ref(false);
const handleEdit = () => {
  // 关闭其他弹框
  document.body.click();
  isEditing.value = true;
};

const handleInputchange = (value: string) => {
  const inputVal = value.trim();
  if (value.length > 32) {
    return;
  }
  const index = tagList.value.findIndex((item: any) => item === value);
  if (index === -1 && inputVal) {
    popShow.value = true;
    createTag.value = inputVal;
  } else {
    popShow.value = false;
    createTag.value = "";
  }
};
// 新建标签
const handleCreateTag = async () => {
  popShow.value = false;
  localTags.value.push(createTag.value);
};
const handleBlur = async () => {
  // 标签输入框更新标签
  await PackageService.SetReleaseLabels({
    generation: 2,
    release_type: props.data.release_type,
    platform: {
      os_type: props.data.os_type,
      cpu_arch: props.data.cpu_arch
    },
    version: props.data.version,
    labels: [...localTags.value],
  });
  // 用于更新标签信息
  packageStore.getPackages();
  emit('blur')
  isEditing.value = false;
}
</script>
