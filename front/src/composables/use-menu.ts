import { computed } from 'vue';
import { useRoute } from 'vue-router';

import { MENU_ROUTE_AUTH_MAP } from '@/constants/auth';
import { i18n } from '../modules/i18n';
import { useAuthStore } from '@/stores/auth';

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
            icon: 'nodeman-icon nc-agent-jiedian',
            title: i18n.global.t('platform.nodeMan.agentStatus.title'),
          },
          {
            routeName: 'proxy',
            icon: 'nodeman-icon nc-proxy-jiedian',
            title: i18n.global.t('platform.nodeMan.proxyStatus.title'),
          },
        ],
      },
      {
        title: i18n.global.t('pluginManagement.pluginGroupName'),
        children: [
          {
            routeName: 'plugin',
            icon: 'nodeman-icon nc-plug-in',
            title: i18n.global.t('pluginManagement.plugin.title'),
          },
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
        title: i18n.global.t('menu.topoGroup'),
        children: [
          {
            routeName: 'workarea',
            icon: 'nodeman-icon nc-workarea',
            title: i18n.global.t('topoManager.workArea.title'),
          },
          {
            routeName: 'topo',
            icon: 'nodeman-icon nc-topo',
            title: i18n.global.t('route.topo'),
          },
        ],
      },
      {
        title: i18n.global.t('menu.recordGroup'),
        children: [
          {
            routeName: 'record',
            icon: 'nodeman-icon nc-record',
            title: i18n.global.t('route.operationRecord'),
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
        title: i18n.global.t('menu.strategyGroup'),
        children: [
          {
            routeName: 'agentStrategy',
            icon: 'nodeman-icon nc-agentcelve',
            title: i18n.global.t('route.agentStrategy'),
          },
          {
            routeName: 'proxyStrategy',
            icon: 'nodeman-icon nc-proxycelve',
            title: i18n.global.t('route.proxyStrategy'),
          },
          // {
          //   routeName: 'pluginStrategy',
          //   icon: 'nodeman-icon nc-plug-in',
          //   title: i18n.global.t('route.pluginStrategy'),
          // },
        ],
      },
      {
        title: i18n.global.t('platform.nodeMan.history'),
        children: [
          {
            routeName: 'strategyTaskHistory',
            icon: 'nodeman-icon nc-history',
            title: i18n.global.t('route.operationRecord'),
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
        title: i18n.global.t('platform.nodeMan.node'),
        children: [
          {
            routeName: 'agentPackageMng',
            icon: 'nodeman-icon nc-package-agent-2',
            title: i18n.global.t('route.agentPackage'),
          },
          {
            routeName: 'proxyPackageMng',
            icon: 'nodeman-icon nc-package-proxy',
            title: i18n.global.t('route.proxyPackage'),
          },
          {
            routeName: 'certPackageMng',
            icon: 'nodeman-icon nc-backstage',
            title: i18n.global.t('route.certManage'),
          },
          {
            routeName: 'bintoolPackageMng',
            icon: 'nodeman-icon nc-manual',
            title: i18n.global.t('route.toolManage'),
          },
        ],
      },
      {
        title: i18n.global.t('pluginManagement.pluginGroupName'),
        children: [
          {
            routeName: 'pluginPackageMng',
            icon: 'nodeman-icon nc-plug-in',
            title: i18n.global.t('menu.pluginPackageMng'),
          },
          {
            routeName: 'plugin_bintoolPackageMng',
            icon: 'nodeman-icon nc-manual',
            title: i18n.global.t('route.pluginToolManage'),
          },
        ],
      },
      {
        title: i18n.global.t('menu.recordGroup'),
        children: [
          {
            routeName: 'operationRecords',
            icon: 'nodeman-icon nc-record',
            title: i18n.global.t('route.operationRecord'),
          },
        ],
      },
    ],
  },
];

export default function useMenu() {
  const route = useRoute();
  const authStore = useAuthStore();
  const currentMainMenu = computed(() => (route.meta?.mainMenu || (route.query.mainMenu as string)));

  const navData = computed<NavItem[]>(() => navList.map(item => ({
    ...item,
    params: {},
  })));

  const subMenuData = computed(() => {
    const groups = navData.value.find(item => item.routeName === currentMainMenu.value)?.group || [];
    // Filter menu items based on authorized permissions
    return groups
      .map(group => ({
        ...group,
        children: group.children.filter((child) => {
          const requiredAction = MENU_ROUTE_AUTH_MAP[child.routeName];
          if (!requiredAction) return true; // No auth config, always show
          if (!authStore.authorizedLoaded) return true; // Not loaded yet, show by default
          return authStore.hasAuthorizedResource(requiredAction);
        }),
      }))
      .filter(group => group.children.length > 0); // Remove empty groups
  });

  return {
    navData,
    subMenuData,
  };
}
