import { Message } from 'bkui-vue';
import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

import type { AuthorizedItem, AuthorizedResult } from '@/@types/auth';
import { AuthService } from '@/api/modules/auth';
import Fetch from '@/api/fetch';
import type { AuthVerifyItem, AuthVerifyResp, PageAuthItem } from '@/constants/auth';
import { getSystemIdForResourceType } from '@/constants/auth';
import type { PermissionData } from '@/stores/permission';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

function isPermissionDeniedResponse(error: unknown): error is { permission: PermissionData } {
  return typeof error === 'object'
    && error !== null
    && 'permission' in error
    && typeof error.permission === 'object'
    && error.permission !== null;
}

export const useAuthStore = defineStore('auth', () => {
  const CACHE_EXPIRE_MS = 5 * 60 * 1000;
  const permissionMap = reactive<Record<string, boolean>>({});
  const permissionTimestampMap = reactive<Record<string, number>>({});
  const permissionDetail = ref<PermissionData | null>(null);
  const deniedActionIds = ref<string[]>([]);
  const needRefresh = ref(true);
  const loading = ref(false);
  const lastVerifiedBizScope = ref<string>('');
  // 并发 verify 请求序号：只让最新一次请求的响应生效，避免旧 tab（如 agent 历史）的
  // 鉴权结果覆盖新 tab（如 proxy 历史）的鉴权结果
  let verifySeq = 0;

  function normalizeBizScope(bizScope?: string | number | Array<string | number>): string {
    if (Array.isArray(bizScope)) {
      const normalized = bizScope
        .map(id => String(id))
        .filter(id => id !== '')
        .sort();
      return normalized.join(',');
    }
    return String(bizScope ?? '');
  }

  async function batchVerify(
    authItems: PageAuthItem[],
    bkBizScope?: string | number | Array<string | number>,
    resourceId?: string | number | Array<string | number>,
    options: { showPermissionDialog?: boolean } = {},
  ): Promise<boolean> {
    const seq = ++verifySeq;
    const bizScope = normalizeBizScope(bkBizScope);
    const bizResources = bizScope
      ? bizScope.split(',').map(id => ({
        system_id: getSystemIdForResourceType('biz'),
        type: 'biz',
        id,
      }))
      : [];

    const items: AuthVerifyItem[] = authItems.map((item) => {
      const verifyItem: AuthVerifyItem = {
        action: item.action,
      };
      if (!item.resourceType) {
        verifyItem.resources = [];
      }

      // For biz-scoped actions, include the biz resource
      if (item.resourceType === 'biz' && bizResources.length > 0) {
        verifyItem.resources = bizResources;
      } else if (item.resourceType && item.resourceType !== 'biz') {
        // For non-biz resource types (networkarea, networkunit, etc.)
        // 支持单个或多个资源实例（如操作记录页申请全部管控区域的查看权限）
        const resourceIds = Array.isArray(resourceId)
          ? resourceId
          : (resourceId !== undefined && resourceId !== null ? [resourceId] : []);
        if (resourceIds.length > 0) {
          verifyItem.resources = resourceIds.map(id => ({
            system_id: getSystemIdForResourceType(item.resourceType),
            type: item.resourceType,
            id: String(id),
          }));
        }
      }

      return verifyItem;
    });

    loading.value = true;
    try {
      const resp = await fetch.post<{ items: AuthVerifyItem[] }, AuthVerifyResp>('/api/v3/auth/verify')(
        { items },
        { irrevocable: true, interceptorErr: false, validateCode: false, needRes: true, ...options },
      ) as unknown as AuthVerifyResp;

      // 过期响应：期间已发起更新的 verify（如任务历史切 tab），本次结果直接丢弃
      if (seq !== verifySeq) return true;

      const data = (resp as any)?.data ?? resp;
      const now = Date.now();
      for (const item of authItems) {
        const cacheKey = `${item.action}:${bizScope}:${resourceId ?? ''}`;
        permissionMap[cacheKey] = true;
        permissionTimestampMap[cacheKey] = now;
      }
      deniedActionIds.value = [];
      // 只有带明细（actions）的 permission 才有效，避免各调用方弹出空白申请弹窗
      const successPermission = data?.permission as PermissionData | undefined;
      permissionDetail.value = successPermission?.actions?.length ? successPermission : null;
      lastVerifiedBizScope.value = bizScope;
      needRefresh.value = false;
      return true;
    } catch (error) {
      // 过期响应：期间已发起更新的 verify，本次结果直接丢弃
      if (seq !== verifySeq) return true;
      if (isPermissionDeniedResponse(error)) {
        const denied = authItems.map(item => item.action);
        const now = Date.now();

        for (const action of denied) {
          const cacheKey = `${action}:${bizScope}:${resourceId ?? ''}`;
          permissionMap[cacheKey] = false;
          permissionTimestampMap[cacheKey] = now;
        }

        // 后端的 permission 明细可能为空（无 actions），此时不算有效明细，
        // 否则调用方会弹出无内容的空白申请弹窗
        permissionDetail.value = error.permission?.actions?.length ? error.permission : null;
        deniedActionIds.value = denied;
        lastVerifiedBizScope.value = bizScope;
        needRefresh.value = false;

        return true;
      }

      // 非权限拒绝的真实异常（如 500 / 网络错误）：本请求关闭了全局错误拦截，这里按拦截器同款格式手动提示
      const errObj = (error ?? {}) as {
        code?: number | string;
        message?: string;
        data?: { code?: number | string };
        datas?: { message?: string };
        error?: { message?: string; details?: Array<{ message?: string }> };
      };
      const showMessage = errObj.message || errObj.datas?.message || errObj.error?.message || '';
      Message({
        theme: 'error',
        message: {
          code: errObj.data?.code ?? errObj.code ?? 500,
          overview: showMessage || '鉴权请求失败',
          suggestion: '',
          type: 'key-value',
          details: {
            code: errObj.data?.code ?? errObj.code,
            message: errObj.error?.details?.[0]?.message ?? showMessage,
          },
        },
      });

      permissionDetail.value = null;
      deniedActionIds.value = [];
      needRefresh.value = true;
      return false;
    } finally {
      loading.value = false;
    }
  }

  function hasPermission(actionId: string, bizScope?: string | number | Array<string | number>, resourceId?: string | number): boolean {
    return permissionMap[`${actionId}:${normalizeBizScope(bizScope)}:${resourceId ?? ''}`] ?? true;
  }

  function hasPermissionCache(actionId: string, bizScope?: string | number | Array<string | number>, resourceId?: string | number): boolean {
    // 注意：不能用 Object.prototype.hasOwnProperty.call(permissionMap, key)，
    // 它走 [[GetOwnProperty]]、不经过 Proxy trap，不会收集响应式依赖，
    // 导致依赖它的 computed（如 workarea 页 hasCreateAuth）首次短路后永不更新
    return permissionMap[`${actionId}:${normalizeBizScope(bizScope)}:${resourceId ?? ''}`] !== undefined;
  }

  function isPermissionCacheExpired(actionId: string, bizScope?: string | number | Array<string | number>, resourceId?: string | number): boolean {
    const cacheKey = `${actionId}:${normalizeBizScope(bizScope)}:${resourceId ?? ''}`;
    if (!Object.prototype.hasOwnProperty.call(permissionTimestampMap, cacheKey)) {
      return true;
    }
    return Date.now() - permissionTimestampMap[cacheKey] >= CACHE_EXPIRE_MS;
  }

  function getPermissionDetail(): PermissionData | null {
    return permissionDetail.value;
  }

  function getDeniedActionIds(): string[] {
    return deniedActionIds.value;
  }

  function setDeniedActionIds(actionIds: string[]) {
    deniedActionIds.value = actionIds;
  }

  function isDifferentBiz(currentBizScope?: string | number | Array<string | number>): boolean {
    return normalizeBizScope(currentBizScope) !== lastVerifiedBizScope.value;
  }

  function refreshPermissions() {
    deniedActionIds.value = [];
    needRefresh.value = true;
  }

  function reset() {
    Object.keys(permissionMap).forEach(key => delete permissionMap[key]);
    Object.keys(permissionTimestampMap).forEach(key => delete permissionTimestampMap[key]);
    permissionDetail.value = null;
    deniedActionIds.value = [];
    needRefresh.value = true;
    loading.value = false;
    lastVerifiedBizScope.value = '';
  }

  // ===== Authorized API =====
  // action → { isAny: boolean; resourceIds: Set<string> }
  const authorizedMap = reactive<Record<string, { isAny: boolean; resourceIds: Set<string> }>>({});
  const authorizedLoaded = ref(false);
  const authorizedLoading = ref(false);
  // 已加载过的模块集合，避免重复请求
  const loadedModules = reactive<Set<string>>(new Set());
  // 加载失败的模块集合，该模块的 action 默认视为有权限（降级）
  const failedModules = reactive<Set<string>>(new Set());
  // 当前正在进行的请求 Promise，按模块/key 独立追踪，不同模块互不阻塞
  const pendingRequests = new Map<string, Promise<void>>();

  /**
   * 调用 /api/v3/auth/authorized 获取当前用户对各 action 有权限的资源范围
   * @param items 要查询的 action-resource_type 对
   * @param moduleName 模块名（可选），用于标记已加载模块，避免重复请求
   * @param options.force 强制刷新：忽略模块级缓存（如权限申请通过后本地缓存已过期）
   */
  async function fetchAuthorized(
    items?: AuthorizedItem[],
    moduleName?: string,
    options: { force?: boolean } = {},
  ) {
    // 如果指定了模块且已加载过（成功或失败都算），跳过
    if (!options.force && moduleName && (loadedModules.has(moduleName) || failedModules.has(moduleName))) return;

    // 内容级去重：所需 action 都已有授权数据时跳过，避免内容相同的兄弟页面重复请求
    if (!options.force && items && items.length > 0 && items.every(item => authorizedMap[item.action] !== undefined)) return;

    // 用 moduleName 或 items 的 action 列表作为 key，相同 key 的请求共享 Promise
    const requestKey = moduleName || (items || []).map(i => i.action).sort().join(',') || '_default';

    // 如果同一 key 已有正在进行的请求，等待它完成
    const existing = pendingRequests.get(requestKey);
    if (existing) {
      await existing;
      return;
    }

    const request = doFetchAuthorized(items, moduleName);
    pendingRequests.set(requestKey, request);
    try {
      await request;
    } finally {
      pendingRequests.delete(requestKey);
    }
  }

  /** 实际执行请求的内部函数 */
  async function doFetchAuthorized(items?: AuthorizedItem[], moduleName?: string) {
    authorizedLoading.value = true;
    try {
      const res = await AuthService.Authorized({
        items: items || [],
      }, { irrevocable: true }); // 权限数据全局共享，不随路由切换取消（如任务历史 setup 的二次 replace 会触发 cancelRequest）

      const results: AuthorizedResult[] = (res as any)?.results || [];
      for (const r of results) {
        const resourceIds = new Set<string>();
        if (!r.is_any) {
          (r.resources || []).forEach(res => resourceIds.add(res.id));
        }
        authorizedMap[r.action] = { isAny: r.is_any, resourceIds };
      }
      if (moduleName) {
        loadedModules.add(moduleName);
      }
      authorizedLoaded.value = true;
    } catch (err) {
      console.error(`fetchAuthorized${moduleName ? ' [' + moduleName + ']' : ''} failed:`, err);
      // 失败时仍标记为已处理，防止无限重试；failedModules 记录用于降级判断
      if (moduleName) {
        failedModules.add(moduleName);
        // 降级：该模块所有 action 默认视为无权限
        (items || []).forEach(item => {
          authorizedMap[item.action] = { isAny: false, resourceIds: new Set() };
        });
      }
      authorizedLoaded.value = true;
    } finally {
      authorizedLoading.value = false;
    }
  }

  /**
   * 判断某个 action 下，指定资源是否有权限
   * @param action IAM action 标识
   * @param resourceId 资源 ID（如 biz_id, networkarea_id, networkunit_id）
   * @returns 有权限返回 true；未加载该 action 时返回 false
   */
  function hasAuthorizedResource(action: string, resourceId?: string | number): boolean {
    if (!action) return true;
    const entry = authorizedMap[action];
    if (!entry) return false; // 未查询过该 action，视为无权限（需先 fetchAuthorized）
    if (entry.isAny) return true;
    if (resourceId === undefined || resourceId === null) return entry.resourceIds.size > 0;
    return entry.resourceIds.has(String(resourceId));
  }

  /**
   * 判断某个 action 下，指定业务是否有权限（兼容旧调用）
   */
  function hasAuthorizedBiz(action: string, bizId?: string | number): boolean {
    return hasAuthorizedResource(action, bizId);
  }

  /**
   * 获取某个 action 下有权限的所有资源 ID
   * @param action IAM action 标识
   * @returns isAny=true 时返回 null（表示全部），否则返回资源 ID 数组
   */
  function getAuthorizedResourceIds(action: string): string[] | null {
    if (!authorizedLoaded.value) return null;
    const entry = authorizedMap[action];
    if (!entry || entry.isAny) return null;
    return Array.from(entry.resourceIds);
  }

  /**
   * 获取某个 action 下有权限的所有业务 ID（兼容旧调用）
   */
  function getAuthorizedBizIds(action: string): string[] | null {
    return getAuthorizedResourceIds(action);
  }

  return {
    permissionMap,
    permissionTimestampMap,
    permissionDetail,
    deniedActionIds,
    needRefresh,
    loading,
    lastVerifiedBizScope,
    authorizedMap,
    authorizedLoaded,
    authorizedLoading,
    batchVerify,
    hasPermission,
    hasPermissionCache,
    isPermissionCacheExpired,
    isDifferentBiz,
    getPermissionDetail,
    getDeniedActionIds,
    setDeniedActionIds,
    refreshPermissions,
    reset,
    fetchAuthorized,
    hasAuthorizedBiz,
    hasAuthorizedResource,
    getAuthorizedBizIds,
    getAuthorizedResourceIds,
  };
});
