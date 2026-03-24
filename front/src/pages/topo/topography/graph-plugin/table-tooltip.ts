/**
 * G6 Table Tooltip 工具类
 * 用于在接入点图标 hover 时显示 endpoints 表格
 * 学习 textTooltip 的实现方式，直接操作 DOM，避免 G6 节点事件冒泡导致的闪烁
 */
import { i18n } from '@/modules/i18n';

interface EndpointRow {
  name: string;
  endpoints: {
    cluster: string[];
    file: string[];
    data: string[];
  };
}

class TableTooltip {
  private tooltipElement: HTMLDivElement | null = null;
  private hideTimer: number | null = null;
  private hideInnerTimer: number | null = null;
  private showTimer: number | null = null;
  private _visible = false; // 用标志位而非 DOM 状态判断

  constructor() {
    this.createTooltipElement();
  }

  private createTooltipElement() {
    const tooltip = document.createElement('div');
    tooltip.className = 'g6-table-tooltip';
    tooltip.style.cssText = `
      position: fixed;
      z-index: 9999;
      padding: 12px 16px;
      background: #fff;
      border-radius: 6px;
      border: 1px solid #dcdee5;
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
      pointer-events: none;
      opacity: 0;
      transition: opacity 0.08s ease;
      display: none;
      max-width: 800px;
      min-width: 400px;
    `;

    // 箭头边框（朝左，tooltip 在右侧）
    const arrowBorder = document.createElement('div');
    arrowBorder.className = 'tooltip-arrow-border';
    arrowBorder.style.cssText = `
      position: absolute;
      left: -6px;
      top: 50%;
      transform: translateY(-50%);
      width: 0;
      height: 0;
      border-top: 6px solid transparent;
      border-bottom: 6px solid transparent;
      border-right: 6px solid #dcdee5;
    `;
    tooltip.appendChild(arrowBorder);

    // 箭头内层
    const arrow = document.createElement('div');
    arrow.className = 'tooltip-arrow';
    arrow.style.cssText = `
      position: absolute;
      left: -5px;
      top: 50%;
      transform: translateY(-50%);
      width: 0;
      height: 0;
      border-top: 5px solid transparent;
      border-bottom: 5px solid transparent;
      border-right: 5px solid #fff;
    `;
    tooltip.appendChild(arrow);

    document.body.appendChild(tooltip);
    this.tooltipElement = tooltip;
  }

  /**
   * 构建表格 HTML
   */
  private buildTableHTML(data: EndpointRow[]): string {
    const headerStyle = 'padding: 8px 12px; font-size: 12px; color: #979ba5; font-weight: normal; text-align: left; border-bottom: 1px solid #dcdee5; white-space: nowrap; background: #fafbfd;';
    const cellStyle = 'padding: 8px 12px; font-size: 12px; color: #63656e; border-bottom: 1px solid #f0f1f5; white-space: pre-line; word-break: break-all;';

    let rows = '';
    data.forEach((row) => {
      rows += `<tr>
        <td style="${cellStyle}">${this.escapeHTML(row.name)}</td>
        <td style="${cellStyle}">${this.escapeHTML(row.endpoints?.cluster?.join('\n') || '')}</td>
        <td style="${cellStyle}">${this.escapeHTML(row.endpoints?.file?.join('\n') || '')}</td>
        <td style="${cellStyle}">${this.escapeHTML(row.endpoints?.data?.join('\n') || '')}</td>
      </tr>`;
    });

    return `
      <div style="border: 1px solid #dcdee5; border-radius: 4px; overflow: hidden;">
        <table style="width: 100%; border-collapse: collapse; table-layout: auto;">
          <thead>
            <tr>
              <th style="${headerStyle}">${i18n.global.t('topoManager.topo.node.accessPoints')}</th>
              <th style="${headerStyle}">cluster</th>
              <th style="${headerStyle}">file</th>
              <th style="${headerStyle}">data</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    `;
  }

  private escapeHTML(str: string): string {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

  /**
   * 显示 tooltip
   */
  show(data: EndpointRow[], targetX: number, targetY: number, targetWidth: number, targetHeight: number) {
    if (!this.tooltipElement || !data || data.length === 0) return;

    // 彻底清除所有定时器，防止竞争
    this.clearAllTimers();

    // 标记为可见
    this._visible = true;

    // 设置表格内容
    const tableContainer = this.tooltipElement.querySelector('.table-content') as HTMLDivElement;
    if (tableContainer) {
      tableContainer.innerHTML = this.buildTableHTML(data);
    } else {
      const div = document.createElement('div');
      div.className = 'table-content';
      div.innerHTML = this.buildTableHTML(data);
      this.tooltipElement.insertBefore(div, this.tooltipElement.firstChild);
    }

    // 直接显示，不走任何延迟
    this.tooltipElement.style.display = 'block';

    requestAnimationFrame(() => {
      if (!this.tooltipElement || !this._visible) return;

        const rect = this.tooltipElement.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        const arrowWidth = 6;
        const gap = 12;

        // tooltip 在目标右侧
        const adjustedX = targetX + targetWidth + arrowWidth + gap;
        // 垂直居中对齐目标
        const targetCenterY = targetY + targetHeight / 2;
        let adjustedY = targetCenterY - rect.height / 2;

        // 垂直边界检查
        if (adjustedY < 10) {
          adjustedY = 10;
        } else if (adjustedY + rect.height > viewportHeight - 10) {
          adjustedY = viewportHeight - rect.height - 10;
        }

        // 计算箭头在气泡内的纵向位置，指向目标中心
        const arrowTop = targetCenterY - adjustedY;
        const clampedArrowTop = Math.max(12, Math.min(arrowTop, rect.height - 12));

        // 箭头朝左
        const arrow = this.tooltipElement.querySelector('.tooltip-arrow') as HTMLDivElement;
        const arrowBorder = this.tooltipElement.querySelector('.tooltip-arrow-border') as HTMLDivElement;

        if (arrow && arrowBorder) {
          // 重置所有方向
          arrowBorder.style.left = '-6px';
          arrowBorder.style.right = 'auto';
          arrowBorder.style.top = `${clampedArrowTop}px`;
          arrowBorder.style.bottom = 'auto';
          arrowBorder.style.transform = 'translateY(-50%)';
          arrowBorder.style.borderTop = '6px solid transparent';
          arrowBorder.style.borderBottom = '6px solid transparent';
          arrowBorder.style.borderRight = '6px solid #dcdee5';
          arrowBorder.style.borderLeft = 'none';

          arrow.style.left = '-5px';
          arrow.style.right = 'auto';
          arrow.style.top = `${clampedArrowTop}px`;
          arrow.style.bottom = 'auto';
          arrow.style.transform = 'translateY(-50%)';
          arrow.style.borderTop = '5px solid transparent';
          arrow.style.borderBottom = '5px solid transparent';
          arrow.style.borderRight = '5px solid #fff';
          arrow.style.borderLeft = 'none';
        }

        // 如果右侧空间不够，改为左侧展示
        if (adjustedX + rect.width > viewportWidth - 10) {
          const leftX = targetX - rect.width - arrowWidth - gap;
          this.tooltipElement.style.left = `${Math.max(10, leftX)}px`;

          // 箭头改为朝右
          if (arrow && arrowBorder) {
            arrowBorder.style.left = 'auto';
            arrowBorder.style.right = '-6px';
            arrowBorder.style.borderRight = 'none';
            arrowBorder.style.borderLeft = '6px solid #dcdee5';

            arrow.style.left = 'auto';
            arrow.style.right = '-5px';
            arrow.style.borderRight = 'none';
            arrow.style.borderLeft = '5px solid #fff';
          }
        } else {
          this.tooltipElement.style.left = `${adjustedX}px`;
        }

        this.tooltipElement.style.top = `${adjustedY}px`;
        this.tooltipElement.style.opacity = '1';
      });
  }

  /**
   * 清除所有定时器
   */
  private clearAllTimers() {
    if (this.hideTimer) {
      clearTimeout(this.hideTimer);
      this.hideTimer = null;
    }
    if (this.hideInnerTimer) {
      clearTimeout(this.hideInnerTimer);
      this.hideInnerTimer = null;
    }
    if (this.showTimer) {
      clearTimeout(this.showTimer);
      this.showTimer = null;
    }
  }

  /**
   * 隐藏 tooltip
   */
  hide(immediate = false) {
    if (!this.tooltipElement) return;

    if (immediate) {
      this.clearAllTimers();
      this._visible = false;
      this.tooltipElement.style.opacity = '0';
      this.tooltipElement.style.display = 'none';
      return;
    }

    // 延迟隐藏：给足够时间让下一个 show() 取消掉
    this.hideTimer = window.setTimeout(() => {
      if (this.tooltipElement && !this._visible) {
        // _visible 已被新的 show 设为 true 时不执行
        this.tooltipElement.style.opacity = '0';
        this.hideInnerTimer = window.setTimeout(() => {
          if (this.tooltipElement && !this._visible) {
            this.tooltipElement.style.display = 'none';
          }
        }, 80);
      }
    }, 150);

    // 先标记不可见，如果 show 在 150ms 内被调用会重置为 true
    this._visible = false;
  }

  destroy() {
    this.clearAllTimers();
    this._visible = false;
    if (this.tooltipElement && this.tooltipElement.parentNode) {
      this.tooltipElement.parentNode.removeChild(this.tooltipElement);
    }
    this.tooltipElement = null;
  }
}

export const tableTooltip = new TableTooltip();
