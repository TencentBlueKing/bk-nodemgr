import { type Group } from '@antv/g';
import { type Point, Polyline } from '@antv/g6';

export default class CustomEdge extends Polyline {
  private finalStartEndPoint: [Point, Point] = [[0, 0], [0, 0]];

  protected getEndpoints(attributes: any) {
    const defaultEndpoints = super.getEndpoints(attributes);
    this.finalStartEndPoint = defaultEndpoints;
    return defaultEndpoints;
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
        // 计算锚点坐标
        const unitX = sourceNodeData.style?.x || 0;
        const unitY = sourceNodeData.style?.y || 0;
        const unitWidth = 240;
        const unitHeight = sourceNodeData.data?.is_direct ? 162 : 214;
        
        const anchorX = unitX + unitWidth / 2;
        const anchorY = unitY + unitHeight + 5;
        
        const apX = targetNodeData.style?.x || 0;
        const apY = targetNodeData.style?.y || 0;
        const apHeight = 30;
        
        const targetX = apX + 80;
        const targetY = apY + apHeight / 2 - 5;
        
        // 计算边的偏移量，按同一单元的接入点顺序累计
        let edgeOffset = 0;
        let yOffset = 0;
        
        // 获取节点编号
        const unitMatch = sourceNodeId.match(/workUnit-(\d+)/);
        const apMatch = targetNodeId.match(/accessPoint-(\d+)/);
        
        if (unitMatch && apMatch) {
          const unitNum = parseInt(unitMatch[1]);
          const apNum = parseInt(apMatch[1]);
          
          // 按接入点编号作为该单元的第几条边来累计偏移
          edgeOffset = apNum * 3; // 第1个接入点+0，第2个+3，第3个+6...
          yOffset = apNum * 3;    // 垂直偏移也按接入点顺序累计
        }
        
        // 检查是否跨区域连接
        const sourceArea = sourceNodeData.data?.area;
        const targetArea = targetNodeData.data?.area;
        const isCrossArea = sourceArea !== targetArea;
        
        // 跨区域时增加额外间距，避免垂直线重合
        const crossAreaOffset = isCrossArea ? 15 : 0; // 跨区域时额外增加15px间距
        
        // 统一垂直线X坐标：接入点targetX + 40 + 偏移量 + 跨区域偏移
        const midX = targetX + 40 + edgeOffset + crossAreaOffset;
        
        const firstVerticalY = anchorY + 15 + yOffset;   // 从锚点向下15px（缩短了）+ 垂直偏移
        
        // 圆弧半径
        const radius = 8;
        
        // 计算路径方向
        const isGoingRight = midX > anchorX;
        const isGoingDown = targetY > firstVerticalY;
        const isTargetRight = targetX > midX;
        
        // 为每个圆弧单独计算正确的sweep-flag
        // 第一个圆弧：垂直→水平
        const arc1SweepFlag = isGoingRight ? 0 : 1;  // 向右时用0，向左时用1
        
        // 第二个圆弧：水平→垂直  
        const arc2SweepFlag = (isGoingRight && isGoingDown) || (!isGoingRight && !isGoingDown) ? 1 : 0;
        
        // 第三个圆弧：垂直→水平
        const arc3SweepFlag = (isGoingDown && isTargetRight) || (!isGoingDown && !isTargetRight) ? 0 : 1;
        
        // 使用A命令实现圆弧
        const zPath = [
          ['M', anchorX, anchorY],                    // 起点：单元底部中心
          ['L', anchorX, firstVerticalY - radius],   // 垂直线到第一个圆弧前
          
          // 第一个圆弧：垂直→水平
          ['A', radius, radius, 0, 0, arc1SweepFlag, 
           anchorX + (isGoingRight ? radius : -radius), firstVerticalY],
          
          ['L', midX - (isGoingRight ? radius : -radius), firstVerticalY],  // 水平线到第二个圆弧前
          
          // 第二个圆弧：水平→垂直
          ['A', radius, radius, 0, 0, arc2SweepFlag, 
           midX, firstVerticalY + (isGoingDown ? radius : -radius)],
          
          ['L', midX, targetY - (isGoingDown ? radius : -radius)],  // 垂直线到第三个圆弧前
          
          // 第三个圆弧：垂直→水平
          ['A', radius, radius, 0, 0, arc3SweepFlag, 
           midX + (isTargetRight ? radius : -radius), targetY],
          
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