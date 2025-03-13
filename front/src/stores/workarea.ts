import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

import type { TopoNetworkAreaCreateReq, TopoNetworkAreaListReq } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';

export const useWorkareaStore = defineStore('workarea', () => {
  const workareaList = ref<NetworkArea[]>([]);
  const allWorkareaList = ref<Map<number, NetworkArea>>(new Map());
  const allWorkUnitList = ref<Map<number, NetworkUnit[]>>(new Map());
  const allAccessPointList = ref<Map<number, AccessPoint[]>>(new Map());
  // const all
  const loading = ref(false);
  const pagination = reactive({ count: 0, limit: 50, current: 1 });

  interface IncludeConditions {
    bk_networkarea_id: number[] // 管控区域ID
    bk_networkarea_name: string[] // 管控区域
    bk_cloud_vendor: number[] // 云服务商
  }

  // 可搜索字段
  const includeConditions = reactive<IncludeConditions>({
    bk_networkarea_id: [],
    bk_networkarea_name: [],
    bk_cloud_vendor: [],
  });

  // 获取管控区域列表
  const handleFetchWorkareaList = async () => {
    loading.value = true;
    const result = await TopoService.NetworkAreaList({
      page: {
        // todo
        // 接口文档示例 offset为0，了解下前端是否需要-1
        offset: pagination.current - 1,
        limit: pagination.limit,
      },
      onlyCount: false,
      includeConditions,
    })
      .catch(() => {})
      .finally(() => loading.value = false);
    workareaList.value = result?.items || [];
    pagination.count = result?.total || 0;
  };

  const handleDeleteWorkarea = async (bk_networkarea_id: number) => {
    const result =  await TopoService.NetworkAreaDelete({
      bk_networkarea_id,
    }).catch(() => {});
    return result;
  };

  const handleCreateWorkarea = async (params: TopoNetworkAreaCreateReq) => {
    const result = await TopoService.NetworkAreaCreate(params).catch(() => {});
    return result;
  };

  // 获取所有管控区域数据，用于复制
  const handleGetAllWorkareaList = async () => {
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0,
      },
    };
    const result = await TopoService.NetworkAreaList(params).catch(() => {});
    const list = result?.items || [];
    return list || [];
  };

  // 将所有管控区域存入Map
  const handleFetchAllWorkarea = async () => {
    allWorkareaList.value.clear();
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0,
      },
    };
    const result = await TopoService.NetworkAreaList(params).catch(() => {});
    const list = result?.items || [];
    for (const area of list) {
      allWorkareaList.value.set(area.bk_networkarea_id, area);
    }
  };

  const handleFetchAllWorkUnit = async () => {
    allWorkUnitList.value.clear();
    allAccessPointList.value.clear();
    const result = await TopoService.NetworkUnitList({
      bk_networkarea_id: null, // null即为获取所有
    });
    const list = result?.items || [];
    const accessPointList = [];
    for (const unit of list) {
      accessPointList.push(...unit.accesspoints);

      allAccessPointList.value.set(unit.bk_networkunit_id, unit.accesspoints);

      const units = allWorkUnitList.value.get(unit.bk_networkarea_id) || [];
      units.push(unit);
      allWorkUnitList.value.set(unit.bk_networkarea_id, units);
    }
  };

  return {
    workareaList,
    loading,
    pagination,
    includeConditions,
    allWorkareaList,
    allWorkUnitList,
    allAccessPointList,
    handleCreateWorkarea,
    handleDeleteWorkarea,
    handleFetchWorkareaList,
    handleGetAllWorkareaList,
    handleFetchAllWorkarea,
    handleFetchAllWorkUnit,
  };
});
