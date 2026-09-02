# bk-nodemgr-deploy-policy-probe

This is a long-lived v3 plugin package for manually validating plugin-related deploy policy behavior.

## Package

- Upload package: `bk-nodemgr-deploy-policy-probe-1.0.1.tgz`
- Checksum: `bk-nodemgr-deploy-policy-probe-1.0.1.tgz.sha256`
- Source tree: `package/bk-nodemgr-deploy-policy-probe`
- Probe executable source: `cmd/bk-nodemgr-deploy-policy-probe`
- Plugin package name: `bk-nodemgr-deploy-policy-probe`
- Version: `1.0.1`
- Platform: `linux_x86_64`

## Maintenance

- Keep the plugin name stable: `bk-nodemgr-deploy-policy-probe`.
- Bump the patch version for any behavior, template field, or observable output change.
- After changing source files, run `./build-package.sh` and commit the regenerated `.tgz` and `.sha256` files together.
- `build-package.sh` compiles the linux `bin/bk-nodemgr-deploy-policy-probe` executable into a temporary staging tree before packaging.
- Verify the packaged artifact with `sha256sum -c bk-nodemgr-deploy-policy-probe-1.0.1.tgz.sha256`.

## Non-goals

- This package does not cover `specify_agent` or `specify_proxy`.
- This package does not cover `specify_plugin_pkg_sub_config`, which is not an executable deploy spec in the current implementation.
- This package does not validate cross-platform plugin packaging; it only provides `linux_x86_64`.
- This package does not simulate real business collection logic.
- This package is not an automation API contract; it is a manual deploy policy probe.

## CustomContext Schema

Use these stable fields in deploy policy `specs[].param.custom_config_context`:

| Field               | Purpose                                      |
| ------------------- | -------------------------------------------- |
| `ProbeMode`         | identifies the deploy spec scenario          |
| `Marker`            | unique marker expected in rendered configs   |
| `Sequence`          | human-readable sequence for multi-step tests |
| `ExpectedFinalName` | expected rendered config filename            |

## Config Templates

| Template name                         | Purpose                                                                                                   |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `bk-nodemgr-deploy-policy-probe.conf` | required main config for plugin install and upgrade specs                                                 |
| `direct-subconfig-template.conf`      | source template for `specify_plugin_sub_config`; request `name` should remain the final config name       |
| `policy-managed-subconfig.conf`       | source template for `specify_plugin_sub_config_template`; final name is deploy-policy derived             |
| `projected-source-subconfig.conf`     | source template for `project_plugin_config_template_to_hosts`; final name includes source target identity |

## Runtime Observation

The package uses the official linux v3 plugin scripts from `script_tools/plugin_scripts/v3/linux` for lifecycle control.
The probe executable itself does not maintain PID files; GSE owns process trusteeship and PID tracking.

| Command   | Behavior                                                                                                                                        |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `start`   | runs `./start.sh bk-nodemgr-deploy-policy-probe`, which starts `./bk-nodemgr-deploy-policy-probe -c ../etc/bk-nodemgr-deploy-policy-probe.conf` |
| `stop`    | runs `./stop.sh bk-nodemgr-deploy-policy-probe`                                                                                                 |
| `restart` | runs `./restart.sh bk-nodemgr-deploy-policy-probe`                                                                                              |
| `reload`  | runs `./reload.sh bk-nodemgr-deploy-policy-probe` and sends `SIGUSR1` to the running process                                                    |
| `debug`   | runs the probe executable in foreground until `SIGTERM` or `SIGINT`                                                                             |
| `version` | prints `bk-nodemgr-deploy-policy-probe 1.0.1`                                                                                                   |

The probe executable also supports SSH-only observation commands that are not declared in `definition.yaml` control:

| Command  | Output                     |
| -------- | -------------------------- |
| `status` | key-value package metadata |
| `dump`   | key-value package metadata |

## Manual Verification Matrix

The installed config root is environment-specific. Use `etc/...` as the stable package-relative path, and locate it under the node plugin config root on the target host.

| Deploy spec                               | Expected process                       | Expected config path                            | Expected config name                                                                            | Expected marker              | Must not happen                                                       |
| ----------------------------------------- | -------------------------------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------- | ---------------------------- | --------------------------------------------------------------------- |
| `specify_plugin`                          | probe process is running               | `etc`                                           | `bk-nodemgr-deploy-policy-probe.conf`                                                           | `install-or-upgrade`         | package-derived plugin instance name is used                          |
| `specify_plugin_pkg`                      | generated plugin process is running    | `etc`                                           | `bk-nodemgr-deploy-policy-probe.conf`                                                           | `generated-plugin-instance`  | fixed plugin name overwrites generated plugin instance naming         |
| `project_plugin_pkg_to_hosts`             | projected plugin process is running    | `etc`                                           | `bk-nodemgr-deploy-policy-probe.conf`                                                           | `projected-plugin-instance`  | placement host is ignored                                             |
| `specify_plugin_sub_config`               | existing plugin process remains usable | `etc/bk-nodemgr-deploy-policy-probe/subconfigs` | `explicit-direct-subconfig.conf`                                                                | `direct-subconfig`           | `direct-subconfig-template_deploy_<deploy_policy_id>.conf` is created |
| `specify_plugin_sub_config_template`      | existing plugin process remains usable | `etc/bk-nodemgr-deploy-policy-probe/subconfigs` | `policy-managed-subconfig_deploy_<deploy_policy_id>.conf`                                       | `policy-managed-subconfig`   | an explicit direct config name is reused                              |
| `project_plugin_config_template_to_hosts` | existing plugin process remains usable | `etc/bk-nodemgr-deploy-policy-probe/subconfigs` | `projected-source-subconfig_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id>.conf` | `projected-source-subconfig` | source module or source host identity is missing from the final name  |

## Suggested Payload Fragments

Use these fragments in deploy policy `specs[].param` after uploading the package.

### specify_plugin

```json
{
  "plugin_name": "bk-nodemgr-deploy-policy-probe",
  "version": "1.0.1",
  "custom_config_context": {
    "ProbeMode": "specify-plugin",
    "Marker": "install-or-upgrade",
    "Sequence": "001",
    "ExpectedFinalName": "bk-nodemgr-deploy-policy-probe.conf"
  }
}
```

### specify_plugin_pkg

```json
{
  "plugin_pkg_name": "bk-nodemgr-deploy-policy-probe",
  "version": "1.0.1",
  "custom_config_context": {
    "ProbeMode": "specify-plugin-pkg",
    "Marker": "generated-plugin-instance",
    "Sequence": "002",
    "ExpectedFinalName": "bk-nodemgr-deploy-policy-probe.conf"
  }
}
```

### project_plugin_pkg_to_hosts

```json
{
  "plugin_pkg_name": "bk-nodemgr-deploy-policy-probe",
  "version": "1.0.1",
  "placement_host_ids": [10001],
  "custom_config_context": {
    "ProbeMode": "project-plugin-pkg-to-hosts",
    "Marker": "projected-plugin-instance",
    "Sequence": "003",
    "ExpectedFinalName": "bk-nodemgr-deploy-policy-probe.conf"
  }
}
```

### specify_plugin_sub_config

```json
{
  "plugin_name": "bk-nodemgr-deploy-policy-probe",
  "config_files_detail": [
    {
      "template_name": "direct-subconfig-template.conf",
      "name": "explicit-direct-subconfig.conf",
      "is_main_config": false
    }
  ],
  "custom_config_context": {
    "ProbeMode": "specify-plugin-sub-config",
    "Marker": "direct-subconfig",
    "Sequence": "004",
    "ExpectedFinalName": "explicit-direct-subconfig.conf"
  }
}
```

Expected direct sub-config name: `explicit-direct-subconfig.conf`.

The direct spec must not derive a name like `direct-subconfig-template_deploy_<deploy_policy_id>.conf`.

### specify_plugin_sub_config_template

```json
{
  "plugin_name": "bk-nodemgr-deploy-policy-probe",
  "config_files_detail": [
    {
      "template_name": "policy-managed-subconfig.conf",
      "is_main_config": false
    }
  ],
  "custom_config_context": {
    "ProbeMode": "specify-plugin-sub-config-template",
    "Marker": "policy-managed-subconfig",
    "Sequence": "005",
    "ExpectedFinalName": "policy-managed-subconfig_deploy_<deploy_policy_id>.conf"
  }
}
```

Expected generated name: `policy-managed-subconfig_deploy_<deploy_policy_id>.conf`.

### project_plugin_config_template_to_hosts

```json
{
  "plugin_name": "bk-nodemgr-deploy-policy-probe",
  "config_files_detail": [
    {
      "template_name": "projected-source-subconfig.conf",
      "is_main_config": false
    }
  ],
  "custom_config_context": {
    "ProbeMode": "project-plugin-config-template-to-hosts",
    "Marker": "projected-source-subconfig",
    "Sequence": "006",
    "ExpectedFinalName": "projected-source-subconfig_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id>.conf"
  }
}
```

Expected generated name: `projected-source-subconfig_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id>.conf`.
