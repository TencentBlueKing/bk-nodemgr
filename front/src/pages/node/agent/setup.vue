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
              [{{ option.bk_networkarea_id }}] {{ option.bk_networkarea_name }}
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
              [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <template #label>
            <div class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.info') }}</div>
            <!-- 导入弹窗 -->
            <Button text theme="primary" @click="handleExcelImport">导入</Button>
          </template>
          <install-table
            ref="installTableRef"
            :max-height="520"
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
            <Table :data="systemData" :border="true" width="568" empty-text="当前无可用版本">
              <TableColumn
                field="os"
                title="操作系统/架构"
                width="200"
              >
                <template #default="{ row }">
                  {{ row.os?.replace('_', '/') }}
                </template>
              </TableColumn>
              <TableColumn field="version" :title="$t('platform.nodeMan.installAgentPage.packageVersion')" width="368">
                <template #header>
                  <span class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.packageVersion') }}</span>
                  <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                  <Button text @click="handleBatchEditVersion">
                    <i class="nodeman-icon nc-bulk-edit cursor-pointer"></i>
                  </Button>
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
      <!-- <Button
        v-show="activeInstallType === 'import' && !isEqual(formData.info, excelImportData)"
        class="w-[100px]"
        theme="primary"
        :disabled="!excelImportData.length"
        @click="handleImport">
        {{ '导入' }}
      </Button> -->
      <Button
        v-show="activeInstallType !== 'import'
          || (excelImportData.length > 0 && isEqual(formData.info, excelImportData))"
        class="w-[120px]"
        theme="primary"
        :disabled="systemData.length === 0"
        v-bk-tooltips="{
          content: '当前无可用版本, 不可安装',
          disabled: systemData.length > 0
        }"
        @click="handlePreview"
      >
        <span>{{ $t("platform.nodeMan.installAgentPage.button.install") }}</span>
        <span
          class="mx-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
        >
          {{ formData.info.length }}
        </span>
      </Button
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
      :is-manual="activeInstallType === 'manual'"
    ></preview>
    <choose-version-dialog
      v-model:is-show="isShowDialog"
      :data="dialogData"
      :batch="isBatch"
      :release-type="'agent'"
      @confirm="handleConfirmVersion"
    ></choose-version-dialog>
    <Dialog
      :is-show="isShowExcelImport"
      :width="1048"
      :title="'Excel 导入'"
      @closed="handleExcelImportCancel">
      <UploadExcel ref="uploadExcelRef" @upload="handleUpload"></UploadExcel>
      <template #footer>
        <Button
          :disabled="!uploadExcelRef?.curFile?.data"
          theme="primary"
          class="mr-[8px]"
          @click="handleExcelImportConfirm">
          {{ $t("action.confirm") }}
        </Button>
        <Button class="" @click="handleExcelImportCancel">{{ $t("action.cancel") }}</Button>
      </template>
    </Dialog>
  </div>
</template>
<script lang="ts" setup>
import { Button, Dialog, Form, Input, Message, Select, Upload } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, debounce, isEqual  } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch  } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Preview from './preview.vue';

import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
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
    os: 'linux_amd64',
    version: '',
  },
  {
    os: 'darwin_amd64',
    version: '',
  },
  {
    os: 'linux_arm64',
    version: '',
  },
  {
    os: 'windows_amd64',
    version: '',
  },
]);
const rules = {};
const isShow = ref(false);
const isShowDialog = ref(false);
const businessList = computed(() => mainStore.businessList);
const isAtBottom = ref(false);
const handleSelect = (newValue: string, oldValue: string) => {
  formData.bk_networkarea_name = networkAreaList.value?.find(item => String(item.bk_networkarea_id) === newValue)?.bk_networkarea_name || '';
};
const dialogData = ref([{
  os: '',
  version: '',
}]);
const isBatch = ref(false);
const handleChooseVersion = (row: { version: string; os: string }) => {
  isShowDialog.value = true;
  dialogData.value = [row];
};
// 批量选择版本
const handleBatchEditVersion = () => {
  isShowDialog.value = true;
  dialogData.value = systemData.value;
  isBatch.value = true;
};

// excel 导入弹窗
const isShowExcelImport = ref(false);
const excelImportData = ref([]);
const uploadExcelRef = ref();
const handleExcelImport = () => {
  isShowExcelImport.value = true;
};
const handleExcelImportConfirm = () => {
  formData.info = excelImportData.value;
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleExcelImportCancel = () => {
  isShowExcelImport.value = false;
  uploadExcelRef.value?.handleDelete();
};
const handleUpload = (data: any) => {
  excelImportData.value = data.info;
};
const handleImport = () => {
  formData.info = excelImportData.value;
};

const handleConfirmVersion = (data: any[]) => {
  systemData.value.forEach((sys: { version: string; os: string }) => {
    const find = data.find((item) => sys.os === `${item.os_type}_${item.cup_arch}`);
    if (find) {
      sys.version = find.version;
    }
  });
  isBatch.value = false;
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
      formData.target_version = systemData.value
        .filter((item: any) => !!item.version)
        .map((item) => {
          const [type, cpu_arch] = item.os.split('_');
          const os_type = type;
          return {
            os_type,
            cpu_arch,
            version: item.version,
          };
        });
      formData.disable_default_target_version = true;
    }
    previewData.data = { ...formData };
  } else {
    scrollToFirstErrorByClassNames();
  }
};
const handleSetpBack = () => {
  formData.info = [];
};
const handleCancel = () => {
  router.push({ name: 'agent' });
};

// 获取版本，用来检查是否有对应架构的包版本去安装
const getVersions = async () => {
  const res = await PackageService.ListReleaseAgent({
    generation: 2,
    exact_include_conditions: {
      release_type: ['agent'],
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const osMap: any = {};
  res.items.forEach((item) => {
    const key = `${item.release.os_type}_${item.release.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        enableVersions: [],
      };
    }
    if (item.release.enabled) {
      osMap[key].enableVersions.push({
        version: item.release.version,
        os_type: item.release.os_type,
        cpu_arch: item.release.cpu_arch,
      });
    }
  });
  systemData.value = systemData.value.filter(item => !!osMap[item.os]?.enableVersions.length);
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
  await getVersions();
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
