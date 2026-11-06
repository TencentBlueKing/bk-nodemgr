/**
 * 按登录用户隔离的 localStorage 封装。
 *
 * 用户偏好类数据（选择的业务、收藏、表格设置、公告已读等）统一使用
 * `<username>-<key>` 作为真实存储 key，例如用户 system 的 `bk_biz_id`
 * 实际存为 `system-bk_biz_id`，避免同一浏览器多用户切换时数据互相污染。
 *
 * 取不到用户名时退回原 key，行为与改造前一致。
 */
function getUserPrefix(): string {
  try {
    return (window.PROJECT_CONFIG?.LOGIN_NAME || '').trim();
  } catch {
    return '';
  }
}

export function userStorageKey(key: string): string {
  const username = getUserPrefix();
  return username ? `${username}-${key}` : key;
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
 * - 用户 key 尚不存在时，把原始 key 的旧值搬到 `<username>-<key>`，避免升级后用户偏好丢失
 * - 搬运后删除原始 key，避免残留无主数据
 * - 同一浏览器多用户时，旧数据仅由首个登录的用户继承
 * - 取不到用户名时整体跳过，原始 key 保持不动（与未迁移时的行为一致）
 */
export function migrateLegacyStorage(): void {
  if (!getUserPrefix()) return;
  // 先收集再删除，避免边遍历边删导致漏项
  const rawKeys: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key === null) continue;
    if (
      (LEGACY_STORAGE_KEYS as readonly string[]).includes(key)
      || LEGACY_STORAGE_KEY_PREFIXES.some(prefix => key.startsWith(prefix))
    ) {
      rawKeys.push(key);
    }
  }
  for (const key of rawKeys) {
    const legacyValue = localStorage.getItem(key);
    if (legacyValue === null) continue;
    const userKey = userStorageKey(key);
    if (localStorage.getItem(userKey) === null) {
      localStorage.setItem(userKey, legacyValue);
    }
    localStorage.removeItem(key);
  }
}
