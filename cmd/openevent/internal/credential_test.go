package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCredentials(t *testing.T) {
	tests := []struct {
		name          string
		opts          ResolveOptions
		wantAppId     string
		wantAppSecret string
		wantErr       bool
		errContains   string
	}{
		{
			name: "both from flags",
			opts: ResolveOptions{
				AppIdFlag:     "flag_id",
				AppSecretFlag: "flag_secret",
				GetEnv:        func(string) string { return "" },
				IsTerminal:    func() bool { return false },
			},
			wantAppId:     "flag_id",
			wantAppSecret: "flag_secret",
		},
		{
			name: "both from env",
			opts: ResolveOptions{
				GetEnv: func(key string) string {
					switch key {
					case "APP_ID":
						return "env_id"
					case "APP_SECRET":
						return "env_secret"
					}
					return ""
				},
				IsTerminal: func() bool { return false },
			},
			wantAppId:     "env_id",
			wantAppSecret: "env_secret",
		},
		{
			name: "flag takes precedence over env",
			opts: ResolveOptions{
				AppIdFlag:     "flag_id",
				AppSecretFlag: "flag_secret",
				GetEnv: func(key string) string {
					return "env_value"
				},
				IsTerminal: func() bool { return false },
			},
			wantAppId:     "flag_id",
			wantAppSecret: "flag_secret",
		},
		{
			name: "mixed sources: id from flag, secret from env",
			opts: ResolveOptions{
				AppIdFlag: "flag_id",
				GetEnv: func(key string) string {
					if key == "APP_SECRET" {
						return "env_secret"
					}
					return ""
				},
				IsTerminal: func() bool { return false },
			},
			wantAppId:     "flag_id",
			wantAppSecret: "env_secret",
		},
		{
			name: "interactive input",
			opts: ResolveOptions{
				GetEnv:     func(string) string { return "" },
				IsTerminal: func() bool { return true },
				ReadLine: func(prompt string) (string, error) {
					return "interactive_id", nil
				},
				ReadSecret: func(prompt string) (string, error) {
					return "interactive_secret", nil
				},
			},
			wantAppId:     "interactive_id",
			wantAppSecret: "interactive_secret",
		},
		{
			name: "non-TTY without credentials fails",
			opts: ResolveOptions{
				GetEnv:     func(string) string { return "" },
				IsTerminal: func() bool { return false },
			},
			wantErr:     true,
			errContains: "--app-id",
		},
		{
			name: "interactive empty input fails",
			opts: ResolveOptions{
				GetEnv:     func(string) string { return "" },
				IsTerminal: func() bool { return true },
				ReadLine: func(prompt string) (string, error) {
					return "", nil
				},
				ReadSecret: func(prompt string) (string, error) {
					return "", nil
				},
			},
			wantErr:     true,
			errContains: "输入不能为空",
		},
		{
			name: "interactive read error propagates",
			opts: ResolveOptions{
				GetEnv:     func(string) string { return "" },
				IsTerminal: func() bool { return true },
				ReadLine: func(prompt string) (string, error) {
					return "", fmt.Errorf("read error")
				},
			},
			wantErr:     true,
			errContains: "读取输入失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appId, appSecret, err := ResolveCredentials(tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantAppId, appId)
			assert.Equal(t, tt.wantAppSecret, appSecret)
		})
	}
}

func TestResolveCredentials_SecretUsesReadSecret(t *testing.T) {
	var usedReadLine, usedReadSecret bool
	opts := ResolveOptions{
		AppIdFlag: "id_from_flag",
		GetEnv:    func(string) string { return "" },
		IsTerminal: func() bool {
			return true
		},
		ReadLine: func(prompt string) (string, error) {
			usedReadLine = true
			return "line", nil
		},
		ReadSecret: func(prompt string) (string, error) {
			usedReadSecret = true
			return "secret", nil
		},
	}

	_, appSecret, err := ResolveCredentials(opts)
	require.NoError(t, err)
	assert.Equal(t, "secret", appSecret)
	assert.False(t, usedReadLine, "ReadLine should not be called for secret")
	assert.True(t, usedReadSecret, "ReadSecret should be called for secret")
}
