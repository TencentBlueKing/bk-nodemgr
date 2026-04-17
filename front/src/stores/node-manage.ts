import { defineStore } from 'pinia';

import type { NodeWorkflowInfo } from '@/@types/node_workflow';

interface SearchSelectItem {
  id: string;
  name: string;
  values: { id: string; name: string }[];
}

export interface HistoryFiltersState {
  active: string;
  searchSelectValue: SearchSelectItem[];
  dateStart: number; // timestamp in seconds
  dateEnd: number;
  page: number;
  limit: number;
  hideAutoTask: boolean;
}

export const useNodeManageStore = defineStore('nodeManageStore', {
  state: () => ({
    taskHistoryTableRowData: {} as NodeWorkflowInfo,
    taskHistoryFilters: null as HistoryFiltersState | null,
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
    assignUnitParams: {
      tableData: [] as Host[],
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
    // 保存任务历史筛选状态
    saveHistoryFilters(filters: HistoryFiltersState) {
      this.taskHistoryFilters = { ...filters };
    },
    // 读取并清除任务历史筛选状态
    consumeHistoryFilters(): HistoryFiltersState | null {
      const filters = this.taskHistoryFilters;
      this.taskHistoryFilters = null;
      return filters;
    },
    updateAgentEditRowData(params: any) {
      this.agentEditParams = { ...params };
    },
    updateAssignUnitParams(params: any) {
      this.assignUnitParams = { ...params };
    },
  },
});

