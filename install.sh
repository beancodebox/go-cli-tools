#!/bin/bash
set -e

# ============================================================================
# go-cli-tools 설치 스크립트 (GitHub Releases 기반)
#
# 사용법:
#   sh | https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh         # 대화형
#   sh | https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh cw      # cw 설치
#   ./install.sh cw                                                                          # 로컬 실행
# ============================================================================

GITHUB_REPO="beancodebox/go-cli-tools"
TOOLS_AVAILABLE="cw"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/go-cli-tools"

# ============================================================================
# 색상 정의
# ============================================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================================
# 함수: 메시지 출력
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
# 함수: OS와 아키텍처 감지
# ============================================================================
detect_platform() {
    local os=""
    local arch=""

    case "$(uname -s)" in
        Linux*)
            os="linux"
            ;;
        Darwin*)
            os="darwin"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        *)
            log_error "Unsupported OS: $(uname -s)"
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# ============================================================================
# 함수: GitHub Releases에서 최신 버전 가져오기
# ============================================================================
get_latest_release() {
    local releases_url="https://api.github.com/repos/$GITHUB_REPO/releases?per_page=10"

    # curl이 없으면 wget 사용
    if command -v curl &> /dev/null; then
        curl -s "$releases_url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "//;s/".*//'
    elif command -v wget &> /dev/null; then
        wget -q -O - "$releases_url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "//;s/".*//'
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# ============================================================================
# 함수: 바이너리 다운로드
# ============================================================================
download_binary() {
    local tool=$1
    local version=$2
    local platform=$3

    local binary_name="${tool}-${version}-${platform}"
    if [ "$(uname -s)" = "MINGW64_NT" ] || [ "$(uname -s)" = "MINGW32_NT" ]; then
        binary_name="${binary_name}.exe"
    fi

    local download_url="https://github.com/$GITHUB_REPO/releases/download/$version/$binary_name"
    local cache_path="$CACHE_DIR/$binary_name"

    # 캐시 디렉토리 생성
    mkdir -p "$CACHE_DIR"

    # 이미 캐시에 있으면 사용
    if [ -f "$cache_path" ]; then
        log_success "Using cached: $binary_name"
        echo "$cache_path"
        return 0
    fi

    log_info "Downloading $binary_name from GitHub Releases..."

    if command -v curl &> /dev/null; then
        if ! curl -fsSL -o "$cache_path" "$download_url"; then
            log_error "Failed to download: $download_url"
            return 1
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q -O "$cache_path" "$download_url"; then
            log_error "Failed to download: $download_url"
            return 1
        fi
    else
        log_error "Neither curl nor wget found"
        return 1
    fi

    chmod +x "$cache_path"
    log_success "Downloaded: $binary_name"
    echo "$cache_path"
}

# ============================================================================
# 함수: 도구 설치
# ============================================================================
install_tool() {
    local tool=$1
    local version=$2
    local platform=$3

    # 유효성 확인
    if ! echo "$TOOLS_AVAILABLE" | grep -q "$tool"; then
        log_error "Unknown tool: $tool"
        return 1
    fi

    log_info "Installing $tool ($version)..."

    # 바이너리 다운로드
    local binary_path
    binary_path=$(download_binary "$tool" "$version" "$platform") || return 1

    # 설치 디렉토리 생성
    mkdir -p "$INSTALL_DIR"

    # 설치
    local install_path="$INSTALL_DIR/$tool"
    if [ "$(uname -s)" = "MINGW64_NT" ] || [ "$(uname -s)" = "MINGW32_NT" ]; then
        install_path="${install_path}.exe"
    fi

    cp "$binary_path" "$install_path"
    chmod +x "$install_path"

    log_success "Installed: $install_path"
}

# ============================================================================
# 함수: PATH 확인
# ============================================================================
check_path() {
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        return 0
    else
        return 1
    fi
}

# ============================================================================
# 함수: PATH 추가 안내
# ============================================================================
suggest_add_to_path() {
    if ! check_path; then
        log_warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add this to your shell config (~/.bashrc, ~/.zshrc, etc):"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
    fi
}

# ============================================================================
# 함수: 도움말
# ============================================================================
show_help() {
    cat << EOF
go-cli-tools Installer

Usage:
  ./install.sh              Install interactively (choose tool)
  ./install.sh cw           Install cw tool
  ./install.sh cw xxx       Install multiple tools
  ./install.sh --help       Show this help message
  ./install.sh --version    Show version info

Web usage:
  sh | curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh

Available tools:
  $TOOLS_AVAILABLE

Examples:
  ./install.sh              # Choose tools interactively
  ./install.sh cw           # Install cw from latest release
EOF
}

# ============================================================================
# 함수: 대화형 모드
# ============================================================================
interactive_mode() {
    echo ""
    log_info "Interactive installation"
    echo ""
    echo "Available tools:"
    local i=1
    for tool in $TOOLS_AVAILABLE; do
        echo "  $i) $tool"
        ((i++))
    done
    echo ""

    read -p "Select tools (space-separated numbers or names, e.g. '1' or 'cw'): " selection

    if [ -z "$selection" ]; then
        log_warn "No tools selected"
        exit 0
    fi

    local tools_to_install=""
    for item in $selection; do
        if [[ "$item" =~ ^[0-9]+$ ]]; then
            local idx=1
            for tool in $TOOLS_AVAILABLE; do
                if [ "$idx" -eq "$item" ]; then
                    tools_to_install="$tools_to_install $tool"
                    break
                fi
                ((idx++))
            done
        else
            if echo "$TOOLS_AVAILABLE" | grep -q "$item"; then
                tools_to_install="$tools_to_install $item"
            else
                log_error "Unknown tool: $item"
            fi
        fi
    done

    echo "$tools_to_install"
}

# ============================================================================
# 메인 로직
# ============================================================================
main() {
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        show_help
        exit 0
    fi

    echo "=========================================="
    echo "  go-cli-tools Installer"
    echo "=========================================="
    echo ""

    # 플랫폼 감지
    local platform
    platform=$(detect_platform)
    log_success "Detected platform: $platform"

    # 최신 버전 가져오기
    log_info "Fetching latest release..."
    local version
    version=$(get_latest_release)

    if [ -z "$version" ]; then
        log_error "Failed to get latest release"
        exit 1
    fi
    log_success "Latest version: $version"
    echo ""

    # 설치할 도구 결정
    local tools_to_install=""

    if [ $# -eq 0 ]; then
        tools_to_install=$(interactive_mode)
    else
        tools_to_install="$@"
    fi

    if [ -z "$tools_to_install" ]; then
        log_warn "No tools to install"
        exit 0
    fi

    echo ""

    # 각 도구 설치
    local failed=0
    for tool in $tools_to_install; do
        if ! install_tool "$tool" "$version" "$platform"; then
            failed=$((failed + 1))
        fi
    done

    echo ""

    # 결과 요약
    if [ $failed -eq 0 ]; then
        log_success "Installation complete!"
        echo ""
        suggest_add_to_path
    else
        log_error "Installation failed for $failed tool(s)"
        exit 1
    fi
}

main "$@"
