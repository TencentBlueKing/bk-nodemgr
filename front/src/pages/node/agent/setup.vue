<template>
  <div class="setup pt-[24px] pb-[48px]">
    <setup-tip @show-setting="handleShowSetting"></setup-tip>
    <div class="m-[24px]">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.type')"
          required
        >
          <install-type></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.business')"
          property="bk_biz_id"
          required
        >
          <Select
            class="w-[568px]"
            v-model="formData.bk_biz_id"
            auto-focus
            filterable
            :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.selectBiz')"
          >
            <Select.Option
              v-for="item in businessList"
              :key="item.bk_biz_id"
              :name="item.bk_biz_name"
              :id="item.bk_biz_id"
            >
              [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.cloud')"
          property="bk_networkarea_id"
          required
        >
          <Select
            class="w-[568px]"
            v-model="formData.bk_networkarea_id"
            auto-focus
            filterable
            @select="handleSelect"
          >
            <Select.Option
              v-for="option in networkAreaList"
              :key="option.bk_networkarea_id"
              :id="String(option.bk_networkarea_id)"
              :name="option.bk_networkarea_name"
            >
              {{ option.bk_networkarea_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.cloud_unit')"
          property="bk_networkunit_id"
          required
        >
          <Select
            class="w-[568px]"
            v-model="formData.bk_networkunit_id"
            auto-focus
            filterable
            :disabled="!formData.bk_networkarea_id"
          >
            <Select.Option
              v-for="option in networkUnitList"
              :key="option.bk_networkarea_id"
              :id="String(option.bk_networkunit_id)"
              :name="option.bk_networkunit_name"
            >
              {{ option.bk_networkunit_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <install-table
            ref="installTableRef"
            v-model:data="formData.info"
          >
            <UploadExcel @upload="handleUpload" v-if="activeInstallType === 'import'"></UploadExcel>
          </install-table>
        </Form.FormItem>
        <Form.FormItem>
          <Button
            text
            theme="primary"
            class="text-[14px]"
            @click="isShow = !isShow"
          >
            <span class="mr-[8.5px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
            <angle-double-down-line
              :class="{ 'transform rotate-180': isShow }"
            />
          </Button>
        </Form.FormItem>
        <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.version')" required v-if="isShow">
          <div class="w-[568px]">
            <Table :data="systemData" :border="true" width="568">
              <TableColumn
                field="os"
                :title="$t('platform.nodeMan.os_type')"
                width="200"
              ></TableColumn>
              <TableColumn field="version" :title="$t('platform.nodeMan.installAgentPage.packageVersion')" width="368">
                <template #header>
                  <span class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.packageVersion') }}</span>
                  <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                  <i class="nodeman-icon nc-bulk-edit"></i>
                </template>
                <template #default="{ row }">
                  <Validate
                    :value="row.version"
                    required
                    :ref="(el) => setInputRef(row.os, el)"
                  >
                    <Input
                      :model-value="row.version"
                      :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.select')"
                      @click="handleChooseVersion(row)"
                    />
                  </Validate>
                </template>
              </TableColumn>
            </Table>
          </div>
        </Form.FormItem>
      </Form>
    </div>
    <div
      :class="[
        'h-[48px] w-full flex gap-[8px] items-center pl-[174px]',
        { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom },
      ]"
      ref="footerRef"
    >
      <Button
        v-if="excelImportData.length && formData.info.length === 0"
        class="w-[100px]"
        theme="primary"
        @click="handleImport">
        {{ '导入' }}
      </Button>
      <Button
        class="w-[100px]"
        theme="primary"
        @click="handlePreview"
      >{{ $t("platform.nodeMan.installAgentPage.button.install") }}</Button
      >
      <Button
        v-if="excelImportData.length && activeInstallType === 'import' && formData.info.length > 0"
        class="w-[88px]"
        @click="handleSetpBack">
        {{ '上一步' }}
      </Button>
      <Button class="w-[88px]" @click="handleCancel">{{ $t("action.cancel") }}</Button>
    </div>
    <preview
      v-model:is-show="previewData.isShow"
      :data="previewData.data"
    ></preview>
    <chooseVersionDialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :release-type="'agent'"
      @confirm="handleComfirmVerion"
    ></chooseVersionDialog>
  </div>
</template>
<script lang="ts" setup>
import { Button, Form, Input, Message, Select, Upload } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, debounce  } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch  } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Preview from './preview.vue';

import { TopoService } from '@/api/modules/topo';
import { capitalizeFirstLetter } from '@/common/util';
import chooseVersionDialog from '@/components/choose-version-dialog.vue';
import Validate from '@/components/validate.vue';
import { useMainStore } from '@/stores/main';

const router = useRouter();
const initData = {
  bk_addressing: 'static',
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  os_type: '',
  login_ip: '',
  login_port: '',
  login_user: '',
  login_mode: 'password',
  login_password: '',
  login_key_file: '',
  bk_networkunit_id: '',
  bk_biz_id: '',
  re_register: false,
  credit: '',
};
const mainStore = useMainStore();
const formData = reactive({
  type: '',
  bk_biz_id: '',
  bk_networkarea_id: '',
  bk_networkunit_id: '',
  bk_networkarea_name: '',
  bk_networkunit_name: '',
  bk_host_name: '',
  info: [cloneDeep(initData)],
  target_version: [] as any[],
  disable_default_target_version: false,
});
const previewData = reactive({
  isShow: false,
  data: null,
});
const systemData = ref([
  {
    os: 'Linux_amd64',
    version: '',
  },
  {
    os: 'Darwin_amd64',
    version: '',
  },
  {
    os: 'Linux_arm64',
    version: '',
  },
  {
    os: 'Windows_amd64',
    version: '',
  },
]);
const rules = {};
const isShow = ref(false);
const isShowDialog = ref(false);
const businessList = computed(() => mainStore.businessList);
const isAtBottom = ref(false);
const handleSelect = (newValue: string, oldValue: string) => {
  formData.bk_networkarea_name =    networkAreaList.value?.find(item => String(item.bk_networkarea_id) === newValue)?.bk_networkarea_name || '';
};
const dialogData = ref({
  os: '',
  version: '',
});
const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = row;
};
const handleComfirmVerion = (val: string) => {
  dialogData.value.version = val;
};
// 安装方式
const activeInstallType = computed(() => mainStore.agentSetupType);
const networkAreaList = ref<NetworkArea[]>([]);
// 管控区域下拉列表获取
const getNetworkAreaList = async () => {
  const res = await TopoService.NetworkAreaList({
    page: {
      limit: 0,
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  networkAreaList.value = res.items;
};

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkarea_id: [Number(formData.bk_networkarea_id)],
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;
};
const excelImportData = ref([]);
// excel 导入
const handleUpload = (data: any) => {
  excelImportData.value = data.info;
};
const handleImport = () => {
  formData.info = excelImportData.value;
};

const formRef = ref(null);
const installTableRef = ref(null);
const inputRefs = ref<Map<string, InstanceType<typeof Validate>>>(new Map());
const setInputRef = (
  os: string,
  el: InstanceType<typeof Validate> | null,
) => {
  if (el) {
    const key = os;
    inputRefs.value.set(key, el);
  }
};
const systemValidate = async () => {
  const refs = Array.from(inputRefs.value.values());
  const validate = [];
  for (const item of refs) {
    validate.push(item.validate('blur'));
  }
  const result = await Promise.all(validate);
  return result.every(item => item);
};

// 显示表格设置
const handleShowSetting = () => {
  installTableRef.value?.showSetting();
};

const handlePreview = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isShow.value ? systemValidate() : true,
  ]);
  // 合并多重Promise
  if (Array.isArray(result[2])) {
    result[2] = result[2].every(item => item);
  }
  if (result.every(item => item)) {
    previewData.isShow = true;
    const modeMap = {
      password: 'login_password',
      key: 'login_key_file',
    };
    formData.info.forEach((item) => {
      item[modeMap[item.login_mode]] = item.credit;
      delete item.bk_host_id;
    });
    if (isShow.value) {
      formData.target_version = systemData.value.map((item) => {
        const [type, cpu_arch] = item.os.split('_');
        const os_type = capitalizeFirstLetter(type);
        return {
          os_type,
          cpu_arch,
          version: item.version,
        };
      });
      formData.disable_default_target_version = true;
    }
    previewData.data = { ...formData };
  }
};
const handleSetpBack = () => {
  formData.info = [];
};
const handleCancel = () => {
  router.push({ name: 'agent' });
};
const footerRef = ref<Element | null>(null);
const checkIfAtBottom = () => {
  if (footerRef.value) {
    let { bottom } = footerRef.value.getBoundingClientRect();
    if (isShow.value) {
      bottom += 224;
    } else {
      bottom -= 224;
    }
    isAtBottom.value = bottom >= window.innerHeight;
  }
};

const debouncedCheck = debounce(checkIfAtBottom, 100);
watch(() => activeInstallType.value, () => {
  excelImportData.value = [];
  if (activeInstallType.value === 'import') {
    formData.info = [];
  } else {
    formData.info = [cloneDeep(initData)];
  }
  formRef.value?.clearValidate();
}, { immediate: true });
watch(
  () => formData.bk_networkarea_id,
  async () => {
    await getNetworkUnitList();
  },
);
watch(
  () => isShow.value,
  () => {
    checkIfAtBottom();
  },
);
onMounted(async () => {
  await getNetworkAreaList();
  if (footerRef.value) {
    window.addEventListener('resize', debouncedCheck);
    checkIfAtBottom();
  }
});
onUnmounted(() => {
  if (footerRef.value) {
    window.removeEventListener('resize', debouncedCheck);
  }
});
</script>
