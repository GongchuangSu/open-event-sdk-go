package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ResolveOptions 凭证解析的依赖注入配置，方便测试
type ResolveOptions struct {
	AppIdFlag     string
	AppSecretFlag string
	GetEnv        func(string) string
	IsTerminal    func() bool
	ReadLine      func(prompt string) (string, error)
	ReadSecret    func(prompt string) (string, error)
}

// DefaultResolveOptions 创建生产环境的默认配置
func DefaultResolveOptions(appIdFlag, appSecretFlag string) ResolveOptions {
	return ResolveOptions{
		AppIdFlag:     appIdFlag,
		AppSecretFlag: appSecretFlag,
		GetEnv:        os.Getenv,
		IsTerminal: func() bool {
			return term.IsTerminal(int(os.Stdin.Fd()))
		},
		ReadLine: func(prompt string) (string, error) {
			fmt.Fprint(os.Stderr, prompt)
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(line), nil
		},
		ReadSecret: func(prompt string) (string, error) {
			fmt.Fprint(os.Stderr, prompt)
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return "", err
			}
			return string(password), nil
		},
	}
}

// ResolveCredentials 按照 flag → env → interactive 优先级解析 appId 和 appSecret
// appId 和 appSecret 独立解析，可以来自不同来源
func ResolveCredentials(opts ResolveOptions) (appId, appSecret string, err error) {
	appId, err = resolveOne(opts.AppIdFlag, "APP_ID", "APP ID: ", false, opts)
	if err != nil {
		return "", "", fmt.Errorf("缺少凭证: %w", err)
	}

	appSecret, err = resolveOne(opts.AppSecretFlag, "APP_SECRET", "APP Secret: ", true, opts)
	if err != nil {
		return "", "", fmt.Errorf("缺少凭证: %w", err)
	}

	return appId, appSecret, nil
}

func resolveOne(flagVal, envKey, prompt string, isSecret bool, opts ResolveOptions) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}

	if val := opts.GetEnv(envKey); val != "" {
		return val, nil
	}

	if !opts.IsTerminal() {
		flagName := strings.ToLower(strings.ReplaceAll(envKey, "_", "-"))
		return "", fmt.Errorf("请通过 --%s 参数或 %s 环境变量提供", flagName, envKey)
	}

	var val string
	var err error
	if isSecret {
		val, err = opts.ReadSecret(prompt)
	} else {
		val, err = opts.ReadLine(prompt)
	}
	if err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}

	if val == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	return val, nil
}
