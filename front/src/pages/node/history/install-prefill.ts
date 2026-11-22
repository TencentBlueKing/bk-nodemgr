import type {
  NodeWorkflowOperationInstallParams,
  NodeWorkflowOperationInstallParamsGetRespData,
} from '@/@types/node_workflow';

export type HistoricalInstallItem = NodeWorkflowOperationInstallParams;
export type HistoricalInstallPrefill = NodeWorkflowOperationInstallParamsGetRespData;

// Undefined means the selected operations disagree; callers must ask for a new shared value.
export function commonInstallValue<T>(values: T[]): T | undefined {
  return values.length && values.every(value => value === values[0]) ? values[0] : undefined;
}

export function historicalInstallMode(items: HistoricalInstallItem[]): string {
  return commonInstallValue(items.map((item) => {
    if (item.is_offline) return 'offline';
    return item.is_manual ? 'manual' : 'setup';
  })) ?? '';
}

// Whitelist editable fields. Historical credentials and host identity never enter a fresh install form.
export function historicalInstallRows(items: HistoricalInstallItem[], proxy = false) {
  return items.map(({ host, missing_fields: missingFields }) => ({
    bk_addressing: host.bk_addressing,
    bk_host_innerip: host.bk_host_innerip.join(','),
    bk_host_innerip_v6: host.bk_host_innerip_v6.join(','),
    os_type: host.os_type,
    login_ip: host.login_ip,
    login_port: host.login_port ? String(host.login_port) : '',
    login_user: host.login_user,
    login_mode: host.login_mode,
    is_sudo: host.is_sudo,
    re_register: host.re_register,
    install_pre_ordered_plugins: host.install_pre_ordered_plugins,
    renew_gse_task: host.renew_gse_task,
    renew_gse_proc: host.renew_gse_proc,
    install_method: missingFields.includes('install_method') ? '' : host.install_method,
    credit: '',
    login_password: '',
    login_key_file: '',
    login_credit_valid: false,
    ...(proxy ? {
      export_ip: host.export_ip,
      export_ip_v6: host.export_ip_v6,
      advertise_ip: host.advertise_ip,
      advertise_ip_v6: host.advertise_ip_v6,
      cpu_arch: host.cpu_arch,
      proxy_tags: [...host.proxy_tags],
      dedicated_installer: host.proxy_tags.includes('dedicated_installer'),
      cluster_tunnel: host.proxy_tags.includes('cluster_tunnel'),
      file_tunnel: host.proxy_tags.includes('file_tunnel'),
      data_tunnel: host.proxy_tags.includes('data_tunnel'),
    } : {}),
  }));
}

// A conflicting or disabled historical version stays empty, never silently becoming the latest version.
export function historicalInstallVersions(items: HistoricalInstallItem[], available: TargetVersion[]) {
  const requested = items.flatMap(item => item.target_version);
  const keys = [...new Set((requested.length ? requested : available)
    .map(item => `${item.os_type}_${item.cpu_arch}`))];
  return keys.map((key) => {
    const versions = items.map(item => item.target_version
      .find(version => `${version.os_type}_${version.cpu_arch}` === key)?.version ?? '');
    const version = commonInstallValue(versions);
    const enabled = available.find(item => `${item.os_type}_${item.cpu_arch}` === key && item.version === version);
    return { os: key, version: enabled?.version ?? '' };
  });
}
