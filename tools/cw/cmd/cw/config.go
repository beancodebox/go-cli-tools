package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// 책임: 설정 파일 관리 및 경로 처리
// ============================================================================
//
// Go에서 설정을 다루는 일반적인 패턴:
// 1. 설정값들을 구조체(struct)로 정의 → 타입 안전성
// 2. 저장소에서 읽고 쓰는 함수 분리 → 유연성
// 3. 우선순위 로직 명확히 → 예측 가능한 동작
//
// 예를 들어 나중에 JSON 파일이나 환경변수 외에
// 홈 폴더의 .config/cw-go/config.json 같은 파일도 지원하려면,
// 이 파일의 getRootPath() 함수만 수정하면 됨.
// ============================================================================

// Config는 cw-go의 실행 옵션을 관리하는 구조체입니다.
//
// 구조체 설계 원칙:
// - 대문자로 시작하는 필드: exported (다른 파일에서 접근 가능)
// - 각 필드의 의미를 명확하게
// - 나중에 추가 옵션이 필요하면 필드만 추가 (하위 호환성 유지)
//
// 향후 확장 가능성:
// - Verbose bool    // 상세 로그 출력
type Config struct {
	Resume  bool   // --resume 플래그: 마지막 세션에서 계속
	Account string // -a / --account 플래그: Claude 계정 선택 (Phase 3)
}

// getRootPath는 사용자가 지정한 루트 디렉토리를 반환합니다.
//
// 우선순위 패턴:
// Go나 다른 도구들도 이렇게 우선순위를 두고 설정값을 읽음:
// 1. 사용자의 명시적 저장소 (~/.cw 파일)
// 2. 환경변수 ($CW_ROOT)
// 3. 기본값 (홈 디렉토리)
//
// 장점:
// - 고급 사용자: 파일이나 환경변수로 커스터마이징
// - 기본 사용자: 아무것도 안 해도 홈 디렉토리에서 시작
// - 테스트: 임시로 환경변수 설정해서 테스트 가능
//
// 참고: 이 함수는 오류를 반환하지 않음
// - ~/.cw 읽기 실패? → 다음 우선순위로
// - $CW_ROOT 없음? → 기본값으로
// 따라서 항상 유효한 경로를 반환하는 것을 보장
func getRootPath() string {
	// 1. ~/.cw 파일 존재 여부 확인 및 읽기
	// 사용자가 "cw --root /my/path"로 설정했다면
	// ~/.cw 파일에 /my/path가 저장되어 있음
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".cw")

	if data, err := os.ReadFile(configPath); err == nil {
		// 파일을 읽을 수 있다면 (err == nil)
		path := strings.TrimSpace(string(data))
		if path != "" {
			// 공백만 있지 않다면 이 경로 사용
			return path
		}
	}
	// 읽기 실패 또는 파일이 비어있으면 다음 단계로

	// 2. 환경변수 $CW_ROOT 확인
	// 사용자가 쉘 설정(.bashrc, .zshrc 등)에 export CW_ROOT=/my/path 설정했다면
	// 이것을 우선적으로 사용할 수 있음
	if env := os.Getenv("CW_ROOT"); env != "" {
		return env
	}

	// 3. 기본값: 홈 디렉토리
	// 설정이 없으면 사용자 홈에서 시작
	return home
}

// expandPath는 ~ 표기법의 경로를 절대 경로로 변환합니다.
//
// 입력 예:
// - "~/projects" → "/home/user/projects"
// - "/absolute/path" → "/absolute/path" (변화 없음)
//
// 왜 이 함수가 필요한가?
// - 사용자는 편의상 ~를 사용하지만 Go 함수들은 절대 경로 필요
// - 이 함수로 사용자 입력 → 절대 경로로 표준화
// - 여러 곳에서 필요하면 함수로 분리해서 재사용
func expandPath(path string) string {
	// strings.HasPrefix 사용: path가 "~/"로 시작하는지 확인
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		// path[2:] = "~/" 뒤의 나머지 문자열
		// 예: "~/projects/myapp" → "projects/myapp"
		return filepath.Join(home, path[2:])
	}
	// ~로 시작하지 않으면 그대로 반환
	return path
}

// saveRootPath는 사용자가 지정한 루트 경로를 ~/.cw 파일에 저장합니다.
//
// 사용 흐름:
// 1. 사용자가 "cw --root /some/path" 실행
// 2. main()에서 saveRootPath() 호출
// 3. 이 함수가 /some/path를 ~/.cw에 저장
// 4. 다음 cw 실행 시 getRootPath()가 ~/.cw에서 읽음
//
// 에러 처리 패턴:
// - 경로가 존재하지 않으면 에러 반환
// - 쓰기 권한이 없으면 에러 반환
// - 호출자(main)가 에러를 처리
func saveRootPath(path string) error {
	// 1. 홈 디렉토리 경로 구성
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".cw")

	// 2. 사용자가 입력한 경로 정규화 (~/로 시작하면 절대 경로로)
	expandedPath := expandPath(path)

	// 3. 경로 유효성 검증
	// 존재하지 않는 경로나 파일(디렉토리가 아닌)이면 에러
	// 이렇게 하면 사용자가 잘못된 경로를 저장하는 것을 방지
	if info, err := os.Stat(expandedPath); err != nil || !info.IsDir() {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	// 4. ~/.cw 파일에 경로 저장
	// os.WriteFile(파일경로, 내용, 파일권한)
	// 0644 = 사용자는 읽기/쓰기, 다른 사용자는 읽기만
	//
	// 저장할 내용: 원본 경로 문자열 (expandedPath 아니라 path)
	// 이유: 나중에 사용자 친화적으로 "cw --root ~"로 쓸 수 있도록
	return os.WriteFile(configPath, []byte(path), 0644)
}
