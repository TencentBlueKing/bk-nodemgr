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
    const res = await PackageService.ListRelease({
      release_type: currentType.value,
      generation: 2,
    });
    const allLabels = res.items.flatMap(item => item.labels || []);
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
