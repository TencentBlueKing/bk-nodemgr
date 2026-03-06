/**
 * G6 文本 Tooltip 工具类
 * 用于在节点文本被截断时显示完整内容
 */

class TextTooltip {
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
    tooltip.className = 'g6-text-tooltip';
    tooltip.style.cssText = `
      position: fixed;
      z-index: 9999;
      padding: 6px 12px;
      background: #fff;
      color: #333;
      font-size: 12px;
      line-height: 1.5;
      border-radius: 4px;
      border: 1px solid #dcdee5;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      pointer-events: none;
      white-space: nowrap;
      min-width: 200px;
      max-width: 600px;
      word-break: break-all;
      opacity: 0;
      transition: opacity 0.2s ease;
      display: none;
    `;

    // 添加箭头边框（外层，灰色）
    const arrowBorder = document.createElement('div');
    arrowBorder.className = 'tooltip-arrow-border';
    arrowBorder.style.cssText = `
      position: absolute;
      left: 50%;
      bottom: -6px;
      transform: translateX(-50%);
      width: 0;
      height: 0;
      border-left: 6px solid transparent;
      border-right: 6px solid transparent;
      border-top: 6px solid #dcdee5;
    `;
    tooltip.appendChild(arrowBorder);

    // 添加小箭头（内层，白色）- 必须在边框之后添加，这样白色箭头会覆盖在边框上
    const arrow = document.createElement('div');
    arrow.className = 'tooltip-arrow';
    arrow.style.cssText = `
      position: absolute;
      left: 50%;
      bottom: -5px;
      transform: translateX(-50%);
      width: 0;
      height: 0;
      border-left: 5px solid transparent;
      border-right: 5px solid transparent;
      border-top: 5px solid #fff;
    `;
    tooltip.appendChild(arrow);

    document.body.appendChild(tooltip);
    this.tooltipElement = tooltip;
  }

  /**
   * 显示 tooltip
   * @param content 提示内容
   * @param targetX 目标元素起始 X 坐标
   * @param targetY 目标元素起始 Y 坐标
   * @param targetWidth 目标元素宽度
   * @param targetHeight 目标元素高度
   * @param minWidth tooltip 最小宽度，默认 200px
   */
  show(content: string, targetX: number, targetY: number, targetWidth: number, targetHeight: number, minWidth = 200) {
    if (!this.tooltipElement || !content) return;

    // 清除之前的定时器
    if (this.hideTimer) {
      clearTimeout(this.hideTimer);
      this.hideTimer = null;
    }
    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = null;
    }

    // 延迟显示，避免快速划过时频繁触发
    this.showTimer = window.setTimeout(() => {
      if (!this.tooltipElement) return;

      // 设置最小宽度
      this.tooltipElement.style.minWidth = `${minWidth}px`;

      // 设置内容
      const textNode = document.createTextNode(content);
      this.tooltipElement.childNodes.forEach((node, index) => {
        if (index === 0 || node.nodeType === Node.TEXT_NODE) {
          this.tooltipElement?.removeChild(node);
        }
      });
      this.tooltipElement.insertBefore(textNode, this.tooltipElement.firstChild);

      // 显示并计算位置
      this.tooltipElement.style.display = 'block';
      this.tooltipElement.style.opacity = '0';

      // 等待一帧后计算位置和显示
      requestAnimationFrame(() => {
        if (!this.tooltipElement) return;

        const rect = this.tooltipElement.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        
        // 箭头高度
        const arrowHeight = 6;
        // tooltip 与目标元素的间距
        const gap = 16;

        // 计算目标元素的中心点 X 坐标
        const targetCenterX = targetX + targetWidth / 2;
        
        // tooltip 水平居中对齐目标元素
        let adjustedX = targetCenterX - rect.width / 2;
        
        // 显示在目标元素上方，包含箭头高度和间距
        const adjustedY = targetY - rect.height - arrowHeight - gap;
        
        // 获取箭头元素，设置为朝下
        const arrow = this.tooltipElement.querySelector('.tooltip-arrow') as HTMLDivElement;
        const arrowBorder = this.tooltipElement.querySelector('.tooltip-arrow-border') as HTMLDivElement;
        
        if (arrow && arrowBorder) {
          // 箭头边框（外层，灰色）
          arrowBorder.style.bottom = '-6px';
          arrowBorder.style.top = 'auto';
          arrowBorder.style.borderTop = '6px solid #dcdee5';
          arrowBorder.style.borderBottom = 'none';
          arrowBorder.style.borderLeft = '6px solid transparent';
          arrowBorder.style.borderRight = '6px solid transparent';
          
          // 箭头（内层，白色）
          arrow.style.bottom = '-5px';
          arrow.style.top = 'auto';
          arrow.style.borderTop = '5px solid #fff';
          arrow.style.borderBottom = 'none';
          arrow.style.borderLeft = '5px solid transparent';
          arrow.style.borderRight = '5px solid transparent';
        }

        // 水平方向边界检查
        if (adjustedX < 10) {
          adjustedX = 10;
        } else if (adjustedX + rect.width > viewportWidth - 10) {
          adjustedX = viewportWidth - rect.width - 10;
        }

        this.tooltipElement.style.left = `${adjustedX}px`;
        this.tooltipElement.style.top = `${adjustedY}px`;
        this.tooltipElement.style.opacity = '0.9';
      });
    }, 300); // 延迟 300ms 显示
  }

  /**
   * 隐藏 tooltip
   * @param immediate 是否立即隐藏（不使用延迟）
   */
  hide(immediate = false) {
    if (!this.tooltipElement) return;

    // 清除显示定时器
    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = null;
    }

    if (immediate) {
      this.tooltipElement.style.opacity = '0';
      this.tooltipElement.style.display = 'none';
      return;
    }

    // 延迟隐藏，避免闪烁
    this.hideTimer = window.setTimeout(() => {
      if (this.tooltipElement) {
        this.tooltipElement.style.opacity = '0';
        setTimeout(() => {
          if (this.tooltipElement) {
            this.tooltipElement.style.display = 'none';
          }
        }, 200);
      }
    }, 300);
  }

  /**
   * 销毁 tooltip
   */
  destroy() {
    if (this.hideTimer) {
      clearTimeout(this.hideTimer);
    }
    if (this.showTimer) {
      clearTimeout(this.showTimer);
    }
    if (this.tooltipElement && this.tooltipElement.parentNode) {
      this.tooltipElement.parentNode.removeChild(this.tooltipElement);
    }
    this.tooltipElement = null;
  }
}

// 创建全局单例
export const textTooltip = new TextTooltip();
