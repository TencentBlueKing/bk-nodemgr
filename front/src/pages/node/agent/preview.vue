<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="1600"
    :title="$t('platform.nodeMan.preview.title')"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #default>
      <div class="py-[24px] px-[40px]">
        <div
          class="flex min-h-[56px] bg-[#FFF4E2] border border-[#FFF4E2] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
        >
          <i class="nodeman-icon nc-tips pt-[2px] text-[#FF9C01]"></i>
          <div class="flex-1 text-[12px] text-[#4D4F56]">
            <p>{{ $t("platform.nodeMan.preview.tipTitle") }}</p>
            <p>{{ $t("platform.nodeMan.preview.firstTip") }}</p>
            <p>{{ $t("platform.nodeMan.preview.secondTip") }}</p>
          </div>
        </div>
        <div class="flex justify-between mt-[16px]">
          <div class="w-[50%] flex gap-[8px]">
            <Input
              type="search"
              v-model="searchValue"
              :placeholder="$t('platform.nodeMan.preview.placeholder')"
            />
            <copy-ip-dropdown
              :type="'agent'"
              :disabled="!selection.length"
              :data="tableData"
            ></copy-ip-dropdown>
          </div>
          <div class="flex gap-[8px]">
            <Button @click="handleBatchInstall">{{
              $t("platform.nodeMan.preview.button.batchInstall")
            }}</Button>
            <Button @click="handleBatchRemove">{{
              $t("platform.nodeMan.preview.button.batchRemove")
            }}</Button>
          </div>
        </div>
        <Tab class="mt-[16px]" v-model:active="active" type="card" :key="tabKey">
          <!-- <template #setting>
            <div class="leading-[50px]"><i class="mr-[16px] nodeman-icon nc-setting"></i></div>
          </template> -->
          <Tab.TabPanel
            v-for="item in tabs"
            :key="item.name"
            :label="item.label"
            :name="item.name"
          >
            <template #label>
              <div class="flex gap-[5px] items-center">
                <close
                  v-if="item.icon === 'wrong'"
                  width="14px"
                  height="14px"
                  :fill="item.iconColor"
                />
                <i
                  v-else
                  :class="`nodeman-icon nc-${item.icon} text-[14px]`"
                  :style="{ color: item.iconColor }"
                ></i>
                <span class="text-[14px] text-[#313238]">{{ item.label }}</span>
                <div
                  :class="`rounded-[8px] w-[23px] h-[16px] border text-[12px] leading-[16px] text-center
                    bg-[${
                    item.name === active
                      ? '#E1ECFF'
                      : '#DCDEE5'
                  }]`"
                >
                  {{ item.count }}
                </div>
              </div>
            </template>
            <template #panel>
              <Table
                :data="tableData"
                :empty-text="$t('table.empty')"
                :column-config="{ resizable: true }"
                show-overflow-tooltip
                :max-height="462"
                :show-settings="isShowSetting"
                :settings="settings"
                @setting-change="handleSettingChange"
                @checkbox-change="handleSelectChange"
                @checkbox-all="handleSelectAllChange"
              >
                <TableColumn
                  type="checkbox"
                  width="80"
                  fixed="left"
                ></TableColumn>
                <TableColumn
                  field="bk_host_innerip"
                  :title="t('platform.nodeMan.inner_ip')"
                  min-width="150"
                  fixed="left"
                ></TableColumn>
                <TableColumn
                  field="bk_host_innerip_v6"
                  :title="t('platform.nodeMan.inner_ipv6')"
                  min-width="150"
                ></TableColumn>
                <TableColumn
                  field="os_type"
                  :title="t('platform.nodeMan.os_type')"
                  min-width="100"
                ></TableColumn>
                <TableColumn
                  field="bk_host_name"
                  :title="t('platform.nodeMan.bk_host_name')"
                  min-width="100"
                ></TableColumn>
                <TableColumn
                  field="bk_networkarea_name"
                  :title="t('platform.nodeMan.bk_cloud_name')"
                  min-width="100"
                ></TableColumn>
                <TableColumn
                  field="elig_status"
                  :title="t('platform.nodeMan.status')"
                  min-width="400"
                >
                  <template #default="{ row }">
                    <div class="flex items-center gap-[4px]">
                      <close
                        v-if="statusMap[row.elig_status]?.icon === 'wrong'"
                        width="14px"
                        height="14px"
                        :fill="statusMap[row.elig_status]?.iconColor"
                      />
                      <i
                        v-else
                        :class="`nodeman-icon nc-${statusMap[row.elig_status]?.icon} text-[14px]`"
                        :style="{ color: statusMap[row.elig_status]?.iconColor }"
                      ></i>
                      <p>{{ statusMap[row.elig_status]?.text }}</p>
                      <PopConfirm
                        width="780"
                        :title="statusMap[row.elig_status]?.text"
                        trigger="click"
                        @confirm="ensure(row)"
                      >
                        <Button
                          v-show="['conflict_ip', 'duplicate_dynamic_ip'].includes(row.elig_status)"
                          theme="primary"
                          text
                          @click="getAgentList(row)"
                        >{{
                          t("platform.nodeMan.preview.button.deel")
                        }}</Button>
                        <template #content>
                          <div class="mt-[6px] mb-[8px] text-[12px]">
                            {{ t("platform.nodeMan.preview.popConfirm.tip") }}
                          </div>
                          <Table
                            class="mb-[24px]"
                            :data="deelTabelData"
                            :empty-text="t('table.empty')"
                            show-overflow-tooltip
                            :column-config="{ resizable: true }"
                            :max-height="462"
                            :show-settings="isShowSetting"
                            :settings="settings"
                            @setting-change="handleSettingChange"
                          >
                            <template #prepend>
                              <div
                                class="bg-[#F0F5FF] h-[32px] flex items-center pl-[16px]"
                              >
                                <Radio v-model="radioValue" label="cmdb">{{
                                  t(
                                    "platform.nodeMan.preview.popConfirm.prepend"
                                  )
                                }}</Radio>
                              </div>
                            </template>
                            <TableColumn
                              field="bk_host_innerip"
                              :title="t('platform.nodeMan.inner_ip')"
                              min-width="200"
                            >
                              <template #default="{ row }">
                                <Radio v-model="radioValue" :label="row.bk_host_innerip">
                                  <span>{{
                                    t(
                                      "platform.nodeMan.preview.popConfirm.install"
                                    )
                                  }}</span>
                                  <span>{{ row.bk_host_innerip }}</span>
                                </Radio>
                              </template>
                            </TableColumn>
                            <TableColumn
                              field="bk_host_innerip_v6"
                              :title="t('platform.nodeMan.inner_ipv6')"
                              min-width="150"
                            ></TableColumn>
                            <TableColumn
                              field="os_type"
                              :title="t('platform.nodeMan.os_type')"
                              min-width="100"
                            ></TableColumn>
                            <TableColumn
                              field="bk_host_name"
                              :title="t('主机名')"
                              min-width="100"
                            ></TableColumn>
                            <TableColumn
                              field="bk_networkarea_name"
                              :title="t('platform.nodeMan.bk_cloud_name')"
                              min-width="150"
                            ></TableColumn>
                          </Table>
                        </template>
                      </PopConfirm>
                    </div>
                  </template>
                </TableColumn>
                <TableColumn
                  field="action"
                  :title="t('platform.nodeMan.operate')"
                  min-width="100"
                  fixed="right"
                >
                  <template #default="{ row }">
                    <Button theme="primary" text @click="handleRemove(row)">{{
                      t("platform.nodeMan.preview.button.remove")
                    }}</Button>
                  </template>
                </TableColumn>
              </Table>
            </template>
          </Tab.TabPanel>
        </Tab>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-start gap-[8px]">
        <Button
          theme="primary"
          @click="handleSetup"
          :disabled="disabledDataNum > 0"
        >{{
          t("platform.nodeMan.preview.button.performInstallation")
        }}</Button
        >
        <Button @click="handleBeforeClose">{{ t("action.cancel") }}</Button>
      </div>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import {
  Button,
  Input,
  Message,
  PopConfirm,
  Radio,
  Sideslider,
  Tab,
  Tag } from 'bkui-vue';
import { Close } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { AgentInstallInfo } from '@/@types/node_agent.d';
import { NodeAgentService } from '@/api/modules/node_agent';
import { TopoService } from '@/api/modules/topo';
import useTableSetting from '@/composables/use-table-setting';
import { useNodeManageStore } from '@/stores/node-manage';

const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps({
  data: {
    type: Object,
    default: () => {},
  },
});

const { t } = useI18n();
const nodeManageStore = useNodeManageStore();
const router = useRouter();
const selection = ref([]);
const searchValue = ref('');
const originData = ref<AgentInstallInfo[]>([]);
const tableData = ref<AgentInstallInfo[]>([]);
const tabKey = ref(Date.now());
const disabledDataNum = ref(0);
const tabs = computed(() => {
  const countResult = originData.value.reduce((acc: any, item: any) => {
    acc.all += 1;

    if (['conflict_ip', 'duplicate_dynamic_ip'].includes(item.elig_status)) {
      acc.confirm += 1;
    }
    if (['exist_proxy', 'exist_agent'].includes(item.elig_status)) {
      acc.error += 1;
    }
    if (item.elig_status === 'clean_install') {
      acc.cleanInstallation += 1;
    }
    if (item.elig_status === 'normal_install') {
      acc.normalInstallation += 1;
    }

    return acc;
  }, {
    all: 0,
    confirm: 0,
    error: 0,
    cleanInstallation: 0,
    normalInstallation: 0,
  });
  disabledDataNum.value = countResult.confirm + countResult.error;
  return [
    {
      label: t('platform.nodeMan.preview.label.all'),
      name: 'all',
      count: countResult.all,
    },
    {
      label: t('platform.nodeMan.preview.label.pendingConfirmation'),
      name: 'confirm',
      count: countResult.confirm,
      icon: 'danger-fill',
      iconColor: '#FF9C01',
      includeSatatus: ['conflict_ip', 'duplicate_dynamic_ip'],
    },
    {
      label: t('platform.nodeMan.preview.label.error'),
      name: 'error',
      count: countResult.error,
      icon: 'wrong',
      iconColor: '#EA3636',
      includeSatatus: ['exist_proxy', 'exist_agent'],
    },
    {
      label: t('platform.nodeMan.preview.label.cleanInstallation'),
      name: 'CMDB',
      count: countResult.cleanInstallation,
      icon: 'check-circle-fill',
      iconColor: '#1CAB88',
      includeSatatus: ['clean_install'],
    },
    {
      label: t('platform.nodeMan.preview.label.normalInstallation'),
      name: 'setup',
      count: countResult.normalInstallation,
      icon: 'check-circle-fill',
      iconColor: '#1CAB88',
      includeSatatus: ['normal_install'],
    },
  ];
});

const active = ref('all');
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_agent_id',
    'bk_host_name',
    'bk_networkarea_name',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'os_type',
    'node_version',
    'elig_status',
    'action',
  ],
  disabled: ['action'],
});
const statusMap = {
  clean_install: {
    text: '可执行：全新安装并导入 CMDB',
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  normal_install: {
    text: '可执行：正常安装',
    icon: 'check-circle-fill',
    iconColor: '#1CAB88',
  },
  exist_proxy: {
    text: '该主机已安装 proxy，若继续安装该 Proxy 将会被覆盖',
    icon: 'wrong',
    iconColor: '#EA3636',
  },
  exist_agent: {
    text: '该 IP 主机在当前管控区域已安装 Agent，无法重复安装',
    icon: 'wrong',
    iconColor: '#EA3636',
  },
  conflict_ip: {
    text: '当前业务下可能已存在您希望安装的相同 IP 主机',
    icon: 'danger-fill',
    iconColor: '#FF9C01',
  },
  duplicate_dynamic_ip: {
    text: '动态寻址下已存在相同 IP 的安装记录',
    icon: 'danger-fill',
    iconColor: '#FF9C01',
  },
};
const handleSelectChange = () => {};
const handleSelectAllChange = () => {};

const radioValue = ref('cmdb');

const handleBeforeClose = () => {
  isShow.value = false;
};
const ensure = (row: AgentInstallInfo) => {
  if (row.elig_status === 'conflict_ip' || row.elig_status === 'duplicate_dynamic_ip') {
    row.elig_status = radioValue.value === 'cmdb' ? 'clean_install' : 'normal_install';
  }
  tabKey.value = Date.now();
  Message({
    theme: 'success',
    message: radioValue.value === 'cmdb' ? '该 Agent 已被手动确认为“全新安装并导入 CMDB' : '该 Agent 已被手动确认为“正常安装',
  });
};
const handleRemove = (row: AgentInstallInfo) => {
  originData.value = originData.value.filter((item: any) => item.bk_host_innerip !== row.bk_host_innerip);
  tabKey.value = Date.now();
  Message({
    theme: 'success',
    message: '全部“错误”Agent 已被移除',
  });
};
const handleBatchInstall = () => {
  tableData.value.forEach((item: any) => {
    if (item.elig_status === 'conflict_ip' || item.elig_status === 'duplicate_dynamic_ip') {
      item.elig_status = 'clean_install';
    }
  });
  tabKey.value = Date.now();
  Message({
    theme: 'success',
    message: '全部“待确认”Agent 已被手动确认为“全新安装并导入 CMDB"',
  });
};
const handleBatchRemove = () => {
  originData.value = originData.value.filter((item: any) => !['exist_proxy', 'exist_agent'].includes(item.elig_status));
  tabKey.value = Date.now();
};
const handleSetup = async () => {
  const res = await NodeAgentService.NodeAgentInstall({
    info: tableData.value,
    target_version: props.data.target_version,
    disable_default_target_version: props.data.disable_default_target_version,
  }).catch(() => ({
    workflow_id: '',
  }));
  if (!res) return;
  if (res.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: res.workflow_id },
    });
  }
};
// 定义排序优先级
const priority = {
  conflict_ip: 0,
  duplicate_dynamic_ip: 0,
  exist_proxy: 1,
  exist_agent: 1,
  clean_install: 2,
  normal_install: 3,
};

// 排序函数
function sortByEligStatus(arr: any[]) {
  // 复制原数组避免修改原数组
  return [...arr].sort((a, b) => {
    // 获取当前元素的优先级，默认最低
    const priorityA = priority[a.elig_status] ?? 6;
    const priorityB = priority[b.elig_status] ?? 6;
    // 按优先级升序排列（数值越小优先级越高）
    return priorityA - priorityB;
  });
}
const installCheck = async () => {
  const res = await NodeAgentService.NodeAgentInstallCheck({
    host: originData.value.map((item: any) => ({
      bk_biz_id: item.bk_biz_id,
      bk_host_innerip: item.bk_host_innerip,
      bk_networkunit_id: item.bk_networkunit_id,
    })),
  }).catch(() => ({
    total_count: 0,
    install_eligs: [],
  }));
  if (res) {
    originData.value = sortByEligStatus(originData.value.map((item: any) => {
      const target = res.install_eligs.find((elig: any) => elig.inner_ip === item.bk_host_innerip);
      return {
        ...item,
        duplicate_host_ids: target?.duplicate_host_ids || [],
        elig_status: target?.elig_status || 'clean_install',
      };
    }));
  }
  tableData.value = cloneDeep(originData.value);
};

const deelTabelData = ref<any[]>([]);
const getAgentList = async (row: any) => {
  const res = await TopoService.HostList({
    exact_include_conditions: {
      bk_host_id: [...row.duplicate_host_ids],
    },
  }).catch(err => ({
    total: 0,
    items: [],
  }));
  deelTabelData.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
  }));
};
watch(() => tableData.value, () => {
  // 每次tableData变化时，重新生成一个key，以强制刷新tab
  tabKey.value = Date.now();
}, { immediate: true });
watch([active, originData], () => {
  tableData.value = originData.value.filter((item: any) => {
    if (active.value === 'all') {
      return true;
    }
    return tabs.value.find((tab: any) => tab.name === active.value)?.includeSatatus?.includes(item.elig_status);
  });
}, { immediate: true });
watch(
  () => isShow,
  async () => {
    if (isShow.value && props.data) {
      originData.value = props.data.info.map(item => ({
        ...item,
        login_port: Number(item.login_port),
        bk_biz_id: props.data.bk_biz_id,
        bk_networkunit_id: Number(props.data.bk_networkunit_id),
        bk_host_id: Number(item.bk_host_id),
        bk_networkarea_name: props.data.bk_networkarea_name,
      }));
      await installCheck();
    }
  },
  { immediate: true, deep: true },
);
</script>
<style lang="postcss" scoped>
:deep(.bk-tab-header) {
  background: #f0f1f5;
}
:deep(.bk-tab-content) {
  padding: 0;
}
</style>
