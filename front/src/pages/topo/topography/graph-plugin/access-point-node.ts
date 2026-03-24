import type { Group, RectStyleProps, TextStyleProps } from '@antv/g';
// 确保引入了 Rect 和 Text
import { Image as GImage, Rect as GRect, Text as GText } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import Downstream from '../../../../../public/images/downstream.svg';
import { tableTooltip } from './table-tooltip';

export interface IAccessPointData {
  name: string;
  downstreamUnits: number;
  type: 'internal' | 'external'; // 内网/外网标识
  area?: string;
  bk_networkunit_id?: number;
  endpoints?: any;
  endpointsData?: any[];
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
      // cursor: 'move',// 暂时把拖动功能注释
      cursor: 'text',
    }, container);
  }

  // 接入点标识图标
  private drawAPLabel(container: Group) {
    const height = AccessPointNode.nodeHeight;

    this.upsert('downstream-ap', GImage, {
      x: 10,
      y: (height - 10) / 2, // 10是图标的高度
      src: Downstream,
      // cursor: 'move',// 暂时把拖动功能注释
      cursor: 'text',
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
      // cursor: 'move', // 暂时把拖动功能注释
      cursor: 'text',
    }, container);
  }

  // 【新增】绘制信息图标 (i)
  private drawInfoIcon(container: Group) {
    const width = this.getCalculatedWidth;
    const height = AccessPointNode.nodeHeight;

    // 放在右侧，垂直居中
    const iconX = width - 12;
    const iconY = height / 2;

    // 1. 先绘制图标圆圈 + 文字（底层）
    this.upsert('info-circle', 'circle', {
      cx: iconX,
      cy: iconY,
      r: 6,
      stroke: '#979ba5',
      lineWidth: 1.5,
      fill: '#979ba5',
      cursor: 'pointer',
      pointerEvents: 'none',
    }, container);

    this.upsert('info-text', GText, {
      x: iconX,
      y: iconY,
      text: 'i',
      fontSize: 12,
      fill: '#ffff',
      fontWeight: 'bold',
      fontFamily: 'serif',
      textAlign: 'center',
      textBaseline: 'middle',
      cursor: 'pointer',
      pointerEvents: 'none',
    }, container);

    // 2. 最后绘制感应区（最上层），确保 cursor: pointer 生效
    const hitAreaElement = this.upsert('info-hit-area', GRect, {
      x: width - 25,
      y: 0,
      width: 25,
      height,
      fill: 'transparent',
      cursor: 'pointer',
    }, container);

    // 3. 只绑定一次事件，避免 render 重复调用导致多个监听器叠加
    if (hitAreaElement && !(hitAreaElement as any).__bindTooltip) {
      (hitAreaElement as any).__bindTooltip = true;

      hitAreaElement.addEventListener('mouseenter', () => {
        // 每次进入时重新读取 data，确保数据最新
        const { endpointsData } = this.data;
        if (!endpointsData || endpointsData.length === 0) return;

        const infoCircle = this.shapeMap['info-circle'];
        const target = infoCircle || hitAreaElement;
        const bounds = target.getRenderBounds();
        const graph = this.context.graph;

        if (graph) {
          const topLeft = graph.getViewportByCanvas([bounds.min[0], bounds.min[1]]);
          const bottomRight = graph.getViewportByCanvas([bounds.max[0], bounds.max[1]]);

          const canvasDom = graph.getCanvas?.()?.getContextService?.()?.getDomElement?.();
          const canvasRect = canvasDom?.getBoundingClientRect?.();
          const offsetX = canvasRect?.left ?? 0;
          const offsetY = canvasRect?.top ?? 0;

          const targetX = offsetX + topLeft[0];
          const targetY = offsetY + topLeft[1];
          const targetWidth = bottomRight[0] - topLeft[0];
          const targetHeight = bottomRight[1] - topLeft[1];
          tableTooltip.show(endpointsData, targetX, targetY, targetWidth, targetHeight);
        }
      });

      hitAreaElement.addEventListener('mouseleave', () => {
        tableTooltip.hide();
      });
    }
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
