import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { BaseLayout } from '@antv/g6';

import { NodeType } from './config';

export interface HorizontalHierarchyLayoutOptions {
  collapsed?: boolean; // 控制孤立区域是否收起
}

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
  private readonly UNIT_COL_WIDTH = 290;
  private readonly AP_COL_WIDTH = 270;
  private readonly AREA_MIN_WIDTH = 300;
  private readonly AREA_HEIGHT_EXTRA = 20;

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
          height: Math.max(areaSize.height, currentH)
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
      col.nodes.forEach((node) => {
        totalHeight += (node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT) + this.NODE_SPACING_ROW;
      });
      return Math.max(max, totalHeight > 0 ? totalHeight - this.NODE_SPACING_ROW : 0);
    }, 0);
    const areaHeight = this.AREA_PADDING * 2 + maxColHeight + this.AREA_HEIGHT_EXTRA + 60;

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
    const childStartY = areaY + this.AREA_PADDING + 60;
    
    let currentColX = childStartX;

    columns.forEach((col) => {
      const { nodes, width } = col;
      nodes.forEach((node, rowIndex) => {
        const nodeHeight = node.type === NodeType.NET_WORK_UNIT ? this.UNIT_NODE_HEIGHT : this.AP_NODE_HEIGHT;
        
        // 算法计算出的理论位置
        const calculatedChildX = currentColX;
        const calculatedChildY = childStartY + rowIndex * (nodeHeight + this.NODE_SPACING_ROW);

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
    // 过滤掉不可见的节点
    const nodeIds = new Set(allNodes.filter(n => n.style?.visibility !== 'hidden').map(n => n.id));
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
