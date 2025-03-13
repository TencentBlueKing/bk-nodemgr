import { defineStore } from 'pinia';

export const useMainStore = defineStore('mainStore', {
  state: () => ({
    globalPageSize: 50, // 全局分页
    windowInnerHeight: 0,
    businessList: [] as Business[], // 业务列表
    selectedBusinessId: [] as number[], // 当前业务id
    selectedBusinessName: [] as string[], // 当前业务名称
  }),
  actions: {
    // 更新全局分页
    updatePageSize(size: number) {
      this.globalPageSize = size;
    },
    // 更新窗口高度
    updateWindowInnerHeight(height: number) {
      this.windowInnerHeight = height;
    },
    updateBusinessList(list: Business[]) {
      console.log("🚀 ~ updateBusinessList ~ list:", list)
      this.businessList = list;
    },
    updateCurBusiness(businessId: number[] = []) {
      this.selectedBusinessId = businessId;
      this.selectedBusinessName = businessId.map(id => this.businessList.find(item => item.bk_biz_id === id)?.bk_biz_name) as string[];
    },
  },
});

