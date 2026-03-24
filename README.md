# go-cli-tools

Go로 만든 CLI 도구 모음입니다.

## 📁 구조

```
go-cli-tools/
├── tools/
│   ├── cw/              # 작업공간 네비게이터
│   │   ├── cmd/cw/
│   │   │   ├── main.go
│   │   │   ├── navigator.go
│   │   │   ├── config.go
│   │   │   └── executor.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   └── (더 추가 예정)
├── scripts/             # 모노레포 전체 빌드/배포 스크립트
├── go.work              # Go 1.18+ 멀티모듈 지원
└── README.md
```

## 🚀 도구들

### cw - Workspace Navigator
- 대화형 폴더 탐색
- Claude 통합 실행
- 크로스플랫폼 지원

자세한 사항은 [`tools/cw/README.md`](tools/cw/README.md)를 참고하세요.

## 설치

### cw 설치
```bash
cd tools/cw
go build -o cw ./cmd/cw
mv cw ~/.local/bin/
```

### 모든 도구 빌드
```bash
# go.work를 사용하면 모든 모듈을 한 번에 관리 가능
go build ./tools/...
```

## 개발

### 로컬 개발
```bash
# go.work가 있으면 모든 모듈이 동시에 관리됨
go mod tidy
go test ./...
```

### 새 도구 추가
1. `tools/new-tool/` 디렉토리 생성
2. `tools/new-tool/go.mod` 작성
3. `go.work`에 추가
4. `tools/new-tool/cmd/new-tool/main.go` 작성

## 라이선스

MIT
