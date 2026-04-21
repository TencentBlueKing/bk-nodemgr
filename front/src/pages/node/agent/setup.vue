<template>
  <div class="setup pt-[24px] pb-[48px]">
    <setup-tip @show-setting="handleShowSetting"></setup-tip>
    <div class="m-[24px]">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.type')"
          required
        >
          <install-type currentNodeType="agent" @change="handleChange"></install-type>
        </Form.FormItem>
        <Form.FormItem
          property="bk_biz_id"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.installAgentPage.installBusiness') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.installBusinessTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <BizSelect
            select-class="w-[568px]"
            v-model="formData.bk_biz_id"
            action="agent_operate"
            :placeholder="$t('platform.nodeMan.installAgentPage.placeholder.selectBiz')"
          />
        </Form.FormItem>
        <Form.FormItem
          property="bk_networkarea_id"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.installAgentPage.cloud') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <AreaSelector class="w-[568px]" :multiple="false" @change="handleSingleChange" />
        </Form.FormItem>
        <Form.FormItem
          property="bk_networkunit_id"
          required
        >
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.installAgentPage.cloud_unit') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudUnitTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <UnitSelector
            class="w-[568px]"
            v-model="formData.bk_networkunit_id"
            :options="networkUnitList"
            :disabled="!formData.bk_networkarea_id"
            @change="handleNetworkUnitChange"
          />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('platform.nodeMan.installAgentPage.info')"
          required
        >
          <template #label>
            <div class="mr-[2px]">{{ $t('platform.nodeMan.installAgentPage.info') }}</div>
            <!-- 导入弹窗 -->
            <Button text theme="primary" @click="handleExcelImport">
              {{ $t('platform.nodeMan.installAgentPage.excelImport') }}
            </Button>
          </template>
          <install-table
            ref="installTableRef"
            :max-height="520"
            v-model:data="formData.info"
            :current-settings="settings"
          >
            <UploadExcel @upload="handleUpload" v-if="activeInstallType === 'import'"></UploadExcel>
          </install-table>
        </Form.FormItem>
        <Form.FormItem>
          <Button
            text
            theme="primary"
            @click="isShow = !isShow"
          >
            <span class="mr-[8.5px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.AdvancedOptions') }}</span>
            <angle-double-down-line
              :class="['text-[14px]', { 'transform rotate-180': isShow }]"
            />
          </Button>
        </Form.FormItem>
        <Form.FormItem required v-if="isShow">
          <template #label>
            <Popover
              theme="light"
              trigger="hover"
              placement="right"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span
                class="cursor-default"
                style="border-bottom: 1px dashed #c4c6cc"
              >{{ $t('platform.nodeMan.installAgentPage.version') }}</span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.versionTooltipDesc') }}</p>
                  <p class="mt-[8px]">{{ $t('platform.nodeMan.installAgentPage.versionTooltipDefault') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <div class="w-[568px]">
            <Table
              :data="systemData"
              :border="true"
              width="568"
              :empty-text="$t('platform.nodeMan.installAgentPage.noAvailableVersion')"
            >
              <TableColumn
                field="os"
                :title="$t('platform.nodeMan.installAgentPage.osArch')"
                width="200"
              >
                <template #header>
                  <span class="text-[14px]">{{ $t('platform.nodeMan.installAgentPage.osArch') }}</span>
                </template>
                <template #default="{ row }">
                  {{ row.os?.replace('_', '/') }}
                </template>
              </TableColumn>
              <TableColumn field="version" :title="$t('platform.nodeMan.installAgentPage.packageVersion')" width="368">
                <template #header>
                  <span class="mr-[2px] text-[14px]">{{ $t('platform.nodeMan.installAgentPage.packageVersion') }}</span>
                  <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                  <Button text @click="handleBatchEditVersion">
                    <i class="nodeman-icon nc-bulk-edit cursor-pointer text-[14px]"></i>
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
          content: $t('platform.nodeMan.installAgentPage.noAvailableVersionTip'),
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
        {{ $t("action.back") }}
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
      :title="$t('action.import')"
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
import { Button, Dialog, Form, Input, Message, Popover, Upload } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, debounce, isEqual  } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch  } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Preview from './preview.vue';

import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { encryptionTool } from '@/common/crypto';
import { PACKAGE_GENERATION } from '@/common/const';
import { getDefaultLoginMode, scrollToFirstErrorByClassNames } from '@/common/util';
import Validate from '@/components/validate.vue';
import BizSelect from '@/components/biz-select.vue';
import UnitSelector from '@/components/unitSelector.vue';
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
  login_mode: getDefaultLoginMode(),
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
});

// 表头设置
const settings = reactive({
  fields: [
    { title: t('components.installTable.innerIPv4'), field: 'bk_host_innerip' },
    { title: t('components.installTable.innerIPv6'), field: 'bk_host_innerip_v6' },
    { title: t('components.installTable.osType'), field: 'os_type' },
    { title: t('components.installTable.loginIP'), field: 'login_ip' },
    { title: t('components.installTable.port'), field: 'login_port' },
    { title: t('components.installTable.account'), field: 'login_user' },
    { title: t('components.installTable.authMethod'), field: 'login_mode' },
    { title: t('components.installTable.passwordKey'), field: 'credit' },
  ],
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'os_type',
    'login_port',
    'login_ip',
    'login_user',
    'login_mode',
    'credit',
  ],
  disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
  size: 'medium',
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
const isAtBottom = ref(false);
const dialogData = ref([{
  os: '',
  version: '',
}]);
const isBatch = ref(false);

// 切换安装方式
const handleChange = (value: string) => {
  formRef.value?.clearValidate();
};

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
  formData.info = [...formData.info, ...excelImportData.value];
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
const handleNetworkUnitChange = (id: string | number, row: any) => {
  formData.bk_networkunit_name = row?.bk_networkunit_name || '';
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

const handleSingleChange = (id: string, rows: any[]) => {
  // 单选通常用于表单赋值
  formData.bk_networkarea_id = id;
  formData.bk_networkunit_id = '';
  formData.bk_networkarea_name = rows[0].bk_networkarea_name;
};

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnitBrief[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: {
      bk_networkarea_id: [Number(formData.bk_networkarea_id)],
    },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return {
      total: 0,
      items: [],
    };
  });
  networkUnitList.value = res.items;
  if (res.items.length === 1) {
    formData.bk_networkunit_id = String(res.items[0].bk_networkunit_id);
    formData.bk_networkunit_name = String(res.items[0].bk_networkunit_name);
  }
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
      keyfile: 'login_key_file',
    };
    // Deep clone to avoid mutating original formData — cancel preview should preserve credit
    const clonedData = cloneDeep(formData);
    clonedData.info.forEach((item: any) => {
      // 获取对应的 key (login_password 或 login_key_file)
      const targetKey = modeMap[item.login_mode];

      if (item.credit) {
        // 密码用 V1 加密，密钥用 V2 加密
        const encryptedValue = item.login_mode === 'keyfile'
          ? encryptionTool.encryptV2Sync(item.credit)
          : encryptionTool.encryptV1Sync(item.credit);

        // 如果加密成功，使用密文；否则使用空字符串
        item[targetKey] = encryptedValue !== false ? encryptedValue : '';
      } else {
        // 如果没有输入值，直接赋值
        item[targetKey] = item.credit;
      }

      // 清理不需要的字段 — only on the cloned copy
      delete item.credit;
      delete item.bk_host_id;
    });
    if (isShow.value) {
      clonedData.target_version = systemData.value
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
    }
    previewData.data = clonedData;
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
  const res = await PackageService.ListReleaseAgentBrief({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      release_type: ['agent'],
    },
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const osMap: any = {};
  res.items.forEach((item) => {
    const key = `${item.os_type}_${item.cpu_arch}`;
    if (!osMap[key]) {
      osMap[key] = {
        name: key,
        enableVersions: [],
      };
    }
    if (item.enabled) {
      osMap[key].enableVersions.push({
        version: item.version,
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
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
    if (formData.bk_networkarea_id) {
      await getNetworkUnitList();
    }
  },
);
watch(
  () => isShow.value,
  () => {
    checkIfAtBottom();
  },
);
onMounted(async () => {
  encryptionTool.initPublicKey();
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
