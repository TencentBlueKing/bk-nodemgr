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
  private showTimer: number | null = null;

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
      transition: opacity 0.2s ease;
      display: none;
      max-width: 800px;
      min-width: 400px;
    `;

    // 箭头边框
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

    // 箭头内层
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

    // 清除定时器
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

      // 设置表格内容
      const tableContainer = this.tooltipElement.querySelector('.table-content') as HTMLDivElement;
      if (tableContainer) {
        tableContainer.innerHTML = this.buildTableHTML(data);
      } else {
        const div = document.createElement('div');
        div.className = 'table-content';
        div.innerHTML = this.buildTableHTML(data);
        // 插到箭头之前
        this.tooltipElement.insertBefore(div, this.tooltipElement.firstChild);
      }

      this.tooltipElement.style.display = 'block';
      this.tooltipElement.style.opacity = '0';

      requestAnimationFrame(() => {
        if (!this.tooltipElement) return;

        const rect = this.tooltipElement.getBoundingClientRect();
        const viewportWidth = window.innerWidth;

        const arrowHeight = 6;
        const gap = 11;

        // 目标中心点
        const targetCenterX = targetX + targetWidth / 2;
        let adjustedX = targetCenterX - rect.width / 2;
        const adjustedY = targetY - rect.height - arrowHeight - gap;

        // 水平边界检查
        if (adjustedX < 10) {
          adjustedX = 10;
        } else if (adjustedX + rect.width > viewportWidth - 10) {
          adjustedX = viewportWidth - rect.width - 10;
        }

        // 计算箭头在气泡内的相对位置，指向目标中心
        const arrowLeft = targetCenterX - adjustedX;
        // 限制箭头不超出气泡边界（留 12px 安全距离）
        const clampedArrowLeft = Math.max(12, Math.min(arrowLeft, rect.width - 12));

        // 箭头朝下
        const arrow = this.tooltipElement.querySelector('.tooltip-arrow') as HTMLDivElement;
        const arrowBorder = this.tooltipElement.querySelector('.tooltip-arrow-border') as HTMLDivElement;

        if (arrow && arrowBorder) {
          arrowBorder.style.bottom = '-6px';
          arrowBorder.style.top = 'auto';
          arrowBorder.style.left = `${clampedArrowLeft}px`;
          arrowBorder.style.transform = 'translateX(-50%)';
          arrowBorder.style.borderTop = '6px solid #dcdee5';
          arrowBorder.style.borderBottom = 'none';

          arrow.style.bottom = '-5px';
          arrow.style.top = 'auto';
          arrow.style.left = `${clampedArrowLeft}px`;
          arrow.style.transform = 'translateX(-50%)';
          arrow.style.borderTop = '5px solid #fff';
          arrow.style.borderBottom = 'none';
        }

        this.tooltipElement.style.left = `${adjustedX}px`;
        this.tooltipElement.style.top = `${adjustedY}px`;
        this.tooltipElement.style.opacity = '1';
      });
    }, 300);
  }

  /**
   * 隐藏 tooltip
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
        }, 200);
      }
    }, 300);
  }

  destroy() {
    if (this.hideTimer) clearTimeout(this.hideTimer);
    if (this.showTimer) clearTimeout(this.showTimer);
    if (this.tooltipElement && this.tooltipElement.parentNode) {
      this.tooltipElement.parentNode.removeChild(this.tooltipElement);
    }
    this.tooltipElement = null;
  }
}

export const tableTooltip = new TableTooltip();
