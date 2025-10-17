import { computed } from 'vue';
import { useRoute } from 'vue-router';

import { i18n } from '../modules/i18n';

export interface NavGroup {
  title: string;
  children: Array<Omit<NavItem, 'group'>>
}
export interface NavItem {
  routeName: MainMenuNames;
  title: string;
  params?: Record<string, any>;
  group?: NavGroup[]
  icon?: string;
}
export type MenuItem = (typeof navList)[number];
export type MainMenuNames = MenuItem['routeName'];

// 菜单配置
const navList = [
  {
    routeName: 'nodeManager',
    title: i18n.global.t('platform.name'),
    group: [
      {
        title: i18n.global.t('platform.nodeMan.node'),
        children: [
          {
            routeName: 'agent',
            icon: 'nodeman-icon nc-state',
            title: i18n.global.t('platform.nodeMan.agentStatus.title'),
          },
          {
            routeName: 'proxy',
            icon: 'nodeman-icon nc-state',
            title: i18n.global.t('platform.nodeMan.proxyStatus.title'),
          },
          // {
          //   routeName: 'plugin',
          //   icon: 'nodeman-icon nc-plug-in',
          //   title: i18n.global.t('platform.nodeMan.pluginManagement'),
          // },
        ],
      },
      {
        title: i18n.global.t('platform.nodeMan.history'),
        children: [
          {
            routeName: 'history',
            icon: 'nodeman-icon nc-history',
            title: i18n.global.t('platform.nodeMan.taskHistory.title.mainTitle'),
          },
        ],
      },
    ],
  },
  {
    routeName: 'topoManager',
    title: i18n.global.t('platform.topoManagerName'),
    group: [
      {
        title: i18n.global.t('拓扑'),
        children: [
          {
            routeName: 'workarea',
            icon: 'nodeman-icon nc-workarea',
            title: i18n.global.t('topoManager.workArea.title'),
          },
          {
            routeName: 'topo',
            icon: 'nodeman-icon nc-topo',
            title: i18n.global.t('拓扑图'),
          },
        ],
      },
      {
        title: i18n.global.t('记录'),
        children: [
          {
            routeName: 'record',
            icon: 'nodeman-icon nc-record',
            title: i18n.global.t('操作记录'),
          },
        ],
      },
    ],
  },
  {
    routeName: 'ruleManager',
    title: i18n.global.t('platform.ruleManagerName'),
    group: [
      {
        title: i18n.global.t('策略'),
        children: [
          {
            routeName: 'agentStrategy',
            icon: 'nodeman-icon nc-state',
            title: i18n.global.t('Agent 策略'),
          },
          {
            routeName: 'proxyStrategy',
            icon: 'nodeman-icon nc-remote-install',
            title: i18n.global.t('Proxy 策略'),
          },
          {
            routeName: 'pluginStrategy',
            icon: 'nodeman-icon nc-plug-in',
            title: i18n.global.t('插件策略'),
          },
        ],
      },
      {
        title: i18n.global.t('历史'),
        children: [
          {
            routeName: 'strategyTaskHistory',
            icon: 'nodeman-icon nc-history',
            title: i18n.global.t('操作记录'),
          },
        ],
      },
    ],
  },
  {
    routeName: 'pkgManager',
    title: i18n.global.t('platform.pkgManagerName'),
    group: [
      {
        title: i18n.global.t('节点'),
        children: [
          {
            routeName: 'agentPackageMng',
            icon: 'nodeman-icon nc-package-agent',
            title: i18n.global.t('Agent 包管理'),
          },
          {
            routeName: 'proxyPackageMng',
            icon: 'nodeman-icon nc-package-2',
            title: i18n.global.t('Proxy 包管理'),
          },
          {
            routeName: 'certPackageMng',
            icon: 'nodeman-icon nc-backstage',
            title: i18n.global.t('证书管理'),
          },
          {
            routeName: 'bintoolPackageMng',
            icon: 'nodeman-icon nc-manual',
            title: i18n.global.t('工具管理'),
          },
        ],
      },
      {
        title: i18n.global.t('插件'),
        children: [
          {
            routeName: 'pluginPackageMng',
            icon: 'nodeman-icon nc-plug-in',
            title: i18n.global.t('插件包管理'),
          },
          {
            routeName: 'plugin_bintoolPackageMng',
            icon: 'nodeman-icon nc-manual',
            title: i18n.global.t('插件包工具管理'),
          },
        ],
      },
      {
        title: i18n.global.t('记录'),
        children: [
          {
            routeName: 'operationRecords',
            icon: 'nodeman-icon nc-record',
            title: i18n.global.t('操作记录'),
          },
        ],
      },
    ],
  },
];

export default function useMenu() {
  const route = useRoute();

  const navData = computed<NavItem[]>(() => navList.map(item => ({
    ...item,
    params: {},
  })));

  const subMenuData = computed(() => navData.value.find(item => item.routeName === route.meta?.mainMenu)?.group || []);

  return {
    navData,
    subMenuData,
  };
}
