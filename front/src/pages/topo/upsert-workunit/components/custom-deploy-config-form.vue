<template>
  <Form.FormItem
    :label="$t('topoManager.workUnit.form.customDeployConfig')"
    label-width="130"
  >
    <div class="w-[488px]">
      <div v-if="osOptions.length > 0" class="os-tab-bar">
        <div
          v-for="option in osOptions"
          :key="option.id"
          :class="['os-tab-item', { 'os-tab-active': activeOs === option.id }]"
          @click="activeOs = option.id"
        >
          {{ option.name }}
        </div>
      </div>
      <div v-else class="text-[12px] leading-[20px] text-[#979BA5]">
        {{ $t('topoManager.workUnit.form.noOsAvailable') }}
      </div>
      <div v-if="currentConfig" class="mt-[12px] rounded-[2px] bg-[#F5F7FA] px-[16px] py-[16px]">
        <div class="mb-[16px]">
          <div class="mb-[8px] text-[12px] font-medium leading-[20px] text-[#313238]">
            {{ $t('topoManager.workUnit.form.installerRuntime') }}
          </div>
          <div class="space-y-[12px]">
            <div class="flex items-center gap-[12px]">
              <div class="w-[140px] shrink-0 text-[12px] leading-[20px] text-[#4D4F56] text-right">
                {{ $t('topoManager.workUnit.form.baseWorkDir') }}
              </div>
              <Input
                v-model="currentConfig.installer_runtime.base_work_dir"
                class="flex-1"
                clearable
                :placeholder="defaultConfig?.installer_runtime?.base_work_dir || ''"
              />
            </div>
          </div>
        </div>

        <div class="mb-[16px]">
          <div class="mb-[8px] text-[12px] font-medium leading-[20px] text-[#313238]">
            {{ $t('topoManager.workUnit.form.nodeRuntime') }}
          </div>
          <div class="space-y-[12px]">
            <div class="flex items-center gap-[12px]" v-for="field in nodeFields" :key="field.key">
              <div class="w-[140px] shrink-0 text-[12px] leading-[20px] text-[#4D4F56] text-right">
                {{ field.label }}
              </div>
              <Input
                v-model="currentConfig.node_runtime[field.key]"
                class="flex-1"
                :clearable="!isReadonlyIpcField(field.key)"
                :disabled="isReadonlyIpcField(field.key)"
                :placeholder="getNodeFieldPlaceholder(field.key)"
              />
            </div>
          </div>
        </div>

        <div>
          <div class="mb-[8px] text-[12px] font-medium leading-[20px] text-[#313238]">
            {{ $t('topoManager.workUnit.form.pluginRuntime') }}
          </div>
          <div class="space-y-[12px]">
            <div class="flex items-center gap-[12px]" v-for="field in pluginFields" :key="field.key">
              <div class="w-[140px] shrink-0 text-[12px] leading-[20px] text-[#4D4F56] text-right">
                {{ field.label }}
              </div>
              <Input
                v-model="currentConfig.plugin_runtime[field.key]"
                class="flex-1"
                clearable
                :placeholder="defaultConfig?.plugin_runtime?.[field.key] || ''"
              />
            </div>
          </div>
        </div>
      </div>
      <div v-if="currentConfig && hasPartialInput" class="mt-[4px] text-[12px] leading-[20px] text-[#EA3636]">
        {{ $t('topoManager.workUnit.form.customDeployConfigPartialError') }}
      </div>
      <div class="mt-[4px] text-[12px] leading-[20px] text-[#979BA5]">
        {{ $t('topoManager.workUnit.form.customDeployConfigTips') }}
      </div>
    </div>
  </Form.FormItem>
</template>

<script lang="ts" setup>
import { Form, Input } from 'bkui-vue';
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { validateCustomDeployConfigAllOrNothing } from '../custom-deploy-config';

const configs = defineModel<Record<string, CustomDeployConfig>>('configs', { required: true });
const activeOs = defineModel<string>('activeOs', { required: true });
const props = defineProps({
  osOptions: {
    type: Array as () => Array<{ id: string; name: string }>,
    default: () => [],
  },
  defaultConfig: {
    type: Object as () => CustomDeployConfig | null,
    default: null,
  },
});
const { t } = useI18n();
const DEFAULT_CUSTOM_DEPLOY_OS = 'linux';
const WINDOWS_OS = 'windows';

const osLabelMap = computed<Record<string, string>>(() => ({
  linux: t('topoManager.workUnit.form.osOptions.linux'),
  windows: t('topoManager.workUnit.form.osOptions.windows'),
  aix: t('topoManager.workUnit.form.osOptions.aix'),
  freebsd: t('topoManager.workUnit.form.osOptions.freebsd'),
  solaris: t('topoManager.workUnit.form.osOptions.solaris'),
  macos: t('topoManager.workUnit.form.osOptions.macos'),
  darwin: t('topoManager.workUnit.form.osOptions.darwin'),
}));

const nodeFields = computed<Array<{ key: keyof NodeRuntime; label: string }>>(() => ([
  { key: 'base_deploy_dir', label: t('topoManager.workUnit.form.baseDeployDir') },
  { key: 'log_dir', label: t('topoManager.workUnit.form.logDir') },
  { key: 'data_ipc', label: t('topoManager.workUnit.form.dataIPC') },
  { key: 'plugin_ipc', label: t('topoManager.workUnit.form.pluginIPC') },
  { key: 'zone_id', label: t('topoManager.workUnit.form.zoneId') },
  { key: 'city_id', label: t('topoManager.workUnit.form.cityId') },
]));

const pluginFields = computed<Array<{ key: keyof PluginRuntime; label: string }>>(() => ([
  { key: 'base_deploy_dir', label: t('topoManager.workUnit.form.baseDeployDir') },
  { key: 'log_dir', label: t('topoManager.workUnit.form.logDir') },
]));

const ipcNodeFieldKeys: Array<keyof NodeRuntime> = ['data_ipc', 'plugin_ipc'];

const isWindowsOs = computed(() => activeOs.value.toLowerCase() === WINDOWS_OS);

const isReadonlyIpcField = (fieldKey: keyof NodeRuntime): boolean => (
  ipcNodeFieldKeys.includes(fieldKey) && !isWindowsOs.value
);

const getNodeFieldPlaceholder = (fieldKey: keyof NodeRuntime): string => {
  if (isReadonlyIpcField(fieldKey)) {
    return t('topoManager.workUnit.form.windowsOnlyIpcEditable');
  }
  return props.defaultConfig?.node_runtime?.[fieldKey] || '';
};

const osOptions = computed(() => {
  const optionMap = new Map<string, { id: string; name: string }>();

  props.osOptions.forEach((item) => {
    if (item.id) {
      optionMap.set(item.id, {
        id: item.id,
        name: osLabelMap.value[item.id] ?? item.name,
      });
    }
  });

  Object.keys(configs.value ?? {}).forEach((item) => {
    if (item && !optionMap.has(item)) {
      optionMap.set(item, {
        id: item,
        name: osLabelMap.value[item] ?? item,
      });
    }
  });

  return Array.from(optionMap.values());
});

const currentConfig = computed(() => {
  if (!activeOs.value) {
    return undefined;
  }
  return configs.value?.[activeOs.value];
});

const hasPartialInput = computed(() => {
  if (!currentConfig.value) {
    return false;
  }
  return !validateCustomDeployConfigAllOrNothing(currentConfig.value);
});

watch(osOptions, (options) => {
  if (options.length === 0) {
    activeOs.value = '';
    return;
  }
  if (!options.some(item => item.id === activeOs.value)) {
    activeOs.value = options.find(item => item.id === DEFAULT_CUSTOM_DEPLOY_OS)?.id ?? options[0].id;
  }
}, { immediate: true });
</script>

<style lang="postcss" scoped>
.os-tab-bar {
  display: flex;
  border-bottom: 1px solid #dcdee5;
  gap: 0;
}

.os-tab-item {
  padding: 6px 16px;
  font-size: 12px;
  line-height: 20px;
  color: #63656e;
  cursor: pointer;
  border: 1px solid transparent;
  border-bottom: none;
  border-radius: 4px 4px 0 0;
  transition: all 0.2s;
  position: relative;
  bottom: -1px;

  &:hover {
    color: #3a84ff;
  }

  &.os-tab-active {
    color: #3a84ff;
    background: #fff;
    border-color: #dcdee5;
    border-bottom-color: #fff;
    font-weight: 500;
  }
}
</style>
