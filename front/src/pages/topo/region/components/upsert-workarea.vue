<template>
  <div>
    <Dialog
      :is-show="isShow"
      :title="dialogTitle"
      @confirm="handleConfirm"
      @closed="isShow = false"
      render-directive="if"
    >
      <Loading :loading="loading">
        <Form :model="form" form-type="vertical" :rules="rules" ref="formRef" :width="480">
          <Form.FormItem
            :label="$t('topoManager.region.workarea.form.workareaName')"
            property="bkNetworkareaName"
            required>
            <Input class="w-[432px]" v-model="form.bkNetworkareaName" />
          </Form.FormItem>
          <Form.FormItem
            :label="$t('topoManager.region.workarea.form.vendor')"
            property="bkCloudVendor"
            required>
            <Select
              v-model="vendor"
              :filterable="false"
            >
              <template #prefix v-if="vendor">
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
    </Dialog>
    <GuideDialog v-model:is-show="isGuideShow"></GuideDialog>
  </div>
</template>

<script lang="ts" setup>
import { Dialog, Form, Input, Loading, Select } from 'bkui-vue';
import { computed, reactive, ref } from 'vue';
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
});

const { t } = useI18n();
const workareaStore = useWorkareaStore();

const dialogTitle = computed(() => (props.isCreate ? t('topoManager.region.workarea.form.create') : t('topoManager.region.workarea.form.edit')));

const vendor = ref();
const curVendor = computed(() => SelectOptions.value.find(item => item.id === vendor.value));

const isGuideShow = ref(false);

const vendorMap = {
  tencent: t('topoManager.region.workarea.vendor.tencent'),
  google: t('topoManager.region.workarea.vendor.google'),
  huawei: t('topoManager.region.workarea.vendor.huawei'),
  microsoft: t('topoManager.region.workarea.vendor.microsoft'),
  aws: 'AWS',
  ali: t('topoManager.region.workarea.vendor.ali'),
};

const SelectOptions = ref([
  {
    id: 'tencent',
    icon: '/tencent-cloud.svg',
    class: 'bg-[#DAE9FD]',
    label: vendorMap.tencent,
  },
  {
    id: 'google',
    icon: '/google-cloud.svg',
    class: 'bg-[#DAF5C8]',
    label: vendorMap.google,
  },
  {
    id: 'huawei',
    icon: '/huawei-cloud.svg',
    class: 'bg-[#FFDDDD]',
    label: vendorMap.huawei,
  },
  {
    id: 'microsoft',
    icon: '/azure.svg',
    class: 'bg-[#D8F4F5]',
    label: vendorMap.microsoft,
  },
  {
    id: 'aws',
    icon: '/aws-cloud.svg',
    class: 'bg-[#FFF2C9]',
    label: vendorMap.aws,
  },
  {
    id: 'ali',
    icon: '/ali-cloud.svg',
    class: 'bg-[#FFE0BF]',
    label: vendorMap.ali,
  },
]);
const formRef = ref();
const form = reactive({
  bkNetworkareaName: '',
  bkCloudVendor: 0,
} as TopoNetworkAreaCreateReq);

const rules = ref({
  bkNetworkareaName: [
    {
      required: true,
      message: t('topoManager.region.workarea.formRule.workareaName'),
      trigger: 'blur',
    },
  ],
  bkCloudVendor: [
    {
      required: true,
      message: t('topoManager.region.workarea.formRule.vendor'),
      trigger: 'blur',
    },
  ],
});

const initForm = () => {
  form.bkCloudVendor = 0;
  form.bkNetworkareaName = '';
  vendor.value = '';
};
const loading = ref(false);
const handleConfirm = async () => {
  const result = await formRef.value.validate();
  if (!result) return;
  loading.value = true;
  const res = await workareaStore.handleCreateWorkarea({
    bk_networkarea_name: form.bkNetworkareaName,
    bk_cloud_vendor: form.bkCloudVendor,
  });
  loading.value = false;
  if (res) {

  }
  isShow.value = false;
  initForm();
  isGuideShow.value = true;
};

</script>
