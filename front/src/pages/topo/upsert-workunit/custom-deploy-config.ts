// DeployConfigMap is an internal UI representation of DeployConfig,
// keyed by os_type for efficient per-OS lookup.
type DeployConfig = CustomDeployConfig & { os_type?: string };
type DeployConfigMap = Record<string, DeployConfig>;
const WINDOWS_OS = 'windows';

const createEmptyInstallerRuntime = (): InstallerRuntime => ({
  base_work_dir: '',
});

const createEmptyNodeRuntime = (): NodeRuntime => ({
  base_deploy_dir: '',
  data_ipc: '',
  plugin_ipc: '',
  log_dir: '',
  zone_id: 'default',
  city_id: 'default',
});

const createEmptyPluginRuntime = (): PluginRuntime => ({
  base_deploy_dir: '',
  log_dir: '',
});

export const createEmptyDeployConfig = (): DeployConfig => ({
  os_type: '',
  installer_runtime: createEmptyInstallerRuntime(),
  node_runtime: createEmptyNodeRuntime(),
  plugin_runtime: createEmptyPluginRuntime(),
});

const normalizeSingleDeployConfig = (config?: Partial<DeployConfig>): DeployConfig => ({
  os_type: config?.os_type ?? '',
  installer_runtime: {
    ...createEmptyInstallerRuntime(),
    ...(config?.installer_runtime ?? {}),
  },
  node_runtime: {
    ...createEmptyNodeRuntime(),
    ...(config?.node_runtime ?? {}),
  },
  plugin_runtime: {
    ...createEmptyPluginRuntime(),
    ...(config?.plugin_runtime ?? {}),
  },
});

const getConfigKeys = (
  config?: DeployConfigMap,
  osTypes: string[] = [],
): string[] => Array.from(new Set([
  ...osTypes,
  ...Object.keys(config ?? {}),
])).filter(Boolean);

const trimConfigValue = (value: string): string => value.trim();

const getAllConfigValues = (config: CustomDeployConfig): string[] => [
  config.installer_runtime.base_work_dir,
  config.node_runtime.base_deploy_dir,
  config.node_runtime.data_ipc,
  config.node_runtime.plugin_ipc,
  config.node_runtime.log_dir,
  config.node_runtime.zone_id,
  config.node_runtime.city_id,
  config.plugin_runtime.base_deploy_dir,
  config.plugin_runtime.log_dir,
];

const getValidationValues = (config: CustomDeployConfig): string[] => {
  if ((config as DeployConfig).os_type?.toLowerCase() !== WINDOWS_OS) {
    return [
      config.installer_runtime.base_work_dir,
      config.node_runtime.base_deploy_dir,
      config.node_runtime.log_dir,
      config.plugin_runtime.base_deploy_dir,
      config.plugin_runtime.log_dir,
    ];
  }

  return getAllConfigValues(config);
};

const hasEffectiveValue = (config: DeployConfig): boolean => (
  getAllConfigValues(config).some(value => trimConfigValue(value) !== '')
);

// validateCustomDeployConfigAllOrNothing returns true when the config is either
// entirely empty (all fields blank) or entirely filled (no field blank). Returns
// false when only some fields are filled, which is disallowed. For non-windows
// configs, data_ipc and plugin_ipc are excluded from this check because they are
// read-only in UI.
export const validateCustomDeployConfigAllOrNothing = (config: DeployConfig): boolean => {
  const values = getValidationValues(config).map(trimConfigValue);
  const filledCount = values.filter(v => v !== '').length;
  return filledCount === 0 || filledCount === values.length;
};

// deployConfigArrayToMap converts a DeployConfig[] from the API into a map
// keyed by os_type for internal UI use.
export const deployConfigArrayToMap = (configs: DeployConfig[] = []): DeployConfigMap => (
  configs.reduce<DeployConfigMap>(
    (result, item) => {
      if (item.os_type) {
        result[item.os_type] = item;
      }
      return result;
    },
    {},
  )
);

// deployConfigMapFromApi converts the map[os_type]CustomDeployConfig shape
// returned by the application service into the internal DeployConfigMap
// (os_type is re-injected as a field inside each entry).
export const deployConfigMapFromApi = (
  apiMap?: Record<string, CustomDeployConfig>,
): DeployConfigMap => {
  if (!apiMap) return {};
  return Object.entries(apiMap).reduce<DeployConfigMap>(
    (result, [os, config]) => ({
      ...result,
      [os]: { ...config, os_type: os },
    }),
    {},
  );
};

export const normalizeCustomDeployConfigs = (
  config?: DeployConfigMap,
  osTypes: string[] = [],
): DeployConfigMap => getConfigKeys(config, osTypes).reduce<DeployConfigMap>(
  (result, os) => ({
    ...result,
    [os]: normalizeSingleDeployConfig({ ...config?.[os], os_type: os }),
  }),
  {},
);

// CustomDeployConfigPayload is the map-value shape the application service expects.
// It matches proto.CustomDeployConfig (no os_type field — os_type is the map key).
type CustomDeployConfigPayload = CustomDeployConfig;

// serializeCustomDeployConfigs converts the internal DeployConfigMap into the
// map[os_type]CustomDeployConfig shape that the application service accepts.
export const serializeCustomDeployConfigs = (config: DeployConfigMap): Record<string, CustomDeployConfigPayload> => (
  Object.entries(config).reduce<Record<string, CustomDeployConfigPayload>>(
    (result, [os, item]) => {
      const normalizedItem = normalizeSingleDeployConfig({ ...item, os_type: os });

      if (!hasEffectiveValue(normalizedItem)) {
        return result;
      }

      return {
        ...result,
        [os]: {
          installer_runtime: {
            base_work_dir: trimConfigValue(normalizedItem.installer_runtime.base_work_dir),
          },
          node_runtime: {
            base_deploy_dir: trimConfigValue(normalizedItem.node_runtime.base_deploy_dir),
            data_ipc: trimConfigValue(normalizedItem.node_runtime.data_ipc),
            plugin_ipc: trimConfigValue(normalizedItem.node_runtime.plugin_ipc),
            log_dir: trimConfigValue(normalizedItem.node_runtime.log_dir),
            zone_id: trimConfigValue(normalizedItem.node_runtime.zone_id ?? ''),
            city_id: trimConfigValue(normalizedItem.node_runtime.city_id ?? ''),
          },
          plugin_runtime: {
            base_deploy_dir: trimConfigValue(normalizedItem.plugin_runtime.base_deploy_dir),
            log_dir: trimConfigValue(normalizedItem.plugin_runtime.log_dir),
          },
        },
      };
    },
    {},
  )
);
