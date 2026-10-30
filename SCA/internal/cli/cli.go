package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devsecurity/sca/internal/analyzer"
	"devsecurity/sca/internal/output"
	"devsecurity/sca/internal/source"
	"devsecurity/sca/internal/version"
	"devsecurity/sca/internal/vuln"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "analyze":
		return runAnalyze(ctx, args[1:], stderr)
	case "version":
		fmt.Fprintln(stdout, version.Current)
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runAnalyze(ctx context.Context, args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("sca analyze", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "source directory or ZIP file")
	outputPath := flags.String("output", "", "result JSON file")
	dataPath := flags.String("data", defaultDataPath(), "vulnerability database JSON")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: sca analyze --input <path> --output <result.json> [--data <vulnerabilities.json>]")
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || strings.TrimSpace(*inputPath) == "" || strings.TrimSpace(*outputPath) == "" {
		flags.Usage()
		return 2
	}

	database, err := vuln.Load(*dataPath)
	if err != nil {
		fmt.Fprintf(stderr, "sca: %v\n", err)
		return 1
	}
	outputDirectory := filepath.Dir(*outputPath)
	if err := os.MkdirAll(outputDirectory, 0o750); err != nil {
		fmt.Fprintf(stderr, "sca: create output directory: %v\n", err)
		return 1
	}
	prepared, err := source.PrepareIn(*inputPath, outputDirectory, source.DefaultLimits())
	if err != nil {
		fmt.Fprintf(stderr, "sca: %v\n", err)
		return 1
	}
	defer prepared.Close()

	mavenSetting := strings.TrimSpace(os.Getenv("SCA_USE_MAVEN"))
	scanner := analyzer.Scanner{
		UseMaven: mavenSetting == "" || strings.EqualFold(mavenSetting, "true"),
		Timeout:  10 * time.Minute,
		Database: database,
	}
	result, err := scanner.Analyze(ctx, prepared.Root)
	if err != nil {
		fmt.Fprintf(stderr, "sca: analyze: %v\n", err)
		return 1
	}
	if err := output.WriteJSON(*outputPath, result); err != nil {
		fmt.Fprintf(stderr, "sca: %v\n", err)
		return 1
	}
	return 0
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, `DevSecurity SCA

Usage:
  sca analyze --input <path> --output <result.json> [--data <vulnerabilities.json>]
  sca version`)
}

func defaultDataPath() string {
	if value := strings.TrimSpace(os.Getenv("SCA_VULN_DB")); value != "" {
		return value
	}
	if info, err := os.Stat(filepath.Join("data", "vulnerabilities.json")); err == nil && info.Mode().IsRegular() {
		return filepath.Join("data", "vulnerabilities.json")
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(executable), "data", "vulnerabilities.json")
		if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return "data/vulnerabilities.json"
}
