#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 """

import re
import argparse
from pathlib import Path

def load_template(path: Path) -> str:
    return path.read_text(encoding="utf-8")

def convert_comments(text: str) -> str:
    return re.sub(r"{#\s*(.*?)\s*#}", r"{{/* \1 */}}", text, flags=re.DOTALL)

def convert_variables(text: str) -> str:
    pattern = r"{{\s*-?\s*(?!(?:if|else|end|range|with|define|template|block|/\*)\b)(.*?)\s*-?\s*}}"

    def _var_repl(m):
        inner = m.group(1).strip()

        if inner.startswith('/*'):
            return m.group(0)

        first_word = inner.split()[0] if inner.split() else ""
        if first_word in ['if', 'else', 'end', 'range', 'with', 'define', 'template', 'block']:
            return m.group(0)

        ternary_match = re.match(r'^(.+?)\s+if\s+(.+?)\s+else\s+(.+)$', inner, re.DOTALL)
        if ternary_match:
            return _convert_ternary(ternary_match, m.group(0))

        if ' or ' in inner and not inner.startswith('or '):
            or_parts = inner.split(' or ', 1)
            var_part = or_parts[0].strip()
            default_part = or_parts[1].strip() if len(or_parts) > 1 else ""

            if default_part in ['""', "''"]:
                inner = var_part
            else:
                normalized_var = _normalize_var_expr(var_part)
                original = m.group(0)
                has_left_dash = '{{-' in original or '{{ -' in original
                has_right_dash = '-}}' in original or '- }}' in original
                ld_s = "-" if has_left_dash else ""
                rd_s = "-" if has_right_dash else ""
                return f"{{{{{ld_s} {normalized_var} | default {default_part} {rd_s}}}}}"

        if "|" in inner:
            parts = inner.split("|", 1)
            expr = parts[0].strip()
            filters = parts[1].strip()
        else:
            expr = inner
            filters = ""

        normalized_expr = _normalize_var_expr(expr)

        original = m.group(0)
        has_left_dash = '{{-' in original or '{{ -' in original
        has_right_dash = '-}}' in original or '- }}' in original

        ld_s = "-" if has_left_dash else ""
        rd_s = "-" if has_right_dash else ""

        if filters:
            return f"{{{{{ld_s} {normalized_expr} | {filters} {rd_s}}}}}"
        else:
            return f"{{{{{ld_s} {normalized_expr} {rd_s}}}}}"

    return re.sub(pattern, _var_repl, text)

def _normalize_var_expr(expr: str) -> str:
    expr = expr.strip()

    if expr.startswith(('.', '$', '"', "'")):
        return expr

    if re.match(r'^-?\d+(\.\d+)?$', expr):
        return expr

    if expr in ('true', 'false', 'nil'):
        return expr

    if ' ' in expr and '.' not in expr:
        return expr

    if re.match(r'^[A-Za-z_]\w*(\.[A-Za-z_]\w*)*$', expr):
        return '.' + expr

    return expr

def _convert_ternary(match, original):
    true_value = match.group(1).strip()
    condition = match.group(2).strip()
    false_value = match.group(3).strip()

    normalized_true = _normalize_complex_expr(true_value)
    normalized_false = _normalize_complex_expr(false_value)
    normalized_condition = _normalize_expr(condition)

    has_left_dash = '{{-' in original or '{{ -' in original
    has_right_dash = '-}}' in original or '- }}' in original

    ld = "-" if has_left_dash else ""
    rd = "-" if has_right_dash else ""

    result = "{{" + ld + " if " + normalized_condition + " " + rd + "}}" + \
             "{{" + ld + " " + normalized_true + " " + rd + "}}" + \
             "{{" + ld + " else " + rd + "}}" + \
             "{{" + ld + " " + normalized_false + " " + rd + "}}" + \
             "{{" + ld + " end " + rd + "}}"

    return result

def _normalize_complex_expr(expr: str) -> str:
    expr = expr.strip()

    if expr.startswith(('.', '$', '"', "'")):
        return expr

    if re.match(r'^-?\d+(\.\d+)?$', expr):
        return expr

    if expr in ('true', 'false', 'nil', 'True', 'False', 'None'):
        return expr

    array_pattern = r'^([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\[(\d+)\]((?:\.[A-Za-z_]\w*)*)$'
    array_match = re.match(array_pattern, expr)

    if array_match:
        base_var = array_match.group(1)
        index = array_match.group(2)
        suffix = array_match.group(3)

        normalized_base = '.' + base_var if not base_var.startswith('.') else base_var

        result = f"(index {normalized_base} {index})"

        if suffix:
            result += suffix

        return result

    if re.match(r'^[A-Za-z_]\w*(\.[A-Za-z_]\w*)*$', expr):
        return '.' + expr

    return expr

FILTER_MAP = {
    "upper": "upper",
    "lower": "lower",
    "default": "default",
    "join": "join",
    "replace": "replace",
    "length": "len",
}

def convert_filters(text: str) -> str:
    pattern = r"{{\s*(-)?\s*([.$\w][^{}|]*?)\s*((?:\|\s*[^{}]+)+)\s*(-)?\s*}}"
    def _filters(m):
        ld, base, filters, rd = m.group(1), m.group(2), m.group(3), m.group(4)
        parts = [p.strip() for p in filters.split("|") if p.strip()]
        new_pipe = []
        for p in parts:
            if '(' in p and ')' in p:
                match = re.match(r'([A-Za-z_]\w*)\((.*)\)', p, re.DOTALL)
                if match:
                    name = match.group(1)
                    args_str = match.group(2).strip()

                    args = _split_filter_args(args_str)

                    if name == 'default':
                        mapped = FILTER_MAP.get(name, name)
                        if mapped == "":
                            continue
                        if args:
                            arg = args[0]
                            if arg.strip().startswith('[') and arg.strip().endswith(']'):
                                arg = _convert_array_to_list(arg.strip())
                            new_pipe.append(mapped + " " + arg)
                        else:
                            new_pipe.append(mapped)
                    else:
                        mapped = FILTER_MAP.get(name, name)
                        if mapped == "":
                            continue
                        if args:
                            new_pipe.append(mapped + " " + " ".join(args))
                        else:
                            new_pipe.append(mapped)
                else:
                    new_pipe.append(p)
            else:
                seg = p.split()
                name = seg[0]
                args = seg[1:]
                mapped = FILTER_MAP.get(name, name)
                if mapped == "":
                    continue
                if args:
                    new_pipe.append(mapped + " " + " ".join(args))
                else:
                    new_pipe.append(mapped)

        ld_s = "-" if ld else ""
        rd_s = "-" if rd else ""
        if not new_pipe:
            return "{{" + ld_s + " " + base.strip() + " " + rd_s + "}}"
        return "{{" + ld_s + " " + base.strip() + " | " + " | ".join(new_pipe) + " " + rd_s + "}}"
    return re.sub(pattern, _filters, text)

def _convert_array_to_list(array_str: str) -> str:
    content = array_str.strip()[1:-1].strip()

    if not content:
        return "(list)"

    elements = []
    current = []
    in_quote = None

    for i, char in enumerate(content):
        if char in ('"', "'") and (i == 0 or content[i-1] != '\\'):
            if in_quote == char:
                in_quote = None
            elif in_quote is None:
                in_quote = char
            current.append(char)
        elif in_quote:
            current.append(char)
        elif char == ',':
            if current:
                elements.append(''.join(current).strip())
                current = []
        else:
            current.append(char)

    if current:
        elements.append(''.join(current).strip())

    converted_elements = []
    for elem in elements:
        elem = elem.strip()
        if elem.startswith("'") and elem.endswith("'"):
            elem = '"' + elem[1:-1].replace('"', '\\"').replace("\\'", "'") + '"'
        converted_elements.append(elem)

    return "(list " + " ".join(converted_elements) + ")"

def _split_filter_args(args_str: str) -> list:
    args = []
    current = []
    depth_paren = 0
    depth_bracket = 0
    depth_brace = 0
    in_quote = None

    i = 0
    while i < len(args_str):
        char = args_str[i]

        if char in ('"', "'") and (i == 0 or args_str[i-1] != '\\'):
            if in_quote == char:
                in_quote = None
            elif in_quote is None:
                in_quote = char
            current.append(char)
        elif in_quote:
            current.append(char)
        elif char == '(':
            depth_paren += 1
            current.append(char)
        elif char == ')':
            depth_paren -= 1
            current.append(char)
        elif char == '[':
            depth_bracket += 1
            current.append(char)
        elif char == ']':
            depth_bracket -= 1
            current.append(char)
        elif char == '{':
            depth_brace += 1
            current.append(char)
        elif char == '}':
            depth_brace -= 1
            current.append(char)
        elif char == ',' and depth_paren == 0 and depth_bracket == 0 and depth_brace == 0:
            if current:
                args.append(''.join(current).strip())
                current = []
        else:
            current.append(char)

        i += 1

    if current:
        args.append(''.join(current).strip())

    return args

def convert_if_statements(text: str) -> str:
    def _if(m):
        ld, cond, rd = m.group(1), m.group(2), m.group(3)
        return "{{" + ("-" if ld else "") + " if " + _normalize_expr(cond) + " " + ("-" if rd else "") + "}}"
    def _elif(m):
        ld, cond, rd = m.group(1), m.group(2), m.group(3)
        return "{{" + ("-" if ld else "") + " else if " + _normalize_expr(cond) + " " + ("-" if rd else "") + "}}"
    def _else(m):
        ld, rd = m.group(1), m.group(2)
        return "{{" + ("-" if ld else "") + " else " + ("-" if rd else "") + "}}"
    def _endif(m):
        ld, rd = m.group(1), m.group(2)
        return "{{" + ("-" if ld else "") + " end " + ("-" if rd else "") + "}}"

    text = re.sub(r"{%\s*(-)?\s*if\s+(.*?)\s*(-)?\s*%}", _if, text)
    text = re.sub(r"{%\s*(-)?\s*elif\s+(.*?)\s*(-)?\s*%}", _elif, text)
    text = re.sub(r"{%\s*(-)?\s*else\s*(-)?\s*%}", _else, text)
    text = re.sub(r"{%\s*(-)?\s*endif\s*(-)?\s*%}", _endif, text)
    return text

def convert_for_loops(text: str) -> str:
    def _for_multi(m):
        ld, vars_part, iterable, rd = m.group(1), m.group(2), m.group(3), m.group(4)
        vars_split = [v.strip() for v in vars_part.split(",")]
        if len(vars_split) == 2:
            v1, v2 = vars_split
            v1 = "$" + v1 if not v1.startswith("$") else v1
            v2 = "$" + v2 if not v2.startswith("$") else v2
            iterable_norm = _normalize_iterable(iterable)
            return "{{" + ("-" if ld else "") + f" range {v1}, {v2} := {iterable_norm} " + ("-" if rd else "") + "}}"
        only = vars_split[0]
        only = "$" + only if not only.startswith("$") else only
        iterable_norm = _normalize_iterable(iterable)
        return "{{" + ("-" if ld else "") + f" range {only} := {iterable_norm} " + ("-" if rd else "") + "}}"

    def _for_single(m):
        ld, var, iterable, rd = m.group(1), m.group(2), m.group(3), m.group(4)
        var_sym = "$" + var if not var.startswith("$") else var
        iterable_norm = _normalize_iterable(iterable)
        return "{{" + ("-" if ld else "") + f" range {var_sym} := {iterable_norm} " + ("-" if rd else "") + "}}"

    def _endfor(m):
        ld, rd = m.group(1), m.group(2)
        return "{{" + ("-" if ld else "") + " end " + ("-" if rd else "") + "}}"

    text = re.sub(r"{%\s*(-)?\s*for\s+([A-Za-z_]\w*\s*,\s*[A-Za-z_]\w*)\s+in\s+(.*?)\s*(-)?\s*%}", _for_multi, text)
    text = re.sub(r"{%\s*(-)?\s*for\s+([A-Za-z_]\w*)\s+in\s+(.*?)\s*(-)?\s*%}", _for_single, text)
    text = re.sub(r"{%\s*(-)?\s*endfor\s*(-)?\s*%}", _endfor, text)
    return text

def _normalize_expr(expr: str) -> str:
    expr = expr.strip()

    ast = _parse_logical_expr(expr)
    return _ast_to_go(ast)

def _parse_logical_expr(expr: str):
    expr = expr.strip()

    or_parts = _split_by_operator(expr, 'or')
    if len(or_parts) > 1:
        return {'op': 'or', 'operands': [_parse_logical_expr(p) for p in or_parts]}

    and_parts = _split_by_operator(expr, 'and')
    if len(and_parts) > 1:
        return {'op': 'and', 'operands': [_parse_logical_expr(p) for p in and_parts]}

    if expr.startswith('not '):
        inner = expr[4:].strip()
        return {'op': 'not', 'operands': [_parse_logical_expr(inner)]}

    if expr.startswith('(') and expr.endswith(')'):
        return _parse_logical_expr(expr[1:-1])

    return _parse_comparison(expr)

def _split_by_operator(expr: str, op: str):
    parts = []
    current = []
    depth = 0
    tokens = re.split(r'(\s+|\(|\))', expr)

    i = 0
    while i < len(tokens):
        token = tokens[i]

        if token == '(':
            depth += 1
            current.append(token)
        elif token == ')':
            depth -= 1
            current.append(token)
        elif depth == 0 and token.strip() == op:
            if current:
                parts.append(''.join(current).strip())
                current = []
        else:
            current.append(token)
        i += 1

    if current:
        parts.append(''.join(current).strip())

    return parts if len(parts) > 1 else [expr]

def _parse_comparison(expr: str):
    expr = expr.strip()

    if expr.endswith(' is not defined'):
        var = expr.replace(' is not defined', '').strip()
        return {
            'op': 'not',
            'operands': [{'op': 'value', 'value': _normalize_variable(var)}]
        }

    if expr.endswith(' is defined'):
        var = expr.replace(' is defined', '').strip()
        return {'op': 'value', 'value': _normalize_variable(var)}

    is_not_type_match = re.match(r'^(.+?)\s+is\s+not\s+(\w+)$', expr)
    if is_not_type_match:
        var = is_not_type_match.group(1).strip()
        type_name = is_not_type_match.group(2).strip()
        return {
            'op': 'not',
            'operands': [{
                'op': 'type_check',
                'var': _normalize_variable(var),
                'type': type_name
            }]
        }

    is_type_match = re.match(r'^(.+?)\s+is\s+(\w+)$', expr)
    if is_type_match:
        var = is_type_match.group(1).strip()
        type_name = is_type_match.group(2).strip()
        return {
            'op': 'type_check',
            'var': _normalize_variable(var),
            'type': type_name
        }

    if ' not in ' in expr:
        left, right = expr.split(' not in ', 1)
        return {
            'op': 'not',
            'operands': [{
                'op': 'has',
                'left': _normalize_variable(right.strip()),
                'right': _normalize_variable(left.strip())
            }]
        }

    if ' in ' in expr:
        left, right = expr.split(' in ', 1)
        return {
            'op': 'has',
            'left': _normalize_variable(right.strip()),
            'right': _normalize_variable(left.strip())
        }

    for op in ['==', '!=', '>=', '<=', '>', '<']:
        if f' {op} ' in expr:
            left, right = expr.split(f' {op} ', 1)
            op_map = {
                '==': 'eq', '!=': 'ne',
                '<': 'lt', '<=': 'le',
                '>': 'gt', '>=': 'ge'
            }
            return {
                'op': op_map[op],
                'left': _normalize_variable(left.strip()),
                'right': _normalize_variable(right.strip())
            }

    return {'op': 'value', 'value': _normalize_variable(expr)}

def _ast_to_go(ast, need_parens=False):
    if ast['op'] == 'value':
        return ast['value']

    if ast['op'] in ['and', 'or']:
        operands = []
        for op in ast['operands']:
            sub_result = _ast_to_go(op)
            operands.append(f"({sub_result})")

        result = f"{ast['op']} {' '.join(operands)}"
        return result

    if ast['op'] == 'not':
        inner = ast['operands'][0]
        inner_result = _ast_to_go(inner)
        return f"not ({inner_result})"

    if ast['op'] == 'type_check':
        var = ast['var']
        type_name = ast['type']
        go_type = _map_jinja_type_to_go(type_name)
        return f'kindIs "{go_type}" {var}'

    if ast['op'] == 'has':
        return f"has {ast['left']} {ast['right']}"

    if ast['op'] in ['eq', 'ne', 'lt', 'le', 'gt', 'ge']:
        return f"{ast['op']} {ast['left']} {ast['right']}"

    return str(ast)

def _map_jinja_type_to_go(jinja_type: str) -> str:
    type_map = {
        'string': 'string',
        'int': 'int',
        'integer': 'int',
        'float': 'float64',
        'bool': 'bool',
        'boolean': 'bool',
        'list': 'slice',
        'dict': 'map',
        'iterable': 'slice',
        'mapping': 'map',
        'none': 'invalid',
        'number': 'int',
    }
    return type_map.get(jinja_type.lower(), jinja_type)

def _normalize_variable(var: str) -> str:
    var = var.strip()

    if var.startswith(('.', '$', '"', "'")):
        return var

    if var in ('True', 'true'):
        return 'true'
    if var in ('False', 'false'):
        return 'false'
    if var in ('None', 'nil', 'null'):
        return 'nil'

    if re.match(r'^-?\d+(\.\d+)?$', var):
        return var

    if re.match(r'^[A-Za-z_]\w*(\.[A-Za-z_]\w*)*$', var):
        return '.' + var

    return var

def _normalize_iterable(expr: str) -> str:
    expr = expr.strip()
    if re.match(r"^[A-Za-z_]\w*(\.[A-Za-z_]\w*)*$", expr) and not expr.startswith((".", "$")):
        return "." + expr
    return expr

def convert_quotes(text: str) -> str:

    def _replace_quote(match):
        content = match.group(1)
        content = content.replace('"', '\\"')
        content = content.replace("\\'", "'")
        return f'"{content}"'

    pattern = r"'((?:[^'\\]|\\.)*)'"
    text = re.sub(pattern, _replace_quote, text)

    return text

def annotate_blocks(text: str) -> str:
    text = re.sub(r"{%\s*block\s+([A-Za-z_]\w*)\s*%}", r"{{/* BLOCK \1 (需手动实现) */}}", text)
    text = re.sub(r"{%\s*endblock\s*%}", r"{{/* END BLOCK */}}", text)
    text = re.sub(r"{%\s*extends\s+['\"](.*?)['\"]\s*%}", r"{{/* EXTENDS \1 (需手动迁移) */}}", text)
    return text

def convert_jinja_to_go(text: str) -> str:
    steps = [
        convert_comments,
        annotate_blocks,
        convert_if_statements,
        convert_for_loops,
        convert_variables,
        convert_filters,
        convert_quotes,
    ]
    for fn in steps:
        text = fn(text)
    return text

def main():
    ap = argparse.ArgumentParser(description="transfer Jinja2 template to Go template")
    ap.add_argument("input", help="input Jinja2 template file path")
    ap.add_argument("output", help=" output Go template file path")
    args = ap.parse_args()

    src = Path(args.input)
    dst = Path(args.output)

    raw = load_template(src)
    converted = convert_jinja_to_go(raw)
    dst.write_text(converted, encoding="utf-8")
    print(f"transfer finished -> {dst}")

if __name__ == "__main__":
    main()
