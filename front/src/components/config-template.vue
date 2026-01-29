<template>
  <div class="mb-[15px] relative">
    <Select
      class="w-[240px]"
      v-model="selectedValue"
      multiple-mode="tag"
      collapse-tags
      multiple
      :list="configMenuList"
      display-key="selectId"
      id-key="selectId"
      @change="handleChange"
    >
      <template #trigger>
        <Button theme="primary" text>
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          <span class="text-[14px]">{{ $t('添加配置项') }}</span>
        </Button>
      </template>
      <template #optionRender="{ item }">
        <Tag v-if="item.pId === 'base_config'">{{ $t('基础') }}</Tag>
        <Tag theme="success" v-else-if="item.pId === 'script_config'">{{ $t('脚本') }}</Tag>
        <Tag theme="info" v-else-if="item.pId === 'data_config'">{{ $t('数据') }}</Tag>
        <Tag theme="warning" v-else-if="item.pId === 'file_config'">{{ $t('文件') }}</Tag>
        <span class="ml-[5px]">{{ item.name }}</span>
      </template>
    </Select>
    <div class="flex items-center text-[#979BA5] absolute left-[120px] top-[2px]">
      <i class="nodeman-icon nc-tips"></i>
      <span class="ml-[9px] text-[12px]"
      >如需修改默认配置，需打开开关后修改</span
      >
    </div>
  </div>
  <div v-for="(temp, index) in configTemplates" :key="index">
    <template v-for="(item, ind) in temp.items">
      <div
        v-if="selectedValue.includes(`${temp.id}-${item.id}`)"
        :key="ind"
        class="flex gap-[15px] mb-[20px] h-[36px] items-center ml-[24px] hover:bg-[#F0F1F5] relative tempItem">
        <div class="text-[14px] w-[180px] text-right">{{ item.name_zh }}</div>
        <span>
          <Tag v-if="temp.id === 'base_config'">{{ $t('基础') }}</Tag>
          <Tag theme="success" v-else-if="temp.id === 'script_config'">{{ $t('脚本') }}</Tag>
          <Tag theme="info" v-else-if="temp.id === 'data_config'">{{ $t('数据') }}</Tag>
          <Tag theme="warning" v-else-if="temp.id === 'file_config'">{{ $t('文件') }}</Tag>
        </span>
        <div class="w-[200px]">
          <Input
            v-if="item.type === 0"
            v-model="item.value_string"
            @change="updateValue"
          />
          <Input
            v-if="item.type === 1"
            v-model="item.value_int"
            type="number"
            @change="updateValue"
          />
          <Switcher
            v-else-if="item.type === 2"
            v-model="item.value_bool"
            theme="primary"
            show-text
            @change="updateValue"
          />
          <Select
            v-else-if="item.type === 3"
            v-model="item.value_string"
            :list="getList(index, ind, 'string')"
            @change="updateValue"
          ></Select>
          <Select
            v-else-if="item.type === 4"
            v-model="item.value_int"
            :list="getList(index, ind, 'number')"
            @change="updateValue"
          ></Select>
        </div>
        <Button
          class="absolute right-[20px] hidden deleteBtn w-[50px]"
          text
          @click="handleDelete(`${temp.id}-${item.id}`)">
          <i class="nodeman-icon nc-delete text-[20px]"></i>
        </Button>
      </div>
    </template>
  </div>
</template>
<script lang="ts" setup>
import { Button, Input, Radio, Select, Switcher, Tag } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import { computed, onMounted, ref, watch  } from 'vue';

import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { useMainStore } from '@/stores/main';

const props = defineProps<{
  visible?: boolean;
  isEdit: boolean;
  configpolicyType: string;
  configs: ConfigPolicyConfigBlock[];
}>();

const emit = defineEmits(['updateConfig', 'updateSelect']);
const configTemplates = ref<ConfigPolicyConfigBlock[]>([]);
const typeMap = {
  0: 'string',
  1: 'number',
  2: 'boolean',
  3: 'stringSelect',
  4: 'numberSelect',
};

const mainStore = useMainStore();
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');
const selectedValue = ref<string[]>([]);
const configMenuList = ref<any[]>([]);
// 获取到配置项模板数据时候更新配置项下拉数据和勾选值
const updateConfigTemplates = (configTemplates: any[]) => {
  const list: any[] = [];
  configTemplates.forEach((temp: any) => {
    list.push(...temp.items.map((item: any) => ({
      pId: temp.id,
      id: item.id,
      selectId: `${temp.id}-${item.id}`,
      name: isZh.value ? item.name_zh : item.name_en,
      enabled: item.enabled,
    })));
  });
  configMenuList.value = list;
  selectedValue.value = list.filter((item: any) => item.enabled).map((item: any) => item.selectId);
};
const handleChange = (val: string[]) => {
  // 勾选后更新配置项的enabled状态
  configTemplates.value = configTemplates.value.map((temp: any) => ({
    ...temp,
    items: temp.items.map((item: any) => ({
      ...item,
      enabled: val.includes(`${temp.id}-${item.id}`),
    })),
  }));
  emit('updateConfig', configTemplates.value);
};
const handleDelete = (selectId: string) => {
  selectedValue.value = selectedValue.value.filter((item: string) => item !== selectId);
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

  const selectList = type === 'string'
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
  if (props.isEdit && props.configs) {
    configTemplates.value = cloneDeep(props.configs);
    updateConfigTemplates(configTemplates.value);
    return;
  }
  const res = await ConfigPolicyAPIService.ConfigPolicyTemplate({
    configpolicy_type: props.configpolicyType,
  }).catch(() => ({
    templates: [],
  }));
  configTemplates.value = cloneDeep(res.templates);
  updateConfigTemplates(configTemplates.value);
};
// 监听visible变化
watch(() => props.visible, (newVal: boolean) => {
  if (newVal) {
    getConfigs();
  }
}, { immediate: true });

// 移除onMounted调用，改为通过visible监听触发
// onMounted(async () => {
//   await getConfigs();
// });
</script>
<style lang="postcss" scoped>
.tempItem {
  &:hover {
    .deleteBtn {
      display: block;
    }
  }
}
</style>
