package dev_runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func HandleCodeRun(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	args := req.GetArguments()
	root, _ := args["root"].(string)
	if root == "" {
		return mcp.NewToolResultError("missing required arg: root"), nil
	}
	//fileRel, _ := args["file"].(string)
	//content, _ := args["content"].(string)
	cmdStr, _ := args["command"].(string)
	if strings.TrimSpace(cmdStr) == "" {
		return mcp.NewToolResultError("missing required arg: command"), nil
	}
	stdin, _ := args["stdin"].(string)
	timeoutF, _ := args["timeout_sec"].(float64)
	timeout := 120 * time.Second
	if timeoutF > 0 {
		timeout = time.Duration(timeoutF) * time.Second
	}

	// 可选：写入文件
	//if fileRel != "" && content != "" {
	//	abs := filepath.Join(root, filepath.Clean(fileRel))
	//	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
	//		return mcp.NewToolResultError("mkdir: " + err.Error()), nil
	//	}
	//	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
	//		return mcp.NewToolResultError("write file: " + err.Error()), nil
	//	}
	//}

	// 运行命令
	stdout, stderr, exitCode, runErr := runShell(ctx, root, cmdStr, stdin, timeout)

	var buf strings.Builder
	buf.WriteString("### code_run\n\n")
	buf.WriteString("**dir:** " + root + "\n\n")
	buf.WriteString("**cmd:**\n```sh\n" + cmdStr + "\n```\n\n")
	buf.WriteString(fmt.Sprintf("**exit_code:** %d\n\n", exitCode))

	if s := strings.TrimSpace(stdout); s != "" {
		buf.WriteString("**stdout:**\n```\n" + tail(s, 10000) + "\n```\n\n")
	} else {
		buf.WriteString("**stdout:** (empty)\n\n")
	}
	if s := strings.TrimSpace(stderr); s != "" {
		buf.WriteString("**stderr:**\n```\n" + tail(s, 10000) + "\n```\n\n")
	} else {
		buf.WriteString("**stderr:** (empty)\n\n")
	}

	if runErr != nil && !errors.Is(runErr, context.DeadlineExceeded) {
		logger.Warnf("code_run: %v", runErr)
	}
	return mcp.NewToolResultText(buf.String()), nil
}

func tail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}

// ===== 辅助：运行命令 =====
func runShell(ctx context.Context, dir, cmdStr, stdin string, timeout time.Duration) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-lc", cmdStr)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			if status, ok := ee.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = 1
			}
		} else if errors.Is(err, context.DeadlineExceeded) {
			exitCode = 124
		} else {
			exitCode = 1
		}
	}
	return stdout.String(), stderr.String(), exitCode, err
}
