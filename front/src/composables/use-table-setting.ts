import type { ISettings } from 'node_modules/@blueking/table/typings/components/setting-column/Index.vue';
import { reactive, ref } from 'vue';

export interface SettingsConfig {
  checked?: string[]
  disabled?: string[]
}

export default function useTableSetting({ checked = [], disabled = [] }: SettingsConfig) {
  const settings = reactive({
    checked,
    disabled,
    size: 'medium', // size: 'medium' 大 | 'mini' 中 | 'small' 小;
  });

  const handleSettingChange = (data: ISettings) => {
    settings.size = data.size as string;
  };
  const isShowSetting = ref(true);

  return {
    isShowSetting,
    settings,
    handleSettingChange,
  };
};
