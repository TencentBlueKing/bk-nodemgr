/**
 * 共享 Agent 节点状态定义
 * 覆盖后端全量 9 个 NodeStatus，提供统一的状态文案、图标类、语义色调
 */

// 后端全量 NodeStatus 联合类型
export const AGENT_NODE_STATUS_LIST = [
  'init',
  'running',
  'damaged',
  'busy',
  'starting',
  'upgrade',
  'stopping',
  'uninit',
  'unknown',
] as const;

export type AgentNodeStatus = (typeof AGENT_NODE_STATUS_LIST)[number];

// 状态语义分组
export type StatusSemantic = 'success' | 'danger' | 'progress' | 'neutral';

// 状态元数据：i18n key、icon 类、色调、语义
interface AgentStatusMeta {
  i18nKey: string;
  iconClass: string;
  color: string;
  semantic: StatusSemantic;
}

const STATUS_META_MAP: Record<AgentNodeStatus, AgentStatusMeta> = {
  running: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.running',
    iconClass: 'nc-running',
    color: '#3FC06D',
    semantic: 'success',
  },
  damaged: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.damaged',
    iconClass: 'nc-damaged',
    color: '#EA3636',
    semantic: 'danger',
  },
  busy: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.busy',
    iconClass: 'nc-busy',
    color: '#FF9C01',
    semantic: 'progress',
  },
  starting: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.starting',
    iconClass: 'nc-starting',
    color: '#3A84FF',
    semantic: 'progress',
  },
  upgrade: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.upgrade',
    iconClass: 'nc-upgrade',
    color: '#3A84FF',
    semantic: 'progress',
  },
  stopping: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.stopping',
    iconClass: 'nc-stopping',
    color: '#FF9C01',
    semantic: 'progress',
  },
  init: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.init',
    iconClass: 'nc-init',
    color: '#979BA5',
    semantic: 'neutral',
  },
  uninit: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.uninit',
    iconClass: 'nc-uninit',
    color: '#979BA5',
    semantic: 'neutral',
  },
  unknown: {
    i18nKey: 'platform.nodeMan.agentNodeStatus.unknown',
    iconClass: 'nc-unknown',
    color: '#C4C6CC',
    semantic: 'neutral',
  },
};

/**
 * 获取状态元数据，未知状态返回 undefined
 */
export const getAgentStatusMeta = (status: string): AgentStatusMeta | undefined =>
  STATUS_META_MAP[status as AgentNodeStatus];

/**
 * 已知状态列表（用于"全部状态"筛选项）
 */
export const KNOWN_STATUS_LIST = AGENT_NODE_STATUS_LIST as unknown as string[];

/**
 * 获取状态对应的 icon class 全名（含 nodeman-icon 前缀）
 */
export const getStatusIconClass = (status: string): string => {
  const meta = getAgentStatusMeta(status);
  return meta ? `nodeman-icon ${meta.iconClass} status-icon` : 'nodeman-icon nc-unknown status-icon';
};

/**
 * 获取状态对应的色值
 */
export const getStatusColor = (status: string): string => {
  const meta = getAgentStatusMeta(status);
  return meta?.color ?? '#C4C6CC';
};
