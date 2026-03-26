package main

import (
	"fmt"
	"os"
	"strings"
)

// GetToken returns the token value for a given account name
func GetToken(accountName string) (string, error) {
	path, err := TokenPath(accountName)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read token for account %s: %w", accountName, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file for account %s is empty", accountName)
	}

	return token, nil
}

// SaveToken saves the token value for a given account name
// Token file is saved with 0600 permissions (user-only read/write)
func SaveToken(accountName string, token string) error {
	if accountName == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Ensure account directory exists
	accountPath, err := AccountPath(accountName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(accountPath, 0700); err != nil {
		return fmt.Errorf("failed to create account directory: %w", err)
	}

	// Write token file with 0600 permissions
	path, err := TokenPath(accountName)
	if err != nil {
		return err
	}

	// Trim trailing newline and write
	token = strings.TrimSpace(token) + "\n"
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token for account %s: %w", accountName, err)
	}

	return nil
}

// GetEnvToken returns the CLAUDE_CODE_OAUTH_TOKEN environment variable
func GetEnvToken() string {
	return os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")
}
