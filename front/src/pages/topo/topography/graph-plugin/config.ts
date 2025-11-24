export enum NodeType {
  NET_WORK_AREA = 'net-work-area-node',
  NET_WORK_UNIT = 'net-work-unit-node',
  ACCESS_POINT = 'access-point-node',
  SERVER = 'server-node'
}

export const iconPathMap = {
  Machine: '/icons/machine.svg',
  Direct: '/icons/direct.svg',
  Proxy: '/icons/proxy.svg',
} as const;

// 节点状态类型
export enum NodeStatus {
  HEALTHY = 'healthy',
  ABNORMAL = 'abnormal'
}

// 单元类型
export enum UnitType {
  DIRECT = 'direct',    // 直连单元
  INDIRECT = 'indirect' // 非直连单元
}
