package main

import (
	"flag"
	"fmt"
	"os"
)

// ============================================================================
// cw-go: 대화형 폴더 탐색 네비게이터
// ============================================================================
//
// 이 파일의 책임: "프로그램의 진입점"과 "전체 흐름 조율"
//
// 파일 분리 구조:
// - main.go: 진입점 (여기) - 프로그램의 흐름 제어
// - navigator.go: 폴더 탐색과 UI 로직
// - config.go: 설정 파일 읽기/쓰기
// - executor.go: Claude 실행
//
// 이 구조의 장점:
// 1. 각 파일이 하나의 책임만 가짐 (Single Responsibility Principle)
// 2. 코드 변경 시 영향 범위가 명확함
// 3. 테스트할 때 각 부분을 독립적으로 테스트 가능
// 4. 새 기능 추가 시 (예: Phase 3 미리보기) 해당 파일만 수정
//
// Go 프로젝트 구조의 실전 팁:
// - 초보자: 파일당 200줄 정도가 읽기 좋은 크기
// - 파일명은 그 파일의 주요 역할을 나타내도록 (navigator.go는 탐색 담당)
// - 같은 package main이므로 import 불필요 (같은 폴더라서 자동으로 컴파일됨)
// ============================================================================

func main() {
	// ============================================================================
	// 1단계: CLI 플래그 파싱 및 초기화
	// ============================================================================
	//
	// Go의 flag 패키지:
	// - 표준 라이브러리로 --flag 형태 파싱
	// - 명령어: cw -r /path (짧은 형태)
	// - 명령어: cw --root /path (긴 형태)
	// - 명령어: cw --resume (bool 플래그)
	//
	// flag.Parse() 전에 flag.Bool() 등으로 플래그 정의
	// flag.Parse() 호출 후 flag.Args()로 나머지 인자들 접근
	// ============================================================================

	resume := flag.Bool("resume", false, "resume from last session")
	rootFlag := flag.String("root", "", "set root directory (saves to ~/.cw)")
	rFlag := flag.String("r", "", "short for --root")
	selectAccount := flag.Bool("a", false, "select Claude account interactively (short form)")
	accountName := flag.String("account", "", "specify Claude account directly (long form)")
	flag.Parse()

	// 계정 선택 모드 결정
	// -a: 대화형 선택, --account: 직접 지정
	account := ""
	if *selectAccount {
		account = "*" // 특수 값: 대화형 선택
	} else if *accountName != "" {
		account = *accountName
	}

	// -r 또는 --root 플래그로 설정 저장
	// "-r"과 "--root" 중 하나가 지정되면 configRoot에 저장
	configRoot := *rootFlag
	if configRoot == "" && *rFlag != "" {
		configRoot = *rFlag
	}

	// 만약 사용자가 "cw --root /some/path"를 입력했다면
	// 경로를 ~/.cw에 저장하고 프로그램 종료
	// (이 경우 폴더 탐색을 하지 않음)
	if configRoot != "" {
		if err := saveRootPath(configRoot); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Root path saved to ~/.cw: %s\n", configRoot)
		return // 여기서 프로그램 종료
	}

	// Config 구조체 초기화
	cfg := Config{
		Resume:  *resume,
		Account: account, // account는 이미 string (포인터 디레퍼런싱 불필요)
	}

	// ============================================================================
	// 2단계: 루트 경로 결정
	// ============================================================================
	//
	// getRootPath() 함수는 우선순위 방식으로 경로 결정:
	// 1. ~/.cw 파일에 저장된 경로
	// 2. $CW_ROOT 환경변수
	// 3. 홈 디렉토리 (기본값)
	//
	// 이렇게 하면:
	// - 파워 유저: 환경변수나 파일로 커스터마이징
	// - 일반 사용자: 아무것도 안 해도 홈에서 시작
	// ============================================================================

	rootPath := getRootPath()
	navigator := Navigator{rootPath, []FolderItem{}}

	// ============================================================================
	// 3단계: 폴더 탐색 루프 (사용자가 선택할 때까지 반복)
	// ============================================================================
	//
	// 루프 설명:
	// - selected := false로 시작: 사용자가 아직 선택 안 함
	// - !selected 동안 반복
	// - RUN_CLAUDE 항목 선택 → selected = true → 루프 탈출
	//
	// 각 반복에서:
	// 1. buildItems(): 현재 폴더의 탐색 가능한 항목들 생성
	// 2. runSelectUI(): 사용자가 선택하도록 huh UI 표시
	// 3. switch: 선택 결과에 따라 처리
	//    - RUN_CLAUDE: 루프 탈출 (Claude 실행)
	//    - FOLDER 또는 PARENT: 경로 이동 후 다시 반복
	// ============================================================================

	for selected := false; !selected; {
		// buildItems()는 폴더를 읽고 탐색 가능한 항목 목록 반환
		items, err := buildItems(navigator.CurrentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory %q: %v\n", navigator.CurrentDir, err)
			os.Exit(1)
		}
		navigator.Items = items

		// runSelectUI()는 huh 라이브러리로 선택 UI 표시
		// 사용자가 화살표키로 항목 선택, Enter로 확정
		currentItem, err := runSelectUI(navigator.CurrentDir, items)
		if err != nil {
			// ESC 눌러서 취소했거나 UI 오류 발생
			fmt.Fprintf(os.Stderr, "Selection cancelled or failed: %v\n", err)
			os.Exit(1)
		}

		// 선택 결과에 따라 처리
		switch currentItem.Type {
		case RUN_CLAUDE:
			// 사용자가 "[Run Claude Here]" 선택 → 루프 탈출
			selected = true

		case PARENT:
			// ".." 선택 → 상위 폴더로 이동 → 루프 계속
			navigator.CurrentDir = currentItem.Path

		case FOLDER:
			// 폴더 선택 → 그 폴더로 이동 → 루프 계속
			navigator.CurrentDir = currentItem.Path
		}
	}

	fmt.Printf("You choose %q\n", navigator.CurrentDir)

	// ============================================================================
	// 4단계: 계정 선택 (Phase 3)
	// ============================================================================
	//
	// -a / --account 플래그가 있으면 Claude 계정 전환
	// ccs 명령어를 통해 계정 관리
	// ============================================================================

	// selectAccountIfNeeded() - 계정 전환 (필요시)
	// cfg.Account = "-a" 플래그 값
	// - "": 계정 선택하지 않음
	// - "계정명": 지정한 계정으로 전환
	if err := selectAccountIfNeeded(cfg); err != nil {
		// 계정 전환 실패해도 계속 진행 (경고만 출력)
		fmt.Fprintf(os.Stderr, "Account selection failed: %v\n", err)
	}

	// ============================================================================
	// 5단계: Claude 실행
	// ============================================================================
	//
	// buildClaudeCmd(): 실행할 명령어 정보 구성
	// - "claude" 프로그램 실행
	// - --resume 플래그 추가 (필요시)
	// - 사용자가 전달한 추가 인자들 포함
	//
	// runClaudeCmd(): 실제 실행
	// - OS와 쉘에 따라 적절한 방식으로 실행
	// - Windows/Unix 차이 처리
	// - 현재 디렉토리에서 실행
	// ============================================================================

	cmd := buildClaudeCmd(navigator.CurrentDir, cfg, flag.Args())

	// Claude 실행
	// 실행 오류 발생 시 에러 메시지 출력 후 종료
	if err := runClaudeCmd(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
