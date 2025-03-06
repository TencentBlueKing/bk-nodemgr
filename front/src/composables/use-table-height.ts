import { onMounted, onUnmounted, ref } from "vue";

export default function useDynamicsHeight(offsetHeight: number) {
  const maxHeight = ref<number>(0);
  function calcHeight() {
    maxHeight.value = window.innerHeight - offsetHeight;
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
