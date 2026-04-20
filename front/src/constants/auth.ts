import type { PermissionData } from '@/stores/permission';
import type { AuthorizedItem } from '@/@types/auth';

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

// AuthResource aligns with backend types.AuthResource structure
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

export interface DeferredBizAuthHandler {
  refreshPermissions: () => void;
}

// IAM system IDs — rendered from server-side template in index.html, with fallback defaults
export const SYSTEM_ID_CMDB = window.PROJECT_CONFIG?.BK_IAM_SYSTEM_ID_BK_CMDB || 'bk_cmdb';
export const SYSTEM_ID_NODEMGR = window.PROJECT_CONFIG?.BK_IAM_SYSTEM_ID_BK_NODEMGR || 'bk_nodemgr';

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
  { id: 'networkarea_view', action: 'networkarea_view', resourceType: 'networkarea', routes: ['workareaDetail'] },
  // networkunit_view 路由级校验已移除（后端 starts_with 兼容问题，由组件内权限控制替代）
  { id: 'networkarea_history_view', action: 'networkarea_history_view', resourceType: 'networkarea', routes: ['record'] },

  // ruleManager — biz-scoped pages
  { id: 'config_policy_view', action: 'config_policy_view', resourceType: 'biz', routes: ['agentStrategy', 'proxyStrategy'] },
  { id: 'config_policy_view', action: 'config_policy_view', resourceType: 'biz', routes: ['pluginStrategy'] },
  { id: 'config_policy_history_view', action: 'config_policy_history_view', resourceType: 'biz', routes: ['strategyTaskHistory'] },

  // pkgManager — non-biz pages
  { id: 'package_view', action: 'package_view', resourceType: 'package', routes: ['agentPackageMng', 'proxyPackageMng', 'certPackageMng', 'bintoolPackageMng', 'plugin_bintoolPackageMng', 'pluginPackageMng'] },
  { id: 'package_history_view', action: 'package_history_view', resourceType: 'package', routes: ['operationRecords'] },
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

export function handleDeferredBizAuthCheck(authStore: DeferredBizAuthHandler): true {
  authStore.refreshPermissions();

  return true;
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

/**
 * 按模块分组的 authorized items 配置
 * key = mainMenu（对应路由 meta.mainMenu）
 * 不同模块进入时只查询该模块需要的 action-resource_type 对
 */
export const AUTHORIZED_MODULE_ITEMS: Record<string, AuthorizedItem[]> = {
  // 业务选择器：仅需 biz_access 判断业务访问权限（所有页面共享，全局预加载）
  bizSelector: [
    { action: 'biz_access', resource_type: 'biz' },
  ],
  // 节点管理：agent / proxy / plugin 相关 biz 权限
  // 注：networkunit_use_for_agent/proxy 因后端 starts_with 兼容问题暂不放此模块，
  //      由使用方（如 list.vue）按需调用 fetchAuthorized 加载
  nodeManager: [
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'agent_operate', resource_type: 'biz' },
    { action: 'agent_history_view', resource_type: 'biz' },
    { action: 'proxy_view', resource_type: 'biz' },
    { action: 'proxy_operate', resource_type: 'biz' },
    { action: 'proxy_history_view', resource_type: 'biz' },
    { action: 'plugin_view', resource_type: 'biz' },
    { action: 'plugin_operate', resource_type: 'biz' },
    { action: 'plugin_history_view', resource_type: 'biz' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
  ],
  // 拓扑管理：管控区域 + 管控单元权限
  // 注：networkunit_use_for_agent/proxy 因后端 starts_with 兼容问题不放此模块，
  //      由使用方（agent/list.vue）按需调用 fetchAuthorized 加载
  topoManager: [
    { action: 'networkarea_view', resource_type: 'networkarea' },
    { action: 'networkarea_create', resource_type: 'networkarea' },
    { action: 'networkarea_edit', resource_type: 'networkarea' },
    { action: 'networkarea_delete', resource_type: 'networkarea' },
    { action: 'networkarea_history_view', resource_type: 'networkarea' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
    { action: 'networkunit_create', resource_type: 'networkunit' },
    { action: 'networkunit_edit', resource_type: 'networkunit' },
    { action: 'networkunit_delete', resource_type: 'networkunit' },
    { action: 'networkunit_history_view', resource_type: 'networkunit' },
  ],
  // 策略管理：config_policy 相关 biz 权限
  ruleManager: [
    { action: 'config_policy_view', resource_type: 'biz' },
    { action: 'config_policy_manage', resource_type: 'biz' },
    { action: 'config_policy_history_view', resource_type: 'biz' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
  ],
  // 包管理：package_type（上传按钮按包类型） + package（其他操作） 权限
  // package_type_upload 按 package_type 查询（根据路由确定是 agent/proxy/cert 等）
  // package_view / package_manage / package_history_view 按 package 查询
  pkgManager: [
    { action: 'package_type_upload', resource_type: 'package_type' },
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_history_view', resource_type: 'package' },
  ],
};

/**
 * 菜单路由名 → 所需 action 的映射
 * authorized 返回有权限时才显示对应菜单项
 * 未在此映射中的菜单项默认显示
 */
export const MENU_ROUTE_AUTH_MAP: Record<string, string> = {
  // pkgManager sub-menus
  agentPackageMng: 'package_view',
  proxyPackageMng: 'package_view',
  certPackageMng: 'package_view',
  bintoolPackageMng: 'package_view',
  plugin_bintoolPackageMng: 'package_view',
  pluginPackageMng: 'package_view',
  operationRecords: 'package_history_view',
  // ruleManager sub-menus
  agentStrategy: 'config_policy_view',
  proxyStrategy: 'config_policy_view',
  pluginStrategy: 'config_policy_view',
  strategyTaskHistory: 'config_policy_history_view',
};

/**
 * 获取指定模块的 authorized items
 * @param moduleName 模块名（mainMenu 值）
 * @returns 该模块需要的 action-resource_type 对列表
 */
export function getModuleAuthorizedItems(moduleName: string): AuthorizedItem[] {
  return AUTHORIZED_MODULE_ITEMS[moduleName] || [];
}
