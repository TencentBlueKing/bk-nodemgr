import { defineStore } from 'pinia';

export const useMainStore = defineStore('mainStore', {
  state: () => ({
    globalPageSize: 50, // 全局分页
    windowInnerHeight: 0,
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
  },
});

