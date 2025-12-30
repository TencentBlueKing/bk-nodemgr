<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    :title="title"
    width="800"
    :before-close="handleBeforeClose"
  >
    <div class="p-[24px] h-full overflow-auto">
      <Form ref="formRef" :model="formData" :rules="rules">
        <Form.FormItem :label="'配置名称'" property="configpolicy_name" required>
          <Input v-model="formData.configpolicy_name"></Input>
        </Form.FormItem>
        <Form.FormItem :label="'业务'" property="biz_id" required>
          <Select
            v-model="formData.biz_id"
            auto-focus
            filterable
            multiple
            placeholder="选择业务"
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
          :label="'是否启用'"
          property="enabled"
          required
          v-if="isEdit"
        >
          <Switcher v-model="formData.enabled" theme="primary"></Switcher>
        </Form.FormItem>
        <Form.FormItem :label="'备注'" property="biz_id">
          <Input type="textarea" v-model="formData.remark"></Input>
        </Form.FormItem>
        <Form.FormItem :label="'作用范围'" property="scopes">
          <div
            v-for="(item, index) in formData.scopes"
            :key="index"
            class="bg-[#F0F1F5] p-[16px] flex items-center gap-[12px]"
          >
            <div class="flex-1 flex flex-col gap-[8px]">
              <AreaSelector
                class="w-[568px]"
                :multiple="false"
                :no-limit="true"
                @change="(id, data) => handleSingleChange(item, id, data)" />
              <Select
                v-model="item.bk_networkunit_id"
                prefix="管控单元"
                :disabled="item.bk_networkarea_id === '-1'"
                auto-focus
                filterable
              >
                <Select.Option label="不限" value="-1"></Select.Option>
                <Select.Group>
                  <Select.Option
                    v-for="option in filterNetworkUnitList(
                      item.bk_networkarea_id
                    )"
                    :key="option.bk_networkarea_id"
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
                prefix="操作系统"
                auto-focus
                filterable
              >
                <Select.Option label="不限" value="-1"></Select.Option>
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
              <Select v-model="item.cpu_arch" prefix="架构" auto-focus filterable>
                <Select.Option label="不限" value="-1"></Select.Option>
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
            <div class="w-[20px] cursor-pointer" @click="handleDelete(index)">
              <i class="nodeman-icon nc-delete-3" v-show="index !== 0"></i>
            </div>
          </div>
          <Button theme="primary" text @click="handleAdd">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">添加范围</span>
          </Button>
        </Form.FormItem>
        <Form.FormItem :label="'配置'" property="configs">
          <div class="flex items-center text-[#979BA5]">
            <i class="nodeman-icon nc-tips"></i>
            <span class="ml-[9px] text-[12px]"
            >如需修改默认配置，需打开开关后修改</span
            >
          </div>
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
            >编辑器内容有改动，保存该配置版本将会由</span
            >
            <Tag theme="warning">{{
              `V${configData.version}`
            }}</Tag>
            <span class="mx-[3px]">升级为</span>
            <Tag theme="success">{{
              `V${configData.version + 1}`
            }}</Tag>
          </div>
        </Form.FormItem>
      </Form>
    </div>

    <template #footer>
      <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleSubmit">
        {{ isEdit ? "保存" : "提交" }}
      </Button>
      <Button class="w-[88px]" @click="handleBeforeClose">取消</Button>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Select, Sideslider, Switcher, Tag } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';

import type { NetworkArea, NetworkUnit } from '@/@types/topo';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { TopoService } from '@/api/modules/topo';
// 导入config-template组件
import ConfigTemplate from '@/components/config-template.vue';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

// Sideslider显示状态
const isShow = defineModel('isShow', { type: Boolean });

// 定义props和emits
const props = defineProps<{
  configpolicyType: string;
  isEdit: boolean;
  configData?: any;
}>();
const emit = defineEmits(['save']);
const mainStore = useMainStore();
const userStore = useUserStore();

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
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

const title = computed(() => {
  const action = props.isEdit ? '编辑' : '新建';
  const role = props.configpolicyType === 'config_policy_agent' ? 'Agent' : 'Proxy';
  return `${action} ${role} 配置`;
});
const businessList = computed(() => mainStore.businessList);

// 初始化数据函数
const formData = reactive({
  configpolicy_name: '',
  configpolicy_type: props.configpolicyType,
  biz_id: ['不限'] as string[] | number[],
  remark: '',
  scopes: [
    {
      bk_networkarea_id: '-1',
      bk_networkunit_id: '-1',
      os_type: '-1',
      cpu_arch: '-1',
    },
  ],
  configs: [] as ConfigPolicyConfigBlock[],
  operator: '',
  enabled: false,
});
const initData = () => {
  formData.configpolicy_name = '';
  formData.configpolicy_type = props.configpolicyType;
  formData.remark = '';
  formData.biz_id = ['不限'];
  formData.scopes = [
    {
      bk_networkarea_id: '-1',
      bk_networkunit_id: '-1',
      os_type: '-1',
      cpu_arch: '-1',
    },
  ];
  formData.configs = [];
  formData.operator = '';
  formData.enabled = false;
  configpolicyId.value = undefined;
};
const formRef = ref();
const configpolicyId = ref();
const rules = {
  configpolicy_name: [
    { required: true, message: '配置名称不能为空', trigger: 'blur' },
    {
      message: '配置名称长度在1-50个字符之间',
      trigger: 'blur',
      validator: (val: string) => val.length <= 50 && val.length >= 1,
    },
  ],
  biz_id: [{ required: true, message: '业务不能为空', trigger: 'change' }],
};
const handleAdd = () => {
  formData.scopes.push({
    bk_networkarea_id: '-1',
    bk_networkunit_id: '-1',
    os_type: '-1',
    cpu_arch: '-1',
  });
};
const handleDelete = (index: number) => {
  if (index === 0) return;
  formData.scopes = formData.scopes.filter((_: any, ind: number) => ind !== index);
};
const handleChangeBiz = (val: any) => {
  const filter = val.filter((item: any) => item !== '不限');
  if (filter.length > 0) {
    formData.biz_id = filter;
  } else {
    formData.biz_id = ['不限'];
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
  const biz_id = formData.biz_id.includes('不限') ? [] : formData.biz_id;
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
        : ['不限'];
      formData.scopes = props.configData.scopes.map((item: ConfigPolicyScope) => ({
        ...item,
        bk_networkarea_id: String(item.bk_networkarea_id),
        bk_networkunit_id: String(item.bk_networkunit_id),
        os_type: item.os_type === '' ? '-1' : item.os_type,
        cpu_arch: item.cpu_arch === '' ? '-1' : item.cpu_arch,
      }));
    }
  }
}, { immediate: true });
</script>
