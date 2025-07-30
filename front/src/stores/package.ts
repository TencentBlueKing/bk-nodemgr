import { defineStore } from 'pinia';
import { ref } from 'vue';

import { PackageService } from "@/api/modules/pkg";
import { useRoute } from "vue-router";

export const usePackageStore = defineStore('package', () => {
  const route = useRoute();
  const tagList = ref<string[]>([]);
  const getPackages = async () => {
    const currentType = route.name === "agentPackageMng" ? "agent" : "proxy";
    const res = await PackageService.ListRelease({
      exact_include_conditions: {
        release_type: [currentType],
      },
    });
    const allLabels = res.items.flatMap(item => item.labels || []);
    tagList.value = Array.from(new Set(allLabels));
  }
  return {
    tagList,
    getPackages,
  };
});
