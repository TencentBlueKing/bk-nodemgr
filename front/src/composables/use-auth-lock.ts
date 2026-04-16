import { computed, onUnmounted, ref } from 'vue';

import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

/**
 * 可复用的权限锁 hook
 * 提供 hover 锁图标跟随鼠标、点击申请权限的通用逻辑
 *
 * @param action IAM action 标识（如 'agent_operate'）
 * @param getResourceId 获取资源 ID 的函数（biz 类型传 bizId，非 biz 类型可传 undefined）
 * @param options 配置项
 */
export default function useAuthLock(
  action: string,
  getResourceId: () => number | string | Array<string | number> | undefined,
  options: {
    /** hover 锁图标水平偏移，默认 15 */
    offsetX?: number;
    /** hover 锁图标垂直偏移，默认 -5 */
    offsetY?: number;
    /** 资源类型，默认 'biz' */
    resourceType?: string;
  } = {},
) {
  const { offsetX = 15, offsetY = -5, resourceType = 'biz' } = options;
  const authStore = useAuthStore();
  const permissionStore = usePermissionStore();

  // ===== 锁光标状态 =====
  const lockCursorVisible = ref(false);
  const lockCursorX = ref(0);
  const lockCursorY = ref(0);

  let lockElement: HTMLDivElement | null = null;
  let suppressLock = false; // 点击后短暂禁止锁显示

  /** 创建跟随鼠标的锁图标 DOM 元素 */
  function ensureLockElement() {
    if (lockElement) return;
    lockElement = document.createElement('div');
    lockElement.className = 'auth-lock-cursor';
    lockElement.style.position = 'fixed';
    lockElement.style.zIndex = '999999';
    lockElement.style.pointerEvents = 'none';
    lockElement.style.width = '12px';
    lockElement.style.height = '16px';
    lockElement.style.display = 'none';
    // 使用 mask-image 实现锁图标，颜色由 backgroundColor 控制
    lockElement.style.maskImage = 'url(/images/lock.svg)';
    lockElement.style.maskSize = 'contain';
    lockElement.style.maskRepeat = 'no-repeat';
    lockElement.style.maskPosition = 'center';
    lockElement.style.webkitMaskImage = 'url(/images/lock.svg)';
    lockElement.style.webkitMaskSize = 'contain';
    lockElement.style.webkitMaskRepeat = 'no-repeat';
    lockElement.style.webkitMaskPosition = 'center';
    lockElement.style.backgroundColor = '#979ba5';
    document.body.appendChild(lockElement);
  }

  function removeLockElement() {
    if (lockElement) {
      lockElement.remove();
      lockElement = null;
    }
  }

  onUnmounted(() => {
    removeLockElement();
  });

  // ===== 权限判断 =====

  /** 判断当前资源是否有操作权限（computed，模板中可直接 v-if） */
  const hasAuth = computed(() => {
    // 直接读取 authorizedMap[action] 确保 Vue 响应式追踪该 key
    const entry = authStore.authorizedMap[action];
    if (!authStore.authorizedLoaded || !entry) return false;
    const resourceId = getResourceId();
    if (entry.isAny) return true;
    if (Array.isArray(resourceId)) {
      if (resourceId.length === 0) return true;
      return resourceId.every(id => entry.resourceIds.has(String(id)));
    }
    if (resourceId === undefined || resourceId === null) return entry.resourceIds.size > 0;
    return entry.resourceIds.has(String(resourceId));
  });

  // ===== 锁 hover 事件处理 =====

  /** mouseenter 事件：authorized=false 时显示锁 */
  function handleMouseEnter(e: MouseEvent, authorized: boolean) {
    if (suppressLock) return;
    if (!authorized) {
      ensureLockElement();
      if (lockElement) {
        lockElement.style.display = 'block';
        lockElement.style.left = `${e.clientX + offsetX}px`;
        lockElement.style.top = `${e.clientY + offsetY}px`;
      }
      lockCursorVisible.value = true;
      lockCursorX.value = e.clientX + offsetX;
      lockCursorY.value = e.clientY + offsetY;
    }
  }

  /** mousemove 事件：authorized=false 时锁跟随鼠标 */
  function handleMouseMove(e: MouseEvent, authorized: boolean) {
    if (!authorized && lockElement) {
      lockElement.style.left = `${e.clientX + offsetX}px`;
      lockElement.style.top = `${e.clientY + offsetY}px`;
      lockCursorX.value = e.clientX + offsetX;
      lockCursorY.value = e.clientY + offsetY;
    }
  }

  /** mouseleave 事件：隐藏锁 */
  function handleMouseLeave() {
    if (lockElement) {
      lockElement.style.display = 'none';
    }
    lockCursorVisible.value = false;
  }

  // ===== 点击申请权限 =====

  /** 点击无权限元素时触发权限申请 */
  async function handleAuthClick(e?: MouseEvent) {
    if (e) {
      e.stopPropagation();
      e.preventDefault();
    }
    // 点击后隐藏锁图标，并短暂禁止锁重新出现
    handleMouseLeave();
    suppressLock = true;
    setTimeout(() => { suppressLock = false; }, 500);
    const resourceId = getResourceId();
    const authItems = [
      { id: action, action, resourceType, routes: [] },
    ];
    await authStore.batchVerify(authItems, resourceType === 'biz' ? resourceId : undefined, resourceType !== 'biz' ? resourceId : undefined);
    const detail = authStore.permissionDetail;
    if (detail) {
      permissionStore.showDialog(detail);
    }
  }

  return {
    /** 当前是否有操作权限 */
    hasAuth,
    /** 锁光标是否可见 */
    lockCursorVisible,
    /** 锁光标 X 坐标 */
    lockCursorX,
    /** 锁光标 Y 坐标 */
    lockCursorY,
    /** mouseenter 事件处理，需传 authorized 参数 */
    handleMouseEnter,
    /** mousemove 事件处理，需传 authorized 参数 */
    handleMouseMove,
    /** mouseleave 事件处理 */
    handleMouseLeave,
    /** 点击无权限元素时触发权限申请 */
    handleAuthClick,
  };
}
