import { keyBy } from "lodash";
import { defineStore } from "pinia";
import { reactive, ref } from "vue";

import type {
  TopoEventListReq,
  TopoNetworkAreaCreateReq,
  TopoNetworkAreaListReq,
  TopoNetworkAreaStaticsRespStaticsInfo,
  TopoNetworkAreaUpdateReq,
} from "@/@types/topo";
import { TopoService } from "@/api/modules/topo";

export type INetWorkArea = NetworkArea & TopoNetworkAreaStaticsRespStaticsInfo;

export const useWorkareaStore = defineStore("workarea", () => {
  const workareaList = ref<INetWorkArea[]>([]);
  const allWorkareaList = ref<Map<number, NetworkArea>>(new Map());
  const allWorkUnitList = ref<Map<number, NetworkUnit[]>>(new Map());
  const allAccessPointList = ref<Map<number, AccessPoint[]>>(new Map());

  const vendorList = ref<string[]>([]);
  const osTypeList = ref<string[]>([]);
  // const all
  const loading = ref(false);
  const pagination = reactive({ count: 0, limit: 50, current: 1 });

  interface IncludeConditions {
    bk_networkarea_id: number[]; // 管控区域ID
    bk_networkarea_name: string[]; // 管控区域
    cloud_vendor: number[]; // 云服务商
  }

  // 可搜索字段
  const includeConditions = reactive<IncludeConditions>({
    bk_networkarea_id: [],
    cloud_vendor: [],
    bk_networkarea_name: [],
  });

  // 获取管控区域列表
  const handleFetchWorkareaList = async () => {
    loading.value = true;
    try {
      const result = await TopoService.NetworkAreaList({
        page: {
          // todo
          // 接口文档示例 offset为0，了解下前端是否需要-1
          offset: (pagination.current - 1) * pagination.limit,
          limit: pagination.limit,
        },
        onlyCount: false,
        exact_include_conditions: {
          bk_networkarea_id: includeConditions.bk_networkarea_id,
          cloud_vendor: includeConditions.cloud_vendor,
        },
        fuzzy_include_conditions: {
          bk_networkarea_name: includeConditions.bk_networkarea_name,
        }
      });
      workareaList.value = (result?.items as INetWorkArea[]) || [];
      pagination.count = result?.total || 0;
    } catch (err) {
      console.error(err);
    } finally {
      loading.value = false;
    }
  };

  // 分页操作
  const pageLimitChange = async (limit: number) => {
    pagination.limit = limit;
    await handleFetchWorkareaList();
  };
  const pageValueChange = async (current: number) => {
    pagination.current = current;
    await handleFetchWorkareaList();
  };

  const handleDeleteWorkarea = async (bk_networkarea_id: number) => {
    const result = await TopoService.NetworkAreaDelete({
      bk_networkarea_id,
    }).catch(() => {});
    return result;
  };

  const handleCreateWorkarea = async (params: TopoNetworkAreaCreateReq) => {
    const result = await TopoService.NetworkAreaCreate(params).catch(() => false);
    return result;
  };

  const handleUpdateWorkarea = async (params: TopoNetworkAreaUpdateReq) => {
    const result = await TopoService.NetworkAreaUpdate(params).catch(() => false);
    return result;
  };

  // 获取所有管控区域数据，用于复制
  const handleGetAllWorkareaList = async () => {
    loading.value = true;
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0,
      },
    };
    const result = await TopoService.NetworkAreaList(params).catch(() => {});
    workareaList.value = (result?.items as INetWorkArea[]) || [];
    pagination.count = result?.total || 0;

    const workareaIds = workareaList.value.map(
      (item) => item.bk_networkarea_id
    );
    const countData = await handleFetchWorkareaInfoCount(workareaIds);
    const lookup = keyBy(countData, "bk_networkarea_id");
    for (const item of workareaList.value) {
      const match = lookup[item.bk_networkarea_id];
      if (match) {
        item.networkunit_count = match.networkunit_count;
        item.proxy_count = match.proxy_count;
        item.agent_count = match.agent_count;
        item.last_operate_time = match.last_operate_time;
        item.last_operator = match.last_operator;
      }
    }
    loading.value = false;
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
  // 将所有管控单元 接入点存入Map
  const handleFetchAllWorkUnit = async (id?: number) => {
    allWorkUnitList.value.clear();
    allAccessPointList.value.clear();
    const result = await TopoService.NetworkUnitList({
      bk_networkarea_id: id || null,
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

  // 获取操作记录列表
  const handleFetchRecordList = async (params: Partial<TopoEventListReq>) => {
    const result = await TopoService.EventList(params);
    return result;
  };

  // 获取管控区域列表中 管控单元数量及节点数量
  const handleFetchWorkareaInfoCount = async (bk_networkarea_id: number[]) => {
    const result = await TopoService.NetworkAreaStatistics({
      bk_networkarea_id,
    });
    return result?.items || [];
  };

  const handleFetchVendorAndOs = async () => {
    const result = await TopoService.ConstantGet({
      cloud_vendor: true,
      os_type: true,
    });
    vendorList.value = result?.cloud_vendor || [];
    osTypeList.value = result?.os_type || [];
  };

  return {
    workareaList,
    loading,
    pagination,
    includeConditions,
    allWorkareaList,
    allWorkUnitList,
    allAccessPointList,
    vendorList,
    osTypeList,
    pageLimitChange,
    pageValueChange,
    handleCreateWorkarea,
    handleUpdateWorkarea,
    handleDeleteWorkarea,
    handleFetchWorkareaList,
    handleGetAllWorkareaList,
    handleFetchAllWorkarea,
    handleFetchAllWorkUnit,
    handleFetchRecordList,
    handleFetchVendorAndOs,
  };
});
