<template>
  <div class="flex">
    <Select
      v-model="link.bk_networkarea_id"
      class="mr-[8px] w-[164px]"
      :allow-empty-values="[0]"
      @change="handleWorkareaChange"
      :disabled="disabled">
      <Select.Option
        v-for="item in curWorkAreaList"
        :key="item.bk_networkarea_id"
        :id="item.bk_networkarea_id"
        :name="item.bk_networkarea_name">
      </Select.Option>
    </Select>
    <Select
      v-model="link.bk_networkunit_id"
      class="mr-[8px] w-[154px]"
      :allow-empty-values="[0]"
      @change="handleWorkUnitChange"
      :disabled="disabled">
      <Select.Option
        v-for="item in curWorkUnitList"
        :key="item.bk_networkunit_id"
        :id="item.bk_networkunit_id"
        :name="item.bk_networkunit_name">
      </Select.Option>
    </Select>
    <Select
      v-model="link.accesspoint_id"
      class="mr-[8px] w-[154px]"
      :allow-empty-values="[0]"
      :disabled="disabled">
      <Select.Option
        v-for="item in curAccessPointList"
        :key="item.accesspoint_id"
        :id="item.accesspoint_id"
        :name="item.accesspoint_name">
      </Select.Option>
    </Select>
  </div>
</template>

<script lang="ts" setup>
import { Select } from 'bkui-vue';
import { computed } from 'vue';

import { useWorkareaStore } from '@/stores/workarea';

const link = defineModel<Partial<Link>>('link', { required: true });

defineProps({
  disabled: {
    type: Boolean,
    default: false,
  },
});

const { allWorkareaList, allWorkUnitList, allAccessPointList } = useWorkareaStore();

const curWorkAreaList = computed(() => Array.from(allWorkareaList.values()));
const curWorkUnitList = computed(() => Array.from(allWorkUnitList.get(link.value.bk_networkarea_id as number) || []));
// eslint-disable-next-line max-len
const curAccessPointList = computed(() => Array.from(allAccessPointList.get(link.value.bk_networkunit_id as number) || []));

const handleWorkareaChange = () => {
  link.value.bk_networkunit_id = undefined;
  link.value.accesspoint_id = undefined;
};

const handleWorkUnitChange = () => {
  link.value.accesspoint_id = undefined;
};

</script>
