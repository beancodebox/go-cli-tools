#!/bin/bash
set -e

# ============================================================================
# go-cli-tools 설치 스크립트 (GitHub Releases 기반)
#
# 사용법:
#   curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh          # 전체 설치
#   curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh -s cw    # cw만 설치
#   ./install.sh cw                                                                                      # 로컬 실행
# ============================================================================

GITHUB_REPO="beancodebox/go-cli-tools"
TOOLS_AVAILABLE="cw ccs"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CACHE_DIR="${CACHE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/go-cli-tools}"

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
    echo -e "${BLUE}ℹ ${1}${NC}" >&2
}

log_success() {
    echo -e "${GREEN}✓ ${1}${NC}" >&2
}

log_warn() {
    echo -e "${YELLOW}⚠ ${1}${NC}" >&2
}

log_error() {
    echo -e "${RED}✗ ${1}${NC}" >&2
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
    if command -v curl >/dev/null 2>&1; then
        curl -s "$releases_url" < /dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "//;s/".*//'
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O - "$releases_url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "//;s/".*//'
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# ============================================================================
# 함수: 바이너리 다운로드 (tar.gz)
# ============================================================================
download_binary() {
    local tool=$1
    local version=$2
    local platform=$3

    local package_name="${tool}-${version}-${platform}.tar.gz"
    local download_url="https://github.com/$GITHUB_REPO/releases/download/$version/$package_name"
    local cache_path="$CACHE_DIR/$package_name"

    # 캐시 디렉토리 생성
    mkdir -p "$CACHE_DIR"

    # 이미 캐시에 있으면 압축 해제된 디렉토리 반환
    if [ -f "$cache_path" ]; then
        log_success "Using cached: $package_name"
    else
        log_info "Downloading $package_name from GitHub Releases..."

        if command -v curl >/dev/null 2>&1; then
            if ! curl -fsSL -o "$cache_path" "$download_url" < /dev/null; then
                log_error "Failed to download: $download_url"
                return 1
            fi
        elif command -v wget >/dev/null 2>&1; then
            if ! wget -q -O "$cache_path" "$download_url"; then
                log_error "Failed to download: $download_url"
                return 1
            fi
        else
            log_error "Neither curl nor wget found"
            return 1
        fi

        log_success "Downloaded: $package_name"
    fi

    # 압축 해제
    local staging_dir="$CACHE_DIR/staging-${tool}-${version}-${platform}"
    mkdir -p "$staging_dir"
    if ! tar -xzf "$cache_path" -C "$staging_dir"; then
        log_error "Failed to extract: $package_name"
        return 1
    fi

    echo "$staging_dir"
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

    # ccs는 Windows 미지원
    if [ "$tool" = "ccs" ]; then
        if [ "$(uname -s)" = "MINGW64_NT" ] || [ "$(uname -s)" = "MINGW32_NT" ]; then
            log_error "ccs does not support Windows (shell wrapper required)"
            return 1
        fi
    fi

    log_info "Installing $tool ($version)..."

    # 바이너리 다운로드 + 압축 해제
    local staging_dir
    staging_dir=$(download_binary "$tool" "$version" "$platform") || return 1

    # 설치 디렉토리 생성
    mkdir -p "$INSTALL_DIR"
    mkdir -p ~/.bash_completion.d 2>/dev/null || true

    # 1. 바이너리 설치
    cp "$staging_dir/$tool" "$INSTALL_DIR/$tool"
    chmod +x "$INSTALL_DIR/$tool"
    log_success "Installed binary: $INSTALL_DIR/$tool"

    # 2. Shell wrapper 설치 (.bashrc.TOOL)
    if [ -f "$staging_dir/.bashrc.$tool" ]; then
        cp "$staging_dir/.bashrc.$tool" ~/.bashrc.$tool
        log_success "Installed shell wrapper: ~/.bashrc.$tool"
    fi

    # 3. Bash completion 설치
    if [ -f "$staging_dir/$tool-completion.sh" ]; then
        cp "$staging_dir/$tool-completion.sh" ~/.bash_completion.d/$tool
        chmod +x ~/.bash_completion.d/$tool
        log_success "Installed completion: ~/.bash_completion.d/$tool"
    fi

    # 4. Shell wrapper를 ~/.bashrc와 ~/.zshrc에 자동 추가
    if [ -f "$staging_dir/.bashrc.$tool" ]; then
        # ~/.bashrc에 추가
        if ! grep -q "source ~/.bashrc.$tool" ~/.bashrc 2>/dev/null; then
            echo "" >> ~/.bashrc
            echo "# $tool configuration" >> ~/.bashrc
            echo "[ -f ~/.bashrc.$tool ] && source ~/.bashrc.$tool" >> ~/.bashrc
            log_success "Added to ~/.bashrc"
        fi

        # zsh 사용자: ~/.zshrc에도 추가
        if [ -f ~/.zshrc ] || [ -n "$ZSH_VERSION" ] || echo "$SHELL" | grep -q "zsh"; then
            if ! grep -q "source ~/.bashrc.$tool" ~/.zshrc 2>/dev/null; then
                echo "" >> ~/.zshrc
                echo "# $tool configuration" >> ~/.zshrc
                echo "[ -f ~/.bashrc.$tool ] && source ~/.bashrc.$tool" >> ~/.zshrc
                log_success "Added to ~/.zshrc (restart shell to apply)"
            fi
        fi
    fi
}

# ============================================================================
# 함수: PATH 확인
# ============================================================================
check_path() {
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
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
  ./install.sh              Install all tools
  ./install.sh cw           Install cw tool only
  ./install.sh cw ccs       Install multiple tools
  ./install.sh --help       Show this help message

Web usage:
  curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh

Available tools:
  $TOOLS_AVAILABLE

Examples:
  ./install.sh              # Install all tools
  ./install.sh cw           # Install cw only
  ./install.sh cw ccs       # Install cw and ccs
EOF
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

    # 설치할 도구와 버전 결정
    local tools_to_install=""
    local specified_version=""

    # 마지막 인자가 v?.?.? 형식이면 버전으로 취급
    if [ $# -gt 0 ]; then
        local last_arg="${!#}"
        if [[ "$last_arg" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
            specified_version="$last_arg"
            # 버전을 제외한 나머지를 도구로 사용
            set -- "${@:1:$(($# - 1))}"
        fi
    fi

    if [ $# -eq 0 ]; then
        tools_to_install="$TOOLS_AVAILABLE"
    else
        tools_to_install="$@"
    fi

    # 버전 결정: 지정된 버전 또는 최신
    local version
    if [ -n "$specified_version" ]; then
        version="$specified_version"
        log_success "Using specified version: $version"
    else
        log_info "Fetching latest release..."
        version=$(get_latest_release)
        if [ -z "$version" ]; then
            log_error "Failed to get latest release"
            exit 1
        fi
        log_success "Latest version: $version"
    fi
    echo ""

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
