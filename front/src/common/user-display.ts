/**
 * 用户名展示翻译工具（复用 bk-user-display-name 的 ts 翻译能力）
 * @author hyfyahuang
 */
import BkUserDisplayName from '@blueking/bk-user-display-name';

/**
 * 批量获取用户名的 display_name
 * @param usernames 用户名（bk_username）数组
 * @returns key 为 username，value 为 display_name；获取失败时返回空 Map
 */
export const batchFetchUserDisplayNames = async (
  usernames: string[],
): Promise<Map<string, string>> => {
  const result = new Map<string, string>();
  const unique = Array.from(new Set(usernames.filter(Boolean)));
  if (!unique.length) return result;
  try {
    const inst = new BkUserDisplayName();
    const resp: any = await inst.fetchUsers(unique);
    ((resp?.data as any[]) || []).forEach((u: any) => {
      if (u?.bk_username && u?.display_name) {
        result.set(u.bk_username, u.display_name);
      }
    });
  } catch {
    // 翻译失败时静默返回空 Map，调用方继续使用原值
  }
  return result;
};

/**
 * 翻译一组筛选项里的用户名显示（兼容 { text, value } 和 { name, id } 两种结构）
 * @param items 筛选项数组（必须含 value 或 id 字段，可含 text 或 name 字段）
 */
export const translateOperatorItems = async (
  items: { text?: string; name?: string; value?: string | number | boolean; id?: string | number | boolean }[],
): Promise<void> => {
  const map = await batchFetchUserDisplayNames(
    items.map((i) => String(i.value ?? i.id ?? '')),
  );
  items.forEach((i) => {
    const k = String(i.value ?? i.id ?? '');
    const display = map.get(k);
    if (!display) return;
    if (i.text !== undefined) i.text = display;
    if (i.name !== undefined) i.name = display;
  });
};