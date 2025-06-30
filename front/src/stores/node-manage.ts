import { defineStore } from 'pinia';

export const useNodeManageStore = defineStore('nodeManageStore', {
  state: () => ({
    taskHistoryTableRowData: null,
    currentStatus: ''
  }),
  actions: {
    // 更新任务历史表格行信息
    updateCurrentRowData(data: NodeWorkflowInfo) {
      this.taskHistoryTableRowData = data;
    },
    updateCurrentStatus(status: string) {
      this.currentStatus = status;
    }
  },
});

