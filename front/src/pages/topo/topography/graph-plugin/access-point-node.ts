import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
// 确保引入了 Rect 和 Text
import { Image as GImage, Rect as GRect, Text as GText } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import Downstream from '../../../../../public/images/downstream.svg';

export interface IAccessPointData {
  name: string;
  downstreamUnits: number;
  type: 'internal' | 'external'; // 内网/外网标识
  area?: string;
  bk_networkunit_id?: number;
}

export default class AccessPointNode extends BaseNode {
  static nodeWidth = 120;
  static nodeHeight = 22;

  // 默认外层节点属性
  static defaultNodeStyle: RectStyleProps = {
    shadowColor: '#dee0e5',
    shadowBlur: 4,
    shadowOffsetX: 0,
    shadowOffsetY: 2,
    fill: '#ffffff',
    cursor: 'pointer',
    fillOpacity: 1,
    width: AccessPointNode.nodeWidth,
    height: AccessPointNode.nodeHeight,
    radius: 99,
  };

  // 顶部栏样式（与节点高度一致）
  static defaultHeaderStyle: Partial<RectStyleProps> = {
    fill: '#FFFFFF',
    height: AccessPointNode.nodeHeight,
    radius: 99, // 全圆角
  };

  // 默认文案样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 12,
    fill: '#4D4F56',
    textBaseline: 'middle',
  };

  get data(): IAccessPointData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as IAccessPointData;
  }

  get estimateTextWidth() {
    const text = this.data.name;
    if (!text) return 0;
    const chineseRegex = /[\u4e00-\u9fa5]/g;
    const chineseChars = text.match(chineseRegex) || [];
    const chineseCount = chineseChars.length;
    const totalLength = text.length;
    const nonChineseCount = totalLength - chineseCount;
    return Math.round(chineseCount * 13 + nonChineseCount * 7);
  };

  get getCalculatedWidth() {
    const minWidth = 80;
    const maxWidth = 160;
    const contentWidth = this.estimateTextWidth;

    // 计算自适应宽度
    let calculatedWidth = contentWidth;
    if (contentWidth < minWidth) {
      calculatedWidth = minWidth;
    } else if (contentWidth > maxWidth) {
      calculatedWidth = maxWidth;
    }
    return calculatedWidth;
  }

  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    return {
      ...super.getKeyStyle(attr),
      ...AccessPointNode.defaultNodeStyle,
      width: this.getCalculatedWidth,
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', GRect, this.getKeyStyle(attr), container);
  }

  // 绘制容器
  private drawContainer(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const calculatedWidth = this.getCalculatedWidth;

    return this.upsert('ap-container', GRect, {
      ...AccessPointNode.defaultHeaderStyle,
      width: calculatedWidth,
      height,
      y: 0,
      cursor: 'move',
    }, container);
  }

  // 接入点标识图标
  private drawAPLabel(container: Group) {
    const height = AccessPointNode.nodeHeight;

    this.upsert('downstream-ap', GImage, {
      x: 10,
      y: (height - 10) / 2, // 10是图标的高度
      src: Downstream,
      cursor: 'move',
    }, container);

    return container;
  }

  // 绘制接入点名称 (修改部分)
  private drawAccessPointName(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const width = this.getCalculatedWidth;
    const { name } = this.data;

    // 计算文字起始位置和可用宽度
    const textStartX = 26;
    const paddingRight = 12; // 右侧留一点空隙
    const availableWidth = width - textStartX - paddingRight - 12; // 140 - 60 - 8 = 72px

    return this.upsert('access-point-name', GText, {
      x: textStartX, // 跟在标识后面
      y: height / 2,
      text: name,
      ...AccessPointNode.defaultTextStyle,
      fontWeight: 400,
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
    const width = this.getCalculatedWidth;
    const height = AccessPointNode.nodeHeight;

    // 放在右侧，垂直居中
    const iconX = width - 12;
    const iconY = height / 2;

    // 1. 扩大鼠标感应区 (透明矩形)
    this.upsert('info-hit-area', GRect, {
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
      r: 6,
      stroke: '#979ba5', // 边框颜色
      lineWidth: 1.5,     // 边框粗细
      fill: '#979ba5',    // 圆圈内部填充白色，盖住下面的线或背景
      pointerEvents: 'none', // 让事件穿透到上面的 hit-area
    }, container);

    this.upsert('info-text', GText, {
      x: iconX,
      y: iconY,
      text: 'i',
      fontSize: 12,
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
    this.drawContainer(container);
    this.drawAPLabel(container);
    this.drawAccessPointName(container);
    this.drawInfoIcon(container);
  }
}
