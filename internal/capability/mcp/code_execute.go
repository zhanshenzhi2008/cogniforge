package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	codeMaxBytes       = 64 * 1024
	codeOutputMaxBytes = 100 * 1024
	codeDefaultTimeout = 30
	codeMaxTimeout     = 60
)

// codeExecute 在本地子进程沙箱中执行代码（超时截断；非完整容器隔离）
func (e *BuiltInExecutor) codeExecute(language string, args map[string]interface{}) (map[string]interface{}, error) {
	if os.Getenv("CODE_EXECUTE_DISABLED") == "1" || strings.EqualFold(os.Getenv("CODE_EXECUTE_DISABLED"), "true") {
		return map[string]interface{}{
			"language": language,
			"output":   "代码执行已禁用（CODE_EXECUTE_DISABLED）",
			"status":   "disabled",
		}, nil
	}

	code, ok := args["code"].(string)
	if !ok || strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("code 参数必填")
	}
	if len(code) > codeMaxBytes {
		return nil, fmt.Errorf("代码过长（最多 %d 字节）", codeMaxBytes)
	}
	if strings.ContainsRune(code, 0) {
		return nil, fmt.Errorf("代码包含非法字符")
	}

	timeoutSec := codeDefaultTimeout
	if t, ok := args["timeout"].(float64); ok {
		timeoutSec = int(t)
	}
	if timeoutSec <= 0 {
		timeoutSec = codeDefaultTimeout
	}
	if timeoutSec > codeMaxTimeout {
		timeoutSec = codeMaxTimeout
	}

	dir, err := os.MkdirTemp("", "cf-code-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(dir)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd, err := buildCodeCommand(ctx, language, code, dir)
	if err != nil {
		return map[string]interface{}{
			"language": language,
			"output":   err.Error(),
			"status":   "error",
		}, nil
	}
	cmd.Dir = dir
	// 收紧环境：不继承敏感变量以外的业务密钥（仍保留 PATH 等基础）
	cmd.Env = sandboxEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	out := trimOutput(stdout.String(), stderr.String())
	status := "ok"
	exitCode := 0
	if ctx.Err() == context.DeadlineExceeded {
		status = "timeout"
		if out == "" {
			out = fmt.Sprintf("执行超时（%ds）", timeoutSec)
		} else {
			out += fmt.Sprintf("\n\n[超时截断 %ds]", timeoutSec)
		}
		exitCode = -1
	} else if runErr != nil {
		status = "error"
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
			if out == "" {
				out = runErr.Error()
			}
		}
	}

	return map[string]interface{}{
		"language":   language,
		"output":     out,
		"status":     status,
		"exit_code":  exitCode,
		"elapsed_ms": elapsed,
	}, nil
}

func buildCodeCommand(ctx context.Context, language, code, dir string) (*exec.Cmd, error) {
	switch strings.ToLower(language) {
	case "python":
		bin, err := lookFirst("python3", "python")
		if err != nil {
			return nil, fmt.Errorf("未找到 python3/python：%w", err)
		}
		path := filepath.Join(dir, "main.py")
		if err := os.WriteFile(path, []byte(code), 0o600); err != nil {
			return nil, err
		}
		// -I：隔离模式，忽略用户 site；-B：不写 pyc
		return exec.CommandContext(ctx, bin, "-I", "-B", path), nil

	case "javascript", "js", "node":
		bin, err := lookFirst("node")
		if err != nil {
			return nil, fmt.Errorf("未找到 node：%w", err)
		}
		path := filepath.Join(dir, "main.js")
		if err := os.WriteFile(path, []byte(code), 0o600); err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, bin, path), nil

	case "go":
		bin, err := lookFirst("go")
		if err != nil {
			return nil, fmt.Errorf("未找到 go：%w", err)
		}
		src := code
		if !strings.Contains(src, "package ") {
			src = "package main\n\n" + src
		}
		path := filepath.Join(dir, "main.go")
		if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
			return nil, err
		}
		// 独立 module，避免污染工作区
		mod := exec.CommandContext(ctx, bin, "mod", "init", "cf-sandbox")
		mod.Dir = dir
		mod.Env = sandboxEnv()
		_ = mod.Run()
		return exec.CommandContext(ctx, bin, "run", path), nil

	default:
		return nil, fmt.Errorf("不支持的语言: %s", language)
	}
}

func lookFirst(names ...string) (string, error) {
	var last error
	for _, n := range names {
		p, err := exec.LookPath(n)
		if err == nil {
			return p, nil
		}
		last = err
	}
	if last == nil {
		last = fmt.Errorf("not found")
	}
	return "", last
}

func sandboxEnv() []string {
	keep := map[string]bool{
		"PATH": true, "HOME": true, "TMPDIR": true, "TEMP": true, "TMP": true,
		"LANG": true, "LC_ALL": true, "LC_CTYPE": true,
		"SystemRoot": true, "COMSPEC": true, // Windows
	}
	var out []string
	for _, e := range os.Environ() {
		i := strings.IndexByte(e, '=')
		if i <= 0 {
			continue
		}
		k := e[:i]
		if keep[k] || strings.HasPrefix(k, "GO") {
			out = append(out, e)
		}
	}
	out = append(out, "PYTHONDONTWRITEBYTECODE=1")
	return out
}

func trimOutput(stdout, stderr string) string {
	var b strings.Builder
	if stdout != "" {
		b.WriteString(stdout)
	}
	if stderr != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(stderr)
	}
	s := b.String()
	if len(s) > codeOutputMaxBytes {
		return s[:codeOutputMaxBytes] + "\n…[输出已截断]"
	}
	return s
}
