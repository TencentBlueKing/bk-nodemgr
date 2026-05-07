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
          <Input v-model="formData.configpolicy_name" :disabled="isViewMode"></Input>
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
          <Switcher v-model="formData.enabled" theme="primary" :disabled="isViewMode"></Switcher>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.remark')" property="remark">
          <Input type="textarea" v-model="formData.remark" show-word-limit :maxlength="100" :disabled="isViewMode"></Input>
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
                :disabled="isViewMode"
                @change="(id, data) => handleSingleChange(item, id, data)" />
              <UnitSelector
                v-model="item.bk_networkunit_id"
                :no-limit="true"
                :disabled="item.bk_networkarea_id === '-1' || isViewMode"
                :filter-by-area-id="item.bk_networkarea_id"
                select-class="w-full"
                @change="(id, row) => handleUnitChange(item, id, row)"
              />
              <Select
                v-model="item.os_type"
                :prefix="t('agentStrategy.form.os')"
                auto-focus
                filterable
                :disabled="isViewMode"
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
              <Select v-model="item.cpu_arch" :prefix="t('agentStrategy.form.arch')" auto-focus filterable :disabled="isViewMode">
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
              v-if="!isViewMode"
              class="delete absolute right-[-10px] top-[-15px] w-[20px] cursor-pointer hidden"
              @click="handleDelete(index)">
              <i class="nodeman-icon nc-minus text-[#EA3636]"></i>
            </div>
          </div>
          <!--eslint-disable-next-line max-len -->
          <div v-if="!isViewMode" class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer" @click="handleAdd">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ $t('agentStrategy.form.addScope') }}</span>
          </div>
        </Form.FormItem>
        <Form.FormItem :label="t('作用IP')" property="IP">
          <!-- 添加IP按钮 -->
          <!--eslint-disable-next-line max-len -->
          <div v-if="!isViewMode" class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer mb-[10px]" @click="handleAddIP">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ t('agentStrategy.form.addHost') }}</span>
          </div>
          
          <!-- IP列表：顶部统计和操作 -->
          <div v-if="allSelectedHostIds.length > 0" class="ip-list-wrapper">
            <div class="flex items-center mb-[10px]">
              <span class="text-[14px] text-[#63656E]">
                {{ t('agentStrategy.form.totalAdded') }} 
                <span class="text-[#3A84FF] font-medium">{{ allSelectedHostIds.length }}</span>
                {{ t('agentStrategy.form.unit') }}
              </span>
              <i v-if="!isViewMode" class="nodeman-icon nc-edit-line text-[16px] text-[#979BA5] cursor-pointer ml-[12px]" @click="handleEditHosts"></i>
              <i v-if="!isViewMode" class="nodeman-icon nc-delete text-[16px] text-[#979BA5] cursor-pointer ml-[8px]" @click="handleDeleteAllHosts"></i>
            </div>
            
            <!-- IP列表表格（无操作列） -->
            <Table
              ref="ipTableRef"
              :data="tableHosts"
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
              :is-edit="!isViewMode"
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
      <!-- 查看模式：显示编辑按钮 -->
      <template v-if="isViewMode">
        <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleRequestEdit">
          {{ t('agentStrategy.form.edit') }}
        </Button>
        <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
      </template>
      <!-- 创建/编辑模式：显示保存按钮 -->
      <template v-else>
        <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleSubmit">
          {{ isEditMode ? t('agentStrategy.form.save') : t('agentStrategy.form.submit') }}
        </Button>
        <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
      </template>
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
import { TopoService } from '@/api/modules/topo';
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
  mode: 'create' | 'edit' | 'view';
  configData?: any;
}>();

const isViewMode = computed(() => props.mode === 'view');
const isEditMode = computed(() => props.mode === 'edit');
const { t } = useI18n();
const emit = defineEmits(['save']);
const mainStore = useMainStore();
const userStore = useUserStore();

const title = computed(() => {
  if (props.mode === 'view') {
    const role = props.configpolicyType === 'config_policy_agent' ? 'Agent' : 'Proxy';
    return t('agentStrategy.form.viewTitle', { role });
  }
  const action = props.mode === 'edit' ? t('agentStrategy.form.edit') : t('agentStrategy.form.create');
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

// IP列表分页（后端分页，参考 agent list 页实现）
const ipPagination = reactive({ count: 0, limit: 10, current: 1, remote: true as const });

// 当前页主机数据（由接口填充）
const tableHosts = ref<ISelectedHost[]>([]);

// 所有选中的主机 ID（用于提交）
const allSelectedHostIds = ref<number[]>([]);

// 根据 allSelectedHostIds 从后端分页获取主机详情（参考 agent list 页后端分页实现）
const fetchHostData = async () => {
  const ids = allSelectedHostIds.value;
  if (!ids.length) {
    tableHosts.value = [];
    ipPagination.count = 0;
    formData.selectedHosts = [];
    return;
  }
  const offset = (ipPagination.current - 1) * ipPagination.limit;
  const res = await TopoService.HostList({
    page: { limit: ipPagination.limit, offset },
    only_count: false,
    exact_include_conditions: {
      bk_host_id: ids,
      bk_biz_id: [formData.bk_biz_id],
      node_role: props.configpolicyType === 'config_policy_agent' ? ['agent', 'blank'] : ['proxy'],
    },
  }).catch(() => ({ total: 0, items: [] as any[] }));
  tableHosts.value = res.items.map((host: any) => ({
    bk_host_id: host.bk_host_id,
    bk_host_innerip: host.info?.bk_host_innerip_list?.[0] || '',
    bk_host_innerip_v6: host.info?.bk_host_innerip_v6_list?.[0] || '',
    bk_host_name: host.info?.bk_host_name || '',
    bk_networkarea_name: host.info?.bk_networkarea_name || '',
    os_type: host.info?.os_type || '',
    cpu_arch: host.info?.cpu_arch || '',
    bk_networkarea_id: host.info?.bk_networkarea_id || 0,
  }));
  ipPagination.count = res.total;
  // 更新 formData.selectedHosts 为当前页数据（用于模板显示）
  formData.selectedHosts = tableHosts.value;
};

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

// IP选择器确认事件 — 提取 hostId 后调后端分页接口
const handleIpSelectorChange = (value: any) => {
  ipSelectorValue.value = value;
  ipPagination.current = 1;
  // 从 hostList 提取所有 hostId
  const ids: number[] = [];
  if (value.hostList && value.hostList.length > 0) {
    value.hostList.forEach((host: any) => {
      const id = host.hostId ?? host.host_id;
      if (id) ids.push(Number(id));
    });
  }
  allSelectedHostIds.value = ids;
  fetchHostData();
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
      allSelectedHostIds.value = [];
      tableHosts.value = [];
      formData.selectedHosts = [];
      ipPagination.count = 0;
      ipPagination.current = 1;
    },
  });
};

const handleIpPageChange = (page: number) => {
  ipPagination.current = page;
  fetchHostData();
};

const handleIpPageLimitChange = (limit: number) => {
  ipPagination.limit = limit;
  ipPagination.current = 1;
  fetchHostData();
};

const handleDelete = (index: number) => {
  formData.scopes = formData.scopes.filter((_: any, ind: number) => ind !== index);
};
const updateConfig = (configs: any[]) => {
  formData.configs = configs;
};

// 查看模式下点击编辑按钮，通知父组件切换到编辑模式
const handleRequestEdit = () => {
  emit('request-edit');
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
  // 提交时使用 allSelectedHostIds 作为 target_host_ids
  const targetHostIds = allSelectedHostIds.value.length
    ? allSelectedHostIds.value
    : formData.selectedHosts.map((h: any) => h.bk_host_id);
  if (isEditMode.value) {
    res = await ConfigPolicyAPIService.ConfigPolicyUpdate({
      ...formData,
      configpolicy_id: configpolicyId.value,
      scopes,
      target_host_ids: targetHostIds,
      operator: userStore.user?.username,
      bk_biz_id,
    }).catch(() => false);
  } else {
    res = await ConfigPolicyAPIService.ConfigPolicyCreate({
      ...formData,
      scopes,
      target_host_ids: targetHostIds,
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



watch(() => isShow.value, async () => {
  if (isShow.value) {
    getPlatform();
    if (props.mode === 'create') {
      initData();
    } else if (props.mode === 'edit' || props.mode === 'view') {
      // 编辑/查看模式，使用props中的配置数据
      if (props.configData) {
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
        // 将 target_host_ids 存入 allSelectedHostIds，然后调后端分页接口
        allSelectedHostIds.value = props.configData.target_host_ids || [];
        ipPagination.current = 1;
        await fetchHostData();
      }
    }
    originData.value = cloneDeep(formData);
  } else {
    // 关闭时清空
    allSelectedHostIds.value = [];
    tableHosts.value = [];
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
