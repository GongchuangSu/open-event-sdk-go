package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	ExitOK              = 0
	ExitCredentialError = 1
	ExitConnectionError = 2
	ExitRuntimeError    = 3
)

// ExitError 携带退出码的错误
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

var rootCmd = &cobra.Command{
	Use:   "openevent",
	Short: "WPS 开放平台事件监听 CLI",
	Long: `openevent 是基于 open-event-sdk-go 的命令行工具，
通过 WebSocket 长连接实时接收并展示解密后的事件数据。

快速开始:
  openevent listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET`,
}

var (
	buildVersion string
	buildCommit  string
	buildDate    string
)

// Execute 执行根命令，返回退出码
func Execute(version, commit, date string) int {
	buildVersion = version
	buildCommit = commit
	buildDate = date

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}
		return ExitRuntimeError
	}
	return ExitOK
}
