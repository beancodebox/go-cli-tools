.PHONY: help build test install clean

# ============================================================================
# 설정
# ============================================================================
TOOL := cw
CMD_DIR := ./cmd/$(TOOL)
BIN_DIR := ../../bin
RELEASE_DIR ?= ../../dist
VERSION ?= dev

# XDG Base Directory 표준
XDG_DATA_HOME ?= $(HOME)/.local/share
BASH_COMPLETION_DIR = $(XDG_DATA_HOME)/bash-completion.d
ZSH_SITE_FUNCTIONS_DIR = $(XDG_DATA_HOME)/zsh/site-functions

# Cross-compile 플랫폼들
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

# 플랫폼별 GOOS, GOARCH
platform-linux-amd64 := GOOS=linux GOARCH=amd64
platform-linux-arm64 := GOOS=linux GOARCH=arm64
platform-darwin-amd64 := GOOS=darwin GOARCH=amd64
platform-darwin-arm64 := GOOS=darwin GOARCH=arm64
platform-windows-amd64 := GOOS=windows GOARCH=amd64

# ============================================================================
# 도움말
# ============================================================================
help:
	@echo "$(TOOL) Makefile targets:"
	@echo "  make build    - Build $(TOOL) binary to ../../bin/"
	@echo "  make test     - Run tests for $(TOOL)"
	@echo "  make install  - Install $(TOOL) to ~/.local/bin"
	@echo "  make clean    - Remove build artifacts"

# ============================================================================
# 빌드
# ============================================================================
build:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(TOOL) $(CMD_DIR)
	@echo "✓ Built: $(BIN_DIR)/$(TOOL)"

# ============================================================================
# 테스트
# ============================================================================
test:
	@echo "Testing $(TOOL)..."
	@go test $(CMD_DIR)/...
	@echo "✓ Tests passed"

# ============================================================================
# 설치 (운영 배포)
# ============================================================================
install:
	@mkdir -p ~/.local/bin
	@mkdir -p $(BASH_COMPLETION_DIR)
	@mkdir -p $(ZSH_SITE_FUNCTIONS_DIR)
	@rm -f ~/.local/bin/$(TOOL)
	@go build -o ~/.local/bin/$(TOOL) $(CMD_DIR)
	@echo "✓ Installed: ~/.local/bin/$(TOOL)"

	@cp .bashrc.cw ~/.bashrc.cw
	@echo "✓ Installed: ~/.bashrc.cw"

	@# ~/.bashrc에 자동 등록
	@if ! grep -q "source ~/.bashrc.cw" ~/.bashrc 2>/dev/null; then \
		echo "" >> ~/.bashrc; \
		echo "# cw configuration" >> ~/.bashrc; \
		echo "[ -f ~/.bashrc.cw ] && source ~/.bashrc.cw" >> ~/.bashrc; \
		echo "✓ Added to ~/.bashrc"; \
	else \
		echo "✓ Already in ~/.bashrc"; \
	fi

	@# ~/.zshrc에 자동 등록 (zsh 사용자)
	@if [ -f ~/.zshrc ]; then \
		if ! grep -q "source ~/.bashrc.cw" ~/.zshrc 2>/dev/null; then \
			echo "" >> ~/.zshrc; \
			echo "# cw configuration" >> ~/.zshrc; \
			echo "[ -f ~/.bashrc.cw ] && source ~/.bashrc.cw" >> ~/.zshrc; \
			echo "✓ Added to ~/.zshrc"; \
		else \
			echo "✓ Already in ~/.zshrc"; \
		fi; \
	fi

	@cp cw-completion.bash $(BASH_COMPLETION_DIR)/cw
	@chmod 755 $(BASH_COMPLETION_DIR)/cw
	@echo "✓ Installed Bash completion: $(BASH_COMPLETION_DIR)/cw"

	@cp _cw $(ZSH_SITE_FUNCTIONS_DIR)/_cw
	@chmod 644 $(ZSH_SITE_FUNCTIONS_DIR)/_cw
	@echo "✓ Installed Zsh completion: $(ZSH_SITE_FUNCTIONS_DIR)/_cw"

	@echo ""
	@echo "Tip: Add ~/.local/bin to your PATH"
	@echo "Completion files installed to:"
	@echo "  Bash: $(BASH_COMPLETION_DIR)/cw"
	@echo "  Zsh:  $(ZSH_SITE_FUNCTIONS_DIR)/_cw"

# ============================================================================
# 청소
# ============================================================================
clean:
	@rm -f $(BIN_DIR)/$(TOOL)
	@rm -f $(RELEASE_DIR)/$(TOOL)-*
	@echo "✓ Cleaned: $(BIN_DIR)/$(TOOL) and $(RELEASE_DIR)/$(TOOL)-*"

# ============================================================================
# 릴리스 빌드 (모든 플랫폼)
# ============================================================================
release-build: $(addprefix release-build-,$(PLATFORMS))
	@echo "✓ Built release binaries for $(TOOL)"

release-build-%:
	@mkdir -p $(abspath $(RELEASE_DIR))/staging
	@echo "Building $(TOOL) for $*..."
	@$(platform-$*) go build \
		-o $(abspath $(RELEASE_DIR))/staging/$(TOOL) \
		$(CMD_DIR)
	@chmod +x $(abspath $(RELEASE_DIR))/staging/$(TOOL)
	@cp .bashrc.cw $(abspath $(RELEASE_DIR))/staging/
	@cp cw-completion.bash $(abspath $(RELEASE_DIR))/staging/
	@cp _cw $(abspath $(RELEASE_DIR))/staging/
	@cd $(abspath $(RELEASE_DIR))/staging && tar -czf ../$(TOOL)-$(VERSION)-$*.tar.gz . && cd ..
	@rm -rf $(abspath $(RELEASE_DIR))/staging
	@echo "  ✓ $(TOOL)-$(VERSION)-$*.tar.gz"
