export interface UpgradeRequestHost {
  bk_host_id: number;
  os_type?: string;
  cpu_arch?: string;
  bk_networkunit_id?: number;
}

export interface UpgradeTargetVersion {
  os_type: string;
  cpu_arch: string;
  version: string;
}

export interface UpgradeRequestFormData {
  networkUnitId?: number;
  targetVersions: UpgradeTargetVersion[];
  force: boolean;
  gracefulRestartTimeoutSec: number;
}

const maybeNetworkUnitField = (networkUnitId?: number) => (
  networkUnitId && networkUnitId > 0
    ? { bk_networkunit_id: networkUnitId }
    : {}
);

export const buildUpgradeCheckParams = (
  hosts: UpgradeRequestHost[],
  formData: UpgradeRequestFormData,
) => ({
  host: hosts.map(host => ({
    bk_host_id: host.bk_host_id,
    cpu_arch: host.cpu_arch ?? '',
    ...maybeNetworkUnitField(host.bk_networkunit_id ?? formData.networkUnitId),
  })),
  target_version: formData.targetVersions,
});

export const buildExecuteUpgradeParams = (
  releaseType: 'agent' | 'proxy',
  hosts: UpgradeRequestHost[],
  formData: UpgradeRequestFormData,
) => {
  const baseHosts = hosts.map(host => ({
    bk_host_id: host.bk_host_id,
    cpu_arch: host.cpu_arch ?? '',
    force: formData.force,
    graceful_restart_timeout_sec: formData.gracefulRestartTimeoutSec,
    ...maybeNetworkUnitField(host.bk_networkunit_id ?? formData.networkUnitId),
  }));

  if (releaseType === 'proxy') {
    return {
      host: baseHosts,
      target_version: formData.targetVersions,
    };
  }

  const versionMap = new Map(
    formData.targetVersions.map(item => [`${item.os_type}:${item.cpu_arch}`, item.version]),
  );

  return {
    host: hosts.map(host => ({
      bk_host_id: host.bk_host_id,
      cpu_arch: host.cpu_arch ?? '',
      target_version: versionMap.get(`${host.os_type ?? ''}:${host.cpu_arch ?? ''}`) ?? '',
      force: formData.force,
      graceful_restart_timeout_sec: formData.gracefulRestartTimeoutSec,
      ...maybeNetworkUnitField(host.bk_networkunit_id ?? formData.networkUnitId),
    })),
  };
};
