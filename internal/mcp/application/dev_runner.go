package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// WithDevRunnerTools 本地开发辅助工具
// 这组工具让 AI 能像本地助手一样：查看项目目录树(fs_tree)、读取文件(fs_cat)、运行项目/脚本(code_run)。
// - fs_tree：列出指定目录的树形结构（可控制深度/忽略模式），帮助 AI 感知项目布局。
// - fs_cat ：读取指定文件的内容（可限制最大字节），帮助 AI 查看未直接提供的代码。
// - code_run：在给定根目录下自动/按命令运行项目或单文件，返回 stdout/stderr/exit code，并给出基于错误输出的建议。
func WithDevRunnerTools() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {

		// fs_tree 目录树查看，让AI感知在哪个目录下运行代码
		toolTree := mcp.NewTool("fs_tree",
			mcp.WithDescription("List a directory as a plain text tree to understand project layout."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Directory path to list")),
			// depth 最大遍历深度
			mcp.WithNumber("depth", mcp.Description("Max depth to traverse (default 4)")),
			// ignore 如 node_modules, *.log
			mcp.WithString("ignore", mcp.Description("Comma-separated glob patterns to ignore (optional)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolTree)
		toolSet.HandlerFunc[toolTree.Name] = HandleFsTree

		// fs_cat 读取文件里的内容
		toolCat := mcp.NewTool("fs_cat",
			mcp.WithDescription("Read a file content to inspect code that was not provided in the prompt."),
			// 文件路径
			mcp.WithString("path", mcp.Required(), mcp.Description("File path to read")),
			// 最大读取字节数
			mcp.WithNumber("max_bytes", mcp.Description("Max bytes to read (default 65536)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolCat)
		toolSet.HandlerFunc[toolCat.Name] = HandleFsCat

		// code_run 运行命令行
		toolRun := mcp.NewTool("code_run",
			// 工具用途：在本地命令行运行项目/脚本，返回 stdout/stderr/exit code，并基于错误输出给建议
			mcp.WithDescription("Run a code file/project locally in the given root directory with the EXACT command provided by the AI,return stdout"),
			// required ：工作目录（项目根目录）
			mcp.WithString("root", mcp.Required(), mcp.Description("Working directory of the project")),
			// 可选参数：运行前将 content 写入到 root 下的 file（相对路径）
			//mcp.WithString("file", mcp.Description("Optional file path relative to root to write/update before running")),
			//mcp.WithString("content", mcp.Description("Optional content to write to `file` before running")),
			// required ：显式运行命令
			mcp.WithString("command", mcp.Description("Explicit shell command to run under the root directory(eg `python main.py`,`go run cmd/host`,`npm run dev`)")),
			// optional ：超时（秒），默认 120s
			mcp.WithNumber("timeout_sec", mcp.Description("Timeout in seconds (default 120)")),
			// optional ：传给程序的标准输入
			mcp.WithString("stdin", mcp.Description("Optional STDIN to pass to the program")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolRun)
		toolSet.HandlerFunc[toolRun.Name] = HandleCodeRun
	}
}

func HandleFsCat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	p, _ := args["path"].(string)
	if p == "" {
		return mcp.NewToolResultError("missing required arg: path"), nil
	}
	maxF, _ := args["max_bytes"].(float64)
	maxBytes := 64 * 1024
	if maxF > 0 {
		maxBytes = int(maxF)
	}
	content, truncated, err := utils.ReadFileMax(p, maxBytes)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	header := fmt.Sprintf("### fs_cat: %s (max_bytes=%d, truncated=%v)\n\n", p, maxBytes, truncated)
	return mcp.NewToolResultText(header + content), nil
}

func HandleFsTree(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	root, _ := args["path"].(string)
	if root == "" {
		return mcp.NewToolResultError("missing required arg: path"), nil
	}
	depthF, _ := args["depth"].(float64)
	depth := 4
	if depthF > 0 {
		depth = int(depthF)
	}
	ignoreStr, _ := args["ignore"].(string)
	var ignores []string
	if ignoreStr != "" {
		for _, s := range strings.Split(ignoreStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				ignores = append(ignores, s)
			}
		}
	}
	out, err := buildTreeText(root, depth, ignores)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

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
