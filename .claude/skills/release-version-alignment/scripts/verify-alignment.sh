#!/bin/bash
# scripts/verify-alignment.sh
# Verify that Helm version alignment changes meet fastpath requirements
#
# Usage:
#   ./verify-alignment.sh <target-version> [base-branch]
#
# Example:
#   ./verify-alignment.sh v3.0.1-alpha.17
#   ./verify-alignment.sh v3.0.1-alpha.17 origin/master
#
# Exit codes:
#   0 - all checks passed
#   1 - file scope check failed
#   2 - version alignment check failed
#   3 - diff sanity check failed
#   4 - changelog check failed
#   6 - gateway resource sync check failed
#   5 - missing required argument

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TARGET_VERSION="${1:-}"
BASE_BRANCH="${2:-master}"

if [ -z "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ Error: target version is required${NC}"
  echo "Usage: $0 <target-version> [base-branch]"
  echo "Example: $0 v3.0.1-alpha.17"
  exit 5
fi

echo "🔍 Verifying Helm version alignment for $TARGET_VERSION against $BASE_BRANCH"
echo ""

# Expected files
EXPECTED_FILES=(
  "install/helm/bk-nodemgr/Chart.yaml"
  "install/helm/bk-nodemgr/values.yaml"
  "install/helm/mock-server/Chart.yaml"
  "install/helm/mock-server/values.yaml"
)

# Check 1: File scope check
echo "📋 Check 1: File scope"
echo "Expected: only the 4 target Helm files"

CHANGED_FILES=$(git diff "$BASE_BRANCH" --name-only | sort)
EXPECTED_FILES_SORTED=$(printf '%s\n' "${EXPECTED_FILES[@]}" | sort)

if [ "$CHANGED_FILES" != "$EXPECTED_FILES_SORTED" ]; then
  echo -e "${RED}❌ File scope check failed${NC}"
  echo ""
  echo "Expected files:"
  printf '%s\n' "${EXPECTED_FILES[@]}"
  echo ""
  echo "Actual changed files:"
  echo "$CHANGED_FILES"
  exit 1
fi

echo -e "${GREEN}✅ File scope check passed${NC}"
echo ""

# Check 2: Version alignment check
echo "📋 Check 2: Version alignment"
echo "Expected: all version/appVersion/image.tag fields = $TARGET_VERSION"

ERRORS=0

# Helper function to extract YAML field value
get_yaml_value() {
  local file=$1
  local field=$2
  grep "^${field}:" "$file" | awk '{print $2}' | tr -d '"' | tr -d "'"
}

# Check bk-nodemgr Chart.yaml
BK_CHART_VERSION=$(get_yaml_value "install/helm/bk-nodemgr/Chart.yaml" "version")
BK_CHART_APPVERSION=$(get_yaml_value "install/helm/bk-nodemgr/Chart.yaml" "appVersion")

if [ "$BK_CHART_VERSION" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ bk-nodemgr Chart.yaml version mismatch: expected $TARGET_VERSION, got $BK_CHART_VERSION${NC}"
  ERRORS=$((ERRORS + 1))
fi

if [ "$BK_CHART_APPVERSION" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ bk-nodemgr Chart.yaml appVersion mismatch: expected $TARGET_VERSION, got $BK_CHART_APPVERSION${NC}"
  ERRORS=$((ERRORS + 1))
fi

# Check bk-nodemgr values.yaml
BK_VALUES_TAG=$(get_yaml_value "install/helm/bk-nodemgr/values.yaml" "tag")

if [ "$BK_VALUES_TAG" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ bk-nodemgr values.yaml image.tag mismatch: expected $TARGET_VERSION, got $BK_VALUES_TAG${NC}"
  ERRORS=$((ERRORS + 1))
fi

# Check mock-server Chart.yaml
MOCK_CHART_VERSION=$(get_yaml_value "install/helm/mock-server/Chart.yaml" "version")
MOCK_CHART_APPVERSION=$(get_yaml_value "install/helm/mock-server/Chart.yaml" "appVersion")

if [ "$MOCK_CHART_VERSION" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ mock-server Chart.yaml version mismatch: expected $TARGET_VERSION, got $MOCK_CHART_VERSION${NC}"
  ERRORS=$((ERRORS + 1))
fi

if [ "$MOCK_CHART_APPVERSION" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ mock-server Chart.yaml appVersion mismatch: expected $TARGET_VERSION, got $MOCK_CHART_APPVERSION${NC}"
  ERRORS=$((ERRORS + 1))
fi

# Check mock-server values.yaml
MOCK_VALUES_TAG=$(get_yaml_value "install/helm/mock-server/values.yaml" "tag")

if [ "$MOCK_VALUES_TAG" != "$TARGET_VERSION" ]; then
  echo -e "${RED}❌ mock-server values.yaml image.tag mismatch: expected $TARGET_VERSION, got $MOCK_VALUES_TAG${NC}"
  ERRORS=$((ERRORS + 1))
fi

if [ $ERRORS -gt 0 ]; then
  echo -e "${RED}❌ Version alignment check failed with $ERRORS error(s)${NC}"
  exit 2
fi

echo -e "${GREEN}✅ Version alignment check passed${NC}"
echo ""

# Check 3: Diff sanity check
echo "📋 Check 3: Diff sanity"
echo "Expected: only version/appVersion/tag lines changed, no template or dependency changes"

DIFF_OUTPUT=$(git diff "$BASE_BRANCH" install/helm/)

# Check for suspicious patterns that indicate non-version changes
SUSPICIOUS_PATTERNS=(
  "templates/"
  "charts/"
  "dependencies:"
  "repository:"
  "condition:"
  "enabled:"
)

WARNINGS=0
for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
  if echo "$DIFF_OUTPUT" | grep -q "$pattern"; then
    echo -e "${YELLOW}⚠️  Warning: diff contains '$pattern' - may indicate non-version changes${NC}"
    WARNINGS=$((WARNINGS + 1))
  fi
done

# Count changed lines (excluding +++ and --- lines)
CHANGED_LINES=$(echo "$DIFF_OUTPUT" | grep -E '^\+|^-' | grep -v -E '^\+\+\+|^---' | wc -l)

# Expect roughly 12 lines changed (2 per file: version + appVersion/tag)
# Allow some tolerance for formatting differences
if [ "$CHANGED_LINES" -gt 20 ]; then
  echo -e "${YELLOW}⚠️  Warning: $CHANGED_LINES lines changed (expected ~12) - may indicate extra changes${NC}"
  WARNINGS=$((WARNINGS + 1))
fi

if [ $WARNINGS -gt 0 ]; then
  echo -e "${YELLOW}⚠️  Diff sanity check completed with $WARNINGS warning(s)${NC}"
  echo "Review the diff manually to confirm only version fields changed:"
  echo "  git diff $BASE_BRANCH install/helm/"
  exit 3
fi

echo -e "${GREEN}✅ Diff sanity check passed${NC}"
echo ""

# Check 4: Changelog check
echo "📋 Check 4: Changelog"
echo "Expected: release.md contains an entry for $TARGET_VERSION"

RELEASE_MD="release.md"

if [ ! -f "$RELEASE_MD" ]; then
  echo -e "${RED}❌ Changelog check failed: release.md not found${NC}"
  echo ""
  echo "A version release must include a changelog entry in release.md."
  echo "Please generate the changelog using the changelog-doc skill:"
  echo "  - Ask the AI to generate changelog for $TARGET_VERSION"
  echo "  - Or manually add an entry to release.md"
  exit 4
fi

# Check if the target version exists in release.md
# Look for pattern: ## [Version: vX.Y.Z] or ## [Version: vX.Y.Z-alpha.N]
if ! grep -q "## \[Version: $TARGET_VERSION\]" "$RELEASE_MD"; then
  echo -e "${RED}❌ Changelog check failed: no entry for $TARGET_VERSION in release.md${NC}"
  echo ""
  echo "A version release must include a changelog entry in release.md."
  echo "Please generate the changelog using the changelog-doc skill:"
  echo "  - Ask the AI to generate changelog for $TARGET_VERSION"
  echo "  - Review and confirm the generated changelog"
  echo "  - Then re-run this verification"
  exit 4
fi

echo -e "${GREEN}✅ Changelog check passed${NC}"
echo ""

# Check 5: Gateway resource sync check
echo "📋 Check 5: Gateway resource sync"
echo "Expected: if backend swagger files changed in the release evidence window, apigw/resources.yaml may need the matching gateway-visible contract update"

CURRENT_DIFF_FILES=$(git diff "$BASE_BRANCH" --name-only)
SWAGGER_PATH_RE='^docs/api/swagger/backend/api/v3/.*\.swagger\.json$'
APIGW_RESOURCES_PATH="apigw/resources.yaml"
CHANGED_SWAGGER_FILES=$(echo "$CURRENT_DIFF_FILES" | grep -E "$SWAGGER_PATH_RE" || true)

if [ -n "$CHANGED_SWAGGER_FILES" ]; then
  if ! echo "$CURRENT_DIFF_FILES" | grep -qx "$APIGW_RESOURCES_PATH"; then
    echo -e "${RED}❌ Gateway resource sync check failed: backend swagger changed but $APIGW_RESOURCES_PATH did not${NC}"
    echo ""
    echo "Changed backend swagger files:"
    echo "$CHANGED_SWAGGER_FILES"
    echo ""
    echo "Review whether these changes affect API Gateway-visible contracts before creating the release PR."
    exit 6
  fi

  echo -e "${GREEN}✅ Gateway resource sync check passed in current diff${NC}"
else
  echo -e "${YELLOW}⚠️  Manual gateway resource sync review required${NC}"
  echo "Compare the previous release version with $TARGET_VERSION as release evidence:"
  echo "  - If any docs/api/swagger/backend/api/v3/*.swagger.json file changed in that version window,"
  echo "  - Then decide whether $APIGW_RESOURCES_PATH must also be updated in that version window."
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ All verification checks passed${NC}"
echo ""
echo "Summary:"
echo "  Target version: $TARGET_VERSION"
echo "  Base branch: $BASE_BRANCH"
echo "  Files changed: 4 (as expected)"
echo "  Version alignment: all fields match target version"
  echo "  Diff sanity: no suspicious patterns detected"
  echo "  Changelog: entry exists in release.md"
  echo "  Gateway resource sync: checked in current diff or explicitly left for manual release-window review"
echo ""
echo "Ready to commit and create PR."
