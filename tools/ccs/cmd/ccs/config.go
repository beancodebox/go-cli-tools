package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// AccountsDir returns the path to ~/.claude-accounts
func AccountsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude-accounts"), nil
}

// ClaudeDir returns the path to ~/.claude
func ClaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

// LastActiveFile returns the path to ~/.claude-accounts/.last-active
func LastActiveFile() (string, error) {
	dir, err := AccountsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".last-active"), nil
}

// CurrentBackupDir returns the path to ~/.claude-accounts/current-backup
func CurrentBackupDir() (string, error) {
	dir, err := AccountsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "current-backup"), nil
}

// AccountPath returns the path to a specific account
func AccountPath(accountName string) (string, error) {
	dir, err := AccountsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, accountName), nil
}

// TokenPath returns the path to the token file for a specific account
func TokenPath(accountName string) (string, error) {
	path, err := AccountPath(accountName)
	if err != nil {
		return "", err
	}
	return filepath.Join(path, ".token"), nil
}

// ensureAccountsDirExists creates ~/.claude-accounts if it doesn't exist
func ensureAccountsDirExists() error {
	dir, err := AccountsDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

// ensureClaudeDirExists creates ~/.claude if it doesn't exist
func ensureClaudeDirExists() error {
	dir, err := ClaudeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}
