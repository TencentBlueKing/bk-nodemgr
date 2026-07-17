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
          v-if="isEditMode"
        >
          <Switcher v-model="formData.enabled" theme="primary" :disabled="isViewMode"></Switcher>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.remark')" property="remark">
          <Input type="textarea" v-model="formData.remark" show-word-limit :maxlength="100" :disabled="isViewMode"></Input>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.form.selectPlugin')" property="plugin_name" required>
          <Select
            v-model="formData.plugin_name"
            :disabled="isViewMode"
            filterable
            clearable
          >
            <Select.Option
              v-for="item in pluginList"
              :key="item.name"
              :name="item.name"
              :id="item.name"
            ></Select.Option>
          </Select>
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
                :bk_networkarea_id="item.bk_networkarea_id"
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
          <div v-if="!isViewMode" class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer" @click="handleAdd">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ $t('agentStrategy.form.addScope') }}</span>
          </div>
        </Form.FormItem>
        <Form.FormItem :label="t('agentStrategy.preview.selectIP')" property="IP">
          <div v-if="!isViewMode" class="bg-[#F0F5FF] border-dashed border-2 border-[#A3C5FD] text-[#3A84FF] h-[30px] flex items-center justify-center cursor-pointer mb-[10px]" @click="handleAddIP">
            <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
            <span class="text-[14px]">{{ t('agentStrategy.form.addHost') }}</span>
          </div>
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
            <Table
              ref="ipTableRef"
              :data="tableHosts"
              :max-height="400"
              :pagination="ipPagination"
              show-overflow-tooltip
              @page-value-change="handleIpPageChange"
              @page-limit-change="handleIpPageLimitChange"
            >
              <TableColumn field="bk_host_innerip" :title="$t('common.ipv4')" :min-width="150"></TableColumn>
              <TableColumn field="bk_host_innerip_v6" :title="$t('common.ipv6')" :min-width="150"></TableColumn>
              <TableColumn field="bk_host_name" :title="$t('common.hostname')" :min-width="150"></TableColumn>
              <TableColumn field="bk_networkarea_name" :title="$t('common.cloudArea')" :width="150"></TableColumn>
              <TableColumn field="os_type" :title="$t('common.osType')" :width="120"></TableColumn>
            </Table>
          </div>
        </Form.FormItem>
        <Form.FormItem property="configs">
          <div class="mt-[15px]">
            <config-template
              :visible="true"
              configpolicy-type="config_policy_plugin"
              :is-edit="!isViewMode"
              :configs="formData.configs"
              @update-config="updateConfig"
            ></config-template>
            <div
              class="text-[#E71818] text-[12px] flex items-center"
              v-if="isEditMode && configData"
            >
              <i class="nodeman-icon nc-remind-fill text-[14px]"></i>
              <span class="mr-[3px] ml-[9px]">{{ t('agentStrategy.form.versionTip') }}</span>
              <Tag theme="warning">{{ `V${configData.version}` }}</Tag>
              <span class="mx-[3px]">{{ t('agentStrategy.form.upgradeTo') }}</span>
              <Tag theme="success">{{ `V${configData.version + 1}` }}</Tag>
            </div>
          </div>
        </Form.FormItem>
      </Form>
    </div>

    <template #footer>
      <template v-if="isViewMode">
        <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleRequestEdit">{{ t('agentStrategy.form.edit') }}</Button>
        <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
      </template>
      <template v-else>
        <Button theme="primary" class="mr-[8px] w-[88px]" @click="handleSubmit">{{ isEditMode ? t('agentStrategy.form.save') : t('agentStrategy.form.submit') }}</Button>
        <Button class="w-[88px]" @click="handleBeforeClose">{{ t('agentStrategy.form.cancel') }}</Button>
      </template>
    </template>

    <IpSelector
      mode="dialog"
      :show-dialog="isShowIpSelector"
      :value="ipSelectorValue"
      @change="handleIpSelectorChange"
      @panel-change="(panel: string) => IpSelectorService.setActiveIpsPanel(panel)"
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
import { PluginAPIService } from '@/api/modules/plugin';
import { TopoService } from '@/api/modules/topo';
import { PACKAGE_GENERATION } from '@/common/const';
import { Table, TableColumn } from '@blueking/table';
import IpSelector from '@/components/IpSelector';
import ConfigTemplate from '@/components/config-template.vue';
import { setPolicyType, setStrategyBizId } from '@/services/ip-selector';
import * as IpSelectorService from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

interface IScope {
  bk_networkarea_id: string | number;
  bk_networkunit_id: string | number;
  os_type: string;
  cpu_arch: string;
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

const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps<{
  mode: 'create' | 'edit' | 'view';
  configData?: any;
}>();

const isViewMode = computed(() => props.mode === 'view');
const isEditMode = computed(() => props.mode === 'edit');
const { t } = useI18n();
const emit = defineEmits(['save', 'request-edit']);
const mainStore = useMainStore();
const userStore = useUserStore();

const title = computed(() => {
  const role = t('agentStrategy.form.pluginRole');
  if (props.mode === 'view') return t('agentStrategy.form.viewTitle', { role });
  const action = props.mode === 'edit' ? t('agentStrategy.form.edit') : t('agentStrategy.form.create');
  return t('agentStrategy.form.configTitle', { action, role });
});
const businessList = computed(() => mainStore.businessList);

const formData = reactive({
  configpolicy_name: '',
  configpolicy_type: 'config_policy_plugin',
  bk_biz_id: mainStore.strategyBizId || 0,
  remark: '',
  plugin_name: '',
  scopes: [] as IScope[],
  selectedHosts: [] as ISelectedHost[],
  configs: [] as ConfigPolicyConfigBlock[],
  operator: '',
  enabled: false,
});

const initData = () => {
  formData.configpolicy_name = '';
  formData.remark = '';
  formData.plugin_name = '';
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
    { message: t('agentStrategy.validate.configNameLength'), trigger: 'blur', validator: (val: string) => val.length <= 50 && val.length >= 1 },
  ],
  bk_biz_id: [{ required: true, message: t('agentStrategy.validate.businessRequired'), trigger: 'change' }],
  plugin_name: [{ required: true, message: t('agentStrategy.validate.pluginRequired'), trigger: 'change' }],
};

const originData = ref(cloneDeep(formData));
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  if (isEqual(formData, originData.value)) { resolve(true); isShow.value = false; return; }
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => { resolve(true); isShow.value = false; },
    onCancel: () => reject(),
  });
});

const handleAdd = () => {
  formData.scopes.push({ bk_networkarea_id: '-1', bk_networkunit_id: '-1', os_type: '-1', cpu_arch: '-1' });
};

// IP selector
const isShowIpSelector = ref(false);
const ipSelectorValue = ref({ hostList: [], nodeList: [], dynamicGroupList: [], serviceTemplateList: [], setTemplateList: [] });
const ipPagination = reactive({ count: 0, limit: 10, current: 1, remote: true as const });
const tableHosts = ref<ISelectedHost[]>([]);
const allSelectedHostIds = ref<number[]>([]);

const mapCachedToRow = (item: any): ISelectedHost => ({
  bk_host_id: item.bk_host_id ?? item.host_id ?? 0,
  bk_host_innerip: item.ip || '',
  bk_host_innerip_v6: item.ipv6 || '',
  bk_host_name: item.host_name || '',
  bk_networkarea_name: item.cloud_area?.name || '',
  os_type: item.os_type || '',
  cpu_arch: item.cpu_arch || '',
  bk_networkarea_id: item.cloud_id ?? item.bk_cloud_id ?? 0,
});

const mapRawToRow = (host: any): ISelectedHost => {
  const mapped = IpSelectorService.mapHostItem(host);
  IpSelectorService.cacheHostItem(mapped);
  return mapCachedToRow(mapped);
};

const fetchHostData = async () => {
  const ids = allSelectedHostIds.value;
  if (!ids.length) { tableHosts.value = []; ipPagination.count = 0; formData.selectedHosts = []; return; }
  const { cached, missIds } = IpSelectorService.getCachedHosts(ids);
  let allRows = cached.map(mapCachedToRow);
  ipPagination.count = ids.length;
  if (missIds.length > 0) {
    try {
      const CHUNK_SIZE = 500;
      const fetchedItems: any[] = [];
      for (let i = 0; i < missIds.length; i += CHUNK_SIZE) {
        const chunkIds = missIds.slice(i, i + CHUNK_SIZE);
        const res = await TopoService.HostList({
          page: { limit: CHUNK_SIZE, offset: 0 },
          only_count: false,
          exact_include_conditions: { bk_host_id: chunkIds, bk_biz_id: [formData.bk_biz_id] },
        });
        fetchedItems.push(...(res.items || []).map(mapRawToRow));
      }
      const rowMap = new Map<number, ISelectedHost>();
      allRows.forEach(row => rowMap.set(row.bk_host_id, row));
      fetchedItems.forEach(row => rowMap.set(row.bk_host_id, row));
      allRows = ids.map(id => rowMap.get(id)).filter(Boolean) as ISelectedHost[];
    } catch { /* fallback to cache */ }
  }
  const offset = (ipPagination.current - 1) * ipPagination.limit;
  tableHosts.value = allRows.slice(offset, offset + ipPagination.limit);
  ipPagination.count = allRows.length;
  formData.selectedHosts = tableHosts.value;
};

const handleAddIP = () => { setPolicyType('config_policy_plugin'); setStrategyBizId(mainStore.strategyBizId); isShowIpSelector.value = true; };
const handleEditHosts = () => { setPolicyType('config_policy_plugin'); setStrategyBizId(mainStore.strategyBizId); isShowIpSelector.value = true; };

const handleIpSelectorChange = (value: any) => {
  ipSelectorValue.value = value;
  ipPagination.current = 1;
  const ids: number[] = [];
  if (value.hostList?.length) value.hostList.forEach((host: any) => { const id = host.hostId ?? host.host_id; if (id) ids.push(Number(id)); });
  allSelectedHostIds.value = ids;
  fetchHostData();
};

const handleDeleteAllHosts = () => InfoBox({
  title: t('agentStrategy.form.confirmDelete'),
  subTitle: t('agentStrategy.form.confirmDeleteAllHosts'),
  onConfirm: () => {
    ipSelectorValue.value = { hostList: [], nodeList: [], dynamicGroupList: [], serviceTemplateList: [], setTemplateList: [] };
    allSelectedHostIds.value = []; tableHosts.value = []; formData.selectedHosts = []; ipPagination.count = 0; ipPagination.current = 1;
  },
});

const handleIpPageChange = (page: number) => { ipPagination.current = page; fetchHostData(); };
const handleIpPageLimitChange = (limit: number) => { ipPagination.limit = limit; ipPagination.current = 1; fetchHostData(); };
const handleDelete = (index: number) => { formData.scopes = formData.scopes.filter((_: any, ind: number) => ind !== index); };
const updateConfig = (configs: any[]) => { formData.configs = configs; };

const handleRequestEdit = () => emit('request-edit');

const handleSubmit = async () => {
  const isValid = await formRef.value?.validate().catch(() => false);
  if (!isValid) return;
  let res;
  const scopes = formData.scopes.map((item: any) => ({
    bk_networkarea_id: Number(item.bk_networkarea_id),
    bk_networkunit_id: Number(item.bk_networkunit_id),
    os_type: item.os_type === '-1' ? '' : item.os_type,
    cpu_arch: item.cpu_arch === '-1' ? '' : item.cpu_arch,
  }));
  const targetHostIds = allSelectedHostIds.value.length ? allSelectedHostIds.value : formData.selectedHosts.map((h: any) => h.bk_host_id);
  if (isEditMode.value) {
    res = await ConfigPolicyAPIService.ConfigPolicyUpdate({
      ...formData, configpolicy_id: configpolicyId.value, scopes, target_host_ids: targetHostIds, operator: userStore.user?.username, bk_biz_id: formData.bk_biz_id,
    }).catch(() => false);
  } else {
    res = await ConfigPolicyAPIService.ConfigPolicyCreate({
      ...formData, scopes, target_host_ids: targetHostIds, operator: userStore.user?.username, bk_biz_id: formData.bk_biz_id,
    }).catch(() => false);
  }
  isShow.value = false;
  if (res) emit('save');
};

const handleSingleChange = (item: any, id: string, rows: any[]) => { item.bk_networkarea_id = id; item.bk_networkunit_id = ''; item.bk_networkarea_name = rows[0].bk_networkarea_name; };
const handleUnitChange = (item: any, id: number | string, row: any) => { item.bk_networkunit_id = id; if (row) item.bk_networkunit_name = row.bk_networkunit_name; };

const osTypeList = ref<{ value: string; label: string }[]>([]);
const cpuArchList = ref<{ value: string; label: string }[]>([]);

// Plugin list
const pluginList = ref<{ name: string }[]>([]);
const fetchPluginList = async () => {
  try {
    const res = await PluginAPIService.ListPlugins({
      page: { limit: 500, offset: 0 },
      exact_include_conditions: { bk_biz_id: [formData.bk_biz_id] },
    } as any).catch(() => ({ items: [] }));
    pluginList.value = res?.items || [];
  } catch { pluginList.value = []; }
};

watch(() => isShow.value, async () => {
  if (isShow.value) {
    fetchPluginList();
    if (props.mode === 'create') {
      initData();
    } else if (props.mode === 'edit' || props.mode === 'view') {
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
        allSelectedHostIds.value = props.configData.target_host_ids || [];
        ipPagination.current = 1;
        await fetchHostData();
      }
    }
    originData.value = cloneDeep(formData);
  } else {
    allSelectedHostIds.value = []; tableHosts.value = [];
  }
}, { immediate: true });
</script>

<style lang="postcss" scoped>
.form-scope {
  &:hover {
    .delete { display: inline-block; }
  }
}
</style>
