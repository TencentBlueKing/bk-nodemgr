import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

import { TopoService } from '@/api/modules/topo';
import { TopoNetworkAreaCreateReq, TopoNetworkAreaListReq } from '@/@types/topo';

export const useWorkareaStore = defineStore('workarea', () => {
  const list = ref([]);
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
    list.value = result?.items;
    pagination.count = result?.total;
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
  // todo 接口type没有给出详细类型 这里需要了解下 暂时写为unknown
  const handleFetchAllWorkAreaList = async (): Promise<unknown[]> => {
    const params: TopoNetworkAreaListReq = {
      page: {
        // todo
        // 接口文档示例 offset为0，了解下前端是否需要-1
        offset: pagination.current - 1,
        limit: pagination.count,
      },
      onlyCount: false,
      includeConditions: {
        bk_networkarea_id: [],
        bk_networkarea_name: [],
        bk_cloud_vendor: [],
      },
    };
    const result = await TopoService.NetworkAreaList(params).catch(() => {});
    return result?.list || [];
  };

  return {
    list,
    loading,
    pagination,
    includeConditions,
    handleCreateWorkarea,
    handleDeleteWorkarea,
    handleFetchWorkareaList,
    handleFetchAllWorkAreaList,
  };
});
