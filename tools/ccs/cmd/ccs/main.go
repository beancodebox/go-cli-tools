package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "list":
		cmdList(args)
	case "use":
		cmdUse(args)
	case "init":
		cmdInit(args)
	case "status":
		cmdStatus(args)
	case "verify":
		cmdVerify(args)
	case "get-current":
		cmdGetCurrent(args)
	case "get-token":
		cmdGetToken(args)
	case "resume":
		cmdResume(args)
	case "save-current":
		cmdSaveCurrent(args)
	case "delete":
		cmdDelete(args)
	case "help":
		printHelp()
	case "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

// cmdList lists all accounts
func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	plainFlag := fs.Bool("plain", false, "Print plain format (one account per line)")
	fs.Parse(args)

	accounts, err := List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts found")
		return
	}

	if *plainFlag {
		for _, acc := range accounts {
			fmt.Println(acc.Name)
		}
	} else {
		for _, acc := range accounts {
			prefix := " "
			if acc.IsActive {
				prefix = "*"
			}
			tokenStatus := "no"
			if acc.TokenExists {
				tokenStatus = "yes"
			}
			fmt.Printf("%s %-20s (token: %s)\n", prefix, acc.Name, tokenStatus)
		}
	}
}

// cmdUse activates an account
func cmdUse(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ccs use <account-name>\n")
		os.Exit(1)
	}

	accountName := args[0]
	if err := Use(accountName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Activated account: %s\n", accountName)
}

// cmdInit initializes a new account
func cmdInit(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ccs init <account-name>\n")
		os.Exit(1)
	}

	accountName := args[0]

	// 1. 계정명 검증
	if accountName == "" {
		fmt.Fprintf(os.Stderr, "Error: Account name is required\n")
		os.Exit(1)
	}

	if !isValidAccountName(accountName) {
		fmt.Fprintf(os.Stderr, "Error: Account name must contain only letters, numbers, underscores, and hyphens\n")
		os.Exit(1)
	}

	// Header
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("  Claude Code Multi-Account Init")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Account: '%s'\n", accountName)
	fmt.Println()

	// 2. claude setup-token 실행
	fmt.Println("Step 1: Setting up Claude Code token...")
	fmt.Println()

	// Check if claude command exists
	cmd := exec.Command("which", "claude")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Claude Code CLI not found. Please install it first.\n")
		os.Exit(1)
	}

	// Run claude setup-token (interactive)
	setupCmd := exec.Command("claude", "setup-token")
	setupCmd.Stdin = os.Stdin
	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr
	if err := setupCmd.Run(); err != nil {
		// Don't exit on error, user might have cancelled
	}

	fmt.Println()
	fmt.Println("Step 2: Getting token...")
	fmt.Println()

	// 3. 토큰 입력받기
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your API token (from https://claude.ai/account/keys): ")
	token, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token: %v\n", err)
		os.Exit(1)
	}

	token = strings.TrimSpace(token)

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: Token is required\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(token, "sk-") {
		fmt.Fprintf(os.Stderr, "Error: Invalid token format (should start with 'sk-')\n")
		os.Exit(1)
	}

	fmt.Println()

	// 4. 후처리: 계정 저장
	fmt.Println("Step 3: Saving account...")

	// Create account directory
	accountPath, err := AccountPath(accountName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(accountPath, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating account directory: %v\n", err)
		os.Exit(1)
	}

	// Create placeholder directories
	for _, dir := range []string{"conversations"} {
		dirPath := filepath.Join(accountPath, dir)
		if err := os.MkdirAll(dirPath, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Save token
	if err := SaveToken(accountName, token); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("  ✅ Account '%s' initialized!\n", accountName)
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("To use this account:")
	fmt.Printf("  ccs use %s\n", accountName)
	fmt.Println("  claude")
}

// cmdStatus prints the current account status
func cmdStatus(args []string) {
	current := GetCurrent()

	if current == "" {
		fmt.Println("Status: No active account")
		os.Exit(0)
	}

	account, err := Get(current)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	tokenEmoji := "❌"
	if account.TokenExists {
		tokenEmoji = "✓"
	}
	fmt.Printf("Status: Last active account '%s' (%s Token)\n", account.Name, tokenEmoji)
}

// cmdVerify verifies account integrity
func cmdVerify(args []string) {
	var accountName string

	if len(args) > 0 {
		accountName = args[0]
	} else {
		accountName = GetCurrent()
		if accountName == "" {
			fmt.Fprintf(os.Stderr, "No active account and no account specified\n")
			os.Exit(1)
		}
	}

	fmt.Printf("Verifying account: '%s'\n", accountName)
	fmt.Println()

	// Token verification
	savedToken := ""
	tokenExists := false
	if token, err := GetToken(accountName); err == nil {
		savedToken = token
		tokenExists = true
		fmt.Printf("  ✓ Saved token: Found (%d chars)\n", len(token))
	} else {
		fmt.Println("  ✗ Saved token: Not found")
	}

	envToken := os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")
	tokenEnvSet := envToken != ""
	if tokenEnvSet {
		fmt.Printf("  ✓ Environment token: Set (%d chars)\n", len(envToken))
	} else {
		fmt.Println("  ✗ Environment token: Not set")
	}

	// Token match check
	if tokenExists && tokenEnvSet {
		if savedToken == envToken {
			fmt.Println("  ✓ Token match: YES ⭐")
		} else {
			fmt.Println("  ✗ Token match: NO ❌ (다른 토큰이 설정되어 있음)")
		}
	}

	fmt.Println()

	// File verification (compare with current ~/.claude)
	claudeDir, _ := ClaudeDir()
	account, _ := Get(accountName)

	filesOk := true

	// Check settings.json
	accountSettingsPath := filepath.Join(account.Path, "settings.json")
	claudeSettingsPath := filepath.Join(claudeDir, "settings.json")

	if _, err := os.Stat(accountSettingsPath); err == nil {
		if _, err := os.Stat(claudeSettingsPath); err == nil {
			accountContent, _ := os.ReadFile(accountSettingsPath)
			claudeContent, _ := os.ReadFile(claudeSettingsPath)
			if string(accountContent) == string(claudeContent) {
				fmt.Println("  ✓ settings.json: Match ⭐")
			} else {
				fmt.Println("  ✗ settings.json: Mismatch ❌")
				filesOk = false
			}
		}
	}

	// Check claude.json
	accountClaudeJsonPath := filepath.Join(account.Path, "claude.json")
	claudeClaudeJsonPath := filepath.Join(claudeDir, "claude.json")

	if _, err := os.Stat(accountClaudeJsonPath); err == nil {
		if _, err := os.Stat(claudeClaudeJsonPath); err == nil {
			accountContent, _ := os.ReadFile(accountClaudeJsonPath)
			claudeContent, _ := os.ReadFile(claudeClaudeJsonPath)
			if string(accountContent) == string(claudeContent) {
				fmt.Println("  ✓ claude.json: Match ⭐")
			} else {
				fmt.Println("  ✗ claude.json: Mismatch ❌")
				filesOk = false
			}
		}
	}

	fmt.Println()

	// Final result
	passed := tokenEnvSet && filesOk && savedToken == envToken
	if passed {
		fmt.Println("✅ Account verification PASSED ✅")
		os.Exit(0)
	} else {
		fmt.Println("❌ Account verification FAILED ❌")
		os.Exit(1)
	}
}

// cmdGetCurrent prints the current account name
func cmdGetCurrent(args []string) {
	current := GetCurrent()
	if current != "" {
		fmt.Println(current)
	}
}

// cmdGetToken prints the token for an account
func cmdGetToken(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ccs get-token <account-name>\n")
		os.Exit(1)
	}

	accountName := args[0]
	token, err := GetToken(accountName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(token)
}

// cmdResume reactivates the last account
func cmdResume(args []string) {
	if err := Resume(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	current := GetCurrent()
	fmt.Printf("ccs: account resumed (%s)\n", current)
}

// cmdSaveCurrent saves the current ~/.claude state as an account
func cmdSaveCurrent(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ccs save-current <account-name>\n")
		os.Exit(1)
	}

	accountName := args[0]
	if err := SaveCurrent(accountName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved current state as account: %s\n", accountName)
}

// cmdDelete deletes an account
func cmdDelete(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ccs delete <account-name>\n")
		os.Exit(1)
	}

	accountName := args[0]

	// Show warning
	fmt.Printf("⚠️  This will permanently delete account '%s'\n", accountName)
	fmt.Println("Files to be deleted:")
	fmt.Println("  - Settings")
	fmt.Println("  - Token")
	fmt.Println("  - Conversation history")
	fmt.Println()

	// Prompt for confirmation (must type account name)
	if !PromptForConfirmation(accountName) {
		fmt.Println("Deletion cancelled")
		os.Exit(0)
	}

	if err := Delete(accountName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Account '%s' deleted\n", accountName)
}

// isValidAccountName validates account name format
func isValidAccountName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// printHelp prints the help message
func printHelp() {
	help := `ccs: Claude Code Multi-Account Switcher

Usage: ccs <command> [args...]

Commands:
  list [--plain]           List all accounts
  use <name>               Activate an account
  init <name>              Initialize a new account
  status                   Show current active account
  verify [name]            Verify account integrity
  get-current              Get active account name (for scripting)
  get-token <name>         Get token for an account (for scripting)
  resume                   Reactivate last active account
  save-current <name>      Save current ~/.claude state as account
  delete <name>            Delete an account
  help                     Show this help message

Examples:
  ccs use work              # Activate 'work' account
  ccs list                  # List all accounts
  ccs status                # Show active account
  ccs verify                # Verify active account
  ccs delete personal       # Delete 'personal' account
`
	fmt.Print(help)
}
