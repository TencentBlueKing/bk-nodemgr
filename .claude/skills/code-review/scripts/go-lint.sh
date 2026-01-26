#!/bin/bash
# 进行 Go 代码 lint 检查

set -e

# 获取项目根目录
PROJECT_ROOT=$(git rev-parse --show-toplevel)

cd ${PROJECT_ROOT}

golangci-lint run --config ${PROJECT_ROOT}/.golangci.yml --path-prefix ${PROJECT_ROOT}