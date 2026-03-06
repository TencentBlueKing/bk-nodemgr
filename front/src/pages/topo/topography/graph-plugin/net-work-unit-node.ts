import type { Group } from '@antv/g';
// 引入需要的图形 Shape
import { Circle as GCircle, Image as GImage, Line as GLine, Rect as GRect, Text as GText } from '@antv/g';
import type { BaseNodeStyleProps } from '@antv/g6';
import { BaseNode } from '@antv/g6';

import JumpLink from '../../../../../public/images/jump-link.svg';
import More from '../../../../../public/images/more.svg';
import VectorDirect from '../../../../../public/images/vector-direct.svg';
import VectorIndirect from '../../../../../public/images/vector-indirect.svg';
import ConnectionPoint from '../../../../../public/images/connection-point.svg';
import { textTooltip } from './text-tooltip.js';
import { i18n } from '@/modules/i18n';

export interface INodeData {
  name: string;
  running_proxy: number;
  total_proxy: number;
  running_agent: number;
  total_agent: number;
  cycle_times: string[];
  is_healthy: Boolean;
  area?: string;
  is_direct?: boolean;
  direct_endpoints?: any;
  accesspoints?: any[];
  links?: any;
}

interface SplitTextConfig {
  prefix: string;
  suffix: string;
  prefixColor: string;
}

export default class NetWorkUnitNode extends BaseNode {
  // --- 1. 尺寸配置 ---
  static gridHeight = 50;   // 单个格子的高度
  static headerHeight = 40; // 标题栏高度 (改为40px)

  // 容器内边距 (关键：让表格和Agent看起来在内部)
  static paddingX = 12;     // 左右内边距
  static paddingY = 12;     // 上下内边距
  static contentGap = 12;   // 表格和Agent之间的间距

  static agentWidth = 216;
  static agentBarHeight = 32;
  static badgeRadius = 2;

  // --- 颜色配置 ---
  static directUnitHeaderBgColor = '#FDEED8';
  static IndirectUnitHeaderBgColor = '#E1ECFF';
  static agentBgColor = '#F5F7FA';
  static linkColor = '#3A84FF';
  static normalColor = '#4D4F56';
  static titleColor = '#000000';
  static greenColor = '#2DCB56';
  static redColor = '#EA3636';
  static gray = '#979ba5';
  static warnColor = '#FFB848';
  static badgeColorP = '#3A84FF';
  static badgeColorA = '#8B5CF6';
  static borderColor = '#E1E4E8';


  // 节点总宽度
  static nodeWidth = 240;

  // 获取节点数据
  get data(): INodeData {
    return this.context.model.getNodeLikeDatum(this.id)?.data as unknown as INodeData;
  }

  // 主容器样式 - 整体背景块
  protected getKeyStyle(attr: Required<BaseNodeStyleProps>) {
    const { is_direct } = this.data;
    const width = NetWorkUnitNode.nodeWidth;
    // 直连：40 + 120 = 160px，非直连：40 + 172 = 212px
    const height = is_direct ? 160 : 212;

    return {
      ...super.getKeyStyle(attr),
      width,
      height,
      fill: '#ffffff',
      stroke: 'none',
      strokeWidth: 0,
      radius: 6,
      shadowColor: 'rgba(0, 0, 0, 0.06)',
      shadowBlur: 8,
      shadowOffsetY: 2,
      cursor: 'text',
    };
  }

  protected drawKeyShape(attr: Required<BaseNodeStyleProps>, container: Group) {
    return this.upsert('key', 'rect', this.getKeyStyle(attr), container);
  }

  // 绘制标题栏块
  private drawHeaderBlock(container: Group) {
    const width = NetWorkUnitNode.nodeWidth;
    const height = NetWorkUnitNode.headerHeight;
    const { directUnitHeaderBgColor, IndirectUnitHeaderBgColor } = NetWorkUnitNode;
    const { is_direct } = this.data;
    
    this.upsert('header-block', 'rect', {
      x: 0,
      y: 0,
      width,
      height,
      fill: is_direct ? directUnitHeaderBgColor : IndirectUnitHeaderBgColor,
      radius: [6, 6, 0, 0],
      // cursor: 'move', // 暂时把拖动功能注释
      cursor: 'text',
    }, container);
  }

  // 绘制内容块 - 白色圆角矩形
  private drawContentBlock(container: Group) {
    const { is_direct } = this.data;
    const width = NetWorkUnitNode.nodeWidth;
    // 直连内容区域：120px，非直连：172px
    const contentHeight = is_direct ? 120 : 172;

    this.upsert('content-block', 'rect', {
      x: 0,
      y: NetWorkUnitNode.headerHeight - 8, // 向上偏移8px与header重叠
      width,
      height: contentHeight + 8, // 补偿向上偏移的8px
      fill: '#ffffff',
      radius: 6, // 完整圆角
      cursor: 'text',
    }, container);
  }

  private drawUnitLabel(container: Group) {
    const { is_direct } = this.data;
    // 标识图标
    this.upsert('unit-icon-direct', GImage, {
      x: 13,
      y: 8,
      src: is_direct ? VectorDirect : VectorIndirect,
      // cursor: 'move', // 暂时把拖动功能注释
    }, container);
  }

  private drawUnitName(container: Group) {
    const { name } = this.data;

    // 计算可用宽度：节点宽度 - 左边距 - 图标宽度 - 间距 - 菜单图标区域宽度 - 额外间距
    const maxWidth = NetWorkUnitNode.nodeWidth - 34 - 30 - 15; // 34是图标+间距，30是右侧菜单区域，15是额外间距

    const textElement = this.upsert('unit-name', GText, {
      x: 34,
      y: 16,
      text: name,
      fontSize: 14,
      fill: '#313238',
      textBaseline: 'middle',
      fontWeight: 700,
      cursor: 'pointer',
      wordWrap: true,
      wordWrapWidth: maxWidth,
      maxLines: 1,
      textOverflow: 'ellipsis',
    }, container);

    // 添加 hover 事件监听
    if (textElement) {
      textElement.addEventListener('mouseenter', (event: any) => {
        const bounds = textElement.getRenderBounds();
        // 使用鼠标事件坐标作为基准
        const mouseX = event.clientX;
        const mouseY = event.clientY;
        
        // 计算元素的宽高
        const targetWidth = bounds.max[0] - bounds.min[0];
        const targetHeight = bounds.max[1] - bounds.min[1];
        
        // 使用鼠标位置作为目标中心点，计算目标的起始位置
        const targetX = mouseX - targetWidth / 2;
        const targetY = mouseY - targetHeight / 2;
        
        textTooltip.show(name, targetX, targetY, targetWidth, targetHeight);
      });

      textElement.addEventListener('mouseleave', () => {
        textTooltip.hide();
      });
    }

    return textElement;
  }

  // --- 内容块绘制 ---

  private drawProxyTable(container: Group) {
    const {
      is_direct, accesspoints,
      running_proxy, total_proxy, cycle_times, is_healthy,
    } = this.data;

    // 计算表格起始位置 (在 Padding 内部)
    const startX = NetWorkUnitNode.paddingX;
    const startY = NetWorkUnitNode.headerHeight + NetWorkUnitNode.paddingY;
    const isZh = i18n.global.locale.value === 'zh-CN';
    
    // 中文：label宽度54px (两个汉字+间距), 英文：label宽度95px (最长的"Access Points")
    const labelWidth = isZh ? 54 : 95;
    const valueX = startX + labelWidth; // 值统一从这个位置开始

    if (is_direct) {
      // 直连
      // Row 1
      this.upsert('Server', GText, {
        x: startX,
        y: startY + 5,
        text: 'Server',
        fontSize: 12,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        fontWeight: 700,
        cursor: 'text',
      }, container);

      const linkIconUnitElement = this.upsert('linkIcon-unit-direct', GImage, {
        x: startX + NetWorkUnitNode.nodeWidth - 43,
        y: startY - 1,
        width: 11,
        height: 11,
        src: JumpLink,
        cursor: 'pointer',
      }, container);

      // 添加 hover 事件监听
      if (linkIconUnitElement) {
        linkIconUnitElement.addEventListener('mouseenter', (event: any) => {
          const bounds = linkIconUnitElement.getRenderBounds();
          const mouseX = event.clientX;
          const mouseY = event.clientY;
          
          const targetWidth = bounds.max[0] - bounds.min[0];
          const targetHeight = bounds.max[1] - bounds.min[1];
          const targetX = mouseX - targetWidth / 2;
          const targetY = mouseY - targetHeight / 2;
          
          textTooltip.show('unit', targetX, targetY, targetWidth, targetHeight, 30);
        });

        linkIconUnitElement.addEventListener('mouseleave', () => {
          textTooltip.hide();
        });
      }
      // Row 2 - 使用统一对齐
      const accessPointsText = i18n.global.t('topoManager.topo.node.accessPoints');
      
      this.upsert('accesspoints-title', GText, {
        x: startX,
        y: startY + 37.5,
        text: `${accessPointsText} :`,
        fontSize: 12,
        fill: NetWorkUnitNode.normalColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
      this.upsert('accesspoints', GText, {
        x: valueX,
        y: startY + 37.5,
        text: `${accesspoints?.length || 0} ${i18n.global.t('topoManager.topo.node.count')}`,
        fontSize: 12,
        fontWeight: 700,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
    } else {
      // 非直连
      // Row 1
      this.upsert('Proxy', GText, {
        x: startX,
        y: startY + 5,
        text: 'Proxy',
        fontSize: 12,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        fontWeight: 700,
        cursor: 'text',
      }, container);

      const stateWidth = isZh ? 32 : 60; // 中文32px，英文60px
      this.upsert('state-bg', 'rect', {
        x: startX + 42,
        y: startY - 3,
        width: stateWidth,
        height: 16,
        fill: is_healthy ? NetWorkUnitNode.greenColor : NetWorkUnitNode.redColor,
        radius: NetWorkUnitNode.badgeRadius,
        cursor: 'text',
      }, container);
      this.upsert('state', GText, {
        x: startX + 42 + stateWidth / 2, // 在状态框中居中显示
        y: startY + 5,
        text: is_healthy ? i18n.global.t('topoManager.topo.node.healthy') : i18n.global.t('topoManager.topo.node.abnormal'),
        fontSize: 10,
        fill: '#FFFFFF',
        textAlign: 'center',
        textBaseline: 'middle',
        fontWeight: 400,
        cursor: 'text',
      }, container);

      const linkIconElement = this.upsert('linkIcon-unit-proxy', GImage, {
        x: startX + NetWorkUnitNode.nodeWidth - 43,
        y: startY - 1,
        width: 11,
        height: 11,
        src: JumpLink,
        cursor: 'pointer',
      }, container);

      // 添加 hover 事件监听
      if (linkIconElement) {
        linkIconElement.addEventListener('mouseenter', (event: any) => {
          const bounds = linkIconElement.getRenderBounds();
          const mouseX = event.clientX;
          const mouseY = event.clientY;
          
          const targetWidth = bounds.max[0] - bounds.min[0];
          const targetHeight = bounds.max[1] - bounds.min[1];
          const targetX = mouseX - targetWidth / 2;
          const targetY = mouseY - targetHeight / 2;
          
          textTooltip.show('unit', targetX, targetY, targetWidth, targetHeight, 30);
        });

        linkIconElement.addEventListener('mouseleave', () => {
          textTooltip.hide();
        });
      }
      // Row 2
      const quantityText = i18n.global.t('topoManager.topo.node.quantity');
      
      if (isZh) {
        // 中文：分开显示 "数" 和 "量" 以对齐
        this.upsert('proxy-count-prefix', GText, {
          x: startX,
          y: startY + 37.5,
          text: '数',
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
        this.upsert('proxy-count-suffix', GText, {
          x: startX + 24,
          y: startY + 37.5,
          text: '量 :',
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
      } else {
        // 英文：完整显示
        this.upsert('proxy-count-label', GText, {
          x: startX,
          y: startY + 37.5,
          text: `${quantityText} :`,
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
      }
      
      this.upsert('proxy-count', GText, {
        x: valueX,
        y: startY + 37.5,
        text: `${running_proxy} / ${total_proxy}`,
        fontSize: 12,
        fontWeight: 700,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
      // Row 3
      const delayText = i18n.global.t('topoManager.topo.node.delay');
      
      if (isZh) {
        // 中文：分开显示 "延" 和 "迟" 以对齐
        this.upsert('proxy-cycle-prefix', GText, {
          x: startX,
          y: startY + 63.5,
          text: '延',
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
        this.upsert('proxy-cycle-suffix', GText, {
          x: startX + 24,
          y: startY + 63.5,
          text: '迟 :',
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
      } else {
        // 英文：完整显示
        this.upsert('proxy-cycle-label', GText, {
          x: startX,
          y: startY + 63.5,
          text: `${delayText} :`,
          fontSize: 12,
          fill: NetWorkUnitNode.normalColor,
          textBaseline: 'middle',
          cursor: 'text',
        }, container);
      }
      
      this.upsert('proxy-cycle', GText, {
        x: valueX,
        y: startY + 63.5,
        text: `${cycle_times}`,
        fontSize: 12,
        fontWeight: 700,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
      // Row 4
      this.upsert('accesspoints-title', GText, {
        x: startX,
        y: startY + 89.5,
        text: `${i18n.global.t('topoManager.topo.node.accessPoints')} :`,
        fontSize: 12,
        fill: NetWorkUnitNode.normalColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
      this.upsert('accesspoints', GText, {
        x: valueX,
        y: startY + 89.5,
        text: `${accesspoints?.length || 0} ${i18n.global.t('topoManager.topo.node.count')}`,
        fontSize: 12,
        fontWeight: 700,
        fill: NetWorkUnitNode.titleColor,
        textBaseline: 'middle',
        cursor: 'text',
      }, container);
    }
  }

  private drawAgentBlock(container: Group) {
    const { is_direct, running_agent, total_agent } = this.data;

    // 计算 Agent 块位置
    const startY = NetWorkUnitNode.headerHeight + (is_direct ? 67.5 : 127);
    const startX = NetWorkUnitNode.paddingX;
    const width = NetWorkUnitNode.agentWidth;
    const height = NetWorkUnitNode.agentBarHeight;

    // 背景框
    this.upsert('agent-bg', 'rect', {
      x: startX,
      y: startY,
      width,
      height,
      fill: NetWorkUnitNode.agentBgColor,
      strokeWidth: 1,
      radius: NetWorkUnitNode.badgeRadius,
      cursor: 'text',
    }, container);

    // 内容文本 (居中)
    const centerY = startY + height / 2;
    // Label
    this.upsert('agent-lbl', GText, {
      x: startX + 12,
      y: centerY,
      text: 'Agent  :',
      fontSize: 12,
      fill: NetWorkUnitNode.normalColor,
      textBaseline: 'middle',
      fontWeight: 400,
      cursor: 'text',
    }, container);
    // 数量
    const prefix = `${running_agent}`;
    const suffix = `${total_agent}`;
    const textColor = running_agent > 0 ? NetWorkUnitNode.greenColor : NetWorkUnitNode.normalColor;

    this.upsert('agent-val-pre', GText, { x: startX + 66, y: centerY, text: prefix, fill: textColor, fontSize: 12, textBaseline: 'middle', cursor: 'text', }, container);
    this.upsert('agent-val-slash', GText, { x: startX + 74, y: centerY, text: '/', fill: NetWorkUnitNode.normalColor, fontSize: 12, textBaseline: 'middle', cursor: 'text', }, container);
    this.upsert('agent-val-suf', GText, { x: startX + 80, y: centerY, text: suffix, fill: NetWorkUnitNode.normalColor, fontSize: 12, textBaseline: 'middle', cursor: 'text', }, container);

    const linkIconAgentElement = this.upsert('linkIcon-agent', GImage, {
      x: startX + NetWorkUnitNode.nodeWidth - 43,
      y: centerY - 5,
      width: 11,
      height: 11,
      src: JumpLink,
      cursor: 'pointer',
    }, container);

    // 添加 hover 事件监听
    if (linkIconAgentElement) {
      linkIconAgentElement.addEventListener('mouseenter', (event: any) => {
        const bounds = linkIconAgentElement.getRenderBounds();
        const mouseX = event.clientX;
        const mouseY = event.clientY;
        
        const targetWidth = bounds.max[0] - bounds.min[0];
        const targetHeight = bounds.max[1] - bounds.min[1];
        const targetX = mouseX - targetWidth / 2;
        const targetY = mouseY - targetHeight / 2;
        
        textTooltip.show('agent', targetX, targetY, targetWidth, targetHeight, 30);
      });

      linkIconAgentElement.addEventListener('mouseleave', () => {
        textTooltip.hide();
      });
    }
  }

  private drawMenuIcon(container: Group) {
    const iconX = NetWorkUnitNode.nodeWidth - 28;
    const iconY = NetWorkUnitNode.headerHeight / 2;

    this.upsert('menu-hit-area', 'rect', {
      x: iconX - 10,
      y: 0,
      width: 30,
      height: NetWorkUnitNode.headerHeight,
      fill: 'transparent',
      cursor: 'pointer',
      zIndex: 50,
      className: 'menu-hit-area', // 关键：用于事件识别
    }, container);

    this.upsert('menu-icon', GImage, {
      x: iconX,
      y: iconY - 8.5,
      src: More,
      cursor: 'pointer',
      zIndex: 50,
    }, container);
  }

  // 绘制连接点图标
  private drawConnectionPoint(container: Group) {
    const { is_direct } = this.data;
    const iconSize = 18; // 图标大小
    const totalHeight = is_direct ? 160 : 212;
    
    // 计算位置：底部中间
    const iconX = (NetWorkUnitNode.nodeWidth - iconSize) / 2; // 水平居中
    const iconY = totalHeight - iconSize / 2; // Y位置

    // 白色背景圆形
    const bgRadius = iconSize / 2 - 1; // 背景半径比图标大一点
    const centerX = NetWorkUnitNode.nodeWidth / 2; // 圆心X
    const centerY = totalHeight - iconSize / 2 + iconSize / 2; // 圆心Y

    this.upsert('connection-point-bg', GCircle, {
      cx: centerX,
      cy: centerY,
      r: bgRadius,
      fill: '#ffffff',
      stroke: 'none',
      strokeWidth: 0,
    }, container);

    this.upsert('connection-point', GImage, {
      x: iconX,
      y: iconY,
      width: iconSize,
      height: iconSize,
      src: ConnectionPoint,
      cursor: 'default',
    }, container);
  }

  // 主渲染函数
  // eslint-disable-next-line @typescript-eslint/member-ordering
  public render(attr: Required<BaseNodeStyleProps>, container: Group) {
    super.render(attr, container);

    // 1. 绘制两个基础块（按层级顺序）
    this.drawHeaderBlock(container);      // 标题块
    this.drawContentBlock(container);     // 内容块（白色圆角矩形）
    
    // 2. 绘制标题内容
    this.drawUnitLabel(container);
    this.drawUnitName(container);
    this.drawMenuIcon(container);

    // 3. 绘制内容模块
    this.drawProxyTable(container);
    this.drawAgentBlock(container);
    
    // 4. 绘制连接点图标
    this.drawConnectionPoint(container);
  }
}
