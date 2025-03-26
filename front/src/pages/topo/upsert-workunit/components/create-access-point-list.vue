<template>
  <div
    v-for="(item, index) in data"
    :key="index"
    class="w-[489px] bg-[#F5F7FA] relative py-[24px] mb-[12px]">
    <!-- 删除接入点 -->
    <i
      class="nodeman-icon nc-delete absolute top-[11px] right-[11px]
        text-[24px] text-[#979BA5] cursor-pointer z-10"
      @click="handleDeleteAccessPoint(index)">
    </i>
    <CreateAccessPoint
      v-model:form="data[index]"
      :ref="el => bindFormRef(el, index)">
    </CreateAccessPoint>
  </div>
  <div
    class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF]
      w-[489px] h-[32px] flex items-center justify-center cursor-pointer"
    @click="handleCreateAccessPoint">
    <i class="nodeman-icon nc-plus-line mr-[7px]"></i>
    <span class="text-[12px]">
      {{ $t('topoManager.workUnit.form.createAccessPoint') }}
    </span>
  </div>
</template>

<script lang="ts" setup>
import { InfoBox } from 'bkui-vue';
import { cloneDeep, isEqual } from 'lodash';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import CreateAccessPoint from './create-access-point.vue';

const data = defineModel<Omit<AccessPoint, 'accesspoint_id' | 'tenant_id'>[]>('accessPoints', { required: true });
const { t } = useI18n();

const initData: Omit<AccessPoint, 'accesspoint_id' | 'tenant_id'> = {
  accesspoint_name: '',
  bk_networkarea_id: 0,
  endpoints: {
    cluster: [''],
    file: [''],
    data: [''],
  },
};

const bindFormRef = (el, index: number) => {
  if (el) {
    formRefs.value[index] = el;
  } else {
    // 找到并移除对应的 null 引用
    const idx = formRefs.value.findIndex(item => item === null);
    if (idx > -1) {
      formRefs.value.splice(idx, 1);
    }
  }
};

const handleDeleteAccessPoint = async (index: number) => {
  if (!isEqual(data.value[index], initData)) {
    InfoBox({
      title: t('topoManager.workUnit.form.deleteTips'),
      onConfirm: () => {
        data.value.splice(index, 1);
        formRefs.value.splice(index, 1);
      },
      cancelText: t('action.cancel'),
    });
  } else {
    data.value.splice(index, 1);
    formRefs.value.splice(index, 1);
  }
};

const handleCreateAccessPoint = () => {
  data.value.push(cloneDeep(initData));
};

const formRefs = ref<typeof CreateAccessPoint[] | null[]>([]);
const validateForms = async () => {
  const results = await Promise.all(formRefs.value.map(formRef => formRef?.validateForm()));
  return results.every(result => result !== false);
};

defineExpose({
  validateForms,
});

</script>
