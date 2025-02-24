import { BaseLayout, ComboData, EdgeData, GraphData, NodeData } from "@antv/g6";

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
 * 1. 支持按combo分组布局
 * 2. 同一combo内的节点按水平方向排列
 * 3. 不同combo的节点按垂直方向错开排列
 * 4. 对于default类型节点特殊处理,保证其布局的合理性
 * 5. 边的连线会根据节点位置自动添加控制点,形成美观的折线效果
 */

export default class ResourceLayout extends BaseLayout {
  id = 'custom';
  combosMap = new Map<string, number>();

  static readonly nodeSize = 64;
  static readonly defaultStyle = {
    horizontalDistance: 226 + ResourceLayout.nodeSize,   // 节点间水平间距
    verticalDistance: 108 + ResourceLayout.nodeSize,   // 节点间垂直间距
  };

  static readonly DEFAULT_NODE_NAME = 'default';
  static readonly maxDefaultNodesPerCombo = 2;
  static readonly singleDefaultWidth = 280;
  static readonly moreDefaultWidth = 520;

  private nodes: NodeData[] = [];
  private edges: EdgeData[] = [];
  private combos: ComboData[] = [];
  
  async execute(data: GraphData) {

    this.nodes = data.nodes as NodeData[];
    this.edges = data.edges as EdgeData[];
    this.combos = data.combos as ComboData[];
    // 统计各combo下nodes数量
    this.setComboCount(this.nodes, this.combos);
    // 根据combo类型排序
    this.sortNodesByCombo(this.nodes);
    // nodes 位置计算
    const nodesStyle = this.getNodesStyle(this.nodes);
    // 折线radius拐点处理
    // const edgesStyle = this.getEdgesStyle(this.edges, nodesStyle);

    return {
      nodes: nodesStyle,
    };
  }
  
  /**
   * @description 根据combo类型排序
   * @param nodes 节点数据
   */
  private sortNodesByCombo(nodes: NodeData[]) {
    // 首先按combo分组
    const comboGroups = new Map<string, NodeData[]>();
    // 将节点按combo分组
    for (const node of nodes) {
      const comboId = node.combo as string;
      if (!comboGroups.has(comboId)) {
        comboGroups.set(comboId, []);
      }
      comboGroups.get(comboId)?.push(node);
    }
    // 对每个combo内的节点进行排序
    // 将default节点放在前面，其他节点按照id排序
    for (const [comboId, comboNodes] of comboGroups) {
      comboNodes.sort((a, b) => {
        // default节点优先
        if (a.data?.name === ResourceLayout.DEFAULT_NODE_NAME && 
            b.data?.name !== ResourceLayout.DEFAULT_NODE_NAME) {
          return -1;
        }
        if (a.data?.name !== ResourceLayout.DEFAULT_NODE_NAME && 
            b.data?.name === ResourceLayout.DEFAULT_NODE_NAME) {
          return 1;
        }
        // 同为default或同为非default，按id排序
        return (a.id as string).localeCompare(b.id as string);
      });
    }
    // 清空原数组
    nodes.length = 0;

    // 按combo顺序重新填充节点数组
    // 先获取所有combo的id并排序
    const comboIds = Array.from(comboGroups.keys()).sort();

    // 按combo顺序添加节点
    for (const comboId of comboIds) {
      const comboNodes = comboGroups.get(comboId) || [];
      for (const node of comboNodes) {
        nodes.push(node);
      }
    }
  }

  /**
   * 
   * @param description 统计每个combo下default节点的数量，记录在combosMap
   */
  private setComboCount(nodes: NodeData[], combos: ComboData[]) {
    for (const o of combos) {
      const curComboCount = this.countDefaultNodesInCombo(nodes, o.id);
      this.combosMap.set(o.id, curComboCount);
    }
  }

  /**
   * 计算combo内default节点的数量
   */
  private countDefaultNodesInCombo(nodes: NodeData[], comboId: string): number {
    return nodes.filter(n => 
      n.combo === comboId && 
      n.data?.name === ResourceLayout.DEFAULT_NODE_NAME
    ).length;
  }

  /**
   * 计算节点样式和位置
   * @param nodes 节点数据
   */
  private getNodesStyle(nodes: NodeData[]) {
    const { verticalDistance } = ResourceLayout.defaultStyle;
    const maxDefaultNodesCount = ResourceLayout.maxDefaultNodesPerCombo;
    // 记录已处理过的combo，用于判断是否进入新的combo
    const processedCombos: string[] = [];
    // 已处理的combo数量
    let comboCounter = 0;
    // 当前combo内的node序号，每进入新combo时重置
    let comboNodeIndex = 0;
    // combo左侧的x坐标
    let comboStartX = 0;
    // combo右侧的x坐标
    let comboEndX = 0;

    return nodes.map((node) => {
      const currentCombo = node.combo as string;
      // 获取当前combo中default节点的数量（最多支持 maxDefaultNodesCount 个）
      const defaultNodeCount = Math.min(this.combosMap.get(currentCombo) as number, maxDefaultNodesCount);
      // 是否为单default
      const isSingleDefault = defaultNodeCount < maxDefaultNodesCount;
      // 根据default数量，获取当前combo的宽度
      const curComboWidth = isSingleDefault ? ResourceLayout.singleDefaultWidth : ResourceLayout.moreDefaultWidth;

      // 每进入一个新combo，初始化数据
      if (!processedCombos.includes(currentCombo)) {
        comboNodeIndex = 0; // 进入新combo，清空节点索引
        comboCounter += 1; // combo数量累加
        processedCombos.push(currentCombo); // 记录新combo的名称

        comboStartX = comboEndX + 10; // 新combo的左侧x位置( 10为margin-right )
        comboEndX = comboStartX + curComboWidth;
      }
      comboNodeIndex += 1; // 节点索引累加

      let segmentWidth = 0;
      let adjustedXPosition = 0; // node的水平偏移
      if (!isSingleDefault) {
        segmentWidth = curComboWidth / (maxDefaultNodesCount + 1); // 将combo width根据nodes数量平分区域
        if (comboNodeIndex % maxDefaultNodesCount === 0) {
          // 若当前节点为水平方向的最后一个节点
          adjustedXPosition = maxDefaultNodesCount * segmentWidth + 5;
        } else {
          adjustedXPosition = (comboNodeIndex % maxDefaultNodesCount) * segmentWidth + 5;
        }
      } else {
        // 若为单default combo 需要获取上一个combo的default count
        // 因为两者(多default combo 与 单default combo) width不同，margin-right有较大的偏差，以下逻辑用于处理该偏差
        const previousDefaultCount = this.combosMap.get(processedCombos[comboCounter - 2]);
        // comboCounter的初始下标为1，因此需要获取previous combo需要 -2

        let adjustmentOffset = 0;
        // 1. 如果previous combo 为single default && current combo 也为single default 无需给出偏移量
        // 2. 如果previous combo 不为single default && current combo 为single default 给出偏移量
        // 3. 如果previous combo 为single default && current combo 不为single default 给出偏移量
        if (previousDefaultCount && previousDefaultCount > defaultNodeCount) {
          adjustmentOffset = 5;
        } else {
          adjustmentOffset = 10;
        }
        adjustedXPosition = ResourceLayout.singleDefaultWidth / 2 + adjustmentOffset;
      }

      // node的垂直偏移，若combo为单个default，直接使用节点索引; 若为多default，需要考虑node在图中是否换行
      const verticalPos = isSingleDefault ? comboNodeIndex : Math.ceil(comboNodeIndex / maxDefaultNodesCount);
      // 节点位置偏下，向上移动
      let adjustedYPosition = -ResourceLayout.nodeSize;

      const xPosition = comboStartX + adjustedXPosition;
      const yPosition = verticalPos * verticalDistance + adjustedYPosition;

      return {
        ...node,
        style: {
          x: xPosition,
          y: yPosition,
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

    return edges.map(edge => {
      const sourceNode = nodeMap.get(edge.source);
      const targetNode = nodeMap.get(edge.target);

      if (!sourceNode || !targetNode) {
        return { ...edge };
      }
      
      let style = { ...edge.style };
      // 当源节点和目标节点位置不在同一水平线或垂直线上，给折线增加控制点，形成折线radius样式
      if (sourceNode.style.x !== targetNode.style.x && sourceNode.style.y !== targetNode.style.y) {
        let x, y = 0;
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
      }

      return {
        ...edge,
        style
      };
    });
  }
}
