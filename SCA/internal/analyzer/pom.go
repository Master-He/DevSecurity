package analyzer

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devsecurity/sca/internal/model"
)

var ErrNoPom = errors.New("no pom.xml found")

type projectXML struct {
	XMLName      xml.Name          `xml:"project"`
	GroupID      string            `xml:"groupId"`
	ArtifactID   string            `xml:"artifactId"`
	Version      string            `xml:"version"`
	Parent       parentXML         `xml:"parent"`
	Dependencies []dependencyXML   `xml:"dependencies>dependency"`
	Properties   projectProperties `xml:"properties"`
}

type parentXML struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type dependencyXML struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
}

type projectProperties map[string]string

func (p *projectProperties) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	values := make(map[string]string)
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch element := token.(type) {
		case xml.StartElement:
			var value string
			if err := decoder.DecodeElement(&value, &element); err != nil {
				return err
			}
			values[element.Name.Local] = strings.TrimSpace(value)
		case xml.EndElement:
			if element.Name == start.Name {
				*p = values
				return nil
			}
		}
	}
}

func FindPoms(root string) ([]string, error) {
	var poms []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "target" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "pom.xml" {
			info, infoErr := entry.Info()
			if infoErr != nil {
				return infoErr
			}
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return fmt.Errorf("pom.xml must be a regular file and not a symbolic link")
			}
			poms = append(poms, path)
		}
		return nil
	})
	sort.Slice(poms, func(i, j int) bool {
		return len(poms[i]) < len(poms[j])
	})
	if err != nil {
		return nil, err
	}
	if len(poms) == 0 {
		return nil, ErrNoPom
	}
	return poms, nil
}

func ParsePoms(root string) ([]model.Component, error) {
	poms, err := FindPoms(root)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]model.Component)
	for _, pom := range poms {
		components, err := parsePOM(root, pom)
		if err != nil {
			continue
		}
		for _, component := range components {
			existing, ok := seen[component.PURL]
			if !ok || (component.Direct && !existing.Direct) {
				seen[component.PURL] = component
			}
		}
	}
	components := make([]model.Component, 0, len(seen))
	for _, component := range seen {
		components = append(components, component)
	}
	sortComponents(components)
	return components, nil
}

func parsePOM(root, path string) ([]model.Component, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var project projectXML
	if err := xml.Unmarshal(content, &project); err != nil {
		return nil, err
	}
	properties := make(map[string]string)
	for key, value := range project.Properties {
		properties[key] = value
	}
	projectGroup := firstNonBlank(project.GroupID, project.Parent.GroupID)
	projectVersion := firstNonBlank(project.Version, project.Parent.Version)
	properties["project.groupId"] = projectGroup
	properties["pom.groupId"] = projectGroup
	properties["project.version"] = projectVersion
	properties["pom.version"] = projectVersion
	properties["project.artifactId"] = project.ArtifactID

	relative, err := filepath.Rel(root, path)
	if err != nil {
		relative = "pom.xml"
	}
	relative = filepath.ToSlash(relative)
	components := make([]model.Component, 0, len(project.Dependencies))
	for _, dependency := range project.Dependencies {
		groupID := resolve(dependency.GroupID, properties)
		artifactID := resolve(dependency.ArtifactID, properties)
		version := strings.TrimSpace(dependency.Version)
		if version == "" {
			version = "unknown"
		} else {
			version = resolve(version, properties)
			if version == "" {
				version = "unknown"
			}
		}
		if groupID == "" {
			groupID = "unknown"
		}
		if artifactID == "" {
			artifactID = "unknown"
		}
		component := model.Component{
			Ecosystem:      "maven",
			GroupID:        groupID,
			ArtifactID:     artifactID,
			Version:        version,
			PURL:           mavenPURL(groupID, artifactID, version),
			Direct:         true,
			DependencyPath: fmt.Sprintf("%s -> %s:%s@%s", relative, groupID, artifactID, version),
			Scope:          firstNonBlank(dependency.Scope, "compile"),
		}
		components = append(components, component)
	}
	return components, nil
}

func resolve(value string, properties map[string]string) string {
	result := strings.TrimSpace(value)
	for iteration := 0; iteration < 8; iteration++ {
		changed := false
		for key, replacement := range properties {
			placeholder := "${" + key + "}"
			if strings.Contains(result, placeholder) {
				result = strings.ReplaceAll(result, placeholder, replacement)
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	if strings.Contains(result, "${") {
		return ""
	}
	return result
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mavenPURL(groupID, artifactID, version string) string {
	return "pkg:maven/" + groupID + "/" + artifactID + "@" + version
}

func sortComponents(components []model.Component) {
	sort.Slice(components, func(i, j int) bool {
		if components[i].GroupID == components[j].GroupID {
			if components[i].ArtifactID == components[j].ArtifactID {
				return components[i].Version < components[j].Version
			}
			return components[i].ArtifactID < components[j].ArtifactID
		}
		return components[i].GroupID < components[j].GroupID
	})
}
