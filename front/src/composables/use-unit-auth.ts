import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

/**
 * 管控单元下拉选项的权限控制 hook
 * 提供：判断选项是否有权限、hover锁图标、点击无权限选项触发 verify 申请
 *
 * 使用方式与 biz-selector 中的权限逻辑一致：
 * - 无权限选项：灰色文字 + hover 锁图标跟随鼠标 + 点击阻止选中并触发权限申请
 * - 有权限选项：正常交互
 */
export default function useUnitAuth(action = 'networkunit_view') {
  const authStore = useAuthStore();
  const permissionStore = usePermissionStore();

  const {
    handleMouseEnter: authLockMouseEnter,
    handleMouseMove: authLockMouseMove,
    handleMouseLeave: authLockMouseLeave,
  } = useAuthLock(action, () => undefined, { resourceType: 'networkunit' });

  /** 判断某个管控单元是否有权限 */
  function isUnitAuthorized(unitId: number | string): boolean {
    if (!authStore.authorizedLoaded) return true; // 未加载完成时默认有权限，避免闪烁
    return authStore.hasAuthorizedResource(action, unitId);
  }

  /** 下拉选项 mouseenter：无权限时显示锁 */
  function handleOptionMouseEnter(e: MouseEvent, unitId: number | string) {
    authLockMouseEnter(e, isUnitAuthorized(unitId));
  }

  /** 下拉选项 mousemove：无权限时锁跟随鼠标 */
  function handleOptionMouseMove(e: MouseEvent, unitId: number | string) {
    authLockMouseMove(e, isUnitAuthorized(unitId));
  }

  /** 下拉选项 mouseleave：隐藏锁 */
  function handleOptionMouseLeave() {
    authLockMouseLeave();
  }

  /** 点击无权限选项：阻止选中并触发权限申请（verify 带 unitId） */
  async function handleOptionClick(e: MouseEvent, unitId: number | string) {
    if (!isUnitAuthorized(unitId)) {
      e.stopPropagation(); // 阻止冒泡到 Select.Option，防止被选中
      e.preventDefault();
      await handleApplyUnitPermission(unitId);
    }
  }

  /** 申请管控单元权限 */
  async function handleApplyUnitPermission(unitId: number | string) {
    const authItems = [
      { id: action, action, resourceType: 'networkunit', routes: [] },
    ];
    await authStore.batchVerify(authItems, undefined, unitId);
    const detail = authStore.permissionDetail;
    if (detail) {
      permissionStore.showDialog(detail);
    }
  }

  return {
    isUnitAuthorized,
    handleOptionMouseEnter,
    handleOptionMouseMove,
    handleOptionMouseLeave,
    handleOptionClick,
    handleApplyUnitPermission,
  };
}
