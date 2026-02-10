import argparse
import json
import requests
import os
import uuid
import subprocess

# =================== utils ===================


def generate_id(tag):
    return tag + ":" + str(uuid.uuid4()).replace("-", "")


def generate_jwt_token(key, expire_hours=24):
    script_dir = os.path.dirname(os.path.abspath(__file__))
    binary_name = "jwt-generator"
    binary_path = os.path.join(script_dir, binary_name)

    if not os.path.exists(binary_path):
        return False, "jwt-generator binary not found: {}".format(binary_path)

    try:
        result = subprocess.run(
            [binary_path, "-k", key, "-e", str(expire_hours)],
            capture_output=True,
            text=True,
            timeout=10,
        )

        if result.returncode != 0:
            error_msg = result.stderr.strip() if result.stderr else "unknown error"
            return False, "jwt-generator failed: {}".format(error_msg)

        token = result.stdout.strip()
        if not token:
            return False, "jwt-generator returned empty token"

        # debug: check token format
        if token.count(".") != 2:
            return (
                False,
                "invalid JWT token format (expected 3 segments, got {}): {}".format(
                    token.count(".") + 1, token[:100]
                ),
            )

        return True, token

    except subprocess.TimeoutExpired:
        return False, "jwt-generator execution timeout"
    except Exception as e:
        return False, "failed to execute jwt-generator: {}".format(str(e))


# =================== http request ===================


def http_request(
    url,
    method="GET",
    headers=None,
    params=None,
    json_data=None,
    form_data=None,
    files=None,
    timeout=30,
):
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
                "error": "HTTP {}: {}".format(resp.status_code, content),
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
        return False, {"error": "unexpected error: {}".format(str(e))}


def http_get(url, params=None, headers=None, timeout=30):
    """GET request"""
    return http_request(
        url, method="GET", params=params, headers=headers, timeout=timeout
    )


def http_post_json(url, data, headers=None, timeout=30):
    """POST JSON request"""
    return http_request(
        url, method="POST", json_data=data, headers=headers, timeout=timeout
    )


def http_upload_file(url, file_path, form_data=None, headers=None, timeout=30):
    """Upload file (automatically handles bk-nodemgr protocol)"""
    import os

    if not os.path.exists(file_path):
        return False, {"error": "file not found: {}".format(file_path)}

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
    ):
        self.jwt_key = jwt_key
        self.expire_hours = expire_hours
        self.tenant_id = tenant_id
        self.base_url = "http://{host}:{port}/api/v3".format(host=host, port=port)

    def _get_common_headers(self):
        headers = {
            "X-Bknodemgr-Request-Id": generate_id("rid"),
        }

        if self.jwt_key:
            ok, jwt_token = generate_jwt_token(self.jwt_key, self.expire_hours)
            if not ok:
                raise Exception("Failed to generate JWT token: {}".format(jwt_token))

            headers["X-Bknodemgr-Authorization"] = jwt_token

        if self.tenant_id:
            headers["X-Bk-Tenant-Id"] = self.tenant_id

        return headers

    def _call_api(self, method, path, json_data=None, file_path=None, metadata=None):
        url = self.base_url + path
        headers = self._get_common_headers()

        if file_path:
            if not os.path.exists(file_path):
                return False, "file not found: {}".format(file_path), None

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
                timeout=300,
            )

        elif json_data is not None:
            ok, resp = http_request(
                url=url, method=method, headers=headers, json_data=json_data, timeout=30
            )
        else:
            return False, "neither file_path nor json_data provided", None

        if not ok:
            error_msg = resp.get("error", "request failed")
            return False, error_msg, None

        code = resp.get("code")
        if code != 0:
            message = resp.get("message", "api failed")
            return False, "code({}), message({})".format(code, message), None

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
            path="/publish/release/server",
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


if __name__ == "__main__":
    p = argparse.ArgumentParser(
        description="bk-nodemgr File Service Upload & Publish Tool"
    )

    p.add_argument(
        "--host",
        action="store",
        dest="host",
        help="file service host (default: localhost)",
        default="localhost",
    )
    p.add_argument(
        "--port",
        action="store",
        dest="port",
        type=int,
        help="file service port (default: 28202)",
        default=28202,
    )
    p.add_argument(
        "--jwt-key",
        action="store",
        dest="jwt_key",
        help="JWT authentication key (optional, leave empty if JWT is disabled)",
        default="",
    )
    p.add_argument(
        "--tenant-id",
        action="store",
        dest="tenant_id",
        help="tenant id",
        default="default",
    )
    p.add_argument(
        "--expire-hours",
        action="store",
        dest="expire_hours",
        type=int,
        help="JWT token expire hours (default: 24)",
        default=24,
    )

    p.add_argument(
        "--init-cert",
        action="store",
        dest="init_cert",
        help="upload and publish cert file",
    )
    p.add_argument(
        "--init-bintool",
        action="store",
        dest="init_bintool",
        help="upload and publish bintool file",
    )
    p.add_argument(
        "--init-plugin-bintool",
        action="store",
        dest="init_plugin_bintool",
        help="upload and publish plugin bintool file",
    )
    p.add_argument(
        "--init-agent",
        action="store",
        dest="init_agent",
        help="upload and publish agent package",
    )
    p.add_argument(
        "--init-proxy",
        action="store",
        dest="init_proxy",
        help="upload and publish proxy package",
    )
    p.add_argument(
        "--init-server",
        action="store",
        dest="init_server",
        help="upload and publish server package",
    )
    p.add_argument(
        "--init-plugin-v2",
        action="store",
        dest="init_plugin_v2",
        help="upload and publish v2 plugin package",
    )
    p.add_argument(
        "--init-external-plugin-v2",
        action="store",
        dest="init_external_plugin_v2",
        help="upload and publish v2 external plugin package",
    )
    p.add_argument(
        "--init-plugin-v3",
        action="store",
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

    args = p.parse_args()

    try:
        client = FileClient(
            host=args.host,
            port=args.port,
            jwt_key=args.jwt_key,
            expire_hours=args.expire_hours,
            tenant_id=args.tenant_id,
        )
    except Exception as e:
        print("failed to create file client: {}".format(str(e)))
        exit(1)

    tasks = [
        (
            "cert",
            args.init_cert,
            client.upload_origin_cert,
            client.publish_release_cert,
            {"overwrite": args.overwrite},
        ),
        (
            "bintool",
            args.init_bintool,
            client.upload_origin_bintool,
            client.publish_release_bintool,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        (
            "plugin_bintool",
            args.init_plugin_bintool,
            client.upload_origin_plugin_bintool,
            client.publish_release_plugin_bintool,
            {"overwrite": args.overwrite},
        ),
        (
            "agent",
            args.init_agent,
            client.upload_origin_agent,
            client.publish_release_agent,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        (
            "proxy",
            args.init_proxy,
            client.upload_origin_proxy,
            client.publish_release_proxy,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        (
            "server",
            args.init_server,
            client.upload_origin_server,
            client.publish_release_server,
            {"generation": args.generation, "overwrite": args.overwrite},
        ),
        (
            "plugin_v2",
            args.init_plugin_v2,
            client.upload_origin_plugin_v2,
            client.publish_release_plugin_v2,
            {"overwrite": args.overwrite},
        ),
        (
            "external_plugin_v2",
            args.init_external_plugin_v2,
            client.upload_origin_external_plugin_v2,
            client.publish_release_external_plugin_v2,
            {"overwrite": args.overwrite},
        ),
        (
            "plugin_v3",
            args.init_plugin_v3,
            client.upload_origin_plugin_v3,
            client.publish_release_plugin_v3,
            {"overwrite": args.overwrite},
        ),
    ]

    success_count = 0
    fail_count = 0

    print("tenant_id: {}".format(args.tenant_id))

    for task_name, file_path, upload_func, publish_func, upload_params in tasks:
        if not file_path:
            continue

        print("\n" + "=" * 60)
        print("processing {}: {}".format(task_name, file_path))
        print("=" * 60)

        print("[1/2] uploading {}...".format(task_name))
        ok, msg, data = upload_func(file_path, **upload_params)
        if not ok:
            print("upload failed: {}".format(msg))
            fail_count += 1
            continue

        upload_id = data.get("upload_id") if data else None
        if not upload_id:
            print("upload failed: no upload_id in response")
            fail_count += 1
            continue

        print("upload success")
        print("  upload_id: {}".format(upload_id))

        if publish_func:
            print("[2/2] publishing {}...".format(task_name))
            ok, msg, data = publish_func(upload_id)
            if not ok:
                print("publish failed: {}".format(msg))
                fail_count += 1
                continue

            print("publish success")
            success_count += 1
        else:
            print("[2/2] skipped (no publish endpoint for {})".format(task_name))
            success_count += 1

    print("\n" + "=" * 60)
    print("summary: {} success, {} failed".format(success_count, fail_count))
    print("=" * 60)

    exit(0 if fail_count == 0 else 1)
