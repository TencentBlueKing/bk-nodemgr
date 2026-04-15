<template>
  <Loading mode="spin" theme="primary" :loading="isLoading">
    <!-- 无权限页面 -->
    <div v-if="noAreaPermission" class="forbidden-page">
      <img class="forbidden-img" src="/images/403.png" alt="403">
      <div class="forbidden-title">{{ $t('components.permission.noPermission') }}</div>
      <Button theme="primary" @click="handleApplyAreaPermission">
        {{ $t('components.permission.apply') }}
      </Button>
    </div>
    <!-- 有权限：正常 topo 图 -->
    <div v-else :class="[mainStore.noticeShow ? 'min-h-[calc(100vh-144px)]' : 'min-h-[calc(100vh-104px)]', 'relative']">
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
        <Select.Group v-if="defaultNetWorkarea" :label="$t('topoManager.topo.select.default')">
          <Select.Option
            :key="defaultNetWorkarea.bk_networkarea_id"
            :id="defaultNetWorkarea.bk_networkarea_id"
            :name="defaultNetWorkarea.bk_networkarea_name"
          >
            <div class="w-[180px] flex">
              <Button
                text
                class="mr-[8px] w-[18px] favorited-item">
                <i class="nodeman-icon nc-collect text-[#C4C6CC] text-[18px]">
                </i>
              </Button>
              <span>
                {{ `[${defaultNetWorkarea.bk_networkarea_id}] ${defaultNetWorkarea.bk_networkarea_name}` }}
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
            v-bk-tooltips="{
              content: isAreaAuthorized(item.bk_networkarea_id)
                ? `[${item.bk_networkarea_id}] ${item.bk_networkarea_name}`
                : t('components.permission.noPermission'),
              disabled: isAreaAuthorized(item.bk_networkarea_id) && !textOverflowMap[item.bk_networkarea_id],
              boundary: 'parent',
              placement: 'right',
              offset: 10
            }"
          >
            <div
              class="w-full flex items-center topo-area-option overflow-hidden"
              :class="{ 'unauthorized-area-row': !isAreaAuthorized(item.bk_networkarea_id) }"
              @click="handleAreaOptionClick($event, item.bk_networkarea_id)"
              @mouseenter="handleAreaOptionMouseEnter($event, item.bk_networkarea_id)"
              @mousemove="handleAreaOptionMouseMove($event, item.bk_networkarea_id)"
              @mouseleave="handleAreaOptionMouseLeave()"
            >
              <Button
                text
                class="mr-[8px] w-[18px] shrink-0"
                @click.stop="handleCollect(item.bk_networkarea_id)">
                <i
                  class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
                  v-if="collectList.includes(item.bk_networkarea_id)">
                </i>
                <i
                  class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
                  v-else>
                </i>
              </Button>
              <div
                class="truncate"
                @mouseenter="handleTextMouseenter($event, item.bk_networkarea_id)">
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

      <!-- =========== 接入点详情：已改为使用 tableTooltip（纯 DOM），无需 Popover =========== -->
    </div>
  </Loading>
</template>

<script setup lang="ts">
import { Button, Loading, OverflowTitle, Popover, Select } from 'bkui-vue';
import { throttle } from 'lodash';
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
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

import AccessPointNode from './graph-plugin/access-point-node';
import { NodeStatus, NodeType, UnitType } from './graph-plugin/config';
import CustomToolbar from './graph-plugin/custom-toolbar.vue';
import { textTooltip } from './graph-plugin/text-tooltip';
import { edgeTooltip } from './graph-plugin/edge-tooltip';

import HorizontalHierarchyLayout from './graph-plugin/HorizontalHierarchyLayout';
import CustomEdge from './graph-plugin/customEdge';
import NetworkAreaNode from './graph-plugin/net-work-area-node';
import NetWorkUnitNode from './graph-plugin/net-work-unit-node';

import useMinLengthRef from '@/composables/use-min-length-ref';
import useAuthLock from '@/composables/use-auth-lock';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';
import { usePermissionStore } from '@/stores/permission';
import { useTopoStore } from '@/stores/topo';
import { getModuleAuthorizedItems } from '@/constants/auth';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const {
  handleFetchTopoWorkareaList,
  handleFetchAllWorkUnit,
} = useTopoStore();
const topoStore = useTopoStore();
const mainStore = useMainStore();
const workareaStore = useWorkareaStore();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const router = useRouter();

// ===== 权限检查 =====
const noAreaPermission = ref(false);

/** 检查是否有任何区域的查看权限 */
async function checkAreaPermission() {
  // 确保 topoManager 模块的 authorized items 已加载
  const topoItems = getModuleAuthorizedItems('topoManager');
  await authStore.fetchAuthorized(topoItems, 'topoManager');

  // 如果 fetchAuthorized 被锁跳过了（其他地方正在请求），等待加载完成
  if (authStore.authorizedLoading) {
    await new Promise<void>((resolve) => {
      const unwatch = watch(() => authStore.authorizedLoading, (loading) => {
        if (!loading) {
          unwatch();
          resolve();
        }
      });
    });
  }

  const hasAny = authStore.hasAuthorizedResource('networkarea_view');
  noAreaPermission.value = !hasAny;
}

/** 申请区域查看权限 */
async function handleApplyAreaPermission(areaId?: number) {
  const authItems = [
    { id: 'networkarea_view', action: 'networkarea_view', resourceType: 'networkarea', routes: [] },
  ];
  await authStore.batchVerify(authItems, undefined, areaId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
}

/** 申请单元查看权限 */
async function handleApplyUnitPermission(unitId: number) {
  const authItems = [
    { id: 'networkunit_view', action: 'networkunit_view', resourceType: 'networkunit', routes: [] },
  ];
  await authStore.batchVerify(authItems, undefined, unitId);
  const detail = authStore.permissionDetail;
  if (detail) {
    permissionStore.showDialog(detail);
  }
}

// ===== 管控区域选择器权限控制（与 areaSelector 一致）=====
const {
  handleMouseEnter: areaAuthMouseEnter,
  handleMouseMove: areaAuthMouseMove,
  handleMouseLeave: areaAuthMouseLeave,
} = useAuthLock(
  'networkarea_view',
  () => undefined,
);

/** 判断某个管控区域是否有权限 */
function isAreaAuthorized(areaId: number): boolean {
  if (!authStore.authorizedLoaded) return true;
  return authStore.hasAuthorizedResource('networkarea_view', areaId);
}

const handleAreaOptionMouseEnter = (e: MouseEvent, areaId: number) => {
  areaAuthMouseEnter(e, isAreaAuthorized(areaId));
};
const handleAreaOptionMouseMove = (e: MouseEvent, areaId: number) => {
  areaAuthMouseMove(e, isAreaAuthorized(areaId));
};
const handleAreaOptionMouseLeave = () => {
  areaAuthMouseLeave();
};

/** 点击无权限区域 → 阻止选中 + 触发权限申请 */
const handleAreaOptionClick = (e: MouseEvent, areaId: number) => {
  if (!isAreaAuthorized(areaId)) {
    e.stopPropagation();
    e.preventDefault();
    handleApplyAreaPermission(areaId);
  }
};

// 选择器逻辑
const regionList = useMinLengthRef(
  ['all'] as Array<string | number>,
  t('topoManager.topo.select.tips', { count: 1 }),
);
const netWorkAreaList = ref<Partial<any>[]>([]);
const defaultNetWorkarea = ref<Partial<any>>();

// 计算属性：排序后的其他区域列表（按照优先级顺序：默认区域第一，有权限优先级第二，勾选优先级第三，收藏优先级第四，ID从大到小第五）
const sortedNetWorkAreaList = computed(() => [...netWorkAreaList.value].sort((a, b) => {
  // bk_networkarea_id为0的始终排在最前面
  if (a.bk_networkarea_id === 0) return -1;
  if (b.bk_networkarea_id === 0) return 1;

  // 有权限的区域排在前面
  const aHasAuth = authStore.hasAuthorizedResource('networkarea_view', a.bk_networkarea_id);
  const bHasAuth = authStore.hasAuthorizedResource('networkarea_view', b.bk_networkarea_id);
  if (aHasAuth && !bHasAuth) return -1;
  if (!aHasAuth && bHasAuth) return 1;

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

// 监听选择状态变化，持久化到 localStorage
watch(regionList, (newVal) => {
  localStorage.setItem('selected_workarea', JSON.stringify(newVal));
}, { deep: true });

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
    edge: {
      type: 'custom-edge',
      style: {
        stroke: '#C4C6CC', // 默认线条颜色
        lineWidth: 1, // 修改：从2改为1，与布局算法保持一致
      },
    },
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
          // if (!['node', 'edge'].includes(event.targetType) || event.target.id.includes('workArea')) return false;

          // const nodeId = event.target.id;

          // // 单元节点：通过计算鼠标是否在顶部 36px 范围内
          // if (String(nodeId).includes('workUnit')) {
          //   const { y: mouseY } = event.canvas;
          //   const bounds = event.target.getRenderBounds();
          //   const nodeTopY = bounds.min[1];
          //   const headerHeight = 36;

          //   // 纯数学比对，瞬间返回 true/false，不需要 await
          //   if (mouseY >= nodeTopY && mouseY <= nodeTopY + headerHeight + 2) {
          //     return true;
          //   }
          //   return false;
          // }
          // return true;
          return false;
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

function handleHoverEdge(evt: any) {
  const { target, client } = evt;
  graph.setElementState(target?.id, 'highlight');

  // hover 边时使用手型光标
  const canvas = graph.getCanvas()?.getContextService?.()?.getDomElement?.();
  if (canvas) canvas.style.cursor = 'pointer';

  // 获取边数据，构建 tooltip 内容
  if (target?.id && graph) {
    const edgeData = graph.getEdgeData(target.id);
    const linkTypes = (edgeData?.data as any)?.linkTypes as string[] | undefined;
    const unitName = (edgeData?.data as any)?.unitName;
    const apName = (edgeData?.data as any)?.apName;
    const apUnitName = (edgeData?.data as any)?.apUnitName;

    if (linkTypes && linkTypes.length > 0 && linkTypes[0] !== 'downstream') {
      // Unit→AP 的上游边
      edgeTooltip.show({
        linkTypes,
        sourceName: unitName,
        targetName: `${apUnitName}/${apName}`,
      }, client?.x ?? 0, client?.y ?? 0);
    } else if (linkTypes && linkTypes[0] === 'downstream') {
      // AP→Unit 的下游边
      edgeTooltip.show({
        linkTypes,
        sourceName: `${apUnitName}/${apName}`,
        targetName: unitName,
      }, client?.x ?? 0, client?.y ?? 0);
    }
  }
}

function handleLeaveEdge(evt: any) {
  const { target } = evt;
  graph.setElementState(target?.id, '');
  edgeTooltip.hide();

  // 恢复默认光标
  const canvas = graph.getCanvas()?.getContextService?.()?.getDomElement?.();
  if (canvas) canvas.style.cursor = 'default';
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

  switch (action) {
    case 'detail':
      // 这里的 nodeData.name 就是 "default-test-1"
      break;

    case 'edit':
      // 编辑单元
      // 如果你需要 ID，可以从 currentNodeId 解析，或者看看 nodeData 里有没有存 ID
      // 你的数据里好像只有 bk_networkunit_id 在 links 里或者需要从 nodeId 解析
      break;

    case 'delete':
      // 你的数据里有 "area": "workArea-0"

      // 执行删除逻辑...
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
      const isDirect = node.data.is_direct as boolean;
      width = 240;
      height = isDirect ? 162 : 214;
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

// 控制节点拖拽是否启用（与 drag-element behavior 的 enable 保持一致）
const isDragElementEnabled = false;

// 3. 拖拽中
function handleNodeDrag(e: any) {
  // drag-element 禁用时，G6 仍会冒泡 node:drag 事件，但不应执行任何副作用
  if (!isDragElementEnabled) return;

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
  // drag-element 禁用时，G6 仍会冒泡 node:dragend 事件，但不应执行任何副作用
  if (!isDragElementEnabled) return;

  const targetNode = e.target;
  if (!targetNode || (targetNode.id && targetNode.id.startsWith('edge-'))) return;

  const nodeId = targetNode.id;
  const nodeData = graph.getNodeData(nodeId);
  const areaId = nodeData?.data?.area;

  if (areaId) {
    // 1. 取消节流队列，防止旧的覆盖新的
    throttledUpdateArea.cancel();

    // 2. 使用 setTimeout(0) 将其推到下一个事件循环
    // 这就是"模拟第二次拖拽"的效果：等待 G6 内部把拖拽后的最终坐标写入数据模型后，
    // 我们立即执行一次计算，把区域框校准到最新位置。
    setTimeout(() => {
      execUpdateArea(areaId as string);
    }, 100);
  }
}

// ------------------ 节点点击逻辑 (保持原样) ------------------
async function handleNodeClick(evt: any) {
  const { target, path, canvas } = evt;

  // 1. 获取点击图形的 className (确认是否点了菜单热区)
  const clickedShape = path?.[0];
  const className = clickedShape?.config?.className;

  // 2. 获取节点 ID (workUnit-xxx 或 accessPoint-xxx)
  let nodeId = target.id;
  if (!nodeId && path) {
    const nodeObj = path.find((p: any) => p.id && (String(p.id).includes(workUnitPrefix) || String(p.id).includes('accessPoint-')));
    nodeId = nodeObj?.id;
  }

  // ==================== 无权限"查看"按钮点击逻辑 ====================
  if (className === 'auth-view-btn' || className === 'auth-view-btn-bg') {
    if (nodeId && nodeId.includes(workUnitPrefix)) {
      const unitId = Number(nodeId.replace(workUnitPrefix, ''));
      await handleApplyUnitPermission(unitId);
    } else if (nodeId && nodeId.includes('accessPoint-')) {
      // 接入点的权限取决于其所属单元
      const nodeData = graph.getNodeData(nodeId);
      const unitId = nodeData?.data?.bk_networkunit_id as number | undefined;
      if (unitId != null) {
        await handleApplyUnitPermission(unitId);
      }
    }
    return;
  }

  // ==================== 点击无权限单元/接入点 → 触发权限申请 ====================
  if (nodeId) {
    const nodeData = graph.getNodeData(nodeId);
    if (nodeData?.data?.has_unit_auth === false) {
      const unitId = nodeId.includes(workUnitPrefix)
        ? Number(nodeId.replace(workUnitPrefix, ''))
        : (nodeData?.data?.bk_networkunit_id as number | undefined);
      if (unitId != null) {
        await handleApplyUnitPermission(unitId);
      }
      return;
    }
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
  // 只有点击跳转图标时才允许跳转
  if (nodeId && nodeId.includes(workUnitPrefix)) {
    const nodeData = graph.getNodeData(nodeId);
    if (!nodeData) return;

    // 解析 ID
    const workareaId = Number(String(nodeData.data?.area).replace(workAreaPrefix, ''));
    const workUnitId = Number(nodeId.replace(workUnitPrefix, ''));

    // 跳转到 Agent 页面（点击 Agent 跳转图标）
    if (className === 'linkIcon-agent') {
      textTooltip.hide(true); // 立即隐藏提示气泡
      router.push({
        name: 'agent',
        query: {
          bk_networkarea_id: workareaId,
          bk_networkunit_id: workUnitId,
          bk_networkunit_name: nodeData.data?.name as string,
        },
      });
      return;
    }

    // 跳转到单元详情页（点击单元跳转图标）
    if (className === 'linkIcon-unit-direct' || className === 'linkIcon-unit-proxy') {
      textTooltip.hide(true); // 立即隐藏提示气泡
      router.push({
        name: 'workareaDetail',
        params: {
          workarea: workareaId,
          workUnit: workUnitId,
        },
      });
      return;
    }

    // 如果不是点击跳转图标，不做任何跳转
  }
}
// 连接关系生成函数
const generateEdgesFromUnitData = () => {
  const edges: Array<{
    id: string;
    target: string;
    source: string;
    data?: Record<string, any>;
  }> = [];

  // 构建接入点映射
  const accessPointMap = new Map();
  topoStore.accessPointData.forEach((ap) => {
    accessPointMap.set(ap.bk_accesspoint_id, ap);
  });

  // 构建单元名映射（unitId -> unitName），用于查询接入点所属单元的名字
  const unitNameMap = new Map<number, string>();
  topoStore.workUnitByArea.forEach((u) => {
    unitNameMap.set(u.bk_networkunit_id, u.bk_networkunit_name);
  });

  topoStore.workUnitByArea.forEach((unit) => {
    const unitId = `workUnit-${unit.bk_networkunit_id}`;

    if (!unit.is_direct) {
      // 非直连单元：连接到上游接入点
      // 检查 cluster/file/data 三个 link 的 accesspoint_id 是否完全一样
      const clusterApId = unit.links?.cluster?.accesspoint_id;
      const fileApId = unit.links?.file?.accesspoint_id;
      const dataApId = unit.links?.data?.accesspoint_id;

      if (clusterApId != null || fileApId != null || dataApId != null) {
        // 收集所有不同的 accesspoint_id 及其对应的 link 类型
        const apTypeMap = new Map<number, string[]>(); // apId -> [linkTypes]
        const linkEntries: Array<[string, number | undefined | null]> = [
          ['cluster', clusterApId],
          ['file', fileApId],
          ['data', dataApId],
        ];
        linkEntries.forEach(([type, apId]) => {
          if (apId != null) {
            if (!apTypeMap.has(apId)) {
              apTypeMap.set(apId, []);
            }
            apTypeMap.get(apId)!.push(type);
          }
        });

        if (apTypeMap.size === 1) {
          // 所有 link 指向同一个接入点 → 生成1条边
          const [apIdNum, types] = [...apTypeMap.entries()][0];
          const apId = `accessPoint-${apIdNum}`;
          const apInfo = accessPointMap.get(apIdNum);
          const apUnitName = apInfo?.bk_networkunit_id != null
            ? unitNameMap.get(apInfo.bk_networkunit_id) || ''
            : '';
          edges.push({
            id: `edge-${apId}-${unitId}`,
            target: apId,
            source: unitId,
            data: {
              linkTypes: types, // e.g. ['cluster', 'file', 'data']
              unitName: unit.bk_networkunit_name,
              unitNum: unit.bk_networkunit_id,
              apName: apInfo?.name || `ap_${apIdNum}`,
              apUnitName,
              apNum: apIdNum,
              subIndex: 0,
              totalSubEdges: 1,
            },
          });
        } else {
          // 不同的 link 指向不同接入点 → 生成多条边
          let subIndex = 0;
          const totalSubEdges = apTypeMap.size;
          apTypeMap.forEach((types, apIdNum) => {
            const apId = `accessPoint-${apIdNum}`;
            const apInfo = accessPointMap.get(apIdNum);
            const apUnitName = apInfo?.bk_networkunit_id != null
              ? unitNameMap.get(apInfo.bk_networkunit_id) || ''
              : '';
            edges.push({
              id: `edge-${apId}-${unitId}-${types.join('_')}`,
              target: apId,
              source: unitId,
              data: {
                linkTypes: types, // e.g. ['cluster'] or ['file', 'data']
                unitName: unit.bk_networkunit_name,
                unitNum: unit.bk_networkunit_id,
                apName: apInfo?.name || `ap_${apIdNum}`,
                apUnitName,
                apNum: apIdNum,
                subIndex,
                totalSubEdges,
              },
            });
            subIndex++;
          });
        }
      }
    }

    // 单元提供接入点：单元→接入点
    if (unit.accesspoints && unit.accesspoints.length > 0) {
      unit.accesspoints.forEach((apInfo) => {
        const apId = `accessPoint-${apInfo.accesspoint_id}`;
        const ap = accessPointMap.get(apInfo.accesspoint_id);
        edges.push({
          id: `edge-${unitId}-${apId}`,
          target: unitId,
          source: apId,
          data: {
            linkTypes: ['downstream'], // AP→Unit 下游连接
            unitName: unit.bk_networkunit_name,
            unitNum: unit.bk_networkunit_id,
            apName: ap?.name || `ap_${apInfo.accesspoint_id}`,
            apUnitName: unit.bk_networkunit_name, // AP所属单元就是当前单元
            apNum: apInfo.accesspoint_id,
            subIndex: 0,
            totalSubEdges: 1,
          },
        });
      });
    }
  });

  return edges;
};

// 初始化区域数据
const initAreaData = async () => {
  // 先获取区域列表（用于下拉）
  await handleFetchTopoWorkareaList().catch(() => {});

  // 获取有权限的单元 ID 列表，传给 handleFetchAllWorkUnit 以区分接入点接口
  // null 表示全部有权限（isAny），[] 表示全部无权限
  const authorizedIds = authStore.authorizedLoaded
    ? authStore.getAuthorizedResourceIds('networkunit_view')
    : [];

  await handleFetchAllWorkUnit([], authorizedIds).catch(() => {});

  const defaultArea = topoStore.allWorkareaList.find(item => item.bk_networkarea_id === 0);
  defaultNetWorkarea.value = defaultArea
    ? { bk_networkarea_id: defaultArea.bk_networkarea_id, bk_networkarea_name: defaultArea.bk_networkarea_name }
    : undefined;

  const defaultAreaId = defaultArea?.bk_networkarea_id;
  netWorkAreaList.value = topoStore.allWorkareaList
    .filter(item => item.bk_networkarea_id !== defaultAreaId)
    .map(item => ({
      bk_networkarea_id: item.bk_networkarea_id,
      bk_networkarea_name: item.bk_networkarea_name,
    }));

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
        width: 260,
        height: 300,
      },
      type: NodeType.NET_WORK_AREA,
    }));

  const filteredUnitNodes = topoStore.workUnitByArea
    .filter(item => targetAreaIds.includes(item.bk_networkarea_id))
    .map((item) => {
      const nodeId = `workUnit-${item.bk_networkunit_id}`;
      // 检查该单元的查看权限
      const hasUnitAuth = authStore.hasAuthorizedResource('networkunit_view', item.bk_networkunit_id);
      return {
        id: nodeId,
        data: {
          name: item.bk_networkunit_name,
          unitType: item.is_direct ? 'direct' : 'indirect',
          area: workAreaPrefix + item.bk_networkarea_id,
          is_direct: item.is_direct,
          accesspoints: item.accesspoints || [],
          links: item.links || {},
          running_proxy: item.running_proxy || 0,
          total_proxy: item.total_proxy || 0,
          running_agent: item.running_agent || 0,
          total_agent: item.total_agent || 0,
          cycle_times: item.cycle_times || [],
          is_healthy: item.is_healthy,
          has_unit_auth: hasUnitAuth,
          bk_networkunit_id: item.bk_networkunit_id,
        },
        type: NodeType.NET_WORK_UNIT,
      };
    });

  const filteredAccessPointNodes = topoStore.accessPointData
    .filter(item => targetAreaIds.includes(item.bk_networkarea_id))
    .map((item) => {
      const nodeId = `accessPoint-${item.bk_accesspoint_id}`;
      // 接入点的权限取决于其所属单元
      const hasUnitAuth = authStore.hasAuthorizedResource('networkunit_view', item.bk_networkunit_id);
      return {
        id: nodeId,
        data: {
          name: hasUnitAuth ? item.name : '***',
          downstreamUnits: item.downstreamUnits || 0,
          area: workAreaPrefix + item.bk_networkarea_id,
          bk_networkunit_id: item.bk_networkunit_id,
          type: item.type,
          // 无权限时脱敏 endpoints
          endpoints: hasUnitAuth ? item.endpoints : {},
          endpointsData: hasUnitAuth
            ? [{ name: item.name, endpoints: item.endpoints }]
            : [{ name: '***', endpoints: {} }],
          has_unit_auth: hasUnitAuth,
          bk_accesspoint_id: item.bk_accesspoint_id,
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
  register(ExtensionCategory.EDGE, 'custom-edge', CustomEdge);
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
  // 权限检查与数据加载并行，authorized 失败不阻塞拓扑图展示
  await checkAreaPermission().catch(() => {});

  // 恢复用户的选择状态（优先从 localStorage 读取）
  const savedSelection = localStorage.getItem('selected_workarea');
  if (savedSelection) {
    try {
      regionList.value = JSON.parse(savedSelection);
    } catch (err) {
      console.error('Failed to parse saved selection:', err);
      // 如果解析失败，默认选中 ID=0 的默认区域
      regionList.value = [0];
    }
  } else {
    // 如果没有保存的选择状态，默认选中 ID=0 的默认区域
    regionList.value = [0];
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
.forbidden-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}
.forbidden-img {
  width: 200px;
  height: auto;
}
.forbidden-title {
  margin-top: 20px;
  font-size: 22px;
  font-weight: 400;
  color: #63656e;
}
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
.topo-area-option {
  min-height: 32px;
  padding: 0 12px;
  box-sizing: border-box;
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
.unauthorized-area-row {
  color: #c4c6cc !important;
}
</style>
