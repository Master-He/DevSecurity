package analyzer

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed rules.sc
var rulesScript string

type Scanner struct {
	JoernHome string
	Timeout   time.Duration
}

func (s Scanner) Available() bool {
	_, err := s.command("javasrc2cpg")
	if err != nil {
		return false
	}
	_, err = s.command("joern")
	if err != nil {
		return false
	}
	info, err := os.Stat(filepath.Join(s.JoernHome, "lib",
		"io.joern.joern-cli-"+EngineVersion+".jar"))
	return err == nil && info.Mode().IsRegular()
}

func (s Scanner) Analyze(ctx context.Context, sourceRoot, workRoot string) (Result, error) {
	if !s.Available() {
		return Result{}, fmt.Errorf("Joern %s is not available", EngineVersion)
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	runContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cpgPath := filepath.Join(workRoot, "cpg.bin.zip")
	frontend, err := s.command("javasrc2cpg")
	if err != nil {
		return Result{}, err
	}
	if err := run(runContext, workRoot, frontend, sourceRoot, "--output", cpgPath); err != nil {
		return Result{}, fmt.Errorf("Joern CPG 构建失败: %w", err)
	}
	if info, err := os.Stat(cpgPath); err != nil || info.Size() == 0 {
		return Result{}, fmt.Errorf("Joern 未生成有效 CPG")
	}

	scriptPath := filepath.Join(workRoot, "rules.sc")
	resultPath := filepath.Join(workRoot, "findings.tsv")
	script := strings.ReplaceAll(rulesScript, "__CPG_PATH__", scalaString(cpgPath))
	script = strings.ReplaceAll(script, "__OUTPUT_PATH__", scalaString(resultPath))
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return Result{}, fmt.Errorf("无法生成 Joern 规则脚本")
	}

	joern, err := s.command("joern")
	if err != nil {
		return Result{}, err
	}
	if err := run(runContext, workRoot, joern, "--nocolors", "--script", scriptPath); err != nil {
		return Result{}, fmt.Errorf("Joern 规则执行失败: %w", err)
	}
	rawFindings, err := os.Open(resultPath)
	if err != nil {
		return Result{}, fmt.Errorf("Joern 未生成扫描结果")
	}
	defer rawFindings.Close()

	findings, err := parseFindings(rawFindings, sourceRoot)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Engine:        "joern",
		EngineVersion: EngineVersion,
		SourceMode:    "JOERN_CPGQL",
		Findings:      findings,
		Diagnostics:   []string{},
		CompletedAt:   time.Now().UTC(),
	}, nil
}

func (s Scanner) command(tool string) (string, error) {
	name := tool
	if runtime.GOOS == "windows" {
		name += ".bat"
	}
	path := filepath.Join(s.JoernHome, name)
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("Joern executable %s not found", tool)
	}
	return path, nil
}

func scalaString(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\r", `\r`,
		"\n", `\n`,
	)
	return replacer.Replace(value)
}

func parseFindings(reader *os.File, sourceRoot string) ([]Finding, error) {
	findings := make([]Finding, 0)
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 4 || fields[0] == "" {
			continue
		}
		meta, ok := rules[fields[0]]
		if !ok {
			continue
		}
		filePath, err := relativePath(sourceRoot, unescape(fields[1]))
		if err != nil {
			continue
		}
		line, _ := strconv.Atoi(fields[2])
		code := unescape(fields[3])
		key := fields[0] + "|" + filePath + "|" + strconv.Itoa(line) + "|" + code
		if seen[key] {
			continue
		}
		seen[key] = true
		findings = append(findings, Finding{
			RuleID:         fields[0],
			Title:          meta.title,
			Severity:       meta.severity,
			Confidence:     meta.confidence,
			Category:       meta.category,
			CWE:            meta.cwe,
			Message:        meta.message,
			FilePath:       filePath,
			StartLine:      line,
			EndLine:        line,
			CodeSnippet:    limitText(code, 2000),
			Recommendation: meta.recommendation,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("SAST 结果读取失败")
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].FilePath == findings[j].FilePath {
			return findings[i].StartLine < findings[j].StartLine
		}
		return findings[i].FilePath < findings[j].FilePath
	})
	return findings, nil
}

func relativePath(root, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "<unknown>" {
		return "", errors.New("unknown location")
	}
	if !filepath.IsAbs(value) {
		cleaned := filepath.Clean(filepath.FromSlash(value))
		if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return "", errors.New("invalid relative path")
		}
		return filepath.ToSlash(cleaned), nil
	}
	relative, err := filepath.Rel(root, value)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("path outside source root")
	}
	return filepath.ToSlash(relative), nil
}

func unescape(value string) string {
	unquoted, err := strconv.Unquote(`"` + value + `"`)
	if err != nil {
		return value
	}
	return unquoted
}

func limitText(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[:maximum]
}

func run(ctx context.Context, workRoot, executable string, args ...string) error {
	var command *exec.Cmd
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(executable), ".bat") {
		commandArgs := append([]string{"/d", "/s", "/c", executable}, args...)
		command = exec.CommandContext(ctx, "cmd.exe", commandArgs...)
	} else if runtime.GOOS != "windows" {
		commandArgs := append([]string{executable}, args...)
		command = exec.CommandContext(ctx, "bash", commandArgs...)
	} else {
		command = exec.CommandContext(ctx, executable, args...)
	}
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	command.Dir = workRoot
	configureProcess(command)
	if err := command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		terminateProcessTree(command)
		<-done
		return ctx.Err()
	}
}
