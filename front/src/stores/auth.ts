import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

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

  return {
    permissionMap,
    permissionTimestampMap,
    permissionDetail,
    deniedActionIds,
    needRefresh,
    loading,
    lastVerifiedBizScope,
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
  };
});
