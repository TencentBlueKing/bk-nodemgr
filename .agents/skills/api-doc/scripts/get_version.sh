#!/bin/bash
# 获取 API 文档使用的版本号
# 规则遵循 issue #1442：v$ArchVer.$Major.$Minor-$Tag.$Patch[[-.]$Mark]

set -euo pipefail

LATEST_TAG=$(git for-each-ref --sort=-creatordate --format='%(refname:short)' refs/tags | head -1)

if [ -z "$LATEST_TAG" ]; then
    echo "未找到任何版本标签"
    exit 1
fi

python - "$LATEST_TAG" <<'PY'
import re
import sys

tag = sys.argv[1]
match = re.match(r'^(v\d+\.\d+\.\d+)-([a-zA-Z]+)\.(\d+)(?:[-.](.+))?$', tag)
if match:
    base_version, tag_name, patch_number, _mark = match.groups()
    print(f"{base_version}-{tag_name}.{int(patch_number) + 1}+")
else:
    print(f"{tag}+")
PY
