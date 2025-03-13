import { BaseNode } from "@antv/g6";
import { Rect, Image, Text } from '@antv/g';
import type { RectStyleProps, Group, ImageStyleProps, TextStyleProps } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { iconPathMap } from "./config";
import { letterAspectRatio } from "./letter-aspect-ratio";

export interface INodeData {
  name: string;
  proxy: number;
  agent: number;
}

export interface IResourceNodeProps extends BaseNodeStyleProps {}

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

export default class ResourceNode extends BaseNode {
  // 默认Logo样式
  static defaultLogoStyle: ImageStyleProps = {
    width: 24,
    height: 24,
    x: 20,
    y: 20,
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
  };

  // 默认tag样式
  static defaultTagStyle: Omit<RectStyleProps, 'width'> = {
    height: 16,
    radius: 8,
    x: 4
  };

  static defaultTagTextStyle: Omit<TextStyleProps, 'text'> = {
    fontSize: 8,
    x: 10,
    lineHeight: 22,
  }

  get data(): Partial<INodeData> {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;// 获取当前节点数据
  }

  protected getKeyStyle(attr: Required<IResourceNodeProps>) {
    return { 
      ...super.getKeyStyle(attr),
      ...ResourceNode.defaultNodeStyle,
    };
  }

  // 重写方法
  protected drawKeyShape(attr: Required<IResourceNodeProps>, container: Group) {
    return this.upsert('key', Rect, this.getKeyStyle(attr), container);
  }

  render(attr: Required<IResourceNodeProps>, container: Group) {
    super.render(attr, container);
    this.drawNodeCenterLogo(container);
    this.drawNodeLabel(attr, container);
    this.drawNodeTag(container, 'primary', this.data.proxy);
    this.drawNodeTag(container, 'warning', this.data.agent);
  }

  private drawNodeCenterLogo(container: Group) {
    return this.upsert('logo', Image, {
      ...ResourceNode.defaultLogoStyle,
      src: iconPathMap.Machine,
    }, container);
  }

  drawNodeLabel(attr: Required<IResourceNodeProps>, container: Group) {
    const textWidth = this.getStrCanvasWidth(this.data.name, 14);
    const offsetX = (ResourceNode.defaultNodeStyle.width - textWidth) / 2;
    return this.upsert('label', Text, {
      ...ResourceNode.defaultLogoTextStyle,
      x: textWidth + offsetX,
      text: this.data.name,
    }, container);
  }

  drawNodeTag(container: Group, theme: 'primary' | 'warning', count: number) {
    const config = THEME_CONFIG[theme];
    const text = `${config.text} ${count}`;
    const { x, height, radius } = ResourceNode.defaultTagStyle;
    const { fontSize, lineHeight } = ResourceNode.defaultTagTextStyle;
    const paddingX = 12;
    const width = this.getStrCanvasWidth(text, fontSize) + paddingX;
    this.upsert('label' + theme, Rect, {
      x,
      y: config.y,
      height,
      width,
      radius,
      fill: config.backgroundColor,
    }, container);

    this.upsert('theme' + theme, Text, {
      x: config.x,
      y: config.y + 20,
      lineHeight,
      fontSize,
      fill: config.color,
      text,
    }, container);
  }

  // 获取ASCII码对应的像素大小 todo 可能有性能问题
  getStrCanvasWidth(str: string|number, fontSize: string|number = 10) {
    let len = 0;
    for (const letter of String(str)) {
      len += Number(fontSize) * (letterAspectRatio[letter] || 1);
    }
    return len;
  };
}