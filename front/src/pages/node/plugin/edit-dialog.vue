<template>
  <div>
    <Dialog
      :is-show="isShow"
      :title="$t('pluginManagement.plugin.edit')"
      render-directive="if"
      :width="600"
      @closed="isShow = false"
    >
      <Form :model="form" form-type="vertical" :rules="rules" ref="formRef" :width="480">
        <Form.FormItem :label="$t('pluginManagement.plugin.table.pluginName')" prop="name" :required="true">
          <Input v-model="form.name" :disabled="true"></Input>
        </Form.FormItem>
        <Form.FormItem :label="$t('pluginManagement.plugin.table.memo')" prop="memo">
          <Input type="textarea" v-model="form.memo" :resize="false" autosize></Input>
        </Form.FormItem>
      </Form>
      <template #footer>
        <Button
          theme="primary"
          :loading="loading"
          class="mr-[8px] w-[64px]"
          @click="handleConfirm">
          {{ $t('action.confirm') }}
        </Button>
        <Button @click="isShow = false" class="w-[64px]">{{ $t('action.cancel') }}</Button>
      </template>
    </Dialog>
  </div>
</template>

<script lang="ts" setup>
import { Button, Dialog, Form, Input } from 'bkui-vue';
import type { PropType } from 'vue';
import { reactive, ref, watch } from 'vue';

import { PluginAPIService } from '@/api/modules/plugin';

interface IPlugin {
  name: string;
  group: string;
  pkg_name: string;
  memo: string;
}

const isShow = defineModel('isShow', { type: Boolean, default: false });

const props = defineProps({
  plugin: {
    type: Object as PropType<IPlugin>,
    default: null,
  },
});

const emit = defineEmits(['confirm']);
const form = reactive<IPlugin>({
  name: '',
  group: '',
  pkg_name: '',
  memo: '',
});

const rules = ref({});

const formRef = ref();
const loading = ref(false);
// 确认
const handleConfirm = async () => {
  try {
    const result = await formRef.value.validate();
    if (!result) return;
    loading.value = true;
    await PluginAPIService.SetPluginMemo({
      plugin_name: form.name,
      memo: form.memo,
    });
    isShow.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
    emit('confirm');
  }
};
  // 清空form
const resetForm = () => {
  form.name = '';
  form.memo = '';
};
  // edit时初始化form数据
const initFormData = async () => {
  form.name = props.plugin?.name || '';
  form.memo = props.plugin?.memo || '';
};

watch(isShow, (curShow: boolean) => {
  if (!curShow) {
    resetForm();
  } else {
    initFormData();
  }
});

</script>
