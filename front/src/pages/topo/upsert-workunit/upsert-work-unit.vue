<template>
  <div>
    <Sideslider
      v-model:is-show="isShow"
      :title="title"
      :before-close="handleBeforeClose"
      width="640"
    >
      <div class="flex items-center w-full text-[14px] text-[#63656e] pt-[28px]" v-if="isCreate">
        <div class="w-[130px] pr-[22px] text-right">管控单元类型</div>
        <Radio.Group v-model="type">
          <Radio label="not_direct">非直连</Radio>
          <Radio label="direct">直连</Radio>
        </Radio.Group>
      </div>
      <Form :model="form" ref="formRef" class="pt-[28px]" :rules="rules" v-if="(!isCreate && isDirect) || (isCreate && type === 'direct')">
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.workUnitName')"
          property="bk_networkunit_name"
          label-width="130"
          required>
          <Input v-model="form.bk_networkunit_name" class="w-[488px]" clearable />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('topoManager.workUnit.form.directConfig')"
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
              v-model:access-points="form.accesspoints"
              ref="accessPointRef">
            </CreateAccessPointList>
          </Form.FormItem>
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
              v-model:access-points="form.accesspoints"
              ref="accessPointRef">
            </CreateAccessPointList>
          </Form.FormItem>
        </template>
        <Button text theme="primary" class="ml-[130px]" @click="toggleExpand">
          <span>{{ $t('topoManager.workUnit.form.senior') }}</span>
          <i class="nodeman-icon nc-angle-double-down text-[24px]" v-if="!isExpand"></i>
          <i class="nodeman-icon nc-double-up text-[24px]" v-else></i>
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
import { isEqual } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import CreateAccessPointList from './components/create-access-point-list.vue';
import CreateDirectAccessPoint from './components/create-direct-access-point.vue';
import SelectGroup from './components/select-group.vue';

import type { TopoNetworkUnitCreateReq, TopoNetworkUnitUpdateReq } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';
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
  bk_networkunit_name: '',
  bk_networkarea_id: workareaId,
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

const initForm = () => {
  form.bk_networkunit_name = '';
  form.accesspoints = [];
  form.direct_endpoints = {
    cluster: [''],
    file: [''],
    data: [''],
  },
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

const title = computed(() => (props.isCreate ? t('topoManager.workUnit.title.create') : t('topoManager.workUnit.title.edit')));

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

// 下游接入点组件 用于表单校验
const accessPointRef = ref();
const saveLoading = ref(false);

const handleConfirm = async () => {
  try {
    saveLoading.value = true;
    const validate = await formRef.value?.validate().catch(() => false);
    if (!validate) {
      scrollToFirstErrorByClassNames();
      return;
    };

    // params配置
    const links = form.links as Links;
    const generalLinkData = generalLink.value as Link;
    if (!isExpand.value) {
      links.cluster = generalLinkData;
      links.file = generalLinkData;
      links.data = generalLinkData;
    }
    const params: TopoNetworkUnitCreateReq = {
      bk_networkunit_name: form.bk_networkunit_name,
      bk_networkarea_id: form.bk_networkarea_id,
      accesspoints: form.accesspoints.map((item: AccessPoint) => ({
        ...item,
        accesspoint_id: item.accesspoint_id ?? -1,
      })),
      direct_endpoints: form.direct_endpoints,
      links,
      is_direct: props.isCreate ? type.value === 'direct' : isDirect.value,
    };

    if (!props.isCreate) (params as TopoNetworkUnitUpdateReq).bk_networkunit_id = props.workUnitId;
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
const handleBeforeClose = () => new Promise((resolve, reject) => {
  InfoBox({
    title: '确认关闭?',
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
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
  form.bk_networkunit_name = curWorkUnit.bk_networkunit_name;

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
    ]);

    if (!props.isCreate) {
      getWorkUnit();
    }
  }
  if (!isCurrentShow) {
    formRef.value.clearValidate();
    initForm();
  }
});
</script>
