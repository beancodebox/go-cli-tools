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
	@echo "  make build          - Build cw for current platform"
	@echo "  make build-all      - Build all tools for current platform"
	@echo "  make test           - Run all tests"
	@echo "  make install        - Install cw to ~/.local/bin"
	@echo "  make install-all    - Install all tools to ~/.local/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Release targets:"
	@echo "  make release-build          - Build all tools for all platforms"
	@echo "  make release-publish        - Publish public release to GitHub"
	@echo "  make release-publish-draft  - Publish draft release (for testing)"
	@echo ""
	@echo "Examples:"
	@echo "  make release-build VERSION=v1.0.0"
	@echo "  make release-publish-draft VERSION=v1.0.0  # Test (not public)"
	@echo "  make release-publish VERSION=v1.0.0        # Official release"

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
# 릴리스 빌드 (모든 플랫폼)
# ============================================================================
release-build: $(addprefix release-build-,$(TOOLS))
	@echo "✓ Release builds complete:"
	@find . -name "*-$(VERSION)-*" -type f 2>/dev/null | sort

release-build-%:
	@echo "Building release binaries for $*..."
	@$(MAKE) -C tools/$* release-build RELEASE_DIR=$(RELEASE_DIR) VERSION=$(VERSION)

# ============================================================================
# 릴리스 발행 (GitHub Releases)
# ============================================================================
release-publish:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required: make release-publish VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if [ ! -f release.sh ]; then \
		echo "❌ release.sh not found"; \
		exit 1; \
	fi
	@echo "Publishing release $(VERSION) to GitHub..."
	@bash release.sh $(VERSION)
	@echo "✓ Release published: $(VERSION)"

# 드래프트 릴리스 (테스트용, 비공개)
release-publish-draft:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required: make release-publish-draft VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if [ ! -f release.sh ]; then \
		echo "❌ release.sh not found"; \
		exit 1; \
	fi
	@echo "Publishing DRAFT release $(VERSION) to GitHub..."
	@bash release.sh $(VERSION) --draft
	@echo "✓ Draft release published: $(VERSION) (not publicly visible)"
