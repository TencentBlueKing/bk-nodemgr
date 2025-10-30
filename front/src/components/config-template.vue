<template>
  <div v-for="(temp, index) in configTemplates" :key="index">
    <div class="flex mb-[22px] text-[#4D4F56]">
      <div class="min-w-[120px] mr-[10px] text-right text-[14px]">{{ temp.title_zh }}</div>
      <div class="flex-1 flex flex-col gap-[12px]">
        <div v-for="(item, ind) in temp.items" :key="ind" class="flex gap-[12px] items-center ml-[24px]">
          <Switcher v-model="item.enabled" @change="updateValue" theme="primary"></Switcher>
          <div class="text-[14px]">{{ item.name_zh }}</div>
          <div class="flex-1">
            <Input
              v-if="item.type === 0"
              v-model="item.value_string"
              :disabled="!item.enabled"
              @change="updateValue"
            />
            <Input
              v-if="item.type === 1"
              v-model="item.value_int"
              :disabled="!item.enabled"
              type="number"
              @change="updateValue"
            />
            <Radio.Group
              v-model="item.value_bool"
              :disabled="!item.enabled"
              v-else-if="item.type === 2"
              @change="updateValue"
            >
              <Radio.Button :label="true">{{ $t('components.configTemplate.enable') }}</Radio.Button>
              <Radio.Button :label="false">{{ $t('components.configTemplate.disabled') }}</Radio.Button>
            </Radio.Group>
            <Select
              v-else-if="item.type === 3"
              v-model="item.value_string"
              :list="getList(index, ind, 'string')"
              :disabled="!item.enabled"
              @change="updateValue"
            ></Select>
            <Select
              v-else-if="item.type === 4"
              v-model="item.value_int"
              :list="getList(index, ind, 'number')"
              :disabled="!item.enabled"
              @change="updateValue"
            ></Select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { Input, Radio, Select, Switcher } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import { computed, onMounted, ref, watch  } from 'vue';
import { useRoute } from 'vue-router';

import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { useMainStore } from '@/stores/main';

const emit = defineEmits(['updateConfig']);
const route = useRoute();
const mainStore = useMainStore();
const configpolicyType = computed(() => route.params.configpolicy_type);
const isEdit = computed(() => route.name === 'editConfig');
const configTemplates = ref<ConfigPolicyConfigBlock[]>([]);
const typeMap = {
  0: 'string',
  1: 'number',
  2: 'boolean',
  3: 'stringSelect',
  4: 'numberSelect',
};
const updateValue = () => {
  emit('updateConfig', configTemplates.value);
};

const getList = (
  index: number,
  ind: number,
  type: 'string' | 'number',
): { value: string | number; label: string }[] => {
  const templateItem = configTemplates.value[index]?.items[ind];

  if (!templateItem) {
    return [];
  }

  const selectList =    type === 'string'
    ? templateItem.value_string_select
    : templateItem.value_int_select;

  const mappedList = selectList?.map(item => ({
    value: item,
    label: String(item),
  })) || []; // 如果未定义，使用空数组以避免错误

  return mappedList;
};

// 获取配置
const getConfigs = async () => {
  if (isEdit.value && mainStore.configEditData) {
    configTemplates.value = cloneDeep(mainStore.configEditData.configs);
    return;
  }
  const res = await ConfigPolicyAPIService.ConfigPolicyTemplate({
    configpolicy_type: configpolicyType.value,
  }).catch(() => ({
    templates: [],
  }));
  configTemplates.value = cloneDeep(res.templates);
};
onMounted(async () => {
  await getConfigs();
});
</script>
