package vuln

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"devsecurity/sca/internal/model"
)

type Record struct {
	ID               string   `json:"id"`
	Aliases          []string `json:"aliases"`
	Title            string   `json:"title"`
	Severity         string   `json:"severity"`
	Package          string   `json:"package"`
	AffectedRange    string   `json:"affectedRange"`
	Introduced       string   `json:"introduced"`
	Fixed            string   `json:"fixed"`
	FixedVersions    []string `json:"fixedVersions"`
	AffectedVersions []string `json:"affectedVersions"`
}

type Database struct {
	records []Record
}

func Load(path string) (*Database, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read vulnerability database: %w", err)
	}
	var records []Record
	if err := json.Unmarshal(content, &records); err != nil {
		return nil, fmt.Errorf("parse vulnerability database: %w", err)
	}
	for index := range records {
		records[index].ID = strings.TrimSpace(records[index].ID)
		records[index].Package = strings.TrimSpace(records[index].Package)
		if records[index].ID == "" || records[index].Package == "" {
			return nil, fmt.Errorf("vulnerability record %d is missing id or package", index)
		}
	}
	return &Database{records: records}, nil
}

func (d *Database) Match(components []model.Component) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, component := range components {
		if component.Version == "unknown" {
			continue
		}
		packageName := component.GroupID + ":" + component.ArtifactID
		for _, record := range d.records {
			if record.Package != packageName || !record.affects(component.Version) {
				continue
			}
			findings = append(findings, model.Finding{
				VulnerabilityID: record.ID,
				Aliases:         nonNil(record.Aliases),
				Title:           record.Title,
				Severity:        record.Severity,
				PackageName:     packageName,
				AffectedVersion: component.Version,
				AffectedRange:   record.AffectedRange,
				FixedVersions:   nonNil(record.FixedVersions),
				PURL:            component.PURL,
				DependencyPath:  component.DependencyPath,
			})
		}
	}
	return findings
}

func (r Record) affects(version string) bool {
	if len(r.AffectedVersions) > 0 && containsExact(r.AffectedVersions, version) {
		return true
	}
	if r.Introduced == "" && r.Fixed == "" {
		return false
	}
	if r.Introduced != "" && CompareVersions(version, r.Introduced) < 0 {
		return false
	}
	if r.Fixed != "" && CompareVersions(version, r.Fixed) >= 0 {
		return false
	}
	return true
}

func containsExact(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// CompareVersions is intentionally conservative for the initial Maven scanner.
// It compares numeric components and treats a release as newer than a qualifier.
func CompareVersions(left, right string) int {
	leftParts := versionParts(left)
	rightParts := versionParts(right)
	length := len(leftParts)
	if len(rightParts) > length {
		length = len(rightParts)
	}
	for index := 0; index < length; index++ {
		var a, b token
		if index < len(leftParts) {
			a = leftParts[index]
		}
		if index < len(rightParts) {
			b = rightParts[index]
		}
		result := compareToken(a, b)
		if result != 0 {
			return result
		}
	}
	return 0
}

type token struct {
	number int64
	text   string
}

func versionParts(value string) []token {
	value = strings.TrimSpace(strings.ToLower(value))
	var parts []token
	for index := 0; index < len(value); {
		start := index
		isDigit := value[index] >= '0' && value[index] <= '9'
		for index < len(value) {
			digit := value[index] >= '0' && value[index] <= '9'
			if digit != isDigit {
				break
			}
			index++
		}
		piece := value[start:index]
		if isDigit {
			number, _ := strconv.ParseInt(piece, 10, 64)
			parts = append(parts, token{number: number})
		} else {
			parts = append(parts, token{text: piece})
		}
		for index < len(value) && (value[index] == '.' || value[index] == '-' || value[index] == '+') {
			index++
		}
	}
	return parts
}

func compareToken(left, right token) int {
	if left.text != "" || right.text != "" {
		if left.text == "" {
			if left.number == 0 {
				return -1
			}
			return 1
		}
		if right.text == "" {
			if right.number == 0 {
				return 1
			}
			return -1
		}
		return strings.Compare(left.text, right.text)
	}
	if left.number < right.number {
		return -1
	}
	if left.number > right.number {
		return 1
	}
	return 0
}
