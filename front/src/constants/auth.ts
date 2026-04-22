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
 *
 * 设计原则：
 * 1. 模块初始化只带「view 类」action（用于菜单/占位页判断是否无权限）
 * 2. 非 view 类（operate/manage/create/edit/delete/upload）由具体菜单页在 onMounted 中按需调用
 * 3. 避免跨模块 action 重复：bizSelector 专责 agent_view，nodeManager 不再重复
 */
export const AUTHORIZED_MODULE_ITEMS: Record<string, AuthorizedItem[]> = {
  // 业务选择器：仅需 biz_access 判断业务访问权限（所有页面共享，全局预加载）
  bizSelector: [
    { action: 'biz_access', resource_type: 'biz' },
  ],
  // 节点管理：模块内所有 view 类 biz 权限
  // 注：业务选择器的可访问业务范围由 biz_access 统一控制，不再依赖 agent_view
  // 注：operate/networkunit 类由具体菜单页按需调用 fetchAuthorized 加载
  nodeManager: [
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'proxy_view', resource_type: 'biz' },
    { action: 'plugin_view', resource_type: 'biz' },
    { action: 'agent_history_view', resource_type: 'biz' },
    { action: 'proxy_history_view', resource_type: 'biz' },
    { action: 'plugin_history_view', resource_type: 'biz' },
  ],
  // 拓扑管理：管控区域 / 管控单元 view 类权限
  // 注：create/edit/delete 由具体菜单页（workarea/topo/workareaDetail）按需加载
  topoManager: [
    { action: 'networkarea_view', resource_type: 'networkarea' },
    { action: 'networkarea_history_view', resource_type: 'networkarea' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
  ],
  // 策略管理：config_policy view 类 biz 权限
  // 注：业务选择器的可访问业务范围由 biz_access 统一控制，不再依赖 agent_view
  // 注：config_policy_manage 由具体策略页按需加载
  ruleManager: [
    { action: 'config_policy_view', resource_type: 'biz' },
    { action: 'config_policy_history_view', resource_type: 'biz' },
  ],
  // 包管理：view 类
  // 注：package_manage / package_type_upload 由具体菜单页（agent-proxy-pkg / cert-bintool-manage 等）按需加载
  pkgManager: [
    { action: 'package_view', resource_type: 'package' },
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
 * 菜单页「按需加载」的 authorized items
 * key = 路由 name，value = 进入该菜单页时需要额外加载的非 view 类 action 列表
 * 进入菜单页（onMounted）时由页面主动调用 fetchAuthorized 加载
 */
export const PAGE_AUTHORIZED_ITEMS: Record<string, AuthorizedItem[]> = {
  // ===== nodeManager =====
  agent: [
    { action: 'agent_operate', resource_type: 'biz' },
  ],
  proxy: [
    { action: 'proxy_operate', resource_type: 'biz' },
  ],
  plugin: [
    { action: 'plugin_operate', resource_type: 'biz' },
  ],
  // ===== topoManager =====
  workarea: [
    { action: 'networkarea_create', resource_type: 'networkarea' },
    { action: 'networkarea_edit', resource_type: 'networkarea' },
    { action: 'networkarea_delete', resource_type: 'networkarea' },
  ],
  workareaDetail: [
    { action: 'networkunit_create', resource_type: 'networkunit' },
    { action: 'networkunit_edit', resource_type: 'networkunit' },
    { action: 'networkunit_delete', resource_type: 'networkunit' },
    { action: 'networkarea_history_view', resource_type: 'networkarea' },
  ],
  topo: [
    { action: 'networkunit_edit', resource_type: 'networkunit' },
  ],
  // ===== ruleManager =====
  agentStrategy: [
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  proxyStrategy: [
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  pluginStrategy: [
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  // ===== pkgManager =====
  agentPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  proxyPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  certPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  bintoolPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  plugin_bintoolPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  pluginPackageMng: [
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
};

/**
 * 获取指定菜单页的按需 authorized items
 * @param routeName 路由 name
 */
export function getPageAuthorizedItems(routeName: string): AuthorizedItem[] {
  return PAGE_AUTHORIZED_ITEMS[routeName] || [];
}

/**
 * 获取指定模块的 authorized items
 * @param moduleName 模块名（mainMenu 值）
 * @returns 该模块需要的 action-resource_type 对列表
 */
export function getModuleAuthorizedItems(moduleName: string): AuthorizedItem[] {
  return AUTHORIZED_MODULE_ITEMS[moduleName] || [];
}
