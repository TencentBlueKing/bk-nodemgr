// @vitest-environment node
import { describe, expect, it } from 'vitest';

import {
  buildExecuteUpgradeParams,
  buildUpgradeCheckParams,
} from '../src/pages/node/agent/upgrade-request';

const hosts = [
  {
    bk_host_id: 1001,
    os_type: 'linux',
    cpu_arch: 'amd64',
  },
  {
    bk_host_id: 1002,
    os_type: 'linux',
    cpu_arch: 'arm64',
  },
];

const formData = {
  networkUnitId: 42,
  targetVersions: [
    { os_type: 'linux', cpu_arch: 'amd64', version: '7.1.0' },
    { os_type: 'linux', cpu_arch: 'arm64', version: '7.1.1' },
  ],
  force: true,
  gracefulRestartTimeoutSec: 30,
};

describe('upgrade request builders', () => {
  it('includes cpu_arch in upgrade check payload', () => {
    expect(buildUpgradeCheckParams(hosts, formData)).toEqual({
      host: [
        { bk_host_id: 1001, bk_networkunit_id: 42, cpu_arch: 'amd64' },
        { bk_host_id: 1002, bk_networkunit_id: 42, cpu_arch: 'arm64' },
      ],
      target_version: formData.targetVersions,
    });
  });

  it('builds agent execute payload with per-host target version', () => {
    expect(buildExecuteUpgradeParams('agent', hosts, formData)).toEqual({
      host: [
        {
          bk_host_id: 1001,
          bk_networkunit_id: 42,
          cpu_arch: 'amd64',
          target_version: '7.1.0',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
        {
          bk_host_id: 1002,
          bk_networkunit_id: 42,
          cpu_arch: 'arm64',
          target_version: '7.1.1',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
      ],
    });
  });

  it('builds proxy execute payload with shared target versions', () => {
    expect(buildExecuteUpgradeParams('proxy', hosts, formData)).toEqual({
      host: [
        {
          bk_host_id: 1001,
          bk_networkunit_id: 42,
          cpu_arch: 'amd64',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
        {
          bk_host_id: 1002,
          bk_networkunit_id: 42,
          cpu_arch: 'arm64',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
      ],
      target_version: formData.targetVersions,
    });
  });

  it('keeps current network unit when user does not choose one', () => {
    const keepCurrentFormData = {
      ...formData,
      networkUnitId: undefined,
    };

    expect(buildUpgradeCheckParams(hosts, keepCurrentFormData)).toEqual({
      host: [
        { bk_host_id: 1001, cpu_arch: 'amd64' },
        { bk_host_id: 1002, cpu_arch: 'arm64' },
      ],
      target_version: formData.targetVersions,
    });

    expect(buildExecuteUpgradeParams('agent', hosts, keepCurrentFormData)).toEqual({
      host: [
        {
          bk_host_id: 1001,
          cpu_arch: 'amd64',
          target_version: '7.1.0',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
        {
          bk_host_id: 1002,
          cpu_arch: 'arm64',
          target_version: '7.1.1',
          force: true,
          graceful_restart_timeout_sec: 30,
        },
      ],
    });
  });
});
