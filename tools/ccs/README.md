# ccs: Claude Code Multi-Account Manager

Go로 구현된 Claude Code 계정 관리 도구입니다. 여러 Anthropic 계정을 하나의 머신에서 관리하고 빠르게 전환할 수 있습니다.

---

## 설치

### 빌드
```bash
cd tools/ccs
make build
```

### 설치
```bash
make install
```

자동 설치되는 것들:
- ✅ `~/.local/bin/ccs` - 메인 바이너리
- ✅ `~/.bashrc.ccs` - Shell wrapper (환경변수 설정)
- ✅ `~/.bash_completion.d/ccs` - Bash 자동완성

### 설정 활성화

설치 후 다음 중 하나 실행:

**옵션 1: 새 터미널 열기** (권장)
```bash
# 새 터미널 탭/윈도우에서 자동으로 로드됨
```

**옵션 2: 현재 터미널에서 적용**
```bash
source ~/.bashrc
```

---

## 사용법

### 첫 계정 초기화
```bash
ccs init work
```
프롬프트 따라:
1. `claude setup-token` 실행 (토큰 받기)
2. 토큰 붙여넣기
3. 계정 저장

### 계정 전환
```bash
ccs use work          # work 계정 활성화
claude                # Claude 실행 (work 토큰으로)
```

### 계정 목록
```bash
ccs list              # 모든 계정 (별표는 현재 활성)
ccs list --plain      # 계정명만 (스크립팅용)
```

### 현재 상태
```bash
ccs status            # 활성 계정 및 토큰 상태
ccs get-current       # 활성 계정명만 (스크립팅용)
```

### 계정 검증
```bash
ccs verify            # 현재 활성 계정 검증
ccs verify work       # 특정 계정 검증
```

### 계정 복구
```bash
ccs resume            # 마지막 활성 계정으로 복구
```

### 현재 설정 저장
```bash
ccs save-current myaccount    # 현재 ~/.claude를 계정으로 저장
```

### 계정 삭제
```bash
ccs delete work       # work 계정 삭제 (계정명 타이핑으로 확인)
```

### 도움말
```bash
ccs help              # 전체 도움말
```

---

## 파일 구조

```
~/.claude-accounts/
├── work/
│   ├── .token                  # API 토큰 (0600 권한)
│   ├── settings.json
│   ├── conversations/
│   └── ...
├── personal/
│   └── ...
├── current-backup/             # 전환 시 백업
└── .last-active                # 마지막 활성 계정명
```

---

## 기능

| 명령어 | 설명 |
|--------|------|
| `init <name>` | 새 계정 초기화 (토큰 입력) |
| `use <name>` | 계정 활성화 + 설정 복사 |
| `list [--plain]` | 계정 목록 |
| `status` | 현재 상태 |
| `verify [name]` | 계정 검증 |
| `get-current` | 활성 계정명 (스크립팅) |
| `get-token <name>` | 토큰 조회 (스크립팅) |
| `resume` | 마지막 계정 복구 |
| `save-current <name>` | 현재 설정 저장 |
| `delete <name>` | 계정 삭제 |

---

## 기술 사항

- **언어**: Go 1.26.1
- **의존성**: 표준 라이브러리만 사용
- **토큰 보안**: 0600 권한으로 저장
- **Cross-platform**: Windows/macOS/Linux 네이티브 바이너리
- **호환성**: 기존 ~/.claude-accounts/ 데이터 100% 호환

---

## 개발

### 디렉토리 구조
```
tools/ccs/
├── cmd/ccs/
│   ├── main.go       # 커맨드 파싱
│   ├── config.go     # 경로 관리
│   ├── token.go      # 토큰 관리
│   └── accounts.go   # 계정 로직
├── go.mod
└── Makefile
```

### 빌드
```bash
make build       # 바이너리 빌드 (bin/ccs)
make test        # 테스트 실행
make install     # ~/.local/bin에 설치
make clean       # 빌드 파일 삭제
```

---

## Shell Wrapper 상세

ccs는 자식 프로세스이기 때문에 부모 shell의 환경변수를 직접 수정할 수 없습니다.
따라서 shell wrapper가 필수입니다.

```bash
# wrapper가 하는 일:
# 1. ccs use <name> 호출 시
#    - ccs get-token으로 토큰을 얻음
#    - 부모 shell에서 CLAUDE_CODE_OAUTH_TOKEN 설정
# 2. ccs resume 호출 시
#    - 마지막 계정의 토큰을 부모 shell에 설정
# 3. 나머지 커맨드는 그냥 ccs binary에 전달
```

---

## 사용 예시

### 다중 계정 관리
```bash
ccs init work
ccs init personal

ccs use work
claude                    # work 토큰으로 실행

ccs use personal
claude                    # personal 토큰으로 실행
```

### 스크립트에서 사용
```bash
account=$(ccs get-current)
token=$(ccs get-token "$account")

# 스크립트에서 토큰 사용...
```

### 계정 상태 확인
```bash
ccs list
  22i                  (token: yes)
  box009               (token: yes)
* work                 (token: yes)     # * = 현재 활성

ccs status
Active account: work
Path: /home/user/.claude-accounts/work
Token: present
```
