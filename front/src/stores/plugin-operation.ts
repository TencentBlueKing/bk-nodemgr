import { defineStore } from 'pinia';
import { ref } from 'vue';

export const usePluginOperationStore = defineStore('plugin-operation', () => {
  /** 缓存的参数配置副本（platform → formValues），供跨页面粘贴 */
  const cachedConfig = ref<Record<string, Record<string, any>> | null>(null);

  function setCopiedConfig(data: Record<string, Record<string, any>>) {
    cachedConfig.value = JSON.parse(JSON.stringify(data));
  }

  function getCopiedConfig(): Record<string, Record<string, any>> | null {
    return cachedConfig.value;
  }

  function clearCopiedConfig() {
    cachedConfig.value = null;
  }

  return { cachedConfig, setCopiedConfig, getCopiedConfig, clearCopiedConfig };
});
