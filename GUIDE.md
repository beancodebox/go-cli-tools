# go-cli-tools 개발 & 릴리스 가이드

## 📋 목차

1. [개발 (로컬)](#-개발-로컬)
2. [릴리스 준비](#-릴리스-준비)
3. [테스트 릴리스](#-테스트-릴리스-draft)
4. [공식 릴리스](#-공식-릴리스)
5. [사용자 설치](#-사용자-설치)
6. [문제 해결](#-문제-해결)

---

## 🛠️ 개발 (로컬)

### 빠른 시작

```bash
# 1. 현재 OS용 빌드 및 실행
make build          # ./bin/cw 생성
./bin/cw            # 실행

# 또는 컴파일 없이 바로 실행
go run ./tools/cw/cmd/cw
```

### 상세 명령어

```bash
# 빌드
make build          # cw만 빌드
make build-all      # 모든 도구 빌드

# 테스트
make test           # 모든 테스트 실행
cd tools/cw && go test ./cmd/cw  # cw만 테스트

# 설치 (개발용)
make install        # cw를 ~/.local/bin에 설치
make install-all    # 모든 도구 설치

# 정리
make clean          # 빌드 파일 제거
```

### 도움말

```bash
make help           # 모든 명령어 확인
```

---

## 🚀 릴리스 준비

### 1단계: 코드 확정

```bash
# 모든 변경사항 커밋
git add .
git commit -m "Release v1.0.0"

# 태그 생성
git tag v1.0.0
git push origin v1.0.0
```

### 2단계: 모든 플랫폼용 바이너리 빌드

```bash
# 5개 플랫폼용 빌드 (Linux/macOS/Windows, amd64/arm64)
make release-build VERSION=v1.0.0

# 빌드 확인
find . -name "cw-v1.0.0-*" -type f
# 출력:
# ./tools/cw/dist/cw-v1.0.0-linux-amd64
# ./tools/cw/dist/cw-v1.0.0-linux-arm64
# ./tools/cw/dist/cw-v1.0.0-darwin-amd64
# ./tools/cw/dist/cw-v1.0.0-darwin-arm64
# ./tools/cw/dist/cw-v1.0.0-windows-amd64
```

---

## 🧪 테스트 릴리스 (Draft)

테스트용 릴리스는 **비공개**이며 언제든 삭제 가능합니다.

### 전체 프로세스

```bash
# 1. 테스트 버전용 바이너리 빌드
make release-build VERSION=v0.0.1-test

# 2. GitHub에 Draft 릴리즈 발행 (비공개)
make release-publish-draft VERSION=v0.0.1-test
```

### 결과 확인

```bash
# 웹 확인 (Draft이므로 목록에 안 보임)
# https://github.com/beancodebox/go-cli-tools/releases
# → 오른쪽 상단 "Edit" 클릭하면 draft 릴리즈 보임

# 또는 API로 확인
TOKEN=$(cat .github-token)
curl -s -H "Authorization: token $TOKEN" \
  https://api.github.com/repos/beancodebox/go-cli-tools/releases | grep v0.0.1-test
```

### 삭제

```bash
# GitHub 웹에서:
# 1. Releases → Draft 릴리즈 → Edit
# 2. Delete 클릭
```

---

## 📦 공식 릴리스

공식 릴리스는 **공개**되며 사용자들에게 보여집니다.

### 전체 프로세스

```bash
# 1. 모든 플랫폼용 바이너리 빌드
make release-build VERSION=v1.0.0

# 2. GitHub에 공식 릴리즈 발행 (공개)
make release-publish VERSION=v1.0.0

# 3. 확인
# https://github.com/beancodebox/go-cli-tools/releases
```

### 한 번에 하기 (완전 자동화)

```bash
# Step 1: 토큰 설정 (한 번만)
echo "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx" > .github-token
chmod 600 .github-token

# Step 2: 매번 릴리스할 때
git tag v1.0.0 && git push origin v1.0.0
make release-build VERSION=v1.0.0
make release-publish VERSION=v1.0.0
```

---

## 💾 GitHub 토큰 설정

### 필수: 토큰 생성

**옵션 A: Classic Personal Access Token** (권장)

```
https://github.com/settings/tokens/new

설정:
- Token name: go-cli-tools-release
- Scopes: ☑ repo (전체 저장소 접근)
- Expiration: 90일
```

**옵션 B: Fine-grained Personal Access Token** (더 안전)

```
https://github.com/settings/personal-access-tokens/new

설정:
- Token name: go-cli-tools-release
- Repository access: Only select repositories → go-cli-tools
- Permissions → Contents: Read and write
- Expiration: 90일
- Organization 승인 대기
```

### 토큰 저장

```bash
# 토큰을 로컬 파일에 저장
echo "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx" > .github-token

# 권한 설정 (선택사항이지만 권장)
chmod 600 .github-token

# .gitignore에 등록되어 있으므로 자동으로 커밋 안 됨
cat .gitignore | grep github-token  # 확인
```

### 토큰 검증

```bash
# 토큰 읽기 테스트
TOKEN=$(cat .github-token)
curl -s -H "Authorization: token $TOKEN" https://api.github.com/user | grep login

# 출력 예:
# "login": "your-username"
```

---

## 👥 사용자 설치

### 웹에서 직접 설치

```bash
# 방법 1: 대화형 (어떤 도구 설치할지 선택)
sh | curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh

# 방법 2: 명령행 (cw만 설치)
curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh -s cw

# 방법 3: 로컬 파일에서 실행
./install.sh cw           # cw만 설치
./install.sh              # 대화형 선택
./install.sh --help       # 도움말
```

### 설치 후 PATH 설정

```bash
# ~/.local/bin이 PATH에 없으면:
export PATH="$HOME/.local/bin:$PATH"

# 영구 설정 (쉘 설정 파일에 추가)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc  # bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc   # zsh
```

---

## 🔧 디렉토리 구조

```
go-cli-tools/
├── Makefile                 # 루트 Makefile (전체 관리)
├── install.sh             # 사용자 설치 스크립트
├── release.sh             # GitHub 릴리스 업로드
├── .github-token          # GitHub 토큰 (gitignore됨)
├── tools/
│   ├── cw/
│   │   ├── Makefile       # cw 전용 Makefile
│   │   ├── cmd/cw/        # 소스 코드
│   │   ├── go.mod
│   │   └── dist/          # 빌드된 바이너리
│   │       ├── cw-v1.0.0-linux-amd64
│   │       ├── cw-v1.0.0-darwin-amd64
│   │       └── ...
│   └── (추가될 도구들...)
└── go.work                # Go 모노레포 설정
```

---

## 🆘 문제 해결

### "토큰을 찾을 수 없음"

```
Error: GitHub token not found
```

**해결:**

```bash
# 토큰 파일 생성
echo "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx" > .github-token

# 또는 환경변수 설정
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### "Release 생성 실패" (403 Forbidden)

```
Error: Resource not accessible by personal access token
```

**원인:** 토큰에 권한 부족

**해결:**

```bash
# Classic PAT: repo 권한 확인
https://github.com/settings/tokens

# Fine-grained PAT: Contents 권한 확인
https://github.com/settings/personal-access-tokens
```

### "바이너리 찾을 수 없음"

```
Error: Release directory not found
```

**해결:**

```bash
# 먼저 빌드
make release-build VERSION=v1.0.0

# 빌드 확인
ls -la tools/cw/dist/
```

### "권한 거부" (Permission denied)

```bash
# .github-token 권한 설정
chmod 600 .github-token

# install.sh 권한 설정
chmod +x install.sh release.sh
```

---

## 📖 추가 정보

### 플랫폼별 바이너리

|  | Linux | macOS | Windows |
|---|-------|-------|---------|
| **x86_64** | ✅ | ✅ | ✅ |
| **ARM64** | ✅ | ✅ (M1/M2) | ❌ |

### 새 도구 추가

```bash
# 1. 디렉토리 생성
mkdir -p tools/newtool/cmd/newtool

# 2. go.mod 생성
go mod init github.com/beancodebox/go-cli-tools/tools/newtool
cd tools/newtool

# 3. Makefile 생성 (tools/cw/Makefile 참조)

# 4. go.work 업데이트
# 루트의 go.work에 ./tools/newtool 추가

# 5. 루트 Makefile 업데이트
# TOOLS := cw newtool
```

### 도움말

```bash
make help              # 모든 명령어
./install.sh --help    # 설치 가이드
bash release.sh        # 릴리스 가이드
```

---

**마지막 업데이트:** 2026-03-26
