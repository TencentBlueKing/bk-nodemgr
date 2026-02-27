
import { setupLayouts } from 'virtual:generated-layouts';
import { createRouter, createWebHashHistory } from 'vue-router';
import { i18n } from '@/modules/i18n';

import { cancelRequest } from '@/api/request-queue';
import NotFound from '@/pages/app/404.vue';
import AgentImport from '@/pages/node/agent/import.vue';
import AgentManager from '@/pages/node/agent/list.vue';
import AgentSetup from '@/pages/node/agent/setup.vue';
import TaskHistory from '@/pages/node/history/history.vue';
import Log from '@/pages/node/history/log.vue';
import TaskDetail from '@/pages/node/history/task-detail.vue';
import NodeManager from '@/pages/node/index.vue';
import PluginManager from '@/pages/node/plugin/plugin.vue';
import AgentPackageMng from '@/pages/pkg/agent-proxy-pkg/list.vue';
import CertBintoolMng from '@/pages/pkg/cert-bintool-manage/list.vue';
import PluginPackageMng from '@/pages/pkg/plugin-package-manage.vue';
import OperationRecords from '@/pages/pkg/record.vue';
import AgentStrategy from '@/pages/rules/agent-strategy/index.vue';
import Rules from '@/pages/rules/index.vue';
import RulesRecord from '@/pages/rules/record/record.vue';
import OperationRecord from '@/pages/topo/record/record.vue';
import Topography from '@/pages/topo/topography/topo.vue';
import WorkArea from '@/pages/topo/workarea/workarea.vue';
import proxyStatus from '@/pages/topo/workarea-detail/components/proxy-info.vue';
import WorkareaDetail from '@/pages/topo/workarea-detail/workarea-detail.vue';
import type { UserModule } from '@/types';

const routes = setupLayouts([
  {
    path: import.meta.env.BK_SITE_URL,
    redirect: { name: 'agent' },
    children: [
      // 节点管理
      {
        name: 'nodeManager',
        path: 'node-manager',
        component: NodeManager,
        redirect: { name: 'agent' },
        children: [
          {
            name: 'agent',
            path: 'agent',
            component: AgentManager,
            meta: {
              subTitle: i18n.global.t('route.agentSubtitle'),
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'proxy',
            path: 'proxy',
            component: proxyStatus,
            meta: {
              title: i18n.global.t('platform.nodeMan.proxyStatus.title'),
              subTitle: i18n.global.t('route.proxyStatusSubtitle'),
              back: false,
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'agentSetup',
            path: 'setup',
            component: AgentSetup,
            meta: {
              title: i18n.global.t('route.installAgent'),
              back: true,
              mainMenu: 'nodeManager',
              parentName: 'agent', // 增加父路由节点名字，用来显示当前菜单
            },
          },
          {
            name: 'agentEdit',
            path: 'edit',
            props: true,
            component: AgentImport,
            meta: {
              title: i18n.global.t('route.installReinstallAgent'),
              back: true,
              mainMenu: 'nodeManager',
              parentName: 'agent',
            },
          },
          {
            name: 'plugin',
            path: 'plugin',
            component: PluginManager,
            meta: {
              mainMenu: 'nodeManager',
              title: i18n.global.t('route.plugin'),
              subTitle: i18n.global.t('route.pluginSubtitle'),
              back: false,
            },
          },
          {
            name: 'history',
            path: 'history',
            component: TaskHistory,
            meta: {
              title: i18n.global.t('route.taskHistory'),
              subTitle: i18n.global.t('route.taskHistorySubtitle'),
              back: false,
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'taskDetail',
            path: 'history/detail/:taskId',
            component: TaskDetail,
            meta: {
              mainMenu: 'nodeManager',
              parentName: 'history',
            },
          },
          {
            name: 'log',
            path: 'history/detail/:taskId/log/:hostId',
            component: Log,
            meta: {
              mainMenu: 'nodeManager',
              parentName: 'history',
            },
          },
        ],
      },
      // 拓扑管理
      {
        name: 'topoManager',
        path: 'topo-manager',
        redirect: { name: 'workarea' },
        children: [
          {
            name: 'workarea',
            path: 'workarea',
            component: WorkArea,
            meta: {
              title: i18n.global.t('route.workarea'),
              subTitle: i18n.global.t('route.workareaSubtitle'),
              back: false,
              mainMenu: 'topoManager',
            },
          },
          {
            name: 'workareaDetail',
            path: 'workarea-detail/:workarea/:workUnit?',
            component: WorkareaDetail,
            meta: {
              back: true,
              mainMenu: 'topoManager',
              title: i18n.global.t('route.workareaDetail'),
              parentName: 'workarea',
            },
          },
          {
            name: 'topo',
            path: 'topo',
            component: Topography,
            meta: {
              title: i18n.global.t('route.topo'),
              subTitle: i18n.global.t('route.topoSubtitle'),
              back: false,
              mainMenu: 'topoManager',
            },
          },
          {
            name: 'record',
            path: 'record',
            component: OperationRecord,
            meta: {
              title: i18n.global.t('route.operationRecord'),
              subTitle: i18n.global.t('route.topoOperationRecordSubtitle'),
              back: false,
              mainMenu: 'topoManager',
            },
          },
        ],
      },
      // 策略管理
      {
        name: 'ruleManager',
        path: 'rule-manager',
        component: Rules,
        redirect: { name: 'agentStrategy' },
        children: [
          {
            name: 'agentStrategy',
            path: 'agentStrategy',
            component: AgentStrategy,
            meta: {
              title: i18n.global.t('route.agentStrategy'),
              subTitle: i18n.global.t('route.agentStrategySubtitle'),
              back: false,
              mainMenu: 'ruleManager',
            },
          },

          {
            name: 'proxyStrategy',
            path: 'proxyStrategy',
            component: AgentStrategy,
            meta: {
              title: i18n.global.t('route.proxyStrategy'),
              subTitle: i18n.global.t('route.proxyStrategySubtitle'),
              back: false,
              mainMenu: 'ruleManager',
            },
          },
          {
            name: 'pluginStrategy',
            path: 'pluginStrategy',
            component: Rules,
            meta: {
              title: i18n.global.t('route.pluginStrategy'),
              back: false,
              mainMenu: 'ruleManager',
            },
          },
          {
            name: 'strategyTaskHistory',
            path: 'strategy-task-history',
            component: RulesRecord,
            meta: {
              title: i18n.global.t('route.operationRecord'),
              subTitle: i18n.global.t('route.strategyTaskHistorySubtitle'),
              back: false,
              mainMenu: 'ruleManager',
            },
          },
        ],
      },
      // 包管理
      {
        name: 'pkgManager',
        path: 'pkg-manager',
        redirect: { name: 'agentPackageMng' },
        children: [
          {
            name: 'agentPackageMng',
            path: 'agentPackageMng',
            component: AgentPackageMng,
            meta: {
              title: i18n.global.t('route.agentPackage'),
              subTitle: i18n.global.t('route.agentPackageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'proxyPackageMng',
            path: 'proxyPackageMng',
            component: AgentPackageMng,
            meta: {
              title: i18n.global.t('route.proxyPackage'),
              subTitle: i18n.global.t('route.proxyPackageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'certPackageMng',
            path: 'certPackageMng',
            component: CertBintoolMng,
            meta: {
              title: i18n.global.t('route.certManage'),
              subTitle: i18n.global.t('route.certManageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'bintoolPackageMng',
            path: 'bintoolPackageMng',
            component: CertBintoolMng,
            meta: {
              title: i18n.global.t('route.toolManage'),
              subTitle: i18n.global.t('route.toolManageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'plugin_bintoolPackageMng',
            path: 'plugin_bintoolPackageMng',
            component: CertBintoolMng,
            meta: {
              title: i18n.global.t('route.pluginToolManage'),
              subTitle: i18n.global.t('route.pluginToolManageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'pluginPackageMng',
            path: 'pluginPackageMng',
            component: PluginPackageMng,
            meta: {
              subTitle: i18n.global.t('route.pluginPackageSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'operationRecords',
            path: 'operationRecords',
            component: OperationRecords,
            meta: {
              title: i18n.global.t('route.operationRecord'),
              subTitle: i18n.global.t('route.pkgOperationRecordsSubtitle'),
              back: false,
              mainMenu: 'pkgManager',
            },
          },
        ],
      },
    ],
  },
  { path: '/:pathMatch(.*)*', name: '404', component: NotFound },
]);

// Setup router
// https://router.vuejs.org/zh/guide/
export const install: UserModule = ({ app }) => {
  const router = createRouter({
    history: createWebHashHistory(import.meta.env.BK_SITE_URL),
    routes,
  });
  router.beforeEach(() => {
    cancelRequest();
  });
  app.use(router);
};
