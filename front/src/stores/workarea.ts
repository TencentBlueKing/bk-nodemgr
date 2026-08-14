import { keyBy } from 'lodash';
import { defineStore } from 'pinia';
import { reactive, ref, watch } from 'vue';

import type {
  TopoEventListReq,
  TopoNetworkAreaCreateReq,
  TopoNetworkAreaListReq,
  TopoNetworkAreaStatisticsRespStatisticsInfo,
  TopoNetworkAreaUpdateReq,
} from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';
import usePage from '@/composables/use-page';
import { useAuthStore } from '@/stores/auth';

export type INetWorkArea = NetworkArea & TopoNetworkAreaStatisticsRespStatisticsInfo;

export const useWorkareaStore = defineStore('workarea', () => {
  const workareaList = ref<INetWorkArea[]>([]);
  const allWorkareaList = ref<Map<number, NetworkArea>>(new Map());
  const allWorkUnitList = ref<Map<number, NetworkUnitDetail[]>>(new Map());
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
    cloud_vendor: string[]; // 云服务商
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

  const waitForNetworkAreaViewAuth = async () => {
    const authStore = useAuthStore();
    if (authStore.authorizedLoaded && authStore.authorizedMap.networkarea_view) return;

    await new Promise<void>((resolve) => {
      const unwatch = watch(
        () => [authStore.authorizedLoaded, !!authStore.authorizedMap.networkarea_view],
        ([loaded, actionReady]) => {
          if (loaded && actionReady) {
            unwatch();
            resolve();
          }
        },
        { immediate: true },
      );
    });
  };

  const updateWorkareaStatistics = async (networkAreaIDs: number[]) => {
    if (networkAreaIDs.length === 0) return;

    await waitForNetworkAreaViewAuth();
    const authStore = useAuthStore();
    const authorizedIDs = networkAreaIDs.filter(id => authStore.hasAuthorizedResource('networkarea_view', id));
    if (authorizedIDs.length === 0) return;

    const countData = await handleFetchWorkareaInfoCount(authorizedIDs).catch(() => []);
    const lookup = keyBy(countData, 'bk_networkarea_id');
    workareaList.value = workareaList.value.map((item) => {
      const statistics = lookup[item.bk_networkarea_id];
      return statistics ? { ...item, ...statistics } : item;
    });
  };

  // 获取当前页的统计信息（不重新请求管控区域数据）
  const handleFetchCurrentPageStatistics = async () => {
    if (workareaList.value.length === 0) return;

    const startIndex = (frontPageConf.current - 1) * frontPageConf.limit;
    const endIndex = startIndex + frontPageConf.limit;
    const currentPageWorkareaIds = workareaList.value
      .slice(startIndex, endIndex)
      .map(item => item.bk_networkarea_id);
    await updateWorkareaStatistics(currentPageWorkareaIds);
  };

  // 获取全部管控区域统计信息，用于基于统计值的前端筛选
  const handleFetchAllWorkareaStatistics = async () => {
    await updateWorkareaStatistics(workareaList.value.map(item => item.bk_networkarea_id));
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
    const list = ((result?.items || []) as INetWorkArea[]).sort((a: INetWorkArea, b: INetWorkArea) => {
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

    const accessPointIDs = [...new Set(list.flatMap(unit => unit.accesspoints || []))];
    const accessPointResult = accessPointIDs.length > 0
      ? await TopoService.AccessPointList({
        page: { offset: 0, limit: accessPointIDs.length },
        only_count: false,
        exact_include_conditions: {
          accesspoint_id: accessPointIDs,
        },
      }).catch(() => ({ total: 0, items: [] }))
      : { total: 0, items: [] };
    const accessPointMap = new Map<number, AccessPoint>((accessPointResult.items || [])
      .map(item => [item.accesspoint_id, item]));

    for (const unit of list) {
      const detailUnit: NetworkUnitDetail = {
        ...unit,
        accesspoints: (unit.accesspoints || [])
          .map(accessPointID => accessPointMap.get(accessPointID))
          .filter((item): item is AccessPoint => !!item),
      };

      allAccessPointList.value.set(detailUnit.bk_networkunit_id, detailUnit.accesspoints);

      const units = allWorkUnitList.value.get(detailUnit.bk_networkarea_id) || [];
      units.push(detailUnit);
      allWorkUnitList.value.set(detailUnit.bk_networkarea_id, units);
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
    handleFetchAllWorkareaStatistics,
    syncFavoriteWorkareaList,
  };
});
