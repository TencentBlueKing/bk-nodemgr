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
        <Form.FormItem :label="t('agentStrategy.form.business')" property="bk_biz_id" required>
          <Select
            v-model="formData.bk_biz_id"
            disabled
            :placeholder="t('agentStrategy.form.selectBusiness')"
          >
            <Select.Option
              v-for="item in businessList"
              :key="item.bk_biz_id"
              :name="`[${item.bk_biz_id}] ${item.bk_biz_name}`"
              :id="item.bk_biz_id"
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
        <Form.FormItem :label="t('agentStrategy.form.remark')" property="remark">
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
              <UnitSelector
                v-model="item.bk_networkunit_id"
                :no-limit="true"
                :disabled="item.bk_networkarea_id === '-1'"
                :filter-by-area-id="item.bk_networkarea_id"
                select-class="w-full"
                @change="(id, row) => handleUnitChange(item, id, row)"
              />
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
        <Form.FormItem :label="t('作用IP')" property="IP">
          <!-- 添加IP按钮 -->
          <!--eslint-disable-next-line max-len -->
          <div class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer mb-[10px]" @click="handleAddIP">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ t('agentStrategy.form.addHost') }}</span>
          </div>
          
          <!-- IP列表：顶部统计和操作 -->
          <div v-if="formData.selectedHosts.length > 0" class="ip-list-wrapper">
            <div class="flex items-center mb-[10px]">
              <span class="text-[14px] text-[#63656E]">
                {{ t('agentStrategy.form.totalAdded') }} 
                <span class="text-[#3A84FF] font-medium">{{ formData.selectedHosts.length }}</span>
                {{ t('agentStrategy.form.unit') }}
              </span>
              <i class="nodeman-icon nc-edit-line text-[16px] text-[#979BA5] cursor-pointer ml-[12px]" @click="handleEditHosts"></i>
              <i class="nodeman-icon nc-delete text-[16px] text-[#979BA5] cursor-pointer ml-[8px]" @click="handleDeleteAllHosts"></i>
            </div>
            
            <!-- IP列表表格（无操作列） -->
            <Table
              ref="ipTableRef"
              :data="paginatedHosts"
              :max-height="400"
              :pagination="ipPagination"
              show-overflow-tooltip
              @page-value-change="handleIpPageChange"
              @page-limit-change="handleIpPageLimitChange"
            >
              <TableColumn
                field="bk_host_innerip"
                :title="$t('IP')"
                :min-width="150"
              ></TableColumn>
              <TableColumn
                field="bk_host_innerip_v6"
                :title="$t('IPv6')"
                :min-width="150"
              ></TableColumn>
              <TableColumn
                field="bk_host_name"
                :title="$t('主机名称')"
                :min-width="150"
              ></TableColumn>
              <TableColumn
                field="bk_networkarea_name"
                :title="$t('云区域')"
                :width="150"
              ></TableColumn>
              <TableColumn
                field="os_type"
                :title="$t('系统')"
                :width="120"
              ></TableColumn>
            </Table>
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
    
    <!-- 蓝鲸IP选择器 (Vue3版本) -->
    <IpSelector
      mode="dialog"
      :show-dialog="isShowIpSelector"
      :value="ipSelectorValue"
      @change="handleIpSelectorChange"
      :keep-host-field-output="true"
      @close-dialog="isShowIpSelector = false"
    />
  </Sideslider>
</template>

<script lang="ts" setup>
import { Button, Form, InfoBox, Input, Select, Sideslider, Switcher, Tag } from 'bkui-vue';
import { cloneDeep, isEqual } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { PACKAGE_GENERATION } from '@/common/const';
import { Table, TableColumn } from '@blueking/table';
// 导入config-template组件
import ConfigTemplate from '@/components/config-template.vue';
import IpSelector from '@/components/IpSelector';
import { setPolicyType, setStrategyBizId } from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

interface IScope {
  bk_networkarea_id: string | number,
  bk_networkunit_id: string | number,
  os_type: string,
  cpu_arch: string,
}

interface ISelectedHost {
  bk_host_id: number;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  bk_host_name: string;
  bk_networkarea_name: string;
  os_type: string;
  cpu_arch: string;
  bk_networkarea_id: number;
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
  bk_biz_id: mainStore.strategyBizId || 0,
  remark: '',
  scopes: [] as IScope[],
  selectedHosts: [] as ISelectedHost[],
  configs: [] as ConfigPolicyConfigBlock[],
  operator: '',
  enabled: false,
});
const initData = () => {
  formData.configpolicy_name = '';
  formData.configpolicy_type = props.configpolicyType;
  formData.remark = '';
  formData.bk_biz_id = mainStore.strategyBizId || 0;
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
  bk_biz_id: [{ required: true, message: t('agentStrategy.validate.businessRequired'), trigger: 'change' }],
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

// IP选择器相关
const isShowIpSelector = ref(false);

// IP选择器的值（蓝鲸IP选择器数据结构）
const ipSelectorValue = ref({
  hostList: [],
  nodeList: [],
  dynamicGroupList: [],
  serviceTemplateList: [],
  setTemplateList: [],
});

// IP列表分页
const ipPagination = reactive({
  current: 1,
  limit: 10,
  count: 0,
});

// 从IP选择器结果提取主机列表（库返回字段：hostId, ip, ipv6, cloudArea{ id, name }, meta）
const extractedHosts = computed(() => {
  const hosts: ISelectedHost[] = [];

  // 从hostList提取主机
  if (ipSelectorValue.value.hostList && ipSelectorValue.value.hostList.length > 0) {
    ipSelectorValue.value.hostList.forEach((host: any) => {
      hosts.push({
        bk_host_id: host.hostId ?? host.host_id,
        bk_host_innerip: host.ip || '',
        bk_host_innerip_v6: host.ipv6 || '',
        bk_host_name: host.hostName || host.host_name || '',
        bk_networkarea_name: host.cloudArea?.name || '',
        os_type: host.osType || host.os_type || '',
        cpu_arch: host.cpuArch || '',
        bk_networkarea_id: host.cloudArea?.id ?? host.cloud_id ?? 0,
      });
    });
  }

  return hosts;
});

// 分页后的主机列表
const paginatedHosts = computed(() => {
  const start = (ipPagination.current - 1) * ipPagination.limit;
  const end = start + ipPagination.limit;
  return extractedHosts.value.slice(start, end);
});

// 监听提取的主机列表变化
watch(extractedHosts, (newHosts) => {
  formData.selectedHosts = newHosts;
  ipPagination.count = newHosts.length;
});

const handleAddIP = () => {
  setPolicyType(props.configpolicyType);
  setStrategyBizId(mainStore.strategyBizId);
  isShowIpSelector.value = true;
};

const handleEditHosts = () => {
  setPolicyType(props.configpolicyType);
  setStrategyBizId(mainStore.strategyBizId);
  isShowIpSelector.value = true;
};

// IP选择器确认事件 — keep-host-field-output=true 已保留完整字段，无需再调 host/list 补全
const handleIpSelectorChange = (value: any) => {
  ipPagination.current = 1;
  ipSelectorValue.value = value;
};

const handleDeleteAllHosts = () => {
  InfoBox({
    title: t('agentStrategy.form.confirmDelete'),
    subTitle: t('agentStrategy.form.confirmDeleteAllHosts'),
    onConfirm: () => {
      ipSelectorValue.value = {
        hostList: [],
        nodeList: [],
        dynamicGroupList: [],
        serviceTemplateList: [],
        setTemplateList: [],
      };
      formData.selectedHosts = [];
      ipPagination.count = 0;
      ipPagination.current = 1;
    },
  });
};

const handleIpPageChange = (page: number) => {
  ipPagination.current = page;
};

const handleIpPageLimitChange = (limit: number) => {
  ipPagination.limit = limit;
  ipPagination.current = 1;
};

const handleDelete = (index: number) => {
  formData.scopes = formData.scopes.filter((_: any, ind: number) => ind !== index);
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
  const bk_biz_id = formData.bk_biz_id;
  if (props.isEdit) {
    res = await ConfigPolicyAPIService.ConfigPolicyUpdate({
      ...formData,
      configpolicy_id: configpolicyId.value,
      scopes,
      target_host_ids: extractedHosts.value.map((h) => h.bk_host_id),
      operator: userStore.user?.username,
      bk_biz_id,
    }).catch(() => false);
  } else {
    res = await ConfigPolicyAPIService.ConfigPolicyCreate({
      ...formData,
      scopes,
      target_host_ids: extractedHosts.value.map((h) => h.bk_host_id),
      operator: userStore.user?.username,
      bk_biz_id,
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

const handleUnitChange = (item: any, id: number | string, row: any) => {
  item.bk_networkunit_id = id;
  if (row) {
    item.bk_networkunit_name = row.bk_networkunit_name;
  }
};
// 操作系统和架构
const osTypeList = ref<{ value: string; label: string }[]>();
const cpuArchList = ref<{ value: string; label: string }[]>();
const getPlatform = async () => {
  const res = await ConfigPolicyAPIService.ConfigPolicyListPlatform({
    configpolicy_type: props.configpolicyType,
    generation: PACKAGE_GENERATION,
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
    getPlatform();
    if (!props.isEdit) {
      initData();
    } else if (props.isEdit && props.configData) {
      // 编辑模式，使用props中的配置数据
      configpolicyId.value = props.configData.configpolicy_id;
      Object.assign(formData, props.configData);
      formData.bk_biz_id = props.configData.bk_biz_id;
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
