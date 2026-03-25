import { defineStore } from 'pinia';

export const useMainStore = defineStore('mainStore', {
  state: (): {
    globalPageSize: number;
    windowInnerHeight: number;
    businessList: Business[];
    selectedBusinessId: number[];
    selectedBusinessName: string[];
    agentSetupType: string;
    proxySetupType: string;
    configEditData: ConfigPolicy | null;
    curLanguage: string;
    routeState: Object,
    isLogRetry: Boolean,
    isLogTerminate: Boolean,
    noticeShow: Boolean,
    isBusinessReady: boolean,
    strategyBizId: number,
  } => ({
    globalPageSize: 50, // 全局分页
    windowInnerHeight: 0,
    businessList: [] as Business[], // 业务列表
    selectedBusinessId: [] as number[], // 当前业务id
    selectedBusinessName: [] as string[], // 当前业务名称
    isBusinessReady: false, // 业务初始化是否完成
    strategyBizId: 0, // 策略管理 - 当前选中的业务ID
    agentSetupType: 'setup', // 代理安装方式
    proxySetupType: 'setup',
    configEditData: null,
    curLanguage: 'zh-CN',
    routeState: {},
    isLogRetry: false,
    isLogTerminate: false,
    noticeShow: window.PROJECT_CONFIG.ENABLE_NOTICE === 'true',
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
      this.businessList = list;
    },

    updateCurBusiness(businessId: number[] = []) {
      this.selectedBusinessId = businessId;
      this.selectedBusinessName = businessId.map(id =>
        this.businessList.find(item => item.bk_biz_id === id)?.bk_biz_name) as string[];
    },
    updateAgentSetupType(type: string) {
      this.agentSetupType = type;
    },
    updateProxySetupType(type: string) {
      this.proxySetupType = type;
    },
    updateConfigEditData(data: ConfigPolicy) {
      this.configEditData = data;
    },
    updateLanguage(language: string) {
      this.curLanguage = language;
    },
    updateRouteState(routeState: {}) {
      this.routeState = routeState;
    },
    updateLogRetry(isRetry: Boolean) {
      this.isLogRetry = isRetry;
    },
    updateLogTerminate(isTerminate: Boolean) {
      this.isLogTerminate = isTerminate;
    },
    updateNoticeShow(isShow: boolean) {
      this.noticeShow = isShow;
    },
    setBusinessReady() {
      this.isBusinessReady = true;
    },
    updateStrategyBizId(bizId: number) {
      this.strategyBizId = bizId;
    },
  },
});

