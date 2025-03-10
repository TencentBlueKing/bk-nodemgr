<template>
  <div class="agent">
    <!-- agnet操作及搜索 -->
    <section class="agent-operate mb15">
      <div class="agent-operate-left">
        <Dropdown theme="light" trigger="manual" placement="bottom" :is-show="dropdownShow">
          <Button class="installButton" theme="primary" @click="handleInstall">{{ $t('platform.nodeMan.installAgent') }}</Button>
          <template #content>
            <Dropdown.DropdownMenu extCls="dropDown-menu">
              <Dropdown.DropdownItem v-for="item in installChannel" :key="item.id" @click="triggerHandler(item.id)">
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <Dropdown theme="light" trigger="click">
          <Button :disabled="!selection.length">
            <span>{{ $t('platform.nodeMan.batchOperate') }}</span>
            <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem v-for="item in operate" :key="item.id" @click="handleOperate(item.id, selection, true)">
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <copy-ip-dropdown :type="'agent'" :disabled="!selection.length" :data="data.list"></copy-ip-dropdown>
      </div>
      <div class="agent-operate-right">
        <Cascader is-remote clearable v-model="topo" :list="topoBizFilterList" :remote-method="topoRemotehandler" ref="topoSelect" :placeholder="$t('业务拓扑')"/>
        <SearchSelect ref="searchSelect" :data="searchSelectData" v-model="searchSelectValue" :uniqueSelect="true"
          :placeholder="$t('platform.nodeMan.agentSearchPlaceholder')" @update:modelValue="handleSearchSelectChange"></SearchSelect>
      </div>
    </section>
    <Table :data="data.list" :empty-text="'暂无数据'" :pagination="pagination" :column-config="{ resizable: true }" show-overflow-tooltip
      :max-height="maxHeight" @checkbox-change="handleSelectChange" @checkbox-all="handleSelectAllChange">
      <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
      <TableColumn field="inner_ip" :title="t('platform.nodeMan.inner_ip')" width="120" fixed="left"></TableColumn>
      <TableColumn field="inner_ipv6" :title="t('platform.nodeMan.inner_ipv6')" width="120"></TableColumn>
      <TableColumn field="bk_agent_id" :title="t('platform.nodeMan.agentId')" width="295"></TableColumn>
      <TableColumn field="bk_cloud_name" :title="t('platform.nodeMan.bk_cloud_name')"></TableColumn>
      <TableColumn field="bk_cloud_unit" :title="t('platform.nodeMan.bk_cloud_unit')"></TableColumn>
      <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')"></TableColumn>
      <TableColumn field="version" :title="t('platform.nodeMan.agent_version')"></TableColumn>
      <TableColumn field="status" :title="t('platform.nodeMan.status')" width="100">
        <template #default="{ row }">
          <div class="col-status" v-if="row.status_display">
            <span :class="`nodeman-icon nc-${row.status.toLowerCase()} status-icon`"></span>
            <span>{{ row.status_display }}</span>
          </div>
          <div class="col-status" v-else>
            <span class="nodeman-icon nc-unknown status-icon"></span>
            <span>{{ row.status_display }}</span>
          </div>
        </template>
      </TableColumn>
      <TableColumn :title="t('platform.nodeMan.operate')" width="120" fixed="right">
        <template #default="{ row }">
          <Button theme="primary" text ext-cls="reinstall" @click="handleOperate('reinstall', [row])">
            {{ row.status === 'not_installed' ? $t('安装') : $t('重装') }}
          </Button>

          <Dropdown theme="light" trigger="click">
            <Button class="ml15" text>
              <span class="nodeman-icon nc-more"></span>
            </Button>
            <template #content>
              <Dropdown.DropdownMenu>
                <Dropdown.DropdownItem v-for="item in operate" :key="item.id" v-show="getOperateShow(row, item)"
                  @click="handleOperate(item.id, [row])">
                  {{ item.name }}
                </Dropdown.DropdownItem>
              </Dropdown.DropdownMenu>
            </template>
          </Dropdown>
        </template>
      </TableColumn>
    </Table>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { useI18n } from 'vue-i18n';
import { toLower } from 'lodash';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Button, Dropdown, Cascader, SearchSelect } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { WorkflowService } from '@/api/modules/workflow';

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const data = reactive({
  list: [
    {
      "status": "RUNNING",
      "version": "v2.1.6-beta.31",
      "bk_biz_id": 3,
      "bk_host_id": 41,
      "login_ip": "10.0.15.104",
      "ap_id": 2,
      "inner_ipv6": "",
      "updated_at": "2025-02-27 14:39:42+0800",
      "dept_name": "",
      "bk_addressing": "static",
      "data_ip": "",
      "outer_ip": "",
      "is_manual": false,
      "bk_cloud_id": 0,
      "bk_agent_id": "0200000000525400c6046216975306648243",
      "install_channel_id": null,
      "extra_data": {
        "bt_speed_limit": null,
        "enable_compression": false,
        "peer_exchange_switch_for_agent": 0
      },
      "outer_ipv6": "",
      "created_at": "2023-10-17 16:15:09+0800",
      "inner_ip": "10.0.15.104",
      "bk_host_name": "10_0_15_104",
      "os_type": "WINDOWS",
      "status_display": "正常",
      "bk_cloud_name": "直连区域",
      "install_channel_name": null,
      "bk_biz_name": "锦华测试专用",
      "identity_info": {
        "account": "Administrator",
        "auth_type": "PASSWORD",
        "port": 445,
        "re_certification": true
      },
      "job_result": {},
      "topology": [
        "锦华测试专用 / 日志检索 / windows"
      ],
      "operate_permission": true
    },
    {
      "status": "RUNNING",
      "version": "v2.1.6-beta.31",
      "bk_biz_id": 3,
      "bk_host_id": 8,
      "login_ip": "",
      "ap_id": 2,
      "inner_ipv6": "",
      "updated_at": "2025-02-27 14:35:17+0800",
      "dept_name": "",
      "bk_addressing": "static",
      "data_ip": "",
      "outer_ip": "",
      "is_manual": false,
      "bk_cloud_id": 0,
      "bk_agent_id": "0200000000525400425f831693884436672l",
      "install_channel_id": null,
      "extra_data": {
        "bt_speed_limit": null,
        "enable_compression": false,
        "peer_exchange_switch_for_agent": 0
      },
      "outer_ipv6": "",
      "created_at": "2024-01-11 11:25:34+0800",
      "inner_ip": "10.0.15.236",
      "bk_host_name": "VM-15-236-centos",
      "os_type": "LINUX",
      "status_display": "正常",
      "bk_cloud_name": "直连区域",
      "install_channel_name": null,
      "bk_biz_name": "锦华测试专用",
      "identity_info": {
        "account": "root",
        "auth_type": "PASSWORD",
        "port": 22,
        "re_certification": true
      },
      "job_result": {},
      "topology": [
        "锦华测试专用 / 日志检索 / Linux"
      ],
      "operate_permission": true
    },
    {
      "status": "TERMINATED",
      "version": "v2.1.6-beta.31",
      "bk_biz_id": 2,
      "bk_host_id": 2430,
      "login_ip": "10.0.220.132",
      "ap_id": 2,
      "inner_ipv6": "",
      "updated_at": "2024-12-25 14:43:45+0800",
      "dept_name": "",
      "bk_addressing": "static",
      "data_ip": "",
      "outer_ip": "",
      "is_manual": false,
      "bk_cloud_id": 0,
      "bk_agent_id": "0200000000525400f5e45d17349451545957",
      "install_channel_id": null,
      "extra_data": {
        "bt_speed_limit": null,
        "enable_compression": false,
        "peer_exchange_switch_for_agent": 0
      },
      "outer_ipv6": "",
      "created_at": "2024-12-25 14:43:45+0800",
      "inner_ip": "10.0.220.132",
      "bk_host_name": "VM-220-132-centos",
      "os_type": "LINUX",
      "status_display": "异常",
      "bk_cloud_name": "直连区域",
      "install_channel_name": null,
      "bk_biz_name": "蓝鲸",
      "identity_info": {
        "account": "root",
        "auth_type": "PASSWORD",
        "port": 22,
        "re_certification": true
      },
      "job_result": {},
      "topology": [
        "蓝鲸 / 空闲机池 / 空闲机"
      ],
      "operate_permission": true
    },
  ],
  total: 5,
});
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
// 分页
const {
  pagination
} = usePage(ref(data.list));
// 跨页全选
const isSelectedAllPages = ref(false);
const loading = ref(false);
// 拓扑级联选择器的选值
const topo = ref([]);
const topoBizFilterList = ref([]);
const topoRemotehandler = () => {}
// 搜索
const searchSelectValue = ref([]);
const searchSelectData = ref([]);
const handleSearchSelectChange = ({id, name, values}: {id: number, name: string, values: []}) => {}
// 批量操作
const operate = [
  {
    id: 'upgrade',
    name: t('升级'),
    disabled: false,
    show: true,
  },
  {
    id: 'uninstall',
    name: t('卸载'),
    disabled: false,
    show: true,
  },
  {
    id: 'reload',
    name: t('重载配置'),
    disabled: false,
    show: true,
  },
  {
    id: 'reboot',
    name: t('重启'),
    disabled: false,
    show: true,
  },
  {
    id: 'log',
    name: t('最新执行日志'),
    disabled: false,
    show: true,
    single: true,
  },
];
// 安装方式
const installChannel = [
  {
    id: 'setup',
    name: t('普通安装'),
  },
  {
    id: 'import',
    name: t('Excel 导入安装'),
  }
];
const dropdownShow = ref(false);
const handleInstall = () => {
  if (selection.value.length) {
    dropdownShow.value = false;
    triggerHandler('reinstall');
  } else {
    dropdownShow.value = !dropdownShow.value;
  }
}
const triggerHandler = (type: string) => {
  switch (type) {
    // 批量重启 批量重装 批量重载配置 批量卸载 批量升级
    case 'reboot':
    case 'reinstall':
    case 'reload':
    case 'uninstall':
    case 'upgrade':
    case 'remove':
      handleOperate(type, selection.value, true);
      break;
      // 普通安装
    case 'setup':
      console.log("🚀 ~ triggerHandler ~ setup:")
      router.push({ name: 'agentSetup' })
      break;
      // Excel 导入
    case 'import':
      console.log("🚀 ~ triggerHandler ~ setup:")
      router.push({ name: 'agentImport' });
      break;
  }
}
/**
 * 当前操作项是否显示
 */
const getOperateShow = (row: IAgentHost, config: IOperateItem) => {
  if (config.id === 'log' && (!row.job_result || !row.job_result.job_id)) {
    return false;
  }
  return config.show;
}
// 操作
const handleOperate = (type: string, data: any[], batch = false) => {
  if (!batch && ['terminated', 'not_installed'].includes(data[0].status) && !(['log', 'reinstall', 'remove'].includes(type))) {
    return;
  }

  let jobType = '';

  switch (type) {
    // 重启
    case 'reboot':
      handleOperatetHost(data, batch, 'RESTART_AGENT');
      break;
    // 移除
    case 'remove':
      handleOperatetHost(data, batch, 'REMOVE_AGENT');
      break;
    // 重装
    case 'reinstall':
      jobType = 'REINSTALL_AGENT';
      break;
    // 重载 只取其中一部分数据
    case 'reload':
      jobType = 'RELOAD_AGENT';
      break;
    // 卸载
    case 'uninstall':
      jobType = 'UNINSTALL_AGENT';
      break;
    // 升级
    case 'upgrade':
      handleOperatetHost(data, batch, 'UPGRADE_AGENT');
      break;
    // 日志详情
    case 'log':
      // handleGotoLog(data[0]);
      break;
  }
  if (!jobType) return;

  router.push({
    name: 'agentEdit',
    params: {
      tableData: data.map(({ identity_info = {}, ...item }) => ({
        ...item,
        ...identity_info,
        install_channel_id: (item.install_channel_id === -1 || !item.install_channel_id)
          ? 'default'
          : item.install_channel_id,
        port: identity_info.port,
      })),
      type: jobType,
      // true：跨页全选（tableData表示标记删除的数据） false：非跨页全选（tableData表示编辑的数据）
      isSelectedAllPages: String(batch && isSelectedAllPages.value),
      condition: [],
    },
  });
};
// 表格勾选
const selection = computed(() => data.list.filter((item: any) => item.checked));
const handleSelectChange = ({ checked, row }: {checked: boolean, row: any}) => {
  row.checked = checked;
}

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean}) => {
  data.list.forEach((item: any) => item.checked = checked);
}
/**
   * Agent操作
   * @param {String} type 操作类型
   * @param {Array} data agent数据
   * @param {Boolean} batch 是否是批量操作
   *
   * 重装都需要经过编辑页面 *****
   * Linux升级走job不需要编辑，windows升级需要编辑不走job， 混合走编辑 *****
   * Linux、window卸载都不需要经过编辑页面 *****
   */

/**
 * @param {Array} data
 */
const handleOperatetHost = async (data: [], batch: boolean, operateType: string) => {
  const titleObj = {
    firstIp: batch ? selection.value[0].inner_ip : data[0].inner_ip,
    num: batch ? selection.value.length : data.length,
  };
  const operateJob = async (data: IAgentHost[]) => {
    // loading.value = true;
    // const params = getOperateHostCondition(data, operateType);
    // const result = await AgentStore.operateJob(params);
    // loading.value = false;
    // if (result.job_id) {
    //   router.push({ name: 'taskDetail', params: { taskId: result.job_id, routerBackName: 'taskList' } });
    // }
  };
  let type = '';
  switch (operateType) {
    // 重启
    case 'RESTART_AGENT':
      type = t('重启');
      break;
    // 升级
    case 'UPGRADE_AGENT':
      type = t('升级');
      break;
    case 'REMOVE_AGENT':
      type = t('移除');
      break;
  }

  InfoBox({
    title: batch
      ? `请确认是否批量${type}`
      : `请确认是否${type}`,
    subTitle: batch
      ? t('批量确认操作提示', {
        ip: titleObj.firstIp,
        num: titleObj.num,
        type,
        suffix: operateType === 'UPGRADE_AGENT' ? t('到最新版本') : ''
      })
      : `${type} ${titleObj.firstIp} 的Agent${operateType === 'UPGRADE_AGENT' ? t('到最新版本') : ''}`,
    extCls: 'wrap-title',
    onConfirm: () => {
      if (operateType === 'REMOVE_AGENT') {
        // handleRemoveHost(data);
      } else {
        operateJob(data);
      }
    },
  });
}
const getAgentList = async () => {
  loading.value = true;
  const res = WorkflowService.WorkflowList();
}
onMounted(async () => {
  // await getAgentList();
});
</script>
<style scoped lang="postcss">
.agent {
  padding: 24px 24px 10px 24px;
  .agent-operate {
    display: flex;
    justify-content: space-between;

    .agent-operate-left {
      display: flex;
      gap: 8px;

      .installButton {
        width: 130px;
      }
    }
    .agent-operate-right {
      display: flex;
      gap: 8px;
    }
  }
}

.ml15 {
  margin-left: 15px;
}

.mb15 {
  margin-bottom: 15px;
}
.dropDown-menu {
  width: 130px;
  .bk-dropdown-item {
    font-size: 14px;
  }
}
.bk-cascader-wrapper {
  width: 250px;

  .bk-cascader-wrapper:not(:last-of-type) {
    margin-bottom: 20px;
  }
}
.bk-search-select {
  width: 480px;
}
.col-status {
  display: flex;
  align-items: center;
}
.status-icon::before {
  content: '';
  display: inline-block;
  margin-right: 8px;
  width: 13px;
  height: 13px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
}
.nc-running {
  &::before {
    background: #3fc06d;
    border-color: #e5f6ea;
  }
}
.nc-terminated {
  &::before {
    border-color: #ffe6e6;
    background: #ea3636;
  }
}
.nc-unknown {
  &::before {
    border-color: #f0f1f5;
    background: #b2b5bd;
  }
}

</style>