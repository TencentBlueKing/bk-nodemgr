import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import { NodeType } from './config';

export interface IServerData {
  name: string;
  area?: string;
}

export default class ServerNode extends BaseNode {
  // Server节点尺寸
  static serverWidth = 60;
  static serverHeight = 200;

  // 默认节点样式
  static defaultNodeStyle: RectStyleProps = {
    stroke: '#c4c6cc',
    strokeWidth: 1,
    fill: '#f0f1f5',
    cursor: 'pointer',
    fillOpacity: 1,
    width: ServerNode.serverWidth,
    height: ServerNode.serverHeight,
    radius: 4,
  };

  // 默认文案样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 12,
    fill: '#313238',
    textAlign: 'center',
    textBaseline: 'middle',
  };

  get data(): IServerData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as IServerData;
  }

  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    return {
      ...super.getKeyStyle(attr),
      ...ServerNode.defaultNodeStyle,
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', 'rect', this.getKeyStyle(attr), container);
  }

  // 绘制Server文字
  private drawServerText(container: Group) {
    const width = ServerNode.serverWidth;
    const height = ServerNode.serverHeight;

    return this.upsert('server-text', 'text', {
      x: width / 2,
      y: height / 2,
      text: 'Server',
      ...ServerNode.defaultTextStyle,
      fontWeight: 600,
    }, container);
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);
    this.drawServerText(container);
  }
}
