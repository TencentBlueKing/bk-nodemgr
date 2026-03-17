/**
 * G6 Edge Tooltip 工具类
 * 用于在 hover 边时展示连接信息（结构化展示 link 类型 + 路径）
 */

interface EdgeTooltipData {
  linkTypes: string[];
  sourceName: string;   // 起始节点名称
  targetName: string;   // 目标节点名称（含路径，如 "单元B/接入点C"）
}

class EdgeTooltip {
  private tooltipElement: HTMLDivElement | null = null;
  private hideTimer: number | null = null;
  private showTimer: number | null = null;

  constructor() {
    this.createTooltipElement();
  }

  /**
   * 创建 tooltip 元素
   */
  private createTooltipElement() {
    const tooltip = document.createElement('div');
    tooltip.className = 'g6-edge-tooltip';
    tooltip.style.cssText = `
      position: fixed;
      z-index: 9999;
      padding: 8px 14px;
      background: #fff;
      color: #333;
      font-size: 12px;
      line-height: 1;
      border-radius: 6px;
      border: 1px solid #e6e8ed;
      box-shadow: 0 3px 12px rgba(0, 0, 0, 0.1);
      pointer-events: none;
      white-space: nowrap;
      width: fit-content;
      opacity: 0;
      transition: opacity 0.15s ease;
      display: none;
    `;

    document.body.appendChild(tooltip);
    this.tooltipElement = tooltip;
  }

  /**
   * 构建结构化 HTML 内容
   */
  private buildHTML(data: EdgeTooltipData): string {
    // 类型标签
    const typeTags = data.linkTypes.map((type) => {
      return `<span style="
        display: inline-block;
        padding: 2px 6px;
        border-radius: 3px;
        background: #f0f1f5;
        color: #63656e;
        font-size: 11px;
        font-weight: 500;
        line-height: 16px;
      ">${type}</span>`;
    }).join('<span style="width: 4px; display: inline-block;"></span>');

    // 箭头 SVG（简洁的右箭头）
    const arrowSvg = `<svg width="16" height="12" viewBox="0 0 16 12" fill="none" style="vertical-align: middle; margin: 0 2px;">
      <path d="M0 6h13M10 2l4 4-4 4" stroke="#c4c6cc" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>`;

    return `<span style="display: inline-flex; align-items: center; gap: 0px;">
      ${typeTags}
      <span style="display: inline-block; width: 12px;"></span>
      <span style="color: #63656e; font-size: 12px; font-weight: 500;">${this.escapeHtml(data.sourceName)}</span>
      <span style="display: inline-flex; align-items: center; margin: 0 6px;">${arrowSvg}</span>
      <span style="color: #63656e; font-size: 12px; font-weight: 500;">${this.escapeHtml(data.targetName)}</span>
    </span>`;
  }

  /**
   * 转义 HTML 特殊字符
   */
  private escapeHtml(str: string): string {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

  /**
   * 显示 tooltip（结构化数据）
   */
  show(data: EdgeTooltipData, clientX: number, clientY: number) {
    if (!this.tooltipElement || !data) return;

    // 清除之前的定时器
    if (this.hideTimer) {
      clearTimeout(this.hideTimer);
      this.hideTimer = null;
    }
    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = null;
    }

    this.showTimer = window.setTimeout(() => {
      if (!this.tooltipElement) return;

      // 设置 HTML 内容
      this.tooltipElement.innerHTML = this.buildHTML(data);

      // 显示并计算位置
      this.tooltipElement.style.display = 'block';
      this.tooltipElement.style.opacity = '0';

      requestAnimationFrame(() => {
        if (!this.tooltipElement) return;

        const rect = this.tooltipElement.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        // tooltip 显示在鼠标上方偏移
        let adjustedX = clientX - rect.width / 2;
        let adjustedY = clientY - rect.height - 10;

        // 边界检查
        if (adjustedX < 5) adjustedX = 5;
        if (adjustedX + rect.width > viewportWidth - 5) {
          adjustedX = viewportWidth - rect.width - 5;
        }
        if (adjustedY < 5) {
          adjustedY = clientY + 15;
        }
        if (adjustedY + rect.height > viewportHeight - 5) {
          adjustedY = viewportHeight - rect.height - 5;
        }

        this.tooltipElement.style.left = `${adjustedX}px`;
        this.tooltipElement.style.top = `${adjustedY}px`;
        this.tooltipElement.style.opacity = '1';
      });
    }, 200);
  }

  /**
   * 更新位置（跟随鼠标）
   */
  updatePosition(clientX: number, clientY: number) {
    if (!this.tooltipElement || this.tooltipElement.style.display === 'none') return;

    const rect = this.tooltipElement.getBoundingClientRect();
    const viewportWidth = window.innerWidth;

    let adjustedX = clientX - rect.width / 2;
    const adjustedY = clientY - rect.height - 10;

    if (adjustedX < 5) adjustedX = 5;
    if (adjustedX + rect.width > viewportWidth - 5) {
      adjustedX = viewportWidth - rect.width - 5;
    }

    this.tooltipElement.style.left = `${adjustedX}px`;
    this.tooltipElement.style.top = `${adjustedY}px`;
  }

  /**
   * 隐藏 tooltip
   * @param immediate 是否立即隐藏
   */
  hide(immediate = false) {
    if (!this.tooltipElement) return;

    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = null;
    }

    if (immediate) {
      this.tooltipElement.style.opacity = '0';
      this.tooltipElement.style.display = 'none';
      return;
    }

    this.hideTimer = window.setTimeout(() => {
      if (this.tooltipElement) {
        this.tooltipElement.style.opacity = '0';
        setTimeout(() => {
          if (this.tooltipElement) {
            this.tooltipElement.style.display = 'none';
          }
        }, 150);
      }
    }, 100);
  }

  /**
   * 销毁 tooltip
   */
  destroy() {
    if (this.hideTimer) clearTimeout(this.hideTimer);
    if (this.showTimer) clearTimeout(this.showTimer);
    if (this.tooltipElement && this.tooltipElement.parentNode) {
      this.tooltipElement.parentNode.removeChild(this.tooltipElement);
    }
    this.tooltipElement = null;
  }
}

// 创建全局单例
export const edgeTooltip = new EdgeTooltip();
