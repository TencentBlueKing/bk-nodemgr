<template>
  <div class="setup pt-[24px] pb-[48px]">
    <div
      class="mx-[24px] flex min-h-[56px] bg-[#F0F8FF] border border-[#C5DAFF] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
    >
      <i class="nodeman-icon nc-tips pt-[2px] text-[#3A84FF]"></i>
      <div class="text-[#4d4f56] text-[12px] leading-[20px] text-left">
        <i18n-t keypath="platform.nodeMan.installAgentPage.tip1" tag="p">
          <span class="text-[#313238] font-bold">{{
            $t("platform.nodeMan.installAgentPage.tip1FirstSlotText")
          }}</span>
          <span class="text-[#313238] font-bold">{{
            $t("platform.nodeMan.installAgentPage.tip1SecondSlotText")
          }}</span>
        </i18n-t>
        <i18n-t keypath="platform.nodeMan.installAgentPage.tip2" tag="p">
          <span class="text-[#313238] font-bold">{{
            $t("platform.nodeMan.installAgentPage.tip2FirstSlotText")
          }}</span>
          <Button text theme="primary" @click="handleShowPanel">{{
            $t("platform.nodeMan.installAgentPage.tip2SecondSlotText")
          }}</Button>
          <Button text theme="primary" @click="handleShowSetting">{{
            $t("platform.nodeMan.installAgentPage.tip2ThirdSlotText")
          }}</Button>
        </i18n-t>
      </div>
    </div>
    <div class="m-[24px]" v-if="activeInstallType === 'import'">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.type')"
          required
        >
          <install-type></install-type>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <install-table ref="installTableRef" v-model:data="formData.info">
            <Upload
              class="p-[24px]"
              :accept="'.xlsx'"
              :handle-res-code="handleRes"
              :select-change="handleSelectChange"
              :url="'https://jsonplaceholder.typicode.com/posts/'"
              :files="fileList"
              with-credentials
              @done="handleDone"
              @error="handleError"
              @progress="handleProgress"
              @success="handleSuccess"
              :before-upload="handleBeforeUpload"
            >
              <template #tip>
                <div class="flex items-center gap-[3px]">
                  <span>{{ '仅支持 .xlsx 类型文件，下载'}}</span>
                  <a :href="url" download="bk_nodeman_info.xlsx">
                    <Button text theme="primary">
                      {{ '模版文件' }}
                    </Button>
                  </a>
                </div>
              </template>
            </Upload>
          </install-table>
        </Form.FormItem>
      </Form>
    </div>
    <div class="m-[24px]" v-else>
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
          ></install-table>
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
                    :ref="`${row.os}_ref`"
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
        'h-[48px] w-full flex items-center pl-[174px]',
        { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom },
      ]"
      ref="footerRef"
    >
      <Button
        class="w-[100px] mr-[8px]"
        theme="primary"
        @click="handlePreview"
      >{{ $t("platform.nodeMan.installAgentPage.button.install") }}</Button
      >
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
  prove: '',
};
const mainStore = useMainStore();
const showRightPanel = ref(false);
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
// 显示侧边栏安装策略
const handleShowPanel = () => {
  showRightPanel.value = true;
};
// 显示表格设置
const handleShowSetting = () => {};

// Excel 导入
const fileList = ref<File[]>([]);
const url = `${window.location.origin}${import.meta.env.BK_SITE_URL}${
  import.meta.env.BK_API_PREFIX
}api/excel/download`;
const handleSuccess = (file: File, fileList: File[]) => {};
const handleProgress = (event: Event, file: File, fileList: File[]) => {};
const handleError = (
  file: File,
  fileList: File[],
  error: { message: string },
) => {
  Message({
    theme: 'error',
    message: error.message,
  });
};
const handleDone = (curFileList: File[]) => {
  fileList.value = [...curFileList];
};
const handleRes = (response: { id: number | string }) => {
  if (response.id) {
    return true;
  }
  return false;
};
const handleBeforeUpload = (file: File, fileList: File[]) => {
  const whiteList = ['xlsx'];
  const AllFiles = fileList.filter(v => whiteList.includes(v.name.substring(file.name.lastIndexOf('.') + 1)));
  if (AllFiles.length !== fileList.length) {
    fileList.pop();
    Message({
      theme: 'warning',
      message: '仅支持 .xlsx 类型文件',
    });
    return false;
  }
  return true;
};
const handleSelectChange = (event: Event) => {};

const formRef = ref(null);
const installTableRef = ref(null);
const Linux_amd64_ref = ref();
const Darwin_amd64_ref = ref();
const Linux_arm64_ref = ref();
const Windows_amd64_ref = ref();
const handlePreview = async () => {
  const result = await Promise.all([
    formRef.value?.validate().catch(() => false),
    installTableRef.value?.tableValidate(),
    isShow.value
      ? Promise.all([
        Linux_amd64_ref.value?.validate('blur').catch(() => false),
        Darwin_amd64_ref.value?.validate('blur').catch(() => false),
        Linux_arm64_ref.value?.validate('blur').catch(() => false),
        Windows_amd64_ref.value?.validate('blur').catch(() => false),
      ])
      : true,
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
      item[modeMap[item.login_mode]] = item.prove;
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

const handleCancel = () => {
  router.push({ name: 'agent' });
};
const debouncedCheck = debounce(checkIfAtBottom, 100);
watch(
  () => formData.bk_networkarea_id,
  async () => {
    await getNetworkUnitList();
  },
);
watch(
  () => isShow.value,
  (val: boolean) => {
    checkIfAtBottom();
  },
);
watch(
  () => activeInstallType.value,
  (val: string) => {
    // installTableRef.value?.clearTableValidate();
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
