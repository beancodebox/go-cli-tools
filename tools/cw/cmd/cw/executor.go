package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"charm.land/huh/v2"
)

// ============================================================================
// 책임: Claude 실행 및 OS별 쉘 관리
// ============================================================================
//
// Go에서 외부 프로세스 실행:
// - os/exec: Go의 표준 라이브러리로 외부 명령어 실행
// - subprocess/process management 패턴
//
// 이 파일이 복잡한 이유:
// Windows, macOS, Linux의 쉘이 다름
// - Windows: cmd.exe, PowerShell, Git Bash
// - Unix계: /bin/sh, /bin/bash
//
// 따라서 런타임에 OS와 환경변수를 감지해서
// 적절한 쉘 명령어를 구성해야 함.
//
// 만약 이 로직이 main()에 섞여있다면:
// 1. main()이 너무 길어짐
// 2. 쉘 호환성 테스트가 어려움
// 3. 쉘 추가(예: nushell)할 때 main()을 수정해야 함
//
// 따라서 이 파일로 분리 → 단일 책임 원칙 준수
// ============================================================================

// ClaudeCmd는 Claude 실행에 필요한 명령어와 경로를 담는 구조체입니다.
//
// 설계 패턴: Command Pattern
// - "어떤 명령을 할 것인가"를 데이터 구조로 표현
// - 실제 실행은 나중에 (runClaudeCmd에서)
// - 이렇게 하면 테스트할 때 실제 실행 없이도 명령어 구성 검증 가능
type ClaudeCmd struct {
	Dir  string   // 작업 디렉토리 (이 폴더에서 Claude를 시작)
	Args []string // 실행할 명령어와 인자들 (예: ["claude", "--resume", "arg1"])
}

// buildClaudeCmd는 Claude를 실행할 명령어를 구성합니다.
//
// 입력 매개변수:
// - dir: Claude를 실행할 작업 디렉토리
// - cfg: Config 구조체 (--resume 플래그 등)
// - extraArgs: 사용자가 전달한 추가 인자들
//
// 반환값:
// - ClaudeCmd: 실행 준비된 명령어 정보 (아직 실행 안 됨)
//
// 설계 사상:
// - "명령어 생성"과 "명령어 실행"을 분리
// - 생성한 명령어를 검증하거나 로깅할 기회 제공
// - Phase 4에서 --account 플래그 추가할 때도
//   이 함수만 수정하면 됨
func buildClaudeCmd(dir string, cfg Config, extraArgs []string) ClaudeCmd {
	args := []string{"claude"}

	// --resume 플래그가 설정되면 추가
	// 이 플래그는 "cw --resume"으로 입력받은 것
	if cfg.Resume {
		args = append(args, "--resume")
	}

	// 사용자가 전달한 추가 인자들을 모두 붙임
	// 예: cw -- --verbose arg1 arg2 → 마지막 3개 포함
	args = append(args, extraArgs...)

	return ClaudeCmd{
		Dir:  dir,
		Args: args,
	}
	// TODO (Phase 4): Account 필드 처리
	// if cfg.Account != "" {
	//     args = append(args, "-a", cfg.Account)
	// }
}

// runClaudeCmd는 ClaudeCmd를 실행합니다.
//
// 이 함수의 복잡함:
// Go의 os/exec은 자동으로 PATH 환경변수를 검색하지만,
// 명령어를 실행하는 방식은 os와 쉘에 따라 다름:
//
// Windows cmd.exe:
//   cmd /c "cd /d C:\path && claude --resume"
//
// Windows PowerShell:
//   powershell -Command "cd 'C:\path'; & claude --resume"
//
// Unix/Linux:
//   bash -c "cd '/path' && claude --resume"
//
// 왜 이렇게 복잡한가?
// - exec.Command()는 PATH에 있는 명령을 직접 실행
// - 하지만 "cd + 명령어"는 쉘이 지원하는 연산자 (&&)를 필요로 함
// - 따라서 "쉘 -c '명령어'"로 감싸야 함
// - cd는 대부분의 쉘에서 빌트인 명령어라 직접 실행 불가
//
// runtime.GOOS를 사용한 OS 감지:
// "darwin" (macOS), "linux", "windows" 등
// 컴파일 시점에는 알 수 없고, 런타임에 결정됨
func runClaudeCmd(cmd ClaudeCmd) error {
	var shellCmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows 쉘 감지
		// 환경변수 SHELL을 확인: SHELL이 없으면 cmd.exe, 있으면 PowerShell/Bash
		//
		// 배경:
		// - cmd.exe: 기본 Windows 쉘
		// - PowerShell: 최신 Windows (보통 명시적으로 설정)
		// - Git Bash: Git for Windows 설치 시 SHELL=/bin/bash 설정
		shell := os.Getenv("SHELL")

		if shell == "" {
			// cmd.exe (가장 흔한 경우)
			// cd /d: Windows에서 드라이브 변경까지 지원하는 옵션
			// /c: 명령어 실행 후 종료
			shellCmd = exec.Command("cmd", "/c",
				fmt.Sprintf("cd /d %q && %s", cmd.Dir, strings.Join(cmd.Args, " ")))

		} else if strings.Contains(shell, "powershell") || strings.Contains(shell, "pwsh") {
			// PowerShell 또는 PowerShell Core
			// cd '경로': 단일 따옴표로 경로 보호 (공백 포함 경로 안전)
			// &: PowerShell의 call operator (명령어 실행)
			claudeArgs := strings.Join(cmd.Args[1:], " ") // "claude" 빼고 나머지만
			shellCmd = exec.Command("powershell", "-Command",
				fmt.Sprintf("cd '%s'; & %s %s", cmd.Dir, cmd.Args[0], claudeArgs))

		} else if strings.Contains(shell, "bash") {
			// Git Bash (bash가 Windows에서 돌아감)
			// Unix와 동일한 방식
			shellCmd = exec.Command("bash", "-c",
				fmt.Sprintf("cd %q && %s", cmd.Dir, strings.Join(cmd.Args, " ")))
		}

	} else {
		// Unix/Linux/macOS
		// 기본 쉘 환경변수 SHELL 확인, 없으면 /bin/sh
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		// 모든 Unix 쉘은 "-c '명령어'" 방식 지원
		execCmd := fmt.Sprintf("cd %q && %s", cmd.Dir, strings.Join(cmd.Args, " "))
		shellCmd = exec.Command(shell, "-c", execCmd)
	}

	// 쉘 명령어를 구성했으면, stdio 연결
	// 이렇게 하면 Claude의 출력이 사용자 터미널에 보임
	shellCmd.Stdin = os.Stdin
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr

	// 실제 실행
	if err := shellCmd.Run(); err != nil {
		// Run()은 프로세스가 0이 아닌 exit code로 종료되면 에러 반환
		// (프로세스 실행 자체는 성공했고, 단지 exit code가 non-zero일 뿐)
		return fmt.Errorf("failed to run claude: %w", err)
	}

	return nil
}

// ============================================================================
// Phase 3: 계정 선택 (Claude Code 계정 관리)
// ============================================================================
//
// CCS (Claude Code's Account Management Tool) 호출
// 기존 쉘 스크립트 cw.sh에서 사용하던 기능을 Go로 포팅
//
// CCS 명령어:
// - ccs list --plain: 저장된 계정 목록 (한 줄에 하나)
// - ccs get-current: 현재 선택된 계정명
// - ccs use <account>: 지정한 계정으로 전환
//
// 흐름:
// 1. selectAccountIfNeeded() 호출 (main에서)
// 2. ccs 명령어 존재 확인
// 3. 계정 목록 조회
// 4. huh/v2로 UI 표시 (사용자 선택)
// 5. ccs use로 계정 전환
// 6. 그 다음 Claude 실행
// ============================================================================

// selectAccountIfNeeded는 -a 플래그가 있으면 계정을 선택하거나 전환합니다.
//
// 입력:
// - cfg.Account: 계정 선택 모드
//   "" (빈 문자열): 계정 선택하지 않음
//   "*": 대화형 선택 UI 표시 (-a 플래그만 입력)
//   "계정명": 지정한 계정으로 직접 전환 (--account 계정명)
//
// 반환:
// - error: 계정 전환 실패 시 (하지만 에러여도 계속 진행)
func selectAccountIfNeeded(cfg Config) error {
	// 계정 선택 안 함
	if cfg.Account == "" {
		return nil
	}

	// ccs 명령어 사용 가능 확인
	if !isCcsAvailable() {
		fmt.Fprintf(os.Stderr, "Warning: ccs not found. Skipping account switch.\n")
		return nil
	}

	// "*"는 대화형 선택 의미 (-a 플래그만 입력)
	if cfg.Account == "*" {
		return interactiveSelectAccount()
	}

	// 그 외는 계정명 지정 (--account 계정명)
	return switchAccount(cfg.Account)
}

// isCcsAvailable은 ccs 명령어가 PATH에 있는지 확인합니다.
func isCcsAvailable() bool {
	cmd := exec.Command("ccs", "--version")
	// 명령어 존재 확인만, 출력은 버림
	return cmd.Run() == nil
}

// interactiveSelectAccount는 huh/v2를 사용해 계정을 대화형으로 선택합니다.
//
// 흐름:
// 1. ccs list --plain으로 계정 목록 조회
// 2. huh/v2 선택 UI 표시
// 3. 사용자 선택 후 ccs use로 전환
func interactiveSelectAccount() error {
	// 계정 목록 조회
	accounts, err := getAccountList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting account list: %v\n", err)
		return err
	}

	if len(accounts) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No saved accounts. Run: ccs init <name>\n")
		return nil
	}

	// 현재 계정 조회
	current, _ := getCurrentAccount()

	// huh/v2로 선택 UI 표시
	// huh.NewSelect는 string 타입으로 선택하도록 설정
	var selectedAccount string
	headerText := "Select account"
	if current != "" {
		headerText = fmt.Sprintf("Current: %s | Select account (Esc to skip)", current)
	}

	err = huh.NewSelect[string]().
		Title(headerText).
		Options(stringToOptions(accounts)...).
		Value(&selectedAccount).
		Run()

	if err != nil {
		// ESC로 취소한 경우 에러지만 계속 진행
		return nil
	}

	// 선택한 계정으로 전환
	if selectedAccount != "" {
		return switchAccount(selectedAccount)
	}

	return nil
}

// getAccountList는 ccs list --plain으로 계정 목록을 조회합니다.
// 반환: []string (계정명들, 한 줄에 하나)
func getAccountList() ([]string, error) {
	cmd := exec.Command("ccs", "list", "--plain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 출력을 줄 단위로 분리
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// 빈 줄 제거
	var accounts []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			accounts = append(accounts, strings.TrimSpace(line))
		}
	}

	return accounts, nil
}

// getCurrentAccount는 ccs get-current로 현재 계정명을 조회합니다.
func getCurrentAccount() (string, error) {
	cmd := exec.Command("ccs", "get-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// switchAccount는 ccs use로 지정한 계정으로 전환합니다.
func switchAccount(accountName string) error {
	cmd := exec.Command("ccs", "use", accountName)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error switching to account %q: %v\n", accountName, err)
		return err
	}

	fmt.Printf("Switched to account: %s\n", accountName)
	return nil
}

// stringToOptions는 []string을 huh.Option[string] 슬라이스로 변환합니다.
// huh/v2의 NewSelect[T]().Options()에 전달하기 위한 헬퍼 함수
func stringToOptions(items []string) []huh.Option[string] {
	options := make([]huh.Option[string], len(items))
	for i, item := range items {
		// label과 value가 같은 경우
		options[i] = huh.NewOption(item, item)
	}
	return options
}
