import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { BaseLayout } from '@antv/g6';

import { NodeType } from './config';

export default class HorizontalHierarchyLayout extends BaseLayout {
  id = 'horizontal-hierarchy-layout';

  // --- 布局常量 ---
  private readonly AREA_SPACING_HORIZONTAL = 300;
  private readonly AREA_SPACING_VERTICAL = 150;
  private readonly AREA_PADDING = 20;
  private readonly UNIT_NODE_HEIGHT = 202;
  private readonly AP_NODE_HEIGHT = 32;
  private readonly NODE_SPACING_ROW = 60;
  private readonly NODE_SPACING_COL = 40;
  private readonly UNIT_COL_WIDTH = 270;
  private readonly AP_COL_WIDTH = 270;
  private readonly AREA_MIN_WIDTH = 300;
  private readonly AREA_HEIGHT_EXTRA = 20;

  async assign(data: GraphData, options?: any) {
    return this.execute(data, options);
  }

  async execute(data: GraphData, options?: any) {
    const nodes = data.nodes || [];
    const edges = data.edges || [];

    const areaNodes = nodes.filter(node => node?.type === NodeType.NET_WORK_AREA);
    const unitNodes = nodes.filter(node => node?.type === NodeType.NET_WORK_UNIT);
    const accessPointNodes = nodes.filter(node => node?.type === NodeType.ACCESS_POINT);

    const nodesByArea = this.groupNodesByArea(areaNodes, unitNodes, accessPointNodes, edges);
    const { areaColumns } = this.buildAreaColumnsByDirectUnitDependency(edges, nodesByArea);

    const globalDependencies = this.buildGlobalDependencies(nodes, edges);
    const sourceNodes = unitNodes.filter(u => (u.data as any).is_direct);
    const globalNodeLevels = this.calculateGlobalNodeLevels(nodes, globalDependencies, sourceNodes);

    const { allNodes, nodeLayoutInfo } = this.layoutAreasByColumnsWithGlobalLevels(
      areaColumns,
      nodesByArea,
      areaNodes,
      globalNodeLevels,
    );

    const edgesWithStyle = this.optimizeEdgeStyle(edges, allNodes, nodeLayoutInfo);

    return { nodes: allNodes, edges: edgesWithStyle };
  }

  private groupNodesByArea(
    areaNodes: NodeData[],
    unitNodes: NodeData[],
    accessPointNodes: NodeData[],
    edges: EdgeData[],
  ) {
    const nodesByArea = new Map<string, {
      area: NodeData;
      units: NodeData[];
      directUnits: NodeData[];
      accessPoints: NodeData[];
      nodeDependencies: Map<string, Set<string>>;
      unitUpstreamAPs: Map<string, string>;
    }>();

    areaNodes.forEach((area) => {
      nodesByArea.set(area.id as string, {
        area,
        units: [],
        directUnits: [],
        accessPoints: [],
        nodeDependencies: new Map(),
        unitUpstreamAPs: new Map(),
      });
    });

    [
      { list: unitNodes, key: 'units' },
      { list: accessPointNodes, key: 'accessPoints' },
    ].forEach(({ list, key }) => {
      list.forEach((node) => {
        const areaId = (node.data as any)?.area;
        const areaData = nodesByArea.get(areaId);
        if (areaData && !areaData[key].some(n => n.id === node.id)) {
          areaData[key].push(node);
          if (key === 'units' && (node.data as any).is_direct) {
            areaData.directUnits.push(node);
          }
        }
      });
    });

    nodesByArea.forEach((areaData) => {
      const areaNodeIds = new Set([
        ...areaData.units.map(n => n.id),
        ...areaData.accessPoints.map(n => n.id),
      ]);

      edges.forEach((edge) => {
        if (!edge.source || !edge.target) return;
        const sourceId = edge.source as string;
        const targetId = edge.target as string;

        if (areaNodeIds.has(sourceId) && areaNodeIds.has(targetId)) {
          if (!areaData.nodeDependencies.has(sourceId)) {
            areaData.nodeDependencies.set(sourceId, new Set());
          }
          areaData.nodeDependencies.get(sourceId)!.add(targetId);

          if (sourceId.startsWith('workUnit-') && targetId.startsWith('accessPoint-') && !areaData.unitUpstreamAPs.has(sourceId)) {
            areaData.unitUpstreamAPs.set(sourceId, targetId);
          }
        }
      });
    });

    return nodesByArea;
  }

  /**
   * 修复无依赖区域的排序逻辑
   */
  private buildAreaColumnsByDirectUnitDependency(edges: EdgeData[], nodesByArea: Map<string, any>) {
    const areaIds = Array.from(nodesByArea.keys());
    const dependencies = new Map(areaIds.map(id => [id, new Set<string>()]));
    const reverseDependencies = new Map(areaIds.map(id => [id, new Set<string>()]));
    const directUnitAreas = new Set<string>();

    nodesByArea.forEach((areaData, areaId) => {
      if (areaData.directUnits.length > 0) {
        directUnitAreas.add(areaId);
      }
    });

    // 构建区域间依赖关系
    edges.forEach((edge) => {
      if (!edge.source || !edge.target) return;
      const sourceId = edge.source as string;
      const targetId = edge.target as string;

      let downstreamAreaId = '';
      let upstreamAreaId = '';

      nodesByArea.forEach((areaData, areaId) => {
        const nodeIds = new Set([...areaData.units, ...areaData.accessPoints].map(n => n.id));
        // 只要边的任何一端是接入点，就认为是跨区域依赖
        if (nodeIds.has(sourceId) && (sourceId.startsWith('accessPoint-') || targetId.startsWith('accessPoint-'))) {
          downstreamAreaId = areaId;
        }
        if (nodeIds.has(targetId) && (sourceId.startsWith('accessPoint-') || targetId.startsWith('accessPoint-'))) {
          upstreamAreaId = areaId;
        }
      });

      if (upstreamAreaId && downstreamAreaId && upstreamAreaId !== downstreamAreaId) {
        dependencies.get(downstreamAreaId)!.add(upstreamAreaId);
        reverseDependencies.get(upstreamAreaId)!.add(downstreamAreaId);
      }
    });

    const visited = new Set<string>();
    const areaColumns: string[][] = [];

    // 第一列：有直连单元的区域
    const firstCol = Array.from(directUnitAreas);
    areaColumns.push(firstCol);
    firstCol.forEach(id => visited.add(id));

    // 后续列：根据依赖关系排列
    let currentLevel = 0;
    while (true) {
      const currentCol = areaColumns[currentLevel] || [];
      if (!currentCol.length) break;

      const nextCol = new Set<string>();
      currentCol.forEach((upAreaId) => {
        reverseDependencies.get(upAreaId)!.forEach((downAreaId) => {
          // 确保下游区域的所有依赖都已被访问
          if (!visited.has(downAreaId) && Array.from(dependencies.get(downAreaId)!).every(u => visited.has(u))) {
            nextCol.add(downAreaId);
          }
        });
      });

      if (!nextCol.size) break;
      const nextColArr = Array.from(nextCol);
      areaColumns.push(nextColArr);
      nextColArr.forEach(id => visited.add(id));
      currentLevel++;
    }

    // 【修复点 1】：正确处理无依赖的区域
    // 无依赖区域是指：不依赖任何其他区域，也不被任何其他区域依赖的区域
    const remainingAreas = areaIds.filter(id => !visited.has(id));
    if (remainingAreas.length > 0) {
      areaColumns.push(remainingAreas);
    }

    return { areaColumns };
  }

  private buildGlobalDependencies(nodes: NodeData[], edges: EdgeData[]): Map<string, Set<string>> {
    const dependencies = new Map<string, Set<string>>();
    nodes.forEach(node => dependencies.set(node.id, new Set()));

    edges.forEach((edge) => {
      if (!edge.source || !edge.target) return;
      const sourceId = edge.source as string;
      const targetId = edge.target as string;
      if (dependencies.has(sourceId)) {
        dependencies.get(sourceId)!.add(targetId);
      }
    });

    return dependencies;
  }

  // eslint-disable-next-line max-len
  private calculateGlobalNodeLevels(nodes: NodeData[], dependencies: Map<string, Set<string>>, sourceNodes: NodeData[]): Map<string, number> {
    const nodeLevels = new Map<string, number>();
    const visited = new Set<string>();
    const queue: { id: string; level: number }[] = [];

    sourceNodes.forEach((node) => {
      if (!visited.has(node.id)) {
        nodeLevels.set(node.id, 0);
        visited.add(node.id);
        queue.push({ id: node.id, level: 0 });
      }
    });

    while (queue.length > 0) {
      const { id: currentId, level: currentLevel } = queue.shift()!;

      nodes.forEach((node) => {
        const nodeId = node.id;
        if (dependencies.get(nodeId)?.has(currentId) && !visited.has(nodeId)) {
          const nextLevel = currentLevel + 1;
          nodeLevels.set(nodeId, nextLevel);
          visited.add(nodeId);
          queue.push({ id: nodeId, level: nextLevel });
        }
      });
    }

    nodes.forEach((node) => {
      if (!nodeLevels.has(node.id)) {
        nodeLevels.set(node.id, 0);
      }
    });

    return nodeLevels;
  }

  private layoutAreasByColumnsWithGlobalLevels(
    areaColumns: string[][],
    nodesByArea: Map<string, any>,
    areaNodes: NodeData[],
    globalNodeLevels: Map<string, number>,
  ) {
    const allNodes: any[] = [];
    const nodeLayoutInfo = new Map<string, { x: number; y: number; areaId: string; rowIndex: number }>();
    let currentColumnX = 150;

    areaColumns.forEach((columnAreaIds) => {
      let currentRowY = 150;

      columnAreaIds.forEach((areaId) => {
        const areaData = nodesByArea.get(areaId);
        if (!areaData) return;

        const allAreaNodes = [...areaData.units, ...areaData.accessPoints].filter(Boolean);
        if (allAreaNodes.length === 0) {
          // 【修复点 2】：处理空区域
          this.layoutEmptyArea(areaId, areaNodes, currentColumnX, currentRowY, allNodes);
          currentRowY += this.UNIT_NODE_HEIGHT + this.AREA_SPACING_VERTICAL;
          return;
        }

        const levelGroups = this.groupNodesByGlobalLevel(allAreaNodes, globalNodeLevels, areaData, nodesByArea);
        const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b);
        const columns = sortedLevels.map((level) => {
          const nodes = levelGroups.get(level)!;
          const firstNode = nodes[0];
          const colWidth = firstNode.type === NodeType.ACCESS_POINT ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH;
          return { nodes: this.sortNodesInColumn(nodes), width: colWidth };
        });

        const areaSize = this.calculateAreaSize(columns);
        this.layoutAreaBackground(areaId, areaNodes, areaSize, currentColumnX, currentRowY, allNodes);
        this.layoutAreaChildNodes(
          areaId,
          columns,
          currentColumnX,
          currentRowY,
          allNodes,
          nodeLayoutInfo,
        );

        currentRowY += areaSize.height + this.AREA_SPACING_VERTICAL;
      });

      // 【修复点 3】：正确计算包含无依赖区域的列宽
      const columnMaxWidth = this.calculateColumnMaxWidth(columnAreaIds, nodesByArea, globalNodeLevels);
      currentColumnX += columnMaxWidth + this.AREA_SPACING_HORIZONTAL;
    });

    return { allNodes, nodeLayoutInfo };
  }

  /**
   * 为无依赖区域计算全局层级
   */
  private groupNodesByGlobalLevel(
    nodes: NodeData[],
    globalNodeLevels: Map<string, number>,
    areaData: any,
    nodesByArea: Map<string, any>,
  ): Map<number, NodeData[]> {
    const levelGroups = new Map<number, NodeData[]>();
    let areaSourceLevel = this.getAreaSourceGlobalLevel(areaData, globalNodeLevels);

    // 如果是无依赖区域，且没有源节点，手动设置一个基础层级
    if (areaSourceLevel === Infinity) {
      areaSourceLevel = 0;
    }

    nodes.forEach((node) => {
      const globalLevel = globalNodeLevels.get(node.id) || 0;
      const localLevel = globalLevel - areaSourceLevel;
      if (!levelGroups.has(localLevel)) {
        levelGroups.set(localLevel, []);
      }
      levelGroups.get(localLevel)!.push(node);
    });

    return levelGroups;
  }

  private getAreaSourceGlobalLevel(areaData: any, globalNodeLevels: Map<string, number>): number {
    let minLevel = Infinity;

    // 优先找直连单元
    areaData.directUnits.forEach((unit: NodeData) => {
      const level = globalNodeLevels.get(unit.id) || 0;
      minLevel = Math.min(minLevel, level);
    });

    // 如果没有直连单元，找无依赖的节点
    if (minLevel === Infinity) {
      const allAreaNodes = [...areaData.units, ...areaData.accessPoints].filter(Boolean);
      allAreaNodes.forEach((node) => {
        if ((areaData.nodeDependencies.get(node.id)?.size || 0) === 0) {
          const level = globalNodeLevels.get(node.id) || 0;
          minLevel = Math.min(minLevel, level);
        }
      });
    }

    return minLevel;
  }

  private sortNodesInColumn(nodes: NodeData[]): NodeData[] {
    return [...nodes].sort((a, b) => {
      if (a.type !== b.type) {
        return a.type === NodeType.ACCESS_POINT ? -1 : 1;
      }
      return a.id.localeCompare(b.id);
    });
  }

  private calculateAreaSize(columns: { nodes: NodeData[], width: number }[]): { width: number, height: number } {
    if (!columns.length) {
      return {
        width: this.AREA_MIN_WIDTH,
        height: this.AREA_PADDING * 2 + this.UNIT_NODE_HEIGHT + this.AREA_HEIGHT_EXTRA,
      };
    }

    const totalColWidth = columns.reduce((sum, col) => sum + col.width, 0);
    const totalColSpacing = (columns.length - 1) * this.NODE_SPACING_COL;
    const areaWidth = Math.max(this.AREA_PADDING * 2 + totalColWidth + totalColSpacing, this.AREA_MIN_WIDTH);

    const maxColHeight = columns.reduce((max, col) => {
      let totalHeight = 0;
      col.nodes.forEach((node) => {
        // eslint-disable-next-line max-len
        totalHeight += (node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT) + this.NODE_SPACING_ROW;
      });
      return Math.max(max, totalHeight > 0 ? totalHeight - this.NODE_SPACING_ROW : 0);
    }, 0);
    const areaHeight = this.AREA_PADDING * 2 + maxColHeight + this.AREA_HEIGHT_EXTRA + 60;

    return { width: areaWidth, height: areaHeight };
  }

  /**
   * 布局空区域
   */
  private layoutEmptyArea(
    areaId: string,
    areaNodes: NodeData[],
    x: number,
    y: number,
    allNodes: any[],
  ) {
    const areaNode = areaNodes.find(n => n.id === areaId);
    if (areaNode) {
      allNodes.push({
        id: areaNode.id,
        type: areaNode.type,
        data: { ...areaNode.data, width: this.AREA_MIN_WIDTH, height: this.UNIT_NODE_HEIGHT },
        style: {
          x,
          y,
          width: this.AREA_MIN_WIDTH,
          height: this.UNIT_NODE_HEIGHT,
          fill: '#f5f7fa',
          stroke: '#dce1e8',
          lineWidth: 2,
          zIndex: -1,
        },
      });
    }
  }

  private layoutAreaBackground(
    areaId: string,
    areaNodes: NodeData[],
    areaSize: { width: number, height: number },
    x: number,
    y: number,
    allNodes: any[],
  ) {
    const areaNode = areaNodes.find(n => n.id === areaId);
    if (areaNode) {
      allNodes.push({
        id: areaNode.id,
        type: areaNode.type,
        data: { ...areaNode.data, width: areaSize.width, height: areaSize.height },
        style: {
          x,
          y,
          width: areaSize.width,
          height: areaSize.height,
          fill: '#f5f7fa',
          stroke: '#dce1e8',
          lineWidth: 2,
          zIndex: -1,
        },
      });
    }
  }

  private layoutAreaChildNodes(
    areaId: string,
    columns: { nodes: NodeData[], width: number }[],
    areaX: number,
    areaY: number,
    allNodes: any[],
    nodeLayoutInfo: Map<string, { x: number; y: number; areaId: string; rowIndex: number }>,
  ) {
    const childStartX = areaX + this.AREA_PADDING;
    const childStartY = areaY + this.AREA_PADDING + 60;
    let currentColX = childStartX;

    columns.forEach((col) => {
      const { nodes, width } = col;
      nodes.forEach((node, rowIndex) => {
        const nodeHeight = node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT;
        const nodeY = childStartY + rowIndex * (nodeHeight + this.NODE_SPACING_ROW);

        allNodes.push({
          id: node.id,
          type: node.type,
          data: { ...node.data },
          style: {
            x: currentColX,
            y: nodeY,
            width,
            height: nodeHeight,
            fill: '#ffffff',
            stroke: '#dce1e8',
            lineWidth: 1.5,
            textAlign: 'center',
            textBaseline: 'middle',
            fontSize: 14,
            fontWeight: 500,
            ...node.style,
          },
          zIndex: 1,
        });

        nodeLayoutInfo.set(node.id, {
          x: currentColX,
          y: nodeY,
          areaId,
          rowIndex,
        });
      });
      currentColX += width + this.NODE_SPACING_COL;
    });
  }

  /**
   * 确保计算列宽时包含无依赖区域
   */
  // eslint-disable-next-line max-len
  private calculateColumnMaxWidth(columnAreaIds: string[], nodesByArea: Map<string, any>, globalNodeLevels: Map<string, number>): number {
    let maxWidth = 0;
    columnAreaIds.forEach((areaId) => {
      const areaData = nodesByArea.get(areaId);
      if (!areaData) return;

      const allAreaNodes = [...areaData.units, ...areaData.accessPoints].filter(Boolean);
      if (allAreaNodes.length === 0) {
        // 空区域使用最小宽度
        maxWidth = Math.max(maxWidth, this.AREA_MIN_WIDTH);
        return;
      }

      const levelGroups = this.groupNodesByGlobalLevel(allAreaNodes, globalNodeLevels, areaData, nodesByArea);
      const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b);

      const colWidths = sortedLevels.map((level) => {
        const nodes = levelGroups.get(level)!;
        return nodes[0].type === NodeType.ACCESS_POINT ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH;
      });

      // eslint-disable-next-line max-len
      const totalWidth = colWidths.reduce((sum, w) => sum + w, 0) + (colWidths.length - 1) * this.NODE_SPACING_COL + this.AREA_PADDING * 2;
      maxWidth = Math.max(maxWidth, totalWidth, this.AREA_MIN_WIDTH);
    });
    return maxWidth;
  }

  private optimizeEdgeStyle(
    edges: EdgeData[],
    allNodes: NodeData[],
    nodeLayoutInfo: Map<string, { x: number; y: number; areaId: string; rowIndex: number }>
  ) {
    const nodeIds = new Set(allNodes.map(n => n.id));
    const apToUnitCount = new Map<string, number>();

    edges.forEach((edge) => {
      if (!edge.source || !edge.target) return;
      const sId = edge.source as string;
      const tId = edge.target as string;
      if (sId.startsWith('accessPoint-') && tId.startsWith('workUnit-')) {
        const key = `${sId}-${tId}`;
        apToUnitCount.set(key, (apToUnitCount.get(key) || 0) + 1);
      }
    });

    return edges.filter((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      return nodeIds.has(sId) && nodeIds.has(tId);
    }).map((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      const source = nodeLayoutInfo.get(sId);
      const target = nodeLayoutInfo.get(tId);

      if (!source || !target) {
        return { ...edge, type: 'line' };
      }

      const isUnitAP = sId.startsWith('workUnit-') && tId.startsWith('accessPoint-') || sId.startsWith('accessPoint-') && tId.startsWith('workUnit-');

      if (isUnitAP) {
        const isAPToUnit = sId.startsWith('accessPoint-') && tId.startsWith('workUnit-');
        const apNode = isAPToUnit ? source : target;
        const horizontalY = apNode.y + (apNode.y === source.y ? this.AP_NODE_HEIGHT : this.UNIT_NODE_HEIGHT) / 2;

        if (isAPToUnit) {
          const key = `${sId}-${tId}`;
          const count = apToUnitCount.get(key) || 0;
          const offset = (count - 1) * 30;
          const adjustedY = horizontalY + offset;

          if (count === 1) {
            return {
              ...edge,
              type: 'line',
              style: {
                stroke: '#666',
                lineWidth: 2,
                endArrow: false,
                strokeOpacity: 0.8,
                startPoint: [source.x + (sId.startsWith('accessPoint-') ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH), adjustedY],
                endPoint: [target.x, adjustedY],
              },
            };
          }
          const turnX = Math.max(source.x, target.x) + 50;
          return {
            ...edge,
            type: 'polyline',
            style: {
              stroke: '#666',
              lineWidth: 2,
              endArrow: false,
              strokeOpacity: 0.8,
              points: [
                [source.x + (sId.startsWith('accessPoint-') ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH), adjustedY],
                [turnX, adjustedY],
                [turnX, target.y + 41 + offset],
                [target.x, target.y + 41 + offset],
              ],
              lineJoin: 'round',
            },
          };
        }
        const verticalOffset = 20;
        return {
          ...edge,
          type: 'cubic',
          style: {
            stroke: '#3182ce',
            lineWidth: 2,
            endArrow: { path: 'M 0,0 L 10,5 L 10,-5 Z', fill: '#3182ce' },
            strokeOpacity: 0.8,
            points: [
              [source.x, source.y + this.UNIT_NODE_HEIGHT / 2],
              [source.x - 50, source.y + this.UNIT_NODE_HEIGHT / 2],
              [source.x - 50, target.y + this.AP_NODE_HEIGHT / 2 + verticalOffset],
              [target.x, target.y + this.AP_NODE_HEIGHT / 2],
            ],
          },
        };
      }

      const targetMidY = target.y + (tId.startsWith('accessPoint-') ? this.AP_NODE_HEIGHT : this.UNIT_NODE_HEIGHT) / 2;
      return {
        ...edge,
        type: 'line',
        style: {
          stroke: '#666',
          lineWidth: 2,
          endArrow: false,
          strokeOpacity: 0.7,
          startPoint: [source.x + (sId.startsWith('accessPoint-') ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH), targetMidY],
          endPoint: [target.x, targetMidY],
        },
      };
    });
  }
}
