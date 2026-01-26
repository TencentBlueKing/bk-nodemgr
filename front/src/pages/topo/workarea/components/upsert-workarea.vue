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
            property="cloud_vendor"
            required>
            <Select
              v-model="form.cloud_vendor"
              :filterable="false"
            >
              <template #prefix v-if="form.cloud_vendor">
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
    <GuideDialog v-model:is-show="isGuideShow" @confirm="handleInstallProxy"></GuideDialog>
  </div>
</template>

<script lang="ts" setup>
import { Button, Dialog, Form, Input, Loading, Select } from 'bkui-vue';
import type { PropType } from 'vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { vendorMap } from '../vendorMap';

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
    default: null
  },
});
const emit = defineEmits(['install-proxy', 'update']);
const { t } = useI18n();
const workareaStore = useWorkareaStore();

const dialogTitle = computed(() => (props.isCreate ? t('topoManager.workArea.form.create') : t('topoManager.workArea.form.edit')));
const isGuideShow = ref(false);

// 云服务商下拉选项列表
const SelectOptions = computed(() => workareaStore.vendorList.map((item: string) => {
  const curItem = vendorMap[item];
  return {
    id: item,
    icon: curItem.icon,
    class: curItem?.class,
    label: t(curItem?.label),
  };
}));

// 根据当前cloud_vendor从vendorMap获取对应class, icon
const curVendor = computed(() => SelectOptions.value.find((item: any) => item.id === form.cloud_vendor));

const form = reactive<TopoNetworkAreaCreateReq>({
  bk_networkarea_name: '',
  cloud_vendor: '',
});

const rules = ref({
  bk_networkarea_name: [
    {
      required: true,
      message: t('topoManager.workArea.formRule.workareaName'),
      trigger: 'blur',
    },
  ],
  cloud_vendor: [
    {
      required: true,
      message: t('topoManager.workArea.formRule.vendor'),
      trigger: 'blur',
    },
  ],
});

const handleInstallProxy = () => {
  emit('install-proxy');
};

const formRef = ref();
const loading = ref(false);
// 新建 workarea
const handleConfirm = async () => {
  try {
    const result = await formRef.value.validate();
    if (!result) return;
    loading.value = true;
    let res;
    if (props.isCreate) {
      res = await workareaStore.handleCreateWorkarea(form);
    } else {
      res = await workareaStore.handleUpdateWorkarea({
        bk_networkarea_id: props.curWorkareaData.bk_networkarea_id,
        ...form,
      });
    }
    if (res) {
      if (props.isCreate) {
        isGuideShow.value = true;
      } else {
        emit('update');
      }
    }
    isShow.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};
// 清空form
const resetForm = () => {
  form.cloud_vendor = '';
  form.bk_networkarea_name = '';
};
// edit时初始化form数据
const initFormData = async () => {
  form.cloud_vendor = props.curWorkareaData?.cloud_vendor || '';
  form.bk_networkarea_name = props.curWorkareaData?.bk_networkarea_name || '';
};

watch(isShow, (curShow: boolean) => {
  if (!curShow) {
    resetForm();
  } else {
    if (!props.isCreate) {
      initFormData();
    }
  }
});

</script>
