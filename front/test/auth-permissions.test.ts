import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { matchPageAuth, PAGE_AUTH_CONFIG, shouldDeferBizAuthCheck } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';

const { verifyRequestMock, postMock } = vi.hoisted(() => {
  const verifyRequestMock = vi.fn();
  const postMock = vi.fn(() => verifyRequestMock);

  return { verifyRequestMock, postMock };
});

vi.mock('@/api/fetch', () => ({
  default: class MockFetch {
    post = postMock;
  },
}));

describe('auth permission helpers', () => {
  it('matches history permissions by active tab', () => {
    expect(matchPageAuth({ name: 'history', query: { active: 'agent' } }, PAGE_AUTH_CONFIG)?.id).toBe('agent_history_view');
    expect(matchPageAuth({ name: 'taskDetail', query: { active: 'proxy' } }, PAGE_AUTH_CONFIG)?.id).toBe('proxy_history_view');
    expect(matchPageAuth({ name: 'log', query: { active: 'plugin' } }, PAGE_AUTH_CONFIG)?.id).toBe('plugin_history_view');
  });

  it('falls back to agent history permission when active tab is missing', () => {
    expect(matchPageAuth({ name: 'history', query: {} }, PAGE_AUTH_CONFIG)?.id).toBe('agent_history_view');
  });

  it('defers biz auth until business context is ready', () => {
    const matched = matchPageAuth({ name: 'agent' }, PAGE_AUTH_CONFIG);

    expect(shouldDeferBizAuthCheck(matched, false, undefined)).toBe(true);
    expect(shouldDeferBizAuthCheck(matched, true, undefined)).toBe(true);
    expect(shouldDeferBizAuthCheck(matched, true, 2)).toBe(false);
  });
});

describe('auth store batch verify', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    verifyRequestMock.mockReset();
    postMock.mockClear();
  });

  it('treats permission denied as a known unauthorized result', async () => {
    const matched = PAGE_AUTH_CONFIG[0];
    verifyRequestMock.mockRejectedValueOnce({
      permission: {
        system: 'bk_nodemgr',
        system_name: 'NodeMan',
        apply_url: '/apply',
        actions: [{
          id: 'agent_view',
          name: 'Agent View',
          related_resource_types: [],
        }],
      },
    });

    const authStore = useAuthStore();
    const success = await authStore.batchVerify([matched], '2');

    expect(success).toBe(true);
    expect(authStore.needRefresh).toBe(false);
    expect(authStore.hasPermission('agent_view', 2)).toBe(false);
    expect(authStore.getDeniedActionIds()).toEqual(['agent_view']);
    expect(authStore.getPermissionDetail()).toEqual({
      system: 'bk_nodemgr',
      system_name: 'NodeMan',
      apply_url: '/apply',
      actions: [{
        id: 'agent_view',
        name: 'Agent View',
        related_resource_types: [],
      }],
    });
  });

  it('does not downgrade unexpected verify failures into denied permissions', async () => {
    verifyRequestMock.mockRejectedValueOnce(new Error('iam backend unavailable'));

    const authStore = useAuthStore();
    const success = await authStore.batchVerify(PAGE_AUTH_CONFIG, '2');

    expect(success).toBe(false);
    expect(authStore.needRefresh).toBe(true);
    expect(authStore.hasPermission('agent_view', 2)).toBe(true);
    expect(authStore.getPermissionDetail()).toBeNull();
  });
});
