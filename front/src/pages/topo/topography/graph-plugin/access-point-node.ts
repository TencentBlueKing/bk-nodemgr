import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
// 确保引入了 Rect 和 Text
import { Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import { NodeType } from './config';

export interface IAccessPointData {
  name: string;
  downstreamUnits: number;
  type: 'internal' | 'external'; // 内网/外网标识
  area?: string;
  bk_networkunit_id?: number;
}

export default class AccessPointNode extends BaseNode {
  static nodeWidth = 140;
  static nodeHeight = 40;

  // 默认外层节点属性
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
    width: AccessPointNode.nodeWidth,
    height: AccessPointNode.nodeHeight,
    radius: 4,
  };

  // 顶部栏样式（与节点高度一致）
  static defaultHeaderStyle: Partial<RectStyleProps> = {
    fill: '#f0f1f5',
    height: AccessPointNode.nodeHeight,
    radius: [4, 4, 4, 4], // 全圆角
    stroke: '#c4c6cc',
    strokeWidth: 1,
  };

  // 接入点标识样式（替换AP为“接入点+内网/外网”）
  static defaultAPLabelStyle: Partial<RectStyleProps> = {
    width: 40,
    height: 20,
    fill: '#8B5CF6', // 原有紫色
    radius: 3,
  };

  // 默认文案样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 12,
    fill: '#313238',
    textBaseline: 'middle',
  };

  get data(): IAccessPointData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as IAccessPointData;
  }

  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    return {
      ...super.getKeyStyle(attr),
      ...AccessPointNode.defaultNodeStyle,
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', 'rect', this.getKeyStyle(attr), container);
  }

  // 绘制顶部栏
  private drawHeader(container: Group) {
    const width = AccessPointNode.nodeWidth;
    const height = AccessPointNode.nodeHeight;

    return this.upsert('header', 'rect', {
      ...AccessPointNode.defaultHeaderStyle,
      width,
      height,
      y: 0,
      cursor: 'move',
    }, container);
  }

  private drawAPLabel(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const labelStyle = AccessPointNode.defaultAPLabelStyle;
    // const { type } = this.data;

    // 标识背景矩形
    this.upsert('ap-label-bg', 'rect', {
      ...labelStyle,
      x: 10,
      y: (height - (labelStyle.height as number)) / 2,
      cursor: 'move',
    }, container);

    // “接入点 内网/外网”文字
    this.upsert('ap-label-text', 'text', {
      x: 10 + (labelStyle.width as number) / 2,
      y: height / 2,
      text: '接入点',
      fontSize: 10,
      fill: '#ffffff',
      fontWeight: 'bold',
      textAlign: 'center',
      textBaseline: 'middle',
      cursor: 'move',
    }, container);

    return container;
  }

  // 绘制接入点名称 (修改部分)
  private drawAccessPointName(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const width = AccessPointNode.nodeWidth;
    const { name } = this.data;

    // 计算文字起始位置和可用宽度
    const textStartX = 60;
    const paddingRight = 8; // 右侧留一点空隙
    const availableWidth = width - textStartX - paddingRight - 12; // 140 - 60 - 8 = 72px

    return this.upsert('access-point-name', 'text', {
      x: textStartX, // 跟在标识后面
      y: height / 2,
      text: name,
      ...AccessPointNode.defaultTextStyle,
      fontWeight: 600,
      textAlign: 'left', // 确保左对齐

      // --- 截断配置 Start ---
      wordWrap: true,             // 开启自动换行/截断逻辑
      wordWrapWidth: availableWidth, // 设置最大宽度
      maxLines: 1,                // 限制最大行数为 1
      textOverflow: 'ellipsis',   // 超出部分显示省略号 '...'
      cursor: 'move',
    }, container);
  }

  // 【新增】绘制信息图标 (i)
  private drawInfoIcon(container: Group) {
    const width = AccessPointNode.nodeWidth;
    const height = AccessPointNode.nodeHeight;

    // 放在右侧，垂直居中
    const iconX = width - 12;
    const iconY = height / 2;

    // 1. 扩大鼠标感应区 (透明矩形)
    this.upsert('info-hit-area', 'rect', {
      x: width - 25,
      y: 0,
      width: 25,
      height,
      fill: 'transparent',
      cursor: 'pointer',
      className: 'ap-info-icon', // 关键标识：用于 topo.vue 识别 Hover
    }, container);

    // 2. 绘制图标 (这里简单画个红色的 i，或者用图片/Iconfont)
    // 也可以画个圆圈+文字

    this.upsert('info-circle', 'circle', {
      cx: iconX,
      cy: iconY,
      r: 7,
      stroke: '#979ba5', // 边框颜色
      lineWidth: 1.5,     // 边框粗细
      fill: '#979ba5',    // 圆圈内部填充白色，盖住下面的线或背景
      pointerEvents: 'none', // 让事件穿透到上面的 hit-area
    }, container);

    this.upsert('info-text', 'text', {
      x: iconX,
      y: iconY,
      text: 'i',
      fontSize: 14,
      fill: '#ffff', // 红色
      fontWeight: 'bold',
      fontFamily: 'serif', // 衬线体看起来像图标
      textAlign: 'center',
      textBaseline: 'middle',
      pointerEvents: 'none', // 让事件穿透到 hit-area
    }, container);
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);

    // 绘制顶部栏、标识、名称
    this.drawHeader(container);
    this.drawAPLabel(container);
    this.drawAccessPointName(container);
    this.drawInfoIcon(container);
  }
}
