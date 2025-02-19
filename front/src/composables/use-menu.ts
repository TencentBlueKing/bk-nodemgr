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
}
export type MenuItem = (typeof navList)[number];
export type MainMenuNames = MenuItem['routeName'];

// 菜单配置
const navList = [
  {
    routeName: 'nodeManager',
    title: i18n.global.t('节点管理'),
    group: [
      {
        title: i18n.global.t('节点'),
        children: [
          {
            routeName: 'agent',
            icon: '',
            title: i18n.global.t('Agent状态'),
          },
          {
            routeName: 'plugin',
            icon: '',
            title: i18n.global.t('插件管理'),
          },
        ],
      },
      {
        title: i18n.global.t('历史'),
        children: [
          {
            routeName: 'history',
            icon: '',
            title: i18n.global.t('任务历史'),
          },
        ],
      },
    ],
  },
  {
    routeName: 'topoManager',
    title: i18n.global.t('拓扑管理'),
  },
  {
    routeName: 'ruleManager',
    title: i18n.global.t('策略管理'),
  },
  {
    routeName: 'pkgManager',
    title: i18n.global.t('包管理'),
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
