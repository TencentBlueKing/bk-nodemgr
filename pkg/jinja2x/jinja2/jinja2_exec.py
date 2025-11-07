# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at https://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import argparse
import copy
import json
import sys
from pathlib import Path

import six
from jinja2.sandbox import SandboxedEnvironment as Environment

"""
jinja2渲染相关的公共函数
"""

TEMPLATE_CACHE = {}


def find_element(element, dict_data):
    """
    根据路径字符串获取字典定义的嵌套key的值
    :param element: 路径字符串，格式 a.b.c.d
    :param dict_data: 字典数据
    :return: value
    """
    keys = element.split(".")
    rv = dict_data
    for key in keys:
        rv = rv[key]
    return rv


def nested_render_data(data, context):
    """
    递归渲染字典中的模板字符串
    """
    if isinstance(data, six.string_types):
        if "{{" not in data:
            # 无 jinja 占位符，直接跳过
            return data
        try:
            # 尝试渲染用户参数，一旦失败，立即返回原数据
            template = TEMPLATE_CACHE.get(data)
            if not template:
                template = Environment().from_string(data)
                TEMPLATE_CACHE[data] = template
            return template.render(context)
        except Exception as err:
            print(f"render template failed: {err}", file=sys.stderr)
            return data
    elif isinstance(data, dict):
        if "$for" in data and "$item" in data and "$body" in data:
            # 循环动态变量解析
            data_list = []
            # 提取列表变量
            for_list = find_element(data["$for"], context)

            for item in for_list:
                # 临时设置上下文
                context[data["$item"]] = item
                # 深拷贝一次，防止原始模板被修改
                body_data = copy.deepcopy(data["$body"])
                data_list.append(nested_render_data(body_data, context))
                # 恢复上下文
                context.pop(data["$item"])
            return data_list

        for key, value in data.items():
            data[key] = nested_render_data(value, context)
    elif isinstance(data, list):
        for index, value in enumerate(data):
            data[index] = nested_render_data(value, context)
    return data


def load_template_file(template_path):
    """
    加载模板文件
    :param template_path: 模板文件路径
    :return: 模板内容
    """
    try:
        template_file = Path(template_path)
        if not template_file.exists():
            raise FileNotFoundError(f"Template file not found: {template_path}")

        return template_file.read_text(encoding="utf-8")
    except Exception as e:
        print(f"Failed to load template file: {e}", file=sys.stderr)
        sys.exit(1)


def load_context_file(context_path):
    """
    加载上下文文件（支持 JSON 格式）
    :param context_path: 上下文文件路径
    :return: 上下文字典
    """
    try:
        context_file = Path(context_path)
        if not context_file.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        content = context_file.read_text(encoding="utf-8")

        # 尝试解析为 JSON
        try:
            return json.loads(content)
        except json.JSONDecodeError as e:
            # 如果不是 JSON，尝试作为 Python 字面量解析（兼容原有的示例数据）
            try:
                import ast
                return ast.literal_eval(content)
            except (ValueError, SyntaxError):
                raise ValueError(f"Context file is not valid JSON or Python literal: {e}")

    except Exception as e:
        print(f"Failed to load context file: {e}", file=sys.stderr)
        sys.exit(1)


def save_output_file(output_path, content):
    """
    保存渲染结果到文件
    :param output_path: 输出文件路径
    :param content: 渲染内容
    """
    try:
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        output_file.write_text(content, encoding="utf-8")
        print(f"Rendered template saved to: {output_path}", file=sys.stderr)
    except Exception as e:
        print(f"Failed to save output file: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    """
    主函数，支持命令行参数
    """
    parser = argparse.ArgumentParser(
        description="Jinja2 Template Renderer - Render Jinja2 templates with context data",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Render template with context and save to file
  python jinjia2.py template.yaml context.json output.yaml

  # Render template with context and print to stdout
  python jinjia2.py template.yaml context.json

  # Use with JSON context data
  python jinjia2.py config.yaml data.json result.yaml
        """
    )

    # 支持位置参数或选项参数
    parser.add_argument("template", nargs="?", help="Input Jinja2 template file path")
    parser.add_argument("context", nargs="?", help="Input context file path (JSON format)")
    parser.add_argument("output", nargs="?", help="Output file path (optional, prints to stdout if not specified)")
    parser.add_argument("-t", "--template", dest="template_file", help="Input Jinja2 template file path")
    parser.add_argument("-c", "--context", dest="context_file", help="Input context file path (JSON format)")
    parser.add_argument("-o", "--output", dest="output_file", help="Output file path")

    args = parser.parse_args()

    # 确定文件路径（优先使用选项参数）
    template_path = args.template_file or args.template
    context_path = args.context_file or args.context
    output_path = args.output_file or args.output

    # 检查必需参数
    if not template_path or not context_path:
        parser.error("Both template and context files are required. Use positional arguments or -t/-c options.")

    # 加载模板文件
    if args.verbose:
        print(f"Loading template from: {template_path}", file=sys.stderr)
    template_content = load_template_file(template_path)

    # 加载上下文文件
    if args.verbose:
        print(f"Loading context from: {context_path}", file=sys.stderr)
    context_data = load_context_file(context_path)

    # 渲染模板
    if args.verbose:
        print("Rendering template...", file=sys.stderr)

    rendered_result = nested_render_data(template_content, context_data)

    # 输出结果
    if output_path:
        # 保存到文件
        save_output_file(output_path, rendered_result)
    else:
        # 输出到标准输出
        print(rendered_result)


if __name__ == "__main__":
    main()

