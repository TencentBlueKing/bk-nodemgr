import { defineStore } from 'pinia';
import type { NodeWorkflowInfo } from '@/@types/node_workflow';

export const useNodeManageStore = defineStore('nodeManageStore', {
  state: () => ({
    taskHistoryTableRowData: {} as NodeWorkflowInfo,
    agentEditParams: {
      tableData: [] as Host[],
      type: '',
      isSelectedAllPages: false,
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

