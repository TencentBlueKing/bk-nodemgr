<template>
  <div class="p-[24px]">
    <Select
      class="w-[240px]"
      v-model="regionList"
      multiple-mode="tag"
      all-option-id="all"
      collapse-tags
      filterable
      multiple
      show-all
      @change="handleSelectChange"
    >
      <Select.Option
        v-for="(item, index) in regionSourceList"
        :id="item.bk_networkarea_id"
        :key="index"
        :name="item.bk_networkarea_name">

      </Select.Option>
    </Select>

    <div id="nodemgr-g6-container" class="mt-[24px]"></div>

    <!-- 自定义工具栏 -->
    <CustomToolbar @trigger-tools="handleClickTool"></CustomToolbar>
  </div>
</template>

<script setup lang="ts">
import { Select } from 'bkui-vue';
import { onMounted, reactive, ref } from 'vue';

import { ExtensionCategory, Graph, register } from '@antv/g6';

import CustomToolbar from './graph-plugin/custom-toolbar.vue';
import ResourceNode from './graph-plugin/resource-node';
import ResourceCombo from './graph-plugin/resource-combo';
import { TopoService } from '@/api/modules/topo';
import { areaList, edgeList, nodeList } from './graph-plugin/mockData';
import ResourceLayout from './graph-plugin/resource-layout';

const regionList = ref(['all']);
const regionSourceList = ref([]);

function handleSelectChange() {};

let graph: Graph;
const graphData = reactive({
  nodes: nodeList,
  edges: edgeList,
  combos: [
    { id: 'combo1', data: { label: '云电脑'} },
    { id: 'combo2', data: { label: 'test'} },
    { id: 'combo3', data: { label: '默认区域'} },
    { id: 'combo4', data: { label: '上海公有云'} },
    { id: 'combo5', data: { label: 'DBM 专用云区域'} },
  ],
});

function handleInitTopo() {
  if (graph) return;
  graph = new Graph({
    container: 'nodemgr-g6-container',
    zoom: 0.95,
    zoomRange: [0.2, 1.5],
    autoFit: 'center',
    animation: false,
    autoResize: true,
    height: window.innerHeight - 210,
    node: {
      type: 'custom-resource',
    },
    edge: {
      type: 'polyline',
      style: {
        startArrow: true,
        startArrowType: 'circle',
        startArrowFill: '#fff',
        endArrow: true,
        stroke: '#C4C6CC',
        radius: 100,
      }
    },
    combo: {
      type: 'custom-combo',
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
    }
  });
  graph.render();
}

function reRender() {
  if (!graph) return;
  graph.setData(graphData)
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

// 注册节点和插件
function handleRegistryCategory() {
  register(ExtensionCategory.NODE, 'custom-resource', ResourceNode);
  register(ExtensionCategory.COMBO, 'custom-combo', ResourceCombo);
  register(ExtensionCategory.LAYOUT, 'custom-layout', ResourceLayout);
}

async function fetchNodeData() {
  const res = await TopoService.HostList();
}

// 
async function fetchComboData() {
  // const res = await TopoService.NetworkAreaList();
  regionSourceList.value = areaList;

  graphData.combos = areaList.map(item => ({
    id: (item.bk_networkarea_id).toString(),
    data: { label: item.bk_networkarea_name }
  }));
}

onMounted(() => {
  // fetchComboData();

  handleRegistryCategory();
  handleInitTopo();

});
</script>
