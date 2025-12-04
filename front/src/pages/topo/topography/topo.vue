<template>
  <Loading mode="spin" theme="primary" :loading="isLoading">
    <div class="min-h-[calc(100vh_-_104px)]">
      <!-- 下拉选择器 -->
      <Select
        class="w-[240px] absolute z-[2] m-[24px]"
        v-model="regionList"
        :clearable="false"
        all-option-id="all"
        collapse-tags
        filterable
        multiple
        show-all
        @change="handleSelectChange"
      >
        <Select.Group :label="$t('topoManager.topo.select.default')">
          <Select.Option
            :key="defaultNetWorkarea?.bk_networkarea_id"
            :id="defaultNetWorkarea?.bk_networkarea_id"
            :name="`[${defaultNetWorkarea?.bk_networkarea_id}] ${defaultNetWorkarea?.bk_networkarea_name}`"
          >
          </Select.Option>
        </Select.Group>
        <Select.Group :label="$t('topoManager.topo.select.other')">
          <Select.Option
            v-for="item in netWorkAreaList"
            :key="item.bk_networkarea_id"
            :id="item.bk_networkarea_id"
            :name="`[${item.bk_networkarea_id}] ${item.bk_networkarea_name}`"
          >
          </Select.Option>
        </Select.Group>
      </Select>

      <div id="nodemgr-g6-container"></div>

      <!-- 工具栏 -->
      <CustomToolbar @trigger-tools="handleClickTool"></CustomToolbar>

      <!-- =========== 新增：菜单弹窗锚点 =========== -->
      <div
        v-show="menuState.visible"
        :style="{
          position: 'absolute',
          left: `${menuState.x}px`,
          top: `${menuState.y}px`,
          width: '1px',
          height: '1px',
          zIndex: 999,
          pointerEvents: 'none' /* 防止遮挡 */
        }"
      >
        <Popover
          v-model:is-show="menuState.showPopover"
          trigger="manual"
          theme="light"
          placement="right"
          :arrow="false"
          :offset="0"
        >
          <!-- 锚点内容为空 -->
          <div style="width: 1px; height: 1px"></div>

          <template #content>
            <!-- 开启 pointer-events 以便点击菜单 -->
            <div class="pointer-events-auto min-w-[120px] py-1 bg-white rounded shadow-md border border-[#DCDEE5]">
              <div class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer" @click="handleMenuAction('detail')">
                查看详情
              </div>
              <div class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer" @click="handleMenuAction('edit')">
                编辑单元
              </div>
              <!-- 可以根据 menuState.nodeData 判断是否显示删除 -->
              <div
                class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer text-red-500"
                @click="handleMenuAction('delete')">
                删除单元
              </div>
            </div>
          </template>
        </Popover>
      </div>

      <!-- =========== 新增：接入点详情气泡 (Hover) =========== -->
      <div
        v-show="detailState.visible"
        :style="{
          position: 'absolute',
          left: `${detailState.x}px`,
          top: `${detailState.y + 30}px`,
          width: '1px',
          height: '1px',
          zIndex: 1000,
          pointerEvents: 'none'
        }"
      >
        <bk-popover
          v-model:is-show="detailState.showPopover"
          trigger="manual"
          theme="light"
          placement="top"
          :arrow="true"
        >
          <!-- 锚点 -->
          <div style="width: 1px; height: 1px"></div>

          <template #content>
            <Table
              :data="detailState.data.endpointsData"
              :min-width="600"
              :max-height="800"
              auto-resize
              :show-overflow="false"
              :row-config="{ isHover: true, height: 'auto' }"
            >
              <TableColumn field="accesspoint_name" title="接入点名称" :min-width="180">
                <template #default="{ row }">
                  {{ row.name }}
                </template>
              </TableColumn>
              <TableColumn field="cluster" title="cluster" :min-width="200">
                <template #default="{ row }">
                  <span style="white-space: pre-line;">{{ row.endpoints.cluster.join('\n') }}</span>
                </template>
              </TableColumn>
              <TableColumn field="file" title="file" :min-width="200">
                <template #default="{ row }">
                  {{ row.endpoints.file.join('\n') }}
                </template>
              </TableColumn>
              <TableColumn field="data" title="data" :min-width="200">
                <template #default="{ row }">
                  {{ row.endpoints.data.join('\n') }}
                </template>
              </TableColumn>
            </Table>
          </template>
        </bk-popover>
      </div>
    </div>
  </Loading>
</template>

<script setup lang="ts">
import { Loading, Popover, Select } from 'bkui-vue';
import { throttle } from 'lodash';
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import type { EdgeData, GraphData, IElementDragEvent, NodeData } from '@antv/g6';
import {
  EdgeEvent,
  ExtensionCategory,
  Graph,
  NodeEvent,
  register,
} from '@antv/g6';
import { Table, TableColumn } from '@blueking/table';

import AccessPointNode from './graph-plugin/access-point-node';
import { NodeStatus, NodeType, UnitType } from './graph-plugin/config';
import CustomToolbar from './graph-plugin/custom-toolbar.vue';
import HorizontalHierarchyLayout from './graph-plugin/HorizontalHierarchyLayout';
import NetworkAreaNode from './graph-plugin/net-work-area-node';
import NetWorkUnitNode from './graph-plugin/net-work-unit-node';

import type { TopoGraphNodeCountRespNodeInfo } from '@/@types/topo';
import useMinLengthRef from '@/composables/use-min-length-ref';
import { useTopoStore } from '@/stores/topo';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const {
  handleFetchTopoWorkareaList,
  handleFetchAllWorkUnit,
} = useTopoStore();
const topoStore = useTopoStore();
const workareaStore = useWorkareaStore();
const router = useRouter();

// 选择器逻辑
const regionList = useMinLengthRef(
  ['all'] as Array<string | number>,
  t('topoManager.topo.select.tips', { count: 1 }),
);
const netWorkAreaList = ref<Partial<any>[]>([]);
const defaultNetWorkarea = ref<Partial<any>>();

let graph: Graph;
const graphData: GraphData = reactive({
  nodes: [],
  edges: [],
});

const isLoading = ref(false);
const workAreaPrefix = 'workArea-';
const workUnitPrefix = 'workUnit-';

function handleSelectChange() {
  filterAreaNodes();
}

// ------------------ 菜单状态与辅助函数 ------------------
const menuState = reactive({
  visible: false,    // 锚点 div 是否渲染
  showPopover: false, // Popover 组件是否显示
  x: 0,
  y: 0,
  currentNodeId: '', // 当前操作的节点 ID, 例如 "workUnit-123"
  nodeData: null as any, // 暂存节点数据，方便菜单操作使用
});

// 关闭菜单
function closeMenu() {
  menuState.showPopover = false;
  setTimeout(() => {
    menuState.visible = false;
    menuState.currentNodeId = '';
    menuState.nodeData = null;
  }, 200);
}

// 菜单项点击回调
function handleMenuAction(action: string) {
  const { currentNodeId, nodeData } = menuState;

  // 打印看看，应该就是你发的那串 JSON
  console.log('当前操作节点:', currentNodeId);
  console.log('节点数据:', nodeData);

  switch (action) {
    case 'detail':
      console.log(`查看详情: ${nodeData.name}`);
      // 这里的 nodeData.name 就是 "default-test-1"
      break;

    case 'edit':
      console.log('编辑单元');
      // 如果你需要 ID，可以从 currentNodeId 解析，或者看看 nodeData 里有没有存 ID
      // 你的数据里好像只有 bk_networkunit_id 在 links 里或者需要从 nodeId 解析
      break;

    case 'delete':
      // 你的数据里有 "area": "workArea-0"
      console.log(`从区域 ${nodeData.area} 删除单元`);

      // 执行删除逻辑...
      // deleteUnit(currentNodeId);
      break;
  }

  closeMenu();
}

// 监听画布点击/拖拽/缩放，关闭菜单
function initGlobalListeners() {
  // 点击空白处关闭
  graph.on('click', (e: any) => {
    if (!e.targetType || e.targetType === 'canvas') closeMenu();
  });
  // 拖拽或缩放画布时关闭，防止菜单位置错乱
  graph.on('drag', closeMenu);
  graph.on('zoom', closeMenu);
}

// ------------------ 3. 详情弹窗状态 (Hover) ------------------
const detailState = reactive({
  visible: false,     // 锚点是否存在
  showPopover: false, // Popover 是否显示
  x: 0,
  y: 0,
  data: null as any,  // 存储接入点数据
  timer: null as any,  // 防抖定时器
});

// 鼠标移入节点 (显示详情)
function handleNodeEnter(evt: any) {
  const { path, canvas } = evt;

  // 1. 获取 Node ID
  // G6 事件冒泡，path[0] 是图形，path[1] 通常是 Node Group (如果不确定，向上查找)
  let nodeId = evt.id;
  if (!nodeId && path) {
    // 尝试找 ID 包含 accessPoint 的对象
    const nodeObj = path.find((p: any) => p.id && String(p.id).includes('accessPoint-'));
    nodeId = nodeObj?.id;
  }

  // 2. 只有接入点才显示详情
  if (nodeId && nodeId.startsWith('accessPoint-')) {
    // 清除隐藏定时器 (防止快速移动时闪烁)
    if (detailState.timer) clearTimeout(detailState.timer);

    const nodeData = graph.getNodeData(nodeId);

    // 坐标转换
    const viewportPoint = graph.getViewportByCanvas([canvas.x, canvas.y]);

    // 更新状态
    detailState.x = viewportPoint[0];
    detailState.y = viewportPoint[1]; // 显示在鼠标位置
    detailState.data = nodeData?.data || {};
    detailState.visible = true;
    detailState.showPopover = true;
  }
}

// 鼠标移出节点 (隐藏详情)
function handleNodeLeave(evt: any) {
  // 延迟隐藏，给用户一点缓冲时间
  detailState.timer = setTimeout(() => {
    detailState.showPopover = false;
    detailState.visible = false;
  }, 100);
}

// 计算节点的包围盒，并更新区域节点及相邻区域的位置
// ---------------------- 最终完整版逻辑 ----------------------

// 计算节点的包围盒，并联动更新区域及其邻居节点（含邻居内部节点）
const updateAreaByChildNodes = throttle((movedNodeId: string, areaId: string) => {
  if (!graph || !areaId) return;

  // 1. 获取所有节点数据
  const allNodes = graph.getNodeData();

  // 【优化】预先构建 "区域ID -> 子节点列表" 的映射，避免在循环中重复遍历
  // 这一步对于性能至关重要，防止 O(n^2) 复杂度
  const childrenByAreaMap = new Map<string, any[]>();

  allNodes.forEach((node) => {
    // 假设 node.data.area 存的是区域ID
    const pId = node.data?.area as string;
    if (pId && node.id !== pId) { // 排除区域节点自己
      if (!childrenByAreaMap.has(pId)) {
        childrenByAreaMap.set(pId, []);
      }
      childrenByAreaMap.get(pId)?.push(node);
    }
  });

  // 2. 找到当前正在变动的区域节点
  const areaNode = graph.getNodeData(areaId);
  if (!areaNode) return;

  const oldWidth = Number(areaNode.style?.width || areaNode.data?.width || 0);
  const oldHeight = Number(areaNode.style?.height || areaNode.data?.height || 0);
  const currentAreaX = Number(areaNode.style?.x ?? 0);
  const currentAreaY = Number(areaNode.style?.y ?? 0);

  // 获取当前区域的子节点
  const currentAreaChildren = childrenByAreaMap.get(areaId) || [];
  if (currentAreaChildren.length === 0) return;

  // 3. 计算当前区域新的包围盒
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;

  currentAreaChildren.forEach((node) => {
    const x = Number(node.style?.x ?? 0);
    const y = Number(node.style?.y ?? 0);
    let width = 0;
    let height = 0;

    if (node.type === NodeType.NET_WORK_UNIT) {
      width = NetWorkUnitNode.gridWidth * 2;
      const isDirect = node.data.is_direct as boolean;
      height = NetWorkUnitNode.getNodeTotalHeight(isDirect);
    } else if (node.type === NodeType.ACCESS_POINT) {
      width = AccessPointNode.nodeWidth;
      height = AccessPointNode.nodeHeight;
    } else {
      width = Number(node.style?.width || node.data?.width || 0);
      height = Number(node.style?.height || node.data?.height || 0);
    }

    if (x < minX) minX = x;
    if (y < minY) minY = y;
    if (x + width > maxX) maxX = x + width;
    if (y + height > maxY) maxY = y + height;
  });

  const PADDING = 20;
  const PADDING_TOP = 60;

  const newX = minX - PADDING;
  const newY = minY - PADDING_TOP;
  const newWidth = (maxX - minX) + (PADDING * 2);
  const newHeight = (maxY - minY) + PADDING_TOP + PADDING;

  // 4. 计算变化量
  const deltaWidth = newWidth - oldWidth;
  const deltaHeight = newHeight - oldHeight;

  // 阈值判断：如果变化极小且位置没变，直接返回
  if (Math.abs(deltaWidth) < 1 && Math.abs(deltaHeight) < 1) {
    if (Math.abs(newX - currentAreaX) > 1 || Math.abs(newY - currentAreaY) > 1) {
      graph.updateNodeData([{
        id: areaId,
        style: { x: newX, y: newY },
      }]);
    }
    return;
  }

  // 5. 准备更新队列
  const updates = [];

  // (1) 更新当前区域自己 (尺寸 + 位置)
  updates.push({
    id: areaId,
    data: { width: newWidth, height: newHeight },
    style: { x: newX, y: newY, width: newWidth, height: newHeight },
  });

  // (2) 调整其他区域及其内部子节点的位置
  if (Math.abs(deltaWidth) > 0.1 || Math.abs(deltaHeight) > 0.1) {
    const otherAreaNodes = allNodes.filter(n => n.type === NodeType.NET_WORK_AREA && n.id !== areaId);

    otherAreaNodes.forEach((otherArea) => {
      const originalOtherX = Number(otherArea.style?.x ?? 0);
      const originalOtherY = Number(otherArea.style?.y ?? 0);

      let shiftX = 0;
      let shiftY = 0;

      // 逻辑A: 水平移动 (右侧的所有区域)
      if (originalOtherX > currentAreaX + 10) {
        shiftX = deltaWidth;
      }

      // 逻辑B: 垂直移动 (同列下方区域)
      const isSameColumn = Math.abs(originalOtherX - currentAreaX) < 50;
      if (isSameColumn && originalOtherY > currentAreaY) {
        shiftY = deltaHeight;
      }

      // 如果需要移动
      if (Math.abs(shiftX) > 0.1 || Math.abs(shiftY) > 0.1) {
        // A. 更新邻居区域本身的位置
        updates.push({
          id: otherArea.id,
          style: {
            x: originalOtherX + shiftX,
            y: originalOtherY + shiftY,
          },
        });

        // B. 【关键新增】更新邻居区域内部所有子节点的位置
        // 从预先构建的 Map 中获取子节点
        const children = childrenByAreaMap.get(otherArea.id);
        if (children && children.length > 0) {
          children.forEach((child) => {
            const childX = Number(child.style?.x ?? 0);
            const childY = Number(child.style?.y ?? 0);
            updates.push({
              id: child.id,
              style: {
                x: childX + shiftX,
                y: childY + shiftY,
              },
            });
          });
        }
      }
    });
  }

  // 6. 执行批量更新
  graph.updateNodeData(updates);
}, 16);

// 拖拽事件回调
function handleNodeDrag(e: any) {
  const targetNode = e.target;
  // 只有拖拽 Node 类型才处理
  if (!targetNode || targetNode.id.startsWith('edge-')) return;

  // 获取所属区域 ID
  const areaId = targetNode.data?.area;
  if (areaId) {
    updateAreaByChildNodes(targetNode.id, areaId);
  }
}
// 初始化拓扑图
function handleInitTopo() {
  if (graph) return;

  const container = document.getElementById('nodemgr-g6-container');
  if (!container) return;

  const minimapContainer = document.getElementById('minimap-container');
  if (minimapContainer) {
    minimapContainer.innerHTML = '';
  }

  graph = new Graph({
    container: 'nodemgr-g6-container',
    width: container.clientWidth,
    height: container.clientHeight,
    zoom: 0.8,
    zoomRange: [0.2, 2],
    autoFit: 'view',
    animation: false,
    autoResize: true,
    data: graphData,
    node: {},
    edge: {},
    behaviors: [
      'scroll-canvas', 'drag-canvas', 'zoom-canvas',
      {
        type: 'drag-element',
        key: 'drag-element-1',
        enableAnimation: true,
        dropEffect: 'move',
        shadow: true, // 启用拖拽幽灵节点
        // 自定义幽灵节点样式
        shadowFill: '#E8F3FF',
        shadowFillOpacity: 0.4,
        shadowStroke: '#1890FF',
        shadowStrokeOpacity: 0.8,
        shadowLineDash: [4, 4],
        // 允许拖拽的元素类型：节点+边
        enable: (event: IElementDragEvent) => ['node', 'edge'].includes(event.targetType) && !event.target.id.includes('workArea'),
        // 拖拽时鼠标样式
        cursor: {
          default: 'default',
          grab: 'grab',
          grabbing: 'grabbing',
        },
      },
    ],
    plugins: [
      {
        type: 'minimap',
        container: 'minimap-container',
        size: [240, 148],
      },
    ],
    layout: {
      type: 'horizontal-hierarchy-layout',
    },
    background: '#FAFBFD',
  });

  graph.render();
  // 监听节点拖拽过程
  graph.on(NodeEvent.DRAG, handleNodeDrag);
}

function reRender() {
  if (!graph) return;
  graph.setData(graphData);
  graph.layout();
  graph.render();
}

// 工具栏处理
function handleClickTool(code: string, value?: number) {
  switch (code) {
    case 'center':
      graph?.fitCenter?.();
      break;
    case 'mapSize':
      graph?.zoomTo?.(value);
      break;
    case 'init':
      graph?.fitView?.();
      graph?.zoomTo?.(0.6);
      break;
  }
}

function handleHoverEdge(evt: Event) {
  const { target } = evt;
  graph.setElementState(target?.id, 'highlight');
}

function handleLeaveEdge(evt: Event) {
  const { target } = evt;
  graph.setElementState(target?.id, '');
}

// 节点点击处理函数：分发菜单点击和跳转逻辑
function handleNodeClick(evt: any) {
  const { target, path, canvas } = evt;

  // 1. 获取点击图形的 className (确认是否点了菜单热区)
  const clickedShape = path?.[0];
  const className = clickedShape?.config?.className;

  // 2. 获取节点 ID (workUnit-xxx)
  let nodeId = target.id;
  if (!nodeId && path) {
    const nodeObj = path.find((p: any) => p.id && String(p.id).includes(workUnitPrefix));
    nodeId = nodeObj?.id;
  }

  // ==================== 菜单点击逻辑 ====================
  if (className === 'menu-hit-area') {
    // 坐标转换
    const viewportPoint = graph.getViewportByCanvas([canvas.x, canvas.y]);

    // 【核心】获取该节点的业务数据, 不传nodeId也可以
    const fullNodeData = graph.getNodeData(nodeId);

    // 更新菜单状态
    menuState.x = viewportPoint[0];
    menuState.y = viewportPoint[1];
    menuState.visible = true;
    menuState.currentNodeId = nodeId;

    // 【新增】将数据存入 state，供菜单操作使用
    // 注意：G6 v5 获取的数据结构通常是 { id: '...', data: { ... }, style: ... }
    // 你发的那个 JSON 应该是 fullNodeData.data
    menuState.nodeData = fullNodeData?.data || {};

    // 延迟显示 Popover
    setTimeout(() => {
      menuState.showPopover = true;
    });

    return; // 阻止跳转
  }

  // ==================== 页面跳转逻辑 ====================
  if (nodeId && nodeId.includes(workUnitPrefix)) {
    const nodeData = graph.getNodeData(nodeId);
    if (!nodeData) return;

    // 解析 ID
    const workareaId = Number(String(nodeData.data.area).replace(workAreaPrefix, ''));
    const workUnitId = Number(nodeId.replace(workUnitPrefix, ''));

    // 特殊跳转逻辑：Agent
    // 注意：检查 path[0] 是否存在及其 config 属性
    if (path?.[0]?.config?.className?.includes('agent')) {
      router.push({
        name: 'agent',
        query: {
          bk_networkarea_id: workareaId,
          bk_networkunit_id: workUnitId,
        },
      });
      return;
    }

    // 默认跳转：区域详情
    router.push({
      name: 'workareaDetail',
      params: {
        workarea: workareaId,
        workUnit: workUnitId,
      },
    });
  }
}
// 连接关系生成函数
const generateEdgesFromUnitData = () => {
  const edges = [];

  // 构建接入点映射
  const accessPointMap = new Map();
  topoStore.accessPointData.forEach((ap) => {
    accessPointMap.set(ap.bk_accesspoint_id, ap);
  });

  topoStore.workUnitByArea.forEach((unit) => {
    const unitId = `workUnit-${unit.bk_networkunit_id}`;

    if (!unit.is_direct) {
      // 非直连单元：连接到上游接入点
      if (unit.links?.cluster?.accesspoint_id !== null) {
        const apId = `accessPoint-${unit.links.cluster.accesspoint_id}`;
        edges.push({
          id: `edge-${apId}-${unitId}`,
          target: apId,
          source: unitId,
        });
      }
    }

    // 单元提供接入点：单元→接入点
    if (unit.accesspoints && unit.accesspoints.length > 0) {
      unit.accesspoints.forEach((apInfo) => {
        const apId = `accessPoint-${apInfo.accesspoint_id}`;
        edges.push({
          id: `edge-${unitId}-${apId}`,
          target: unitId,
          source: apId,
        });
      });
    }
  });

  return edges;
};

// 初始化区域数据
const initAreaData = async () => {
  await Promise.all([
    handleFetchTopoWorkareaList(),
    handleFetchAllWorkUnit(),
  ]);

  netWorkAreaList.value = topoStore.allWorkareaList
    .filter(item => item.bk_networkarea_id !== 0)
    .map(item => ({
      bk_networkarea_id: item.bk_networkarea_id,
      bk_networkarea_name: item.bk_networkarea_name,
    }));

  const defaultArea = topoStore.allWorkareaList.find(item => item.bk_networkarea_id === 0);
  defaultNetWorkarea.value = {
    bk_networkarea_id: defaultArea!.bk_networkarea_id,
    bk_networkarea_name: defaultArea!.bk_networkarea_name,
  };

  // 直接调用filterAreaNodes来设置初始的节点数据，避免重复的数据处理
  filterAreaNodes();
};

// 筛选区域数据（删除 Server 相关逻辑）
function filterAreaNodes() {
  const selectedValues = regionList.value;
  let targetAreaIds: number[] = [];

  if (selectedValues.length === 1 && selectedValues[0] === 'all') {
    targetAreaIds = topoStore.allWorkareaList.map(item => item.bk_networkarea_id);
  } else {
    targetAreaIds = selectedValues.flatMap((areaId) => {
      const numId = Number(areaId);
      const arr = topoStore.areaDependencyMap.get(numId) || [numId];
      return arr;
    });
  }

  const filteredAreaNodes = topoStore.allWorkareaList
    .filter(item => targetAreaIds.includes(item.bk_networkarea_id))
    .map(item => ({
      id: workAreaPrefix + item.bk_networkarea_id,
      data: {
        name: item.bk_networkarea_name,
        bk_networkarea_id: item.bk_networkarea_id,
        bk_networkarea_name: item.bk_networkarea_name,
        width: 600,
        height: 400,
      },
      type: NodeType.NET_WORK_AREA,
    }));

  const filteredUnitNodes = topoStore.workUnitByArea
    .filter(item => targetAreaIds.includes(item.bk_networkarea_id))
    .map((item) => {
      const nodeId = `workUnit-${item.bk_networkunit_id}`;
      return {
        id: nodeId,
        data: {
          name: item.bk_networkunit_name,
          unitType: item.is_direct ? 'direct' : 'indirect',
          area: workAreaPrefix + item.bk_networkarea_id,
          is_direct: item.is_direct,
          direct_endpoints: item.direct_endpoints || { cluster: ['未知'], count: 0 },
          accesspoints: item.accesspoints || [],
          links: item.links || {},
          running_proxy: item.running_proxy || 0,
          total_proxy: item.total_proxy || 0,
          running_agent: item.running_agent || 0,
          total_agent: item.total_agent || 0,
          cycle_times: item.cycle_times || ['0', '0', '0'],
          is_healthy: item.is_healthy,
        },
        type: NodeType.NET_WORK_UNIT,
      };
    });

  const filteredAccessPointNodes = topoStore.accessPointData
    .filter(item => targetAreaIds.includes(item.bk_networkarea_id))
    .map((item) => {
      const nodeId = `accessPoint-${item.bk_accesspoint_id}`;
      return {
        id: nodeId,
        data: {
          name: item.name,
          downstreamUnits: item.downstreamUnits || 0,
          area: workAreaPrefix + item.bk_networkarea_id,
          bk_networkunit_id: item.bk_networkunit_id,
          type: item.type,
          endpoints: item.endpoints,
          endpointsData: [{
            name: item.name,
            endpoints: item.endpoints,
          }],
        },
        type: NodeType.ACCESS_POINT,
      };
    });


  const existingNodeIds = new Set([
    ...filteredAreaNodes.map(n => n.id),
    ...filteredUnitNodes.map(n => n.id),
    ...filteredAccessPointNodes.map(n => n.id),
  ]);

  const generateValidEdges = () => {
    const originalEdges = generateEdgesFromUnitData();
    return originalEdges.filter(edge => existingNodeIds.has(edge.source as string) && existingNodeIds.has(edge.target as string));
  };
  const validUnitEdges = generateValidEdges();

  graphData.nodes = [
    ...filteredAreaNodes,
    ...filteredUnitNodes,
    ...filteredAccessPointNodes,
  ];
  graphData.edges = [
    ...validUnitEdges,
  ];
  graph = null;
  handleInitTopo();
}

// 注册Graph插件
function handleRegistryCategory() {
  register(ExtensionCategory.NODE, NodeType.NET_WORK_AREA, NetworkAreaNode);
  register(ExtensionCategory.NODE, NodeType.NET_WORK_UNIT, NetWorkUnitNode);
  register(ExtensionCategory.NODE, NodeType.ACCESS_POINT, AccessPointNode);
  register(ExtensionCategory.LAYOUT, 'horizontal-hierarchy-layout', HorizontalHierarchyLayout);
}

// 窗口大小变化处理
function handleResize() {
  if (graph) {
    const container = document.getElementById('nodemgr-g6-container');
    if (container) {
      graph.setSize(container.clientWidth, container.clientHeight);
      graph.fitView();
    }
  }
}

onMounted(async () => {
  isLoading.value = true;
  // 确保收藏状态同步
  workareaStore.syncFavoriteWorkareaList();

  // 如果有收藏的管控区域，则选中收藏的区域，否则选中所有区域
  if (workareaStore.favoriteWorkareaList.length > 0) {
    regionList.value = workareaStore.favoriteWorkareaList.map(id => Number(id));
  } else {
    regionList.value = ['all'];
  }
  try {
    // 注册插件
    handleRegistryCategory();
    // 初始化数据
    await initAreaData();
    // 初始化拓扑图
    handleInitTopo();
    initGlobalListeners(); // 注册全局关闭菜单的监听

    // 事件监听
    window.addEventListener('resize', handleResize);
    graph.on(EdgeEvent.POINTER_OVER, handleHoverEdge);
    graph.on(EdgeEvent.POINTER_OUT, handleLeaveEdge);

    // 监听节点的移入移出 (显示详情)
    graph.on(NodeEvent.POINTER_OVER, handleNodeEnter);
    graph.on(NodeEvent.POINTER_OUT, handleNodeLeave);
    // 【修改】使用新的整合函数
    graph.on(NodeEvent.CLICK, handleNodeClick);
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  graph?.off();
  graph?.destroy();
});
</script>

<style scoped>
#nodemgr-g6-container {
  width: 100%;
  height: calc(100vh - 104px);
  position: relative;
  overflow: hidden;
}
</style>
