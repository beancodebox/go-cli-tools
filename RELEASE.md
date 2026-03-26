# Release & Distribution

go-cli-tools의 릴리즈 및 배포 전략 문서입니다.

---

## 📦 Release 구조

### **컴포넌트 분류**

| 타입 | 관리 방식 | 예시 |
|------|---------|------|
| **메인 프로그램** | GitHub Releases (버전 고정) | ccs, cw 바이너리 |
| **보조 파일** | GitHub Releases 번들 포함 (tar.gz) | .bashrc.ccs, ccs-completion.sh |
| **설정 문서** | GitHub Raw (main 최신) | README.md, 설치 문서 |

---

## 🏗️ Release 번들 구조

### **각 도구별 릴리즈 패키지**

```
TOOL-VERSION-PLATFORM.tar.gz
├── TOOL              # 실행 바이너리
├── .bashrc.TOOL      # Shell wrapper (선택, tool에 따라)
├── TOOL-completion.sh # Bash 자동완성 (선택, tool에 따라)
└── .install-manifest # 설치 정보 (향후)
```

### **예시**

```
ccs-v1.0.0-linux-amd64.tar.gz
├── ccs
├── .bashrc.ccs
└── ccs-completion.sh

cw-v1.0.0-darwin-amd64.tar.gz
├── cw
```

---

## 🔨 Release 빌드 프로세스

### **1. Makefile (tools/TOOL/Makefile) 수정**

```makefile
# 변경 전: 단순 바이너리만
release-build-%:
	@go build -o $(RELEASE_DIR)/$(TOOL)-$(VERSION)-$* ...

# 변경 후: tar.gz 번들
release-build-%:
	@mkdir -p $(RELEASE_DIR)/staging
	@go build -o $(RELEASE_DIR)/staging/$(TOOL) ...

	# 보조 파일 복사 (있으면)
	@[ -f .bashrc.$(TOOL) ] && cp .bashrc.$(TOOL) $(RELEASE_DIR)/staging/ || true
	@[ -f $(TOOL)-completion.sh ] && cp $(TOOL)-completion.sh $(RELEASE_DIR)/staging/ || true

	# tar.gz 압축
	@cd $(RELEASE_DIR)/staging && tar -czf ../$(TOOL)-$(VERSION)-$*.tar.gz . && cd ..
	@rm -rf $(RELEASE_DIR)/staging
```

### **2. 실행**

```bash
make release-build VERSION=v1.0.0
```

결과:
```
dist/
├── ccs-v1.0.0-linux-amd64.tar.gz
├── ccs-v1.0.0-linux-arm64.tar.gz
├── ccs-v1.0.0-darwin-amd64.tar.gz
├── ccs-v1.0.0-darwin-arm64.tar.gz
└── ccs-v1.0.0-windows-amd64.zip  # Windows는 zip
```

---

## 📥 설치 스크립트 (install.sh) 수정

### **다운로드 및 압축 해제**

```bash
download_binary() {
    local tool=$1
    local version=$2
    local platform=$3

    # 1. 압축 파일 다운로드
    local package_name="${tool}-${version}-${platform}.tar.gz"
    local download_url="https://github.com/$GITHUB_REPO/releases/download/$version/$package_name"

    curl -fsSL -o "$cache_path" "$download_url"

    # 2. 압축 해제
    local staging_dir="$CACHE_DIR/staging-${tool}-${version}"
    mkdir -p "$staging_dir"
    tar -xzf "$cache_path" -C "$staging_dir"

    echo "$staging_dir"  # 압축 해제된 디렉토리 반환
}
```

### **설치**

```bash
install_tool() {
    local tool=$1
    local version=$2
    local platform=$3

    # 다운로드 + 압축 해제
    local staging_dir=$(download_binary "$tool" "$version" "$platform")

    # 1. 바이너리 설치
    cp "$staging_dir/$tool" ~/.local/bin/$tool
    chmod +x ~/.local/bin/$tool

    # 2. 보조 파일 설치
    [ -f "$staging_dir/.bashrc.$tool" ] && cp "$staging_dir/.bashrc.$tool" ~/.bashrc.$tool
    [ -f "$staging_dir/$tool-completion.sh" ] && cp "$staging_dir/$tool-completion.sh" ~/.bash_completion.d/$tool
}
```

---

## 🖥️ 플랫폼별 지원

### **Linux / macOS**

```
패키지: tar.gz
포함:
  ✅ 실행 바이너리
  ✅ Shell wrapper (.bashrc.TOOL)
  ✅ Bash 자동완성 (.sh)
```

### **Windows**

```
패키지: zip (또는 exe)
포함:
  ✅ 실행 바이너리
  ❌ Shell wrapper (bash 미지원)
  ❌ Bash 자동완성 (bash 미지원)
```

**참고:** ccs는 shell wrapper가 필수이므로 **Windows 미지원**

---

## 📋 도구별 지원 매트릭스

| 도구 | 바이너리 | Shell Wrapper | 자동완성 | 플랫폼 |
|------|---------|---------------|---------|--------|
| **cw** | ✅ | ❌ | ✅ | 모두 |
| **ccs** | ✅ | ✅ | ✅ | Linux/macOS만 |

---

## 🔄 버전 관리 정책

### **메인 프로그램**
- Semantic Versioning (v1.0.0, v1.1.0 등)
- GitHub Releases에서 버전 고정
- 구 버전도 유지 (하위 호환성)

### **보조 파일**
- 메인 프로그램과 함께 번들 (버전 일치)
- 긴급 패치: main 브랜치 먼저 업데이트, 다음 릴리즈에 포함
- GitHub Raw는 "최신 권장" (선택사항)

---

## 📝 Release 체크리스트

```
릴리즈 준비:
  ☐ 모든 도구 테스트 완료 (make test)
  ☐ README 최신화
  ☐ 버전 번호 결정 (v?.?.?)

릴리즈 빌드:
  ☐ make release-build VERSION=v?.?.?
  ☐ dist/ 디렉토리 확인
    - 모든 플랫폼별 tar.gz/zip 생성
    - 파일 크기 합리적 (압축 정상)
  ☐ 압축 파일 무결성 확인
    - tar -tzf 로 내용 확인

GitHub Release:
  ☐ Release Notes 작성
  ☐ dist/ 파일들 업로드
  ☐ v?.?.? 태그 생성 및 푸시
  ☐ release.sh 실행 (자동화)

사후:
  ☐ 설치 스크립트로 테스트 설치
    ./install.sh ccs v?.?.?
  ☐ 최신 버전 동작 확인
  ☐ 공지/문서 업데이트
```

---

## 🚀 설치 명령어 예시

### **GitHub Release에서 설치**

```bash
# 최신 버전
curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh -s ccs

# 특정 버전
./install.sh ccs v1.0.0

# 로컬 개발 빌드 및 설치 (Source)
cd tools/ccs
make install
```

---

## 🔮 향후 계획

### **자동화**
- [ ] GitHub Actions에서 Release 자동 빌드
- [ ] .install-manifest로 tool별 요구사항 명시
- [ ] install.sh에서 manifest 기반 동적 처리

### **호환성**
- [ ] Windows PowerShell wrapper 추가 (선택)
- [ ] Zsh completion 추가 (선택)
- [ ] Docker 이미지 배포 (선택)

---

## 📚 관련 문서

- [README.md](./README.md) - 프로젝트 개요
- [tools/ccs/README.md](./tools/ccs/README.md) - ccs 사용법
- [tools/cw/README.md](./tools/cw/README.md) - cw 사용법
- [CONTRIBUTING.md](./CONTRIBUTING.md) - 개발 기여 가이드
