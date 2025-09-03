<template>
  <div class="flex flex-wrap">
    <SelectItem
      class="mr-[8px]"
      v-for="(item, index) in list"
      :key="index"
      :value="item.value"
      :icon="item.icon"
      :title="item.title"
      :content="item.content"
      :checked="data.includes(item.value)"
      @change="handleChange"
    ></SelectItem>
  </div>
</template>

<script lang="ts" setup>
import { type PropType, ref } from 'vue';

import SelectItem from './select-item.vue';

interface SelectItemList {
  value: number | string
  icon: string
  title: string
  content: string
}

const props = defineProps({
  list: {
    type: Array as PropType<SelectItemList[]>,
    default: [],
  },
  multiple: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(['change']);
const data = ref<Array<number | string>>([props.list[0].value]);

const handleChange = (state: boolean, value: string | number) => {
  if (!props.multiple) {
    data.value = [];
    if (state) {
      data.value.push(value);
    }
  } else {
    if (state) {
      data.value.push(value);
    } else {
      const index = data.value.findIndex(item => item === value);
      data.value.splice(index, 1);
    }
  }
  emit('change', data.value);
};
</script>
