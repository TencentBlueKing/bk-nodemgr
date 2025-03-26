import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { BaseLayout } from '@antv/g6';

import { NodeType } from './config';

interface NodeStyle extends NodeData {
  style: {
    x: number;
    y: number;
  };
}

/**
 * @description 自定义资源布局
 * 该布局主要用于节点管理的资源拓扑图展示
 * 布局特点:
 * 1. 支持按area分组布局
 * 2. 同一area内的节点按水平方向排列
 * 3. 不同area的节点按垂直方向错开排列
 * 4. 对节点位置处理,保证其布局的合理性
 */

export default class ResourceLayout extends BaseLayout {
  static readonly nodeSize = 64;
  static readonly nodeVerticalDistance = 108 + ResourceLayout.nodeSize; // 节点间垂直间距

  static readonly maxNodesPerArea = 2;
  static readonly singleWorkUnitWidth = 280;
  static readonly moreWorkUnitWidth = 520;
  static readonly commonWorkAreaHeight = 792;
  static readonly maxAreaCol = 2;
  static readonly maxAreaRow = 4;
  static readonly heightSegment = ResourceLayout.commonWorkAreaHeight / ResourceLayout.maxAreaRow;

  static readonly WORK_AREA_TYPE = NodeType.NET_WORK_AREA;
  static readonly WORK_UNIT_TYPE = NodeType.NET_WORK_UNIT;

  id = 'custom';
  areasMap = new Map<string, {
    count: number
    xPosition: number
  }>();

  private nodes: NodeData[] = [];
  private edges: EdgeData[] = [];
  private areas: NodeData[] = [];

  async execute(data: GraphData) {
    this.edges = data.edges as EdgeData[];
    this.areas = this.getWorkAreaNode(data.nodes || []);
    this.nodes = this.getWorkUnitNode(data.nodes || []);

    // 抽出type为net-work-area的节点

    // 统计各area下nodes数量
    this.setUnitCountByArea(this.nodes, this.areas);
    // // 根据area类型排序
    this.sortNodesByArea(this.nodes);

    const areasStyle = this.getAreasStyle(this.areas);
    // nodes 位置计算
    const nodesStyle = this.getNodesStyle(this.nodes);
    // 折线radius拐点处理
    const edgesStyle = this.getEdgesStyle(this.edges, nodesStyle);

    return {
      nodes: [...areasStyle, ...nodesStyle],
      edges: edgesStyle,
    };
  }

  getWorkAreaNode(nodes: NodeData[]) {
    return nodes.filter(item => item.type === ResourceLayout.WORK_AREA_TYPE);
  }

  getWorkUnitNode(nodes: NodeData[]) {
    return nodes.filter(item => item.type === ResourceLayout.WORK_UNIT_TYPE);
  }

  /**
   * @description 根据workarea类型排序
   * @param nodes 单元节点数据
   */
  private sortNodesByArea(nodes: NodeData[]) {
    // 首先按workarea分组
    const areaGroups = new Map<string, NodeData[]>();
    // 将节点按areaId分组
    for (const unit of nodes) {
      const areaId = unit.data?.area as string;
      if (!areaGroups.has(areaId)) {
        areaGroups.set(areaId, []);
      }
      areaGroups.get(areaId)?.push(unit);
    }
    // 对每个area内的节点进行排序
    for (const [areaId, areaNodes] of areaGroups) {
      areaNodes.sort((a, b) => (a.id as string).localeCompare(b.id as string));
    }
    // 清空原数组
    nodes.length = 0;

    // 按area顺序重新填充节点数组
    // 先获取所有area的id并排序
    const areaIds = Array.from(areaGroups.keys()).sort();

    // 按area顺序添加节点
    for (const areaId of areaIds) {
      const areaNodes = areaGroups.get(areaId) || [];
      for (const node of areaNodes) {
        nodes.push(node);
      }
    }
  }

  /**
   *
   * @param description 统计每个area下节点的数量，记录在areasMap
   */
  private setUnitCountByArea(nodes: NodeData[], areas: NodeData[]) {
    for (const area of areas) {
      const curUnitCount = this.countNodesInArea(nodes, area.id);
      this.areasMap.set(area.id, {
        count: curUnitCount,
        xPosition: 0,
      });
    }
  }

  /**
   * 计算area内单元节点的数量
   */
  private countNodesInArea(nodes: NodeData[], areaId: string): number {
    return nodes.filter(n => n.data?.area === areaId).length;
  }

  private getCurrentAreaNodesMaxCount(id: string) {
    const { maxNodesPerArea } = ResourceLayout;
    const currentAreaId = id;
    const count = Math.min(this.areasMap.get(currentAreaId)?.count as number, maxNodesPerArea);
    return count;
  }

  private getAreaHeight(nodesInAreaCount: number) {
    let height  = ResourceLayout.commonWorkAreaHeight;
    const maxNodesInArea = ResourceLayout.maxAreaCol * ResourceLayout.maxAreaRow;
    if (nodesInAreaCount > maxNodesInArea) {
      height = Math.ceil(nodesInAreaCount / ResourceLayout.maxAreaCol) * ResourceLayout.heightSegment;
    }
    return height;
  }

  private getAreasStyle(areas: NodeData[]) {
    let xPosition = 0;
    return areas.map((area) => {
      const currentAreaId = area.id as string;
      const count = this.areasMap.get(currentAreaId)?.count as number;
      this.areasMap.set(currentAreaId, {
        count,
        xPosition,
      });
      const curHeight = this.getAreaHeight(count);
      const curWidth = count >= ResourceLayout.maxNodesPerArea
        ? ResourceLayout.moreWorkUnitWidth
        : ResourceLayout.singleWorkUnitWidth;
      const curX = xPosition;
      xPosition += curWidth + 8;
      return {
        ...area,
        style: {
          x: curX,
          y: 0,
          width: curWidth,
          height: curHeight,
        },
      };
    });
  }

  /**
   * 计算节点样式和位置
   * @param nodes 节点数据
   */
  private getNodesStyle(nodes: NodeData[]) {
    const { maxNodesPerArea } = ResourceLayout;
    // 记录已处理过的area，用于判断是否进入新的area
    const processedAreas: string[] = [];
    // 当前area内的node序号，每进入新area时重置
    let areaNodeIndex = 0;
    // area左侧的x坐标
    let areaStartX = 0;
    return nodes.map((node) => {
      const currentAreaId = node.data?.area as string;
      // 获取当前area中节点的数量（最多支持 maxNodesPerArea 个）
      const curNodeCount = this.getCurrentAreaNodesMaxCount(currentAreaId);
      // 是否为单个管控单元
      const isSingleNode = curNodeCount < maxNodesPerArea;
      // 根据node数量，获取当前area的宽度
      const curAreaWidth = isSingleNode ? ResourceLayout.singleWorkUnitWidth : ResourceLayout.moreWorkUnitWidth;

      // 每进入一个新area，初始化数据
      if (!processedAreas.includes(currentAreaId)) {
        areaNodeIndex = 0; // 进入新area，清空节点索引
        processedAreas.push(currentAreaId); // 记录新area的名称

        areaStartX = this.areasMap.get(currentAreaId)?.xPosition as number; // 新area的左侧x位置( 10为margin-right )
      }
      areaNodeIndex += 1; // 节点索引累加

      let segmentWidth = 0;
      let adjustedXPosition = 0; // node的水平偏移
      if (!isSingleNode) {
        segmentWidth = curAreaWidth / (maxNodesPerArea + 1); // 将area width根据nodes数量平分区域
        if (areaNodeIndex % maxNodesPerArea === 0) {
          // 若当前节点为水平方向的最后一个节点
          adjustedXPosition = maxNodesPerArea * segmentWidth;
        } else {
          adjustedXPosition = (areaNodeIndex % maxNodesPerArea) * segmentWidth;
        }
      } else {
        adjustedXPosition = ResourceLayout.singleWorkUnitWidth / 2;
      }

      // node的垂直偏移，若area为单个node，直接使用节点索引; 若为多node，需要考虑node在图中是否换行
      const verticalPos = isSingleNode ? areaNodeIndex : Math.ceil(areaNodeIndex / maxNodesPerArea);
      // 节点位置偏下，向上移动
      const adjustedYPosition = -ResourceLayout.nodeSize;

      const xPosition = areaStartX + adjustedXPosition - ResourceLayout.nodeSize / 2;
      const yPosition = verticalPos * ResourceLayout.nodeVerticalDistance + adjustedYPosition;
      return {
        ...node,
        style: {
          x: xPosition,
          y: yPosition,
          zIndex: 100,
        },
      };
    });
  }

  /**
   *
   * @param description 边的样式处理，计算radius的拐点位置
   */
  private getEdgesStyle(edges: EdgeData[], nodesStyle: NodeStyle[]) {
    const nodeMap = new Map(nodesStyle.map(node => [node.id, node]));
    const halfNodeSize = ResourceLayout.nodeSize / 2;

    return edges.map((edge) => {
      const sourceNode = nodeMap.get(edge.source);
      const targetNode = nodeMap.get(edge.target);

      if (!sourceNode || !targetNode) {
        return { ...edge };
      }

      const style = { ...edge.style };
      // 当源节点和目标节点位置不在同一水平线或垂直线上，给折线增加控制点，形成折线radius样式
      if (sourceNode.style.x !== targetNode.style.x && sourceNode.style.y !== targetNode.style.y) {
        let x; let y = 0;
        const sourceNodeX = sourceNode.style.x;
        const sourceNodeY = sourceNode.style.y;
        const targetNodeX = targetNode.style.x;
        const targetNodeY = targetNode.style.y;

        if (sourceNodeX > targetNodeX) {
          x = Math.min(sourceNodeX, targetNodeX) + halfNodeSize;
        } else {
          x = Math.max(sourceNodeX, targetNodeX) + halfNodeSize;
        }

        if (sourceNodeY > targetNodeY) {
          y = Math.max(sourceNodeY, targetNodeY) + halfNodeSize;
        } else {
          y = Math.min(sourceNodeY, targetNodeY) + halfNodeSize;
        }

        style.controlPoints = [[x, y]];
        // 避免尾部箭头被proxy、agent标签遮挡，设置偏移量50
        style.endArrowOffset = 50;
      }

      return {
        ...edge,
        style,
      };
    });
  }
}
