import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Line, Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import type { UnitType } from './config';
import { fill } from 'lodash';

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

// 拆分文本配置类型（支持部分文字染色）
interface SplitTextConfig {
  prefix: string; // 染色前缀（如 running_proxy）
  suffix: string; // 普通后缀（如 total_proxy）
  prefixColor: string; // 前缀颜色
}

export default class NetWorkUnitNode extends BaseNode {
  // 布局常量
  static gridWidth = 135; // 单元格宽度
  static gridHeight = 50; // 单元格高度
  static headerHeight = 32; // 标题栏高度
  static space = 5;
  static agentBarHeight = 50; // Agent栏高度
  static agentBarRadius = 12; // Agent栏圆角
  static agentMarginTop = 12; // Agent栏与Unit间距
  static greenColor = '#45E35F'; // 统一绿色
  static gray = '#979ba5'; // 灰色
  static warnColor = '#FFB848'; // 黄色

  // Unit节点主容器样式（仅包含Unit主体，不含Agent栏）
  static defaultNodeStyle: RectStyleProps = {
    stroke: '#c4c6cc',
    strokeWidth: 1,
    shadowColor: '#dee0e5',
    shadowBlur: 4,
    shadowOffsetX: 0,
    shadowOffsetY: 2,
    fill: '#ffffff',
    cursor: 'pointer',
    fillOpacity: 1,
    width: NetWorkUnitNode.gridWidth * 2,
    height: 0, // 动态计算（仅Unit主体高度）
    radius: [4, 4, 0, 0], // 仅顶部圆角（与Agent栏视觉分离）
  };

  // 标题栏样式
  static defaultHeaderStyle: Partial<RectStyleProps> = {
    fill: '#f0f1f5',
    height: NetWorkUnitNode.headerHeight,
    radius: [4, 4, 0, 0],
    stroke: '#c4c6cc',
    strokeWidth: 1,
  };

  // 单元标识样式
  static defaultUnitLabelStyle: Partial<RectStyleProps> = {
    width: 36,
    height: 20,
    fill: '#c2d5f6',
    radius: 3,
  };

  // 文本样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 16,
    fill: NetWorkUnitNode.gray,
    textBaseline: 'middle',
  };

  static defaultSmallTextStyle: Partial<TextStyleProps> = {
    fontSize: 10,
    fill: '#979ba5',
    textBaseline: 'top',
  };

  // 格子样式
  static defaultGridStyle: Partial<RectStyleProps> = {
    stroke: '#e1e4e8',
    strokeWidth: 1,
    fill: 'transparent',
    radius: 2,
  };

  // Agent栏样式（独立样式，与Unit分离）
  static defaultAgentBarStyle: Partial<RectStyleProps> = {
    fill: '#ffffff',
    stroke: '#e1e4e8',
    strokeWidth: 1,
    radius: NetWorkUnitNode.agentBarRadius,
    shadowColor: '#dee0e5',
    shadowBlur: 2,
    shadowOffsetX: 0,
    shadowOffsetY: 1,
  };

  // 获取节点数据
  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;
  }

  // 计算节点总高度（Unit主体高度 + Agent栏高度 + 间距）
  static getNodeTotalHeight(isDirect: Boolean): number {
    const unitHeight = NetWorkUnitNode.getUnitInnerHeight(isDirect);
    return unitHeight + NetWorkUnitNode.agentMarginTop + NetWorkUnitNode.agentBarHeight;
  }

  // 计算Unit主体高度（标题栏 + 内容区）
  static getUnitInnerHeight(isDirect: Boolean): number {
    const contentHeight = isDirect
      ? NetWorkUnitNode.gridHeight
      : NetWorkUnitNode.gridHeight * 2;
    return NetWorkUnitNode.headerHeight + contentHeight;
  }

  // 配置节点主容器（keyShape）样式（仅包含Unit主体）
  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    const { is_direct } = this.data;
    const unitHeight = NetWorkUnitNode.getUnitInnerHeight(is_direct as Boolean);

    return {
      ...super.getKeyStyle(attr),
      ...NetWorkUnitNode.defaultNodeStyle,
      height: unitHeight, // 主容器高度 = Unit主体高度（不含Agent栏）
    };
  }

  // 绘制节点主容器（白色背景，仅Unit主体）
  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', 'rect', this.getKeyStyle(attr), container);
  }

  // 绘制标题栏
  private drawHeader(container: Group) {
    const width = NetWorkUnitNode.gridWidth * 2;
    return this.upsert('header', 'rect', {
      ...NetWorkUnitNode.defaultHeaderStyle,
      width,
      y: 0,
    }, container);
  }

  // 绘制单元标识（“单元”文字）
  private drawUnitLabel(container: Group) {
    const { headerHeight } = NetWorkUnitNode;
    const labelStyle = NetWorkUnitNode.defaultUnitLabelStyle;

    // 标识背景
    this.upsert('unit-label-bg', 'rect', {
      ...labelStyle,
      x: 10,
      y: (headerHeight - (labelStyle.height as number)) / 2,
    }, container);

    // 标识文字
    this.upsert('unit-label-text', 'text', {
      x: 10 + (labelStyle.width as number) / 2,
      y: headerHeight / 2,
      text: '单元',
      fontSize: 10,
      fill: '#ffffff',
      fontWeight: 'bold',
      textAlign: 'center',
      textBaseline: 'middle',
    }, container);
  }

  // 绘制单元名称
  private drawUnitName(container: Group) {
    const { headerHeight } = NetWorkUnitNode;
    const { name } = this.data;

    return this.upsert('unit-name', 'text', {
      x: 56,
      y: headerHeight / 2,
      text: name,
      fontSize: 14,
      fill: '#313238',
      textBaseline: 'middle',
      fontWeight: 600,
    }, container);
  }

  // 绘制格子（支持普通文本/拆分文本染色 + 竖线延长）
  private drawGrid(
    container: Group,
    x: number,
    y: number,
    mainText: string,
    subText: string,
    mainTextColor = '#313238',
    splitTextConfig?: SplitTextConfig,
    latency?: Boolean,
  ) {
    // 绘制格子边框
    this.upsert(`grid-${x}-${y}`, 'rect', {
      x,
      y,
      width: NetWorkUnitNode.gridWidth,
      height: NetWorkUnitNode.gridHeight,
      ...NetWorkUnitNode.defaultGridStyle,
    }, container);

    const mainTextY = y + NetWorkUnitNode.gridHeight / 2 - (subText ? 8 : 0);
    const gridKeyPrefix = `grid-main-${x}-${y}`;

    // 拆分文本模式（部分染色）
    if (splitTextConfig) {
      const { prefix, suffix, prefixColor } = splitTextConfig;
      const centerX = x + NetWorkUnitNode.gridWidth / 2;

      // 染色前缀
      this.upsert(`${gridKeyPrefix}-prefix`, 'text', {
        x: centerX - prefix.length * 9 - 4,
        y: mainTextY,
        text: prefix,
        ...NetWorkUnitNode.defaultTextStyle,
        fill: prefixColor,
        fontWeight: 600,
        textAlign: 'left',
        textBaseline: 'middle',
      }, container);

      // 斜杠
      if (suffix) {
        this.upsert(`${gridKeyPrefix}-slash`, 'text', {
          x: centerX - 2,
          y: mainTextY,
          text: '/',
          fontSize: 14,
          fill: '#979ba5',
          textAlign: 'left',
          textBaseline: 'middle',
        }, container);
      }

      // 普通后缀
      if (suffix) {
        this.upsert(`${gridKeyPrefix}-suffix`, 'text', {
          x: centerX + 3,
          y: mainTextY,
          text: suffix,
          ...NetWorkUnitNode.defaultTextStyle,
          textAlign: 'left',
          textBaseline: 'middle',
        }, container);
      }
    } else if (latency && mainText) {
      const centerX = x + NetWorkUnitNode.gridWidth / 2;
      const parts = mainText.replace('ms', '').split(','); // 按逗号分割 ["120", "80", "150"]
      const textParts: any[] = [];

      // 1. 准备所有要绘制的文本片段（数字、逗号、ms）
      parts.forEach((part, index) => {
        const num = parseInt(part, 10);
        let color = NetWorkUnitNode.gray;
        if (!isNaN(num)) {
          color = num < 100 ? NetWorkUnitNode.greenColor : NetWorkUnitNode.warnColor;
        }
        textParts.push({
          text: part,
          color,
          isNumber: true,
        });
        // 如果不是最后一个数字，添加逗号
        if (index < parts.length - 1) {
          textParts.push({ text: ',', color: NetWorkUnitNode.gray, isNumber: false });
        }
      });
      // 添加最后的 "ms"
      textParts.push({ text: 'ms', color: NetWorkUnitNode.gray, isNumber: false });

      // 2. 计算所有文本的总宽度，以便居中
      let totalWidth = 0;
      textParts.forEach(p => {
        // 这里假设数字和非数字的字宽相同，用一个平均宽度估算
        totalWidth += p.text.length * 8;
      });

      // 3. 计算起始绘制位置
      let currentX = centerX - totalWidth / 2;

      // 4. 逐个绘制文本片段
      textParts.forEach((part, index) => {
        this.upsert(`${gridKeyPrefix}-latency-${index}`, 'text', {
          x: currentX,
          y: mainTextY,
          text: part.text,
          ...NetWorkUnitNode.defaultTextStyle,
          fill: part.color,
          textAlign: 'left',
          textBaseline: 'middle',
        }, container);

        // 更新下一个片段的X坐标
        currentX += part.text.length * 8;
      });
    } else { // 普通文本模式
      this.upsert(`${gridKeyPrefix}`, 'text', {
        x: x + NetWorkUnitNode.gridWidth / 2,
        y: mainTextY,
        text: mainText,
        ...NetWorkUnitNode.defaultTextStyle,
        fill: mainTextColor,
        textAlign: 'center',
        textBaseline: 'middle',
      }, container);
    }

    // 小字
    if (subText) {
      this.upsert(`grid-sub-${x}-${y}`, 'text', {
        x: x + NetWorkUnitNode.gridWidth / 2,
        y: y + NetWorkUnitNode.gridHeight / 2 + 10,
        text: subText,
        ...NetWorkUnitNode.defaultSmallTextStyle,
        textAlign: 'center',
        textBaseline: 'middle',
      }, container);
    }
  }

  // 绘制独立Agent栏（外置，不包含在keyShape中）
  private drawAgentStatBar(container: Group) {
    const { running_agent, total_agent } = this.data;
    const unitHeight = NetWorkUnitNode.getUnitInnerHeight(this.data.is_direct as Boolean);
    const barY = unitHeight + NetWorkUnitNode.agentMarginTop; // Agent栏Y坐标 = Unit主体底部 + 间距
    const barWidth = NetWorkUnitNode.gridWidth * 2 - 20; // 左右留边，比Unit窄
    const barX = 10;
    const centerX = barX + barWidth / 2;
    const centerY = barY + NetWorkUnitNode.agentBarHeight / 2;

    // 1. Agent栏背景（独立矩形）
    this.upsert('agent-stat-bar', 'rect', {
      ...NetWorkUnitNode.defaultAgentBarStyle,
      x: barX,
      y: barY,
      width: barWidth,
      height: NetWorkUnitNode.agentBarHeight,
    }, container);

    // 2. Agent数值（拆分文本，running_agent绿色）
    const prefix = `${running_agent}`;
    const suffix = `${total_agent}`;

    // 绿色前缀
    this.upsert('agent-stat-prefix', 'text', {
      x: centerX - prefix.length * 9 - 1,
      y: centerY - 4,
      text: prefix,
      ...NetWorkUnitNode.defaultTextStyle,
      fill: running_agent === 0 ? NetWorkUnitNode.gray : NetWorkUnitNode.greenColor,
      textAlign: 'left',
      textBaseline: 'middle',
    }, container);

    // 斜杠
    this.upsert('agent-stat-slash', 'text', {
      x: centerX - 2,
      y: centerY - 4,
      text: '/',
      fontSize: 14,
      fill: '#979ba5',
      textAlign: 'left',
      textBaseline: 'middle',
    }, container);

    // 普通后缀
    this.upsert('agent-stat-suffix', 'text', {
      x: centerX + 3,
      y: centerY - 4,
      text: suffix,
      ...NetWorkUnitNode.defaultTextStyle,
      fill: NetWorkUnitNode.gray,
      textAlign: 'left',
      textBaseline: 'middle',
    }, container);

    // 3. Agent文字
    this.upsert('agent-stat-subtext', 'text', {
      x: centerX,
      y: centerY + 12,
      text: 'Agent',
      ...NetWorkUnitNode.defaultSmallTextStyle,
      textAlign: 'center',
      textBaseline: 'middle',
    }, container);
  }


  // 绘制直连单元内容
  // 绘制直连单元内容，并返回竖线起点Y坐标
  private drawDirectUnitContent(container: Group) {
    const { direct_endpoints, accesspoints } = this.data;
    const contentStartY = NetWorkUnitNode.headerHeight;

    // 左侧格子：Server
    this.drawGrid(container, 0, contentStartY, 'Server', '', '#1768EF');

    // 右侧格子：接入点数量
    this.drawGrid(
      container, NetWorkUnitNode.gridWidth, contentStartY,
      accesspoints?.length.toString() || '0', '接入点数量', NetWorkUnitNode.greenColor,
    );

    // 绘制外置Agent栏
    this.drawAgentStatBar(container);
  }

  // 绘制非直连单元内容，并返回竖线起点Y坐标
  private drawIndirectUnitContent(container: Group) {
    const { is_healthy, cycle_times, running_proxy, total_proxy, accesspoints } = this.data;
    const contentStartY = NetWorkUnitNode.headerHeight;

    const latencyText = cycle_times.some(item => item !== '0') ? cycle_times.join(', ') : '-';
    const statusText = is_healthy ? '健康' : '异常';
    const statusColor = is_healthy ? NetWorkUnitNode.greenColor : '#EA3536';

    // 第一行：Proxy
    this.drawGrid(
      container, 0, contentStartY,
      `${running_proxy}/${total_proxy}`, 'Proxy', '#313238', {
        prefix: `${running_proxy}`,
        suffix: `${total_proxy}`,
        prefixColor: NetWorkUnitNode.greenColor,
      },
    );

    // 第一行：接入点数量
    this.drawGrid(
      container, NetWorkUnitNode.gridWidth, contentStartY,
      String(accesspoints?.length), '接入点数量', NetWorkUnitNode.greenColor,
    );

    // 第二行：延迟
    this.drawGrid(
      container, 0, contentStartY + NetWorkUnitNode.gridHeight,
      latencyText, '延迟', '', undefined, true,
    );

    // 第二行：状态
    const statusGridY = contentStartY + NetWorkUnitNode.gridHeight;
    this.drawGrid(
      container, NetWorkUnitNode.gridWidth, statusGridY,
      statusText, '状态', statusColor,
    );

    // 绘制外置Agent栏
    this.drawAgentStatBar(container);
  }

  // 绘制agent和unit的连线
  private drawVerticalLine(container: Group) {
    const keyShape = this.getShape('key');
    if (keyShape) {
      const unitWidth = keyShape.attr('width');
      const x = unitWidth / 2; // 竖线的X坐标：单元的正中心

      let startY: number;
      if (this.data.is_direct) {
        // 直连单元：从“接入点数量”格子的底边开始
        startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.gridHeight;
      } else {
        // 非直连单元：从“状态”格子的底边开始
        startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.gridHeight * 2;
      }

      this.upsert('vertical-divider-line', 'line', {
        x1: x,
        y1: startY,
        x2: x,
        y2: startY + 12,
        stroke: '#AAAAAA', // 颜色
        strokeWidth: 2,    // 宽度
      }, container);
    }
  }

  // ---------------- 新增：绘制菜单图标 ----------------
  private drawMenuIcon(container: Group) {
    const { headerHeight } = NetWorkUnitNode;
    // 整个单元的宽度 (gridWidth * 2)
    const totalWidth = NetWorkUnitNode.gridWidth * 2;

    // 菜单按钮的位置：右上角
    // 留出一些边距，比如距右边 10px
    const iconX = totalWidth - 24;
    const iconY = headerHeight / 2;

    // 1. 绘制一个透明的背景矩形，作为点击的热区 (Hit Area)
    // 这样用户稍微点歪一点也能触发，体验更好
    this.upsert('menu-hit-area', 'rect', {
      x: iconX - 10,
      y: 0,
      width: 34, // 足够宽
      height: headerHeight,
      fill: 'transparent',
      cursor: 'pointer',
      zIndex: 10, // 保证在最上层
      class: 'menu-hit-area',
      className: 'menu-hit-area',
    }, container);

    // 2. 绘制三个小圆点 (或者你也可以用 Text 写 "...")
    // 这里我们用三个 Circle 来画，比较精致
    const dotColor = '#979ba5';
    const r = 2; // 半径
    const gap = 5; // 间距

    [-1, 0, 1].forEach((i) => {
      this.upsert(`menu-dot-${i}`, 'circle', {
        cx: iconX + i * gap,
        cy: iconY,
        r: r,
        fill: dotColor,
        cursor: 'pointer',
        // 让圆点不捕获事件，让事件穿透到下面的 menu-hit-area
        // 这样我们在事件监听时只需要判断 hit-area 即可
        pointerEvents: 'none',
      }, container);
    });
  }
  // 核心渲染入口
  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);

    // 绘制Unit主体内容
    this.drawHeader(container);
    this.drawUnitLabel(container);
    this.drawUnitName(container);

    // 根据单元类型绘制内容 + 外置Agent栏
    if (this.data.is_direct) {
      this.drawDirectUnitContent(container);
    } else {
      this.drawIndirectUnitContent(container);
    }

    this.drawVerticalLine(container);
    // 【新增】绘制菜单
    this.drawMenuIcon(container);
  }
}
