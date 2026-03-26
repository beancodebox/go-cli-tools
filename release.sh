#!/bin/bash

# ============================================================================
# GitHub Releases에 바이너리 업로드 스크립트
#
# 사용법:
#   bash release.sh v1.0.0
#
# 요구사항:
#   1. dist/ 디렉토리에 빌드된 바이너리 존재
#   2. GitHub 토큰: GITHUB_TOKEN 환경변수
#   3. git 저장소 설정됨
# ============================================================================

VERSION="${1:-}"
GITHUB_REPO="beancodebox/go-cli-tools"
TOKEN_FILE=".github-token"
DRAFT_MODE=false

# 바이너리 찾기 (도구별 dist 디렉토리)
find_release_dir() {
    # 루트 dist 디렉토리 우선, 없으면 tools/*/dist 찾기
    if [ -d "./dist" ]; then
        echo "./dist"
    else
        local dir=$(find ./tools -name "dist" -type d | head -1)
        if [ -n "$dir" ]; then
            echo "$dir"
        else
            echo ""
        fi
    fi
}

# --draft 옵션 처리
if [ "$2" = "--draft" ]; then
    DRAFT_MODE=true
fi

# ============================================================================
# 색상
# ============================================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================================
# 함수
# ============================================================================
log_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

log_success() {
    echo -e "${GREEN}✓ ${1}${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠ ${1}${NC}"
}

log_error() {
    echo -e "${RED}✗ ${1}${NC}"
}

# ============================================================================
# 검증
# ============================================================================
if [ -z "$VERSION" ]; then
    log_error "Usage: bash release.sh <version> [--draft]"
    echo ""
    echo "Examples:"
    echo "  bash release.sh v1.0.0              # Create public release"
    echo "  bash release.sh v1.0.0 --draft      # Create draft (for testing)"
    echo ""
    exit 1
fi

# 토큰 읽기 (파일 → 환경변수 → 없음)
if [ -f "$TOKEN_FILE" ]; then
    GITHUB_TOKEN=$(cat "$TOKEN_FILE" | tr -d '\n\r ')
    if [ -n "$GITHUB_TOKEN" ]; then
        log_info "Using token from: $TOKEN_FILE"
    fi
elif [ -n "$GITHUB_TOKEN" ]; then
    log_info "Using token from: GITHUB_TOKEN environment variable"
fi

if [ -z "$GITHUB_TOKEN" ]; then
    log_error "GitHub token not found"
    echo ""
    echo "Please provide your GitHub token in one of these ways:"
    echo ""
    echo "  Option 1: Create .github-token file"
    echo "    echo 'ghp_xxxxxxxxxxxx' > .github-token"
    echo "    chmod 600 .github-token"
    echo ""
    echo "  Option 2: Set environment variable"
    echo "    export GITHUB_TOKEN=ghp_xxxxxxxxxxxx"
    echo ""
    exit 1
fi

# RELEASE_DIR 동적 설정
RELEASE_DIR=$(find_release_dir)

if [ -z "$RELEASE_DIR" ] || [ ! -d "$RELEASE_DIR" ]; then
    log_error "Release directory not found"
    echo "Run 'make release-build VERSION=$VERSION' first"
    exit 1
fi

# ============================================================================
# 릴리스 생성/업데이트 및 파일 업로드
# ============================================================================
echo "=========================================="
echo "  GitHub Releases Uploader"
echo "=========================================="
echo ""

log_info "Version: $VERSION"
log_info "Repository: $GITHUB_REPO"
log_info "Release directory: $RELEASE_DIR"
echo ""

# 릴리스가 존재하는지 확인
log_info "Checking if release exists..."
release_response=$(curl -s \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$VERSION" 2>/dev/null || echo "")

if echo "$release_response" | grep -q "\"id\""; then
    log_warn "Release already exists: $VERSION"
    release_id=$(echo "$release_response" | grep '"id"' | head -1 | sed 's/.*"id": //;s/,.*//')
else
    # 새 릴리스 생성
    log_info "Creating new release: $VERSION"
    if [ "$DRAFT_MODE" = true ]; then
        log_warn "Draft mode enabled (not publicly visible)"
    fi

    create_response=$(curl -s -X POST \
      -H "Authorization: token $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      -d "{\"tag_name\":\"$VERSION\",\"name\":\"$VERSION\",\"draft\":$DRAFT_MODE,\"prerelease\":false}" \
      "https://api.github.com/repos/$GITHUB_REPO/releases")

    release_id=$(echo "$create_response" | grep '"id"' | head -1 | sed 's/.*"id": //;s/,.*//')

    if [ -z "$release_id" ]; then
        log_error "Failed to create release"
        echo "$create_response"
        exit 1
    fi
    log_success "Created release: $VERSION (ID: $release_id)"
fi

echo ""

# 바이너리 파일 업로드
log_info "Uploading binaries..."
uploaded=0
for binary in "$RELEASE_DIR"/*; do
    if [ ! -f "$binary" ]; then
        continue
    fi

    filename=$(basename "$binary")
    log_info "Uploading: $filename"

    upload_response=$(curl -s -X POST \
      -H "Authorization: token $GITHUB_TOKEN" \
      -H "Content-Type: application/octet-stream" \
      --data-binary "@$binary" \
      "https://uploads.github.com/repos/$GITHUB_REPO/releases/$release_id/assets?name=$filename" 2>/dev/null)

    if echo "$upload_response" | grep -q "\"id\""; then
        log_success "Uploaded: $filename"
        ((uploaded++))
    else
        log_warn "Failed to upload: $filename"
    fi
done

echo ""

if [ $uploaded -gt 0 ]; then
    log_success "Uploaded $uploaded file(s) successfully"
    echo ""
    echo "Release URL: https://github.com/$GITHUB_REPO/releases/tag/$VERSION"
else
    log_error "No files were uploaded"
    exit 1
fi
