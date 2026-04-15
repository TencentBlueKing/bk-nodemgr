/**
 * G6 拓扑图无权限锁光标工具类
 * 在无权限节点 hover 时显示锁图标跟随鼠标，替代 useAuthLock（G6 节点无法使用 Vue composable）
 */

class AuthCursor {
  private cursorElement: HTMLDivElement | null = null;

  constructor() {
    this.createCursorElement();
  }

  private createCursorElement() {
    const el = document.createElement('div');
    el.className = 'g6-auth-cursor';
    el.style.cssText = `
      position: fixed;
      z-index: 999999;
      pointer-events: none;
      width: 12px;
      height: 16px;
      display: none;
      mask-image: url(/images/lock.svg);
      mask-size: contain;
      mask-repeat: no-repeat;
      mask-position: center;
      -webkit-mask-image: url(/images/lock.svg);
      -webkit-mask-size: contain;
      -webkit-mask-repeat: no-repeat;
      -webkit-mask-position: center;
      background-color: #979ba5;
    `;
    document.body.appendChild(el);
    this.cursorElement = el;
  }

  show(clientX: number, clientY: number) {
    if (!this.cursorElement) return;
    this.cursorElement.style.display = 'block';
    this.cursorElement.style.left = `${clientX + 15}px`;
    this.cursorElement.style.top = `${clientY - 5}px`;
  }

  move(clientX: number, clientY: number) {
    if (!this.cursorElement || this.cursorElement.style.display === 'none') return;
    this.cursorElement.style.left = `${clientX + 15}px`;
    this.cursorElement.style.top = `${clientY - 5}px`;
  }

  hide() {
    if (!this.cursorElement) return;
    this.cursorElement.style.display = 'none';
  }

  destroy() {
    if (this.cursorElement && this.cursorElement.parentNode) {
      this.cursorElement.parentNode.removeChild(this.cursorElement);
    }
    this.cursorElement = null;
  }
}

export const authCursor = new AuthCursor();
