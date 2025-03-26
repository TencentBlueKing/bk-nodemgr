import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';

export const useRouteSubTitle = defineStore('routeSubTile', () => {
  const route = useRoute();
  const subTitle = ref(route.meta.subTitle);

  watch(() => route.fullPath, () => {
    subTitle.value = route.meta?.subTitle;
  });

  return {
    subTitle,
  };
});
