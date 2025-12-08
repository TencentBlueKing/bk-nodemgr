import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
// 引入需要的图形 Shape
import { Circle, Line, Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import type { UnitType } from './config';

export interface INodeData {
  name: string;
  running_proxy: number;
  total_proxy: number;
  running_agent: number;
  total_agent: number;
  cycle_times: string[];
  unitType: UnitType;
  is_healthy: Boolean;
  area?: string;
  is_direct?: boolean;
  direct_endpoints?: any;
  accesspoints?: any[];
  links?: any;
}

interface SplitTextConfig {
  prefix: string;
  suffix: string;
  prefixColor: string;
}

export default class NetWorkUnitNode extends BaseNode {
  // --- 1. 尺寸配置 ---
  static gridWidth = 135;   // 单个格子的宽度
  static gridHeight = 50;   // 单个格子的高度
  static headerHeight = 36; // 标题栏高度

  // 容器内边距 (关键：让表格和Agent看起来在内部)
  static paddingX = 12;     // 左右内边距
  static paddingY = 12;     // 上下内边距
  static contentGap = 12;   // 表格和Agent之间的间距

  static agentBarHeight = 50;
  static badgeRadius = 8;

  // --- 颜色配置 ---
  static greenColor = '#45E35F';
  static gray = '#979ba5';
  static warnColor = '#FFB848';
  static badgeColorP = '#3A84FF';
  static badgeColorA = '#8B5CF6';
  static borderColor = '#E1E4E8';

  // --- 动态计算属性 ---

  // 节点内容宽度 (表格宽度) = 2个格子宽
  static get contentWidth(): number {
    return this.gridWidth * 2;
  }

  // 节点总宽度 = 内容宽 + 左右内边距
  static get nodeWidth(): number {
    return this.contentWidth + (this.paddingX * 2);
  }

  // 获取表格部分的高度
  static getTableHeight(isDirect: boolean): number {
    return isDirect ? this.gridHeight : (this.gridHeight * 2);
  }

  // 获取节点总高度
  static getNodeTotalHeight(isDirect: boolean): number {
    const tableH = this.getTableHeight(isDirect);
    // 总高度 = 标题栏 + 上内边距 + 表格高度 + 间距 + Agent高度 + 下内边距
    return this.headerHeight + this.paddingY + tableH + this.contentGap + this.agentBarHeight + this.paddingY;
  }

  /**
   * 【核心逻辑】自定义锚点
   * 计算表格部分在整个节点高度中的相对位置 (0-1)
   */
  public getAnchorPoints() {
    const isDirect = this.data.is_direct as boolean;

    const totalHeight = NetWorkUnitNode.getNodeTotalHeight(isDirect);
    const tableHeight = NetWorkUnitNode.getTableHeight(isDirect);

    // 表格区域的起始 Y 坐标 (标题栏 + 上内边距)
    const tableStartY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.paddingY;

    // 表格区域的中心 Y 坐标
    const tableCenterY = tableStartY + (tableHeight / 2);

    // 转换为相对比例 (0~1)
    const ratio = tableCenterY / totalHeight;

    return [
      [0, ratio], // 左侧锚点
      [1, ratio], // 右侧锚点
    ];
  }

  // 获取节点数据
  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;
  }

  // 主容器样式
  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    const { is_direct } = this.data;
    const width = NetWorkUnitNode.nodeWidth;
    const height = NetWorkUnitNode.getNodeTotalHeight(is_direct as boolean);

    return {
      ...super.getKeyStyle(attr),
      width,
      height,
      fill: '#ffffff',
      stroke: '#DCDEE5',
      strokeWidth: 1,
      radius: 6, // 整体大圆角
      shadowColor: 'rgba(0, 0, 0, 0.06)',
      shadowBlur: 8,
      shadowOffsetY: 2,
      cursor: 'pointer',
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', 'rect', this.getKeyStyle(attr), container);
  }

  // 绘制标题栏
  private drawHeader(container: Group) {
    const width = NetWorkUnitNode.nodeWidth;
    const height = NetWorkUnitNode.headerHeight;

    // 标题背景 (带顶部圆角)
    this.upsert('header-bg', 'rect', {
      x: 0,
      y: 0,
      width,
      height,
      fill: '#F0F1F5',
      radius: [6, 6, 0, 0],
      stroke: NetWorkUnitNode.borderColor,
      strokeWidth: 1,
    }, container);

    // 底部边框线 (确保标题栏和内容区分割)
    this.upsert('header-border', 'line', {
      x1: 0,
      y1: height,
      x2: width,
      y2: height,
      stroke: NetWorkUnitNode.borderColor,
      strokeWidth: 1,
    }, container);
  }

  private drawUnitLabel(container: Group) {
    const h = NetWorkUnitNode.headerHeight;
    // 标识
    this.upsert('unit-tag-bg', 'rect', {
      x: 12,
      y: 6,
      width: 36,
      height: 20,
      fill: '#E1ECFF',
      radius: 2,
    }, container);

    this.upsert('unit-tag-text', 'text', {
      x: 12 + 18,
      y: 6 + 10,
      text: '单元',
      fontSize: 10,
      fill: '#3A84FF',
      textAlign: 'center',
      textBaseline: 'middle',
      fontWeight: 'bold',
    }, container);
  }

  private drawUnitName(container: Group) {
    const { name } = this.data;
    const h = NetWorkUnitNode.headerHeight;

    this.upsert('unit-name', 'text', {
      x: 58,
      y: h / 2,
      text: name,
      fontSize: 14,
      fill: '#313238',
      textBaseline: 'middle',
      fontWeight: 600,
    }, container);
  }

  // --- 通用绘制方法 ---

  // 绘制角标 (P 或 A) - 悬挂在左上角
  private drawBadge(container: Group, startX: number, startY: number, text: string, color: string) {
    const r = NetWorkUnitNode.badgeRadius;
    // 圆心位置：稍微往外突出一点点，更有“角标”感
    const cx = startX;
    const cy = startY;

    this.upsert(`badge-bg-${text}`, 'circle', {
      cx,
      cy,
      r,
      fill: color,
      zIndex: 20,
      stroke: '#fff', // 加个白边，分离感更好
      strokeWidth: 1.5,
    }, container);

    this.upsert(`badge-text-${text}`, 'text', {
      x: cx,
      y: cy,
      text,
      fontSize: 10,
      fill: '#fff',
      textAlign: 'center',
      textBaseline: 'middle',
      fontWeight: 'bold',
      zIndex: 21,
    }, container);
  }

  // 绘制格子
  private drawGrid(
    container: Group,
    x: number,
    y: number,
    mainText: string,
    subText: string,
    mainColor: string,
    split?: SplitTextConfig,
    isLatency?: boolean,
  ) {
    // 格子背景框
    this.upsert(`grid-rect-${x}-${y}`, 'rect', {
      x,
      y,
      width: NetWorkUnitNode.gridWidth,
      height: NetWorkUnitNode.gridHeight,
      fill: '#fff',
      stroke: NetWorkUnitNode.borderColor,
      strokeWidth: 1,
    }, container);

    const cx = x + NetWorkUnitNode.gridWidth / 2;
    const textY = y + NetWorkUnitNode.gridHeight / 2 - (subText ? 8 : 0);
    const key = `grid-content-${x}-${y}`;

    // 1. 绘制主文本
    if (split) {
      // 拆分文本 (e.g. 2/2)
      const { prefix, suffix, prefixColor } = split;
      const fullW = (prefix.length + suffix.length + 1) * 9; // 估算宽度
      const startX = cx - fullW / 2;

      this.upsert(`${key}-pre`, 'text', { x: cx - 5, y: textY, text: prefix, fill: prefixColor, fontSize: 16, fontWeight: 600, textAlign: 'right', textBaseline: 'middle' }, container);
      this.upsert(`${key}-slash`, 'text', { x: cx, y: textY, text: '/', fill: '#979ba5', fontSize: 14, textAlign: 'center', textBaseline: 'middle' }, container);
      this.upsert(`${key}-suf`, 'text', { x: cx + 5, y: textY, text: suffix, fill: '#979ba5', fontSize: 16, textAlign: 'left', textBaseline: 'middle' }, container);
    } else if (isLatency && mainText) {
      // 延迟文本 (带颜色)
      const parts = mainText.replace('ms', '').split(',');
      const offsetX = 0;
      // 简单处理：直接画在中间
      // 实际项目可能需要更精细的测量
      this.upsert(`${key}-lat`, 'text', {
        x: cx,
        y: textY,
        text: mainText, // 简单处理，暂不拆分颜色
        fill: parseInt(parts[0]) > 100 ? NetWorkUnitNode.warnColor : NetWorkUnitNode.greenColor,
        fontSize: 16,
        textAlign: 'center',
        textBaseline: 'middle',
      }, container);
    } else {
      // 普通文本
      this.upsert(`${key}-txt`, 'text', {
        x: cx,
        y: textY,
        text: mainText,
        fill: mainColor,
        fontSize: 16,
        fontWeight: 500,
        textAlign: 'center',
        textBaseline: 'middle',
      }, container);
    }

    // 2. 绘制副标题
    if (subText) {
      this.upsert(`${key}-sub`, 'text', {
        x: cx,
        y: y + 36, // 靠下
        text: subText,
        fill: '#979ba5',
        fontSize: 12,
        textAlign: 'center',
        textBaseline: 'middle',
      }, container);
    }
  }

  // --- 内容块绘制 ---

  private drawProxyTable(container: Group) {
    const {
      is_direct, direct_endpoints, accesspoints,
      running_proxy, total_proxy, cycle_times, is_healthy,
    } = this.data;

    // 计算表格起始位置 (在 Padding 内部)
    const startX = NetWorkUnitNode.paddingX;
    const startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.paddingY;

    if (is_direct) {
      // 直连：一行
      this.drawGrid(container, startX, startY, 'Server', '', '#1768EF');
      this.drawGrid(
        container, startX + NetWorkUnitNode.gridWidth, startY,
        `${accesspoints?.length || 0}`, '接入点数量', NetWorkUnitNode.greenColor,
      );
    } else {
      // 非直连：两行
      // Row 1
      this.drawGrid(
        container, startX, startY,
        `${running_proxy}/${total_proxy}`, 'Proxy', '#333',
        { prefix: `${running_proxy}`, suffix: `${total_proxy}`, prefixColor: NetWorkUnitNode.greenColor },
      );
      this.drawGrid(
        container, startX + NetWorkUnitNode.gridWidth, startY,
        `${accesspoints?.length || 0}`, '接入点数量', NetWorkUnitNode.greenColor,
      );
      // Row 2
      const y2 = startY + NetWorkUnitNode.gridHeight;
      const latencyStr = cycle_times.some(t => t !== '0') ? cycle_times.join(', ') : '-';
      this.drawGrid(container, startX, y2, latencyStr, '延迟', '', undefined, true);

      const statusColor = is_healthy ? NetWorkUnitNode.greenColor : '#EA3536';
      const statusText = is_healthy ? '健康' : '异常';
      this.drawGrid(container, startX + NetWorkUnitNode.gridWidth, y2, statusText, '状态', statusColor);
    }

    // 绘制 Proxy 角标 "P" (在表格左上角)
    this.drawBadge(container, startX, startY, 'P', NetWorkUnitNode.badgeColorP);
  }

  private drawAgentBlock(container: Group) {
    const { is_direct, running_agent, total_agent } = this.data;

    // 计算 Agent 块位置
    // Y = 标题 + 上Pad + 表格高 + 间距
    const tableH = NetWorkUnitNode.getTableHeight(is_direct as boolean);
    const startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.paddingY + tableH + NetWorkUnitNode.contentGap;
    const startX = NetWorkUnitNode.paddingX;
    const width = NetWorkUnitNode.contentWidth; // 270
    const height = NetWorkUnitNode.agentBarHeight; // 50

    // 背景框
    this.upsert('agent-bg', 'rect', {
      x: startX,
      y: startY,
      width,
      height,
      fill: '#FAFBFD',
      stroke: NetWorkUnitNode.borderColor,
      strokeWidth: 1,
      radius: 4,
    }, container);

    // 绘制 Agent 角标 "A" (在块的左上角)
    this.drawBadge(container, startX, startY, 'A', NetWorkUnitNode.badgeColorA);

    // 内容文本 (居中)
    const centerX = startX + width / 2;
    const centerY = startY + height / 2;

    // 数量
    const prefix = `${running_agent}`;
    const suffix = `${total_agent}`;
    const textColor = running_agent > 0 ? NetWorkUnitNode.greenColor : '#979ba5';

    this.upsert('agent-val-pre', 'text', { x: centerX - 5, y: centerY - 6, text: prefix, fill: textColor, fontSize: 16, fontWeight: 600, textAlign: 'right', textBaseline: 'middle' }, container);
    this.upsert('agent-val-slash', 'text', { x: centerX, y: centerY - 6, text: '/', fill: '#979ba5', fontSize: 14, textAlign: 'center', textBaseline: 'middle' }, container);
    this.upsert('agent-val-suf', 'text', { x: centerX + 5, y: centerY - 6, text: suffix, fill: '#979ba5', fontSize: 16, textAlign: 'left', textBaseline: 'middle' }, container);

    // Label
    this.upsert('agent-lbl', 'text', {
      x: centerX,
      y: centerY + 10,
      text: 'Agent',
      fontSize: 12,
      fill: '#979ba5',
      textAlign: 'center',
      textBaseline: 'middle',
    }, container);
  }

  // 绘制连接线 (表格 -> Agent)
  private drawConnectLine(container: Group) {
    const { is_direct } = this.data;
    const tableH = NetWorkUnitNode.getTableHeight(is_direct as boolean);

    // 线条起点：表格底部中心
    const startX = NetWorkUnitNode.nodeWidth / 2;
    const startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.paddingY + tableH;

    // 线条终点：Agent 顶部
    const endY = startY + NetWorkUnitNode.contentGap;

    this.upsert('connect-line', 'line', {
      x1: startX,
      y1: startY,
      x2: startX,
      y2: endY,
      stroke: '#3182ce',
      strokeWidth: 1,
    }, container);
  }

  private drawMenuIcon(container: Group) {
    const iconX = NetWorkUnitNode.nodeWidth - 28;
    const iconY = NetWorkUnitNode.headerHeight / 2;

    this.upsert('menu-hit-area', 'rect', {
      x: iconX - 10,
      y: 0,
      width: 30,
      height: NetWorkUnitNode.headerHeight,
      fill: 'transparent',
      cursor: 'pointer',
      zIndex: 50,
      className: 'menu-hit-area', // 关键：用于事件识别
    }, container);

    // 三个点
    [-1, 0, 1].forEach((i) => {
      this.upsert(`menu-dot-${i}`, 'circle', {
        cx: iconX + i * 5,
        cy: iconY,
        r: 2,
        fill: '#979ba5',
        pointerEvents: 'none',
      }, container);
    });
  }

  // 主渲染函数
  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);

    // 1. 绘制框架
    this.drawHeader(container);
    this.drawUnitLabel(container);
    this.drawUnitName(container);
    this.drawMenuIcon(container);

    // 2. 绘制内容模块
    this.drawProxyTable(container);
    this.drawAgentBlock(container);

    // 3. 绘制内部连接线
    this.drawConnectLine(container);
  }
}
