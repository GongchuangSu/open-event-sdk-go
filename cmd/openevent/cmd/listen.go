package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	openevent "github.com/GongchuangSu/open-event-sdk-go"
	"github.com/GongchuangSu/open-event-sdk-go/cmd/openevent/internal"
	"github.com/GongchuangSu/open-event-sdk-go/ws"
	"github.com/spf13/cobra"
)

var (
	appIdFlag     string
	appSecretFlag string
	eventsFlag    string
	jsonFlag      bool
	outputFlag    string
	endpointFlag  string
	noColorFlag   bool
	verboseFlag   bool
	noAckFlag     bool
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "监听并实时打印事件",
	Long: `建立 WebSocket 长连接，实时接收并展示解密后的事件推送。

凭证解析优先级: 命令行参数 > 环境变量 > 交互式输入

示例:
  # 使用命令行参数
  openevent listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET

  # 使用环境变量
  APP_ID=xxx APP_SECRET=yyy openevent listen

  # 交互式输入（仅限 TTY）
  openevent listen

  # JSON 输出 + 管道
  openevent listen --json | jq .

  # 过滤特定事件
  openevent listen --events "kso.app_chat.message.create,kso.app_chat.create"`,
	RunE: runListen,
}

func init() {
	listenCmd.Flags().StringVar(&appIdFlag, "app-id", "", "应用 ID")
	listenCmd.Flags().StringVar(&appSecretFlag, "app-secret", "", "应用密钥")
	listenCmd.Flags().StringVarP(&eventsFlag, "events", "e", "", "事件类型过滤，逗号分隔")
	listenCmd.Flags().BoolVar(&jsonFlag, "json", false, "以 NDJSON 格式输出（适合管道）")
	listenCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "同时输出到文件")
	listenCmd.Flags().StringVar(&endpointFlag, "endpoint", "", "自定义 WebSocket 端点")
	listenCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "禁用彩色输出")
	listenCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "显示详细日志（DEBUG 级别）")
	listenCmd.Flags().BoolVar(&noAckFlag, "no-ack", false, "禁用 ACK 模式（默认开启）")

	rootCmd.AddCommand(listenCmd)
}

func runListen(cmd *cobra.Command, args []string) error {
	printer, err := internal.NewPrinter(internal.PrinterConfig{
		JsonMode: jsonFlag,
		NoColor:  noColorFlag,
		FilePath: outputFlag,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖ 错误: %s\n", err)
		return &ExitError{Code: ExitRuntimeError, Err: err}
	}
	printer.Start()
	defer printer.Close()

	appId, appSecret, err := internal.ResolveCredentials(
		internal.DefaultResolveOptions(appIdFlag, appSecretFlag),
	)
	if err != nil {
		printer.SendStatus(internal.StatusError, fmt.Sprintf("✖ %s", err))
		return &ExitError{Code: ExitCredentialError, Err: err}
	}
	printer.SendStatus(internal.StatusInfo, "✔ 凭证已确认")

	filter := internal.NewEventFilter(eventsFlag)

	opts := []openevent.Option{
		openevent.WithEventHandlerFunc(func(ctx context.Context, e *openevent.Event) error {
			if !filter.Match(e.EventCode()) {
				return nil
			}
			printer.SendEvent(e)
			return nil
		}),
		openevent.WithAckMode(!noAckFlag),
	}

	if verboseFlag {
		opts = append(opts, openevent.WithLogger(internal.NewStderrLogger(openevent.LogLevelDebug)))
	} else {
		opts = append(opts, openevent.WithLogger(openevent.NewNopLogger()))
	}

	if endpointFlag != "" {
		opts = append(opts, openevent.WithEndpoint(endpointFlag))
	}

	client := openevent.NewClient(appId, appSecret, opts...)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		client.Stop()
	}()

	printer.SendStatus(internal.StatusInfo, "正在连接到 WebSocket 服务...")

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Start(ctx)
	}()

	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case err := <-errCh:
			tick.Stop()
			return handleStartError(ctx, printer, err)
		case <-tick.C:
			if client.IsConnected() {
				tick.Stop()
				printer.SendStatus(internal.StatusInfo, "✔ 已连接，正在监听事件... (按 Ctrl+C 退出)")
				select {
				case err := <-errCh:
					return handleStartError(ctx, printer, err)
				case <-ctx.Done():
					<-errCh
					return nil
				}
			}
		case <-ctx.Done():
			<-errCh
			return nil
		}
	}
}

func handleStartError(ctx context.Context, printer *internal.Printer, err error) error {
	if err == nil || ctx.Err() != nil {
		return nil
	}

	var clientErr *ws.ClientError
	if errors.As(err, &clientErr) {
		printer.SendStatus(internal.StatusError,
			fmt.Sprintf("✖ 连接失败: %s — 请检查 APP_ID 和 APP_SECRET", err))
		return &ExitError{Code: ExitCredentialError, Err: err}
	}

	if errors.Is(err, ws.ErrReconnectExceeded) {
		printer.SendStatus(internal.StatusError,
			fmt.Sprintf("✖ 重连次数已耗尽: %s", err))
		return &ExitError{Code: ExitConnectionError, Err: err}
	}

	printer.SendStatus(internal.StatusWarn,
		fmt.Sprintf("⚠ 连接已关闭: %s", err))
	return &ExitError{Code: ExitConnectionError, Err: err}
}
