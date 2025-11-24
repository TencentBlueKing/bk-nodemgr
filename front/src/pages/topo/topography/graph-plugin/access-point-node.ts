import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
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
    }, container);
  }

  private drawAPLabel(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const labelStyle = AccessPointNode.defaultAPLabelStyle;
    const { type } = this.data; // type为'internal'/'external'

    // 标识背景矩形
    this.upsert('ap-label-bg', 'rect', {
      ...labelStyle,
      x: 10,
      y: (height - (labelStyle.height as number)) / 2,
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
    }, container);

    return container;
  }

  // 绘制接入点名称
  private drawAccessPointName(container: Group) {
    const height = AccessPointNode.nodeHeight;
    const { name, bk_networkunit_id } = this.data;

    return this.upsert('access-point-name', 'text', {
      x: 60, // 跟在标识后面
      y: height / 2,
      text: name,
      ...AccessPointNode.defaultTextStyle,
      fontWeight: 600,
      maxWidth: 150,
      ellipsis: true,
    }, container);
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);

    // 绘制顶部栏、标识、名称
    this.drawHeader(container);
    this.drawAPLabel(container);
    this.drawAccessPointName(container);
  }
}
