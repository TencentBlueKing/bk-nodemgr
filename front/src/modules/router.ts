
import { setupLayouts } from 'virtual:generated-layouts';
import { createRouter, createWebHashHistory } from 'vue-router';

import NotFound from '@/pages/app/404.vue';
import AgentImport from '@/pages/node/agent/import.vue';
import AgentManager from '@/pages/node/agent/list.vue';
import AgentSetup from '@/pages/node/agent/setup.vue';
import TaskHistory from '@/pages/node/history/history.vue';
import Log from '@/pages/node/history/log.vue';
import TaskDetail from '@/pages/node/history/task-detail.vue';
import NodeManager from '@/pages/node/index.vue';
import PluginManager from '@/pages/node/plugin.vue';
import AgentPackageMng from '@/pages/pkg/agent-proxy-pkg/list.vue';
import CertBintoolMng from '@/pages/pkg/cert-bintool-manage/list.vue';
import PluginPackageMng from '@/pages/pkg/plugin-package-manage.vue';
import OperationRecords from '@/pages/pkg/record.vue';
import CreateConfig from '@/pages/rules/agent-strategy/create-config.vue';
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
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'proxy',
            path: 'proxy',
            component: proxyStatus,
            meta: {
              title: 'Proxy状态',
              back: false,
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'agentSetup',
            path: 'setup',
            component: AgentSetup,
            meta: {
              title: '安装 Agent',
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
              title: '安装/重装 Agent',
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
            },
          },
          {
            name: 'history',
            path: 'history',
            component: TaskHistory,
            meta: {
              title: '任务历史',
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
            path: 'history/detail/:taskId/log/:ip',
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
              title: '管控区域',
              subTitle: '管控区域是互相之间能直接通信的一组服务器单元，如企业内的局域网、公有云VPC（虚拟私有网络）。',
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
              title: '管控区域详情',
              parentName: 'workarea',
            },
          },
          {
            name: 'topo',
            path: 'topo',
            component: Topography,
            meta: {
              title: '拓扑图',
              subTitle: '拓扑图展示各个管控单元之间的拓扑联系',
              back: false,
              mainMenu: 'topoManager',
            },
          },
          {
            name: 'record',
            path: 'record',
            component: OperationRecord,
            meta: {
              title: '操作记录',
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
              title: 'Agent 策略',
              back: false,
              mainMenu: 'ruleManager',
            },
          },

          {
            name: 'proxyStrategy',
            path: 'proxyStrategy',
            component: AgentStrategy,
            meta: {
              title: 'Proxy 策略',
              back: false,
              mainMenu: 'ruleManager',
            },
          },
          {
            name: 'pluginStrategy',
            path: 'pluginStrategy',
            component: Rules,
            meta: {
              title: '插件策略',
              back: false,
              mainMenu: 'ruleManager',
            },
          },
          {
            name: 'strategyTaskHistory',
            path: 'strategy-task-history',
            component: RulesRecord,
            meta: {
              title: '操作记录',
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
              title: 'Agent 包管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'proxyPackageMng',
            path: 'proxyPackageMng',
            component: AgentPackageMng,
            meta: {
              title: 'Proxy 包管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'certPackageMng',
            path: 'certPackageMng',
            component: CertBintoolMng,
            meta: {
              title: '证书管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'bintoolPackageMng',
            path: 'bintoolPackageMng',
            component: CertBintoolMng,
            meta: {
              title: '工具管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'plugin_bintoolPackageMng',
            path: 'plugin_bintoolPackageMng',
            component: CertBintoolMng,
            meta: {
              title: '插件包工具管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'pluginPackageMng',
            path: 'pluginPackageMng',
            component: PluginPackageMng,
            meta: {
              title: '插件包管理',
              back: false,
              mainMenu: 'pkgManager',
            },
          },
          {
            name: 'operationRecords',
            path: 'operationRecords',
            component: OperationRecords,
            meta: {
              title: '操作记录',
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
  app.use(router);
};
