import { type Group } from '@antv/g';
import { type Point, Polyline } from '@antv/g6';

export default class CustomEdge extends Polyline {
  private finalStartEndPoint: [Point, Point] = [[0, 0], [0, 0]];

  protected getEndpoints(attributes: any) {
    const defaultEndpoints = super.getEndpoints(attributes);
    this.finalStartEndPoint = defaultEndpoints;
    return defaultEndpoints;
  }

  protected getKeyStyle(attributes: any) {
    const style = super.getKeyStyle(attributes);
    // 从布局阶段设置的 edge data style 中读取 stroke 颜色
    const edgeData = this.context?.graph?.getEdgeData(this.id);
    const dataStroke = edgeData?.style?.stroke as string;
    if (dataStroke) {
      return { ...style, stroke: dataStroke };
    }
    return style;
  }

  protected getKeyPath(attributes: any) {
    const keyPathStyle = super.getKeyPath(attributes) as any;
    
    // 检查是否有布局算法设置的startPoint和endPoint
    const startPoint = this.parsedAttributes.startPoint;
    const endPoint = this.parsedAttributes.endPoint;
    
    // 如果布局算法设置了具体的起点和终点，直接使用（主要是AP->Unit的水平直线）
    if (startPoint && endPoint) {
      return [
        ['M', startPoint[0], startPoint[1]],
        ['L', endPoint[0], endPoint[1]]
      ];
    }
    
    // 获取源节点和目标节点ID
    const sourceNodeId = this.parsedAttributes.sourceNode;
    const targetNodeId = this.parsedAttributes.targetNode;
    
    // 检查是否是Unit->AP的边（需要L型路径）
    const isUnitToAPEdge = sourceNodeId?.startsWith('workUnit-') && targetNodeId?.startsWith('accessPoint-');
    
    // 对Unit->AP的边应用L型路径
    if (isUnitToAPEdge && this.context && this.context.graph) {
      const sourceNodeData = this.context.graph.getNodeData(sourceNodeId);
      const targetNodeData = this.context.graph.getNodeData(targetNodeId);
      
      if (sourceNodeData && targetNodeData) {
        // 获取子边信息（同一单元到不同接入点的多条边时的偏移）
        const edgeData = this.context.graph.getEdgeData(this.id);
        const subIndex = (edgeData?.data as any)?.subIndex ?? 0;
        const totalSubEdges = (edgeData?.data as any)?.totalSubEdges ?? 1;
        // 纵向偏移：多条子边时，以中心为基准上下分布，间距3px
        const subEdgeYOffset = totalSubEdges > 1
          ? (subIndex - (totalSubEdges - 1) / 2) * 3
          : 0;

        // 计算锚点坐标
        const unitX = sourceNodeData.style?.x || 0;
        const unitY = sourceNodeData.style?.y || 0;
        const unitWidth = 240;
        const unitHeight = sourceNodeData.data?.is_direct ? 162 : 214;
        
        const anchorX = unitX + unitWidth / 2;
        const anchorY = unitY + unitHeight + 5 + subEdgeYOffset;
        
        const apX = targetNodeData.style?.x || 0;
        const apY = targetNodeData.style?.y || 0;
        const apHeight = 30;
        
        const targetX = apX + 80;
        const targetY = apY + apHeight / 2 - 5 + subEdgeYOffset;
        
        // 从布局阶段传入的 edgeIndex 计算偏移
        // - edgeIndex: 按 source unit 的 x 分桶，控制垂直线 X 偏移
        // - apEdgeIndex: 按目标 AP 分组，控制水平线 Y 偏移（避免同AP多条入边水平线重合）
        const edgeIndex = this.parsedAttributes.edgeIndex ?? 0;
        const apEdgeIndex = this.parsedAttributes.apEdgeIndex ?? 0;
        const edgeOffset = edgeIndex * 5;
        const yOffset = apEdgeIndex * 5; // 用 apEdgeIndex 控制水平线Y偏移
        
        // 检查是否跨区域连接
        const sourceArea = sourceNodeData.data?.area;
        const targetArea = targetNodeData.data?.area;
        const isCrossArea = sourceArea !== targetArea;
        
        // 判断是否在两个区域的边上：接入点靠近其区域右边界，且源单元靠近其区域左边界
        let isOnAreaBorder = false;
        if (isCrossArea && this.context.graph) {
          const targetAreaNode = this.context.graph.getNodeData(`${targetArea}`);
          const sourceAreaNode = this.context.graph.getNodeData(`${sourceArea}`);
          if (targetAreaNode && sourceAreaNode) {
            const targetAreaRight = Number(targetAreaNode.style?.x || 0) + Number(targetAreaNode.style?.width || targetAreaNode.data?.width || 260);
            const sourceAreaLeft = Number(sourceAreaNode.style?.x || 0);
            // 接入点右边缘（apX + AP宽度140）接近区域右边界，且源单元在其区域左侧
            const apRight = apX + 140;
            const isApNearAreaRight = targetAreaRight - apRight < 60;
            const isUnitNearAreaLeft = unitX - sourceAreaLeft < 60;
            isOnAreaBorder = isApNearAreaRight && isUnitNearAreaLeft;
          }
        }
        
        // 跨区域时增加额外间距（两种情况独立处理）
        // 1. 普通跨区域：+15，避免垂直线与其他区域的元素重叠
        // 2. 跨区域且在区域边界上：+50，垂直线需要绕过区域边界
        const crossAreaOffset = isCrossArea ? (isOnAreaBorder ? 50 : 15) : 0;
        
        // 统一垂直线X坐标：接入点targetX + 40 + 偏移量 + 跨区域偏移
        const midX = targetX + 40 + edgeOffset + crossAreaOffset;
        
        const firstVerticalY = anchorY + 15 + yOffset;   // 从锚点向下15px（缩短了）+ 垂直偏移
        
        // 圆弧半径（自适应）
        const maxRadius = 8;
        
        // 计算路径方向
        const isGoingRight = midX > anchorX;
        const isGoingDown = targetY > firstVerticalY;
        const isTargetRight = targetX > midX;
        
        // 计算各段可用距离，自适应调整圆弧半径避免间隙
        const verticalDist1 = Math.abs(firstVerticalY - anchorY); // 第一段垂直距离
        const horizontalDist1 = Math.abs(midX - anchorX); // 第一段水平距离
        const verticalDist2 = Math.abs(targetY - firstVerticalY); // 第二段垂直距离
        const horizontalDist2 = Math.abs(targetX - midX); // 最后水平距离
        
        // 圆弧半径取各段可用距离的一半（保证不超出），且不超过 maxRadius
        const radius = Math.min(
          maxRadius,
          verticalDist1 / 2,   // 第一段垂直可用
          horizontalDist1 / 2, // 第一段水平可用
          verticalDist2 / 2,   // 第二段垂直可用（关键：防止中间间隙）
          horizontalDist2 / 2, // 最后水平可用
        );
        // 确保半径至少为 1px，防止退化
        const safeRadius = Math.max(radius, 1);
        
        // 为每个圆弧单独计算正确的sweep-flag
        // 第一个圆弧：垂直→水平
        const arc1SweepFlag = isGoingRight ? 0 : 1;  // 向右时用0，向左时用1
        
        // 第二个圆弧：水平→垂直  
        const arc2SweepFlag = (isGoingRight && isGoingDown) || (!isGoingRight && !isGoingDown) ? 1 : 0;
        
        // 第三个圆弧：垂直→水平
        const arc3SweepFlag = (isGoingDown && isTargetRight) || (!isGoingDown && !isTargetRight) ? 0 : 1;
        
        // 使用A命令实现圆弧（使用自适应 safeRadius）
        const zPath = [
          ['M', anchorX, anchorY],                    // 起点：单元底部中心
          ['L', anchorX, firstVerticalY - safeRadius],   // 垂直线到第一个圆弧前
          
          // 第一个圆弧：垂直→水平
          ['A', safeRadius, safeRadius, 0, 0, arc1SweepFlag, 
           anchorX + (isGoingRight ? safeRadius : -safeRadius), firstVerticalY],
          
          ['L', midX - (isGoingRight ? safeRadius : -safeRadius), firstVerticalY],  // 水平线到第二个圆弧前
          
          // 第二个圆弧：水平→垂直
          ['A', safeRadius, safeRadius, 0, 0, arc2SweepFlag, 
           midX, firstVerticalY + (isGoingDown ? safeRadius : -safeRadius)],
          
          ['L', midX, targetY - (isGoingDown ? safeRadius : -safeRadius)],  // 垂直线到第三个圆弧前
          
          // 第三个圆弧：垂直→水平
          ['A', safeRadius, safeRadius, 0, 0, arc3SweepFlag, 
           midX + (isTargetRight ? safeRadius : -safeRadius), targetY],
          
          ['L', targetX, targetY]                     // 最后的水平线到接入点
        ];
        
        return zPath;
      }
    }
    
    return keyPathStyle;
  }

  drawCircleToStartNode(container: Group) {
    const { isSourceSubProcess, sourceNodeData, targetNodeData } = this.parsedAttributes as any;

    if (isSourceSubProcess && sourceNodeData?.level < targetNodeData?.level) {
      this.upsert(
        'subprocessStartCricle',
        'circle',
        {
          cx: this.finalStartEndPoint[0][0],
          cy: this.finalStartEndPoint[0][1],
          fill: '#fff',
          r: 3,
          stroke: '#C4C6CC',
        },
        container,
      );
    }
  }

  render(attributes = this.parsedAttributes as any, container: Group) {
    super.render(attributes, container);
    this.drawCircleToStartNode(container);
  }
}