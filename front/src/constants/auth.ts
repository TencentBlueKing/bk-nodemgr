import type { PermissionData } from '@/stores/permission';

export interface PageAuthItem {
  id: string;
  action: string;
  resourceType: string;
  routes: string[];
  activeScopes?: HistoryActiveScope[];
}

export type HistoryActiveScope = 'agent' | 'proxy' | 'plugin';

export interface RouteAuthTarget {
  name?: string | null | symbol;
  query?: Record<string, unknown>;
}

// AuthResource aligns with backend auth.Resource structure
export interface AuthResource {
  system_id: string;
  type: string;
  id: string;
}

// AuthVerifyItem represents action + resources pair
export interface AuthVerifyItem {
  action: string;
  resources?: AuthResource[];
}

export interface AuthVerifyResult {
  action: string;
  authorized: boolean;
}

export interface AuthVerifyResp {
  results: AuthVerifyResult[];
  permission?: PermissionData;
}

// IAM system IDs
export const SYSTEM_ID_CMDB = 'bk_cmdb';
export const SYSTEM_ID_NODEMGR = 'bk_nodemgr';

// Resource type to system ID mapping
export function getSystemIdForResourceType(resourceType: string): string {
  if (resourceType === 'biz') {
    return SYSTEM_ID_CMDB;
  }

  return SYSTEM_ID_NODEMGR;
}

export const PAGE_AUTH_CONFIG: PageAuthItem[] = [
  // nodeManager — biz-scoped pages
  { id: 'agent_view', action: 'agent_view', resourceType: 'biz', routes: ['agent'] },
  { id: 'proxy_view', action: 'proxy_view', resourceType: 'biz', routes: ['proxy'] },
  { id: 'agent_operate', action: 'agent_operate', resourceType: 'biz', routes: ['agentSetup', 'agentEdit', 'assignUnit'] },
  { id: 'plugin_view', action: 'plugin_view', resourceType: 'biz', routes: ['plugin'] },
  { id: 'agent_history_view', action: 'agent_history_view', resourceType: 'biz', routes: ['history', 'taskDetail', 'log'], activeScopes: ['agent'] },
  { id: 'proxy_history_view', action: 'proxy_history_view', resourceType: 'biz', routes: ['history', 'taskDetail', 'log'], activeScopes: ['proxy'] },
  { id: 'plugin_history_view', action: 'plugin_history_view', resourceType: 'biz', routes: ['history', 'taskDetail', 'log'], activeScopes: ['plugin'] },

  // topoManager — non-biz pages
  { id: 'networkarea_view', action: 'networkarea_view', resourceType: 'networkarea', routes: ['workarea', 'workareaDetail'] },
  { id: 'networkunit_view', action: 'networkunit_view', resourceType: 'networkunit', routes: ['topo'] },
  { id: 'networkarea_history_view', action: 'networkarea_history_view', resourceType: 'networkarea', routes: ['record'] },

  // ruleManager — biz-scoped pages
  { id: 'deploy_policy_view', action: 'deploy_policy_view', resourceType: 'biz', routes: ['agentStrategy', 'proxyStrategy'] },
  { id: 'config_policy_view', action: 'config_policy_view', resourceType: 'biz', routes: ['pluginStrategy'] },
  { id: 'deploy_policy_history_view', action: 'deploy_policy_history_view', resourceType: 'biz', routes: ['strategyTaskHistory'] },

  // pkgManager — non-biz pages
  { id: 'package_view', action: 'package_view', resourceType: 'package_type', routes: ['agentPackageMng', 'proxyPackageMng', 'certPackageMng', 'bintoolPackageMng', 'plugin_bintoolPackageMng', 'pluginPackageMng'] },
  { id: 'package_history_view', action: 'package_history_view', resourceType: 'package_type', routes: ['operationRecords'] },
];

export function normalizeHistoryActive(active: unknown): HistoryActiveScope {
  if (active === 'proxy' || active === 'plugin') return active;

  return 'agent';
}

export function shouldDeferBizAuthCheck(
  authItem: PageAuthItem | undefined,
  isBusinessReady: boolean,
  currentBizId?: string | number,
): boolean {
  if (!authItem || authItem.resourceType !== 'biz') return false;

  return !isBusinessReady || currentBizId === undefined || currentBizId === null || currentBizId === '';
}

export function matchPageAuth(
  route: string | undefined | null | symbol | RouteAuthTarget,
  config: PageAuthItem[],
): PageAuthItem | undefined {
  const routeName = typeof route === 'object' && route !== null ? route.name : route;
  if (!routeName || typeof routeName !== 'string') return undefined;

  const activeScope = typeof route === 'object' && route !== null
    ? normalizeHistoryActive(route.query?.active)
    : 'agent';

  return config.find(item => item.routes.includes(routeName)
    && (!item.activeScopes?.length || item.activeScopes.includes(activeScope)));
}
