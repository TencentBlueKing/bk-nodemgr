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
    this.bindGlobalMouseTracker();
  }

  /**
   * 全局追踪鼠标位置，用于 hide() 时判断鼠标是否在 tooltip 上
   */
  private bindGlobalMouseTracker() {
    document.addEventListener('mousemove', (e: MouseEvent) => {
      (window as any).__tooltipMouseX = e.clientX;
      (window as any).__tooltipMouseY = e.clientY;
    });
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
      white-space: nowrap;
      min-width: 200px;
      max-width: 600px;
      max-height: 60vh;
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

    // 鼠标悬停 tooltip 时取消隐藏，离开时延迟隐藏
    tooltip.addEventListener('mouseenter', () => {
      if (this.hideTimer) {
        clearTimeout(this.hideTimer);
        this.hideTimer = null;
      }
    });
    tooltip.addEventListener('mouseleave', () => {
      this.hide();
    });

    document.body.appendChild(tooltip);
    this.tooltipElement = tooltip;
  }

  /**
   * 显示 tooltip（纯文本）
   * @param content 提示内容
   * @param targetX 目标元素起始 X 坐标
   * @param targetY 目标元素起始 Y 坐标
   * @param targetWidth 目标元素宽度
   * @param targetHeight 目标元素高度
   * @param minWidth tooltip 最小宽度，默认 200px
   */
  show(content: string, targetX: number, targetY: number, targetWidth: number, targetHeight: number, minWidth = 200) {
    this._show(content, targetX, targetY, targetWidth, targetHeight, minWidth, false);
  }

  /**
   * 显示 tooltip（HTML 内容，支持彩色标记）
   * @param htmlContent HTML 内容字符串
   * @param targetX 目标元素起始 X 坐标
   * @param targetY 目标元素起始 Y 坐标
   * @param targetWidth 目标元素宽度
   * @param targetHeight 目标元素高度
   * @param minWidth tooltip 最小宽度，默认 200px
   */
  showHTML(htmlContent: string, targetX: number, targetY: number, targetWidth: number, targetHeight: number, minWidth = 200) {
    this._show(htmlContent, targetX, targetY, targetWidth, targetHeight, minWidth, true);
  }

  private _show(content: string, targetX: number, targetY: number, targetWidth: number, targetHeight: number, minWidth = 200, isHTML = false) {
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
      if (isHTML) {
        // HTML 模式：保留箭头元素，替换其余内容
        this.tooltipElement.innerHTML = content;
        // 确保箭头元素存在（重新添加，因为 innerHTML 清空了）
        if (!this.tooltipElement.querySelector('.tooltip-arrow')) {
          const arrowBorder = document.createElement('div');
          arrowBorder.className = 'tooltip-arrow-border';
          arrowBorder.style.cssText = `
            position: absolute; left: 50%; bottom: -6px; transform: translateX(-50%);
            width: 0; height: 0;
            border-left: 6px solid transparent; border-right: 6px solid transparent;
            border-top: 6px solid #dcdee5;
          `;
          const arrow = document.createElement('div');
          arrow.className = 'tooltip-arrow';
          arrow.style.cssText = `
            position: absolute; left: 50%; bottom: -5px; transform: translateX(-50%);
            width: 0; height: 0;
            border-left: 5px solid transparent; border-right: 5px solid transparent;
            border-top: 5px solid #fff;
          `;
          this.tooltipElement.appendChild(arrowBorder);
          this.tooltipElement.appendChild(arrow);
        }
      } else {
        // 纯文本模式
        const textNode = document.createTextNode(content);
        this.tooltipElement.childNodes.forEach((node, index) => {
          if (index === 0 || node.nodeType === Node.TEXT_NODE) {
            this.tooltipElement?.removeChild(node);
          }
        });
        this.tooltipElement.insertBefore(textNode, this.tooltipElement.firstChild);
      }

      // 显示并计算位置
      this.tooltipElement.style.display = 'block';
      this.tooltipElement.style.opacity = '0';

      // 等待一帧后计算位置和显示
      requestAnimationFrame(() => {
        if (!this.tooltipElement) return;

        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        // 箭头高度
        const arrowHeight = 6;
        // tooltip 与目标元素的间距
        const gap = 16;

        // 获取箭头元素
        const arrow = this.tooltipElement.querySelector('.tooltip-arrow') as HTMLDivElement;
        const arrowBorder = this.tooltipElement.querySelector('.tooltip-arrow-border') as HTMLDivElement;

        // ---- 第一步：测量内容自然高度 ----
        this.tooltipElement.style.maxHeight = 'none';
        this.tooltipElement.style.overflowY = 'visible';
        const naturalHeight = this.tooltipElement.getBoundingClientRect().height;

        // ---- 第二步：根据视口约束决定 max-height + overflow ----
        const cssMaxHeightPx = (60 * viewportHeight) / 100;
        const topGap = 10; // 视口顶部留白
        const bottomGap = 10; // 视口底部留白

        let finalMaxHeight = cssMaxHeightPx;
        let needScroll = naturalHeight > cssMaxHeightPx;
        let placedBelow = false;
        let adjustedY = 0;

        // 尝试放在目标上方
        const aboveY = targetY - naturalHeight - arrowHeight - gap;
        if (aboveY >= topGap) {
          // 上方空间足够，放在上方
          adjustedY = aboveY;
          if (needScroll) {
            this.tooltipElement.style.overflowY = 'auto';
          }
        } else {
          // 上方空间不够，检查下方
          const belowY = targetY + targetHeight + arrowHeight + gap;
          const bottomAvailable = viewportHeight - bottomGap - belowY;
          if (bottomAvailable >= Math.min(naturalHeight, cssMaxHeightPx)) {
            // 下方空间足够
            adjustedY = belowY;
            placedBelow = true;
            if (needScroll) {
              this.tooltipElement.style.overflowY = 'auto';
            }
          } else if (bottomAvailable > 60) {
            // 下方有部分空间，限制高度 + 滚动
            adjustedY = belowY;
            placedBelow = true;
            finalMaxHeight = bottomAvailable;
            needScroll = true;
            this.tooltipElement.style.overflowY = 'auto';
          } else {
            // 下方空间也很少，放在上方并限制高度
            const topAvailable = targetY - arrowHeight - gap - topGap;
            adjustedY = topGap;
            if (topAvailable > 60) {
              finalMaxHeight = topAvailable;
              needScroll = true;
              this.tooltipElement.style.overflowY = 'auto';
            } else {
              finalMaxHeight = Math.max(40, topAvailable);
              needScroll = true;
              this.tooltipElement.style.overflowY = 'auto';
            }
          }
        }

        this.tooltipElement.style.maxHeight = `${finalMaxHeight}px`;

        // ---- 第三步：用最终尺寸重新测量做水平定位 ----
        const rect = this.tooltipElement.getBoundingClientRect();

        // 计算目标元素的中心点 X 坐标
        const targetCenterX = targetX + targetWidth / 2;
        let adjustedX = targetCenterX - rect.width / 2;
        
        // 设置箭头方向
        if (arrow && arrowBorder) {
          if (placedBelow) {
            // tooltip 在目标下方，箭头朝上
            arrowBorder.style.top = '-6px';
            arrowBorder.style.bottom = 'auto';
            arrowBorder.style.borderBottom = '6px solid #dcdee5';
            arrowBorder.style.borderTop = 'none';
            arrowBorder.style.borderLeft = '6px solid transparent';
            arrowBorder.style.borderRight = '6px solid transparent';
            
            arrow.style.top = '-5px';
            arrow.style.bottom = 'auto';
            arrow.style.borderBottom = '5px solid #fff';
            arrow.style.borderTop = 'none';
            arrow.style.borderLeft = '5px solid transparent';
            arrow.style.borderRight = '5px solid transparent';
          } else {
            // tooltip 在目标上方，箭头朝下
            arrowBorder.style.bottom = '-6px';
            arrowBorder.style.top = 'auto';
            arrowBorder.style.borderTop = '6px solid #dcdee5';
            arrowBorder.style.borderBottom = 'none';
            arrowBorder.style.borderLeft = '6px solid transparent';
            arrowBorder.style.borderRight = '6px solid transparent';
            
            arrow.style.bottom = '-5px';
            arrow.style.top = 'auto';
            arrow.style.borderTop = '5px solid #fff';
            arrow.style.borderBottom = 'none';
            arrow.style.borderLeft = '5px solid transparent';
            arrow.style.borderRight = '5px solid transparent';
          }
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
      if (!this.tooltipElement) return;

      // 如果鼠标当前在 tooltip 上，不隐藏
      const tooltipRect = this.tooltipElement.getBoundingClientRect();
      const mouseX = (window as any).__tooltipMouseX ?? -1;
      const mouseY = (window as any).__tooltipMouseY ?? -1;
      if (
        mouseX >= tooltipRect.left && mouseX <= tooltipRect.right
        && mouseY >= tooltipRect.top && mouseY <= tooltipRect.bottom
      ) {
        return;
      }

      this.tooltipElement.style.opacity = '0';
      setTimeout(() => {
        if (this.tooltipElement) {
          this.tooltipElement.style.display = 'none';
        }
      }, 200);
    }, 150);
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
