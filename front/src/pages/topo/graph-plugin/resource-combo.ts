import { BaseCombo } from "@antv/g6";
import { Rect, Text } from '@antv/g';
import type { RectStyleProps, Group, TextStyleProps } from '@antv/g';
import type { BaseComboStyleProps } from '@antv/g6';

export interface IResourceComboProps extends BaseComboStyleProps {}

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
  static maxDefaultWidth = 520;

  static defaultLabelStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 12,
    lineHeight: 20,
    fill: '#979BA5',
    x: 12,
    y: 26,
  };

  constructor (options) {
    super(options);
  }


  get data(): Partial<IComboData> {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as IComboData;// 获取当前节点数据
  }

  protected getKeyStyle(attr: Required<IResourceComboProps>) {
    const { width, offsetX, offsetY } = this.getComboOffset(attr);
    return { 
      ...super.getKeyStyle(attr),
      ...ResourceCombo.defaultComboStyle,
      width,
      x: offsetX,
      y: offsetY
    };
  }

  // 实现 drawKeyShape 方法
  protected drawKeyShape(attr: Required<IResourceComboProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  render(attr: Required<IResourceComboProps>, container: Group) {
    super.render(attr, container);
    this.drawComboLabel(attr, container);
  }

  protected drawComboLabel(attr: Required<IResourceComboProps>, container: Group) {
    const text = this.data.label;
    const { offsetX, offsetY } = this.getComboOffset(attr);
    const { x, y } = ResourceCombo.defaultLabelStyle;
    return this.upsert('label', Text, {
      ...ResourceCombo.defaultLabelStyle,
      x: offsetX + (x as number),
      y: offsetY + (y as number),
      text,
    }, container);
  }

  getComboWidth(nodes) {
    const nodeCount = nodes.length;
    if (nodeCount >= 2) {
      return ResourceCombo.maxDefaultWidth;
    }
    return ResourceCombo.defaultComboStyle.width;
  }

  getComboOffset(attr: Required<IResourceComboProps>) {
    const comboPosition = this.getComboPosition(attr);
    const width = this.getComboWidth(attr.childrenData) as number;
    const offsetX = (width / 2) * -1;
    const offsetY = (comboPosition[1]) * -1;
    return { width, offsetX, offsetY };
  }
}
