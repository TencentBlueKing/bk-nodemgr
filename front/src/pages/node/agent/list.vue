<template>
  <div class="p-[24px]">
    <!-- agnet操作及搜索 -->
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[8px]">
        <Dropdown
          theme="light"
          trigger="manual"
          placement="bottom-start"
          :is-show="dropdownShow"
        >
          <Button class="w-[130px]" theme="primary" @click="handleInstall">{{
            $t("platform.nodeMan.installAgent")
          }}</Button>
          <template #content>
            <Dropdown.DropdownMenu extCls="dropDown-menu">
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in agentInstallType"
                :key="item.id"
                @click="triggerHandler('setup', item.id)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <Dropdown theme="light" trigger="click">
          <Button :disabled="!selection.length">
            <span>{{ $t("platform.nodeMan.batchOperate") }}</span>
            <i
              class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
            ></i>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem
                v-for="item in operate"
                :key="item.id"
                @click="handleOperate(item.id, selection, true)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <copy-ip-dropdown
          :type="'agent'"
          :disabled="!selection.length"
          :data="tableData"
        ></copy-ip-dropdown>
      </div>
      <div class="flex gap-[8px]">
        <Cascader
          class="w-[250px]"
          is-remote
          clearable
          v-model="topo"
          :list="topoBizFilterList"
          id-key="bk_biz_id"
          name-key="bk_biz_name"
          :remote-method="topoRemotehandler"
          ref="topoSelect"
          :placeholder="$t('业务拓扑')"
        />
        <SearchSelect
          class="w-[480px] z-99"
          ref="searchSelect"
          :data="searchSelectData"
          v-model="searchSelectValue"
          :uniqueSelect="true"
          :placeholder="$t('platform.nodeMan.agentSearchPlaceholder')"
          @update:modelValue="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <bk-loading
      title="数据加载中"
      :loading="loading"
      class="w-full overflow-auto"
    >
      <Table
        :data="tableData"
        :empty-text="'暂无数据'"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
        <TableColumn
          field="bk_host_innerip"
          :title="t('platform.nodeMan.inner_ip')"
          :min-width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip_v6"
          :title="t('platform.nodeMan.inner_ipv6')"
          :min-width="150"
        ></TableColumn>
        <TableColumn
          field="bk_agent_id"
          :title="t('platform.nodeMan.agentId')"
          :min-width="320"
        ></TableColumn>
        <TableColumn
          field="bk_networkarea_name"
          :title="t('platform.nodeMan.bk_cloud_name')"
          :filter="areaFilterOption"
          :min-width="120"
          show-overflow
        ></TableColumn>
        <TableColumn
          show-overflow
          field="bk_networkunit_name"
          :title="t('platform.nodeMan.bk_cloud_unit')"
          :filter="filterOptionSource.bk_networkunit_id"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          show-overflow
          field="os_type"
          :title="t('platform.nodeMan.os_type')"
          :filter="filterOptionSource.os_type"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ osMap[row.os_type] }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="node_version"
          :title="t('platform.nodeMan.agent_version')"
          :filter="filterOptionSource.node_version"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          field="node_status"
          :title="t('platform.nodeMan.status')"
          :min-width="150"
          :filter="filterOptionSource.node_status"
        >
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <span
                :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"
              ></span>
              <span>{{ row.node_status }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.node_status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="t('platform.nodeMan.operate')"
          :min-width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <Button
              theme="primary"
              text
              ext-cls="reinstall"
              @click="handleOperate('reinstall', [row])"
            >
              {{ $t("platform.nodeMan.agentStatus.button.reinstall") }}
            </Button>

            <Dropdown theme="light" trigger="click">
              <Button class="ml-[15px]" text>
                <span class="nodeman-icon nc-more"></span>
              </Button>
              <template #content>
                <Dropdown.DropdownMenu>
                  <Dropdown.DropdownItem
                    v-for="item in operate"
                    :key="item.id"
                    v-show="getOperateShow(row, item)"
                    @click="handleOperate(item.id, [row])"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                </Dropdown.DropdownMenu>
              </template>
            </Dropdown>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}
import { ref, reactive, computed, onMounted, shallowRef, watch } from "vue";
import { Table, TableColumn } from "@blueking/table";
import { useI18n } from "vue-i18n";
import { toLower } from "lodash";
import { useRoute, useRouter } from "vue-router";
import { InfoBox, Button, Dropdown, Cascader, SearchSelect } from "bkui-vue";
import { useMainStore } from "@/stores/main";
import { useNodeManageStore } from "@/stores/node-manage";
import { WorkflowService } from "@/api/modules/workflow";
import { TopoService } from "@/api/modules/topo";
import useTableSetting from "@/composables/use-table-setting";
import { TopoHostDistinctRespData } from "@/@types/topo";
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from "@/@types/topo.d";
import { capitalizeFirstLetter } from "@/common/util";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const tableData = ref<Host[]>([]);
const agentList = ref<Host[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
// 后端分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
// 跨页全选
const isSelectedAllPages = ref(false);
const loading = ref(false);
// 拓扑级联选择器的选值
const topo = ref([]);
const topoBizFilterList = computed(() => mainStore.businessList);
const topoRemotehandler = () => {};
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>(
  []
);
const handleSearchSelectChange = async (
  data: { id: string; name: string; values: { id: string; name: string }[] }[]
) => {
  // 给筛选器添加选中值
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id]) {
      filterOptionSource[item.id].checked = item.values.map(
        (item: any) => item.id
      );
    }
  });
};
const hostDistinct = ref<TopoHostDistinctRespData | null>();
// 筛选
const getHostDistinct = async () => {
  const params = {
    exact_include_conditions: {
      node_role: ['agent', 'blank']
    }
  };
  const res = await TopoService.HostDistinct(params).catch(() => null);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      const curUniqueValues = res[key] || [];
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = curUniqueValues
          .filter((item: any) => item !== "")
          .map((value: string) => ({
            text: key === "os_type" && osMap[value] ? osMap[value] : value,
            value: value,
          }));
      }
    });
  }
};
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  if (field === "bk_networkarea_name") {
    if (checked.length) {
      tableData.value = agentList.value.filter((item: any) =>
        checked.includes(item[field])
      );
    } else {
      tableData.value = agentList.value;
    }
  } else {
    const index = searchSelectValue.value.findIndex(
      (item: any) => item.id === field
    );
    index > -1 && searchSelectValue.value.splice(index, 1);
    if (checked.length) {
      searchSelectValue.value.push({
        id: field,
        name: t(field),
        values: checked.map((item: any) => ({
          id: item,
          name: field === "os_type" && osMap[item] ? osMap[item] : item,
        })),
      });
    }
  }
};
// 分页操作
const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  await getAgentList();
};
const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getAgentList();
};

// 批量操作
const operate = [
  {
    id: "upgrade",
    name: t("升级"),
    disabled: false,
    show: true,
  },
  {
    id: "uninstall",
    name: t("卸载"),
    disabled: false,
    show: true,
  },
  {
    id: "reload",
    name: t("重载配置"),
    disabled: false,
    show: true,
  },
  {
    id: "reboot",
    name: t("重启"),
    disabled: false,
    show: true,
  },
  {
    id: "log",
    name: t("最新执行日志"),
    disabled: false,
    show: true,
    single: true,
  },
];
// 安装方式
const agentInstallType = [
  {
    id: "setup",
    name: t("普通远程安装"),
  },
  {
    id: "import",
    name: t("Excel 导入远程安装"),
  },
  {
    id: "manual",
    name: t("手动安装"),
  },
];
const osMap = {
  windows: "Windows",
  unknown: "Unknown",
  darwin: "Darwin",
  linux: "Linux",
};
const dropdownShow = ref(false);
const handleInstall = () => {
  if (selection.value.length) {
    dropdownShow.value = false;
    triggerHandler("reinstall");
  } else {
    dropdownShow.value = !dropdownShow.value;
  }
};

// 搜索
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(
    new Set(
      agentList.value.map((item: any) => item[prop]).filter((item: any) => item)
    )
  );
  return uniqueValues.map((value) => ({
    id: value,
    name: String(value),
  }));
};
const getUniqueChildrenFrom = <K extends keyof TopoHostDistinctRespData>(
  prop: K,
  keyMap?: Record<string, any>
) => {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== "")
    .map((value: any) => ({
      id: value,
      name: keyMap && keyMap[value] ? keyMap[value] : String(value),
    }));
};
const searchSelectData = computed(() => [
  {
    id: "bk_host_innerip",
    name: t("platform.nodeMan.inner_ip"),
    multiple: true,
  },
  {
    id: "bk_host_innerip_v6",
    name: t("platform.nodeMan.inner_ipv6"),
    multiple: true,
  },
  {
    id: "bk_networkarea_id",
    name: "管控区域ID:IP",
    children: getUniqueChildrenFrom("bk_networkarea_id"),
    multiple: true,
  },
  { id: "bk_agent_id", name: "Agent ID", multiple: true },
  // {id: 'bk_networkarea_name', name: '管控区域', children: getUniqueChildren('bk_networkarea_name')},
  {
    id: "bk_networkunit_id",
    name: "管控单元",
    children: getUniqueChildrenFrom("bk_networkunit_id"),
    multiple: true,
  },
  {
    id: "os_type",
    name: "操作系统",
    children: getUniqueChildrenFrom("os_type", osMap),
    multiple: true,
  },
  {
    id: "node_version",
    name: "Agent版本",
    children: getUniqueChildrenFrom("node_version"),
    multiple: true,
  },
  {
    id: "node_status",
    name: "Agent 状态",
    children: getUniqueChildrenFrom("node_status"),
    multiple: true,
  },
]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    "bk_host_innerip",
    "bk_host_innerip_v6",
    "bk_agent_id",
    "bk_networkarea_name",
    "bk_networkarea_id",
    "bk_networkunit_id",
    "os_type",
    "node_version",
    "node_status",
    "action",
  ],
  disabled: ["action"],
});

const areaFilterOption = reactive({
  list: computed(() =>
    Array.from(
      new Set(
        agentList.value
          .map((item: any) => item["bk_networkarea_name"])
          .filter((item: any) => item !== "")
      )
    ).map((value) => ({
      text: value,
      value: value,
    }))
  ),
  checked: [] as string[],
  filterScope: "all",
});
const filterOptionSource: Record<string, FilterOption> = reactive({
  bk_networkarea_id: {
    list: [],
    checked: [],
    filterScope: "all",
  },
  bk_networkunit_id: {
    list: [],
    checked: [],
    filterScope: "all",
  },
  os_type: {
    list: [],
    checked: [],
    filterScope: "all",
  },
  node_version: {
    list: [],
    checked: [],
    filterScope: "all",
  },
  node_status: {
    list: [],
    checked: [],
    filterScope: "all",
  },
});

const triggerHandler = (type: string, setupType: string = "setup") => {
  switch (type) {
    // 批量重启 批量重装 批量重载配置 批量卸载 批量升级
    case "reboot":
    case "reinstall":
    case "reload":
    case "uninstall":
    case "upgrade":
    case "remove":
      handleOperate(type, selection.value, true);
      break;
    case "setup":
      router.push({ name: "agentSetup" });
      mainStore.updateAgentSetupType(setupType);
      break;
  }
};
/**
 * 当前操作项是否显示
 */
const getOperateShow = (row: Host, config: IOperateItem) => {
  if (config.id === "log" && (!row.job_result || !row.job_result.job_id)) {
    return false;
  }
  return config.show;
};
// 操作
const handleOperate = (type: string, data: Host[], batch = false) => {
  if (
    !batch &&
    ["terminated", "not_installed"].includes(data[0].state.node_version) &&
    !["log", "reinstall", "remove"].includes(type)
  ) {
    return;
  }

  let jobType = "";

  switch (type) {
    // 重启
    case "reboot":
      handleOperatetHost(data, batch, "RESTART_AGENT");
      break;
    // 移除
    case "remove":
      handleOperatetHost(data, batch, "REMOVE_AGENT");
      break;
    // 重装
    case "reinstall":
      jobType = "REINSTALL_AGENT";
      break;
    // 重载 只取其中一部分数据
    case "reload":
      jobType = "RELOAD_AGENT";
      break;
    // 卸载
    case "uninstall":
      jobType = "UNINSTALL_AGENT";
      break;
    // 升级
    case "upgrade":
      handleOperatetHost(data, batch, "UPGRADE_AGENT");
      break;
    // 日志详情
    case "log":
      // handleGotoLog(data[0]);
      break;
  }
  if (!jobType) return;

  router.push({ name: "agentEdit" });
  const params = {
    tableData: data.map((item: any) => ({
      ...item,
    })),
    type: jobType,
    // true：跨页全选（tableData表示标记删除的数据） false：非跨页全选（tableData表示编辑的数据）
    isSelectedAllPages: String(batch && isSelectedAllPages.value),
  };
  nodeManageStore.updateAgentEditRowData(params);
};
// 表格勾选
const selection = computed(() =>
  tableData.value.filter((item: any) => item.checked)
);
const handleSelectChange = ({
  checked,
  row,
}: {
  checked: boolean;
  row: any;
}) => {
  row.checked = checked;
};

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  tableData.value.forEach((item: any) => (item.checked = checked));
};
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
const handleOperatetHost = async (
  data: Host[],
  batch: boolean,
  operateType: string
) => {
  const titleObj = {
    firstIp: data[0].info.bk_host_innerip,
    num: tableData.value.length,
  };
  const operateJob = async (data: Host[]) => {
    loading.value = true;
    const params = {
      hosts: data.map((item: any) => ({
        bk_host_id: item.info.bk_host_id,
        topo: {
          bk_biz_id: item.info.bk_biz_id,
          bk_networkarea_id: item.bk_networkarea_id,
          bk_networkunit_id: item.bk_networkunit_id,
        },
        attributes: {
          bk_host_innerip: item.bk_host_innerip,
          bk_host_innerip_v6: item.bk_host_innerip_v6,
          bk_host_outerip: item.bk_host_outerip,
          bk_host_outerip_v6: item.bk_host_outerip_v6,
          login_ip: item.login_ip,
          login_port: item.login_port,
          login_password: item.login_password,
        },
        config: {
          version: item.node_version,
        },
      })),
    };
    const result = await WorkflowService.AgentUpgrade(params);
    loading.value = false;
    if (result.workflow_id) {
      router.push({
        name: "taskDetail",
        params: { taskId: result.workflow_id, routerBackName: "taskList" },
      });
    }
    return;
  };
  let type = "";
  switch (operateType) {
    // 重启
    case "RESTART_AGENT":
      type = t("重启");
      break;
    // 升级
    case "UPGRADE_AGENT":
      type = t("升级");
      break;
    case "REMOVE_AGENT":
      type = t("移除");
      break;
  }

  InfoBox({
    title: batch ? `请确认是否批量${type}` : `请确认是否${type}`,
    subTitle: batch
      ? t("批量确认操作提示", {
          ip: titleObj.firstIp,
          num: titleObj.num,
          type,
          suffix: operateType === "UPGRADE_AGENT" ? t("到最新版本") : "",
        })
      : `${type} ${titleObj.firstIp} 的Agent${
          operateType === "UPGRADE_AGENT" ? t("到最新版本") : ""
        }`,
    extCls: "wrap-title",
    onConfirm: () => {
      if (operateType === "REMOVE_AGENT") {
        // handleRemoveHost(data);
      } else {
        operateJob(data);
      }
    },
  });
};
const fuzzyKeys = new Set([
  "bk_host_innerip",
  "bk_host_innerip_v6",
  "bk_host_name",
  "dept_name",
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1)*pagination.limit,
    },
    exact_include_conditions: {} as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  params.exact_include_conditions.node_role = ["agent", "blank"];
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values.map((value: any) => value.id);
  });
  return params;
};
const getAgentList = async () => {
  loading.value = true;
  const res = await TopoService.HostList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;
  tableData.value = res.items.map((item: any) => ({
    ...item.state,
    ...item.info,
    ...item,
  }));
  agentList.value = tableData.value;
  loading.value = false;
};
watch(
  searchSelectValue,
  async () => {
    await getAgentList();
  },
  { deep: true }
);
watch(route, async () => {
  if (route.query.os_type){
    searchSelectValue.value.push(...[
      {
        id: "os_type",
        name: "操作系统",
        values: [{
          id: route.query.os_type,
          name: capitalizeFirstLetter(route.query.os_type)
        }]
      },
      {
        id: "node_version",
        name: "Agent版本",
        values: [{
          id: route.query.node_version,
          name: route.query.node_version
        }]
      },
    ])
    await getAgentList();
  }
}, {immediate: true, deep: true})
onMounted(async () => {
  await getAgentList();
  await getHostDistinct();
});
</script>
<style lang="postcss" scoped>
:deep(.vxe-table--empty-content) {
  height: 200px;
  line-height: 200px;
}
.dropDown-menu {
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

.status-icon::before {
  content: "";
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