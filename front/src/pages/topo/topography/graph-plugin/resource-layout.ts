// resource-layout.ts
import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { BaseLayout } from '@antv/g6';

import { NodeType } from './config';

export default class ResourceLayout extends BaseLayout {
  id = 'custom-layout';

  async execute(data: GraphData) {
    const nodes = data.nodes || [];
    const edges = data.edges || [];

    // 分离不同类型的节点
    const areaNodes = nodes.filter(node => node.type === NodeType.NET_WORK_AREA);
    const unitNodes = nodes.filter(node => node.type === NodeType.NET_WORK_UNIT);
    const accessPointNodes = nodes.filter(node => node.type === NodeType.ACCESS_POINT);
    const serverNodes = nodes.filter(node => node.type === NodeType.SERVER);

    // 构建节点关系图
    const nodeRelations = this.buildNodeRelations(nodes, edges);

    // 排序区域：有Server的区域排第一
    const sortedAreaNodes = this.sortAreaNodes(areaNodes, serverNodes, nodeRelations);

    // 布局区域节点
    const areaNodesWithStyle = this.layoutAreaNodes(sortedAreaNodes, unitNodes, accessPointNodes);

    // 布局内部节点（单元、接入点、Server）
    const internalNodesWithStyle = this.layoutInternalNodes(
      unitNodes,
      accessPointNodes,
      serverNodes,
      areaNodesWithStyle,
      nodeRelations,
    );

    // 布局边（使用贝塞尔曲线）- 只处理现有边，不自动创建
    const edgesWithStyle = this.layoutEdges(edges);

    return {
      nodes: [...areaNodesWithStyle, ...internalNodesWithStyle],
      edges: edgesWithStyle,
    };
  }

  private buildNodeRelations(nodes: NodeData[], edges: EdgeData[]) {
    const relations = new Map<string, { sources: string[], targets: string[] }>();

    nodes.forEach((node) => {
      relations.set(node.id as string, { sources: [], targets: [] });
    });

    edges.forEach((edge) => {
      const sourceId = edge.source as string;
      const targetId = edge.target as string;

      const sourceRelation = relations.get(sourceId);
      const targetRelation = relations.get(targetId);

      if (sourceRelation) sourceRelation.targets.push(targetId);
      if (targetRelation) targetRelation.sources.push(sourceId);
    });

    return relations;
  }

  private sortAreaNodes(areaNodes: NodeData[], serverNodes: NodeData[], relations: Map<string, any>) {
    return [...areaNodes].sort((a, b) => {
      const aId = a.id as string;
      const bId = b.id as string;

      // 有Server的区域排前面
      const aHasServer = serverNodes.some(server => (server.data as any)?.area === aId);
      const bHasServer = serverNodes.some(server => (server.data as any)?.area === bId);

      if (aHasServer && !bHasServer) return -1;
      if (!aHasServer && bHasServer) return 1;

      return 0;
    });
  }

  private layoutAreaNodes(areaNodes: NodeData[], unitNodes: NodeData[], accessPointNodes: NodeData[]) {
    const MARGIN = 20;

    return areaNodes.map((node, index) => {
      const nodeId = node.id as string;

      // 统计区域内的节点数量
      const areaUnits = unitNodes.filter(unit => (unit.data as any)?.area === nodeId);
      const areaAccessPoints = accessPointNodes.filter(ap => (ap.data as any)?.area === nodeId);

      const totalInternalNodes = areaUnits.length + areaAccessPoints.length;

      // 动态计算区域大小
      const baseWidth = 600;
      const baseHeight = 400;
      const minNodeWidth = 240;
      const minNodeHeight = 160;

      // 计算需要的行数和列数
      const maxNodesPerRow = 3;
      const rows = Math.ceil(totalInternalNodes / maxNodesPerRow);

      const unitWidth = minNodeWidth + MARGIN * 2;
      const unitHeight = minNodeHeight + MARGIN * 2;
      const horizontalSpacing = 30;
      const verticalSpacing = 30;

      const contentWidth = Math.max(
        baseWidth,
        Math.min(totalInternalNodes, maxNodesPerRow) * unitWidth
        + (Math.min(totalInternalNodes, maxNodesPerRow) - 1) * horizontalSpacing + MARGIN * 2,
      );

      const contentHeight = Math.max(
        baseHeight,
        rows * unitHeight + (rows - 1) * verticalSpacing + 100,
      );

      // 布局位置 - 有Server的区域在左侧，其他在右侧上下排列
      const startX = 50;
      const startY = 50;
      const areaSpacing = 150;

      let x; let y;
      if (index === 0) {
        // 第一个区域（有Server）放在左侧
        x = startX;
        y = startY;
      } else {
        // 其他区域在右侧上下排列
        const row = index - 1;
        x = startX + baseWidth + areaSpacing;
        y = startY + row * (contentHeight + areaSpacing);
      }

      return {
        ...node,
        style: {
          x,
          y,
          width: contentWidth,
          height: contentHeight,
        },
        data: {
          ...node.data,
          width: contentWidth,
          height: contentHeight,
        },
      };
    });
  }

  private layoutInternalNodes(
    unitNodes: NodeData[],
    accessPointNodes: NodeData[],
    serverNodes: NodeData[],
    areaNodes: any[],
    relations: Map<string, any>,
  ) {
    const MARGIN = 20;
    const result: any[] = [];

    areaNodes.forEach((areaNode) => {
      const areaId = areaNode.id as string;
      const areaStyle = areaNode.style;

      // 获取区域内的所有节点
      const areaUnits = unitNodes.filter(unit => (unit.data as any)?.area === areaId);
      const areaAccessPoints = accessPointNodes.filter(ap => (ap.data as any)?.area === areaId);
      const areaServers = serverNodes.filter(server => (server.data as any)?.area === areaId);

      // 先布局Server（如果有）
      areaServers.forEach((server, index) => {
        result.push({
          ...server,
          style: {
            x: areaStyle.x + MARGIN,
            y: areaStyle.y + 80 + index * 220,
            width: (server.data as any)?.width || 60,
            height: (server.data as any)?.height || 200,
          },
        });
      });

      // 布局单元和接入点
      const internalNodes = [...areaUnits, ...areaAccessPoints];
      const maxNodesPerRow = 3;
      const nodeWidth = 240 + MARGIN * 2;
      const nodeHeight = 160 + MARGIN * 2;
      const horizontalSpacing = 30;
      const verticalSpacing = 30;

      const startX = areaStyle.x + (areaServers.length > 0 ? 100 : MARGIN);
      const startY = areaStyle.y + 60;

      internalNodes.forEach((node, index) => {
        const row = Math.floor(index / maxNodesPerRow);
        const col = index % maxNodesPerRow;

        result.push({
          ...node,
          style: {
            x: startX + col * (nodeWidth + horizontalSpacing),
            y: startY + row * (nodeHeight + verticalSpacing),
            width: nodeWidth,
            height: nodeHeight,
          },
        });
      });
    });

    return result;
  }

  private layoutEdges(edges: EdgeData[]) {
    // 只处理现有边，不自动创建新边
    return edges.map(edge => ({
      ...edge,
      style: {
        ...edge.style,
        curveOffset: 30, // 贝塞尔曲线偏移
      },
    }));
  }
}
