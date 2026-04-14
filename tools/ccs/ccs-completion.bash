#!/bin/bash

################################################################################
# Claude Code Multi-Account Manager - Bash Completion
#
# 설치 위치: ~/.local/share/bash-completion.d/ccs
# .bashrc에 추가:
#   [ -f ~/.bashrc.ccs ] && source ~/.bashrc.ccs
#
################################################################################

# ccs 명령어 자동완성
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
