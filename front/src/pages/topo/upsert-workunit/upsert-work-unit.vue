<template>
  <div>
    <Sideslider
      v-model:is-show="isShow"
      :title="title"
      :before-close="handleBeforeClose"
      width="640"
    >
      <div class="flex items-center w-full text-[14px] text-[#63656e] pt-[28px]" v-if="isCreate">
        <div class="w-[130px] pr-[22px] text-right">{{ t('topoManager.workUnit.form.workUnitType') }}</div>
        <Radio.Group v-model="type">
          <Radio label="not_direct">{{ $t('topoManager.workUnit.form.notDirect') }}</Radio>
          <Radio label="direct">{{ $t('topoManager.workUnit.form.direct') }}</Radio>
        </Radio.Group>
      </div>
      <Form
        :model="form"
        ref="formRef"
        class="pt-[28px]"
        :rules="rules"
        v-if="(!isCreate && isDirect) || (isCreate && type === 'direct')">
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.workUnitName')"
          property="bk_networkunit_name"
          label-width="130"
          required>
          <Input v-model="form.bk_networkunit_name" class="w-[488px]" clearable />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.directConfig')"
          required
          label-width="130"
        >
          <div class="w-[489px] bg-[#F5F7FA] relative py-[24px] mb-[12px]">
            <CreateDirectAccessPoint v-model:data="form.direct_endpoints">
            </CreateDirectAccessPoint>
          </div>
        </Form.FormItem>
        <template v-if="isExpand">
          <Form.FormItem
            :label="$t('topoManager.workUnit.form.downstream')"
            label-width="130">
            <CreateAccessPointList
              v-model:access-points="form.accesspoints">
            </CreateAccessPointList>
          </Form.FormItem>
          <CustomDeployConfigForm
            v-model:configs="form.custom_deploy_configs"
            v-model:active-os="activeCustomDeployOs"
            :os-options="customDeployOsOptions"
            :default-config="defaultDeployConfigs[activeCustomDeployOs] ?? null"
          />
        </template>
        <Button text theme="primary" class="ml-[130px]" @click="toggleExpand">
          <span>{{ $t('topoManager.workUnit.form.senior') }}</span>
          <i class="nodeman-icon nc-angle-double-down text-[24px]" v-if="!isExpand"></i>
          <i class="nodeman-icon nc-double-up text-[24px]" v-else></i>
        </Button>
      </Form>
      <Form :model="form" class="pt-[28px]" ref="formRef" :rules="rules" v-else>
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.workUnitName')"
          property="bk_networkunit_name"
          label-width="130"
          required>
          <Input v-model="form.bk_networkunit_name" class="w-[488px]" clearable />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.upstream')"
          property="generalLink"
          label-width="130"
          required>
          <SelectGroup
            v-model:link="generalLink"
            :disabled="isExpand"
            :is-create="isCreate"
            :work-unit-id="workUnitId">
          </SelectGroup>
        </Form.FormItem>
        <template v-if="isExpand">
          <Form.FormItem
            class="mt-[-12px]"
            label="cluster"
            property="cluster"
            label-width="130">
            <SelectGroup
              v-model:link="form.links.cluster"
              :work-unit-id="workUnitId"
              :is-create="isCreate">
            </SelectGroup>
          </Form.FormItem>
          <Form.FormItem
            class="mt-[-12px]"
            label="file"
            property="file"
            label-width="130">
            <SelectGroup
              v-model:link="form.links.file"
              :work-unit-id="workUnitId"
              :is-create="isCreate">
            </SelectGroup>
          </Form.FormItem>
          <Form.FormItem
            class="mt-[-12px]"
            label="data"
            property="data"
            label-width="130">
            <SelectGroup
              v-model:link="form.links.data"
              :work-area-list="workAreaList"
              :work-unit-id="workUnitId"
              :is-create="isCreate">
            </SelectGroup>
          </Form.FormItem>
          <Form.FormItem
            :label="$t('topoManager.workUnit.form.downstream')"
            label-width="130">
            <CreateAccessPointList
              v-model:access-points="form.accesspoints">
            </CreateAccessPointList>
          </Form.FormItem>
          <CustomDeployConfigForm
            v-model:configs="form.custom_deploy_configs"
            v-model:active-os="activeCustomDeployOs"
            :os-options="customDeployOsOptions"
            :default-config="defaultDeployConfigs[activeCustomDeployOs] ?? null"
          />
        </template>
        <Button text theme="primary" class="ml-[130px]" @click="toggleExpand">
          <span class="text-[14px]">{{ $t('topoManager.workUnit.form.senior') }}</span>
          <i class="nodeman-icon nc-angle-double-down text-[20px]" v-if="!isExpand"></i>
          <i class="nodeman-icon nc-double-up text-[20px]" v-else></i>
        </Button>
      </Form>
      <template #footer>
        <div class="flex">
          <Button
            theme="primary"
            class="mr-[8px] w-[88px]"
            :loading="saveLoading"
            @click="handleConfirm">
            {{ $t('action.save') }}
          </Button>
          <Button class="w-[88px]" @click="handleBeforeClose">
            {{ $t('action.cancel') }}
          </Button>
        </div>
      </template>
    </Sideslider>
  </div>
</template>

<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Message, Radio, Sideslider } from 'bkui-vue';
import { cloneDeep, isEqual } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import CreateAccessPointList from './components/create-access-point-list.vue';
import CreateDirectAccessPoint from './components/create-direct-access-point.vue';
import CustomDeployConfigForm from './components/custom-deploy-config-form.vue';
import SelectGroup from './components/select-group.vue';
import {
  deployConfigMapFromApi,
  normalizeCustomDeployConfigs,
  validateCustomDeployConfigAllOrNothing,
} from './custom-deploy-config';
import { buildNetworkUnitPayload } from './network-unit-payload';

import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { PACKAGE_GENERATION } from '@/common/const';
import { scrollToFirstErrorByClassNames } from '@/common/util';
import { useWorkareaStore } from '@/stores/workarea';

const isShow = defineModel<boolean>('isShow', { required: true });

const props = defineProps({
  isCreate: {
    type: Boolean,
    default: false,
  },
  workUnitId: {
    type: Number,
    required: true,
  },
});
const emit = defineEmits(['save']);
const {
  handleFetchAllWorkarea,
  handleFetchAllWorkUnit,
} = useWorkareaStore();
const workareaStore = useWorkareaStore();

const { t } = useI18n();
const route = useRoute();
const workareaId = Number(route.params.workarea);

const type = ref('not_direct');
const isDirect = ref(false);
const isExpand = ref(false);
const toggleExpand = () => {
  const curState = !isExpand.value;
  isExpand.value = curState;
  if (curState) {
    Object.assign(form.links.cluster, generalLink.value);
    Object.assign(form.links.file, generalLink.value);
    Object.assign(form.links.data, generalLink.value);
  }
  formRef.value.clearValidate();
};

const formRef = ref();

const generalLink = ref<Record<keyof Link, number | undefined>>({
  bk_networkarea_id: undefined,
  bk_networkunit_id: undefined,
  accesspoint_id: undefined,
});

const form = reactive({
  tenant_id: '',
  bk_networkunit_name: '',
  bk_networkarea_id: workareaId,
  generation: PACKAGE_GENERATION,
  custom_deploy_configs: {} as Record<string, CustomDeployConfig>,
  accesspoints: [] as AccessPoint[],
  direct_endpoints: {
    cluster: [''],
    file: [''],
    data: [''],
  },
  links: {
    cluster: {},
    file: {},
    data: {},
  },
});

const activeCustomDeployOs = ref('');
const customDeployOsOptions = ref<{ id: string; name: string }[]>([]);
const defaultDeployConfigs = ref<Record<string, CustomDeployConfig>>({});
const DEFAULT_CUSTOM_DEPLOY_OS = 'linux';

const syncCustomDeployConfigs = (
  config?: Record<string, CustomDeployConfig>,
  preferredOs?: string,
) => {
  const configMap = deployConfigMapFromApi(config);
  form.custom_deploy_configs = normalizeCustomDeployConfigs(
    configMap,
    customDeployOsOptions.value.map(item => item.id),
  );
  const configuredOsList = Object.keys(configMap);
  const normalizedOsList = Object.keys(form.custom_deploy_configs);
  const defaultOs = normalizedOsList.includes(DEFAULT_CUSTOM_DEPLOY_OS)
    ? DEFAULT_CUSTOM_DEPLOY_OS
    : undefined;
  activeCustomDeployOs.value = preferredOs
    || defaultOs
    || configuredOsList[0]
    || normalizedOsList[0]
    || '';
};

const initForm = () => {
  form.tenant_id = '';
  form.bk_networkunit_name = '';
  form.generation = PACKAGE_GENERATION;
  syncCustomDeployConfigs();
  form.accesspoints = [];
  form.direct_endpoints = {
    cluster: [''],
    file: [''],
    data: [''],
  };
  form.links = {
    cluster: {},
    file: {},
    data: {},
  };
  Object.assign(generalLink.value, {
    bk_networkarea_id: undefined,
    bk_networkunit_id: undefined,
    accesspoint_id: undefined,
  });
  isExpand.value = false;
};

const fetchCustomDeployOsOptions = async () => {
  const distinctParams = {
    generation: PACKAGE_GENERATION,
    exact_include_conditions: {
      enabled: [true],
    },
    distinct_field: {
      os_type: true,
    },
  };
  const res = await PackageService.DistinctReleaseAgent(distinctParams as any).catch(() => null);
  customDeployOsOptions.value = res?.os_type?.map(item => ({ id: item, name: item })) ?? [];
};

const fetchDefaultDeployConfig = async (osType: string) => {
  if (!osType || defaultDeployConfigs.value[osType] !== undefined) {
    return;
  }
  const res = await TopoService.DefaultDeployConstantGet({
    generation: PACKAGE_GENERATION,
    os_type: osType,
  }).catch(() => null);
  if (res?.default_deploy_config) {
    defaultDeployConfigs.value[osType] = res.default_deploy_config;
  }
};

const title = computed(() => (
  props.isCreate
    ? t('topoManager.workUnit.title.create')
    : t('topoManager.workUnit.title.edit')
));

const rules = reactive({
  bk_networkunit_name: [{
    required: true,
    trigger: 'blur',
  }],
  generalLink: [{
    trigger: 'blur',
    required: true,
    validator: () => validateLink(generalLink.value, true),
  }],
  cluster: [{
    trigger: 'blur',
    required: true,
    validator: () => validateLink(form.links.cluster),
  }],
  file: [{
    trigger: 'blur',
    required: true,
    validator: () => validateLink(form.links.file),
  }],
  data: [{
    trigger: 'blur',
    required: true,
    validator: () => validateLink(form.links.data),
  }],
});
// 上游接入点表单校验
function validateLink(
  link: {
    bk_networkarea_id?: number | undefined;
    bk_networkunit_id?: number | undefined;
    accesspoint_id?: number | undefined;
  },
  isGeneralLink = false,
): boolean {
  // 判断是否所有必填字段都有值
  const isAllFieldsValid = (
    link.bk_networkarea_id !== undefined
    && link.bk_networkunit_id !== undefined
    && link.accesspoint_id !== undefined
  );

  // generalLink 的验证逻辑与其他链接相反
  if (isGeneralLink) {
    // 如果未展开高级选项，需要验证所有字段
    return isExpand.value || isAllFieldsValid;
  }
  // 其他链接的验证逻辑：仅在展开高级选项时需要验证
  return !isExpand.value || isAllFieldsValid;
}

const saveLoading = ref(false);

const handleConfirm = async () => {
  try {
    saveLoading.value = true;
    const validate = await formRef.value?.validate().catch(() => false);
    if (!validate) {
      scrollToFirstErrorByClassNames();
      return;
    };

    // Validate all-or-nothing constraint for custom deploy configs
    const partialConfigs = Object.entries(form.custom_deploy_configs)
      .filter(([, config]) => !validateCustomDeployConfigAllOrNothing(config));
    if (partialConfigs.length > 0) {
      Message({ theme: 'error', message: t('topoManager.workUnit.form.customDeployConfigPartialError') });
      return;
    }

    // params配置
    const links = form.links as Links;
    const generalLinkData = generalLink.value as Link;
    if (!isExpand.value) {
      links.cluster = generalLinkData;
      links.file = generalLinkData;
      links.data = generalLinkData;
    }
    const params = buildNetworkUnitPayload({
      form,
      links,
      isCreate: props.isCreate,
      workUnitId: props.workUnitId,
      type: type.value,
      isDirect: isDirect.value,
    });
    let res;
    if (props.isCreate) {
      res = await TopoService.NetworkUnitCreate(params).catch(() => ({
        bk_networkunit_id: null,
      }));
      Message({
        theme: 'success',
        message: t('message.success.create'),
      });
    } else {
      res = await TopoService.NetworkUnitUpdate(params).catch(() => ({
        bk_networkunit_id: null,
      }));
      Message({
        theme: 'success',
        message: t('message.success.edit'),
      });
    }
    if (res.bk_networkunit_id === null) return;
    // 关闭侧栏
    handleClose();
    // 初始化表单
    initForm();
    // 触发save 父组件刷新list
    emit('save', res.bk_networkunit_id);
  } catch (err) {
    console.error(err);
  } finally {
    saveLoading.value = false;
  }
};

const handleClose = () => {
  isShow.value = false;
};

const originData = ref<any>();
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve) => {
  // 没有修改，直接关闭
  if (isEqual(form, originData.value)) {
    resolve(true);
    isShow.value = false;
    return;
  }
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => resolve(false),
  });
});
const workAreaList = ref<{
  label: string
  value: number
}[]>([]);

const getWorkUnit = async () => {
  // 从缓存获取当前管控区的所有管控单元
  const workUnitList = workareaStore.allWorkUnitList.get(workareaId);
  // 找到当前管控单元数据
  const curWorkUnit = workUnitList?.find((unit: NetworkUnit) => unit.bk_networkunit_id === props.workUnitId) as NetworkUnit;
  isDirect.value = curWorkUnit.is_direct;
  // 数据回填
  form.tenant_id = curWorkUnit.tenant_id;
  form.bk_networkunit_name = curWorkUnit.bk_networkunit_name;
  form.generation = curWorkUnit.generation ?? PACKAGE_GENERATION;
  syncCustomDeployConfigs(curWorkUnit.custom_deploy_config);

  // 判断 cluster, file, data 数据是否一致
  // 以及是否有下游接入点 决定高级是否展开
  // 并完成表单数据初始化
  const { cluster, file, data } = curWorkUnit.links;
  const equal = isEqual(cluster, file) && isEqual(file, data);
  if (curWorkUnit.accesspoints.length !== 0 || !equal) {
    isExpand.value = true;
    Object.assign(form.links.cluster, cluster);
    Object.assign(form.links.file, file);
    Object.assign(form.links.data, data);
    Object.assign(form.accesspoints, curWorkUnit.accesspoints);
  }
  if (curWorkUnit.accesspoints.length === 0 && equal) {
    Object.assign(generalLink.value, cluster);
  }
  Object.assign(form.direct_endpoints, curWorkUnit.direct_endpoints);
};

watch(isShow, async (isCurrentShow: boolean) => {
  if (isCurrentShow) {
    type.value = 'not_direct';
    await Promise.all([
      handleFetchAllWorkarea(),
      handleFetchAllWorkUnit(),
      fetchCustomDeployOsOptions(),
    ]);

    if (props.isCreate) {
      syncCustomDeployConfigs();
    }
    if (!props.isCreate) {
      getWorkUnit();
    }
    originData.value = cloneDeep(form);
  }
  if (!isCurrentShow) {
    formRef.value.clearValidate();
    initForm();
  }
});

watch(activeCustomDeployOs, (osType) => {
  fetchDefaultDeployConfig(osType);
});
</script>
