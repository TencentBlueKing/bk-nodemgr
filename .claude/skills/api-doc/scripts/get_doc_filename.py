#!/usr/bin/env python3
"""
根据 API URL 查找对应的 swagger operationId，生成文档文件名

使用方法:
    get_doc_filename.py <url> [method]

参数:
    url     - API URL 路径，如 /api/v3/node/agent/install
    method  - HTTP 方法（可选，默认 POST），如 GET, POST, PUT, DELETE

示例:
    get_doc_filename.py /api/v3/node/agent/install
    get_doc_filename.py /api/v3/node/agent/install POST
"""

import json
import sys
import os
from pathlib import Path


def find_operation_id(url, method='POST'):
    """在 swagger 文件中查找 operationId"""
    # 项目根目录
    project_root = Path(__file__).parent.parent.parent.parent.parent
    swagger_dir = project_root / 'docs' / 'api' / 'swagger'

    if not swagger_dir.exists():
        print(f"错误: Swagger 目录不存在: {swagger_dir}", file=sys.stderr)
        sys.exit(1)

    # 遍历所有 swagger.json 文件
    swagger_files = list(swagger_dir.glob('**/*.swagger.json'))

    if not swagger_files:
        print(f"错误: 在 {swagger_dir} 中未找到 swagger.json 文件", file=sys.stderr)
        sys.exit(1)

    method = method.lower()

    for swagger_file in swagger_files:
        try:
            with open(swagger_file, 'r', encoding='utf-8') as f:
                data = json.load(f)

            # 查找路径
            paths = data.get('paths', {})
            if url in paths:
                path_item = paths[url]
                if method in path_item:
                    operation = path_item[method]
                    operation_id = operation.get('operationId')
                    if operation_id:
                        return operation_id, swagger_file
        except Exception as e:
            print(f"警告: 读取 {swagger_file} 时出错: {e}", file=sys.stderr)
            continue

    return None, None


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)

    url = sys.argv[1]
    method = sys.argv[2].upper() if len(sys.argv) > 2 else 'POST'

    # 查找 operationId
    operation_id, swagger_file = find_operation_id(url, method)

    if not operation_id:
        print(f"错误: 未找到 {method} {url} 对应的 operationId", file=sys.stderr)
        print(f"提示: 请检查 URL 和 HTTP 方法是否正确", file=sys.stderr)
        sys.exit(1)

    # 直接使用 operationId 作为文件名
    filename = operation_id

    # 输出结果
    print(f"{filename}.md")

    # 输出详细信息到 stderr（不影响主输出）
    print(f"", file=sys.stderr)
    print(f"✓ 找到匹配:", file=sys.stderr)
    print(f"  URL: {url}", file=sys.stderr)
    print(f"  Method: {method}", file=sys.stderr)
    print(f"  OperationId: {operation_id}", file=sys.stderr)
    print(f"  Swagger: {swagger_file.name}", file=sys.stderr)
    print(f"  文档文件名: {filename}.md", file=sys.stderr)


if __name__ == '__main__':
    main()
