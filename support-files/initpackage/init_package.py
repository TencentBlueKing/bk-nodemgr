import argparse
import json
import os
import subprocess
import sys
import uuid


# =================== utils ===================


def generate_id(tag):
    return tag + ":" + str(uuid.uuid4()).replace("-", "")


DEFAULT_CONFIG_FILES = (
    "/bk-nodemgr/etc/file_conf.yaml",
    "/bk-nodemgr/etc/bk-nodemgr-file.yml",
)
DEFAULT_PACKAGES_DIR = "/bk-nodemgr/file/packages"
TASK_ORDER = (
    "cert",
    "bintool",
    "plugin_bintool",
    "agent",
    "proxy",
    "server",
    "plugin_v2",
    "external_plugin_v2",
    "plugin_v3",
)


def load_file_config(config_file=None):
    """Load the first available File service configuration."""
    try:
        import yaml
    except ImportError as error:
        return {}, None, [f"PyYAML is unavailable: {error}"]

    config_files = (config_file,) if config_file else DEFAULT_CONFIG_FILES
    errors = []

    for path in config_files:
        if not os.path.isfile(path):
            continue

        try:
            with open(path, "r", encoding="utf-8") as file:
                config = yaml.safe_load(file) or {}
            if not isinstance(config, dict):
                raise ValueError("top-level YAML value must be a mapping")
            return config, path, errors
        except (OSError, ValueError, yaml.YAMLError) as error:
            errors.append(f"{path}: {error}")

    return {}, None, errors


def get_nested_value(mapping, *keys):
    """Return a nested mapping value, or None when it is unavailable."""
    value = mapping
    for key in keys:
        if not isinstance(value, dict):
            return None
        value = value.get(key)
    return value


def config_int(value, fallback):
    """Return an integer configuration value, using fallback for invalid values."""
    try:
        return int(value)
    except (TypeError, ValueError):
        return fallback


def resolve_runtime_config(args, config):
    """Resolve CLI values, File configuration values, and safe defaults."""
    basic_server = config.get("basicServer", {})
    auth_identity = get_nested_value(basic_server, "authIdentity")
    config_jwt_key = ""
    if auth_identity != "none":
        config_jwt_key = (
            get_nested_value(basic_server, "jwtServerConfig", "symmetricKey") or ""
        )

    tenant_mode = config.get("tenantMode")
    return {
        "host": args.host if args.host is not None else "localhost",
        "port": (
            args.port
            if args.port is not None
            else config_int(get_nested_value(basic_server, "port"), 28202)
        ),
        "jwt_key": args.jwt_key if args.jwt_key is not None else config_jwt_key,
        "expire_hours": (
            args.expire_hours
            if args.expire_hours is not None
            else config_int(
                get_nested_value(
                    basic_server, "jwtServerConfig", "tokenExpirationHour"
                ),
                24,
            )
        ),
        "bk_username": args.bk_username if args.bk_username is not None else "admin",
        "login_name": args.login_name if args.login_name is not None else "admin",
        "tenant_id": (
            args.tenant_id
            if args.tenant_id is not None
            else ("system" if tenant_mode == "multiple" else "default")
        ),
    }


PACKAGE_DIRECTORY_TYPES = {
    "cert": "cert",
    "agent": "agent",
    "proxy": "proxy",
    "server": "server",
    "plugin-v2": "plugin_v2",
    "external-plugin-v2": "external_plugin_v2",
    "plugin-v3": "plugin_v3",
}


def discover_packages(packages_dir):
    """Discover regular package files from the conventional package directories."""
    discovered = {package_type: [] for package_type in TASK_ORDER}

    for directory_name, package_type in PACKAGE_DIRECTORY_TYPES.items():
        directory = os.path.join(packages_dir, directory_name)
        if not os.path.isdir(directory):
            continue
        for filename in sorted(os.listdir(directory)):
            file_path = os.path.join(directory, filename)
            if os.path.isfile(file_path):
                discovered[package_type].append(file_path)

    bintool_directory = os.path.join(packages_dir, "bintool")
    if os.path.isdir(bintool_directory):
        for filename in sorted(os.listdir(bintool_directory)):
            file_path = os.path.join(bintool_directory, filename)
            if not os.path.isfile(file_path):
                continue
            package_type = (
                "plugin_bintool" if filename.startswith("plugin_bintool") else "bintool"
            )
            discovered[package_type].append(file_path)

    return discovered


def generate_jwt_token(key, expire_hours=24, bk_username=None, login_name=None):
    script_dir = os.path.dirname(os.path.abspath(__file__))
    binary_name = "jwt-generator"
    binary_path = os.path.join(script_dir, binary_name)

    if not os.path.exists(binary_path):
        return False, f"jwt-generator binary not found: {binary_path}"

    try:
        result = subprocess.run(
            [
                binary_path,
                "-k",
                key,
                "-e",
                str(expire_hours),
                "-u",
                bk_username,
                "-l",
                login_name,
            ],
            capture_output=True,
            text=True,
            timeout=10,
        )

        if result.returncode != 0:
            error_msg = result.stderr.strip() if result.stderr else "unknown error"
            return False, f"jwt-generator failed: {error_msg}"

        token = result.stdout.strip()
        if not token:
            return False, "jwt-generator returned empty token"

        # debug: check token format
        if token.count(".") != 2:
            return (
                False,
                f"invalid JWT token format (expected 3 segments, got {token.count('.') + 1}): {token[:100]}",
            )

        return True, token

    except subprocess.TimeoutExpired:
        return False, "jwt-generator execution timeout"
    except Exception as e:
        return False, f"failed to execute jwt-generator: {str(e)}"


# =================== http request ===================


PACKAGE_REQUEST_TIMEOUT = 30 * 60


def http_request(
    url,
    method="GET",
    headers=None,
    params=None,
    json_data=None,
    form_data=None,
    files=None,
    timeout=PACKAGE_REQUEST_TIMEOUT,
):
    import requests

    try:
        files_to_upload = None
        files_to_close = []

        if files:
            files_to_upload = {}
            for field_name, file_info in files.items():
                if isinstance(file_info, str):
                    f = open(file_info, "rb")
                    files_to_close.append(f)
                    filename = file_info.split("/")[-1]
                    files_to_upload[field_name] = (filename, f)
                elif isinstance(file_info, tuple):
                    files_to_upload[field_name] = file_info
                else:
                    files_to_upload[field_name] = file_info

        resp = requests.request(
            method=method.upper(),
            url=url,
            headers=headers,
            params=params,
            json=json_data if json_data is not None else None,
            data=form_data if files_to_upload else None,
            files=files_to_upload,
            timeout=timeout,
            verify=False,
        )

        for f in files_to_close:
            f.close()

        if resp.status_code != 200:
            content = (
                resp.content.decode("utf-8", errors="ignore") if resp.content else ""
            )
            return False, {
                "error": f"HTTP {resp.status_code}: {content}",
                "status_code": resp.status_code,
            }

        try:
            return True, resp.json()
        except json.JSONDecodeError:
            return True, {"raw": resp.content}

    except requests.exceptions.Timeout:
        return False, {"error": "request timeout"}

    except requests.exceptions.RequestException as e:
        return False, {"error": str(e)}

    except Exception as e:
        return False, {"error": f"unexpected error: {str(e)}"}


def http_get(url, params=None, headers=None, timeout=PACKAGE_REQUEST_TIMEOUT):
    """GET request"""
    return http_request(
        url, method="GET", params=params, headers=headers, timeout=timeout
    )


def http_post_json(url, data, headers=None, timeout=PACKAGE_REQUEST_TIMEOUT):
    """POST JSON request"""
    return http_request(
        url, method="POST", json_data=data, headers=headers, timeout=timeout
    )


def http_upload_file(
    url,
    file_path,
    form_data=None,
    headers=None,
    timeout=PACKAGE_REQUEST_TIMEOUT,
):
    """Upload file (automatically handles bk-nodemgr protocol)"""
    import os

    if not os.path.exists(file_path):
        return False, {"error": f"file not found: {file_path}"}

    filename = os.path.basename(file_path)
    data = {"filename": filename}

    if form_data:
        data.update(form_data)

    return http_request(
        url=url,
        method="POST",
        form_data=data,
        files={"file": file_path},
        headers=headers,
        timeout=timeout,
    )


# =================== file client ===================


class FileClient(object):
    """bk-nodemgr File Service Client (refer to pkg/thirdparty/file/file.go)"""

    def __init__(
        self,
        host="localhost",
        port=28202,
        jwt_key="",
        expire_hours=24,
        tenant_id=None,
        bk_username=None,
        login_name=None,
    ):
        self.jwt_key = jwt_key
        self.expire_hours = expire_hours
        self.tenant_id = tenant_id
        self.bk_username = bk_username
        self.login_name = login_name
        self.base_url = f"http://{host}:{port}/api/v3"

    def _get_common_headers(self):
        headers = {
            "X-Bknodemgr-Request-Id": generate_id("rid"),
        }

        if self.jwt_key:
            ok, jwt_token = generate_jwt_token(
                self.jwt_key, self.expire_hours, self.bk_username, self.login_name
            )
            if not ok:
                raise Exception(f"Failed to generate JWT token: {jwt_token}")

            headers["X-Bknodemgr-Authorization"] = jwt_token

        if self.tenant_id:
            headers["X-Bk-Tenant-Id"] = self.tenant_id

        return headers

    def _call_api(self, method, path, json_data=None, file_path=None, metadata=None):
        url = self.base_url + path
        headers = self._get_common_headers()

        if file_path:
            if not os.path.exists(file_path):
                return False, f"file not found: {file_path}", None

            filename = os.path.basename(file_path)
            form_data = {"filename": filename}

            if metadata:
                form_data["metadata"] = json.dumps(metadata)

            ok, resp = http_request(
                url=url,
                method="POST",
                headers=headers,
                form_data=form_data,
                files={"file": file_path},
                timeout=PACKAGE_REQUEST_TIMEOUT,
            )

        elif json_data is not None:
            ok, resp = http_request(
                url=url,
                method=method,
                headers=headers,
                json_data=json_data,
                timeout=PACKAGE_REQUEST_TIMEOUT,
            )
        else:
            return False, "neither file_path nor json_data provided", None

        if not ok:
            error_msg = resp.get("error", "request failed")
            return False, error_msg, None

        code = resp.get("code")
        if code != 0:
            message = resp.get("message", "api failed")
            return False, f"code({code}), message({message})", None

        data = resp.get("data")
        return True, "ok", data

    # ---------- upload

    def upload_origin_agent(self, file_path, generation=2, overwrite=False):
        """Upload Agent package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/agent",
            file_path=file_path,
            metadata={"generation": generation, "overwrite": overwrite},
        )

    def upload_origin_proxy(self, file_path, generation=2, overwrite=False):
        """Upload Proxy package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/proxy",
            file_path=file_path,
            metadata={"generation": generation, "overwrite": overwrite},
        )

    def upload_origin_server(self, file_path, generation=2, overwrite=False):
        """Upload Server package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/server",
            file_path=file_path,
            metadata={"generation": generation, "overwrite": overwrite},
        )

    def upload_origin_cert(self, file_path, overwrite=False):
        """Upload certificate file"""
        return self._call_api(
            method="POST",
            path="/upload/origin/cert",
            file_path=file_path,
            metadata={"overwrite": overwrite},
        )

    def upload_origin_bintool(self, file_path, generation=2, overwrite=False):
        """Upload BinTool"""
        return self._call_api(
            method="POST",
            path="/upload/origin/bintool",
            file_path=file_path,
            metadata={"generation": generation, "overwrite": overwrite},
        )

    def upload_origin_plugin_bintool(self, file_path, overwrite=False):
        """Upload Plugin BinTool"""
        return self._call_api(
            method="POST",
            path="/upload/origin/plugin_bintool",
            file_path=file_path,
            metadata={"overwrite": overwrite},
        )

    def upload_origin_plugin_v2(self, file_path, overwrite=False):
        """Upload V2 plugin package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/v2/plugin",
            file_path=file_path,
            metadata={"overwrite": overwrite},
        )

    def upload_origin_external_plugin_v2(self, file_path, overwrite=False):
        """Upload V2 external plugin package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/v2/external_plugin",
            file_path=file_path,
            metadata={"overwrite": overwrite},
        )

    def upload_origin_plugin_v3(self, file_path, overwrite=False):
        """Upload V3 plugin package"""
        return self._call_api(
            method="POST",
            path="/upload/origin/v3/plugin",
            file_path=file_path,
            metadata={"overwrite": overwrite},
        )

    # ---------- publish

    def publish_release_agent(self, upload_id):
        """Publish Agent package"""
        return self._call_api(
            method="POST",
            path="/publish/release/agent",
            json_data={"upload_id": upload_id},
        )

    def publish_release_proxy(self, upload_id, upload_origin_pkg_type="origin_proxy"):
        """Publish Proxy/Server package (distinguished by upload_origin_pkg_type)"""
        return self._call_api(
            method="POST",
            path="/publish/release/proxy",
            json_data={
                "upload_id": upload_id,
                "upload_origin_pkg_type": upload_origin_pkg_type,
            },
        )

    def publish_release_server(self, upload_id):
        """Publish Server package (internally calls publish_release_proxy)"""
        return self.publish_release_proxy(
            upload_id, upload_origin_pkg_type="origin_server"
        )

    def publish_release_cert(self, upload_id):
        """Publish certificate"""
        return self._call_api(
            method="POST",
            path="/publish/release/cert",
            json_data={"upload_id": upload_id},
        )

    def publish_release_bintool(self, upload_id):
        """Publish BinTool"""
        return self._call_api(
            method="POST",
            path="/publish/release/bintool",
            json_data={"upload_id": upload_id},
        )

    def publish_release_plugin_bintool(self, upload_id):
        """Publish Plugin BinTool"""
        return self._call_api(
            method="POST",
            path="/publish/release/plugin_bintool",
            json_data={"upload_id": upload_id},
        )

    def publish_release_plugin_v2(self, upload_id):
        """Publish V2 plugin"""
        return self._call_api(
            method="POST",
            path="/publish/release/v2/plugin",
            json_data={"upload_id": upload_id},
        )

    def publish_release_external_plugin_v2(self, upload_id):
        """Publish V2 external plugin"""
        return self._call_api(
            method="POST",
            path="/publish/release/v2/external_plugin",
            json_data={"upload_id": upload_id},
        )

    def publish_release_plugin_v3(self, upload_id):
        """Publish V3 plugin"""
        return self._call_api(
            method="POST",
            path="/publish/release/v3/plugin",
            json_data={"upload_id": upload_id},
        )

    # ---------- release state

    def enable_release_agent(self, identifier):
        """Enable an Agent release."""
        return self._call_api("POST", "/release/agent/enable", json_data=identifier)

    def set_as_default_release_agent(self, identifier):
        """Set an Agent release as default."""
        return self._call_api(
            "POST", "/release/agent/set_as_default", json_data=identifier
        )

    def enable_release_proxy(self, identifier):
        """Enable a Proxy release."""
        return self._call_api("POST", "/release/proxy/enable", json_data=identifier)

    def set_as_default_release_proxy(self, identifier):
        """Set a Proxy release as default."""
        return self._call_api(
            "POST", "/release/proxy/set_as_default", json_data=identifier
        )

    def enable_release_plugin(self, identifier):
        """Enable a Plugin release."""
        return self._call_api("POST", "/release/plugin/enable", json_data=identifier)

    def set_as_default_release_plugin(self, identifier):
        """Set a Plugin release as default."""
        return self._call_api(
            "POST", "/release/plugin/set_as_default", json_data=identifier
        )


def enable_default_package(client, task_name, upload_data):
    """Enable and set default for every platform produced by the selected package."""
    release_type = "plugin"
    if task_name == "agent":
        release_type = "agent"
    elif task_name in ("proxy", "server"):
        release_type = "proxy"

    version = upload_data.get("version")
    platforms = upload_data.get("platforms")
    if not version or not platforms:
        print("release state failed: missing version or platforms in upload response")
        return False

    # File publishes all supported origin package formats as Generation2 releases.
    identifier = {"generation": 2, "version": version}
    if release_type == "plugin":
        identifier["name"] = upload_data.get("plugin_pkg_name")
        if not identifier["name"]:
            print("release state failed: missing plugin name in upload response")
            return False

    enable = getattr(client, "enable_release_" + release_type)
    set_default = getattr(client, "set_as_default_release_" + release_type)
    succeeded = True
    for platform in platforms:
        try:
            key = dict(identifier, platform=platform)
            print(f"setting release state: {release_type} {key}")
            ok, msg, _ = enable(key)
            if not ok:
                print(f"enable failed: {msg}")
                succeeded = False
                continue
            ok, msg, _ = set_default(key)
            if not ok:
                print(f"set as default failed: {msg}")
                succeeded = False
                continue
            print("enable and set as default success")
        except Exception as error:
            print(f"release state failed: {error}")
            succeeded = False
    return succeeded


if __name__ == "__main__":
    p = argparse.ArgumentParser(
        description="bk-nodemgr File Service Upload & Publish Tool"
    )

    p.add_argument(
        "--host",
        action="store",
        dest="host",
        help="file service host (falls back to localhost)",
        default=None,
    )
    p.add_argument(
        "--port",
        action="store",
        dest="port",
        type=int,
        help="file service port (uses File config, falls back to 28202)",
        default=None,
    )
    p.add_argument(
        "--jwt-key",
        action="store",
        dest="jwt_key",
        help="JWT authentication key (uses File config when JWT is enabled)",
        default=None,
    )
    p.add_argument(
        "--tenant-id",
        action="store",
        dest="tenant_id",
        help="tenant id (multiple: system; otherwise: default)",
        default=None,
    )
    p.add_argument(
        "--bk-username",
        action="store",
        dest="bk_username",
        help="bk username (falls back to admin)",
        default=None,
    )
    p.add_argument(
        "--login-name",
        action="store",
        dest="login_name",
        help="login name (falls back to admin)",
        default=None,
    )
    p.add_argument(
        "--expire-hours",
        action="store",
        dest="expire_hours",
        type=int,
        help="JWT token expire hours (uses File config, falls back to 24)",
        default=None,
    )

    p.add_argument(
        "--config-file",
        help="File service YAML configuration path",
    )
    p.add_argument(
        "--auto-select",
        action="store_true",
        help="discover packages under the conventional package directory",
    )
    p.add_argument(
        "--packages-dir",
        default=DEFAULT_PACKAGES_DIR,
        help=f"package discovery directory (default: {DEFAULT_PACKAGES_DIR})",
    )

    p.add_argument(
        "--init-cert",
        action="append",
        dest="init_cert",
        help="upload and publish cert file",
    )
    p.add_argument(
        "--init-bintool",
        action="append",
        dest="init_bintool",
        help="upload and publish bintool file",
    )
    p.add_argument(
        "--init-plugin-bintool",
        action="append",
        dest="init_plugin_bintool",
        help="upload and publish plugin bintool file",
    )
    p.add_argument(
        "--init-agent",
        action="append",
        dest="init_agent",
        help="upload and publish agent package",
    )
    p.add_argument(
        "--init-proxy",
        action="append",
        dest="init_proxy",
        help="upload and publish proxy package",
    )
    p.add_argument(
        "--init-server",
        action="append",
        dest="init_server",
        help="upload and publish server package",
    )
    p.add_argument(
        "--init-plugin-v2",
        action="append",
        dest="init_plugin_v2",
        help="upload and publish v2 plugin package",
    )
    p.add_argument(
        "--init-external-plugin-v2",
        action="append",
        dest="init_external_plugin_v2",
        help="upload and publish v2 external plugin package",
    )
    p.add_argument(
        "--init-plugin-v3",
        action="append",
        dest="init_plugin_v3",
        help="upload and publish v3 plugin package",
    )

    p.add_argument(
        "--generation",
        action="store",
        dest="generation",
        type=int,
        help="generation version (default: 2)",
        default=2,
    )
    p.add_argument(
        "--overwrite",
        action="store_true",
        dest="overwrite",
        help="overwrite existing files",
    )

    p.add_argument(
        "--set-as-default",
        action="store_true",
        help="enable and set default for the last package in each group after publishing",
    )

    args = p.parse_args()

    file_config, config_source, config_errors = load_file_config(args.config_file)
    if args.config_file and not config_source:
        errors = config_errors or [f"{args.config_file}: file not found"]
        for config_error in errors:
            print(f"failed to load config: {config_error}")
        sys.exit(1)

    for config_error in config_errors:
        print(f"warning: unable to load config: {config_error}")
    if config_source:
        print(f"config source: {config_source}")

    runtime_config = resolve_runtime_config(args, file_config)
    print(
        "runtime config: tenant_id={}, host={}, port={}".format(
            runtime_config["tenant_id"],
            runtime_config["host"],
            runtime_config["port"],
        )
    )

    explicit_packages = {
        task_name: getattr(args, f"init_{task_name}") or [] for task_name in TASK_ORDER
    }
    discovered_packages = (
        discover_packages(args.packages_dir) if args.auto_select else {}
    )
    package_paths = {
        task_name: explicit_packages[task_name] + discovered_packages.get(task_name, [])
        for task_name in TASK_ORDER
    }
    package_count = sum(len(paths) for paths in package_paths.values())
    if package_count == 0:
        print("No packages to upload")
        sys.exit(0)

    try:
        client = FileClient(
            host=runtime_config["host"],
            port=runtime_config["port"],
            jwt_key=runtime_config["jwt_key"],
            expire_hours=runtime_config["expire_hours"],
            tenant_id=runtime_config["tenant_id"],
            bk_username=runtime_config["bk_username"],
            login_name=runtime_config["login_name"],
        )
    except Exception as e:
        print(f"failed to create file client: {str(e)}")
        sys.exit(1)

    task_handlers = {
        "cert": (
            client.upload_origin_cert,
            client.publish_release_cert,
            {"overwrite": args.overwrite},
        ),
        "bintool": (
            client.upload_origin_bintool,
            client.publish_release_bintool,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        "plugin_bintool": (
            client.upload_origin_plugin_bintool,
            client.publish_release_plugin_bintool,
            {"overwrite": args.overwrite},
        ),
        "agent": (
            client.upload_origin_agent,
            client.publish_release_agent,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        "proxy": (
            client.upload_origin_proxy,
            client.publish_release_proxy,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        "server": (
            client.upload_origin_server,
            client.publish_release_server,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        "plugin_v2": (
            client.upload_origin_plugin_v2,
            client.publish_release_plugin_v2,
            {"overwrite": args.overwrite},
        ),
        "external_plugin_v2": (
            client.upload_origin_external_plugin_v2,
            client.publish_release_external_plugin_v2,
            {"overwrite": args.overwrite},
        ),
        "plugin_v3": (
            client.upload_origin_plugin_v3,
            client.publish_release_plugin_v3,
            {"overwrite": args.overwrite},
        ),
    }

    tasks = []
    for task_name in TASK_ORDER:
        upload_func, publish_func, upload_params = task_handlers[task_name]
        for file_path in package_paths[task_name]:
            tasks.append(
                (task_name, file_path, upload_func, publish_func, upload_params)
            )

    results = []
    default_targets = {}

    for task_name, file_path, upload_func, publish_func, upload_params in tasks:
        task_succeeded = False
        result = {"task_name": task_name, "file_path": file_path, "succeeded": False}
        results.append(result)
        if args.set_as_default and task_name in ("agent", "proxy", "server"):
            group = "agent" if task_name == "agent" else "proxy"
            default_targets[group] = result
        try:
            print("\n" + "=" * 60)
            print(f"processing {task_name}: {file_path}")
            print("=" * 60)

            print(f"[1/2] uploading {task_name}...")
            ok, msg, data = upload_func(file_path, **upload_params)
            if not ok:
                print(f"upload failed: {msg}")
            else:
                upload_id = data.get("upload_id") if data else None
                if not upload_id:
                    print("upload failed: no upload_id in response")
                else:
                    result["upload_data"] = data
                    if args.set_as_default and task_name in (
                        "plugin_v2",
                        "external_plugin_v2",
                        "plugin_v3",
                    ):
                        plugin_name = data.get("plugin_pkg_name")
                        if not plugin_name:
                            raise ValueError("missing plugin name in upload response")
                        default_targets[(task_name, plugin_name)] = result
                    print("upload success")
                    print(f"  upload_id: {upload_id}")
                    print(f"[2/2] publishing {task_name}...")
                    ok, msg, data = publish_func(upload_id)
                    if not ok:
                        print(f"publish failed: {msg}")
                    else:
                        print("publish success")
                        task_succeeded = True
        except Exception as error:
            print(f"task failed: {error}")

        result["succeeded"] = task_succeeded

    if args.set_as_default:
        for result in default_targets.values():
            if not result["succeeded"]:
                continue
            print("\nactivating package: {}".format(result["file_path"]))
            result["succeeded"] = enable_default_package(
                client, result["task_name"], result["upload_data"]
            )

    success_files = [result["file_path"] for result in results if result["succeeded"]]
    failed_files = [
        result["file_path"] for result in results if not result["succeeded"]
    ]

    success_count = len(success_files)
    fail_count = len(failed_files)

    print("\n" + "=" * 60)
    print(f"summary: {success_count} success, {fail_count} failed")
    print("successful files:")
    for file_path in success_files:
        print(f"  - {file_path}")
    print("failed files:")
    for file_path in failed_files:
        print(f"  - {file_path}")
    print("=" * 60)

    sys.exit(0 if fail_count == 0 else 1)
