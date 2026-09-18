#!/bin/bash
set -euo pipefail

# 配置
FROM_VERSION="${1:-v3.0.1-alpha.17}"
TO_VERSION="${2:-master}"
OUTPUT_DIR=".diff/${FROM_VERSION}~${TO_VERSION}"

echo "Extracting backend swagger changes from ${FROM_VERSION} to ${TO_VERSION}"
echo "Output directory: ${OUTPUT_DIR}"

# 创建输出目录
mkdir -p "${OUTPUT_DIR}"

# 获取所有变更的 backend swagger 文件
changed_files=$(git diff "${FROM_VERSION}..${TO_VERSION}" --name-only -- 'docs/api/swagger/backend/**/*.swagger.json' || true)

if [ -z "$changed_files" ]; then
    echo "No backend swagger files changed between ${FROM_VERSION} and ${TO_VERSION}"
    exit 0
fi

echo "Found changed backend swagger files:"
echo "$changed_files"
echo ""

# 复制每个变更的文件
while IFS= read -r file; do
    if [ -z "$file" ]; then
        continue
    fi
    
    # 检查文件是否存在于目标版本
    if git cat-file -e "${TO_VERSION}:${file}" 2>/dev/null; then
        echo "Extracting: ${file}"
        
        # 创建目标目录结构
        target_dir="${OUTPUT_DIR}/$(dirname "${file}")"
        mkdir -p "${target_dir}"
        
        # 提取文件
        git show "${TO_VERSION}:${file}" > "${OUTPUT_DIR}/${file}"
    else
        echo "Skipping (deleted): ${file}"
    fi
done <<< "$changed_files"

echo ""
echo "Extraction complete!"
echo "Changed backend swagger files are in: ${OUTPUT_DIR}"
echo ""
echo "Summary:"
find "${OUTPUT_DIR}" -name "*.swagger.json" -type f | sort
