import { keyBy } from 'lodash';
import { defineStore } from 'pinia';
import { onUpdated, reactive, ref } from 'vue';
import usePage from '@/composables/use-page';

import type {
  TopoEventListReq,
  TopoNetworkAreaCreateReq,
  TopoNetworkAreaListReq,
  TopoNetworkAreaStaticsRespStaticsInfo,
  TopoNetworkAreaUpdateReq,
} from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';

export type INetWorkArea = NetworkArea & TopoNetworkAreaStaticsRespStaticsInfo;

export const useWorkareaStore = defineStore('workarea', () => {
  const workareaList = ref<INetWorkArea[]>([]);
  const allWorkareaList = ref<Map<number, NetworkArea>>(new Map());
  const allWorkUnitList = ref<Map<number, NetworkUnit[]>>(new Map());
  const allAccessPointList = ref<Map<number, AccessPoint[]>>(new Map());

  // 收藏的管控区域
  const favoriteWorkareaList = ref<number[]>(JSON.parse(localStorage.getItem('collect_workarea') || '[]'));

  const vendorList = ref<string[]>([]);
  const osTypeList = ref<string[]>([]);
  // const all
  const loading = ref(false);
  const pagination = reactive({ count: 0, limit: 10, current: 1 });

  const { 
    pagination: frontPagination,
    pageConf: frontPageConf,
  } = usePage(workareaList);

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
      // 先同步收藏状态，确保排序使用最新数据
      syncFavoriteWorkareaList();

      const result = await TopoService.NetworkAreaList({
        page: { offset: 0, limit: 0 },
        onlyCount: false,
        exact_include_conditions: {
          bk_networkarea_id: includeConditions.bk_networkarea_id,
          cloud_vendor: includeConditions.cloud_vendor,
        },
        fuzzy_include_conditions: {
          bk_networkarea_name: includeConditions.bk_networkarea_name,
        },
      });
      const allWorkareaList = (result?.items as INetWorkArea[]) || [];
      allWorkareaList.sort((a: INetWorkArea, b: INetWorkArea) => {
        // bk_networkarea_id为0的始终排在最前面
        if (a.bk_networkarea_id === 0) return -1;
        if (b.bk_networkarea_id === 0) return 1;
        const aIsFavorite = favoriteWorkareaList.value.includes(a.bk_networkarea_id);
        const bIsFavorite = favoriteWorkareaList.value.includes(b.bk_networkarea_id);
        if (aIsFavorite && bIsFavorite) return b.bk_networkarea_id - a.bk_networkarea_id;
        if (aIsFavorite && !bIsFavorite) return -1;
        if (!aIsFavorite && bIsFavorite) return 1;
        return b.bk_networkarea_id - a.bk_networkarea_id;
      });
      
      // 设置所有数据到store
      workareaList.value = allWorkareaList;
      pagination.count = result?.total || 0;
      loading.value = false; // 立即结束加载，先渲染

      // 异步获取当前页的统计信息
      Promise.resolve().then(async () => {
        await handleFetchCurrentPageStatistics();
      });
    } catch (err) {
      console.error(err);
    } finally {
      loading.value = false;
    }
  };

  // 获取当前页的统计信息（不重新请求管控区域数据）
  const handleFetchCurrentPageStatistics = async () => {
    if (workareaList.value.length === 0) return;
    
    // 获取当前页的数据（分页切片）
    const startIndex = (frontPageConf.current - 1) * frontPageConf.limit;
    const endIndex = startIndex + frontPageConf.limit;
    const currentPageData = workareaList.value.slice(startIndex, endIndex);
    const currentPageWorkareaIds = currentPageData.map(item => item.bk_networkarea_id);
    
    if (currentPageWorkareaIds.length > 0) {
      const countData = await handleFetchWorkareaInfoCount(currentPageWorkareaIds).catch(() => []);
      const lookup = keyBy(countData, 'bk_networkarea_id');

      // 只更新当前页的数据（触发响应式更新）
      workareaList.value = workareaList.value.map(item => {
        if (currentPageWorkareaIds.includes(item.bk_networkarea_id)) {
          return {
            ...item,
            ...lookup[item.bk_networkarea_id],
          };
        }
        return item;
      });
    }
  };

  // 分页操作
  const pageLimitChange = async (limit: number) => {
    pagination.limit = limit;
    await handleFetchCurrentPageStatistics();
  };
  const pageValueChange = async (current: number) => {
    pagination.current = current;
    await handleFetchCurrentPageStatistics();
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
      page: { offset: 0, limit: 0 },
    };

    try {
      // 先同步收藏状态，确保排序使用最新数据
      syncFavoriteWorkareaList();

      // 1. 第一次请求：获取所有基础列表数据
      const result = await TopoService.NetworkAreaList(params);
      const allWorkareaList = (result?.items as INetWorkArea[]) || [];
      
      // 排序所有数据
      allWorkareaList.sort((a: INetWorkArea, b: INetWorkArea) => {
        // bk_networkarea_id为0的始终排在最前面
        if (a.bk_networkarea_id === 0) return -1;
        if (b.bk_networkarea_id === 0) return 1;
        const aIsFavorite = favoriteWorkareaList.value.includes(a.bk_networkarea_id);
        const bIsFavorite = favoriteWorkareaList.value.includes(b.bk_networkarea_id);
        if (aIsFavorite && bIsFavorite) return b.bk_networkarea_id - a.bk_networkarea_id;
        if (aIsFavorite && !bIsFavorite) return -1;
        if (!aIsFavorite && bIsFavorite) return 1;
        return b.bk_networkarea_id - a.bk_networkarea_id;
      });
      
      // 设置所有数据到store
      workareaList.value = allWorkareaList;
      pagination.count = result?.total || 0;
      loading.value = false; // 立即结束加载，先渲染

      // 2. 用微任务异步执行第二次请求：只请求当前页的统计信息
      // 异步获取当前页的统计信息
      Promise.resolve().then(async () => {
        await handleFetchCurrentPageStatistics();
      });
    } catch (error) {
      workareaList.value = [];
      pagination.count = 0;
      loading.value = false;
    }
  };

  // 将所有管控区域存入Map
  const handleFetchAllWorkarea = async () => {
    // 先同步收藏状态，确保排序使用最新数据
    syncFavoriteWorkareaList();

    allWorkareaList.value.clear();
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0,
      },
    };
    const result = await TopoService.NetworkAreaList(params).catch(() => {});
    const list = result?.items.sort((a: INetWorkArea, b: INetWorkArea) => {
      // bk_networkarea_id为0的始终排在最前面
      if (a.bk_networkarea_id === 0) return -1;
      if (b.bk_networkarea_id === 0) return 1;
      const aIsFavorite = favoriteWorkareaList.value.includes(a.bk_networkarea_id);
      const bIsFavorite = favoriteWorkareaList.value.includes(b.bk_networkarea_id);
      if (aIsFavorite && bIsFavorite) return b.bk_networkarea_id - a.bk_networkarea_id;
      if (aIsFavorite && !bIsFavorite) return -1;
      if (!aIsFavorite && bIsFavorite) return 1;
      return b.bk_networkarea_id - a.bk_networkarea_id;
    }) || [];
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
    }).catch(() => ({
      total: 0,
      items: [],
    }));
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
    const result = await TopoService.EventList(params).catch(() => ({
      total: 0,
      items: [],
    }));
    return result;
  };

  // 获取管控区域列表中 管控单元数量及节点数量
  const handleFetchWorkareaInfoCount = async (bk_networkarea_id: number[]) => {
    const result = await TopoService.NetworkAreaStatistics({
      bk_networkarea_id,
    }).catch(() => ({
      items: [],
    }));
    return result?.items;
  };

  const handleFetchVendorAndOs = async () => {
    const result = await TopoService.ConstantGet({
      cloud_vendor: true,
      os_type: true,
    }).catch(() => ({
      cloud_vendor: [],
      os_type: [],
    }));;
    vendorList.value = result?.cloud_vendor;
    osTypeList.value = result?.os_type;
  };

  // 同步更新收藏状态
  const syncFavoriteWorkareaList = () => {
    favoriteWorkareaList.value = JSON.parse(localStorage.getItem('collect_workarea') || '[]');
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
    favoriteWorkareaList,
    frontPagination,
    frontPageConf,
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
    handleFetchCurrentPageStatistics,
    syncFavoriteWorkareaList,
  };
});
