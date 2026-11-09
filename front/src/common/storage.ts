/**
 * 按「租户 + 登录用户」隔离的 localStorage 封装。
 *
 * 用户偏好类数据（选择的业务、收藏、表格设置、公告已读等）统一使用
 * `<tenant>-<username>-<key>` 作为真实存储 key，例如租户 default、用户 system
 * 的 `bk_biz_id` 实际存为 `default-system-bk_biz_id`。
 * 不同租户下用户名可能相同，必须带上租户前缀才能避免数据互相污染。
 *
 * 取不到用户名时退回原 key，行为与改造前一致。
 */
function getProjectUserIdentity(): { username: string; tenant: string } {
  try {
    const username = (window.PROJECT_CONFIG?.LOGIN_NAME || '').trim();
    const tenant = (window.PROJECT_CONFIG?.BK_TENANT || '').trim();

    return { username, tenant };
  } catch {
    return { username: '', tenant: '' };
  }
}

function getUserPrefix(): string {
  const { username, tenant } = getProjectUserIdentity();
  if (!username) return '';

  return tenant ? `${tenant}-${username}` : username;
}

export function userStorageKey(key: string): string {
  const prefix = getUserPrefix();
  return prefix ? `${prefix}-${key}` : key;
}

export const userStorage = {
  getItem(key: string): string | null {
    return localStorage.getItem(userStorageKey(key));
  },
  setItem(key: string, value: string): void {
    localStorage.setItem(userStorageKey(key), value);
  },
  removeItem(key: string): void {
    localStorage.removeItem(userStorageKey(key));
  },
};

/** 改造前直接以原始形式写入 localStorage 的业务 key 清单 */
const LEGACY_STORAGE_KEYS = [
  'bk_biz_id',
  'strategy_biz_id',
  'collect',
  'collect_workarea',
  'selected_workarea',
  'bk-dialog-notice-ids',
  'bk-table-default-settings',
  // 插件状态侧滑表格设置的旧版命名（新版为 nodeMng-*-status-sideslider）
  'frontend-plugin-status-sideslider',
] as const;

/**
 * 表格设置类自定义 storageKey 的模块前缀。
 * useTableSetting 各调用点按 `<模块前缀>-<动态后缀>` 传 key（如 pkgMng-agent、
 * topoMng-workarea-detail-xxx），后缀动态变化，只能按前缀匹配。
 */
const LEGACY_STORAGE_KEY_PREFIXES = ['pkgMng-', 'rulesMng-', 'topoMng-', 'nodeMng-'] as const;

/**
 * 一次性迁移改造前的历史数据：
 * - 用户 key 尚不存在时，把原始 key 的旧值搬到 `<tenant>-<username>-<key>`，避免升级后用户偏好丢失
 * - 上一版按 `<username>-<key>` 存的用户数据（无租户前缀）同样搬到新 key
 * - 搬运后删除原始 key，避免残留无主数据
 * - 同一浏览器多用户时，旧数据仅由首个登录的用户继承
 * - 取不到用户名时整体跳过，原始 key 保持不动（与未迁移时的行为一致）
 */
export function migrateLegacyStorage(): void {
  const { username, tenant } = getProjectUserIdentity();
  const prefix = getUserPrefix();
  if (!prefix) return;
  // 先收集再删除，避免边遍历边删导致漏项
  const migratePairs: Array<{ from: string; to: string }> = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key === null) continue;
    if (
      (LEGACY_STORAGE_KEYS as readonly string[]).includes(key)
      || LEGACY_STORAGE_KEY_PREFIXES.some(item => key.startsWith(item))
    ) {
      migratePairs.push({ from: key, to: userStorageKey(key) });
      continue;
    }
    // 上一版 <username>-<key> 格式：补上租户前缀，避免升级后丢失
    if (tenant && key.startsWith(`${username}-`)) {
      migratePairs.push({ from: key, to: `${prefix}-${key.slice(username.length + 1)}` });
    }
  }
  for (const { from, to } of migratePairs) {
    if (from === to) continue;
    const legacyValue = localStorage.getItem(from);
    if (legacyValue === null) continue;
    if (localStorage.getItem(to) === null) {
      localStorage.setItem(to, legacyValue);
    }
    localStorage.removeItem(from);
  }
}
