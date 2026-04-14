#!/bin/bash

################################################################################
# cw - Workspace Navigator - Bash Completion
#
# 설치 위치: ~/.local/share/bash-completion.d/cw
# .bashrc에 추가:
#   [ -f ~/.bashrc.ccs ] && source ~/.bashrc.ccs  (ccs 포함되어 있음)
#
################################################################################

_cw_completion() {
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
  _get_saved_accounts() {
    if command -v ccs &> /dev/null; then
      ccs list --plain 2>/dev/null
    fi
  }

  # 1. 첫 번째 인자 (플래그) 자동완성
  if [[ "$cur" == -* ]]; then
    local flags="--resume --root -r --account -a --help"
    COMPREPLY=($(compgen -W "$flags" -- "$cur"))
    return 0
  fi

  # 2. 이전 인자에 따라 값 자동완성
  case "$prev" in
    --root|-r)
      # 디렉토리 자동완성
      COMPREPLY=($(compgen -d -- "$cur"))
      ;;
    --account|-a)
      # 계정명 자동완성
      local accounts=$(_get_saved_accounts)
      COMPREPLY=($(compgen -W "$accounts" -- "$cur"))
      ;;
    *)
      # 다른 경우는 자동완성 없음
      COMPREPLY=()
      ;;
  esac
}

# 자동완성 함수 등록
complete -o bashdefault -o default -o nospace -F _cw_completion cw
