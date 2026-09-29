import type { AuthorizedItem } from '@/@types/auth';
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
  // query.active 可能是数组（URL 重复参数），取第一个元素
  const value = Array.isArray(active) ? active[0] : active;
  const normalized = typeof value === 'string' ? value.trim().toLowerCase() : '';
  if (normalized === 'proxy' || normalized === 'plugin') return normalized;

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
 *
 * 设计原则：
 * 1. 所有「view 类」权限不再模块级预加载，统一收敛到具体菜单页（PAGE_AUTHORIZED_ITEMS）按需加载；
 *    biz 类 view（agent_view/proxy_view/plugin_view/history 系列/config_policy_view）由路由守卫 verify 实时校验，
 *    页面渲染不读 authorizedMap（少数需要的页面自行 fetchAuthorized），无需模块级预加载
 * 2. 非 view 类（operate/manage/create/edit/delete/upload）由具体菜单页按需加载
 * 3. 唯一保留的模块级项：bizSelector 的 biz_access（业务选择器全局共享，所有页面都需要）
 */
export const AUTHORIZED_MODULE_ITEMS: Record<string, AuthorizedItem[]> = {
  // 业务选择器：仅需 biz_access 判断业务访问权限（所有页面共享，全局预加载）
  bizSelector: [
    { action: 'biz_access', resource_type: 'biz' },
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
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'agent_operate', resource_type: 'biz' },
  ],
  proxy: [
    { action: 'proxy_view', resource_type: 'biz' },
    { action: 'proxy_operate', resource_type: 'biz' },
    { action: 'networkunit_use_for_proxy', resource_type: 'networkunit' },
  ],
  plugin: [
    { action: 'plugin_view', resource_type: 'biz' },
    { action: 'plugin_operate', resource_type: 'biz' },
    // 插件进程侧栏 / 插件安装的 ip-selector 依赖 agent_view/proxy_view 计算 node_role 过滤，
    // 必须随页面级 items 一并加载，否则 authorizedLoaded 已置 true 时兜底不会触发，导致 Proxy 主机被误过滤
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'proxy_view', resource_type: 'biz' },
  ],
  pluginOperate: [
    { action: 'plugin_operate', resource_type: 'biz' },
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'proxy_view', resource_type: 'biz' },
  ],
  createPluginOperation: [
    { action: 'plugin_operate', resource_type: 'biz' },
    { action: 'agent_view', resource_type: 'biz' },
    { action: 'proxy_view', resource_type: 'biz' },
  ],
  assignUnit: [
    { action: 'networkunit_use_for_agent', resource_type: 'networkunit' },
  ],
  // 任务历史：同一路由按 active 参数区分 agent/proxy/plugin，三个历史 view 都要
  history: [
    { action: 'agent_history_view', resource_type: 'biz' },
    { action: 'proxy_history_view', resource_type: 'biz' },
    { action: 'plugin_history_view', resource_type: 'biz' },
  ],
  taskDetail: [
    { action: 'agent_history_view', resource_type: 'biz' },
    { action: 'proxy_history_view', resource_type: 'biz' },
    { action: 'plugin_history_view', resource_type: 'biz' },
  ],
  log: [
    { action: 'agent_history_view', resource_type: 'biz' },
    { action: 'proxy_history_view', resource_type: 'biz' },
    { action: 'plugin_history_view', resource_type: 'biz' },
  ],
  // ===== topoManager =====
  workarea: [
    { action: 'networkarea_view', resource_type: 'networkarea' },
    { action: 'networkarea_edit', resource_type: 'networkarea' },
    { action: 'networkarea_delete', resource_type: 'networkarea' },
    { action: 'networkunit_create', resource_type: 'networkarea' },
  ],
  workareaDetail: [
    { action: 'networkarea_view', resource_type: 'networkarea' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
    { action: 'networkarea_history_view', resource_type: 'networkarea' },
    { action: 'networkunit_create', resource_type: 'networkarea' },
    { action: 'networkunit_edit', resource_type: 'networkunit' },
    { action: 'networkunit_delete', resource_type: 'networkunit' },
  ],
  record: [
    { action: 'networkarea_history_view', resource_type: 'networkarea' },
  ],
  topo: [
    { action: 'networkarea_view', resource_type: 'networkarea' },
    { action: 'networkunit_view', resource_type: 'networkunit' },
    { action: 'networkunit_edit', resource_type: 'networkunit' },
  ],
  // ===== ruleManager =====
  agentStrategy: [
    { action: 'config_policy_view', resource_type: 'biz' },
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  proxyStrategy: [
    { action: 'config_policy_view', resource_type: 'biz' },
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  pluginStrategy: [
    { action: 'config_policy_view', resource_type: 'biz' },
    { action: 'config_policy_manage', resource_type: 'biz' },
  ],
  strategyTaskHistory: [
    { action: 'config_policy_history_view', resource_type: 'biz' },
  ],
  // ===== pkgManager =====
  agentPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  proxyPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  certPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  bintoolPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  plugin_bintoolPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  pluginPackageMng: [
    { action: 'package_view', resource_type: 'package' },
    { action: 'package_manage', resource_type: 'package' },
    { action: 'package_type_upload', resource_type: 'package_type' },
  ],
  operationRecords: [
    { action: 'package_history_view', resource_type: 'package' },
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
