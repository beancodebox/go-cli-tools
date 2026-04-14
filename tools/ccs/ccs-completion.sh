#!/bin/sh

################################################################################
# Claude Code Multi-Account Manager - Shell Completion (Bash & Zsh)
#
# 설치 위치: ~/.bash_completion.d/ccs
# .bashrc에 추가:
#   [ -f ~/.bash_completion.d/ccs ] && source ~/.bash_completion.d/ccs
#
################################################################################

if [ -n "$ZSH_VERSION" ]; then
  # ============================================================================
  # === ZSH COMPLETION ===
  # ============================================================================
  _ccs_completion_zsh() {
    local -a commands accounts

    commands=(
      'init:새 계정 초기화'
      'use:계정 활성화'
      'verify:계정 검증'
      'save-current:현재 설정 저장'
      'list:계정 목록 표시'
      'status:현재 활성 계정 상태'
      'get-current:활성 계정명 출력'
      'resume:마지막 활성 계정 복구'
      'delete:계정 삭제'
      'get-token:토큰 조회 (스크립팅용)'
      'help:도움말 표시'
    )

    # 1. 첫 번째 인자: 명령어 완성
    if (( CURRENT == 2 )); then
      _describe 'command' commands
      return 0
    fi

    # 2. 두 번째 이상 인자: 명령어에 따라 완성
    case "$words[2]" in
      init|use|verify|save-current|delete|get-token)
        # 이 명령어들은 계정명을 인자로 받음
        accounts=($(~/.local/bin/ccs list --plain 2>/dev/null))
        _describe 'account' accounts
        ;;
      *)
        # 다른 명령어는 추가 인자 없음
        ;;
    esac
  }

  # zsh 자동완성 함수 등록
  compdef _ccs_completion_zsh ccs

elif [ -n "$BASH_VERSION" ]; then
  # ============================================================================
  # === BASH COMPLETION ===
  # ============================================================================
  _ccs_completion() {
    local cur prev words cword

    # bash-completion이 제공하는 변수들 (호환성)
    _get_comp_words_by_ref -n ':' cur prev words cword 2>/dev/null || {
      COMPREPLY=()
      cur="${COMP_WORDS[COMP_CWORD]}"
      prev="${COMP_WORDS[COMP_CWORD-1]}"
      words=("${COMP_WORDS[@]}")
      cword=$COMP_CWORD
    }

    # 저장된 계정 목록 조회
    local accounts_dir="${HOME}/.claude-accounts"
    _get_saved_accounts() {
      if [ -d "$accounts_dir" ]; then
        ls -d "$accounts_dir"/*/ 2>/dev/null | while read -r dir; do
          local name=$(basename "$dir")
          # current-backup 제외
          [ "$name" != "current-backup" ] && echo "$name"
        done
      fi
    }

    # 1. 첫 번째 인자 (서브명령어) 자동완성
    if [ $cword -eq 1 ]; then
      local commands="init use verify save-current list status delete resume get-current get-token help"
      COMPREPLY=($(compgen -W "$commands" -- "$cur"))
      return 0
    fi

    # 2. 두 번째 인자부터는 명령어에 따라 다름
    case "$prev" in
      init|use|verify|save-current|delete|get-token)
        # 계정명 자동완성
        local accounts=$(_get_saved_accounts)
        COMPREPLY=($(compgen -W "$accounts" -- "$cur"))
        ;;
      *)
        # 다른 명령어는 자동완성 없음
        COMPREPLY=()
        ;;
    esac
  }

  # 자동완성 함수 등록
  complete -o bashdefault -o default -o nospace -F _ccs_completion ccs
fi
