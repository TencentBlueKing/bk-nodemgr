<template>
  <div>
    <Dialog
      :is-show="isShow"
      :title="dialogTitle"
      render-directive="if"
      @closed="isShow = false"
    >
      <Loading :loading="loading">
        <Form :model="form" form-type="vertical" :rules="rules" ref="formRef" :width="480">
          <Form.FormItem
            :label="$t('topoManager.workArea.form.workareaName')"
            property="bk_networkarea_name"
            required>
            <Input class="w-[432px]" v-model="form.bk_networkarea_name" />
          </Form.FormItem>
          <Form.FormItem
            :label="$t('topoManager.workArea.form.vendor')"
            property="bk_cloud_vendor"
            required>
            <Select
              v-model="form.bk_cloud_vendor"
              :filterable="false"
            >
              <template #prefix v-if="form.bk_cloud_vendor">
                <div class="flex items-center">
                  <img
                    class="h-[18px] w-[18px] rounded-[50px] p-[2px] ml-[8px]"
                    :src="curVendor?.icon"
                    :class="curVendor?.class" />
                </div>
              </template>
              <Select.Option
                v-for="(item, index) in SelectOptions"
                class="!hover:bg-[#E1ECFF] !hover:text-[#3A84FF]"
                :id="item.id"
                :key="index"
                :name="item.label">
                <div class="flex">
                  <img
                    :src="item.icon"
                    class="h-[18px] w-[18px] rounded-[50px] p-[2px] mr-[10px]"
                    :class="item.class" />
                  <span>{{ item.label }}</span>
                </div>
              </Select.Option>
            </Select>
          </Form.FormItem>
        </Form>
      </Loading>
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
    <GuideDialog v-model:is-show="isGuideShow"></GuideDialog>
  </div>
</template>

<script lang="ts" setup>
import { Button, Dialog, Form, Input, Loading, Select } from 'bkui-vue';
import { computed, PropType, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import GuideDialog from './guide-dialog.vue';

import type { TopoNetworkAreaCreateReq } from '@/@types/topo';
import { useWorkareaStore } from '@/stores/workarea';

const isShow = defineModel('isShow', { type: Boolean, default: false });

const props = defineProps({
  isCreate: {
    type: Boolean,
    default: true,
  },
  curWorkareaData: {
    type: Object as PropType<NetworkArea>,
  },
});

const { t } = useI18n();
const workareaStore = useWorkareaStore();

const dialogTitle = computed(() => (props.isCreate ? t('topoManager.workArea.form.create') : t('topoManager.workArea.form.edit')));

const isGuideShow = ref(false);

const vendorMap = {
  tencent: t('topoManager.workArea.vendor.tencent'),
  google: t('topoManager.workArea.vendor.google'),
  huawei: t('topoManager.workArea.vendor.huawei'),
  microsoft: t('topoManager.workArea.vendor.microsoft'),
  aws: 'AWS',
  ali: t('topoManager.workArea.vendor.ali'),
};

const SelectOptions = ref([
  {
    id: 'tencent',
    icon: '/images/tencent-cloud.svg',
    class: 'bg-[#DAE9FD]',
    label: vendorMap.tencent,
  },
  {
    id: 'google',
    icon: '/images/google-cloud.svg',
    class: 'bg-[#DAF5C8]',
    label: vendorMap.google,
  },
  {
    id: 'huawei',
    icon: '/images/huawei-cloud.svg',
    class: 'bg-[#FFDDDD]',
    label: vendorMap.huawei,
  },
  {
    id: 'microsoft',
    icon: '/images/azure.svg',
    class: 'bg-[#D8F4F5]',
    label: vendorMap.microsoft,
  },
  {
    id: 'aws',
    icon: '/images/aws-cloud.svg',
    class: 'bg-[#FFF2C9]',
    label: vendorMap.aws,
  },
  {
    id: 'ali',
    icon: '/images/ali-cloud.svg',
    class: 'bg-[#FFE0BF]',
    label: vendorMap.ali,
  },
]);
const formRef = ref();
const form = reactive<TopoNetworkAreaCreateReq>({
  bk_networkarea_name: '',
  bk_cloud_vendor: '',
});
const curVendor = computed(() => SelectOptions.value.find(item => item.id === form.bk_cloud_vendor));

const rules = ref({
  bk_networkarea_name: [
    {
      required: true,
      message: t('topoManager.workArea.formRule.workareaName'),
      trigger: 'blur',
    },
  ],
  bk_cloud_vendor: [
    {
      required: true,
      message: t('topoManager.workArea.formRule.vendor'),
      trigger: 'blur',
    },
  ],
});

const resetForm = () => {
  form.bk_cloud_vendor = '';
  form.bk_networkarea_name = '';
};
const loading = ref(false);
const handleConfirm = async () => {
  try {
    const result = await formRef.value.validate();
    if (!result) return;
    loading.value = true;
    const res = await workareaStore.handleCreateWorkarea(form);
    if (res && props.isCreate) {
      isGuideShow.value = true;
    }
    isShow.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const initFormData = async () => {
  form.bk_cloud_vendor = props.curWorkareaData?.bk_cloud_vendor || '';
  form.bk_networkarea_name = props.curWorkareaData?.bk_networkarea_name || '';
}

watch(isShow, (curShow) => {
  if (!curShow) {
    resetForm();
  }
  if (!props.isCreate) {
    initFormData();
  }
});

</script>
