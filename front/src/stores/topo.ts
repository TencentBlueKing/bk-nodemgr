import { defineStore } from 'pinia';
import { ref } from 'vue';

import type { TopoNetworkAreaListReq } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';

export const useTopoStore = defineStore('topo', () => {
  const allWorkareaList = ref<NetworkArea[]>([]);
  const workareaTotalCount = ref(0);
  const allWorkGraphNodes = ref<NetworkUnitGraph[]>([]);
  const allWorkGraphEdges = ref<LinkGraph[]>([]);
  const allWorkGraphInfos = ref<{
    bk_networkunit_id: number,
    proxy: number,
    agent: number
  }[]>([]);

  const handleFetchTopoWorkareaList = async () => {
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0, // limit为0即获取所有
      },
    };

    const result = await TopoService.NetworkAreaList(params);
    allWorkareaList.value = result.items;
    workareaTotalCount.value = result.total;
  };

  const handleFetchTopoWorkGraphNode = async () => {
    const result = await TopoService.GraphGet({
      bk_networkarea_id: null, // null即为获取所有
    });
    allWorkGraphNodes.value = result.networkunit;
    allWorkGraphEdges.value = result.links;
  };

  const handleFetchTopoWorkGraphInfo = async () => {
    const result = await TopoService.GraphNodeCount({
      bk_networkunit_id: null,
    });
    allWorkGraphInfos.value = result.networkunits;
  };

  return {
    allWorkareaList,
    workareaTotalCount,
    allWorkGraphNodes,
    allWorkGraphEdges,
    allWorkGraphInfos,
    handleFetchTopoWorkareaList,
    handleFetchTopoWorkGraphNode,
    handleFetchTopoWorkGraphInfo,
  };
});
