.PHONY: help build build-all test install install-all clean

# ============================================================================
# 도구 목록 (새 도구 추가시 여기에만 추가하면 됨)
# ============================================================================
TOOLS := cw ccs
BIN_DIR := ./bin
RELEASE_DIR := ./dist
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

# ============================================================================
# 도움말
# ============================================================================
help:
	@echo "Development targets:"
	@echo "  make build          - Build current platform"
	@echo "  make build-all      - Build all tools for current platform"
	@echo "  make test           - Run all tests"
	@echo "  make install        - Install cw to ~/.local/bin"
	@echo "  make install-all    - Install all tools to ~/.local/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Release targets:"
	@echo "  make build VERSION=v0.1.0      - Build for all platforms (tools/*/dist/)"
	@echo "  make release VERSION=v0.1.0    - Full release (build → dist → GitHub publish)"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=v0.1.0      # Build only"
	@echo "  make release VERSION=v0.1.0    # Build + dist + publish to GitHub"

# ============================================================================
# 빌드: 모든 도구
# ============================================================================
build-all: $(addprefix build-,$(TOOLS))
	@echo "✓ All tools built to ./bin/"

# 동적 빌드 타겟 (각 도구)
build-%:
	@echo "Building $*..."
	@$(MAKE) -C tools/$* build

# ============================================================================
# 빌드: 주요 도구들 (자주 사용하는 것)
# ============================================================================
build: build-cw

# ============================================================================
# 테스트
# ============================================================================
test: $(addprefix test-,$(TOOLS))
	@echo "✓ All tests passed"

test-%:
	@echo "Testing $*..."
	@$(MAKE) -C tools/$* test

# ============================================================================
# 설치: 모든 도구
# ============================================================================
install-all: $(addprefix install-,$(TOOLS))
	@echo "✓ All tools installed to ~/.local/bin"
	@echo "Tip: Add ~/.local/bin to your PATH if not already done"

# 동적 설치 타겟 (각 도구)
install-%:
	@echo "Installing $*..."
	@$(MAKE) -C tools/$* install

# ============================================================================
# 설치: 주요 도구 (자주 사용하는 것)
# ============================================================================
install: install-cw

# ============================================================================
# 청소
# ============================================================================
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -rf $(RELEASE_DIR)
	@echo "✓ Clean complete"

# ============================================================================
# 빌드 (모든 플랫폼)
# ============================================================================
build: VERSION?=
build:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION required: make build VERSION=v0.1.0"; \
		exit 1; \
	fi
	@for tool in $(TOOLS); do \
		$(MAKE) -C tools/$$tool release-build VERSION=$(VERSION); \
	done
	@echo ""
	@echo "✓ Build complete. Files in tools/*/dist/"

# ============================================================================
# 배포 (빌드 → dist 복사 → GitHub Release)
# ============================================================================
release: VERSION?=
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION required: make release VERSION=v0.1.0"; \
		exit 1; \
	fi
	@$(MAKE) build VERSION=$(VERSION)
	@echo ""
	@echo "Copying to $(RELEASE_DIR)..."
	@cp tools/*/dist/*-$(VERSION)-* $(RELEASE_DIR)/ 2>/dev/null || true
	@echo "✓ Copied to $(RELEASE_DIR)"
	@echo ""
	@echo "Publishing to GitHub..."
	@bash release.sh $(VERSION)
	@echo "✓ Release published: $(VERSION)"
