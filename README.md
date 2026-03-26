# go-cli-tools

Go로 만든 CLI 도구 모음입니다.

## 🚀 빠른 시작

```bash
# 설치
sh | curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh

# 또는 cw만 설치
curl -fsSL https://raw.githubusercontent.com/beancodebox/go-cli-tools/main/install.sh | sh -s cw
```

## 📦 도구들

### cw - Workspace Navigator
- 대화형 폴더 탐색
- Claude IDE 통합
- 크로스플랫폼 지원

더 자세한 정보: [`tools/cw/README.md`](tools/cw/README.md)

---

## 📚 문서

| 문서 | 설명 |
|------|------|
| **[GUIDE.md](GUIDE.md)** | 개발, 빌드, 릴리스 전체 가이드 |
| **[tools/cw/README.md](tools/cw/README.md)** | cw 도구 상세 정보 |

---

## 💻 개발자용

### 로컬 빌드 및 테스트

```bash
# 빌드
make build              # 현재 OS용
make build-all          # 모든 도구
make release-build VERSION=v1.0.0  # 모든 플랫폼

# 테스트
make test

# 설치
make install            # ~/.local/bin에 설치
```

### 릴리스

```bash
# 1. 토큰 설정 (한 번만)
echo "ghp_token..." > .github-token

# 2. 테스트 릴리스 (draft, 비공개)
make release-build VERSION=v0.0.1-test
make release-publish-draft VERSION=v0.0.1-test

# 3. 공식 릴리스
make release-build VERSION=v1.0.0
make release-publish VERSION=v1.0.0
```

더 자세한 내용은 [GUIDE.md](GUIDE.md)를 참고하세요.

---

## 📁 구조

```
go-cli-tools/ (Go 모노레포)
├── tools/
│   ├── cw/                      # Workspace Navigator
│   │   ├── cmd/cw/              # 소스 코드
│   │   ├── Makefile
│   │   ├── go.mod
│   │   └── dist/                # 빌드된 바이너리
│   └── (추가될 도구들...)
├── Makefile                     # 루트 Makefile
├── install.sh                   # 사용자 설치 스크립트
├── release.sh                   # GitHub 릴리스 업로드
├── GUIDE.md                     # 📖 개발 & 릴리스 가이드
├── go.work                      # Go 1.18+ 모노레포
└── README.md                    # (이 파일)
```

---

## 🔗 링크

- GitHub: https://github.com/beancodebox/go-cli-tools
- Releases: https://github.com/beancodebox/go-cli-tools/releases

---

## 📄 라이선스

MIT
