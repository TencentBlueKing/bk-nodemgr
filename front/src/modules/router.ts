
import { setupLayouts } from 'virtual:generated-layouts';
import { createRouter, createWebHashHistory } from 'vue-router';

import NotFound from '@/pages/app/404.vue';
import AgentImport from '@/pages/node/agent/import.vue';
import AgentManager from '@/pages/node/agent/list.vue';
import AgentSetup from '@/pages/node/agent/setup.vue';
import TaskHistory from '@/pages/node/history.vue';
import NodeManager from '@/pages/node/index.vue';
import PluginManager from '@/pages/node/plugin.vue';
import OperationRecord from '@/pages/topo/record.vue';
import Topography from '@/pages/topo/topo.vue';
import WorkArea from '@/pages/topo/workarea/workarea.vue';
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
              title: 'Agent状态',
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
            },
          },
          {
            name: 'agentImport',
            path: 'import',
            props: true,
            component: AgentImport,
            meta: {
              title: 'Excel导入安装',
              back: true,
              mainMenu: 'nodeManager',
            },
          },
          {
            name: 'agentEdit',
            path: 'edit',
            props: true,
            component: AgentImport,
            meta: {
              title: '重装 Agent',
              back: true,
              mainMenu: 'nodeManager',
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
              mainMenu: 'nodeManager',
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
              back: false,
              mainMenu: 'topoManager',
              title: '管控区域',
            },
          },
          {
            name: 'workareaDetail',
            path: 'workarea-detail/:workarea',
            component: WorkareaDetail,
            meta: {
              back: true,
              mainMenu: 'topoManager',
              title: '管控区域详情',
            },
          },
          {
            name: 'topo',
            path: 'topo',
            component: Topography,
            meta: {
              back: false,
              mainMenu: 'topoManager',
            },
          },
          {
            name: 'record',
            path: 'record',
            component: OperationRecord,
            meta: {
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
        children: [],
      },
      // 包管理
      {
        name: 'pkgManager',
        path: 'pkg-manager',
        children: [],
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
