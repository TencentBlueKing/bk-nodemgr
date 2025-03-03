import type { Group, ImageStyleProps, RectStyleProps, TextStyleProps } from '@antv/g';
import { Image, Rect, Text } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import { iconPathMap } from './config';
import { letterAspectRatio } from './letter-aspect-ratio';

export interface INodeData {
  name: string;
  proxy: number;
  agent: number;
}

export type IResourceNodeProps = BaseNodeStyleProps;

type Theme = 'primary' | 'warning';
interface ThemeConfig {
  text: string;
  color: string;
  backgroundColor: string;
  x: number;
  y: number;
}
const THEME_CONFIG: Record<Theme, ThemeConfig> = {
  primary: {
    text: 'Proxy:',
    color: '#1768EF',
    backgroundColor: '#E1ECFF',
    x: 10,
    y: 70,
  },
  warning: {
    text: 'Agent:',
    color: '#E38B02',
    backgroundColor: '#FDEED8',
    x: 10,
    y: 90,
  },
};

export default class NetWorkUnitNode extends BaseNode {
  // 默认Logo样式
  static defaultLogoStyle: ImageStyleProps = {
    width: 24,
    height: 24,
    x: 20,
    y: 20,
    zIndex: 2,
  };
  // 默认Logo文案样式
  static defaultLogoTextStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 14,
    lineHeight: 10,
    fontWeight: 700,
    // x: 5,
    y: 80,
    textAlign: 'end',
    fill: '#4D4F56',
    zIndex: 2,
  };
  // 默认外层节点属性
  static defaultNodeStyle: RectStyleProps = {
    stroke: '#000',
    shadowColor: '#dee0e5',
    shadowBlur: 15,
    shadowType: 'outer',
    fill: '#fff',
    cursor: 'pointer',
    fillOpacity: 1,
    height: 64,
    width: 64,
    radius: 200,
    zIndex: 2,
  };

  // 默认tag样式
  static defaultTagStyle: Omit<RectStyleProps, 'width'> = {
    height: 16,
    radius: 8,
    x: 12,
    zIndex: 2,
  };

  static defaultTagTextStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 8,
    x: 10,
    lineHeight: 22,
    zIndex: 2,
  };

  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;// 获取当前节点数据
  }

  protected getKeyStyle(attr: Required<IResourceNodeProps>) {
    return {
      ...super.getKeyStyle(attr),
      ...NetWorkUnitNode.defaultNodeStyle,
    };
  }

  // 重写方法
  protected drawKeyShape(attr: Required<IResourceNodeProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<IResourceNodeProps>, container: Group) {
    const { agent, proxy } = this.data;
    super.render(attr, container);
    this.drawNodeCenterLogo(container);
    this.drawNodeLabel(container);
    this.drawNodeTag(container, 'primary', proxy);
    this.drawNodeTag(container, 'warning', agent);
  }

  private drawNodeCenterLogo(container: Group) {
    return this.upsert('logo', Image, {
      ...NetWorkUnitNode.defaultLogoStyle,
      src: iconPathMap.Machine,
    }, container);
  }

  private drawNodeLabel(container: Group) {
    const text = this.data.name;
    const textWidth = this.getStrCanvasWidth(text, 14);
    const width = NetWorkUnitNode.defaultNodeStyle.width as number;
    const offsetX = (width - textWidth) / 2;
    return this.upsert('label', Text, {
      ...NetWorkUnitNode.defaultLogoTextStyle,
      x: textWidth + offsetX,
      text,
    }, container);
  }

  private drawNodeTag(container: Group, theme: 'primary' | 'warning', count: number) {
    const config = THEME_CONFIG[theme];
    const text = `${config.text} ${count}`;
    const { x, height, radius } = NetWorkUnitNode.defaultTagStyle;
    const { fontSize, lineHeight } = NetWorkUnitNode.defaultTagTextStyle;
    const paddingX = 12;
    const width = this.getStrCanvasWidth(text, fontSize) + paddingX;
    this.upsert(`label${theme}`, Rect, {
      x,
      y: config.y,
      height,
      width,
      radius,
      fill: config.backgroundColor,
    }, container);

    this.upsert(`theme${theme}`, Text, {
      x: config.x + 8,
      y: config.y + 20,
      lineHeight,
      fontSize,
      fill: config.color,
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
}
