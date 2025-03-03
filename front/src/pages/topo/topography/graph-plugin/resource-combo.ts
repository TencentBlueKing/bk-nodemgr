import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Rect, Text } from '@antv/g';
import type { BaseComboStyleProps, NodeData } from '@antv/g6';
import { BaseCombo } from '@antv/g6';

export type IResourceComboProps = BaseComboStyleProps;

export interface IComboData {
  label: string;
}
export default class ResourceCombo extends BaseCombo {
  // 默认外层节点属性
  static defaultComboStyle: RectStyleProps = {
    height: 792,
    width: 280,
    stroke: '#F5F7FA',
    // stroke: '#000',
  };
  static readonly maxComboWidth = 520;
  static readonly maxComboCol = 2;
  static readonly maxComboRow = 4;
  static readonly heightSegment = (ResourceCombo.defaultComboStyle.height as number) / ResourceCombo.maxComboRow;

  static defaultLabelStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 12,
    lineHeight: 20,
    fill: '#979BA5',
    x: 12,
    y: 26,
  };

  get data(): IComboData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as IComboData;// 获取当前节点数据
  }

  protected getKeyStyle(attr: Required<IResourceComboProps>) {
    const { offsetX, offsetY } = this.getComboOffset(attr);
    const nodes = attr.childrenData;
    const width = this.getComboWidth(nodes) as number;
    const height = this.getComboHeight(nodes) as number;
    console.log('size', width, height)
    return {
      ...super.getKeyStyle(attr),
      ...ResourceCombo.defaultComboStyle,
      width,
      height,
      x: offsetX,
      y: offsetY,
    };
  }

  // 实现 drawKeyShape 方法
  protected drawKeyShape(attr: Required<IResourceComboProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  protected render(attr: Required<IResourceComboProps>, container: Group) {
    super.render(attr, container);
    this.drawComboLabel(attr, container);
  }

  protected drawComboLabel(attr: Required<IResourceComboProps>, container: Group) {
    const text = this.data.label;
    const { offsetX, offsetY } = this.getComboOffset(attr);
    const { x, y } = ResourceCombo.defaultLabelStyle;
    console.log('label', offsetX + (x as number), offsetY + (y as number))
    return this.upsert('label', Text, {
      ...ResourceCombo.defaultLabelStyle,
      x: offsetX + (x as number),
      y: offsetY + (y as number),
      text,
    }, container);
  }

  protected getComboWidth(nodes: NodeData[]) {
    const nodeCount = nodes.length;
    if (nodeCount >= 2) {
      return ResourceCombo.maxComboWidth;
    }
    return ResourceCombo.defaultComboStyle.width;
  }

  protected getComboHeight(nodes: NodeData[]) {
    let { height } = ResourceCombo.defaultComboStyle;
    const maxNodesInCombo = ResourceCombo.maxComboCol * ResourceCombo.maxComboRow;
    if (nodes.length > maxNodesInCombo) {
      height = Math.ceil(nodes.length / ResourceCombo.maxComboCol) * ResourceCombo.heightSegment;
    }
    return height;
  }

  protected getComboOffset(attr: Required<IResourceComboProps>) {
    const comboPosition = this.getComboPosition(attr);
    console.log('comboPosition', comboPosition)
    const width = this.getComboWidth(attr.childrenData) as number;
    const offsetX = (width / 2) * -1;
    const offsetY = (comboPosition[1]) * -1;
    return { offsetX, offsetY };
  }
}
