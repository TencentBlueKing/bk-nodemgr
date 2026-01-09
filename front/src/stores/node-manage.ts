import { defineStore } from 'pinia';

import type { NodeWorkflowInfo } from '@/@types/node_workflow';

export const useNodeManageStore = defineStore('nodeManageStore', {
  state: () => ({
    taskHistoryTableRowData: {} as NodeWorkflowInfo,
    agentEditParams: {
      tableData: [] as Host[],
      type: '',
      isCrossPageSelection: false,
      queryParams: {
        exact_include_conditions: {} as Record<string, string>,
        exact_exclude_conditions: {} as Record<string, string>,
        fuzzy_include_conditions: {} as Record<string, string>,
      },
    },
  }),
  actions: {
    // 更新任务历史表格行信息
    updateCurrentRowData(data: NodeWorkflowInfo) {
      this.taskHistoryTableRowData = data;
    },
    updateAgentEditRowData(params: any) {
      this.agentEditParams = { ...params };
    },
  },
});

