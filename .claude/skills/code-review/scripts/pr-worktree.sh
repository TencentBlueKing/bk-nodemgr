#!/bin/bash
# PR Worktree Creator
# 根据 PR URL 自动创建 git worktree，用于在独立工作区审查 PR
#
# 支持平台:
#   - GitHub / GitHub Enterprise
#   - GitLab / 腾讯工蜂 (git.code.tencent.com)
#   - Gitee
#
# 用法: pr-worktree.sh <PR_URL>
# 示例:
#   pr-worktree.sh https://github.com/TencentBlueKing/bk-nodemgr/pull/123
#   pr-worktree.sh https://gitlab.com/owner/repo/-/merge_requests/456
#   pr-worktree.sh https://git.code.tencent.com/owner/repo/merge_requests/789

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 打印函数
info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 显示帮助
show_help() {
    cat << 'EOF'
用法: pr-worktree.sh <PR_URL>

根据 PR URL 自动创建 git worktree，用于在独立工作区审查 PR。

支持的平台:
  - GitHub / GitHub Enterprise  (URL 包含 /pull/)
  - GitLab / 腾讯工蜂           (URL 包含 /merge_requests/)
  - Gitee                       (URL 包含 /pulls/)

示例:
  pr-worktree.sh https://github.com/owner/repo/pull/123
  pr-worktree.sh https://gitlab.com/owner/repo/-/merge_requests/456
  pr-worktree.sh https://git.code.tencent.com/owner/repo/merge_requests/789

前提条件:
  - 当前目录位于 git 仓库中
  - 已配置对应的 remote（通过 git remote add）
  - 已配置访问权限（SSH key 或 token）
EOF
    exit 0
}

# 解析 PR URL
# 设置: PR_ID, REF_PATH, URL_HOST, URL_REPO_PATH, PLATFORM
parse_pr_url() {
    local url="$1"

    # 提取 host
    URL_HOST=$(echo "$url" | sed -E 's|https?://([^/]+)/.*|\1|')

    # 提取 owner/repo 路径（用于精确匹配 remote）
    # 格式: host/owner/repo/...
    URL_REPO_PATH=$(echo "$url" | sed -E 's|https?://[^/]+/([^/]+/[^/]+).*|\1|' | sed 's|\.git$||')

    # GitHub: /pull/{id}
    if [[ "$url" =~ /pull/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="pull/${PR_ID}/head"
        PLATFORM="GitHub"
        return 0
    fi

    # GitLab/工蜂: /merge_requests/{id} 或 /-/merge_requests/{id}
    if [[ "$url" =~ /merge_requests/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="merge-requests/${PR_ID}/head"
        PLATFORM="GitLab"
        return 0
    fi

    # Gitee: /pulls/{id}
    if [[ "$url" =~ /pulls/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="pull/${PR_ID}/head"
        PLATFORM="Gitee"
        return 0
    fi

    error "无法解析 PR URL: $url\n支持的格式: /pull/{id}, /merge_requests/{id}, /pulls/{id}"
}

# 查找匹配的 remote
# 设置: REMOTE_NAME
find_matching_remote() {
    local url_host="$1"
    local url_repo_path="$2"

    # 优先精确匹配 owner/repo
    while read -r name url _; do
        # 去掉 .git 后缀进行匹配
        local clean_url=$(echo "$url" | sed 's|\.git$||')
        if [[ "$clean_url" == *"$url_host"*"$url_repo_path"* ]]; then
            REMOTE_NAME="$name"
            return 0
        fi
    done < <(git remote -v | grep "(fetch)")

    # 如果精确匹配失败，回退到只匹配 host
    while read -r name url _; do
        if [[ "$url" == *"$url_host"* ]]; then
            REMOTE_NAME="$name"
            warn "未找到精确匹配的 remote，使用第一个匹配 host 的: $name"
            return 0
        fi
    done < <(git remote -v | grep "(fetch)")

    error "找不到匹配的 remote\n请先添加 remote: git remote add <name> <url>\n当前配置的 remote:\n$(git remote -v)"
}

# 主函数
main() {
    local pr_url="$1"

    # 检查参数
    [[ -z "$pr_url" || "$pr_url" == "-h" || "$pr_url" == "--help" ]] && show_help

    # 检查是否在 git 仓库中
    git rev-parse --show-toplevel &>/dev/null || error "当前目录不是 git 仓库"

    # 获取项目信息
    PROJECT_ROOT=$(git rev-parse --show-toplevel)
    PROJECT_NAME=$(basename "$PROJECT_ROOT")
    WORKTREE_BASE="${PROJECT_ROOT}-worktrees"

    info "项目: $PROJECT_NAME"
    info "Worktree 目录: $WORKTREE_BASE"
    echo ""

    # 解析 PR URL
    parse_pr_url "$pr_url"
    info "平台: $PLATFORM"
    info "PR 编号: #$PR_ID"
    info "仓库: $URL_REPO_PATH"
    echo ""

    # 查找匹配的 remote
    find_matching_remote "$URL_HOST" "$URL_REPO_PATH"
    info "使用 Remote: $REMOTE_NAME"
    echo ""

    # 构建 worktree 路径
    LOCAL_BRANCH="pr-${PR_ID}"
    WORKTREE_PATH="${WORKTREE_BASE}/${LOCAL_BRANCH}"

    # 检查 worktree 是否已存在
    if [[ -d "$WORKTREE_PATH" ]]; then
        warn "Worktree 已存在: $WORKTREE_PATH"
        echo ""
        echo "进入现有 worktree:"
        echo -e "  ${YELLOW}cd $WORKTREE_PATH${NC}"
        exit 0
    fi

    # 检查本地分支是否已存在
    if git show-ref --verify --quiet "refs/heads/${LOCAL_BRANCH}"; then
        warn "本地分支 ${LOCAL_BRANCH} 已存在，将使用现有分支"
    else
        # Fetch PR 分支
        info "获取 PR 分支: ${REMOTE_NAME} ${REF_PATH}"
        if ! git fetch "$REMOTE_NAME" "${REF_PATH}:${LOCAL_BRANCH}" 2>&1; then
            error "无法获取 PR 分支\n可能原因:\n  1. PR 不存在或已被删除\n  2. 没有访问权限\n  3. refs 路径不正确"
        fi
        success "已获取 PR 分支"
    fi

    # 创建 worktree 目录
    mkdir -p "$WORKTREE_BASE"

    # 创建 worktree
    info "创建 Worktree: $WORKTREE_PATH"
    if ! git worktree add "$WORKTREE_PATH" "$LOCAL_BRANCH" 2>&1; then
        error "无法创建 worktree"
    fi
    success "Worktree 创建成功！"

    # 输出结果
    echo ""
    echo "========================================="
    echo -e "${GREEN}PR Worktree 已准备就绪${NC}"
    echo "========================================="
    echo ""
    echo "PR: #$PR_ID ($PLATFORM)"
    echo "分支: $LOCAL_BRANCH"
    echo "路径: $WORKTREE_PATH"
    echo ""
    echo "进入 worktree:"
    echo -e "  ${YELLOW}cd $WORKTREE_PATH${NC}"
    echo ""
    echo "清理 worktree:"
    echo "  git worktree remove $WORKTREE_PATH"
    echo "  git branch -D $LOCAL_BRANCH"
    echo ""
}

main "$@"
