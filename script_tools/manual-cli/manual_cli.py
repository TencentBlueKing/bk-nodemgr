#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
manual_cli.py - trigger bk-nodemgr manual agent install via APIGateway and
fetch the manual install command.

Subcommands:
  check        call /api/v3/node/agent/install_check, tell reinstall vs fresh install
  install      call /api/v3/node/agent/install with is_manual=true (auto install_check
               first; reuses bk_host_id of the matched host for reinstall), then
               optionally poll for the manual bootstrap command
  manual-info  query /api/v3/node/workflow/operation/manual/info/get for an existing
               workflow/operation, once or with blocking poll

Auth: app code/secret plus bk_username, passed through the X-Bkapi-Authorization
header (same shape as pkg/thirdparty/apigw/client VirtualUserConfig un mode).

Examples:
  # check only
  python3 manual_cli.py --apigw-url http://bkapi.example.com/api/bk-nodemgr/prod \
      --app-code myapp --app-secret mysecret --bk-username admin \
      check --biz-id 100 --networkunit-id 1 --ip 10.0.0.1

  # trigger manual install and wait for the bootstrap command
  python3 manual_cli.py --apigw-url http://bkapi.example.com/api/bk-nodemgr/prod \
      --app-code myapp --app-secret mysecret --bk-username admin \
      install --biz-id 100 --networkunit-id 1 --ip 10.0.0.1 --os-type linux --wait

  # query manual info of an existing operation
  python3 manual_cli.py ... manual-info --workflow-id wf-xxx --operation-id op-yyy --wait
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

DEFAULT_POLL_INTERVAL = 5
DEFAULT_POLL_TIMEOUT = 300
DEFAULT_HTTP_TIMEOUT = 30

# install_check categories that allow continuing without extra confirmation
CATEGORY_NORMAL_INSTALL = "normal_install"
CATEGORY_REGISTER_AND_INSTALL = "register_to_cmdb_and_install"
CATEGORY_NEED_CONFIRM = "need_confirm"
CATEGORY_ERROR = "error"


class ApiError(Exception):
    """Error returned by the apigw/backend envelope (code != 0) or HTTP layer."""


class AmbiguousOperationError(ApiError):
    """The operation locating conditions matched more than one operation."""


class ApiGwClient:
    """Minimal bk-apigateway client using app code/secret authentication."""

    def __init__(self, base_url, app_code, app_secret, bk_username,
                 http_timeout=DEFAULT_HTTP_TIMEOUT):
        self.base_url = base_url.rstrip("/")
        self.app_code = app_code
        self.app_secret = app_secret
        # the backend requires a non-empty username claim in the apigw jwt
        self.bk_username = bk_username
        self.http_timeout = http_timeout

    def post(self, path, payload):
        url = self.base_url + path
        auth = {"bk_app_code": self.app_code, "bk_app_secret": self.app_secret}
        if self.bk_username:
            auth["bk_username"] = self.bk_username
        headers = {
            "Content-Type": "application/json",
            "X-Bkapi-Authorization": json.dumps(auth),
        }
        req = urllib.request.Request(
            url, data=json.dumps(payload).encode("utf-8"), headers=headers, method="POST"
        )
        try:
            with urllib.request.urlopen(req, timeout=self.http_timeout) as resp:
                body = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as e:
            detail = e.read().decode("utf-8", errors="replace")
            raise ApiError("POST {} http {}: {}".format(path, e.code, detail))
        except urllib.error.URLError as e:
            raise ApiError("POST {} failed: {}".format(path, e.reason))

        code = body.get("code", 0)
        if code != 0:
            message = body.get("message", "")
            permission = body.get("permission") or {}
            apply_url = permission.get("apply_url", "")
            suffix = " (apply url: {})".format(apply_url) if apply_url else ""
            raise ApiError("POST {} code={} message={}{}".format(path, code, message, suffix))
        return body.get("data") or {}


def eprint(*args):
    print(*args, file=sys.stderr)


def build_check_host(args):
    host = {"bk_biz_id": args.biz_id}
    host["bk_host_id"] = args.host_id if args.host_id is not None else -1
    if args.ip:
        host["bk_host_innerip_list"] = args.ip
    if args.ipv6:
        host["bk_host_innerip_v6_list"] = args.ipv6
    if args.networkunit_id is not None:
        host["bk_networkunit_id"] = args.networkunit_id
    return host


def run_install_check(client, args):
    """Call install_check and return the single check result."""
    data = client.post("/api/v3/node/agent/install_check", {"host": [build_check_host(args)]})
    results = data.get("results") or []
    if not results:
        raise ApiError("install_check returned empty results")
    return results[0]


def resolve_install_target(args, result):
    """
    Decide fresh install vs reinstall from an install_check result.

    Returns (bk_host_id, os_type). bk_host_id is None for a fresh install.
    """
    category = result.get("category", "")
    status = result.get("status", "")
    message = result.get("message_zh") or result.get("message_en") or ""
    matched = result.get("matched") or {}

    eprint("install_check: status={} category={} message={}".format(status, category, message))

    if category == CATEGORY_ERROR:
        raise ApiError("install_check failed, cannot install: {} ({})".format(message, status))

    if category == CATEGORY_NEED_CONFIRM and not args.force:
        raise ApiError(
            "install_check requires confirmation: {} ({}); "
            "re-run with --force to reuse the matched host".format(message, status)
        )

    # explicit --host-id always wins over the auto-detected matched host
    if args.host_id is not None and args.host_id >= 0:
        matched_host_id = matched.get("bk_host_id")
        if matched_host_id is not None and matched_host_id >= 0 and matched_host_id != args.host_id:
            raise ApiError(
                "host id conflict: --host-id={} but install_check matched bk_host_id={}; "
                "please check the parameters".format(args.host_id, matched_host_id))
        eprint("reinstall: use user provided bk_host_id={}".format(args.host_id))
        return args.host_id, args.os_type or matched.get("os_type") or "linux"

    host_id = matched.get("bk_host_id")
    if host_id is not None and host_id >= 0:
        eprint("reinstall: matched existing host bk_host_id={} node_role={}".format(
            host_id, matched.get("node_role", "")))
        os_type = args.os_type or matched.get("os_type") or "linux"
        return host_id, os_type

    if category == CATEGORY_NEED_CONFIRM:
        raise ApiError(
            "install_check requires confirmation: {} ({}), but no host could be "
            "identified to reinstall; pass --host-id explicitly".format(message, status))

    eprint("fresh install: no existing host matched")
    return None, args.os_type or "linux"


def build_install_payload(args, host_id, os_type):
    host = {
        "bk_addressing": args.addressing,
        "bk_biz_id": args.biz_id,
        "bk_networkunit_id": args.networkunit_id,
        "os_type": os_type,
        # credentials are not used in manual mode, but the fields are still validated
        "login_ip": args.login_ip or (args.ip[0] if args.ip else args.ipv6[0]),
        "login_port": args.login_port,
        "login_user": args.login_user or ("Administrator" if os_type == "windows" else "root"),
        "login_mode": args.login_mode,
        "re_register": args.re_register,
    }
    if args.ip:
        host["bk_host_innerip"] = args.ip
    if args.ipv6:
        host["bk_host_innerip_v6"] = args.ipv6
    if host_id is not None:
        host["bk_host_id"] = host_id
    return {"host": [host], "is_manual": True}


def find_operation_id(client, workflow_id, ips, ips_v6):
    """Locate the operation by host IPs; returns None when not created yet."""
    conditions = {}
    if ips:
        conditions["bk_host_innerip"] = ips
    if ips_v6:
        conditions["bk_host_innerip_v6"] = ips_v6
    data = client.post(
        "/api/v3/node/workflow/operation/list",
        {"workflow_id": workflow_id,
         "page": {"offset": 0, "limit": 100},
         "exact_include_conditions": conditions},
    )
    operations = data.get("operations") or []
    if len(operations) > 1:
        raise AmbiguousOperationError(
            "conditions {} matched {} operations in workflow_id={}; "
            "please pass --operation-id explicitly".format(conditions, len(operations), workflow_id))
    if not operations:
        return None
    return operations[0].get("operation_id")


def get_manual_commands(client, workflow_id, operation_id):
    data = client.post(
        "/api/v3/node/workflow/operation/manual/info/get",
        {"workflow_id": workflow_id, "operation_id": operation_id},
    )
    return data.get("commands") or []


def wait_manual_commands(client, workflow_id, operation_id, ips, ips_v6, interval, timeout):
    """Poll until the manual bootstrap commands are generated."""
    deadline = time.time() + timeout
    while True:
        # the operation instance is created asynchronously by the workflow engine,
        # wait one interval before the first poll to avoid the transient
        # "operation has no instances" error
        time.sleep(interval)
        try:
            if operation_id is None:
                operation_id = find_operation_id(client, workflow_id, ips, ips_v6)
                if operation_id:
                    eprint("found operation_id={}".format(operation_id))
            if operation_id is not None:
                commands = get_manual_commands(client, workflow_id, operation_id)
                if commands:
                    return operation_id, commands
        except AmbiguousOperationError:
            # ambiguous locating conditions will not fix themselves, fail fast
            raise
        except ApiError as e:
            # operation/instance/private data may not be ready yet, keep polling
            eprint("manual info not ready: {}".format(e))
        if time.time() >= deadline:
            raise ApiError("timed out waiting for manual commands ({}s)".format(timeout))


def print_commands(commands, output=None):
    if output:
        try:
            with open(output, "w", encoding="utf-8") as f:
                for cmd in commands:
                    f.write(cmd.get("command", "") + "\n")
        except OSError as e:
            raise ApiError("failed to write commands to {}: {}".format(output, e))
        eprint("commands saved to {}".format(output))
        return
    for cmd in commands:
        print("# type: {}".format(cmd.get("type", "")))
        print(cmd.get("command", ""))


def cmd_check(client, args):
    result = run_install_check(client, args)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    if result.get("category") == CATEGORY_ERROR:
        return 1
    return 0


def cmd_install(client, args):
    host_id = args.host_id
    os_type = args.os_type
    if not args.skip_check:
        host_id, os_type = resolve_install_target(args, run_install_check(client, args))
    # fresh install: bk_host_id omitted, the server side defaults it to -1
    if os_type is None:
        os_type = "linux"
        eprint("warning: --os-type not specified and no matched host to infer from, "
               "defaulting to linux")

    data = client.post("/api/v3/node/agent/install", build_install_payload(args, host_id, os_type))
    workflow_id = data.get("workflow_id", "")
    eprint("manual install launched, workflow_id={}".format(workflow_id))
    print(json.dumps({"workflow_id": workflow_id}, ensure_ascii=False))

    if not args.wait:
        return 0
    _, commands = wait_manual_commands(
        client, workflow_id, None, args.ip, args.ipv6, args.poll_interval, args.poll_timeout
    )
    print_commands(commands, args.output)
    return 0


def cmd_manual_info(client, args):
    operation_id = args.operation_id
    if args.wait:
        operation_id, commands = wait_manual_commands(
            client, args.workflow_id, operation_id, args.ip, args.ipv6,
            args.poll_interval, args.poll_timeout
        )
    else:
        if operation_id is None:
            operation_id = find_operation_id(client, args.workflow_id, args.ip, args.ipv6)
            if operation_id is None:
                raise ApiError("no operation found for workflow_id={}".format(args.workflow_id))
        commands = get_manual_commands(client, args.workflow_id, operation_id)
    eprint("workflow_id={} operation_id={}".format(args.workflow_id, operation_id))
    print_commands(commands, args.output)
    return 0


def add_common_args(parser):
    parser.add_argument("--apigw-url", required=True,
                        help="gateway stage root, e.g. http://bkapi.example.com/api/bk-nodemgr/prod")
    parser.add_argument("--app-code", required=True, help="bk_app_code")
    # prefer the env var so the secret does not leak into shell history / ps
    parser.add_argument("--app-secret", default=os.environ.get("BK_APP_SECRET"),
                        help="bk_app_secret; defaults to the BK_APP_SECRET env var")
    parser.add_argument("--bk-username", required=True,
                        help="tenant-scoped bk_username passed to apigw (e.g. admin, or "
                             "bk_admin in multi-tenant mode); the backend requires a "
                             "non-empty username claim in the apigw jwt")
    parser.add_argument("--http-timeout", type=int, default=DEFAULT_HTTP_TIMEOUT,
                        help="single http request timeout in seconds")


def add_host_args(parser, need_networkunit=True):
    parser.add_argument("--biz-id", type=int, required=True, help="bk_biz_id")
    parser.add_argument("--networkunit-id", type=int, required=need_networkunit,
                        help="bk_networkunit_id (管控单元ID)")
    parser.add_argument("--host-id", type=int, default=None,
                        help="bk_host_id; omit/-1 means a new host")
    parser.add_argument("--ip", action="append", default=[],
                        help="inner IPv4, repeatable")
    parser.add_argument("--ipv6", action="append", default=[],
                        help="inner IPv6, repeatable")
    parser.add_argument("--os-type", choices=["linux", "windows", "darwin"], default=None,
                        help="os type; defaults to the matched host os, then linux")


def add_poll_args(parser):
    parser.add_argument("--wait", action="store_true",
                        help="block and poll until the manual command is generated")
    parser.add_argument("--poll-interval", type=int, default=DEFAULT_POLL_INTERVAL)
    parser.add_argument("--poll-timeout", type=int, default=DEFAULT_POLL_TIMEOUT)
    parser.add_argument("--output", default=None,
                        help="write the manual commands to this file instead of stdout")


def build_parser():
    parser = argparse.ArgumentParser(
        description="bk-nodemgr manual install CLI via APIGateway")
    add_common_args(parser)
    sub = parser.add_subparsers(dest="command", required=True)

    p_check = sub.add_parser("check", help="install check: reinstall vs fresh install")
    add_host_args(p_check, need_networkunit=False)

    p_install = sub.add_parser("install", help="trigger a manual install")
    add_host_args(p_install)
    p_install.add_argument("--addressing", choices=["static", "dynamic"], default="static")
    p_install.add_argument("--login-ip", default=None, help="defaults to the first --ip")
    p_install.add_argument("--login-port", type=int, default=22)
    p_install.add_argument("--login-user", default=None,
                           help="defaults to root (linux/darwin) or Administrator (windows)")
    p_install.add_argument("--login-mode", choices=["password_vault", "password", "keyfile"],
                           default="password_vault",
                           help="not used in manual mode, kept for validation only")
    p_install.add_argument("--re-register", action="store_true",
                           help="re-register the host when reinstalling")
    p_install.add_argument("--skip-check", action="store_true",
                           help="skip install_check and install directly")
    p_install.add_argument("--force", action="store_true",
                           help="continue when install_check asks for confirmation")
    add_poll_args(p_install)

    p_info = sub.add_parser("manual-info", help="get the manual bootstrap command")
    p_info.add_argument("--workflow-id", required=True)
    p_info.add_argument("--operation-id", default=None,
                        help="omit to locate the operation by --ip/--ipv6")
    p_info.add_argument("--ip", action="append", default=[],
                        help="inner IPv4 used to locate the operation, repeatable")
    p_info.add_argument("--ipv6", action="append", default=[],
                        help="inner IPv6 used to locate the operation, repeatable")
    add_poll_args(p_info)

    return parser


def main():
    parser = build_parser()
    args = parser.parse_args()

    if args.command in ("check", "install") and not args.ip and not args.ipv6:
        parser.error("--ip or --ipv6 is required")
    if args.command == "manual-info" and args.operation_id is None \
            and not args.ip and not args.ipv6:
        parser.error("--operation-id or at least one of --ip/--ipv6 is required")
    if not args.app_secret:
        parser.error("--app-secret is required (or set the BK_APP_SECRET env var)")

    client = ApiGwClient(args.apigw_url, args.app_code, args.app_secret, args.bk_username,
                         args.http_timeout)
    handlers = {"check": cmd_check, "install": cmd_install, "manual-info": cmd_manual_info}
    try:
        return handlers[args.command](client, args)
    except ApiError as e:
        eprint("error: {}".format(e))
        return 1


if __name__ == "__main__":
    sys.exit(main())
