import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import { letterAspectRatio } from './letter-aspect-ratio';

export interface INodeData {
  name: string;
}

export default class NetworkAreaNode extends BaseNode {
  // 默认外层节点属性
  static defaultNodeStyle: RectStyleProps = {
    fill: '#f5f7fa',
    height: 792,
    zIndex: 0,
  };

  // 默认文案样式
  static defaultLogoTextStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 12,
    lineHeight: 10,
    textAlign: 'end',
    fill: '#979BA5',
  };

  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;// 获取当前节点数据
  }

  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    return {
      ...super.getKeyStyle(attr),
      ...NetworkAreaNode.defaultNodeStyle,
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  private drawNodeLabel(container: Group) {
    const text = this.data.name;
    const textWidth = this.getStrCanvasWidth(text, 14);
    return this.upsert('label', Text, {
      ...NetworkAreaNode.defaultLogoTextStyle,
      x: textWidth,
      y: 34,
      text,
    }, container);
  }

  // 获取ASCII码对应的像素大小 todo 可能有性能问题
  private getStrCanvasWidth(str: string|number, fontSize: string|number = 10) {
    let len = 0;
    for (const letter of String(str)) {
      len += Number(fontSize) * (letterAspectRatio[letter] || 1);
    }
    return len;
  };

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);
    this.drawNodeLabel(container);
  }
}
