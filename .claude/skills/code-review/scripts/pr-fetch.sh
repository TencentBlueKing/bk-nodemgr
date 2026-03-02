#!/bin/bash
# PR Branch Fetcher
# Fetches a PR branch to a local ref WITHOUT switching the current branch.
# Worktree creation is handled separately by the using-git-worktrees skill.
#
# Supported platforms:
#   - GitHub / GitHub Enterprise  (/pull/<id>)
#   - GitLab / Tencent Coding     (/merge_requests/<id>)
#   - Gitee                       (/pulls/<id>)
#
# Usage: pr-fetch.sh <PR_URL>
# Output: prints the local branch name (pr-<id>) on success

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $1" >&2; }
success() { echo -e "${GREEN}[OK]${NC} $1" >&2; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1" >&2; }
error()   { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }

show_help() {
    cat << 'EOF'
Usage: pr-fetch.sh <PR_URL>

Fetches a PR branch to a local ref without switching the current branch.
Prints the local branch name to stdout on success.

Supported platforms:
  GitHub / GitHub Enterprise  (URL contains /pull/)
  GitLab / Tencent Coding     (URL contains /merge_requests/)
  Gitee                       (URL contains /pulls/)

Examples:
  pr-fetch.sh https://github.com/owner/repo/pull/123
  pr-fetch.sh https://gitlab.com/owner/repo/-/merge_requests/456
  pr-fetch.sh https://git.code.tencent.com/owner/repo/merge_requests/789

After running, use the using-git-worktrees skill to create an isolated worktree
from the fetched branch.
EOF
    exit 0
}

parse_pr_url() {
    local url="$1"

    URL_HOST=$(echo "$url" | sed -E 's|https?://([^/]+)/.*|\1|')
    URL_REPO_PATH=$(echo "$url" | sed -E 's|https?://[^/]+/([^/]+/[^/]+).*|\1|' | sed 's|\.git$||')

    if [[ "$url" =~ /pull/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="pull/${PR_ID}/head"
        PLATFORM="GitHub"
        return 0
    fi

    if [[ "$url" =~ /merge_requests/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="merge-requests/${PR_ID}/head"
        PLATFORM="GitLab"
        return 0
    fi

    if [[ "$url" =~ /pulls/([0-9]+) ]]; then
        PR_ID="${BASH_REMATCH[1]}"
        REF_PATH="pull/${PR_ID}/head"
        PLATFORM="Gitee"
        return 0
    fi

    error "Cannot parse PR URL: $url\nSupported formats: /pull/<id>, /merge_requests/<id>, /pulls/<id>"
}

find_matching_remote() {
    local url_host="$1"
    local url_repo_path="$2"

    # Prefer exact owner/repo match
    while read -r name url _; do
        local clean_url=$(echo "$url" | sed 's|\.git$||')
        if [[ "$clean_url" == *"$url_host"*"$url_repo_path"* ]]; then
            REMOTE_NAME="$name"
            return 0
        fi
    done < <(git remote -v | grep "(fetch)")

    # Fall back to host-only match
    while read -r name url _; do
        if [[ "$url" == *"$url_host"* ]]; then
            REMOTE_NAME="$name"
            warn "No exact remote match; using first host match: $name"
            return 0
        fi
    done < <(git remote -v | grep "(fetch)")

    error "No matching remote found for $url_host/$url_repo_path\nConfigured remotes:\n$(git remote -v)"
}

main() {
    local pr_url="$1"
    [[ -z "$pr_url" || "$pr_url" == "-h" || "$pr_url" == "--help" ]] && show_help

    git rev-parse --show-toplevel &>/dev/null || error "Not inside a git repository"

    parse_pr_url "$pr_url"
    info "Platform: $PLATFORM | PR: #$PR_ID"

    find_matching_remote "$URL_HOST" "$URL_REPO_PATH"
    info "Remote: $REMOTE_NAME"

    LOCAL_BRANCH="pr-${PR_ID}"

    if git show-ref --verify --quiet "refs/heads/${LOCAL_BRANCH}"; then
        warn "Local branch ${LOCAL_BRANCH} already exists, skipping fetch"
    else
        info "Fetching ${REMOTE_NAME} ${REF_PATH} -> ${LOCAL_BRANCH}"
        git fetch "$REMOTE_NAME" "${REF_PATH}:${LOCAL_BRANCH}" \
            || error "Fetch failed. Check PR existence, access rights, and remote refs path."
        success "Branch fetched: ${LOCAL_BRANCH}"
    fi

    # Print branch name to stdout so callers can capture it
    echo "$LOCAL_BRANCH"
}

main "$@"
