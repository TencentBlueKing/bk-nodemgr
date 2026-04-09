import { ref } from 'vue';

import { PluginAPIService } from '@/api/modules/plugin';

export interface PluginPermissionMap {
  /** pluginName → Set<actionName> */
  [pluginName: string]: Set<string>;
}

// 接口返回 reconfigv，前端按钮 id 是 reload，做一次映射
const ACTION_ALIAS: Record<string, string> = {
  reconfigv: 'reload',
};

/**
 * 插件操作权限 composable
 * 调用 list_permitted_operations 获取每个插件可执行的操作
 */
export default function usePluginAuth() {
  const permissionMap = ref<PluginPermissionMap>({});
  const allActions = ref(new Set<string>());
  const loaded = ref(false);
  const loading = ref(false);

  const load = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
      const res = await PluginAPIService.ListPluginPermittedOperation({
        page: { limit: 500, offset: 0 },
      }).catch(() => ({ operations: [] }));

      const map: PluginPermissionMap = {};
      const all = new Set<string>();

      ((res as any)?.operations || []).forEach((op: any) => {
        const perms = op.permissions || op.permission || [];
        const actions = new Set<string>();
        perms.forEach((p: string) => {
          actions.add(p);
          all.add(p);
          if (ACTION_ALIAS[p]) {
            actions.add(ACTION_ALIAS[p]);
            all.add(ACTION_ALIAS[p]);
          }
        });
        map[op.name] = actions;
      });

      permissionMap.value = map;
      allActions.value = all;
      loaded.value = true;
    } finally {
      loading.value = false;
    }
  };

  /**
   * 判断某个操作是否有权限
   * @param action 操作名，如 install / upgrade / uninstall / restart / stop / reload
   * @param pluginName 可选，指定插件名；不传则检查全局是否有任一插件拥有该操作权限
   */
  const hasAction = (action: string, pluginName?: string): boolean => {
    if (!loaded.value) return false; // 未加载完成，默认无权限
    if (pluginName) {
      return permissionMap.value[pluginName]?.has(action) ?? false;
    }
    return allActions.value.has(action);
  };

  return {
    permissionMap,
    allActions,
    loaded,
    loading,
    load,
    hasAction,
  };
}
