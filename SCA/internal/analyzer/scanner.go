package analyzer

import (
	"context"
	"fmt"
	"time"

	"devsecurity/sca/internal/model"
	"devsecurity/sca/internal/version"
	"devsecurity/sca/internal/vuln"
)

type Scanner struct {
	UseMaven bool
	Timeout  time.Duration
	Database *vuln.Database
}

func (s Scanner) Analyze(ctx context.Context, root string) (model.Result, error) {
	poms, err := FindPoms(root)
	if err != nil {
		return model.Result{}, err
	}
	diagnostics := make([]string, 0)
	components := []model.Component{}
	mode := "POM_FALLBACK"

	if s.UseMaven && MavenAvailable() {
		tree, treeErr := GenerateDependencyTree(ctx, poms[0], s.Timeout)
		if treeErr == nil {
			components = tree
			mode = "MAVEN_TREE"
		} else {
			diagnostics = append(diagnostics, fmt.Sprintf(
				"mvn dependency:tree 未成功，已回退到 pom.xml 解析: %s", treeErr.Error()))
		}
	} else if s.UseMaven {
		diagnostics = append(diagnostics, "环境中没有 mvn，使用 pom.xml 解析")
	}

	if len(components) == 0 {
		components, err = ParsePoms(root)
		if err != nil {
			return model.Result{}, err
		}
	}
	return model.Result{
		Engine:        "sca-maven",
		EngineVersion: version.Current,
		SourceMode:    mode,
		Components:    components,
		Findings:      s.Database.Match(components),
		Diagnostics:   diagnostics,
	}, nil
}
