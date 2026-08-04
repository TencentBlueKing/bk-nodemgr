#!/bin/bash
# scripts/verify-alignment.sh
# Verify that release version alignment changes meet fastpath requirements.
#
# Usage:
#   ./verify-alignment.sh <target-version> [base-branch]
#
# Example:
#   ./verify-alignment.sh v3.0.1-alpha.40
#   ./verify-alignment.sh v3.0.1-alpha.40 origin/master
#
# Exit codes:
#   0 - all checks passed
#   1 - file scope check failed
#   2 - version alignment check failed
#   3 - diff sanity check failed
#   4 - changelog check failed
#   5 - missing required argument
#   6 - gateway resource sync check failed

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TARGET_VERSION="${1:-}"
BASE_BRANCH="${2:-master}"

if [ -z "$TARGET_VERSION" ]; then
  echo -e "${RED}Error: target version is required${NC}"
  echo "Usage: $0 <target-version> [base-branch]"
  echo "Example: $0 v3.0.1-alpha.40"
  exit 5
fi

TARGET_VERSION_NO_V="${TARGET_VERSION#v}"

PRIMARY_FILES=(
  "apigw/definition.yaml"
  "apigw/resources.yaml"
  "install/helm/bk-nodemgr/Chart.yaml"
  "install/helm/bk-nodemgr/values.yaml"
  "install/helm/mock-server/Chart.yaml"
  "install/helm/mock-server/values.yaml"
)

HELM_FILES=(
  "install/helm/bk-nodemgr/Chart.yaml"
  "install/helm/bk-nodemgr/values.yaml"
  "install/helm/mock-server/Chart.yaml"
  "install/helm/mock-server/values.yaml"
)

APIGW_DEFINITION_PATH="apigw/definition.yaml"
APIGW_RESOURCES_PATH="apigw/resources.yaml"

is_in_list() {
  local needle=$1
  shift
  local item
  for item in "$@"; do
    if [ "$needle" = "$item" ]; then
      return 0
    fi
  done
  return 1
}

is_changelog_file() {
  local file=$1
  [[ "$file" =~ ^support-files/changelog/(en|zh)/${TARGET_VERSION}_[0-9]{4}-[0-9]{2}-[0-9]{2}\.md$ ]]
}

has_changed_file() {
  local file=$1
  printf '%s\n' "$CHANGED_FILES" | grep -qx "$file"
}

read_top_yaml_value() {
  local file=$1
  local key=$2
  awk -v key="$key" '$1 == key ":" { gsub(/["'\'' ]/, "", $2); print $2; exit }' "$file"
}

read_nested_yaml_value() {
  local file=$1
  local section=$2
  local key=$3
  awk -v section="$section" -v key="$key" '
    $0 ~ "^" section ":" { in_section=1; next }
    in_section && $0 ~ "^[^[:space:]]" { in_section=0 }
    in_section && $1 == key ":" {
      value=$2
      gsub(/["'\'' ]/, "", value)
      print value
      exit
    }
  ' "$file"
}

read_apigw_release_value() {
  local file=$1
  local key=$2
  awk -v key="$key" '
    /^release:/ { in_release=1; next }
    in_release && /^[^[:space:]]/ { in_release=0 }
    in_release && $1 == key ":" {
      value=$2
      gsub(/["'\'' ]/, "", value)
      print value
      exit
    }
  ' "$file"
}

read_values_apigw_release_value() {
  local file=$1
  local key=$2
  awk -v key="$key" '
    /^apigwSync:/ { in_apigw=1; next }
    in_apigw && /^[^#[:space:]]/ { in_apigw=0 }
    in_apigw && /^[[:space:]]+release:/ { in_release=1; next }
    in_release && /^[[:space:]]{2}[^[:space:]]/ && $1 != "release:" { in_release=0 }
    in_release && $1 == key ":" {
      value=$2
      gsub(/["'\'' ]/, "", value)
      print value
      exit
    }
  ' "$file"
}

values_apigw_release_changed() {
  git diff "$BASE_BRANCH" --unified=0 -- install/helm/bk-nodemgr/values.yaml |
    grep -Eq '^[+-][[:space:]]{6}(version|comment):'
}

check_equals() {
  local label=$1
  local actual=$2
  local expected=$3
  if [ "$actual" != "$expected" ]; then
    echo -e "${RED}Mismatch: $label expected $expected, got ${actual:-<empty>}${NC}"
    ERRORS=$((ERRORS + 1))
  fi
}

CHANGED_FILES=$(git diff "$BASE_BRANCH" --name-only | sort)

if [ -z "$CHANGED_FILES" ]; then
  echo -e "${YELLOW}No changed files against $BASE_BRANCH${NC}"
  exit 0
fi

MODE="helm-followup"
if printf '%s\n' "$CHANGED_FILES" | grep -Eq '^(apigw/|support-files/changelog/)'; then
  MODE="complete-release"
fi

APIGW_METADATA_CHANGED=false
if has_changed_file "$APIGW_DEFINITION_PATH" || values_apigw_release_changed; then
  APIGW_METADATA_CHANGED=true
fi

echo "Verifying release version alignment for $TARGET_VERSION against $BASE_BRANCH"
echo "Detected mode: $MODE"
echo ""

# Check 1: File scope check
echo "Check 1: File scope"
SCOPE_ERRORS=0
while IFS= read -r file; do
  [ -z "$file" ] && continue
  if [ "$MODE" = "complete-release" ]; then
    if ! is_in_list "$file" "${PRIMARY_FILES[@]}" && ! is_changelog_file "$file"; then
      echo -e "${RED}Unexpected complete-release file: $file${NC}"
      SCOPE_ERRORS=$((SCOPE_ERRORS + 1))
    fi
  else
    if ! is_in_list "$file" "${HELM_FILES[@]}"; then
      echo -e "${RED}Unexpected Helm-only follow-up file: $file${NC}"
      SCOPE_ERRORS=$((SCOPE_ERRORS + 1))
    fi
  fi
done <<< "$CHANGED_FILES"

if [ $SCOPE_ERRORS -gt 0 ]; then
  echo ""
  echo "Actual changed files:"
  echo "$CHANGED_FILES"
  exit 1
fi

echo -e "${GREEN}File scope check passed${NC}"
echo ""

# Check 2: Version alignment check
echo "Check 2: Version alignment"
ERRORS=0

if has_changed_file "install/helm/bk-nodemgr/Chart.yaml" || [ "$MODE" = "complete-release" ]; then
  check_equals "bk-nodemgr Chart.yaml version" \
    "$(read_top_yaml_value "install/helm/bk-nodemgr/Chart.yaml" "version")" "$TARGET_VERSION"
  check_equals "bk-nodemgr Chart.yaml appVersion" \
    "$(read_top_yaml_value "install/helm/bk-nodemgr/Chart.yaml" "appVersion")" "$TARGET_VERSION"
fi

if has_changed_file "install/helm/bk-nodemgr/values.yaml" || [ "$MODE" = "complete-release" ]; then
  BK_IMAGE_TAG=$(read_nested_yaml_value "install/helm/bk-nodemgr/values.yaml" "image" "tag")
  BK_API_MANAGER_TAG=$(read_nested_yaml_value "install/helm/bk-nodemgr/values.yaml" "apiManagerImage" "tag")

  if [ -n "$BK_IMAGE_TAG" ]; then
    check_equals "bk-nodemgr values.yaml image.tag" "$BK_IMAGE_TAG" "$TARGET_VERSION"
  fi
  if [ -n "$BK_API_MANAGER_TAG" ]; then
    check_equals "bk-nodemgr values.yaml apiManagerImage.tag" "$BK_API_MANAGER_TAG" "$TARGET_VERSION"
  fi
fi

if has_changed_file "install/helm/mock-server/Chart.yaml" || [ "$MODE" = "complete-release" ]; then
  check_equals "mock-server Chart.yaml version" \
    "$(read_top_yaml_value "install/helm/mock-server/Chart.yaml" "version")" "$TARGET_VERSION"
  check_equals "mock-server Chart.yaml appVersion" \
    "$(read_top_yaml_value "install/helm/mock-server/Chart.yaml" "appVersion")" "$TARGET_VERSION"
fi

if has_changed_file "install/helm/mock-server/values.yaml" || [ "$MODE" = "complete-release" ]; then
  MOCK_IMAGE_TAG=$(read_nested_yaml_value "install/helm/mock-server/values.yaml" "image" "tag")
  check_equals "mock-server values.yaml image.tag" "$MOCK_IMAGE_TAG" "$TARGET_VERSION"
fi

if [ "$MODE" = "complete-release" ]; then
  APIGW_VERSION=$(read_apigw_release_value "$APIGW_DEFINITION_PATH" "version")
  APIGW_COMMENT=$(read_apigw_release_value "$APIGW_DEFINITION_PATH" "comment")
  VALUES_APIGW_VERSION=$(read_values_apigw_release_value "install/helm/bk-nodemgr/values.yaml" "version")
  VALUES_APIGW_COMMENT=$(read_values_apigw_release_value "install/helm/bk-nodemgr/values.yaml" "comment")

  check_equals "apigw/definition.yaml release.version" "$APIGW_VERSION" "$TARGET_VERSION_NO_V"
  check_equals "apigw/definition.yaml release.comment" "$APIGW_COMMENT" "$TARGET_VERSION_NO_V"
  check_equals "bk-nodemgr values.yaml apigwSync.config.release.version" "$VALUES_APIGW_VERSION" "$TARGET_VERSION_NO_V"
  check_equals "bk-nodemgr values.yaml apigwSync.config.release.comment" "$VALUES_APIGW_COMMENT" "$TARGET_VERSION_NO_V"
fi

if [ $ERRORS -gt 0 ]; then
  echo -e "${RED}Version alignment check failed with $ERRORS error(s)${NC}"
  exit 2
fi

echo -e "${GREEN}Version alignment check passed${NC}"
echo ""

# Check 3: Diff sanity check
echo "Check 3: Diff sanity"
WARNINGS=0

HELM_DIFF=$(git diff "$BASE_BRANCH" --unified=0 -- install/helm/ || true)
SUSPICIOUS_PATTERNS=(
  "templates/"
  "charts/"
  "dependencies:"
  "condition:"
  "enabled:"
)

for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
  if echo "$HELM_DIFF" | grep -q "$pattern"; then
    echo -e "${YELLOW}Warning: Helm diff contains '$pattern' - may indicate non-version changes${NC}"
    WARNINGS=$((WARNINGS + 1))
  fi
done

if [ $WARNINGS -gt 0 ]; then
  echo "Review the diff manually: git diff $BASE_BRANCH -- install/helm/"
  exit 3
fi

echo -e "${GREEN}Diff sanity check passed${NC}"
echo ""

# Check 4: Changelog check
echo "Check 4: Changelog"
if [ "$MODE" = "complete-release" ]; then
  EN_CHANGELOG=$(find support-files/changelog/en -maxdepth 1 -type f -name "${TARGET_VERSION}_*.md" | head -1)
  ZH_CHANGELOG=$(find support-files/changelog/zh -maxdepth 1 -type f -name "${TARGET_VERSION}_*.md" | head -1)

  if [ -z "$EN_CHANGELOG" ] || [ -z "$ZH_CHANGELOG" ]; then
    echo -e "${RED}Changelog check failed: missing localized changelog files for $TARGET_VERSION${NC}"
    echo "Expected:"
    echo "  support-files/changelog/en/${TARGET_VERSION}_YYYY-MM-DD.md"
    echo "  support-files/changelog/zh/${TARGET_VERSION}_YYYY-MM-DD.md"
    exit 4
  fi

  if ! grep -q "## \[Version: $TARGET_VERSION\]" "$EN_CHANGELOG"; then
    echo -e "${RED}Changelog check failed: missing heading in $EN_CHANGELOG${NC}"
    exit 4
  fi
  if ! grep -q "## \[Version: $TARGET_VERSION\]" "$ZH_CHANGELOG"; then
    echo -e "${RED}Changelog check failed: missing heading in $ZH_CHANGELOG${NC}"
    exit 4
  fi

  echo -e "${GREEN}Changelog check passed${NC}"
else
  echo -e "${YELLOW}Skipped for Helm-only follow-up${NC}"
fi
echo ""

# Check 5: Gateway resource sync check
echo "Check 5: Gateway resource sync"
CURRENT_DIFF_FILES=$(git diff "$BASE_BRANCH" --name-only)
SWAGGER_PATH_RE='^docs/api/swagger/backend/api/v3/.*\.swagger\.json$'
CHANGED_SWAGGER_FILES=$(echo "$CURRENT_DIFF_FILES" | grep -E "$SWAGGER_PATH_RE" || true)

if [ -n "$CHANGED_SWAGGER_FILES" ] && ! echo "$CURRENT_DIFF_FILES" | grep -qx "$APIGW_RESOURCES_PATH"; then
  echo -e "${RED}Gateway resource sync check failed: backend swagger changed but $APIGW_RESOURCES_PATH did not${NC}"
  echo "Changed backend swagger files:"
  echo "$CHANGED_SWAGGER_FILES"
  exit 6
fi

if [ "$MODE" = "complete-release" ]; then
  if has_changed_file "$APIGW_RESOURCES_PATH"; then
    if [ "$APIGW_METADATA_CHANGED" != true ]; then
      echo -e "${RED}Gateway resource sync check failed: $APIGW_RESOURCES_PATH changed but API Gateway release metadata did not${NC}"
      echo "Update $APIGW_DEFINITION_PATH release.* and install/helm/bk-nodemgr/values.yaml apigwSync.config.release.* to ${TARGET_VERSION_NO_V}."
      exit 6
    fi
    echo -e "${GREEN}Gateway resource sync check passed in current diff${NC}"
  else
    echo -e "${YELLOW}No apigw/resources.yaml change in complete-release mode; confirm no gateway-visible swagger contract changed${NC}"
  fi
else
  echo -e "${YELLOW}Skipped current-diff gateway expansion for Helm-only follow-up${NC}"
  echo "If release-window swagger evidence requires $APIGW_RESOURCES_PATH, ask before expanding scope."
fi
echo ""

echo "------------------------------------------------------------"
echo -e "${GREEN}All verification checks passed${NC}"
echo "Summary:"
echo "  Target version: $TARGET_VERSION"
echo "  Base branch: $BASE_BRANCH"
echo "  Mode: $MODE"
echo "  Changed files:"
echo "$CHANGED_FILES" | sed 's/^/    - /'
echo "  Changelog: $([ "$MODE" = "complete-release" ] && echo "localized version files checked" || echo "skipped for Helm-only follow-up")"
echo "  Gateway resource sync: checked according to mode"
echo ""
echo "Ready to commit and create PR."
