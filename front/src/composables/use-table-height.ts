import { onMounted, onUnmounted, ref } from 'vue';

import { useMainStore } from '@/stores/main';

export default function useDynamicsHeight(offsetHeight: number) {
  const mainStore = useMainStore();
  const maxHeight = ref<number>(0);
  function calcHeight() {
    maxHeight.value = window.innerHeight - offsetHeight - (mainStore.noticeShow ? 40 : 0);
  };

  onMounted(() => {
    calcHeight();
    window.addEventListener('resize', calcHeight);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', calcHeight);
  });

  return {
    maxHeight,
  };
};
