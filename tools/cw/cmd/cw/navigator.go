package main

import (
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
)

// ============================================================================
// 책임: 폴더 탐색 로직 및 사용자 인터페이스
// ============================================================================
//
// Go의 "관심사 분리(Separation of Concerns)" 원칙:
// - main.go: 프로그램의 흐름 제어 (어떤 순서로 무엇을 할 것인가)
// - navigator.go: 폴더 탐색과 사용자 입력 처리 (사용자와의 상호작용)
// - config.go: 설정값 관리 (데이터 저장/로드)
// - executor.go: Claude 실행 (외부 프로세스 관리)
//
// 이렇게 파일을 나누면:
// 1. 각 파일이 하나의 역할에만 집중 (단일 책임 원칙)
// 2. 코드 수정 시 영향 범위가 명확해짐
// 3. 테스트할 때 각 부분을 독립적으로 테스트 가능
// ============================================================================

// ItemType은 폴더 항목의 종류를 구분하는 열거형입니다.
// Go에서 const + iota를 사용한 "열거형" 패턴입니다.
//
// 왜 이렇게 할까?
// - 문자열("folder", "run", "parent")보다 타입 안전성 제공
// - 컴파일러가 잘못된 값을 미리 감지
// - switch문에서 모든 경우를 처리했는지 체크 가능
type ItemType int

const (
	FOLDER    ItemType = iota // iota = 0, 1, 2, ... (자동 증가)
	RUN_CLAUDE               // = 1
	PARENT                   // = 2
)

// UI에 표시할 라벨들. 상수로 정의하면:
// - 여러 곳에서 같은 값 사용 시 수정이 한 곳만 하면 됨
// - 오타 방지 (문자열 하드코딩 피하기)
const (
	runClaudeHereLabel = "[Run Claude Here]"
	parentDirLabel     = ".."
)

// FolderItem은 폴더 탐색 시 표시할 하나의 항목을 나타냅니다.
//
// Go 구조체(struct) 설계 원칙:
// - 필드는 대문자로 시작: exported (다른 파일에서도 사용 가능)
// - 각 필드에 명확한 목적을 가짐
//
// 왜 이런 구조체가 필요할까?
// huh 라이브러리가 []huh.Option[T]를 사용하는데,
// 사용자 선택지(Name)와 실제 데이터(Path, Type)를 함께 관리하려면
// 이렇게 묶어야 함.
type FolderItem struct {
	Name string   // UI에 표시될 이름 (폴더명 또는 "[Run Claude Here]")
	Path string   // 실제 파일시스템 경로
	Type ItemType // FOLDER, RUN_CLAUDE, PARENT 중 하나
}

// Navigator는 현재 탐색 상태를 관리하는 구조체입니다.
//
// 상태 관리 패턴:
// - 프로그램이 "지금 어디에 있는가"를 추적
// - 사용자가 폴더를 이동할 때마다 CurrentDir 업데이트
// - Items는 현재 폴더의 항목들을 캐시 (UI 렌더링에 사용)
type Navigator struct {
	CurrentDir string       // 현재 디렉토리 절대 경로
	Items      []FolderItem // 현재 폴더의 탐색 가능한 항목들
}

// buildItems는 주어진 디렉토리를 읽어 FolderItem 목록을 생성합니다.
//
// 함수 설계 원칙:
// - 입력(dir string) 명확: 어디를 읽을 것인가
// - 출력([]FolderItem, error) 명확: 뭘 반환하고, 에러 가능성 명시
// - 단일 책임: 디렉토리 읽기와 필터링만 담당
//
// 왜 main()에서 분리했을까?
// - 탐색 로직이 재사용될 가능성이 있음 (미리보기 기능 등)
// - 테스트하기 쉬움 (실제 폴더 대신 가짜 경로로 테스트)
// - main()은 "무엇을" 할지만 명시, 구현은 여기서
func buildItems(dir string) ([]FolderItem, error) {
	optionItems := []FolderItem{}

	// "[Run Claude Here]" 선택지를 항상 맨 앞에 추가
	// 사용자가 이 폴더에서 Claude를 실행하려면 이 항목 선택
	optionItems = append(optionItems, FolderItem{
		Name: runClaudeHereLabel,
		Path: dir,
		Type: RUN_CLAUDE,
	})

	// ".." (상위 폴더) 선택지 추가 (루트가 아닐 때만)
	// filepath.Dir은 부모 디렉토리를 반환함
	// 같은 경로가 반환되면 루트 디렉토리라는 뜻
	parentDir := filepath.Dir(dir)
	if parentDir != dir {
		optionItems = append(optionItems, FolderItem{
			Name: parentDirLabel,
			Path: parentDir,
			Type: PARENT,
		})
	}

	// 디렉토리 읽기
	// os.ReadDir은 Go 1.16+의 권장 방식 (이전의 ioutil.ReadDir는 deprecated)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// 디렉토리만 필터링하고 숨김 폴더는 제외
	// 이 필터링으로 클린한 목록만 사용자에게 보여줌
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			// "."로 시작하는 폴더는 숨김 (예: .git, .env)
			optionItems = append(optionItems, FolderItem{
				Name: entry.Name(),
				Path: filepath.Join(dir, entry.Name()),
				Type: FOLDER,
			})
		}
	}

	return optionItems, nil
}

// runSelectUI는 huh 라이브러리를 사용해 사용자에게 선택지를 제시합니다.
//
// UI 라이브러리 추상화:
// - huh/v2는 구체적인 UI 구현
// - 이 함수로 감싸면, 나중에 다른 라이브러리로 바꾸기 쉬워짐
// - 예: huh → bubbletea로 전환해도 runSelectUI만 수정하면 됨
//
// 입력 매개변수:
// - title: 선택 화면 상단에 표시할 현재 경로
// - items: 선택지 목록
//
// 반환값:
// - FolderItem: 사용자가 선택한 항목
// - error: ESC 눌러서 취소했거나 UI 오류 발생 시
func runSelectUI(title string, items []FolderItem) (FolderItem, error) {
	// huh.Option[T]는 "표시 텍스트"와 "실제 값"을 쌍으로 관리
	// Go의 제네릭(Generic) 패턴으로 타입 안전성 보장
	//
	// 제네릭이란?
	// - huh.Option[FolderItem]은 "FolderItem만 다룬다"는 뜻
	// - 컴파일러가 다른 타입 실수를 방지
	selects := make([]huh.Option[FolderItem], len(items))
	for i, item := range items {
		// huh.NewOption(label, value)
		// - label: UI에 보여줄 텍스트 (item.Name)
		// - value: 선택 시 반환될 실제 값 (item 자체)
		selects[i] = huh.NewOption(item.Name, item)
	}

	// 선택지 UI 구성
	var currentItem FolderItem
	err := huh.NewSelect[FolderItem]().
		Title(title). // 현재 디렉토리 경로 표시
		Options(selects...).
		Value(&currentItem). // 선택 결과를 이 변수에 저장
		Run()               // 실제 UI 실행

	return currentItem, err
}
