package analyzer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devsecurity/sca/internal/model"
)

func MavenAvailable() bool {
	_, err := exec.LookPath("mvn")
	return err == nil
}

func GenerateDependencyTree(ctx context.Context, pomPath string, timeout time.Duration) ([]model.Component, error) {
	runContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	command := exec.CommandContext(runContext, "mvn",
		"-B", "-DskipTests", "-Dstyle.color=never", "dependency:tree")
	command.Dir = filepath.Dir(pomPath)
	command.Env = safeEnvironment()
	output, err := command.CombinedOutput()
	if runContext.Err() != nil {
		return nil, fmt.Errorf("maven dependency:tree timed out")
	}
	if err != nil {
		return nil, fmt.Errorf("maven dependency:tree failed: %w", err)
	}
	components := ParseDependencyTree(string(output))
	if len(components) == 0 {
		return nil, fmt.Errorf("maven dependency:tree produced no components")
	}
	return components, nil
}

func ParseDependencyTree(output string) []model.Component {
	var components []model.Component
	var stack []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \r")
		if info := strings.Index(line, "[INFO]"); info >= 0 {
			line = line[info+len("[INFO]"):]
		}
		if strings.HasPrefix(line, " ") {
			line = line[1:]
		}
		plus := strings.Index(line, "+-")
		last := strings.Index(line, `\-`)
		marker := plus
		if marker < 0 || (last >= 0 && last < marker) {
			marker = last
		}
		if marker < 0 {
			continue
		}
		depth := len(line[:marker])/3 + 1
		coordinate := strings.TrimSpace(line[marker+2:])
		parts := strings.Split(coordinate, ":")
		if len(parts) < 4 {
			continue
		}
		groupID, artifactID, version := parts[0], parts[1], parts[3]
		if groupID == "" || artifactID == "" {
			continue
		}
		if len(stack) >= depth {
			stack = stack[:depth-1]
		}
		if len(stack) < depth-1 {
			stack = stack[:0]
		}
		label := groupID + ":" + artifactID + "@" + version
		path := append(append([]string{}, stack...), label)
		scope := "compile"
		if len(parts) >= 5 && parts[4] != "" {
			scope = parts[4]
		}
		purl := mavenPURL(groupID, artifactID, version)
		if !seen[purl] {
			components = append(components, model.Component{
				Ecosystem:      "maven",
				GroupID:        groupID,
				ArtifactID:     artifactID,
				Version:        version,
				PURL:           purl,
				Direct:         depth == 1,
				DependencyPath: strings.Join(path, " -> "),
				Scope:          scope,
			})
			seen[purl] = true
		}
		stack = path
	}
	sortComponents(components)
	return components
}

func safeEnvironment() []string {
	allowedExact := map[string]bool{
		"PATH": true, "JAVA_HOME": true, "MAVEN_HOME": true,
		"SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true,
	}
	result := make([]string, 0)
	for _, value := range os.Environ() {
		key := strings.SplitN(value, "=", 2)[0]
		upper := strings.ToUpper(key)
		if allowedExact[upper] || strings.HasPrefix(upper, "LC_") {
			result = append(result, value)
		}
	}
	result = append(result, "HOME=/tmp/maven-home", "MAVEN_CONFIG=/tmp/maven-home/.m2")
	return result
}
