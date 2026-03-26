.PHONY: help build test install clean

# ============================================================================
# 설정
# ============================================================================
TOOL := ccs
CMD_DIR := ./cmd/$(TOOL)
BIN_DIR := ../../bin
RELEASE_DIR ?= ../../dist
VERSION ?= dev

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
	@mkdir -p ~/.bash_completion.d
	@rm -f ~/.local/bin/$(TOOL)
	@go build -o ~/.local/bin/$(TOOL) $(CMD_DIR)
	@echo "✓ Installed: ~/.local/bin/$(TOOL)"
	@cp .bashrc.ccs ~/.bashrc.ccs
	@echo "✓ Installed: ~/.bashrc.ccs"
	@cp ccs-completion.sh ~/.bash_completion.d/ccs
	@chmod +x ~/.bash_completion.d/ccs
	@echo "✓ Installed: ~/.bash_completion.d/ccs"
	@echo ""
	@echo "Setup complete! Add to ~/.bashrc or ~/.zshrc:"
	@echo "  [ -f ~/.bashrc.ccs ] && source ~/.bashrc.ccs"
	@echo ""
	@echo "Then run: source ~/.bashrc"

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
	@echo "✓ Built release bundles for $(TOOL)"

release-build-%:
	@mkdir -p $(abspath $(RELEASE_DIR))/staging
	@echo "Building $(TOOL) for $*..."
	@$(platform-$*) go build \
		-o $(abspath $(RELEASE_DIR))/staging/$(TOOL) \
		$(CMD_DIR)
	@chmod +x $(abspath $(RELEASE_DIR))/staging/$(TOOL)
	@cp .bashrc.ccs $(abspath $(RELEASE_DIR))/staging/
	@cp ccs-completion.sh $(abspath $(RELEASE_DIR))/staging/
	@cd $(abspath $(RELEASE_DIR))/staging && tar -czf ../$(TOOL)-$(VERSION)-$*.tar.gz . && cd ..
	@rm -rf $(abspath $(RELEASE_DIR))/staging
	@echo "  ✓ $(TOOL)-$(VERSION)-$*.tar.gz"
