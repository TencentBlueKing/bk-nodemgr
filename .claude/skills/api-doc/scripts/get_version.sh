#!/bin/bash
# 获取项目的最新正式版本号
# 用于 API 文档"该接口提供版本"字段

set -e

# 获取最新的正式版本号（排除 alpha、beta、test 等预发布版本）
LATEST_VERSION=$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1)

if [ -z "$LATEST_VERSION" ]; then
    echo "未找到正式版本标签"
    echo "提示: 可能只存在预发布版本 (alpha/beta/test)"
    exit 1
fi

echo "$LATEST_VERSION"
