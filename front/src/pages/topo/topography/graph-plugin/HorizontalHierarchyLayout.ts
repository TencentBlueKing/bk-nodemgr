import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { BaseLayout } from '@antv/g6';

import { NodeType } from './config';

export interface HorizontalHierarchyLayoutOptions {
  collapsed?: boolean; // 控制孤立区域是否收起
}

export default class HorizontalHierarchyLayout extends BaseLayout {
  id = 'horizontal-hierarchy-layout';

  // --- 布局常量 ---
  private readonly AREA_SPACING_HORIZONTAL = 50;
  private readonly AREA_SPACING_VERTICAL = 50;
  private readonly AREA_PADDING = 20;
  private readonly UNIT_NODE_HEIGHT = 202;
  private readonly AP_NODE_HEIGHT = 22; // 修正：与AccessPointNode.nodeHeight保持一致
  private readonly NODE_SPACING_ROW = 8; // 接入点纵向间距保持8px
  private readonly UNIT_SPACING_ROW = 100; // 单元纵向间距 120px，为边水平段预留足够空间
  private readonly NODE_SPACING_COL = 20;
  private readonly UNIT_COL_WIDTH = 240;
  private readonly AP_COL_WIDTH = 140;
  private readonly AREA_MIN_WIDTH = 260;
  private readonly AREA_HEIGHT_EXTRA = 150;

  private options: HorizontalHierarchyLayoutOptions = { collapsed: false };

  async assign(data: GraphData, options?: HorizontalHierarchyLayoutOptions) {
    if (options) {
      this.options = { ...this.options, ...options };
    }
    return this.execute(data, this.options);
  }

  async execute(data: GraphData, options?: HorizontalHierarchyLayoutOptions) {
    const nodes = data.nodes || [];
    const edges = data.edges || [];
    const isCollapsed = options?.collapsed ?? this.options.collapsed ?? false;

    const areaNodes = nodes.filter(node => node?.type === NodeType.NET_WORK_AREA);
    const unitNodes = nodes.filter(node => node?.type === NodeType.NET_WORK_UNIT);
    const accessPointNodes = nodes.filter(node => node?.type === NodeType.ACCESS_POINT);

    const nodesByArea = this.groupNodesByArea(areaNodes, unitNodes, accessPointNodes, edges);

    // 构建列，并分离出孤立区域
    const { areaColumns, isolatedAreas } = this.buildAreaColumnsByDirectUnitDependency(edges, nodesByArea);

    // 根据折叠状态决定参与布局的列
    const columnsToLayout = [...areaColumns];
    if (!isCollapsed && isolatedAreas.length > 0) {
      columnsToLayout.push(isolatedAreas);
    }

    const globalDependencies = this.buildGlobalDependencies(nodes, edges);
    const sourceNodes = unitNodes.filter(u => (u.data as any).is_direct);
    const globalNodeLevels = this.calculateGlobalNodeLevels(nodes, globalDependencies, sourceNodes);

    const { allNodes, nodeLayoutInfo } = this.layoutAreasByColumnsWithGlobalLevels(
      columnsToLayout,
      nodesByArea,
      areaNodes,
      globalNodeLevels,
    );

    // 处理被折叠的隐藏节点
    if (isCollapsed && isolatedAreas.length > 0) {
      isolatedAreas.forEach((areaId) => {
        const areaData = nodesByArea.get(areaId);
        if (areaData) {
          // 隐藏区域节点
          allNodes.push({
            id: areaData.area.id,
            style: { visibility: 'hidden', x: 0, y: 0 },
            data: { ...areaData.area.data },
          });
          // 隐藏区域内的子节点
          [...areaData.units, ...areaData.accessPoints].forEach((node) => {
            allNodes.push({
              id: node.id,
              style: { visibility: 'hidden', x: 0, y: 0 },
              data: { ...node.data },
            });
          });
        }
      });
    }

    const { edges: edgesWithStyle, apColorMap } = this.optimizeEdgeStyle(edges, allNodes, nodeLayoutInfo);

    // 将接入点节点的边框颜色设置为与其边颜色一致
    allNodes.forEach((node) => {
      if (node.id?.startsWith('accessPoint-') && apColorMap.has(node.id)) {
        const color = apColorMap.get(node.id)!;
        node.style = { ...node.style, stroke: color, lineWidth: 1.5 };
      }
    });

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

    edges.forEach((edge) => {
      if (!edge.source || !edge.target) return;
      const sourceId = edge.source as string;
      const targetId = edge.target as string;

      let downstreamAreaId = '';
      let upstreamAreaId = '';

      nodesByArea.forEach((areaData, areaId) => {
        const nodeIds = new Set([...areaData.units, ...areaData.accessPoints].map(n => n.id));
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

    const firstCol = Array.from(directUnitAreas);
    if (firstCol.length > 0) {
      areaColumns.push(firstCol);
      firstCol.forEach(id => visited.add(id));
    }

    let currentLevel = 0;
    while (true) {
      const currentCol = areaColumns[currentLevel] || [];

      // 兜底逻辑：如果依赖图有环或无直连单元入口，找入度为0的
      if (areaColumns.length === 0 && dependencies.size > 0 && directUnitAreas.size === 0) {
        const roots = areaIds.filter(id => dependencies.get(id)!.size === 0 && !visited.has(id));
        if (roots.length > 0) {
          areaColumns.push(roots);
          roots.forEach(id => visited.add(id));
          continue;
        } else {
          break;
        }
      }

      if (!currentCol.length) break;

      const nextCol = new Set<string>();
      currentCol.forEach((upAreaId) => {
        reverseDependencies.get(upAreaId)!.forEach((downAreaId) => {
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

    // 孤立区域
    const isolatedAreas = areaIds.filter(id => !visited.has(id));

    return { areaColumns, isolatedAreas };
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
          // 对空区域也做同样的位置保持处理 (略，原理同下)
          this.layoutEmptyArea(areaId, areaNodes, currentColumnX, currentRowY, allNodes);
          currentRowY += this.UNIT_NODE_HEIGHT + this.AREA_SPACING_VERTICAL;
          return;
        }

        // ... (中间的分组和排序逻辑保持不变) ...
        const levelGroups = this.groupNodesByGlobalLevel(allAreaNodes, globalNodeLevels, areaData, nodesByArea);
        const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b);
        const columns = sortedLevels.map((level) => {
          const nodes = this.sortNodesInColumn(levelGroups.get(level)!);
          const colWidth = nodes[0].type === NodeType.ACCESS_POINT ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH;
          return { nodes, width: colWidth };
        });

        // 1. 计算尺寸 (取最大值，保持宽度不缩回)
        let areaSize = this.calculateAreaSize(columns);
        const areaNode = areaNodes.find(n => n.id === areaId);
        const currentW = Number(areaNode?.style?.width || areaNode?.data?.width || 0);
        const currentH = Number(areaNode?.style?.height || areaNode?.data?.height || 0);

        areaSize = {
          width: Math.max(areaSize.width, currentW),
          height: Math.max(areaSize.height, currentH),
        };

        // 2. 【核心修改】确定区域位置 (X, Y)
        // 算法计算出的理论位置
        const calculatedX = currentColumnX;
        const calculatedY = currentRowY;

        // 尝试读取现有位置
        // 注意：G6 中 style.x 可能是 0，所以要判断 undefined
        const existingX = areaNode?.style?.x;
        const existingY = areaNode?.style?.y;

        // 如果有现有位置，就用现有的；否则用算出来的
        const finalAreaX = (existingX !== undefined) ? Number(existingX) : calculatedX;
        const finalAreaY = (existingY !== undefined) ? Number(existingY) : calculatedY;

        // 3. 传递 finalAreaX / finalAreaY 给背景绘制和子节点布局
        this.layoutAreaBackground(areaId, areaNodes, areaSize, finalAreaX, finalAreaY, allNodes);

        this.layoutAreaChildNodes(
          areaId,
          columns,
          finalAreaX, // 使用最终确定的 X
          finalAreaY, // 使用最终确定的 Y
          allNodes,
          nodeLayoutInfo,
        );

        // 更新下一行的理论 Y 坐标 (依然按算法累加，保证新加入的区域不会重叠)
        // 如果你希望被拖走的区域原来的位置“空出来”，就保持这样。
        // 如果你希望被拖走的区域原来的位置“被填补”，这里逻辑会更复杂，目前保持这样最稳妥。
        currentRowY += areaSize.height + this.AREA_SPACING_VERTICAL;
      });

      // 计算列宽 (保持不变)
      const columnMaxWidth = this.calculateColumnMaxWidth(columnAreaIds, nodesByArea, globalNodeLevels, areaNodes);
      currentColumnX += columnMaxWidth + this.AREA_SPACING_HORIZONTAL;
    });

    return { allNodes, nodeLayoutInfo };
  }

  private groupNodesByGlobalLevel(
    nodes: NodeData[],
    globalNodeLevels: Map<string, number>,
    areaData: any,
    nodesByArea: Map<string, any>,
  ): Map<number, NodeData[]> {
    const levelGroups = new Map<number, NodeData[]>();
    let areaSourceLevel = this.getAreaSourceGlobalLevel(areaData, globalNodeLevels);

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

    areaData.directUnits.forEach((unit: NodeData) => {
      const level = globalNodeLevels.get(unit.id) || 0;
      minLevel = Math.min(minLevel, level);
    });

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
      let lastNodeType = NodeType.ACCESS_POINT; // 默认使用接入点类型
      col.nodes.forEach((node) => {
        totalHeight += (node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT) + 
                      (node.type === NodeType.NET_WORK_UNIT ? this.UNIT_SPACING_ROW : this.NODE_SPACING_ROW);
        lastNodeType = node.type; // 记录最后一个节点类型
      });
      return Math.max(max, totalHeight > 0 ? totalHeight - (lastNodeType === NodeType.NET_WORK_UNIT ? this.UNIT_SPACING_ROW : this.NODE_SPACING_ROW) : 0);
    }, 0);
    const areaHeight = this.AREA_PADDING * 2 + maxColHeight + this.AREA_HEIGHT_EXTRA + 50;

    return { width: areaWidth, height: areaHeight };
  }

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
          visibility: 'visible',
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
          visibility: 'visible',
        },
      });
    }
  }

  private layoutAreaChildNodes(
    areaId: string,
    columns: { nodes: NodeData[], width: number }[],
    areaX: number, // 这是上面传进来的 finalAreaX
    areaY: number, // 这是上面传进来的 finalAreaY
    allNodes: any[],
    nodeLayoutInfo: Map<string, { x: number; y: number; areaId: string; rowIndex: number }>,
  ) {
    // 算法计算出的子节点起始基准点
    const childStartX = areaX + this.AREA_PADDING;
    const childStartY = areaY + this.AREA_PADDING + 50;

    let currentColX = childStartX;

    columns.forEach((col) => {
      const { nodes, width } = col;
      nodes.forEach((node, rowIndex) => {
        const nodeHeight = node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT;

        // 算法计算出的理论位置
        const calculatedChildX = currentColX;
        const calculatedChildY = childStartY + rowIndex * (nodeHeight + 
          (node.type === NodeType.NET_WORK_UNIT ? this.UNIT_SPACING_ROW : this.NODE_SPACING_ROW));

        // 【核心修改】优先使用子节点现有的位置
        const existingChildX = node.style?.x;
        const existingChildY = node.style?.y;

        const finalChildX = (existingChildX !== undefined) ? Number(existingChildX) : calculatedChildX;
        const finalChildY = (existingChildY !== undefined) ? Number(existingChildY) : calculatedChildY;

        allNodes.push({
          id: node.id,
          type: node.type,
          data: { ...node.data },
          style: {
            x: finalChildX, // 使用最终位置
            y: finalChildY, // 使用最终位置
            width,
            height: nodeHeight,
            fill: '#ffffff',
            stroke: '#dce1e8',
            lineWidth: 1.5,
            textAlign: 'center',
            textBaseline: 'middle',
            fontSize: 14,
            fontWeight: 500,
            visibility: 'visible',
            ...node.style, // 这里的 ...node.style 会包含之前的 x,y，但我们显式指定了 finalX/Y 更清晰
          },
          zIndex: 1,
        });

        nodeLayoutInfo.set(node.id, {
          x: finalChildX,
          y: finalChildY,
          areaId,
          rowIndex,
        });
      });
      currentColX += width + this.NODE_SPACING_COL;
    });
  }

  private calculateColumnMaxWidth(
    columnAreaIds: string[],
    nodesByArea: Map<string, any>,
    globalNodeLevels: Map<string, number>,
    areaNodes: NodeData[], // 新增参数
  ): number {
    let maxWidth = 0;
    columnAreaIds.forEach((areaId) => {
      const areaData = nodesByArea.get(areaId);
      if (!areaData) return;

      // 1. 获取算法计算的宽度
      let calculatedWidth = 0;
      const allAreaNodes = [...areaData.units, ...areaData.accessPoints].filter(Boolean);

      if (allAreaNodes.length === 0) {
        calculatedWidth = this.AREA_MIN_WIDTH;
      } else {
        const levelGroups = this.groupNodesByGlobalLevel(allAreaNodes, globalNodeLevels, areaData, nodesByArea);
        const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b);
        const colWidths = sortedLevels.map((level) => {
          const nodes = levelGroups.get(level)!;
          return nodes[0].type === NodeType.ACCESS_POINT ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH;
        });
        // eslint-disable-next-line max-len
        calculatedWidth = colWidths.reduce((sum, w) => sum + w, 0) + (colWidths.length - 1) * this.NODE_SPACING_COL + this.AREA_PADDING * 2;
        calculatedWidth = Math.max(calculatedWidth, this.AREA_MIN_WIDTH);
      }

      // ------------------ 【核心修改开始】 ------------------
      // 2. 获取当前实际宽度
      const areaNode = areaNodes.find(n => n.id === areaId);
      const currentWidth = Number(areaNode?.style?.width || areaNode?.data?.width || 0);

      // 3. 取最大值作为该区域对列宽的贡献
      const finalWidth = Math.max(calculatedWidth, currentWidth);
      // ------------------ 【核心修改结束】 ------------------

      maxWidth = Math.max(maxWidth, finalWidth);
    });
    return maxWidth;
  }

  // ---------------------- 重点修改：还原旧的边样式逻辑 ----------------------
  private optimizeEdgeStyle(
    edges: EdgeData[],
    allNodes: NodeData[],
    nodeLayoutInfo: Map<string, { x: number; y: number; areaId: string; rowIndex: number }>,
  ) {
    const nodeIds = new Set(allNodes.filter(n => n.style?.visibility !== 'hidden').map(n => n.id));
    const apToUnitCount = new Map<string, number>();

    // 预处理连线统计
    edges.forEach((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      if (sId.startsWith('accessPoint-') && tId.startsWith('workUnit-')) {
        const key = `${sId}-${tId}`;
        apToUnitCount.set(key, (apToUnitCount.get(key) || 0) + 1);
      }
    });

    // 预处理：为 Unit→AP 的边分配 edgeIndex（控制垂直线 X 偏移）
    // 全局按 source unit 的 x 坐标排序后统一分配，保持所有垂直线间距一致
    const unitToApEdgeIndexMap = new Map<string, number>();
    const visibleEdges = edges.filter(edge => nodeIds.has(edge.source as string) && nodeIds.has(edge.target as string));

    // 收集所有 Unit→AP 的边（同时保留 xBuckets 供后续 optimizeEdgeStyle 使用）
    const unitToApEdges: EdgeData[] = [];
    const xBuckets = new Map<number, EdgeData[]>();
    visibleEdges.forEach((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      if (sId.startsWith('workUnit-') && tId.startsWith('accessPoint-')) {
        unitToApEdges.push(edge);
        const sourceInfo = nodeLayoutInfo.get(sId);
        if (sourceInfo) {
          const bucketKey = sourceInfo.x;
          if (!xBuckets.has(bucketKey)) {
            xBuckets.set(bucketKey, []);
          }
          xBuckets.get(bucketKey)!.push(edge);
        }
      }
    });

    // 按 source unit 的 x 坐标排序，全局分配 edgeIndex
    unitToApEdges.sort((a, b) => {
      const aX = nodeLayoutInfo.get(a.source as string)?.x ?? 0;
      const bX = nodeLayoutInfo.get(b.source as string)?.x ?? 0;
      return aX - bX;
    });
    unitToApEdges.forEach((edge, idx) => {
      unitToApEdgeIndexMap.set(edge.id as string, idx);
    });

    // 按 source unit 分组，按 anchorY 排序后分配阶梯式 unitYOffset
    // 设计原则：同一水平（anchorY 相同）的单元，水平线可能交叉重合，需要错开
    // 第二个单元的 baseOffset = 第一个单元的最大子边跨度 + UNIT_GAP
    const unitToApYIndexMap = new Map<string, number>();
    const unitEdgeGroups = new Map<string, EdgeData[]>();
    unitToApEdges.forEach((edge) => {
      const sId = edge.source as string;
      if (!unitEdgeGroups.has(sId)) unitEdgeGroups.set(sId, []);
      unitEdgeGroups.get(sId)!.push(edge);
    });

    const UNIT_GAP = 5; // 不同冲突单元之间的额外间距（px），和同一单元内子边间距一致
    const SUB_EDGE_GAP = 5; // 同一单元内子边间距（px）
    const unitHeight = 214; // 单元节点高度（与 customEdge.ts 保持一致）

    // 收集所有 unit 的 anchorY，全局按 anchorY 排序后分配偏移
    const unitAnchorInfos: { unitId: string; anchorY: number; edgeCount: number }[] = [];
    unitEdgeGroups.forEach((edges, unitId) => {
      const unitY = nodeLayoutInfo.get(unitId)?.y ?? 0;
      const anchorY = unitY + unitHeight + 5;
      unitAnchorInfos.push({ unitId, anchorY, edgeCount: edges.length });
    });

    // 按 anchorY 排序
    unitAnchorInfos.sort((a, b) => a.anchorY - b.anchorY);

    let currentGroupOffset = 0;
    let lastAnchorY = -Infinity;
    let lastUnitSpan = 0;

    unitAnchorInfos.forEach(({ unitId, anchorY, edgeCount }) => {
      // 当前单元的线跨度 = (edgeCount - 1) * SUB_EDGE_GAP
      const unitSpan = Math.max(edgeCount - 1, 0) * SUB_EDGE_GAP;
      // 当前单元的最小子边偏移（可能为负数）
      const unitMinSubOffset = unitSpan > 0 ? -unitSpan / 2 : 0;

      // 判断是否与上一个单元在同一水平（anchorY 相同或非常接近）
      if (Math.abs(anchorY - lastAnchorY) < 1) {
        // 同一水平：累加偏移，确保当前单元的最小子边线也在前面单元线的下方
        // 前面单元最后一条线的相对位置 = currentGroupOffset + lastUnitSpan/2 (当lastUnitSpan>0) 或 currentGroupOffset (当lastUnitSpan=0)
        // 简化为：currentGroupOffset 需要增加的量 = (lastUnitSpan/2 - unitMinSubOffset) + UNIT_GAP
        // 当 lastUnitSpan=0 时，lastUnitSpan/2=0，增量 = -unitMinSubOffset + UNIT_GAP
        const neededIncrement = Math.max(lastUnitSpan, 0) / 2 - unitMinSubOffset + UNIT_GAP;
        currentGroupOffset += Math.max(lastUnitSpan + UNIT_GAP, neededIncrement);
      } else {
        // 不同水平：重置
        currentGroupOffset = 0;
      }
      lastAnchorY = anchorY;
      lastUnitSpan = unitSpan;

      const edgesInUnit = unitEdgeGroups.get(unitId)!;
      edgesInUnit.forEach((edge) => {
        unitToApYIndexMap.set(edge.id as string, currentGroupOffset);
      });
    });

    // ===== 动态扩展区域高度：估算边水平段最大 Y 并扩展区域 =====
    const areaMaxVerticalY = new Map<string, number>();
    unitToApEdges.forEach((edge) => {
      const sId = edge.source as string;
      const unitInfo = nodeLayoutInfo.get(sId);
      if (!unitInfo) return;
      const unitY = unitInfo.y;
      const unitData = allNodes.find(n => n.id === sId);
      const isDirect = unitData?.data?.is_direct ?? false;
      const unitH = isDirect ? 162 : 214;
      const unitYOffset = unitToApYIndexMap.get(edge.id as string) ?? 0;
      const subIndex = (edge.data as any)?.subIndex ?? 0;
      const totalSubEdges = (edge.data as any)?.totalSubEdges ?? 1;
      const subEdgeYOffset = totalSubEdges > 1 ? (subIndex - (totalSubEdges - 1) / 2) * 5 : 0;
      // firstVerticalY ≈ unitY + unitH + 5(anchor gap) + 15(vertical segment) + subEdgeYOffset + unitYOffset
      const estimatedFirstVerticalY = unitY + unitH + 20 + subEdgeYOffset + unitYOffset;
      const areaId = unitInfo.areaId;
      if (!areaMaxVerticalY.has(areaId) || estimatedFirstVerticalY > areaMaxVerticalY.get(areaId)!) {
        areaMaxVerticalY.set(areaId, estimatedFirstVerticalY);
      }
    });

    // 扩展区域高度（若边水平段超出区域底部）
    areaMaxVerticalY.forEach((maxVerticalY, areaId) => {
      const areaNode = allNodes.find(n => n.id === areaId);
      if (!areaNode) return;
      const areaY = Number(areaNode.style?.y ?? 0);
      const currentH = Number(areaNode.style?.height ?? areaNode.data?.height ?? 0);
      const requiredH = maxVerticalY - areaY + 80;
      if (requiredH > currentH) {
        areaNode.style = { ...areaNode.style, height: requiredH };
        if (areaNode.data) areaNode.data.height = requiredH;
      }
    });

    // 按接入点 ID 全局分配颜色：同一接入点的所有边颜色一致
    // 分层策略：数量少时高对比好分辨，数量多时逐渐降低区分度但仍可辨
    // 第一梯队(1-30)：手工精选高对比色
    const TIER1_COLORS = [
      '#3A84FF', '#FF9C01', '#2DCB8D', '#8B5CF6', '#F36DB9',
      '#14B8A6', '#E97F0A', '#6366F1', '#0EA5E9', '#D97706',
      '#10B981', '#A855F7', '#EC4899', '#0891B2', '#7C3AED',
      '#059669', '#EF4444', '#F59E0B', '#06B6D4', '#8B5E3C',
      '#DC2626', '#84CC16', '#E879F9', '#FB923C', '#22D3EE',
      '#A3E635', '#F43F5E', '#818CF8', '#34D399', '#FBBF24',
    ];
    // 第二梯队(31-100)：HSL色环均匀分布，高饱和高亮度，还好分辨
    // 第三梯队(101+)：HSL色环更细分，饱和度和亮度微调，能分辨一点
    const hslToHex = (h: number, s: number, l: number): string => {
      h = ((h % 360) + 360) % 360;
      s = Math.max(0, Math.min(100, s)) / 100;
      l = Math.max(0, Math.min(100, l)) / 100;
      const c = (1 - Math.abs(2 * l - 1)) * s;
      const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
      const m = l - c / 2;
      let r = 0; let g = 0; let b = 0;
      if (h < 60) { r = c; g = x; b = 0; }
      else if (h < 120) { r = x; g = c; b = 0; }
      else if (h < 180) { r = 0; g = c; b = x; }
      else if (h < 240) { r = 0; g = x; b = c; }
      else if (h < 300) { r = x; g = 0; b = c; }
      else { r = c; g = 0; b = x; }
      const toHex = (v: number) => Math.round((v + m) * 255).toString(16).padStart(2, '0');
      return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
    };
    // 使用黄金角度偏移避免相邻颜色太接近
    const GOLDEN_ANGLE = 137.508;
    const generateColor = (index: number, total: number): string => {
      if (index < TIER1_COLORS.length) {
        return TIER1_COLORS[index];
      }
      // 超出手工精选范围，用算法生成
      const algoIndex = index - TIER1_COLORS.length;
      // 色相：黄金角度均匀分布，起始偏移避开手工色
      const hue = (algoIndex * GOLDEN_ANGLE + 15) % 360;
      if (total <= 100) {
        // 第二梯队：高饱和(70-85%)、中高亮度(45-55%)，交替变化增加区分
        const saturation = 70 + (algoIndex % 3) * 7;
        const lightness = 45 + (algoIndex % 4) * 4;
        return hslToHex(hue, saturation, lightness);
      }
      // 第三梯队：饱和度和亮度有更大波动，尽可能区分
      const saturation = 55 + (algoIndex % 5) * 8;
      const lightness = 38 + (algoIndex % 6) * 5;
      return hslToHex(hue, saturation, lightness);
    };

    // 收集所有不同的接入点 ID，按出现顺序分配颜色
    const apIds: string[] = [];
    visibleEdges.forEach((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      let apId = '';
      if (sId.startsWith('accessPoint-')) apId = sId;
      else if (tId.startsWith('accessPoint-')) apId = tId;
      if (apId && !apIds.includes(apId)) {
        apIds.push(apId);
      }
    });
    const totalAps = apIds.length;
    const apColorMap = new Map<string, string>(); // apId -> color
    apIds.forEach((apId, idx) => {
      apColorMap.set(apId, generateColor(idx, totalAps));
    });
    const edgeColorMap = new Map<string, string>(); // edgeId -> color
    visibleEdges.forEach((edge) => {
      const sId = edge.source as string;
      const tId = edge.target as string;
      let apId = '';
      if (sId.startsWith('accessPoint-')) apId = sId;
      else if (tId.startsWith('accessPoint-')) apId = tId;
      if (apId) {
        edgeColorMap.set(edge.id as string, apColorMap.get(apId)!);
      }
    });

    // 弹性间距：用和 customEdge.ts 相同的公式计算每条垂直线的真实 midX
    // midX = apX + 80 + 40 + edgeIndex * 5 + (isCrossArea ? 15 : 0)
    // 然后确保最大 midX 到右侧最近 unit 有足够间距
    const MIN_GAP = 30; // 最后一条垂直线到右侧 unit 的最小间距
    const nodeShifts = new Map<string, number>();

    // 收集所有 Unit→AP 边的 source area 信息，用于判断跨区域
    const nodeAreaMap = new Map<string, string>();
    allNodes.forEach((node) => {
      if (node.data?.area) {
        nodeAreaMap.set(node.id, String(node.data.area));
      }
    });

    // 按垂直线实际经过的 x 范围找到影响区域
    // 关键：垂直线的 x 取决于 target AP 的 x，不是 source unit 的 x
    // 所以需要找到所有垂直线中，midX 最大且右侧有 unit 的情况
    
    // 收集所有 unit 的 x 坐标（用于判断哪些 unit 在垂直线右边）
    const unitXPositions = new Map<string, number>(); // unitId -> x
    allNodes.forEach((node) => {
      if (node.id?.startsWith('workUnit-') && node.style?.visibility !== 'hidden') {
        unitXPositions.set(node.id, Number(node.style?.x ?? 0));
      }
    });

    // 按"垂直线经过的 x 区间"影响的右侧 unit 列来计算需要的右移
    // 对于每个 unique 的 unit x 坐标（每列 unit），检查是否有垂直线太靠近
    const unitXSet = new Set(unitXPositions.values());
    const sortedUnitXs = Array.from(unitXSet).sort((a, b) => a - b);

    // 构建区域边界信息 map: areaId -> { x, width, right }
    const areaBoundsMap = new Map<string, { x: number; width: number; right: number }>();
    allNodes.forEach((node) => {
      if (node.id?.startsWith('area-') && node.style?.visibility !== 'hidden') {
        const ax = Number(node.style?.x ?? 0);
        const aw = Number(node.style?.width ?? node.data?.width ?? 260);
        areaBoundsMap.set(node.id, { x: ax, width: aw, right: ax + aw });
      }
    });

    // 计算所有垂直线的实际 midX（使用与 customEdge.ts 一致的 edgeIndex）
    const allVertLineMidXs: { midX: number; edgeId: string }[] = [];
    xBuckets.forEach((bucketEdges) => {
      bucketEdges.forEach((edge) => {
        const tId = edge.target as string;
        const sId = edge.source as string;
        const targetInfo = nodeLayoutInfo.get(tId);
        const sourceInfo = nodeLayoutInfo.get(sId);
        if (targetInfo) {
          const apX = targetInfo.x;
          const sourceArea = nodeAreaMap.get(sId) ?? '';
          const targetArea = nodeAreaMap.get(tId) ?? '';
          const isCrossArea = sourceArea !== targetArea && sourceArea !== '' && targetArea !== '';

          let isOnAreaBorder = false;
          if (isCrossArea) {
            const targetAreaBounds = areaBoundsMap.get(targetArea);
            const sourceAreaBounds = areaBoundsMap.get(sourceArea);
            if (targetAreaBounds && sourceAreaBounds && sourceInfo) {
              const apRight = apX + 140;
              const isApNearAreaRight = targetAreaBounds.right - apRight < 60;
              const isUnitNearAreaLeft = sourceInfo.x - sourceAreaBounds.x < 60;
              isOnAreaBorder = isApNearAreaRight && isUnitNearAreaLeft;
            }
          }

          // 跨区域标记保留，但不再额外增加垂直线X偏移
          // 所有垂直线统一按 edgeIndex * 5px 间距排列
          const crossAreaOffset = 0;
          const edgeIndex = unitToApEdgeIndexMap.get(edge.id as string) ?? 0;
          const midX = apX + 80 + 40 + edgeIndex * 5 + crossAreaOffset;
          allVertLineMidXs.push({ midX, edgeId: edge.id as string });
        }
      });
    });

    // 对每列 unit，找到所有 midX < unitX 且距离不够 MIN_GAP 的垂直线
    // 取最靠近该列 unit 的垂直线的 midX，计算需要的右移量
    // 优化：预排序 midX 列表，用二分查找替代线性扫描
    const sortedMidXs = allVertLineMidXs.map(item => item.midX).sort((a, b) => a - b);
    // 收集所有需要右移的 (unitX, needed) 对，最后批量计算总 shift
    const pendingShifts: { unitX: number; needed: number }[] = [];
    sortedUnitXs.forEach((unitX) => {
      // 二分查找最大的 midX < unitX + 已累积的 shift
      const accumulatedShift = nodeShifts.get('__unitX_' + unitX) ?? 0;
      const shiftedUnitX = unitX + accumulatedShift;
      // 找到 sortedMidXs 中 < shiftedUnitX 的最大值（upperBound - 1）
      let lo = 0;
      let hi = sortedMidXs.length - 1;
      let maxMidIdx = -1;
      while (lo <= hi) {
        const mid = (lo + hi) >> 1;
        if (sortedMidXs[mid] < shiftedUnitX) {
          maxMidIdx = mid;
          lo = mid + 1;
        } else {
          hi = mid - 1;
        }
      }

      if (maxMidIdx >= 0) {
        const maxMidXBeforeUnit = sortedMidXs[maxMidIdx];
        const currentGap = shiftedUnitX - maxMidXBeforeUnit;
        if (currentGap < MIN_GAP) {
          const needed = MIN_GAP - currentGap;
          pendingShifts.push({ unitX, needed });
          // 立即更新 __unitX_ 的累积 shift，供后续迭代使用
          nodeShifts.set('__unitX_' + unitX, accumulatedShift + needed);
        }
      }
    });
    // 批量应用 pendingShifts：对每个节点，累加所有 unitX <= nodeX 的 needed
    if (pendingShifts.length > 0) {
      // 按 unitX 排序，计算前缀和
      pendingShifts.sort((a, b) => a.unitX - b.unitX);
      // 对每个节点，二分查找最大的 unitX <= nodeX，累加对应 needed
      const sortedUnitXWithNeeded = pendingShifts;
      allNodes.forEach((node) => {
        if (node.style?.visibility === 'hidden') return;
        const nodeX = Number(node.style?.x ?? 0);
        // 二分查找最后一个 unitX <= nodeX
        let lo = 0;
        let hi = sortedUnitXWithNeeded.length - 1;
        let idx = -1;
        while (lo <= hi) {
          const mid = (lo + hi) >> 1;
          if (sortedUnitXWithNeeded[mid].unitX <= nodeX) {
            idx = mid;
            lo = mid + 1;
          } else {
            hi = mid - 1;
          }
        }
        if (idx >= 0) {
          // 累加所有 needed[0..idx]
          let totalNeeded = 0;
          for (let i = 0; i <= idx; i++) totalNeeded += sortedUnitXWithNeeded[i].needed;
          const currentShift = nodeShifts.get(node.id) ?? 0;
          nodeShifts.set(node.id, currentShift + totalNeeded);
        }
      });
    }

    // 应用右移
    // 同时收集每个区域需要额外扩大的宽度
    const areaExtraWidth = new Map<string, number>();
    allNodes.forEach((node) => {
      const shift = nodeShifts.get(node.id);
      if (shift && shift > 0) {
        const currentX = Number(node.style?.x ?? 0);
        const newX = currentX + shift;
        node.style = { ...node.style, x: newX };
        const info = nodeLayoutInfo.get(node.id);
        if (info) {
          info.x = newX;
        }
        // 记录该节点所属区域需要额外扩大的宽度
        const areaId = (node.data as any)?.area;
        if (areaId) {
          const currentExtra = areaExtraWidth.get(areaId) ?? 0;
          areaExtraWidth.set(areaId, Math.max(currentExtra, shift));
        }
      }
    });
    // 同步扩大区域背景的宽度（区域背景节点的 id 就是 areaId）
    allNodes.forEach((node) => {
      const nodeId = node.id as string;
      if (nodeId.startsWith('area-')) {
        const extra = areaExtraWidth.get(nodeId) ?? 0;
        if (extra > 0) {
          const currentW = Number(node.style?.width ?? node.data?.width ?? 0);
          node.style = { ...node.style, width: currentW + extra };
          if (node.data) node.data.width = currentW + extra;
        }
      }
    });






    const styledEdges = edges.filter(edge => nodeIds.has(edge.source as string) && nodeIds.has(edge.target as string))
      .map((edge) => {
        const sId = edge.source as string;
        const tId = edge.target as string;
        const source = nodeLayoutInfo.get(sId);
        const target = nodeLayoutInfo.get(tId);

        if (!source || !target) return { ...edge, type: 'line' };

        // 识别节点类型
        const isS_AP = sId.startsWith('accessPoint-');
        const isT_AP = tId.startsWith('accessPoint-');
        const isS_Unit = sId.startsWith('workUnit-');
        const isT_Unit = tId.startsWith('workUnit-');
        const isS_Area = sId.startsWith('area-');
        const isT_Area = tId.startsWith('area-');

        // 计算 AP 节点的水平中心 Y 坐标 (AP 高度 32, 中心点 +16)
        const apCenterY = (isS_AP ? source.y : target.y) + this.AP_NODE_HEIGHT / 2;

        // --- 情况 A: AccessPoint 连向 WorkUnit (期望水平直线) ---
        if (isS_AP && isT_Unit) {
          const edgeColor = edgeColorMap.get(edge.id as string) ?? '#C4C6CC';
          const key = `${sId}-${tId}`;
          const count = apToUnitCount.get(key) || 0;

          if (count === 1) {
            const edgeWithStyle = {
              ...edge,
              type: 'custom-edge',
              style: {
                stroke: edgeColor,
                lineWidth: 1,
                endArrow: false,
                startPoint: [source.x, apCenterY],
                endPoint: [target.x, apCenterY],
              },
              startPoint: [source.x, apCenterY],
              endPoint: [target.x, apCenterY],
            };

            return edgeWithStyle;
          }

          // 多条线走折线
          const offset = (count - 1) * 30;
          const turnX = Math.max(source.x, target.x) + 50;
          const startPoint = [source.x, apCenterY];
          const endPoint = [target.x, target.y + 41 + offset];

          return {
            ...edge,
            type: 'polyline',
            style: {
              stroke: edgeColor,
              lineWidth: 1,
              points: [
                startPoint,
                [turnX, apCenterY],
                [turnX, target.y + 41 + offset],
                endPoint,
              ],
            },
            startPoint: startPoint,
            endPoint: endPoint,
          };
        }

        // --- 情况 B: WorkUnit 连向 AccessPoint (从连接锚点出发) ---
        if (isS_Unit && isT_AP) {
          const edgeColor = edgeColorMap.get(edge.id as string) ?? '#C4C4CC';
          return {
            ...edge,
            type: 'custom-edge',
            style: {
              stroke: edgeColor,
              lineWidth: 1,
              endArrow: {
                path: 'M 0,0 L 4,2 L 4,-2 Z',
                fill: edgeColor
              },
              edgeIndex: unitToApEdgeIndexMap.get(edge.id as string) ?? 0,
              unitYIndex: unitToApYIndexMap.get(edge.id as string) ?? 0,
            },
          };
        }

        // --- 情况 C: 涉及区域节点的连线 ---
        if (isS_Area || isT_Area) {
          // 计算区域节点的端点
          const sourceY = isS_Area ? source.y + 20 : // 区域节点假设高度40，中心+20
                         isS_AP ? source.y + this.AP_NODE_HEIGHT / 2 : 
                         source.y + this.UNIT_NODE_HEIGHT / 2;
          
          const targetY = isT_Area ? target.y + 20 : // 区域节点假设高度40，中心+20
                         isT_AP ? target.y + this.AP_NODE_HEIGHT / 2 : 
                         target.y + this.UNIT_NODE_HEIGHT / 2;
          
          const sourceWidth = isS_Area ? 200 : // 区域节点假设宽度200
                             isS_AP ? this.AP_COL_WIDTH : 
                             this.UNIT_COL_WIDTH;
          
          const areaStartPoint = [source.x + sourceWidth, sourceY];
          const areaEndPoint = [target.x, targetY];
          
          return {
            ...edge,
            type: 'custom-edge',
            style: {
              stroke: '#999',
              lineWidth: 1,
              strokeDasharray: [5, 5], // 虚线表示区域连接
              startPoint: areaStartPoint,
              endPoint: areaEndPoint,
            },
            startPoint: areaStartPoint,
            endPoint: areaEndPoint,
          };
        }

        // --- 情况 D: 默认其他连线 ---
        // 计算默认端点，确保所有边缘都有自定义端点
        const sourceY = source.y + (isS_AP ? this.AP_NODE_HEIGHT / 2 : this.UNIT_NODE_HEIGHT / 2);
        const targetY = target.y + (isT_AP ? this.AP_NODE_HEIGHT / 2 : this.UNIT_NODE_HEIGHT / 2);
        
        const defaultStartPoint = [
          source.x + (isS_AP ? this.AP_COL_WIDTH : this.UNIT_COL_WIDTH), 
          sourceY
        ];
        const defaultEndPoint = [
          target.x, 
          targetY
        ];

        return {
          ...edge,
          type: 'custom-edge',
          style: {
            stroke: '#C4C6CC', // 修改：使用设计稿的颜色
            lineWidth: 1, // 改细：从2改为1
            startPoint: defaultStartPoint,
            endPoint: defaultEndPoint,
          },
          // 同时在根级别也设置这些属性
          startPoint: defaultStartPoint,
          endPoint: defaultEndPoint,
        };
      });

    // ========== 修复问题1：按子节点实际位置重新计算区域包围盒 ==========
    // 构建 areaId -> children 映射（只包含可见节点）
    const areaIdToChildren = new Map<string, typeof allNodes>();
    allNodes.forEach((node) => {
      const areaId = node.data?.area;
      if (areaId && !String(node.id).startsWith('area-') && node.style?.visibility !== 'hidden') {
        if (!areaIdToChildren.has(areaId)) areaIdToChildren.set(areaId, []);
        areaIdToChildren.get(areaId)!.push(node);
      }
    });

    const RE_PAD = 20;
    const RE_PAD_TOP = 60;

    areaIdToChildren.forEach((children, areaId) => {
      if (children.length === 0) return;
      const areaNode = allNodes.find(n => n.id === areaId);
      if (!areaNode) return;

      const areaX = Number(areaNode.style?.x ?? 0);
      const areaY = Number(areaNode.style?.y ?? 0);

      let maxX = -Infinity;
      let maxY = -Infinity;
      children.forEach((node) => {
        const x = Number(node.style?.x ?? 0);
        const y = Number(node.style?.y ?? 0);
        const w = Number(node.style?.width ?? node.data?.width ?? 0);
        const h = Number(node.style?.height ?? node.data?.height ?? 0);
        if (x + w > maxX) maxX = x + w;
        if (y + h > maxY) maxY = y + h;
      });

      const requiredW = maxX - areaX + RE_PAD;
      const requiredH = maxY - areaY + RE_PAD_TOP + RE_PAD;

      const currentW = Number(areaNode.style?.width ?? areaNode.data?.width ?? 0);
      const currentH = Number(areaNode.style?.height ?? areaNode.data?.height ?? 0);

      const newW = Math.max(currentW, requiredW);
      const newH = Math.max(currentH, requiredH);

      if (newW > currentW || newH > currentH) {
        areaNode.style = { ...areaNode.style, width: newW, height: newH };
        if (areaNode.data) { areaNode.data.width = newW; areaNode.data.height = newH; }
      }
    });

    return { edges: styledEdges, apColorMap };
  }
}
