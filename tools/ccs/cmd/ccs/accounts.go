package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Account represents a Claude account
type Account struct {
	Name       string // Account name
	Path       string // Full path to account directory
	IsActive   bool   // Is this the current active account
	TokenExists bool   // Does .token file exist
}

// List returns all available accounts
func List() ([]Account, error) {
	if err := ensureAccountsDirExists(); err != nil {
		return nil, err
	}

	dir, err := AccountsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts directory: %w", err)
	}

	currentAccount := GetCurrent()
	var accounts []Account

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip special directories
		if name == "current-backup" || strings.HasPrefix(name, ".") {
			continue
		}

		accountPath := filepath.Join(dir, name)
		tokenPath := filepath.Join(accountPath, ".token")

		// Check if token file exists
		_, err := os.Stat(tokenPath)
		tokenExists := err == nil

		accounts = append(accounts, Account{
			Name:       name,
			Path:       accountPath,
			IsActive:   name == currentAccount,
			TokenExists: tokenExists,
		})
	}

	return accounts, nil
}

// Get returns a specific account by name
func Get(name string) (Account, error) {
	if name == "" {
		return Account{}, fmt.Errorf("account name cannot be empty")
	}

	if err := ensureAccountsDirExists(); err != nil {
		return Account{}, err
	}

	accountPath, err := AccountPath(name)
	if err != nil {
		return Account{}, err
	}

	// Check if account directory exists
	info, err := os.Stat(accountPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Account{}, fmt.Errorf("account %s does not exist", name)
		}
		return Account{}, fmt.Errorf("failed to stat account directory: %w", err)
	}
	if !info.IsDir() {
		return Account{}, fmt.Errorf("account %s is not a directory", name)
	}

	// Check if token file exists
	tokenPath := filepath.Join(accountPath, ".token")
	_, err = os.Stat(tokenPath)
	tokenExists := err == nil

	currentAccount := GetCurrent()

	return Account{
		Name:        name,
		Path:        accountPath,
		IsActive:    name == currentAccount,
		TokenExists: tokenExists,
	}, nil
}

// Use activates an account by updating .last-active
// The shell wrapper sets the token environment variable
func Use(name string) error {
	// Verify account exists
	_, err := Get(name)
	if err != nil {
		return err
	}

	// Update .last-active file
	lastActiveFile, err := LastActiveFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(lastActiveFile, []byte(name+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to update .last-active: %w", err)
	}

	return nil
}

// Delete removes an account
// Prompts for confirmation in interactive mode
func Delete(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	// Verify account exists
	_, err := Get(name)
	if err != nil {
		return err
	}

	accountPath, err := AccountPath(name)
	if err != nil {
		return err
	}

	// Remove directory recursively
	if err := os.RemoveAll(accountPath); err != nil {
		return fmt.Errorf("failed to delete account %s: %w", name, err)
	}

	return nil
}

// SaveCurrent saves the current token as an account
func SaveCurrent(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	accountPath, err := AccountPath(name)
	if err != nil {
		return err
	}

	// Ensure account directory exists
	if err := os.MkdirAll(accountPath, 0700); err != nil {
		return fmt.Errorf("failed to create account directory: %w", err)
	}

	// Save token from environment
	token := os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")
	if token == "" {
		return fmt.Errorf("no token in environment")
	}

	tokenPath := filepath.Join(accountPath, ".token")
	if err := os.WriteFile(tokenPath, []byte(token+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// GetCurrent returns the name of the currently active account
// Returns "" if no account is active
func GetCurrent() string {
	lastActiveFile, err := LastActiveFile()
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(lastActiveFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// Resume reactivates the last active account
func Resume() error {
	lastActiveFile, err := LastActiveFile()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(lastActiveFile)
	if err != nil {
		return fmt.Errorf("failed to read .last-active file: %w", err)
	}

	accountName := strings.TrimSpace(string(data))
	if accountName == "" {
		return fmt.Errorf("no previous account to resume")
	}

	return Use(accountName)
}

// Verify checks the integrity of an account
// Compares saved account state with current ~/.claude state
// Returns a list of issues found, empty if no issues
func Verify(name string) []string {
	account, err := Get(name)
	if err != nil {
		return []string{fmt.Sprintf("Account error: %v", err)}
	}

	var issues []string

	// 1. Token verification
	var savedToken, envToken string
	var tokenExists, tokenEnvSet bool

	// Check saved token
	if token, err := GetToken(name); err == nil {
		savedToken = token
		tokenExists = true
	}

	// Check environment token
	if envToken = os.Getenv("CLAUDE_CODE_OAUTH_TOKEN"); envToken != "" {
		tokenEnvSet = true
	}

	// Report token status
	if !tokenExists {
		issues = append(issues, "Saved token: Not found")
	}

	if !tokenEnvSet {
		issues = append(issues, "Environment token: Not set")
	}

	if tokenExists && tokenEnvSet && savedToken != envToken {
		issues = append(issues, "Token mismatch: saved token != environment token")
	}

	// 2. Configuration file verification (compare with current ~/.claude)
	claudeDir, err := ClaudeDir()
	if err != nil {
		return append(issues, fmt.Sprintf("Failed to get ~/.claude path: %v", err))
	}

	// Check if current ~/.claude exists
	if _, err := os.Stat(claudeDir); err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, "Current ~/.claude directory not found")
		}
	}

	// Compare settings.json
	accountSettingsPath := filepath.Join(account.Path, "settings.json")
	claudeSettingsPath := filepath.Join(claudeDir, "settings.json")

	if _, err := os.Stat(accountSettingsPath); err == nil {
		// Account has settings.json, compare with current
		if _, err := os.Stat(claudeSettingsPath); err == nil {
			// Both exist, compare content
			accountContent, _ := os.ReadFile(accountSettingsPath)
			claudeContent, _ := os.ReadFile(claudeSettingsPath)
			if string(accountContent) != string(claudeContent) {
				issues = append(issues, "settings.json: Mismatch (account != current)")
			}
		}
	}

	// Compare claude.json
	accountClaudeJsonPath := filepath.Join(account.Path, "claude.json")
	claudeClaudeJsonPath := filepath.Join(claudeDir, "claude.json")

	if _, err := os.Stat(accountClaudeJsonPath); err == nil {
		// Account has claude.json, compare with current
		if _, err := os.Stat(claudeClaudeJsonPath); err == nil {
			// Both exist, compare content
			accountContent, _ := os.ReadFile(accountClaudeJsonPath)
			claudeContent, _ := os.ReadFile(claudeClaudeJsonPath)
			if string(accountContent) != string(claudeContent) {
				issues = append(issues, "claude.json: Mismatch (account != current)")
			}
		}
	}

	return issues
}

// PromptForConfirmation prompts user to type account name to confirm deletion
func PromptForConfirmation(accountName string) bool {
	reader := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("Are you sure? (type '%s' to confirm): ", accountName)
	fmt.Print(prompt)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	return response == accountName
}
