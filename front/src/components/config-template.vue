<template>
  <div v-for="(temp, index) in configTemplates" :key="index">
    <div class="flex mb-[22px] text-[#4D4F56]">
      <div class="min-w-[120px] mr-[10px] text-right text-[14px]">{{ temp.title_zh }}</div>
      <div class="flex-1 flex flex-col gap-[12px]">
        <div v-for="(item, ind) in temp.items" :key="ind" class="flex gap-[12px] items-center ml-[24px]">
          <Switcher v-model="item.enabled" theme="primary"></Switcher>
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
              <Radio.Button :label="true">开启</Radio.Button>
              <Radio.Button :label="false">关闭</Radio.Button>
            </Radio.Group>
            <Select
              v-else-if="item.type === 3"
              v-model="item.value_string_select"
              :list="getList(index, ind, 'string')"
              multiple
              :disabled="!item.enabled"
              @change="updateValue"
            ></Select>
            <Select
              v-else-if="item.type === 4"
              v-model="item.value_int_select"
              :list="getList(index, ind, 'number')"
              :disabled="!item.enabled"
              multiple
              @change="updateValue"
            ></Select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { ref, computed, watch } from "vue";
import { Switcher, Input, Radio, Select } from "bkui-vue";
import { cloneDeep } from "lodash";
import { useRoute } from 'vue-router';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { onMounted } from "vue";

const emit = defineEmits(["updateConfig"]);
const route = useRoute();
const nodeRole = computed(() => route.params.node_role);
let originTemplates: ConfigPolicyConfigBlock[] = [];
const configTemplates = ref<ConfigPolicyConfigBlock[]>([]);
const typeMap = {
  0: "string",
  1: "number",
  2: "boolean",
  3: "stringSelect",
  4: "numberSelect",
};
const updateValue = () => {
  emit("updateConfig", configTemplates);
};

const getList = (
  index: number,
  ind: number,
  type: 'string' | 'number'
): { value: string | number; label: string }[] => {
  const templateItem = originTemplates[index]?.items[ind];

  if (!templateItem) {
    // return [{ value: "-1", label: "不限" }];
    return [];
  }

  const selectList =
    type === 'string'
      ? templateItem.value_string_select
      : templateItem.value_int_select;

  const mappedList = selectList?.map((item) => ({
    value: item,
    label: String(item),
  })) || []; // 如果未定义，使用空数组以避免错误

  return [
    // { value: "-1", label: "不限" },
    ...mappedList,
  ];
};

// 获取配置
const getConfigs = async () => {
  const res = await ConfigPolicyAPIService.ConfigPolicyTemplate({
    node_role: nodeRole.value
  }).catch(() => ({
    templates: []
  }));
  originTemplates = res.templates;
  configTemplates.value = cloneDeep(res.templates);
}
onMounted(async () => {
  await getConfigs();
});
</script>
