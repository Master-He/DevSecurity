package vuln

import (
	"testing"

	"devsecurity/sca/internal/model"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		left, right string
		expected    int
	}{
		{"2.14.1", "2.15.0", -1},
		{"2.0", "2.0.0", 0},
		{"4.1.109.Final", "4.1.109", 1},
		{"10.0.0", "9.9.9", 1},
	}
	for _, test := range cases {
		actual := CompareVersions(test.left, test.right)
		if actual < 0 {
			actual = -1
		} else if actual > 0 {
			actual = 1
		}
		if actual != test.expected {
			t.Errorf("compare(%s, %s)=%d, want %d", test.left, test.right, actual, test.expected)
		}
	}
}

func TestDatabaseMatchesMavenComponent(t *testing.T) {
	database := &Database{records: []Record{{
		ID:            "CVE-TEST-1",
		Title:         "test",
		Severity:      "HIGH",
		Package:       "org.example:demo",
		AffectedRange: ">=1.0, <2.0",
		Introduced:    "1.0",
		Fixed:         "2.0",
		FixedVersions: []string{"2.0"},
	}}}
	findings := database.Match([]model.Component{{
		Ecosystem: "maven", GroupID: "org.example", ArtifactID: "demo",
		Version: "1.5.0", PURL: "pkg:maven/org.example/demo@1.5.0",
	}})
	if len(findings) != 1 || findings[0].VulnerabilityID != "CVE-TEST-1" {
		t.Fatalf("expected one known finding, got %#v", findings)
	}
}
