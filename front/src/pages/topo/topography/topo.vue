<template>
  <Loading mode="spin" theme="primary" :loading="isLoading">
    <div :class="[mainStore.noticeShow ? 'min-h-[calc(100vh-144px)]' : 'min-h-[calc(100vh-104px)]', 'relative']">
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
            :name="defaultNetWorkarea?.bk_networkarea_name"
          >
            <div class="w-[180px] flex">
              <Button
                text
                class="mr-[8px] w-[18px] favorited-item">
                <i class="nodeman-icon nc-collect text-[#C4C6CC] text-[18px]">
                </i>
              </Button>
              <span>
                {{ `[${defaultNetWorkarea?.bk_networkarea_id}] ${defaultNetWorkarea?.bk_networkarea_name}` }}
              </span>
            </div>
          </Select.Option>
        </Select.Group>
        <Select.Group :label="$t('topoManager.topo.select.other')">
          <Select.Option
            v-for="item in sortedNetWorkAreaList"
            :key="item.bk_networkarea_id"
            :id="item.bk_networkarea_id"
            :name="item.bk_networkarea_name"
          >
            <div class="w-[180px] flex favorited-item">
              <Button
                text
                class="mr-[8px] w-[18px]"
                @click.stop="handleCollect(item.bk_networkarea_id)">
                <i
                  class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
                  v-if="collectList.includes(item.bk_networkarea_id)">
                </i>
                <i
                  class="nodeman-icon nc-not-favorited text-[#C4C6CC] text-[18px] hidden"
                  v-else>
                </i>
              </Button>
              <div
                class="w-[154px] truncate"
                @mouseenter="handleTextMouseenter($event, item.bk_networkarea_id)"
                v-bk-tooltips="{
                  content: item.bk_networkarea_name,
                  placement: 'top',
                  boundary: 'body',
                  extCls: 'force-tooltip-z-index',
                  disabled: !textOverflowMap[item.bk_networkarea_id] // 没超长就禁用 Tooltip
                }"
              >
                {{ `[${item.bk_networkarea_id}] ${item.bk_networkarea_name}` }}
              </div>
            </div>
          </Select.Option>
        </Select.Group>
      </Select>

      <!-- 【新增】收起/展开无关联区域按钮 -->
      <div class="absolute top-[24px] right-[80px] z-[2]">
        <Button
          theme="primary"
          :outline="true"
          @click="toggleIsolatedAreas"
        >
          {{ isIsolatedCollapsed ? $t('topoManager.topo.expand') : $t('topoManager.topo.collapse') }}
        </Button>
      </div>

      <div
        id="nodemgr-g6-container"
        :class="[mainStore.noticeShow ? 'h-[calc(100vh-144px)]' : 'h-[calc(100vh-144px)]']">
      </div>

      <!-- 工具栏 -->
      <CustomToolbar @trigger-tools="handleClickTool"></CustomToolbar>

      <!-- =========== 新增：菜单弹窗锚点 =========== -->
      <!-- <div
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
          <div style="width: 1px; height: 1px"></div>

          <template #content>
            <div class="pointer-events-auto min-w-[120px] py-1 bg-white rounded shadow-md border border-[#DCDEE5]">
              <div class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer" @click="handleMenuAction('detail')">
                查看详情
              </div>
              <div class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer" @click="handleMenuAction('edit')">
                编辑单元
              </div>
              <div
                class="px-3 py-2 hover:bg-[#F0F1F5] text-sm cursor-pointer text-red-500"
                @click="handleMenuAction('delete')">
                删除单元
              </div>
            </div>
          </template>
        </Popover>
      </div> -->

      <!-- =========== 新增：接入点详情气泡 (Hover) =========== -->
      <div
        v-show="detailState.visible"
        :style="{
          position: 'absolute',
          left: `${detailState.x}px`,
          top: `${detailState.y}px`,
          width: '1px',
          height: '1px',
          zIndex: 1000,
          pointerEvents: 'none'
        }"
      >
        <Popover
          :is-show="detailState.showPopover"
          trigger="manual"
          theme="light"
          placement="right"
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
              <TableColumn field="accesspoint_name" :title="$t('topoManager.topo.accessPointName')" :min-width="180">
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
        </Popover>
      </div>
    </div>
  </Loading>
</template>

<script setup lang="ts">
import { Button, Loading, OverflowTitle, Popover, Select } from 'bkui-vue';
import { throttle } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
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
import { useMainStore } from '@/stores/main';
import { useTopoStore } from '@/stores/topo';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const {
  handleFetchTopoWorkareaList,
  handleFetchAllWorkUnit,
} = useTopoStore();
const topoStore = useTopoStore();
const mainStore = useMainStore();
const workareaStore = useWorkareaStore();
const router = useRouter();

// 选择器逻辑
const regionList = useMinLengthRef(
  ['all'] as Array<string | number>,
  t('topoManager.topo.select.tips', { count: 1 }),
);
const netWorkAreaList = ref<Partial<any>[]>([]);
const defaultNetWorkarea = ref<Partial<any>>();

// 计算属性：排序后的其他区域列表（按照优先级顺序：默认区域第一，勾选优先级第二，收藏优先级第三，ID从大到小第四）
const sortedNetWorkAreaList = computed(() => [...netWorkAreaList.value].sort((a, b) => {
  // bk_networkarea_id为0的始终排在最前面
  if (a.bk_networkarea_id === 0) return -1;
  if (b.bk_networkarea_id === 0) return 1;

  // 勾选的区域排在前面（regionList中存在的区域）
  const aIsSelected = regionList.value.includes(a.bk_networkarea_id);
  const bIsSelected = regionList.value.includes(b.bk_networkarea_id);
  if (aIsSelected && !bIsSelected) return -1;
  if (!aIsSelected && bIsSelected) return 1;

  // 收藏的区域排在前面
  const aIsFavorite = workareaStore.favoriteWorkareaList.includes(a.bk_networkarea_id);
  const bIsFavorite = workareaStore.favoriteWorkareaList.includes(b.bk_networkarea_id);
  if (aIsFavorite && !bIsFavorite) return -1;
  if (!aIsFavorite && bIsFavorite) return 1;

  // 其他情况按ID从大到小排序
  return b.bk_networkarea_id - a.bk_networkarea_id;
}));

// 收藏管控区域
const collectList = ref<number[]>(JSON.parse(localStorage.getItem('collect_workarea') || '[]'));
const handleCollect = (val: number) => {
  if (collectList.value.includes(val)) {
    collectList.value = collectList.value.filter(item => item !== val);
  } else {
    collectList.value.push(val);
  }
  localStorage.setItem('collect_workarea', JSON.stringify(collectList.value));
  // 同步更新store中的收藏状态
  workareaStore.syncFavoriteWorkareaList();
};

// 1. 定义一个响应式对象，用来存储每个区域 ID 是否超长
// Key: networkarea_id, Value: boolean (true=超长, false=未超长)
const textOverflowMap = reactive<Record<number, boolean>>({});

// 2. 鼠标移入时的检测函数
const handleTextMouseenter = (e: MouseEvent, id: number) => {
  const el = e.target as HTMLElement;
  // 核心逻辑：内容宽度 > 可视宽度 = 发生了截断
  const isOverflow = el.scrollWidth > el.clientWidth;

  // 更新状态
  textOverflowMap[id] = isOverflow;
};

// 【新增】孤立区域折叠状态
const isIsolatedCollapsed = ref(false);

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

// 【新增】切换折叠状态的方法
function toggleIsolatedAreas() {
  isIsolatedCollapsed.value = !isIsolatedCollapsed.value;
  if (graph) {
    graph.layout({
      type: 'horizontal-hierarchy-layout',
      collapsed: isIsolatedCollapsed.value,
    });
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
        enable: (event: IElementDragEvent) => {
          if (!['node', 'edge'].includes(event.targetType) || event.target.id.includes('workArea')) return false;

          const nodeId = event.target.id;

          // 单元节点：通过计算鼠标是否在顶部 36px 范围内
          if (String(nodeId).includes('workUnit')) {
            const { y: mouseY } = event.canvas;
            const bounds = event.target.getRenderBounds();
            const nodeTopY = bounds.min[1];
            const headerHeight = 36;

            // 纯数学比对，瞬间返回 true/false，不需要 await
            if (mouseY >= nodeTopY && mouseY <= nodeTopY + headerHeight + 2) {
              return true;
            }
            return false;
          }
          return true;
        },
        // 拖拽时鼠标样式
        cursor: {
          default: 'default',
          grab: 'pointer',
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
      // 【修改】传入初始折叠状态
      collapsed: isIsolatedCollapsed.value,
    },
    background: '#FAFBFD',
  });

  graph.render();
  graph.on(NodeEvent.DRAG, handleNodeDrag); // 注册拖拽事件
  graph.on(NodeEvent.DRAG_END, handleNodeDragEnd); // 注册拖拽结束事件
  initGlobalListeners(); // 注册全局关闭菜单的监听
  graph.on(EdgeEvent.POINTER_OVER, handleHoverEdge);
  graph.on(EdgeEvent.POINTER_OUT, handleLeaveEdge);

  // 监听节点的移入移出 (显示详情)
  graph.on(NodeEvent.POINTER_OVER, handleNodeEnter);
  graph.on(NodeEvent.POINTER_OUT, handleNodeLeave);
  // 【修改】使用新的整合函数
  graph.on(NodeEvent.CLICK, handleNodeClick);
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
  // console.log('当前操作节点:', currentNodeId);
  // console.log('节点数据:', nodeData);

  switch (action) {
    case 'detail':
      // console.log(`查看详情: ${nodeData.name}`);
      // 这里的 nodeData.name 就是 "default-test-1"
      break;

    case 'edit':
      // console.log('编辑单元');
      // 如果你需要 ID，可以从 currentNodeId 解析，或者看看 nodeData 里有没有存 ID
      // 你的数据里好像只有 bk_networkunit_id 在 links 里或者需要从 nodeId 解析
      break;

    case 'delete':
      // 你的数据里有 "area": "workArea-0"
      // console.log(`从区域 ${nodeData.area} 删除单元`);

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
  const { path, canvas, target } = evt;
  if (!canvas) return;

  // 1. 校验类名 (保持不变)
  const shapeClass = target.className || target.attributes?.class || path[0]?.config?.className;
  if (shapeClass !== 'ap-info-icon' && shapeClass !== 'info-hit-area') return;

  // 2. 获取 ID (保持不变)
  let nodeId = evt.id;
  if (!nodeId && path) {
    const nodeObj = path.find((p: any) => p.id && String(p.id).includes('accessPoint-'));
    nodeId = nodeObj?.id;
  }

  if (nodeId && nodeId.startsWith('accessPoint-')) {
    const nodeData = graph.getNodeData(nodeId);

    // --- 【简化】坐标计算：改为右侧 ---
    const bbox = target.getRenderBounds();

    // 取图标的【最右侧】X 坐标
    const rightX = bbox.max[0];
    // 取图标的【垂直中心】Y 坐标
    const centerY = (bbox.min[1] + bbox.max[1]) / 2;

    // 转为屏幕坐标
    const viewportPoint = graph.getViewportByCanvas([rightX, centerY]);

    // x: 图标右边缘 + 10px 间距
    detailState.x = viewportPoint[0];
    // y: 垂直居中
    detailState.y = viewportPoint[1];

    detailState.data = nodeData?.data || {};
    detailState.visible = true;
    detailState.showPopover = true;
  }
}

// 鼠标移出节点 (隐藏详情)
function handleNodeLeave(evt: any) {
  // 目前没有但是之后要设计：移到pop上时候pop可以不消失
  detailState.showPopover = false;
  detailState.visible = false;
}

// ---------------------- 节点拖拽逻辑 (保持原样) ----------------------

// 计算节点的包围盒，并联动更新区域及其邻居节点（含邻居内部节点）
const execUpdateArea = (areaId: string) => {
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
};

// 2. 节流版 (保持简单)
const throttledUpdateArea = throttle((areaId: string) => {
  execUpdateArea(areaId);
}, 16);

// 3. 拖拽中
function handleNodeDrag(e: any) {
  const targetNode = e.target;
  // 注意：G6 5.0 中 e.target 是 Group，ID 就在上面
  if (!targetNode || (targetNode.id && targetNode.id.startsWith('edge-'))) return;

  const nodeId = targetNode.id;
  // 这里需要从 graph 里拿数据找 areaId
  const nodeData = graph.getNodeData(nodeId);
  const areaId = nodeData?.data?.area;

  if (areaId) {
    throttledUpdateArea(areaId as string);
  }
}

// 拖拽结束事件回调
function handleNodeDragEnd(e: any) {
  const targetNode = e.target;
  if (!targetNode || (targetNode.id && targetNode.id.startsWith('edge-'))) return;

  const nodeId = targetNode.id;
  const nodeData = graph.getNodeData(nodeId);
  const areaId = nodeData?.data?.area;

  if (areaId) {
    // 1. 取消节流队列，防止旧的覆盖新的
    throttledUpdateArea.cancel();

    // 2. 使用 setTimeout(0) 将其推到下一个事件循环
    // 这就是“模拟第二次拖拽”的效果：等待 G6 内部把拖拽后的最终坐标写入数据模型后，
    // 我们立即执行一次计算，把区域框校准到最新位置。
    setTimeout(() => {
      execUpdateArea(areaId as string);
    }, 100);
  }
}

// ------------------ 节点点击逻辑 (保持原样) ------------------
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

    // 事件监听
    window.addEventListener('resize', handleResize);
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

<style lang="postcss" scoped>
#nodemgr-g6-container {
  width: 100%;
  position: relative;
  overflow: hidden;
}
.favorited-item {
  &:hover {
    .nc-not-favorited {
      display: inline;
    }
  }
}
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
::-webkit-scrollbar-thumb {
  border-radius: 7px;
  border: 3px solid transparent;
  -webkit-box-shadow: inset 0 0 8px 8px #c4c6cc;
  box-shadow: inset 0 0 8px 8px #c4c6cc;
}
</style>
<style>
.force-tooltip-z-index {
  z-index: 99999 !important;
}
</style>
