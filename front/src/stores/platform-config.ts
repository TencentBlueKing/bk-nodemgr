import { defineStore } from 'pinia';

export const usePlatformConfigStore = defineStore('platform-config', {
  state: () => ({
    bkAppCode: '', // appcode
    name: '蓝鲸节点管理', // 站点的名称，通常显示在页面左上角，也会出现在网页title中
    nameEn: 'bk-nodemgr', // 站点的名称-英文
    appLogo: '', // 站点logo
    favicon: '/static/images/favicon.icon', // 站点favicon
    helperText: '',
    helperTextEn: '',
    helperLink: '',
    brandImg: '',
    brandImgEn: '',
    brandName: '', // 品牌名，会用于拼接在站点名称后面显示在网页title中
    favIcon: '',
    brandNameEn: '', // 品牌名-英文
    productName: '蓝鲸节点管理',
    productNameEn: 'bk-nodemgr',
    footerInfo: '', // 页脚的内容，仅支持 a 的 markdown 内容格式
    footerInfoEn: '', // 页脚的内容-英文
    footerCopyright: '', // 版本信息，包含 version 变量，展示在页脚内容下方
    footerInfoHTML: '',
    footerInfoHTMLEn: '',
    footerCopyrightContent: '',
    version: '', // 版本号
    // 需要国际化的字段，根据当前语言cookie自动匹配，页面中应该优先使用这里的字段
    i18n: {
      name: '',
      productName: '',
      helperText: '...',
      brandImg: '...',
      brandName: '...',
      footerInfoHTML: '...',
    },
  }),
  actions: {},
});
