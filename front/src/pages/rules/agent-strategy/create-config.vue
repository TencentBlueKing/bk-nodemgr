<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    :title="title"
    width="800"
    :before-close="handleBeforeClose"
  >
    <div class="p-[24px] h-full overflow-auto">
      <Form ref="formRef" :model="formData" :rules="rules" form-type="vertical">
        <Form.FormItem :label="t('agentStrategy.form.configName')" property="configpolicy_name" required>
          <Input v-model="formData.configpolicy_name"></Input>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.business')" property="biz_id" required>
          <Select
            v-model="formData.biz_id"
            auto-focus
            filterable
            multiple
            :placeholder="t('agentStrategy.form.selectBusiness')"
            @change="handleChangeBiz"
          >
            <Select.Option
              v-for="item in businessList"
              :key="item.bk_biz_id"
              :name="item.bk_biz_name"
              :id="item.bk_biz_id"
            >
              <span v-show="item.bk_biz_id !== -1"
              >[{{ item.bk_biz_id }}] {{ item.bk_biz_name }}</span
              >
            </Select.Option>
          </Select>
        </Form.FormItem>
        <Form.FormItem
          :label="t('agentStrategy.form.enabled')"
          property="enabled"
          required
          v-if="isEdit"
        >
          <Switcher v-model="formData.enabled" theme="primary"></Switcher>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.remark')" property="biz_id">
          <Input type="textarea" v-model="formData.remark" show-word-limit :maxlength="100"></Input>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.scope')" property="scopes">
          <div
            v-for="(item, index) in formData.scopes"
            :key="index"
            class="bg-[#F0F1F5] p-[16px] flex items-center gap-[12px] mb-[10px] relative form-scope"
          >
            <div class="flex-1 flex flex-col gap-[8px]">
              <AreaSelector
                class="w-full"
                :multiple="false"
                :no-limit="true"
                @change="(id, data) => handleSingleChange(item, id, data)" />
              <Select
                v-model="item.bk_networkunit_id"
                :prefix="t('agentStrategy.form.workUnit')"
                :disabled="item.bk_networkarea_id === '-1'"
                auto-focus
                filterable
              >
                <Select.Option :label="t('agentStrategy.form.unlimited')" value="-1"></Select.Option>
                <Select.Group>
                  <Select.Option
                    v-for="option in filterNetworkUnitList(
                      item.bk_networkarea_id
                    )"
                    :key="option.bk_networkunit_id"
                    :id="String(option.bk_networkunit_id)"
                    :name="option.bk_networkunit_name"
                  >
                    [{{ option.bk_networkunit_id }}]
                    {{ option.bk_networkunit_name }}
                  </Select.Option>
                </Select.Group>
              </Select>
              <Select
                v-model="item.os_type"
                :prefix="t('agentStrategy.form.os')"
                auto-focus
                filterable
              >
                <Select.Option :label="t('agentStrategy.form.unlimited')" value="-1"></Select.Option>
                <Select.Group>
                  <Select.Option
                    v-for="option in osTypeList"
                    :key="option.value"
                    :id="option.value"
                    :name="option.label"
                  >
                  </Select.Option>
                </Select.Group>
              </Select>
              <Select v-model="item.cpu_arch" :prefix="t('agentStrategy.form.arch')" auto-focus filterable>
                <template #prefix>
                  <div class="w-[65px] text-center text-[#63656E] text-[12px] border-r border-r-[#c4c6cc]">
                    架<span class="ml-[24px]">构</span>
                  </div>
                </template>
                <Select.Option :label="t('agentStrategy.form.unlimited')" value="-1"></Select.Option>
                <Select.Group>
                  <Select.Option
                    v-for="option in cpuArchList"
                    :key="option.value"
                    :id="option.value"
                    :name="option.label"
                  >
                  </Select.Option>
                </Select.Group>
              </Select>
            </div>
            <div
              class="delete absolute right-[-10px] top-[-15px] w-[20px] cursor-pointer hidden"
              @click="handleDelete(index)">
              <i class="nodeman-icon nc-minus text-[#EA3636]"></i>
            </div>
          </div>
          <!--eslint-disable-next-line max-len -->
          <div class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer" @click="handleAdd">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ $t('agentStrategy.form.addScope') }}</span>
          </div>
        </Form.FormItem>
        <Form.FormItem property="configs">
          <div class="mt-[15px]">
            <config-template
              :visible="true"
              :configpolicy-type="configpolicyType"
              :is-edit="isEdit"
              :configs="formData.configs"
              @update-config="updateConfig"
            ></config-template>
            <div
              class="text-[#E71818] text-[12px] flex items-center"
              v-if="isEdit && configData"
            >
              <i class="nodeman-icon nc-remind-fill text-[14px]"></i>
              <span class="mr-[3px] ml-[9px]"
              >{{ t('agentStrategy.form.versionTip') }}</span
              >
              <Tag theme="warning">{{
                `V${configData.version}`
              }}</Tag>
              <span class="mx-[3px]">{{ t('agentStrategy.form.upgradeTo') }}</span>
              <Tag theme="success">{{
                `V${configData.version + 1}`
              }}</Tag>
            </div>
          </div>
        </Form.FormItem>
      </Form>
    </div>

    <template #footer>
      <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleSubmit">
        {{ isEdit ? t('agentStrategy.form.save') : t('agentStrategy.form.submit') }}
      </Button>
      <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Select, Sideslider, Switcher, Tag } from 'bkui-vue';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { cloneDeep, isEqual } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { NetworkArea, NetworkUnit } from '@/@types/topo';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { TopoService } from '@/api/modules/topo';
// 导入config-template组件
import ConfigTemplate from '@/components/config-template.vue';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

interface IScope {
  bk_networkarea_id: string | number,
  bk_networkunit_id: string | number,
  os_type: string,
  cpu_arch: string,
}
// Sideslider显示状态
const isShow = defineModel('isShow', { type: Boolean });

// 定义props和emits
const props = defineProps<{
  configpolicyType: string;
  isEdit: boolean;
  configData?: any;
}>();
const { t } = useI18n();
const emit = defineEmits(['save']);
const mainStore = useMainStore();
const userStore = useUserStore();

const title = computed(() => {
  const action = props.isEdit ? t('agentStrategy.form.edit') : t('agentStrategy.form.create');
  const role = props.configpolicyType === 'config_policy_agent' ? 'Agent' : 'Proxy';
  return t('agentStrategy.form.configTitle', { action, role });
});
const businessList = computed(() => mainStore.businessList);

// 初始化数据函数
const formData = reactive({
  configpolicy_name: '',
  configpolicy_type: props.configpolicyType,
  biz_id: [t('agentStrategy.form.unlimited')] as string[] | number[],
  remark: '',
  scopes: [] as IScope[],
  configs: [] as ConfigPolicyConfigBlock[],
  operator: '',
  enabled: false,
});
const initData = () => {
  formData.configpolicy_name = '';
  formData.configpolicy_type = props.configpolicyType;
  formData.remark = '';
  formData.biz_id = [t('agentStrategy.form.unlimited')];
  formData.scopes = [];
  formData.configs = [];
  formData.operator = '';
  formData.enabled = false;
  configpolicyId.value = undefined;
};
const formRef = ref();
const configpolicyId = ref();
const rules = {
  configpolicy_name: [
    { required: true, message: t('agentStrategy.validate.configNameRequired'), trigger: 'blur' },
    {
      message: t('agentStrategy.validate.configNameLength'),
      trigger: 'blur',
      validator: (val: string) => val.length <= 50 && val.length >= 1,
    },
  ],
  biz_id: [{ required: true, message: t('agentStrategy.validate.businessRequired'), trigger: 'change' }],
};

const originData = ref(cloneDeep(formData));
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  // 没有修改数据，直接关闭
  if (isEqual(formData, originData.value)) {
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
    onCancel: () => reject(),
  });
});

const handleAdd = () => {
  formData.scopes.push({
    bk_networkarea_id: '-1',
    bk_networkunit_id: '-1',
    os_type: '-1',
    cpu_arch: '-1',
  });
};
const handleDelete = (index: number) => {
  formData.scopes = formData.scopes.filter((_: any, ind: number) => ind !== index);
};
const handleChangeBiz = (val: any) => {
  const filter = val.filter((item: any) => item !== t('agentStrategy.form.unlimited'));
  if (filter.length > 0) {
    formData.biz_id = filter;
  } else {
    formData.biz_id = [t('agentStrategy.form.unlimited')];
  }
};
const updateConfig = (configs: any[]) => {
  formData.configs = configs;
};

const handleSubmit = async () => {
  // 表单验证
  const isValid = await formRef.value?.validate().catch(() => false);
  if (!isValid) return;

  let res;
  const scopes = formData.scopes.map((item: any) => ({
    bk_networkarea_id: Number(item.bk_networkarea_id),
    bk_networkunit_id: Number(item.bk_networkunit_id),
    os_type: item.os_type === '-1' ? '' : item.os_type,
    cpu_arch: item.cpu_arch === '-1' ? '' : item.cpu_arch,
  }));
  const biz_id = formData.biz_id.includes(t('agentStrategy.form.unlimited')) ? [] : formData.biz_id;
  if (props.isEdit) {
    res = await ConfigPolicyAPIService.ConfigPolicyUpdate({
      configpolicy_id: configpolicyId.value,
      ...formData,
      scopes,
      operator: userStore.user?.username,
      biz_id,
    }).catch(() => false);
  } else {
    res = await ConfigPolicyAPIService.ConfigPolicyCreate({
      ...formData,
      scopes,
      operator: userStore.user?.username,
      biz_id,
    }).catch(() => false);
  }
  isShow.value = false;
  if (res && props.configpolicyType) {
    // 提交成功后通知父组件
    emit('save');
  }
};

const handleSingleChange = (item: any, id: string, rows: any[]) => {
  // 单选通常用于表单赋值
  item.bk_networkarea_id = id;
  item.bk_networkunit_id = '';
  item.bk_networkarea_name = rows[0].bk_networkarea_name;
};
// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
// eslint-disable-next-line max-len
const filterNetworkUnitList = (id: number | string) => networkUnitList.value.filter((item: NetworkUnit) => item.bk_networkarea_id === Number(id) || item.bk_networkunit_id === -1);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({}).catch(() => ({
    total: 0,
    items: [],
  }));
  networkUnitList.value = res.items;
};
// 操作系统和架构
const osTypeList = ref<{ value: string; label: string }[]>();
const cpuArchList = ref<{ value: string; label: string }[]>();
const getPlatform = async () => {
  const res = await ConfigPolicyAPIService.ConfigPolicyListPlatform({
    configpolicy_type: props.configpolicyType,
    generation: 2,
  }).catch(() => ({
    os_type: [],
    cpu_arch: [],
  }));
  osTypeList.value = res.os_type.map(item => ({
    value: item,
    label: item,
  }));
  cpuArchList.value = res.cpu_arch.map(item => ({
    value: item,
    label: item,
  }));
};


watch(() => isShow.value, () => {
  if (isShow.value) {
    Promise.all([getNetworkUnitList(), getPlatform()]);
    if (!props.isEdit) {
      initData();
    } else if (props.isEdit && props.configData) {
      // 编辑模式，使用props中的配置数据
      configpolicyId.value = props.configData.configpolicy_id;
      Object.assign(formData, props.configData);
      formData.biz_id = props.configData.biz_id.length
        ? props.configData.biz_id
        : [t('agentStrategy.form.unlimited')];
      formData.scopes = props.configData.scopes.map((item: ConfigPolicyScope) => ({
        ...item,
        bk_networkarea_id: String(item.bk_networkarea_id),
        bk_networkunit_id: String(item.bk_networkunit_id),
        os_type: item.os_type === '' ? '-1' : item.os_type,
        cpu_arch: item.cpu_arch === '' ? '-1' : item.cpu_arch,
      }));
    }
    originData.value = cloneDeep(formData);
  }
}, { immediate: true });
</script>
<style lang="postcss" scoped>
.form-scope {
  &:hover {
    .delete {
      display: inline-block;
    }
  }
}
</style>
