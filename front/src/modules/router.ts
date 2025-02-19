
import { setupLayouts } from 'virtual:generated-layouts';
import { createRouter, createWebHashHistory } from 'vue-router';

import NotFound from '@/pages/app/404.vue';
import AgentManager from '@/pages/node/agent.vue';
import TaskHistory from '@/pages/node/history.vue';
import NodeManager from '@/pages/node/index.vue';
import PluginManager from '@/pages/node/plugin.vue';
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
        children: [],
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
