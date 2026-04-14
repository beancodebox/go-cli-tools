# cw - Workspace Navigator - Shell Configuration
# 자동 로드: ~/.bashrc 또는 ~/.zshrc에서
# [ -f ~/.bashrc.cw ] && source ~/.bashrc.cw

# === Zsh completion fpath 설정 및 초기화 ===
if [ -n "$ZSH_VERSION" ]; then
  fpath=("${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions" $fpath)
  autoload -U compinit && compinit -i
fi

# === Bash completion 로드 ===
# XDG 표준 경로 우선, 레거시 경로도 지원
for dir in \
    "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion.d" \
    "$HOME/.bash_completion.d"; do
  if [ -f "$dir/cw" ]; then
    source "$dir/cw" 2>/dev/null
    break
  fi
done
