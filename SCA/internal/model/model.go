package model

type Component struct {
	Ecosystem      string `json:"ecosystem"`
	GroupID        string `json:"groupId"`
	ArtifactID     string `json:"artifactId"`
	Version        string `json:"version"`
	PURL           string `json:"purl"`
	Direct         bool   `json:"direct"`
	DependencyPath string `json:"dependencyPath"`
	Scope          string `json:"scope"`
}

type Finding struct {
	VulnerabilityID string   `json:"vulnerabilityId"`
	Aliases         []string `json:"aliases"`
	Title           string   `json:"title"`
	Severity        string   `json:"severity"`
	PackageName     string   `json:"packageName"`
	AffectedVersion string   `json:"affectedVersion"`
	AffectedRange   string   `json:"affectedRange"`
	FixedVersions   []string `json:"fixedVersions"`
	PURL            string   `json:"purl"`
	DependencyPath  string   `json:"dependencyPath"`
}

type Result struct {
	Engine        string      `json:"engine"`
	EngineVersion string      `json:"engineVersion"`
	SourceMode    string      `json:"sourceMode"`
	Components    []Component `json:"components"`
	Findings      []Finding   `json:"findings"`
	Diagnostics   []string    `json:"diagnostics"`
}
