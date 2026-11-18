# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making
蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
Copyright (C) Tencent. All rights reserved.
Licensed under the MIT License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the
specific language governing permissions and limitations under the License.

We undertake not to change the open source license (MIT license) applicable

to the current version of the project delivered to anyone in the future.
"""

import hashlib
import json
import re
import shutil
import sys
import typing
from datetime import datetime, timezone
from pathlib import Path

from django.core.exceptions import FieldDoesNotExist
from django.core.management.base import BaseCommand

from apps.node_man.models import (
    Host,
    Packages,
    PluginConfigTemplate,
    ProcControl,
    ProcessStatus,
    Subscription,
    SubscriptionStep,
)

FALLBACK_TENANT_ID = "default"
OUTPUT_DIR = "/tmp/nodeman_export"
SELECTED_FILE = "import_selected.js"
BATCH_SIZE = 100
MONITOR_CHECK_SECS = 5
RESOURCE_LIMIT_PERCENT = 10.0
PROCESS_GENERATION = 2

# 与 pkg/format/platform 的 StandardOSMap 对齐，覆盖 nodeman 主机可能出现的系统类型
OS_MAP = {
    "linux": "linux",
    "windows": "windows",
    "win": "windows",
    "win32": "windows",
    "win64": "windows",
    "aix": "aix",
    "aix6": "aix6",
    "aix7": "aix7",
    "solaris": "solaris",
    "darwin": "darwin",
    "macos": "darwin",
    "freebsd": "freebsd",
    "ubuntu": "linux",
    "debian": "linux",
    "centos": "linux",
    "rhel": "linux",
    "redhat": "linux",
    "alpine": "linux",
    "wsl": "linux",
    "cygwin": "windows",
    "msys": "windows",
    "msys2": "windows",
    "mingw": "windows",
}

# 与 pkg/format/platform 的 StandardArchMap 对齐，覆盖 nodeman 主机可能出现的 CPU 架构
ARCH_MAP = {
    "x86_64": "amd64",
    "x86-64": "amd64",
    "amd64": "amd64",
    "x64": "amd64",
    "aarch64": "arm64",
    "arm64": "arm64",
    "armv8": "arm64",
    "arm": "arm",
    "armv5": "arm",
    "armv6": "arm",
    "armv7": "arm",
    "x86": "386",
    "386": "386",
    "i386": "386",
    "i486": "386",
    "i586": "386",
    "i686": "386",
    "ia32": "386",
    "ppc64": "ppc64",
    "ppc64le": "ppc64le",
    "s390x": "s390x",
    "mips64": "mips64",
    "mips64le": "mips64le",
    "loong64": "loong64",
    "riscv64": "riscv64",
}

COMMAND_FIELDS = ("start_cmd", "stop_cmd", "restart_cmd", "reload_cmd", "debug_cmd", "kill_cmd", "version_cmd", "health_cmd")
DATE_KEY_SUFFIXES = ("last_sync_at", "basic.created_at", "basic.updated_at")


def log_to_stderr(message: str) -> None:
    """输出迁移日志到 stderr。"""
    print(message, file=sys.stderr)


def normalize_os(os_name: typing.Optional[str]) -> typing.Optional[str]:
    """将 nodeman 操作系统名称转换为 nodemgr 标准值。"""
    return OS_MAP.get((os_name or "").strip().lower())


def normalize_arch(arch: typing.Optional[str]) -> typing.Optional[str]:
    """将 nodeman CPU 架构转换为 nodemgr 标准值。"""
    normalized = (arch or "").strip().lower()
    if normalized in ARCH_MAP:
        return ARCH_MAP[normalized]
    if normalized.startswith("armv"):
        match = re.match(r"armv(\d+)", normalized)
        if match and int(match.group(1)) >= 8:
            return "arm64"
        return "arm"
    return None


def json_column(value: typing.Any, default: typing.Any) -> typing.Any:
    """解析数据库中的 JSON 字段，空值返回调用方指定的默认值。"""
    if value is None or value == "":
        return default
    if isinstance(value, (list, dict)):
        return value
    if isinstance(value, bytes):
        value = value.decode("utf-8")
    return json.loads(value)


def tenant_value_fields(model: typing.Any) -> typing.Tuple[str, ...]:
    """返回 tenant_id 字段名；nodeman 表没有该字段时返回空元组，按默认租户导出。"""
    try:
        model._meta.get_field("tenant_id")
    except FieldDoesNotExist:
        return ()
    return ("tenant_id",)


class ExportHandler(object):
    """nodeman 主插件进程导出处理器。"""

    PROCESS_FILTER_FIELDS = ("host_id", "name")
    CONFIG_FILTER_FIELDS = ("name", "process_name", "host_id", "is_main_config")

    def __init__(self) -> None:
        """初始化导出时间戳。"""
        self.now_iso = self.get_now_iso()

    @staticmethod
    def get_now_iso() -> str:
        """生成带毫秒精度的 UTC ISO 时间字符串。"""
        # 兼容 Python 3.6，使用 timezone.utc；datetime.UTC 需要 Python 3.11
        return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%f")[:-3] + "Z"

    @staticmethod
    def select_config_template(
        templates: typing.List[typing.Dict[str, typing.Any]], row: typing.Dict[str, typing.Any]
    ) -> typing.Optional[typing.Dict[str, typing.Any]]:
        """选择主配置模板：优先精确匹配插件版本，允许 * 通配，再按正式版本和模板 ID 选择。"""
        candidates = [
            template for template in templates
            if template["os"] == row["os_type"].lower() and template["cpu_arch"] == row["cpu_arch"]
            and (row["version"] == "" or template["plugin_version"] == row["version"] or template["plugin_version"] == "*")
        ]
        if not candidates:
            return None
        return max(candidates, key=lambda template: (
            row["version"] != "" and template["plugin_version"] == row["version"], template["is_release_version"], template["id"]))

    @staticmethod
    def safe_path_component(value: typing.Optional[str]) -> str:
        """将租户ID转换为合法目录名。"""
        value = (value or "").strip()
        if not value:
            return FALLBACK_TENANT_ID
        return re.sub(r"[^A-Za-z0-9._-]", "_", value)

    @staticmethod
    def trim_bin_suffix(path: str) -> str:
        """去掉路径末尾的 bin 目录并返回安装根目录。"""
        separator = "\\" if "\\" in path else "/"
        suffix = separator + "bin"
        if path.endswith(suffix):
            return path[: -len(suffix)]
        return path

    @staticmethod
    def normalize_command(command: typing.Optional[str]) -> str:
        """把以 ./ 开头的命令改写为 bin/ 相对路径。"""
        command = (command or "").strip()
        if command.startswith("./"):
            return "bin/" + command[2:]
        return command

    @classmethod
    def config_path(cls, setup_path: str, config_name: str) -> str:
        """根据安装路径和配置文件名生成主配置的完整路径。"""
        if not config_name:
            return ""
        if not setup_path:
            return config_name
        separator = "\\" if "\\" in setup_path else "/"
        base = cls.trim_bin_suffix(setup_path).rstrip(separator)
        return base + separator + "etc" + separator + config_name

    @staticmethod
    def matching_step_context(
        params: typing.Dict[str, typing.Any], os_type: str, cpu_arch: str
    ) -> typing.Dict[str, typing.Any]:
        """选择与主机 OS/CPU 架构匹配的订阅渲染上下文。"""
        details = params.get("details") or []
        target_os = os_type.lower()
        target_arch = normalize_arch(cpu_arch)

        for detail in details:
            if (detail.get("os", "").lower() == target_os and normalize_arch(detail.get("cpu_arch", "")) == target_arch
                    and target_arch is not None):
                return detail.get("context") or {}
        for detail in details:
            if detail.get("os", "").lower() == target_os and not detail.get("cpu_arch"):
                return detail.get("context") or {}
        return {}

    @staticmethod
    def js_literal(value: typing.Any, key_path: str = "") -> str:
        """将 Python 对象转换为 mongosh 可执行的字面量，日期字段保持 BSON Date 类型。"""
        if isinstance(value, dict):
            items = []
            for key in sorted(value):
                child_path = key_path + "." + key if key_path else key
                items.append(f"{json.dumps(key)}: {ExportHandler.js_literal(value[key], child_path)}")
            return "{" + ", ".join(items) + "}"
        if isinstance(value, list):
            return "[" + ", ".join(ExportHandler.js_literal(item) for item in value) + "]"
        if isinstance(value, bool):
            return "true" if value else "false"
        if isinstance(value, int):
            # mongosh 已废弃数字入参，使用字符串避免精度丢失
            return f'NumberLong("{value}")'
        if isinstance(value, str) and key_path.endswith(DATE_KEY_SUFFIXES):
            return f"new Date({json.dumps(value)})"
        encoded = json.dumps(value, ensure_ascii=False)
        # U+2028/U+2029 是合法 JSON，但在 ES2019 之前不能直接出现在 JS 字符串中
        return encoded.replace("\u2028", "\\u2028").replace("\u2029", "\\u2029")

    @staticmethod
    def render_update_one(filter_doc: typing.Dict[str, typing.Any], data_doc: typing.Dict[str, typing.Any], now_iso: str) -> str:
        """生成一条 mongosh upsert 更新操作。"""
        update = {
            "$set": {"basic.is_deleted": False, "basic.updated_at": now_iso, "data": data_doc},
            "$setOnInsert": {"basic.created_at": now_iso},
        }
        return f"{{updateOne: {{filter: {ExportHandler.js_literal(filter_doc)}, update: {ExportHandler.js_literal(update)}, upsert: true}}}}"

    @classmethod
    def write_bulk_batches(
        cls,
        handle: typing.TextIO,
        collection: str,
        items: typing.List[typing.Dict[str, typing.Any]],
        filter_fields: typing.Tuple[str, ...],
        batch_size: int,
        now_iso: str,
    ) -> None:
        """按批次写入 bulkWrite 语句，避免单次操作过大。"""
        for start in range(0, len(items), batch_size):
            batch = items[start : start + batch_size]
            rendered_operations = []
            for item in batch:
                filter_doc = {"basic.is_deleted": False}
                for field in filter_fields:
                    filter_doc["data." + field] = item[field]
                rendered_operations.append(cls.render_update_one(filter_doc, item, now_iso))
            operations = ",\n".join(rendered_operations)
            handle.write(f"db.{collection}.bulkWrite([\n{operations}\n], {{ordered: false}});\n")

    def export(self) -> None:
        """执行完整导出流程并写出 mongosh 导入文件。"""
        rows = self.fetch_rows()
        templates = self.fetch_config_templates()
        controls = self.fetch_proc_controls()
        plan = self.build_plan(rows, templates, controls)
        self.print_summary(plan)

        output_dir = Path(OUTPUT_DIR)
        if output_dir.exists():
            shutil.rmtree(str(output_dir))

        paths = self.write_js(plan)
        log_to_stderr(f"wrote {len(paths)} files under {OUTPUT_DIR} ({plan['process_count']} processes, {plan['config_count']} configs)")
        log_to_stderr(
            "import inside the bk-nodemgr pod: ./mongo.sh login, then "
            f'load("<dir>/<tenant>/<biz_id>/<plugin>.js") or load("<dir>/{SELECTED_FILE}")'
        )

    def fetch_subscriptions(self) -> typing.Dict[int, typing.Dict[str, typing.Any]]:
        """读取主插件订阅并按订阅 ID 建立索引。"""
        subscriptions = {}
        records = Subscription.objects.filter(is_main=True, is_deleted=False).values(
            "id", "plugin_name", "object_type", "node_type", "nodes", *tenant_value_fields(Subscription)
        )
        for record in records:
            plugin = record["plugin_name"] or ""
            if not plugin:
                continue
            subscriptions[record["id"]] = {
                "plugin": plugin, "tenant": record.get("tenant_id") or "", "object_type": record["object_type"] or "",
                "node_type": record["node_type"] or "", "nodes": json_column(record["nodes"], []),
            }
        return subscriptions

    def fetch_step_params(self, subscription_ids: typing.List[int]) -> typing.Dict[typing.Tuple[int, str], typing.Any]:
        """读取订阅步骤参数并按订阅和步骤建立索引。"""
        params = {}
        records = SubscriptionStep.objects.filter(subscription_id__in=subscription_ids).values(
            "subscription_id", "step_id", "params"
        )
        for record in records:
            params[(record["subscription_id"], record["step_id"] or "")] = json_column(record["params"], {})
        return params

    def fetch_rows(self) -> typing.List[typing.Dict[str, typing.Any]]:
        """组装进程导出所需的订阅、主机和步骤参数数据。"""
        subscriptions = self.fetch_subscriptions()
        if not subscriptions:
            return []

        subscription_ids = sorted(subscriptions)
        step_params = self.fetch_step_params(subscription_ids)
        processes = ProcessStatus.objects.filter(
            source_type=ProcessStatus.SourceType.DEFAULT, is_latest=True, source_id__isnull=False,
            source_id__in=[str(subscription_id) for subscription_id in subscription_ids],
        ).values(
            "bk_host_id", "name", "status", "version", "is_auto", "listen_ip", "listen_port", "setup_path",
            "log_path", "pid_path", "configs", "source_id", *tenant_value_fields(ProcessStatus),
        )

        process_records = list(processes)
        host_ids = {process["bk_host_id"] for process in process_records if process["bk_host_id"]}
        host_records = Host.objects.filter(bk_host_id__in=host_ids).values(
            "bk_host_id", "bk_biz_id", "os_type", "cpu_arch", "bk_agent_id", *tenant_value_fields(Host)
        )
        hosts = {record["bk_host_id"]: record for record in host_records}

        rows = []
        for process in process_records:
            try:
                subscription_id = int(process["source_id"])
            except (TypeError, ValueError):
                continue
            subscription = subscriptions.get(subscription_id)
            if subscription is None or subscription["plugin"] != (process["name"] or ""):
                continue

            host = hosts.get(process["bk_host_id"])
            if host is None:
                continue

            tenant = subscription["tenant"] or process.get("tenant_id") or host.get("tenant_id") or FALLBACK_TENANT_ID
            params = step_params.get((subscription_id, subscription["plugin"]), {})
            rows.append({
                "host_id": int(process["bk_host_id"]), "biz_id": int(host["bk_biz_id"]), "os_type": (host["os_type"] or "").lower(),
                "cpu_arch": host["cpu_arch"] or "", "agent_id": host["bk_agent_id"] or "", "status": process["status"] or "",
                "version": process["version"] or "", "is_auto": process["is_auto"] or "", "listen_ip": process["listen_ip"] or "",
                "listen_port": int(process["listen_port"] or 0), "setup_path": process["setup_path"] or "",
                "log_path": process["log_path"] or "", "pid_path": process["pid_path"] or "", "configs": json_column(process["configs"], []),
                "node_type": subscription["node_type"], "object_type": subscription["object_type"], "scope": subscription["nodes"],
                "params": params, "plugin": subscription["plugin"], "tenant": tenant,
            })
        rows.sort(key=lambda row: (row["biz_id"], row["plugin"], row["host_id"]))
        return rows

    def fetch_config_templates(self) -> typing.Dict[str, typing.List[typing.Dict[str, typing.Any]]]:
        """读取主配置模板并按插件名称分组。"""
        templates = {}
        records = PluginConfigTemplate.objects.filter(is_main=True).order_by("id").values(
            "id", "plugin_name", "name", "file_path", "os", "cpu_arch", "plugin_version", "is_release_version"
        )
        for record in records:
            templates.setdefault(record["plugin_name"] or "", []).append({
                "id": record["id"], "name": record["name"] or "", "file_path": record["file_path"] or "",
                "os": (record["os"] or "").lower(), "cpu_arch": record["cpu_arch"] or "",
                "plugin_version": record["plugin_version"] or "", "is_release_version": bool(record["is_release_version"]),
            })
        return templates

    def fetch_proc_controls(self) -> typing.Dict[typing.Tuple[str, str, str, str, str], typing.Dict[str, str]]:
        """读取插件包进程控制命令并按包信息建立索引。"""
        controls_by_package = {}
        records = ProcControl.objects.values("plugin_package_id", *COMMAND_FIELDS)
        for record in records:
            controls_by_package[record["plugin_package_id"]] = {field: record[field] or "" for field in COMMAND_FIELDS}

        controls = {}
        packages = Packages.objects.order_by("id").values(
            "id", "project", "version", "os", "cpu_arch", *tenant_value_fields(Packages))
        for record in packages:
            commands = controls_by_package.get(record["id"])
            if commands is None:
                continue
            # 按包ID升序处理，后续记录覆盖前面的记录，最终取ID最大的包
            controls[
                (record.get("tenant_id") or FALLBACK_TENANT_ID, record["project"] or "", record["version"] or "",
                 (record["os"] or "").lower(), record["cpu_arch"] or "")
            ] = commands
        return controls

    def build_process(
        self,
        row: typing.Dict[str, typing.Any],
        template: typing.Optional[typing.Dict[str, typing.Any]],
        commands: typing.Dict[str, str],
    ) -> typing.Dict[str, typing.Any]:
        """根据源进程记录构造 process_v2 文档。"""
        os_name = normalize_os(row["os_type"])
        arch_name = normalize_arch(row["cpu_arch"])
        if os_name is None or arch_name is None:
            os_name, arch_name = "unknown", "unknown"

        is_auto = row["is_auto"].upper() == "AUTO"
        config_name = template["name"] if template else ""
        # 与 nodemgr 保持一致：操作用户只由操作系统决定，unix 为 root，windows 为 system
        operate_user = "system" if os_name == "windows" else "root"
        identity_name = row["plugin"] + ".exe" if os_name == "windows" else row["plugin"]
        # nodemgr 只同步 running/stopped，其他源状态统一写为 stopped，避免导入记录无法被接管
        process_status = "running" if row["status"].strip().upper() == "RUNNING" else "stopped"

        return {
            "tenant_id": row["tenant"], "host_id": row["host_id"], "biz_id": row["biz_id"], "name": row["plugin"],
            "group": "default", "plugin_pkg_name": row["plugin"], "generation": PROCESS_GENERATION,
            "platform": {"os": os_name, "arch": arch_name}, "bind_ip": row["listen_ip"], "bind_port": row["listen_port"],
            "info": {
                "pid": -1, "version": row["version"], "agent_id": row["agent_id"], "trusteeship": is_auto,
                "status": process_status, "last_sync_at": self.now_iso,
            },
            "identity": {
                "name": identity_name, "setup_path": self.trim_bin_suffix(row["setup_path"]), "pid_path": row["pid_path"],
                "config_path": self.config_path(row["setup_path"], config_name), "log_path": row["log_path"], "user": operate_user,
            },
            "controller": {field: self.normalize_command(commands.get(field, "")) for field in COMMAND_FIELDS},
            "resource": {"cpu_limit_percent": RESOURCE_LIMIT_PERCENT, "mem_limit_percent": RESOURCE_LIMIT_PERCENT},
            "monitor_policy": {
                "restart_type": "auto" if is_auto else "manual", "start_check_secs": MONITOR_CHECK_SECS,
                "stop_check_secs": MONITOR_CHECK_SECS, "op_timeout_secs": MONITOR_CHECK_SECS,
            },
        }

    def build_configs(
        self, row: typing.Dict[str, typing.Any], template: typing.Optional[typing.Dict[str, typing.Any]]
    ) -> typing.List[typing.Dict[str, typing.Any]]:
        """根据主配置模板和源配置内容构造 process_v2_config 文档。"""
        if not template or not template["name"]:
            return []

        context = self.matching_step_context(row["params"], row["os_type"], row["cpu_arch"])
        # nodemgr 会将 custom context 的顶层字段平铺到模板变量空间
        custom_context = context or None

        configs = []
        for config in row["configs"]:
            if config.get("name") != template["name"]:
                continue
            content = config.get("content") or ""
            md5_value = config.get("md5") or hashlib.md5(content.encode("utf-8")).hexdigest()
            configs.append({
                "name": config["name"], "template_name": config["name"], "process_name": row["plugin"],
                "host_id": row["host_id"], "set": "", "is_main_config": True, "content": content, "md5": md5_value,
                "file_path": template["file_path"], "custom_config_context": custom_context,
            })
        return configs

    def build_plan(
        self,
        rows: typing.List[typing.Dict[str, typing.Any]],
        templates: typing.Dict[str, typing.List[typing.Dict[str, typing.Any]]],
        controls: typing.Dict[typing.Tuple[str, str, str, str, str], typing.Dict[str, str]],
    ) -> typing.Dict[str, typing.Any]:
        """过滤无效记录并按租户、业务和插件生成导出计划。"""
        groups = {}
        skipped = 0
        duplicates = 0

        for row in rows:
            if row["host_id"] <= 0:
                skipped += 1
                continue
            if row["node_type"].upper() == "INSTANCE" and row["object_type"].upper() == "HOST":
                if not any(node.get("bk_host_id") == row["host_id"] for node in row["scope"]):
                    skipped += 1
                    continue

            plugin = row["plugin"]
            if not plugin:
                skipped += 1
                continue

            tenant = row["tenant"] or FALLBACK_TENANT_ID
            group_key = (tenant, row["biz_id"], plugin)
            group = groups.setdefault(group_key, {"processes": {}, "configs": {}})

            template = self.select_config_template(templates.get(plugin, []), row)
            control_key = (tenant, plugin, row["version"], row["os_type"].lower(), row["cpu_arch"])
            commands = controls.get(control_key) or controls.get((FALLBACK_TENANT_ID,) + control_key[1:], {})

            if row["host_id"] in group["processes"]:
                duplicates += 1
            group["processes"][row["host_id"]] = self.build_process(row, template, commands)
            group["configs"][row["host_id"]] = self.build_configs(row, template)

        ordered_groups = []
        for tenant, biz_id, plugin in sorted(groups):
            group = groups[(tenant, biz_id, plugin)]
            ordered_hosts = sorted(group["processes"])
            ordered_groups.append({
                "tenant": tenant, "biz_id": biz_id, "plugin": plugin,
                "processes": [group["processes"][host] for host in ordered_hosts],
                "configs": [config for host in ordered_hosts for config in sorted(group["configs"][host], key=lambda item: item["name"])],
            })

        return {
            "groups": ordered_groups,
            "process_count": sum(len(group["processes"]) for group in ordered_groups),
            "config_count": sum(len(group["configs"]) for group in ordered_groups),
            "skipped": skipped,
            "duplicates": duplicates,
            "total": len(rows),
        }

    def write_group_body(self, handle: typing.TextIO, group: typing.Dict[str, typing.Any]) -> None:
        """把一个导出分组的进程和配置写成 mongosh 语句。"""
        handle.write(
            f"// tenant: {group['tenant']}, biz: {group['biz_id']}, plugin: {group['plugin']}, "
            f"processes: {len(group['processes'])}, configs: {len(group['configs'])}\n"
        )
        self.write_bulk_batches(
            handle, "process_v2_" + group["tenant"], group["processes"],
            self.PROCESS_FILTER_FIELDS, BATCH_SIZE, self.now_iso
        )
        if group["configs"]:
            self.write_bulk_batches(
                handle, "process_v2_config_" + group["tenant"], group["configs"],
                self.CONFIG_FILTER_FIELDS, BATCH_SIZE, self.now_iso
            )

    def write_group_files(self, plan: typing.Dict[str, typing.Any]) -> typing.List[Path]:
        """按租户、业务和插件生成独立的 JS 导入文件。"""
        output_dir = Path(OUTPUT_DIR)
        written = []
        for group in plan["groups"]:
            tenant_dir = self.safe_path_component(group["tenant"])
            biz_dir = output_dir / tenant_dir / str(group["biz_id"])
            biz_dir.mkdir(parents=True, exist_ok=True)
            path = biz_dir / f"{group['plugin']}.js"
            with path.open("w") as handle:
                title = f"tenant: {group['tenant']}, biz: {group['biz_id']}, plugin: {group['plugin']}"
                handle.write(f"// Generated by export_nodeman_process.py at {self.now_iso}\n")
                handle.write(f"// {title}, tenant: {group['tenant']}\n")
                handle.write("// Uses the currently connected MongoDB database.\n")
                self.write_group_body(handle, group)
            written.append(path)
        return written

    def write_selection_file(self, plan: typing.Dict[str, typing.Any]) -> Path:
        """生成批量导入清单，运维可按需注释掉不需要导入的行。"""
        output_dir = Path(OUTPUT_DIR)
        output_dir.mkdir(parents=True, exist_ok=True)
        path = output_dir / SELECTED_FILE
        with path.open("w") as handle:
            handle.write(f"// Generated by export_nodeman_process.py at {self.now_iso}\n")
            handle.write("// Uses the currently connected MongoDB database.\n")
            handle.write("// Comment out any line below to skip that file, then run:\n")
            handle.write(f'//   load("<path to this {SELECTED_FILE}>")\n')
            handle.write("var files = [\n")
            for group in plan["groups"]:
                handle.write(f'  "{self.safe_path_component(group["tenant"])}/{group["biz_id"]}/{group["plugin"]}.js",\n')
            handle.write("];\n")
            handle.write('files.forEach(function (file) { load(__dirname + "/" + file); });\n')
            handle.write(f'print("imported {len(plan["groups"])} files");\n')
        return path

    def write_js(self, plan: typing.Dict[str, typing.Any]) -> typing.List[Path]:
        """按业务/插件生成导入文件，并生成批量导入清单。"""
        paths = self.write_group_files(plan)
        if paths:
            paths.append(self.write_selection_file(plan))
        return paths

    def print_summary(self, plan: typing.Dict[str, typing.Any]) -> None:
        """输出导出计划摘要和分组统计。"""
        log_to_stderr(f"output dir: {OUTPUT_DIR}")
        log_to_stderr(f"source rows: {plan['total']}, skipped: {plan['skipped']}, duplicate hosts: {plan['duplicates']}")
        log_to_stderr(
            f"planned groups: {len(plan['groups'])}, processes: {plan['process_count']}, configs: {plan['config_count']}"
        )
        for group in plan["groups"]:
            log_to_stderr(
                f"  tenant={group['tenant']} biz_id={group['biz_id']} plugin={group['plugin']} "
                f"processes={len(group['processes'])} configs={len(group['configs'])}"
            )


class Command(BaseCommand):
    help = "导出 nodeman 主插件进程记录为 mongosh JS 文件"

    def handle(self, *args, **options):
        """Django 管理命令入口，执行插件进程导出。"""
        ExportHandler().export()
