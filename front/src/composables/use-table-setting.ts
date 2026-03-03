import type { ISettings } from 'node_modules/@blueking/table/typings/components/setting-column/Index.vue';
import { reactive, ref, watch } from 'vue';

export interface SettingsConfig {
  checked?: string[];
  disabled?: string[];
}

// 每个 Hook 调用可传入一个自定义 key，用于隔离不同表格/模块的设置
export default function useTableSetting(
  { checked = [], disabled = [] }: SettingsConfig,
  storageKey = 'bk-table-default-settings', // 允许自定义 key
) {
  // 1. 从 localStorage 读取保存的设置
  const savedSettingsStr = localStorage.getItem(storageKey);
  const savedSettings: {
    checked?: string[];
    disabled?: string[];
    size?: string;
    knownFields?: string[];
  } = savedSettingsStr ? JSON.parse(savedSettingsStr) : {};

  // 2. 初始化 settings，优先使用用户保存的值，其次使用传入的默认值
  //    通过 knownFields 追踪用户已知字段，仅自动添加真正的新增字段
  //    迁移兼容：旧格式无 knownFields 时，视当前默认字段为已知，避免重置用户偏好
  const savedKnownFields = savedSettings.knownFields
    ?? (savedSettings.checked ? checked : []);
  const newFields = savedSettings.checked
    ? checked.filter(field => !savedKnownFields.includes(field))
    : [];
  const mergedChecked = savedSettings.checked
    ? [...new Set([...savedSettings.checked, ...newFields])]
    : checked;
  const settings = reactive({
    checked: mergedChecked,
    disabled: savedSettings.disabled ?? disabled,
    size: savedSettings.size ?? 'medium',
  });

  const isShowSetting = ref(true);

  const handleSettingChange = (data: ISettings) => {
    if (typeof data.size === 'string') {
      settings.size = data.size;
    }
    if (Array.isArray(data.checked)) settings.checked = data.checked;
  };

  // 3. 监听 settings 变更，自动保存到 localStorage
  watch(
    () => ({
      checked: settings.checked,
      disabled: settings.disabled,
      size: settings.size,
    }),
    (newVal) => {
      localStorage.setItem(
        storageKey,
        JSON.stringify({
          checked: newVal.checked,
          disabled: newVal.disabled,
          size: newVal.size,
          knownFields: [...new Set([...checked, ...newVal.checked])],
        }),
      );
    },
    { deep: true },
  );

  return {
    isShowSetting,
    settings,
    handleSettingChange,
  };
}
