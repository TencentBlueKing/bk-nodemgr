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
  ): Promise<boolean> {
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

      // For biz-scoped actions, include the biz resource
      if (item.resourceType === 'biz' && bizResources.length > 0) {
        verifyItem.resources = bizResources;
      }

      return verifyItem;
    });

    loading.value = true;
    try {
      const resp = await fetch.post<{ items: AuthVerifyItem[] }, AuthVerifyResp>('/api/v3/auth/verify')(
        { items },
        { interceptorErr: false, validateCode: false, needRes: true },
      ) as unknown as AuthVerifyResp;

      const data = (resp as any)?.data ?? resp;
      const now = Date.now();
      for (const item of authItems) {
        const cacheKey = `${item.action}:${bizScope}`;
        permissionMap[cacheKey] = true;
        permissionTimestampMap[cacheKey] = now;
      }
      deniedActionIds.value = [];
      if (data?.permission) {
        permissionDetail.value = data.permission as PermissionData;
      } else {
        permissionDetail.value = null;
      }
      lastVerifiedBizScope.value = bizScope;
      needRefresh.value = false;
      return true;
    } catch (error) {
      if (isPermissionDeniedResponse(error)) {
        const denied = authItems.map(item => item.action);
        const now = Date.now();

        for (const action of denied) {
          const cacheKey = `${action}:${bizScope}`;
          permissionMap[cacheKey] = false;
          permissionTimestampMap[cacheKey] = now;
        }

        permissionDetail.value = error.permission;
        deniedActionIds.value = denied;
        lastVerifiedBizScope.value = bizScope;
        needRefresh.value = false;

        return true;
      }

      permissionDetail.value = null;
      deniedActionIds.value = [];
      needRefresh.value = true;
      return false;
    } finally {
      loading.value = false;
    }
  }

  function hasPermission(actionId: string, bizScope?: string | number | Array<string | number>): boolean {
    return permissionMap[`${actionId}:${normalizeBizScope(bizScope)}`] ?? true;
  }

  function hasPermissionCache(actionId: string, bizScope?: string | number | Array<string | number>): boolean {
    return Object.prototype.hasOwnProperty.call(permissionMap, `${actionId}:${normalizeBizScope(bizScope)}`);
  }

  function isPermissionCacheExpired(actionId: string, bizScope?: string | number | Array<string | number>): boolean {
    const cacheKey = `${actionId}:${normalizeBizScope(bizScope)}`;
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
  // action → { isAny: boolean; bizIds: Set<string> }
  const authorizedMap = reactive<Record<string, { isAny: boolean; bizIds: Set<string> }>>({});
  const authorizedLoaded = ref(false);
  const authorizedLoading = ref(false);

  /**
   * 调用 /api/v3/auth/authorized 获取当前用户对各 action 有权限的业务范围
   * @param items 要查询的 action-resource_type 对，默认查所有 biz 类型的页面权限
   */
  async function fetchAuthorized(items?: AuthorizedItem[]) {
    if (authorizedLoading.value) return;
    authorizedLoading.value = true;

    const defaultItems: AuthorizedItem[] = [
      { action: 'agent_view', resource_type: 'biz' },
      { action: 'agent_operate', resource_type: 'biz' },
      { action: 'proxy_view', resource_type: 'biz' },
      { action: 'plugin_view', resource_type: 'biz' },
      { action: 'plugin_operate', resource_type: 'biz' },
      { action: 'agent_history_view', resource_type: 'biz' },
      { action: 'proxy_history_view', resource_type: 'biz' },
      { action: 'plugin_history_view', resource_type: 'biz' },
      { action: 'deploy_policy_view', resource_type: 'biz' },
      { action: 'config_policy_view', resource_type: 'biz' },
      { action: 'deploy_policy_history_view', resource_type: 'biz' },
    ];

    try {
      const res = await AuthService.Authorized({
        items: items || defaultItems,
      });

      const results: AuthorizedResult[] = (res as any)?.results || [];
      for (const r of results) {
        const bizIds = new Set<string>();
        if (!r.is_any) {
          (r.resources || []).forEach(res => bizIds.add(res.id));
        }
        authorizedMap[r.action] = { isAny: r.is_any, bizIds };
      }
      authorizedLoaded.value = true;
    } catch (err) {
      console.error('fetchAuthorized failed:', err);
    } finally {
      authorizedLoading.value = false;
    }
  }

  /**
   * 判断某个 action 下，指定业务是否有权限
   * @param action IAM action 标识
   * @param bizId 业务 ID
   * @returns 有权限返回 true；未加载完成时默认 false（显示锁图标），靠 authorized 接口返回后更新
   */
  function hasAuthorizedBiz(action: string, bizId?: string | number): boolean {
    if (!action) return true; // 无 action 匹配时不做权限限制
    if (!authorizedLoaded.value) return false; // 未加载完成时默认无权限，显示锁图标
    const entry = authorizedMap[action];
    if (!entry) return true; // 不在查询列表中的 action，默认放行
    if (entry.isAny) return true;
    if (bizId === undefined || bizId === null) return entry.bizIds.size > 0;
    return entry.bizIds.has(String(bizId));
  }

  /**
   * 获取某个 action 下有权限的所有业务 ID
   * @param action IAM action 标识
   * @returns isAny=true 时返回 null（表示全部），否则返回 bizId 数组
   */
  function getAuthorizedBizIds(action: string): string[] | null {
    if (!authorizedLoaded.value) return null;
    const entry = authorizedMap[action];
    if (!entry || entry.isAny) return null;
    return Array.from(entry.bizIds);
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
    getAuthorizedBizIds,
  };
});
