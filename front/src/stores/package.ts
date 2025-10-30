import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';

import { PackageService } from '@/api/modules/pkg';

export const usePackageStore = defineStore('package', () => {
  const route = useRoute();
  const tagList = ref<string[]>([]);
  const currentType = computed(() => {
    const routeName = route.name?.toString() || '';
    const type = routeName.split('PackageMng')[0];
    return type;
  });
  const getPackages = async () => {
    let res;
    if (currentType.value === 'agent') {
      res = await PackageService.ListReleaseAgent({
        generation: 2,
        exact_include_conditions: {
          release_type: [currentType.value],
        },
      }).catch(() => ({
        total: 0,
        items: [],
      }));
    } else {
      res = await PackageService.ListReleaseProxy({
        generation: 2,
        exact_include_conditions: {
          release_type: [currentType.value],
        },
      }).catch(() => ({
        total: 0,
        items: [],
      }));
    }
    const allLabels = res.items.flatMap(item => item.release.labels || []);
    tagList.value = Array.from(new Set(allLabels));
  };

  const updateTagList = (list: string[]) => {
    tagList.value = list;
  };

  return {
    tagList,
    getPackages,
    updateTagList,
  };
});
