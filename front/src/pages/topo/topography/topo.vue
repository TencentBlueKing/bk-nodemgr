<template>
  <Loading
    mode="spin"
    theme="primary"
    :loading="isLoading"
  >
    <div class="min-h-[calc(100vh_-_104px)]">
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
        <Select.Option
          v-for="(item, index) in netWorkAreaList"
          :id="item.bk_networkarea_id"
          :key="index"
          :name="item.bk_networkarea_name">

        </Select.Option>
      </Select>

      <div id="nodemgr-g6-container"></div>

      <!-- 自定义工具栏 -->
      <CustomToolbar @trigger-tools="handleClickTool"></CustomToolbar>
    </div>
  </Loading>
</template>

<script setup lang="ts">
import { Loading, Select } from 'bkui-vue';
import { filter, throttle } from 'lodash';
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { EdgeData, GraphData, NodeData } from '@antv/g6';
import { EdgeEvent, ExtensionCategory, Graph, register } from '@antv/g6';

import { NodeType } from './graph-plugin/config';
import CustomToolbar from './graph-plugin/custom-toolbar.vue';
import NetworkAreaNode from './graph-plugin/net-work-area-node';
import NetWorkUnitNode from './graph-plugin/net-work-unit-node';
import ResourceLayout from './graph-plugin/resource-layout';

import useMinLengthRef from '@/composables/use-min-length-ref';
import { useTopoStore } from '@/stores/topo';

const { t } = useI18n();
const {
  handleFetchTopoWorkareaList,
  handleFetchTopoWorkGraphNode,
  handleFetchTopoWorkGraphInfo,
} = useTopoStore();
const topoStore = useTopoStore();
const regionList = useMinLengthRef(['all'] as Array<string | number>, t('topoManager.topo.select.tips', { count: 1 }));
const netWorkAreaList = ref<Partial<NetworkArea>[]>([]);

function handleSelectChange() {};

let graph: Graph;
const graphData: GraphData = reactive({
  nodes: [],
  edges: [],
});
let allAreaNodes: NodeData[] = [];
let allUnitNodes: NodeData[] = [];
let allUnitNodesInfo = [];
let allLinkEdges: EdgeData[] = [];

const isLoading = ref(false);
const workAreaPrefix = 'workArea-';
const workUnitPrefix = 'workUnit-';
const workUnitLinkPrefix = 'link-';

const NavigatorHeight = 104;
const graphHeight = window.innerHeight - NavigatorHeight;

// graph参数配置，初始化graph
function handleInitTopo() {
  if (graph) return;
  graph = new Graph({
    container: 'nodemgr-g6-container',
    zoom: 0.95,
    zoomRange: [0.2, 1.5],
    autoFit: 'center',
    animation: false,
    autoResize: true,
    height: graphHeight,
    edge: {
      type: 'polyline',
      style: {
        startArrow: true,
        startArrowType: 'circle',
        startArrowFill: '#fff',
        endArrow: true,
        stroke: '#C4C6CC',
        radius: 100,
        zIndex: 1,
      },
    },
    data: graphData,
    behaviors: [
      {
        type: 'scroll-canvas',
        key: 'scroll-canvas',
        direction: 'y',
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
      type: 'custom-layout',
    },
    background: '#FAFBFD',
  });
  graph.render();
}

function reRender() {
  if (!graph) return;
  graph.setData(graphData);
  graph.render();
}

function handleClickTool(code: string, value: number) {
  switch (code) {
    case 'center':
      graph?.fitCenter?.();
      break;
    case 'mapSize':
      graph?.zoomTo?.(value);
      break;
  };
}

function handleHoverEdge(evt) {
  const { target } = evt;
  graph.setElementState(target.id, 'highlight');
}

function handleLeaveEdge(evt) {
  const { target } = evt;
  graph.setElementState(target.id, '');
}

// graph插件注册
function handleRegistryCategory() {
  register(ExtensionCategory.NODE, NodeType.NET_WORK_UNIT, NetWorkUnitNode);
  register(ExtensionCategory.NODE, NodeType.NET_WORK_AREA, NetworkAreaNode);
  register(ExtensionCategory.LAYOUT, 'custom-layout', ResourceLayout);
}

const initGraphData = async () => {
  await Promise.all([
    handleFetchTopoWorkareaList(),
    handleFetchTopoWorkGraphNode(),
  ]);
  // 初始化select option
  netWorkAreaList.value = topoStore.allWorkareaList.map(item => ({
    bk_networkarea_id: item.bk_networkarea_id,
    bk_networkarea_name: item.bk_networkarea_name,
  }));

  // 初始化区域数据
  allAreaNodes = topoStore.allWorkareaList.map((item) => {
    const nodeId = workAreaPrefix + item.bk_networkarea_id;
    return {
      id: nodeId,
      data: {
        name: item.bk_networkarea_name,
      },
      type: NodeType.NET_WORK_AREA,
    };
  });

  // 初始化管控单元
  allUnitNodes = topoStore.allWorkGraphNodes.map((item) => {
    const nodeId = workUnitPrefix + item.bk_networkunit_id;
    return {
      id: nodeId,
      data: {
        name: item.bk_networkunit_name,
        area: workAreaPrefix + item.bk_networkarea_id,
        proxy: 0,
        agent: 0,
      },
      type: NodeType.NET_WORK_UNIT,
    };
  });

  // 初始化连接关系
  allLinkEdges = topoStore.allWorkGraphEdges.map((item, index) => ({
    id: workUnitLinkPrefix + index,
    source: workUnitPrefix + item.source_networkunit_id,
    target: workUnitPrefix + item.target_networkunit_id,
  }));

  // 初始化节点数据
  graphData.nodes = [...allAreaNodes, ...allUnitNodes];
  graphData.edges = allLinkEdges;
};

const initTopoCount = async () => {
  await handleFetchTopoWorkGraphInfo();
  allUnitNodesInfo = topoStore.allWorkGraphInfos;
  updateNodeProxyAgentInfo();
  reRender();
};

function updateNodeProxyAgentInfo() {
  // 创建一个查找表
  const lookup = new Map();
  // 填充查找表
  for (const unit of allUnitNodesInfo) {
    const unitId = `${workUnitPrefix}${unit.bk_networkunit_id}`;
    lookup.set(unitId, {
      proxy: unit.proxy,
      agent: unit.agent,
    });
  }
  // 补充allUnitNodes 信息
  for (const unit of allUnitNodes as NodeData[]) {
    const matchingUnit = lookup.get(unit.id);
    if (matchingUnit) {
      unit.data = {
        ...unit.data,
        proxy: matchingUnit.proxy,
        agent: matchingUnit.agent,
      };
    }
  }
}

// 根据workarea_id筛选数据
function filterAreaNodes() {
  const newValSet = new Set(regionList.value);
  const curAreaNodes = allAreaNodes.filter((item) => {
    const targetId = Number(item.id.replace(workAreaPrefix, ''));
    const result = newValSet.has(targetId);
    return result;
  });
  const curUnitNodes = allUnitNodes.filter((item) => {
    const targetId = Number(item.data.area.replace(workAreaPrefix, ''));
    const result = newValSet.has(targetId);
    return result;
  });
  const workUnitIds = new Set(curUnitNodes.map(item => item.id));
  const curEdges = allLinkEdges.filter(item => workUnitIds.has(item.source) && workUnitIds.has(item.target));

  graphData.nodes = [...curAreaNodes, ...curUnitNodes];
  graphData.edges = curEdges;
  reRender();
}

watch(regionList, throttle(async (newVal) => {
  isLoading.value = true;
  graphData.edges = []; // 清空edges 防止filter nodes后，edge找不到对应的node
  try {
    if (newVal[0] === 'all') {
      // 复用allNodes
      graphData.nodes = [...allAreaNodes, ...allUnitNodes];
      // 复用allEdges
      graphData.edges = allLinkEdges;
      reRender();
    } else {
      filterAreaNodes();
    }
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
}, 100, { leading: true, trailing: true }));

onMounted(async () => {
  isLoading.value = true;
  try {
    // Graph插件注册
    handleRegistryCategory();
    // 初始化graphData
    await initGraphData();
    // 初始化graph配置并渲染
    handleInitTopo();
    // 为edges增加hover效果
    graph.on(EdgeEvent.POINTER_OVER, handleHoverEdge);
    graph.on(EdgeEvent.POINTER_OUT, handleLeaveEdge);
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
  // 获取管控单元详细信息并重新渲染
  initTopoCount();
});

onUnmounted(() => {
  graph?.off();
  graph?.destroy();
});
</script>
