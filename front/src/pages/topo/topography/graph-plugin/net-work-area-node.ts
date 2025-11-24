import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

export interface INodeData {
  name: string;
  bk_networkarea_id?: number;
  bk_networkarea_name?: string;
  width: number;
  height: number;
}

export default class NetworkAreaNode extends BaseNode {
  // 区域节点样式：无标题栏、浅背景、无边框
  static defaultNodeStyle: Partial<RectStyleProps> = {
    fill: '#f0f0f0', // 浅灰色背景
    stroke: 'transparent', // 无边框
    radius: 4,
    zIndex: -1, // 层级低于内部节点
    width: 300,
    height: 200,
  };

  // 区域名称样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 14,
    fill: '#333',
    textBaseline: 'middle',
    fontWeight: 'bold',
  };

  // 区域标识样式（替换NA为“区域”文字）
  static defaultAreaLabelStyle: Partial<RectStyleProps> = {
    width: 40,
    height: 24,
    fill: '#14a568',
    radius: 4,
  };

  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;
  }

  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    const { width, height } = this.data;
    return {
      ...super.getKeyStyle(attr),
      ...NetworkAreaNode.defaultNodeStyle,
      width,
      height,
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  // 绘制“区域”标识（修复坐标，确保文字显示）
  private drawAreaLabel(container: Group) {
    const labelStyle = NetworkAreaNode.defaultAreaLabelStyle;
    const labelX = 20; // 标识左上角x坐标（区域内边距10px）
    const labelY = 20; // 标识左上角y坐标（区域内边距10px）

    // 标识背景矩形（位置：区域左上角，边距10px）
    this.upsert('area-label-bg', Rect, {
      ...labelStyle,
      x: labelX,
      y: labelY,
    }, container);

    // “区域”文字（位置：背景矩形正中央）
    this.upsert('area-label-text', Text, {
      x: labelX + (labelStyle.width as number) / 2, // 矩形水平中心
      y: labelY + (labelStyle.height as number) / 2, // 矩形垂直中心
      text: '区域',
      fontSize: 12,
      fill: '#ffffff',
      fontWeight: 'bold',
      textAlign: 'center',
      textBaseline: 'middle',
    }, container);

    return container;
  }

  // 绘制区域名称（位置：“区域”标识右侧，垂直居中）
  private drawAreaInfo(container: Group) {
    const { width, height } = this.data;
    const { name, bk_networkarea_name, bk_networkarea_id } = this.data;
    const labelStyle = NetworkAreaNode.defaultAreaLabelStyle;

    const areaName = bk_networkarea_name || name;
    const areaId = bk_networkarea_id || 0;
    const areaText = `#${areaId} ${areaName}`;

    // 名称位置：标识右侧10px，与标识垂直居中
    const textX = 20 + (labelStyle.width as number) + 10; // 10(边距)+40(标识宽)+10(间距)
    const textY = 20 + (labelStyle.height as number) / 2; // 与标识垂直居中

    return this.upsert('area-info', Text, {
      x: textX,
      y: textY,
      text: areaText,
      ...NetworkAreaNode.defaultTextStyle,
    }, container);
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);
    this.drawAreaLabel(container); // 绘制“区域”标识
    this.drawAreaInfo(container); // 绘制区域名称
  }
}
