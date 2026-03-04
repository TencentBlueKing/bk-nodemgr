import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
import { Image as GImage, Rect as GRect, Text as GText } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import Area from '../../../../../public/images/area.svg';
import { textTooltip } from './text-tooltip.js';

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
    fill: '#F0F1F5', // 浅灰色背景
    stroke: 'transparent', // 无边框
    radius: 4,
    zIndex: -1, // 层级低于内部节点
    width: 260,
    height: 300,
  };

  // 区域名称样式
  static defaultTextStyle: Partial<TextStyleProps> = {
    fontSize: 14,
    fill: '#333',
    textBaseline: 'middle',
    fontWeight: '',
  };

  // 区域标识样式（替换NA为"区域"文字）
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
    return this.upsert('key', GRect, this.getKeyStyle(attr), container);
  }

  // 绘制"区域"标识（修复坐标，确保文字显示）
  private drawAreaLabel(container: Group) {
    const labelX = 18; // 标识左上角x坐标（区域内边距10px）
    const labelY = 16; // 标识左上角y坐标（区域内边距10px）

    // 标识背景矩形（位置：区域左上角，边距10px）
    this.upsert('area-logo', GImage, {
      x: labelX,
      y: labelY,
      src: Area,
    }, container);

    return container;
  }

  // 绘制区域名称（位置："区域"标识右侧，垂直居中）
  private drawAreaInfo(container: Group) {
    const { name, bk_networkarea_name, bk_networkarea_id, width } = this.data;

    const areaName = bk_networkarea_name || name;
    const areaId = bk_networkarea_id || 0;
    const areaText = `#${areaId} ${areaName}`;

    // 名称位置：标识右侧10px，与标识垂直居中
    const textX = 42; // 10(边距)+40(标识宽)+10(间距)
    const textY = 24; // 与标识垂直居中
    
    // 计算可用宽度：区域总宽度 - 左边距 - 标识宽度 - 间距 - 右边距 - 额外间距
    const maxWidth = (width || 260) - 42 - 20 - 15;

    const textElement = this.upsert('area-info', GText, {
      x: textX,
      y: textY,
      text: areaText,
      ...NetworkAreaNode.defaultTextStyle,
      wordWrap: true,
      wordWrapWidth: maxWidth,
      maxLines: 1,
      textOverflow: 'ellipsis',
      cursor: 'default',
    }, container);

    // 添加 hover 事件监听
    if (textElement) {
      textElement.addEventListener('mouseenter', (evt: any) => {
        const mouseEvent = evt.client;
        if (mouseEvent) {
          // 获取文本元素的中心坐标
          const bounds = textElement.getRenderBounds();
          const centerY = (bounds.min[1] + bounds.max[1]) / 2;
          textTooltip.show(areaText, centerY, mouseEvent.x, mouseEvent.y);
        }
      });

      textElement.addEventListener('mouseleave', () => {
        textTooltip.hide();
      });
    }

    return textElement;
  }

  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);
    this.drawAreaLabel(container); // 绘制"区域"标识
    this.drawAreaInfo(container); // 绘制区域名称
  }
}
